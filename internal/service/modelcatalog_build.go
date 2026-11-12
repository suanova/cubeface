package service

import (
	"errors"
	"sort"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/suanova/medrouter/internal/kube"
)

const (
	annotationDescription = "ai.cubestack.io/description"
	annotationDisplayName = "medrouter.io/display-name"
	annotationPublisher   = "medrouter.io/publisher"
	annotationTask        = "medrouter.io/task"
	annotationLicense     = "medrouter.io/license"
	annotationTags        = "medrouter.io/tags"
	annotationUsage       = "medrouter.io/usage-boundary"
	annotationFeatured    = "medrouter.io/featured"

	labelServingMode = "inference.ai.cubestack.io/serving-mode"

	conditionStorageResolved = "StorageResolved"
	conditionInUse           = "InUse"
	conditionAssetsResolved  = "AssetsResolved"
	conditionDeprecated      = "ProfileDeprecated"
	conditionDrifted         = "ProfileDrifted"

	maxDescription = 2048
	maxTagCount    = 12
	maxTagLength   = 64
	maxTagsInput   = 64
)

// BuildCatalogList projects raw CRs into the list response. It is pure and is
// the single place where annotation/fallback rules live.
func BuildCatalogList(mvs []kube.ModelVersion, profiles []kube.RuntimeProfile, profilesErr error, now time.Time) *ModelList {
	list := &ModelList{
		CatalogStatus: CatalogStatus{
			Available:         true,
			ProfilesAvailable: profilesErr == nil,
			GeneratedAt:       now.UTC(),
		},
		Models: []ModelSummary{},
	}
	if profilesErr != nil {
		reason, _ := statusFromError(profilesErr)
		list.Warning = "Runtime profiles are unavailable: " + catalogMessage(reason)
	}

	groups := groupModelVersions(mvs)
	list.VersionTotal = len(mvs)
	if profilesErr == nil {
		list.ProfileTotal = len(profiles)
	}

	for _, group := range groups {
		list.Models = append(list.Models, summarizeGroup(group, profiles))
	}
	sort.SliceStable(list.Models, func(i, j int) bool {
		a, b := list.Models[i], list.Models[j]
		if a.Featured != b.Featured {
			return a.Featured
		}
		if !a.UpdatedAt.Equal(b.UpdatedAt) {
			return a.UpdatedAt.After(b.UpdatedAt)
		}
		return strings.ToLower(a.DisplayName) < strings.ToLower(b.DisplayName)
	})
	list.Total = len(list.Models)
	return list
}

// BuildModelDetail finds ref in the projected list and expands it. ref may be
// spec.model (the public page identity) or a ModelVersion metadata.name.
func BuildModelDetail(list *ModelList, mvs []kube.ModelVersion, profiles []kube.RuntimeProfile, ref string) (*ModelDetail, bool) {
	groups := groupModelVersions(mvs)
	group, ok := findGroup(groups, ref)
	if !ok {
		return nil, false
	}
	summary := summarizeGroup(group, profiles)
	compatibleVersions := map[string][]string{}
	for i := range group.versions {
		// group.versions is newest-first, so versions appear in that order.
		v := group.versions[i]
		for _, profile := range profiles {
			if !compatible(profile, v) {
				continue
			}
			name := profile.Metadata.Name
			// Two ModelVersions may legitimately share spec.version; list each
			// version name only once per profile.
			if !containsString(compatibleVersions[name], v.Spec.Version) {
				compatibleVersions[name] = append(compatibleVersions[name], v.Spec.Version)
			}
		}
	}

	detail := &ModelDetail{
		CatalogStatus:      list.CatalogStatus,
		ModelSummary:       summary,
		LatestRef:          group.versions[0].Metadata.Name,
		Versions:           []ModelVersionDetail{},
		CompatibleProfiles: []RuntimeProfileSummary{},
		UsedBy:             unionUsedBy(group),
	}

	names := make([]string, 0, len(compatibleVersions))
	for name := range compatibleVersions {
		names = append(names, name)
	}
	sort.Strings(names)
	profileByName := make(map[string]kube.RuntimeProfile, len(profiles))
	for _, p := range profiles {
		profileByName[p.Metadata.Name] = p
	}
	for _, name := range names {
		if profile, ok := profileByName[name]; ok {
			detail.CompatibleProfiles = append(detail.CompatibleProfiles, profileSummary(profile, compatibleVersions[name]))
		}
	}

	for i := range group.versions {
		v := group.versions[i]
		detail.Versions = append(detail.Versions, versionDetail(v, i == 0, profiles))
	}
	return detail, true
}

