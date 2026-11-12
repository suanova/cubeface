package service

import (
	"sort"
	"strings"
	"time"

	"github.com/suanova/medrouter/internal/kube"
)

const (
	annotationAccess           = "medrouter.io/access"
	annotationDeidentification = "medrouter.io/deidentification"
	annotationEthics           = "medrouter.io/ethics"
	annotationSourceRepo       = "medrouter.io/source-repo"
)

// BuildDatasetList projects raw Dataset CRs into the list response. It is pure
// and is the single place where annotation/fallback rules live.
func BuildDatasetList(datasets []kube.Dataset, now time.Time) *DatasetList {
	list := &DatasetList{
		DatasetCatalogStatus: DatasetCatalogStatus{
			Available:   true,
			GeneratedAt: now.UTC(),
		},
		Datasets: []DatasetSummary{},
	}

	groups := groupDatasets(datasets)
	list.VersionTotal = len(datasets)
	for _, group := range groups {
		list.Datasets = append(list.Datasets, summarizeDatasetGroup(group))
	}
	sort.SliceStable(list.Datasets, func(i, j int) bool {
		a, b := list.Datasets[i], list.Datasets[j]
		if a.Featured != b.Featured {
			return a.Featured
		}
		if !a.UpdatedAt.Equal(b.UpdatedAt) {
			return a.UpdatedAt.After(b.UpdatedAt)
		}
		return strings.ToLower(a.DisplayName) < strings.ToLower(b.DisplayName)
	})
	list.Total = len(list.Datasets)
	return list
}

// BuildDatasetDetail finds ref in the projected list and expands it. ref may be
// spec.dataset (the public page identity) or a Dataset metadata.name.
func BuildDatasetDetail(list *DatasetList, datasets []kube.Dataset, ref string) (*DatasetDetail, bool) {
	groups := groupDatasets(datasets)
	group, ok := findDatasetGroup(groups, ref)
	if !ok {
		return nil, false
	}

	detail := &DatasetDetail{
		DatasetCatalogStatus: list.DatasetCatalogStatus,
		DatasetSummary:       summarizeDatasetGroup(group),
		LatestRef:            group.versions[0].Metadata.Name,
		Versions:             []DatasetVersionDetail{},
	}
	for i := range group.versions {
		detail.Versions = append(detail.Versions, datasetVersionDetail(group.versions[i], i == 0))
	}
	return detail, true
}

// summarizeDatasetGroup returns the dataset-level card built from the newest
// version.
func summarizeDatasetGroup(group datasetGroup) DatasetSummary {
	latest := group.versions[0]
	annotations := latest.Metadata.Annotations

	summary := DatasetSummary{
		Dataset:          latest.Spec.Dataset,
		DisplayName:      annotationText(annotations, annotationDisplayName, 200),
		Publisher:        annotationText(annotations, annotationPublisher, 120),
		Description:      annotationText(annotations, annotationDescription, maxDescription),
		Format:           latest.Spec.Format,
		Task:             annotationText(annotations, annotationTask, 80),
		License:          annotationText(annotations, annotationLicense, 80),
		Access:           annotationText(annotations, annotationAccess, 80),
		Deidentification: annotationText(annotations, annotationDeidentification, 120),
		Ethics:           annotationText(annotations, annotationEthics, 240),
		SourceRepo:       annotationText(annotations, annotationSourceRepo, 240),
		UsageBoundary:    annotationText(annotations, annotationUsage, 240),
		Size:             strings.TrimSpace(latest.Spec.Size),
		RecordCount:      copyInt64(latest.Spec.RecordCount),
		StorageStrategy:  latest.Spec.Storage.Strategy,
		LatestVersion:    latest.Spec.Version,
		VersionCount:     len(group.versions),
		Featured:         strings.EqualFold(strings.TrimSpace(annotations[annotationFeatured]), "true"),
		Tags:             datasetTags(annotations, latest),
	}
	if summary.DisplayName == "" {
		summary.DisplayName = latest.Spec.Dataset
	}
	summary.PublisherInitials = initials(summary.Publisher, summary.DisplayName)
	summary.CreatedAt, summary.UpdatedAt = datasetGroupTimes(group)
	return summary
}

type datasetGroup struct {
	dataset  string
	versions []kube.Dataset // newest/latest first
}

