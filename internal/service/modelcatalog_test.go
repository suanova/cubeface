package service

import (
	"testing"
	"time"

	"github.com/cubeface/cubeface/internal/kube"
)

func testModelVersion(name, model, version, arch, quant string, created time.Time, ann map[string]string) kube.ModelVersion {
	return kube.ModelVersion{
		Metadata: kube.ObjectMeta{
			Name:              name,
			Annotations:       ann,
			CreationTimestamp: created,
		},
		Spec: kube.ModelVersionSpec{
			Model:        model,
			Version:      version,
			Architecture: arch,
			Quantization: quant,
			Storage: kube.ModelStorage{
				Strategy: "S3",
				S3: &kube.S3Storage{
					URI:            "s3://bucket/private/model",
					CredentialsRef: &kube.S3CredentialsRef{Name: "secret-name"},
				},
			},
		},
		Status: kube.ModelVersionStatus{
			RootPath: "/internal/resolved/path",
		},
	}
}

func intPtr(v int) *int { return &v }

func testProfile(name, arch, quant string) kube.RuntimeProfile {
	return kube.RuntimeProfile{
		Metadata: kube.ObjectMeta{
			Name:   name,
			Labels: map[string]string{labelServingMode: "pd-separation"},
			Annotations: map[string]string{
				annotationDescription: "A runtime profile",
			},
		},
		Spec: kube.RuntimeProfileSpec{
			Accelerator:       kube.Accelerator{Vendor: "nvidia", Models: []string{"H100"}},
			Engine:            kube.Engine{Name: "vllm", Version: "0.10.1"},
			ModelRequirements: kube.ModelRequirements{Architectures: []string{arch}, Quantization: []string{quant}},
			Roles: []kube.Role{
				{
					Name: "server",
					PodTemplate: kube.PodTemplate{
						Resources: &kube.PodResources{GPUPerPod: &kube.IntOrString{Int: intPtr(2)}},
					},
				},
			},
		},
		Status: kube.ProfileStatus{
			UsedBy: []kube.ObjectRef{{Namespace: "ml", Name: "svc"}},
			Conditions: []kube.Condition{
				{Type: conditionAssetsResolved, Status: "True"},
			},
		},
	}
}

func testCatalog() ([]kube.ModelVersion, []kube.RuntimeProfile, *ModelList) {
	base := time.Date(2026, 9, 1, 8, 0, 0, 0, time.UTC)
	ann := map[string]string{
		annotationDescription: "Qwen instruct model",
		annotationDisplayName: "Qwen2.5-7B-Instruct",
		annotationPublisher:   "Alibaba Cloud",
		annotationTask:        "chat",
		annotationLicense:     "Apache-2.0",
		annotationTags:        "chat, qwen2, chat",
		annotationUsage:       "Research and commercial use",
		annotationFeatured:    "true",
	}
	v19 := testModelVersion("qwen-v1-9", "qwen2.5-7b-instruct", "v1.9", "qwen2", "bf16", base, nil)
	v19.Status.UsedBy = []kube.ObjectRef{{Namespace: "ml", Name: "svc-a"}}
	v110 := testModelVersion("qwen-v1-10", "qwen2.5-7b-instruct", "v1.10", "qwen2", "bf16", base.Add(48*time.Hour), ann)
	v110.Status.UsedBy = []kube.ObjectRef{
		{Namespace: "ml", Name: "svc-a"},
		{Namespace: "ml", Name: "svc-b"},
	}
	v110.Status.Conditions = []kube.Condition{
		{Type: conditionStorageResolved, Status: "True", LastTransitionTime: base.Add(72 * time.Hour)},
		{Type: conditionInUse, Status: "False"},
	}
	profiles := []kube.RuntimeProfile{testProfile("nvidia-vllm", "qwen2", "bf16")}
	mvs := []kube.ModelVersion{v19, v110}
	return mvs, profiles, BuildCatalogList(mvs, profiles, nil, base.Add(96*time.Hour))
}

func TestBuildCatalogListGroupsAndProjects(t *testing.T) {
	_, _, list := testCatalog()

	if !list.Available || list.Total != 1 || list.VersionTotal != 2 || list.ProfileTotal != 1 {
		t.Fatalf("list = %#v", list)
	}
	m := list.Models[0]
	if m.Model != "qwen2.5-7b-instruct" || m.DisplayName != "Qwen2.5-7B-Instruct" {
		t.Fatalf("model fields = %#v", m)
	}
	if m.LatestVersion != "v1.10" || m.VersionCount != 2 {
		t.Fatalf("latest/version count = %s/%d", m.LatestVersion, m.VersionCount)
	}
	if m.Publisher != "Alibaba Cloud" || m.PublisherInitials != "AC" {
		t.Fatalf("publisher = %q initials = %q", m.Publisher, m.PublisherInitials)
	}
	if m.Description != "Qwen instruct model" || m.Task != "chat" || m.License != "Apache-2.0" {
		t.Fatalf("annotations = %#v", m)
	}
	if len(m.Tags) != 2 || m.Tags[0] != "chat" || m.Tags[1] != "qwen2" {
		t.Fatalf("tags = %#v", m.Tags)
	}
	if !m.Featured || m.CompatibleProfileCount != 1 || m.UsedByCount != 2 {
		t.Fatalf("counts = %#v", m)
	}
	if m.StorageStrategy != "S3" {
		t.Fatalf("storage = %s", m.StorageStrategy)
	}
	if !m.UpdatedAt.Equal(time.Date(2026, 9, 4, 8, 0, 0, 0, time.UTC)) {
		t.Fatalf("updated_at = %s", m.UpdatedAt)
	}
	if !m.CreatedAt.Equal(time.Date(2026, 9, 1, 8, 0, 0, 0, time.UTC)) {
		t.Fatalf("created_at = %s", m.CreatedAt)
	}
}