// summarizeGroup returns the model-level card built from the newest version.
func summarizeGroup(group modelGroup, profiles []kube.RuntimeProfile) ModelSummary {
	latest := group.versions[0]
	annotations := latest.Metadata.Annotations

	summary := ModelSummary{
		Model:           latest.Spec.Model,
		DisplayName:     annotationText(annotations, annotationDisplayName, 200),
		Publisher:       annotationText(annotations, annotationPublisher, 120),
		Description:     annotationText(annotations, annotationDescription, maxDescription),
		Architecture:    latest.Spec.Architecture,
		Quantization:    latest.Spec.Quantization,
		Task:            annotationText(annotations, annotationTask, 80),
		License:         annotationText(annotations, annotationLicense, 80),
		UsageBoundary:   annotationText(annotations, annotationUsage, 240),
		StorageStrategy: latest.Spec.Storage.Strategy,
		LatestVersion:   latest.Spec.Version,
		VersionCount:    len(group.versions),
		Featured:        strings.EqualFold(strings.TrimSpace(annotations[annotationFeatured]), "true"),
		Tags:            tags(annotations, latest),
	}
	if summary.DisplayName == "" {
		summary.DisplayName = latest.Spec.Model
	}
	summary.PublisherInitials = initials(summary.Publisher, summary.DisplayName)
	summary.CreatedAt, summary.UpdatedAt = groupTimes(group)

	compatibleProfiles := map[string]struct{}{}
	for _, version := range group.versions {
		for _, profile := range profiles {
			if compatible(profile, version) {
				compatibleProfiles[profile.Metadata.Name] = struct{}{}
			}
		}
	}
	summary.CompatibleProfileCount = len(compatibleProfiles)
	summary.UsedByCount = len(unionUsedBy(group))
	return summary
}

type modelGroup struct {
	model    string
	versions []kube.ModelVersion // newest/latest first
}

func groupModelVersions(mvs []kube.ModelVersion) []modelGroup {
	byModel := map[string][]kube.ModelVersion{}
	order := make([]string, 0, len(mvs))
	for _, mv := range mvs {
		key := mv.Spec.Model
		if _, seen := byModel[key]; !seen {
			order = append(order, key)
		}
		byModel[key] = append(byModel[key], mv)
	}
	groups := make([]modelGroup, 0, len(byModel))
	for _, key := range order {
		versions := byModel[key]
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
		groups = append(groups, modelGroup{model: key, versions: versions})
	}
	sort.SliceStable(groups, func(i, j int) bool { return groups[i].model < groups[j].model })
	return groups
}