func groupDatasets(datasets []kube.Dataset) []datasetGroup {
	byDataset := map[string][]kube.Dataset{}
	order := make([]string, 0, len(datasets))
	for _, ds := range datasets {
		key := ds.Spec.Dataset
		if _, seen := byDataset[key]; !seen {
			order = append(order, key)
		}
		byDataset[key] = append(byDataset[key], ds)
	}
	groups := make([]datasetGroup, 0, len(byDataset))
	for _, key := range order {
		versions := byDataset[key]
		sort.SliceStable(versions, func(i, j int) bool {
			if c := compareVersions(versions[i].Spec.Version, versions[j].Spec.Version); c != 0 {
				return c > 0
			}
			ti, tj := versions[i].Metadata.CreationTimestamp, versions[j].Metadata.CreationTimestamp
			if !ti.Equal(tj) {
				return ti.After(tj)
			}
			return versions[i].Metadata.Name > versions[j].Metadata.Name
		})
		groups = append(groups, datasetGroup{dataset: key, versions: versions})
	}
	sort.SliceStable(groups, func(i, j int) bool { return groups[i].dataset < groups[j].dataset })
	return groups
}

func findDatasetGroup(groups []datasetGroup, ref string) (datasetGroup, bool) {
	ref = strings.TrimSpace(ref)
	if ref == "" {
		return datasetGroup{}, false
	}
	for _, group := range groups {
		if group.dataset == ref {
			return group, true
		}
	}
	for _, group := range groups {
		for _, version := range group.versions {
			if version.Metadata.Name == ref {
				return group, true
			}
		}
	}
	// URLs are case-sensitive, but accepting a case-insensitive identity makes
	// hand-typed references forgiving without merging distinct groups.
	for _, group := range groups {
		if strings.EqualFold(group.dataset, ref) {
			return group, true
		}
	}
	return datasetGroup{}, false
}

func datasetGroupTimes(group datasetGroup) (created, updated time.Time) {
	for _, version := range group.versions {
		t := version.Metadata.CreationTimestamp
		if !t.IsZero() && (created.IsZero() || t.Before(created)) {
			created = t
		}
		if t.After(updated) {
			updated = t
		}
	}
	return created, updated
}

func datasetVersionDetail(ds kube.Dataset, latest bool) DatasetVersionDetail {
	return DatasetVersionDetail{
		Ref:         ds.Metadata.Name,
		Version:     ds.Spec.Version,
		IsLatest:    latest,
		Format:      ds.Spec.Format,
		Size:        strings.TrimSpace(ds.Spec.Size),
		RecordCount: copyInt64(ds.Spec.RecordCount),
		Description: annotationText(ds.Metadata.Annotations, annotationDescription, maxDescription),
		Storage:     storageSummary(ds.Spec.Storage),
		CreatedAt:   ds.Metadata.CreationTimestamp.UTC(),
	}
}

func datasetTags(annotations map[string]string, ds kube.Dataset) []string {
	out := make([]string, 0, 4)
	seen := map[string]struct{}{}
	add := func(raw string) {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			return
		}
		raw = truncateText(stripControl(raw), maxTagLength)
		key := strings.ToLower(raw)
		if _, ok := seen[key]; ok {
			return
		}
		seen[key] = struct{}{}
		out = append(out, raw)
	}
	if raw := annotations[annotationTags]; raw != "" {
		for i, part := range strings.Split(raw, ",") {
			if i >= maxTagsInput || len(out) >= maxTagCount {
				break
			}
			add(part)
		}
	}
	if len(out) == 0 {
		add(ds.Spec.Format)
		add(ds.Spec.Storage.Strategy)
	}
	return out
}

// copyInt64 detaches the value so callers cannot mutate the cached CR data.
func copyInt64(in *int64) *int64 {
	if in == nil {
		return nil
	}
	v := *in
	return &v
}

func datasetCatalogMessage(reason string) string {
	switch reason {
	case CatalogDisabled:
		return "The dataset library is disabled."
	case CatalogNoCluster:
		return "The dataset library has no Kubernetes API access. Run in-cluster or set DATASET_LIBRARY_API_URL."
	case CatalogCRDMissing:
		return "The CubeStack Dataset CRD is not installed in this cluster."
	case CatalogForbidden:
		return "The dataset library cannot list CubeStack datasets. Check the ServiceAccount RBAC."
	case CatalogUnauthorized:
		return "The Kubernetes API rejected the service account credentials."
	case CatalogTimeout:
		return "The Kubernetes API did not respond in time."
	default:
		return "The dataset library could not read CubeStack resources."
	}
}