func TestBuildModelDetail(t *testing.T) {
	mvs, profiles, list := testCatalog()

	detail, ok := BuildModelDetail(list, mvs, profiles, "qwen2.5-7b-instruct")
	if !ok {
		t.Fatal("model not found by spec.model")
	}
	if detail.LatestRef != "qwen-v1-10" || len(detail.Versions) != 2 || len(detail.CompatibleProfiles) != 1 {
		t.Fatalf("detail = %#v", detail)
	}
	if detail.Versions[0].Version != "v1.10" || !detail.Versions[0].IsLatest {
		t.Fatalf("latest version = %#v", detail.Versions[0])
	}
	if detail.Versions[1].Version != "v1.9" || detail.Versions[1].IsLatest {
		t.Fatalf("second version = %#v", detail.Versions[1])
	}
	latest := detail.Versions[0]
	if latest.Storage.Strategy != "S3" || !latest.Storage.URIConfigured || !latest.Storage.CredentialsConfigured {
		t.Fatalf("storage summary = %#v", latest.Storage)
	}
	if latest.StorageResolved == nil || !*latest.StorageResolved {
		t.Fatalf("storage_resolved = %v", latest.StorageResolved)
	}
	if latest.InUse == nil || *latest.InUse {
		t.Fatalf("in_use = %v", latest.InUse)
	}
	if !latest.RootPathResolved {
		t.Fatal("root_path_resolved should be true")
	}
	if len(latest.CompatibleProfiles) != 1 || latest.CompatibleProfiles[0] != "nvidia-vllm" {
		t.Fatalf("compatible profiles = %#v", latest.CompatibleProfiles)
	}
	profilesOut := detail.CompatibleProfiles[0]
	if len(profilesOut.CompatibleVersions) != 2 || profilesOut.CompatibleVersions[0] != "v1.10" {
		t.Fatalf("compatible versions = %#v", profilesOut.CompatibleVersions)
	}
	if profilesOut.ServingMode != "pd-separation" || profilesOut.GPUPerPod == nil || *profilesOut.GPUPerPod != 2 {
		t.Fatalf("profile summary = %#v", profilesOut)
	}
	if len(detail.UsedBy) != 2 {
		t.Fatalf("used_by = %#v", detail.UsedBy)
	}
	if detail.UsedBy[0].Name != "svc-a" || detail.UsedBy[0].Version != "v1.10" {
		t.Fatalf("used_by order/version = %#v", detail.UsedBy)
	}

	if _, ok := BuildModelDetail(list, mvs, profiles, "qwen-v1-9"); !ok {
		t.Fatal("model not found by metadata.name")
	}
	if _, ok := BuildModelDetail(list, mvs, profiles, "missing"); ok {
		t.Fatal("missing model was found")
	}
}

func TestBuildModelDetailDeduplicatesSharedVersionNames(t *testing.T) {
	base := time.Date(2026, 10, 1, 8, 0, 0, 0, time.UTC)
	a := testModelVersion("dup-a", "dup-model", "fp16", "qwen3_5", "fp16", base, nil)
	b := testModelVersion("dup-b", "dup-model", "fp16", "qwen3_5", "fp16", base.Add(time.Hour), nil)
	mvs := []kube.ModelVersion{a, b}
	profiles := []kube.RuntimeProfile{testProfile("p", "qwen3_5", "fp16")}

	list := BuildCatalogList(mvs, profiles, nil, base.Add(2*time.Hour))
	detail, ok := BuildModelDetail(list, mvs, profiles, "dup-model")
	if !ok {
		t.Fatal("model not found")
	}
	if got := detail.CompatibleProfiles[0].CompatibleVersions; len(got) != 1 || got[0] != "fp16" {
		t.Fatalf("compatible versions = %#v", got)
	}
}