func findGroup(groups []modelGroup, ref string) (modelGroup, bool) {
	ref = strings.TrimSpace(ref)
	if ref == "" {
		return modelGroup{}, false
	}
	for _, group := range groups {
		if group.model == ref {
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
		if strings.EqualFold(group.model, ref) {
			return group, true
		}
	}
	return modelGroup{}, false
}

func groupTimes(group modelGroup) (created, updated time.Time) {
	for _, version := range group.versions {
		t := version.Metadata.CreationTimestamp
		if !t.IsZero() && (created.IsZero() || t.Before(created)) {
			created = t
		}
		if t.After(updated) {
			updated = t
		}
		for _, c := range version.Status.Conditions {
			if c.LastTransitionTime.After(updated) {
				updated = c.LastTransitionTime
			}
		}
	}
	return created, updated
}

func versionDetail(mv kube.ModelVersion, latest bool, profiles []kube.RuntimeProfile) ModelVersionDetail {
	names := make([]string, 0)
	for _, profile := range profiles {
		if compatible(profile, mv) {
			names = append(names, profile.Metadata.Name)
		}
	}
	sort.Strings(names)
	usedBy := make([]UsedByRef, 0, len(mv.Status.UsedBy))
	for _, ref := range mv.Status.UsedBy {
		usedBy = append(usedBy, UsedByRef{Namespace: ref.Namespace, Name: ref.Name, Version: mv.Spec.Version})
	}
	return ModelVersionDetail{
		Ref:                mv.Metadata.Name,
		Version:            mv.Spec.Version,
		IsLatest:           latest,
		Architecture:       mv.Spec.Architecture,
		Quantization:       mv.Spec.Quantization,
		Description:        annotationText(mv.Metadata.Annotations, annotationDescription, maxDescription),
		Storage:            storageSummary(mv.Spec.Storage),
		StorageResolved:    conditionState(mv.Status.Conditions, conditionStorageResolved),
		InUse:              conditionState(mv.Status.Conditions, conditionInUse),
		RootPathResolved:   strings.TrimSpace(mv.Status.RootPath) != "",
		UsedBy:             usedBy,
		CompatibleProfiles: names,
		CreatedAt:          mv.Metadata.CreationTimestamp.UTC(),
	}
}

func compatible(profile kube.RuntimeProfile, mv kube.ModelVersion) bool {
	return contains(profile.Spec.ModelRequirements.Architectures, mv.Spec.Architecture) &&
		contains(profile.Spec.ModelRequirements.Quantization, mv.Spec.Quantization)
}

func contains(values []string, want string) bool {
	for _, v := range values {
		if v == want {
			return true
		}
	}
	return false
}

func storageSummary(storage kube.ModelStorage) StorageSummary {
	out := StorageSummary{Strategy: storage.Strategy}
	switch storage.Strategy {
	case "HostPath":
		out.PathConfigured = storage.HostPath != nil && strings.TrimSpace(storage.HostPath.Path) != ""
	case "Dynamic":
		if storage.Dynamic != nil {
			out.StorageClass = storage.Dynamic.StorageClassName
			out.Capacity = storage.Dynamic.Capacity
		}
	case "Static":
		if storage.Static != nil {
			out.StorageClass = storage.Static.StorageClassName
			out.Capacity = storage.Static.Capacity
		}
	case "S3":
		if storage.S3 != nil {
			out.URIConfigured = strings.TrimSpace(storage.S3.URI) != ""
			out.CredentialsConfigured = storage.S3.CredentialsRef != nil && strings.TrimSpace(storage.S3.CredentialsRef.Name) != ""
		}
	}
	return out
}

func profileSummary(profile kube.RuntimeProfile, compatibleVersions []string) RuntimeProfileSummary {
	out := RuntimeProfileSummary{
		Name:               profile.Metadata.Name,
		Description:        annotationText(profile.Metadata.Annotations, annotationDescription, maxDescription),
		Engine:             EngineSummary{Name: profile.Spec.Engine.Name, Version: profile.Spec.Engine.Version},
		Vendor:             profile.Spec.Accelerator.Vendor,
		AcceleratorModels:  cloneStrings(profile.Spec.Accelerator.Models),
		Architectures:      cloneStrings(profile.Spec.ModelRequirements.Architectures),
		Quantizations:      cloneStrings(profile.Spec.ModelRequirements.Quantization),
		CompatibleVersions: cloneStrings(compatibleVersions),
		AssetsResolved:     conditionState(profile.Status.Conditions, conditionAssetsResolved),
		Deprecated:         conditionState(profile.Status.Conditions, conditionDeprecated),
		Drifted:            conditionState(profile.Status.Conditions, conditionDrifted),
		UsedByCount:        len(profile.Status.UsedBy),
		Roles:              []ProfileRoleSummary{},
	}
	switch mode := profile.Metadata.Labels[labelServingMode]; mode {
	case "standard", "pd-separation":
		out.ServingMode = mode
	}
	maxGPU := 0
	hasGPU := false
	for _, role := range profile.Spec.Roles {
		accelerator := profile.Spec.Accelerator
		if role.Accelerator != nil {
			accelerator = *role.Accelerator
		}
		var gpu *int
		if role.PodTemplate.Resources != nil && role.PodTemplate.Resources.GPUPerPod != nil {
			if n, ok := role.PodTemplate.Resources.GPUPerPod.Number(); ok && n >= 0 {
				value := n
				gpu = &value
				if !hasGPU || n > maxGPU {
					maxGPU = n
					hasGPU = true
				}
			}
		}
		out.Roles = append(out.Roles, ProfileRoleSummary{
			Name:              role.Name,
			Vendor:            accelerator.Vendor,
			AcceleratorModels: cloneStrings(accelerator.Models),
			GPUPerPod:         gpu,
		})
	}
	if hasGPU {
		value := maxGPU
		out.GPUPerPod = &value
	}
	return out
}

func unionUsedBy(group modelGroup) []UsedByRef {
	seen := map[string]struct{}{}
	out := make([]UsedByRef, 0)
	// versions are newest-first; the first reference to a workload wins, so it
	// carries the newest version label.
	for _, version := range group.versions {
		for _, ref := range version.Status.UsedBy {
			key := ref.Namespace + "/" + ref.Name
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			out = append(out, UsedByRef{Namespace: ref.Namespace, Name: ref.Name, Version: version.Spec.Version})
		}
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Namespace != out[j].Namespace {
			return out[i].Namespace < out[j].Namespace
		}
		return out[i].Name < out[j].Name
	})
	return out
}

func conditionState(conditions []kube.Condition, condType string) *bool {
	for _, c := range conditions {
		if c.Type != condType {
			continue
		}
		switch c.Status {
		case "True":
			v := true
			return &v
		case "False":
			v := false
			return &v
		default:
			return nil
		}
	}
	return nil
}

func tags(annotations map[string]string, mv kube.ModelVersion) []string {
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
		add(mv.Spec.Architecture)
		add(mv.Spec.Quantization)
		add(mv.Spec.Storage.Strategy)
	}
	return out
}

func annotationText(annotations map[string]string, key string, limit int) string {
	raw := strings.TrimSpace(annotations[key])
	if raw == "" {
		return ""
	}
	return truncateText(stripControl(raw), limit)
}

func stripControl(raw string) string {
	return strings.Map(func(r rune) rune {
		if r == '\n' || r == '\t' {
			return ' '
		}
		if unicode.IsControl(r) {
			return -1
		}
		return r
	}, raw)
}

func truncateText(raw string, limit int) string {
	if limit <= 0 || utf8.RuneCountInString(raw) <= limit {
		return raw
	}
	runes := []rune(raw)
	return string(runes[:limit])
}

func initials(publisher, fallback string) string {
	source := strings.TrimSpace(publisher)
	if source == "" {
		source = strings.TrimSpace(fallback)
	}
	if source == "" {
		return "?"
	}
	fields := strings.Fields(source)
	if len(fields) >= 2 {
		return strings.ToUpper(firstRune(fields[0]) + firstRune(fields[1]))
	}
	runes := []rune(source)
	if len(runes) == 1 {
		return strings.ToUpper(string(runes[0]))
	}
	return strings.ToUpper(string(runes[:2]))
}

func firstRune(s string) string {
	for _, r := range s {
		return string(r)
	}
	return ""
}

func cloneStrings(in []string) []string {
	if in == nil {
		return []string{}
	}
	return append([]string(nil), in...)
}

func containsString(haystack []string, needle string) bool {
	for _, value := range haystack {
		if value == needle {
			return true
		}
	}
	return false
}

// compareVersions is a natural comparison: digit runs compare numerically,
// other runs compare case-insensitively. It handles v1.9 > v1.10 and dates.
func compareVersions(a, b string) int {
	ra, rb := []rune(a), []rune(b)
	i, j := 0, 0
	for i < len(ra) && j < len(rb) {
		if unicode.IsDigit(ra[i]) && unicode.IsDigit(rb[j]) {
			si, sj := i, j
			for i < len(ra) && unicode.IsDigit(ra[i]) {
				i++
			}
			for j < len(rb) && unicode.IsDigit(rb[j]) {
				j++
			}
			if c := compareNumericRun(ra[si:i], rb[sj:j]); c != 0 {
				return c
			}
			continue
		}
		ca, cb := unicode.ToLower(ra[i]), unicode.ToLower(rb[j])
		if ca != cb {
			if ca < cb {
				return -1
			}
			return 1
		}
		i++
		j++
	}
	switch {
	case i < len(ra):
		return 1
	case j < len(rb):
		return -1
	default:
		return 0
	}
}

func compareNumericRun(a, b []rune) int {
	trim := func(r []rune) []rune {
		for len(r) > 1 && r[0] == '0' {
			r = r[1:]
		}
		return r
	}
	a, b = trim(a), trim(b)
	if len(a) != len(b) {
		if len(a) < len(b) {
			return -1
		}
		return 1
	}
	sa, sb := string(a), string(b)
	if sa < sb {
		return -1
	}
	if sa > sb {
		return 1
	}
	return 0
}

func statusFromError(err error) (string, bool) {
	var statusErr *kube.StatusError
	if errors.As(err, &statusErr) {
		return statusErr.Reason, statusErr.Retryable
	}
	if errors.Is(err, kube.ErrNoCluster) {
		return CatalogNoCluster, true
	}
	return CatalogAPIError, true
}

func catalogMessage(reason string) string {
	switch reason {
	case CatalogDisabled:
		return "The model library is disabled."
	case CatalogNoCluster:
		return "The model library has no Kubernetes API access. Run in-cluster or set MODEL_LIBRARY_API_URL."
	case CatalogCRDMissing:
		return "The CubeStack CRDs are not installed in this cluster."
	case CatalogForbidden:
		return "The model library cannot list CubeStack resources. Check the ServiceAccount RBAC."
	case CatalogUnauthorized:
		return "The Kubernetes API rejected the service account credentials."
	case CatalogTimeout:
		return "The Kubernetes API did not respond in time."
	default:
		return "The model library could not read CubeStack resources."
	}
}