func TestCompatibilityRequiresArchitectureAndQuantization(t *testing.T) {
	mv := testModelVersion("m-v1", "m", "v1", "llama", "w8a8", time.Now(), nil)

	cases := []struct {
		name string
		spec kube.ModelRequirements
		want bool
	}{
		{"both match", kube.ModelRequirements{Architectures: []string{"llama"}, Quantization: []string{"w8a8"}}, true},
		{"architecture only", kube.ModelRequirements{Architectures: []string{"llama"}}, false},
		{"quantization only", kube.ModelRequirements{Quantization: []string{"w8a8"}}, false},
		{"empty is not a wildcard", kube.ModelRequirements{}, false},
		{"case sensitive", kube.ModelRequirements{Architectures: []string{"Llama"}, Quantization: []string{"w8a8"}}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			profile := testProfile("p", "", "")
			profile.Spec.ModelRequirements = tc.spec
			if got := compatible(profile, mv); got != tc.want {
				t.Fatalf("compatible = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestFallbacksWithoutAnnotations(t *testing.T) {
	mv := testModelVersion("llama-v1", "llama-3-8b", "v1", "llama", "bf16", time.Now(), nil)
	mv.Spec.Storage = kube.ModelStorage{Strategy: "Dynamic", Dynamic: &kube.DynamicStorage{StorageClassName: "ceph", Capacity: "10Gi"}}
	list := BuildCatalogList([]kube.ModelVersion{mv}, nil, nil, time.Now())
	m := list.Models[0]

	if m.DisplayName != "llama-3-8b" || m.Publisher != "" || m.License != "" {
		t.Fatalf("fallbacks = %#v", m)
	}
	if m.PublisherInitials != "LL" {
		t.Fatalf("initials = %q", m.PublisherInitials)
	}
	if len(m.Tags) != 3 || m.Tags[0] != "llama" || m.Tags[1] != "bf16" || m.Tags[2] != "Dynamic" {
		t.Fatalf("derived tags = %#v", m.Tags)
	}
	if m.Featured {
		t.Fatal("featured should default to false")
	}

	detail, _ := BuildModelDetail(list, []kube.ModelVersion{mv}, nil, "llama-3-8b")
	if detail.Versions[0].Storage.StorageClass != "ceph" || detail.Versions[0].Storage.Capacity != "10Gi" {
		t.Fatalf("storage = %#v", detail.Versions[0].Storage)
	}
}

func TestConditionsAreTriState(t *testing.T) {
	conds := []kube.Condition{
		{Type: conditionStorageResolved, Status: "True"},
		{Type: conditionInUse, Status: "False"},
		{Type: conditionAssetsResolved, Status: "Unknown"},
	}
	if got := conditionState(conds, conditionStorageResolved); got == nil || !*got {
		t.Fatalf("True = %v", got)
	}
	if got := conditionState(conds, conditionInUse); got == nil || *got {
		t.Fatalf("False = %v", got)
	}
	if got := conditionState(conds, conditionAssetsResolved); got != nil {
		t.Fatalf("Unknown = %v", got)
	}
	if got := conditionState(conds, "Missing"); got != nil {
		t.Fatalf("missing = %v", got)
	}
}

func TestProfileRoleAcceleratorAndGPUMax(t *testing.T) {
	profile := testProfile("multi", "llama", "bf16")
	profile.Metadata.Labels[labelServingMode] = "invalid-mode"
	profile.Spec.Roles = []kube.Role{
		{
			Name: "prefill",
			PodTemplate: kube.PodTemplate{
				Resources: &kube.PodResources{GPUPerPod: &kube.IntOrString{Int: intPtr(1)}},
			},
		},
		{
			Name:        "decode",
			Accelerator: &kube.Accelerator{Vendor: "metax", Models: []string{"C500"}},
			PodTemplate: kube.PodTemplate{
				Resources: &kube.PodResources{GPUPerPod: &kube.IntOrString{String: "{{ overrides.gpu }}"}},
			},
		},
	}
	out := profileSummary(profile, nil)
	if out.ServingMode != "" {
		t.Fatalf("invalid serving mode accepted: %q", out.ServingMode)
	}
	if out.GPUPerPod == nil || *out.GPUPerPod != 1 {
		t.Fatalf("gpu_per_pod = %v", out.GPUPerPod)
	}
	if len(out.Roles) != 2 || out.Roles[1].Vendor != "metax" || out.Roles[1].GPUPerPod != nil {
		t.Fatalf("roles = %#v", out.Roles)
	}
}

func TestCompareVersionsNaturalOrder(t *testing.T) {
	cases := []struct {
		a, b string
		want int
	}{
		{"v1.10", "v1.9", 1},
		{"v1.9", "v1.10", -1},
		{"v2.0", "v2.0", 0},
		{"2026-10-01", "2026-09-30", 1},
		{"v1.0", "v1.0.1", -1},
	}
	for _, tc := range cases {
		got := compareVersions(tc.a, tc.b)
		if (got > 0) != (tc.want > 0) || (got < 0) != (tc.want < 0) {
			t.Fatalf("compareVersions(%q,%q) = %d, want sign %d", tc.a, tc.b, got, tc.want)
		}
	}
}

func TestUnavailableListShape(t *testing.T) {
	list := BuildCatalogList(nil, nil, nil, time.Now())
	if !list.Available || len(list.Models) != 0 || list.Models == nil {
		t.Fatalf("empty list = %#v", list)
	}
}
