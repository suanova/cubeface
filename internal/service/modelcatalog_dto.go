package service

import "time"

// Catalog availability reasons. The frontend translates these; Message is a
// safe English fallback that never contains response bodies or credentials.
const (
	CatalogAvailable    = ""
	CatalogDisabled     = "DISABLED"
	CatalogNoCluster    = "NO_CLUSTER"
	CatalogCRDMissing   = "CRD_MISSING"
	CatalogForbidden    = "FORBIDDEN"
	CatalogUnauthorized = "UNAUTHORIZED"
	CatalogTimeout      = "TIMEOUT"
	CatalogAPIError     = "API_ERROR"
)

// CatalogStatus is shared by the list and detail responses.
type CatalogStatus struct {
	Available         bool      `json:"available"`
	Reason            string    `json:"reason,omitempty"`
	Message           string    `json:"message,omitempty"`
	Retryable         bool      `json:"retryable"`
	Stale             bool      `json:"stale"`
	Warning           string    `json:"warning,omitempty"`
	ProfilesAvailable bool      `json:"profiles_available"`
	GeneratedAt       time.Time `json:"generated_at"`
}

// ModelSummary is one model group (all versions of spec.model).
type ModelSummary struct {
	Model                  string    `json:"model"`
	DisplayName            string    `json:"display_name"`
	Publisher              string    `json:"publisher"`
	PublisherInitials      string    `json:"publisher_initials"`
	Description            string    `json:"description"`
	Architecture           string    `json:"architecture"`
	Quantization           string    `json:"quantization"`
	Task                   string    `json:"task"`
	License                string    `json:"license"`
	Tags                   []string  `json:"tags"`
	UsageBoundary          string    `json:"usage_boundary"`
	StorageStrategy        string    `json:"storage_strategy"`
	LatestVersion          string    `json:"latest_version"`
	VersionCount           int       `json:"version_count"`
	CompatibleProfileCount int       `json:"compatible_profile_count"`
	UsedByCount            int       `json:"used_by_count"`
	Featured               bool      `json:"featured"`
	CreatedAt              time.Time `json:"created_at"`
	UpdatedAt              time.Time `json:"updated_at"`
}

// ModelList is the payload of GET /api/models.
type ModelList struct {
	CatalogStatus
	Total        int            `json:"total"`
	VersionTotal int            `json:"version_total"`
	ProfileTotal int            `json:"profile_total"`
	Models       []ModelSummary `json:"models"`
}

// StorageSummary deliberately omits host paths, S3 URIs and credential names.
type StorageSummary struct {
	Strategy              string `json:"strategy"`
	StorageClass          string `json:"storage_class,omitempty"`
	Capacity              string `json:"capacity,omitempty"`
	PathConfigured        bool   `json:"path_configured"`
	URIConfigured         bool   `json:"uri_configured"`
	CredentialsConfigured bool   `json:"credentials_configured"`
}

type UsedByRef struct {
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
	Version   string `json:"version,omitempty"`
}

type EngineSummary struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type ProfileRoleSummary struct {
	Name              string   `json:"name"`
	Vendor            string   `json:"vendor"`
	AcceleratorModels []string `json:"accelerator_models"`
	GPUPerPod         *int     `json:"gpu_per_pod"`
}

type RuntimeProfileSummary struct {
	Name               string               `json:"name"`
	Description        string               `json:"description"`
	Engine             EngineSummary        `json:"engine"`
	Vendor             string               `json:"vendor"`
	AcceleratorModels  []string             `json:"accelerator_models"`
	Architectures      []string             `json:"architectures"`
	Quantizations      []string             `json:"quantizations"`
	ServingMode        string               `json:"serving_mode,omitempty"`
	GPUPerPod          *int                 `json:"gpu_per_pod"`
	Roles              []ProfileRoleSummary `json:"roles"`
	CompatibleVersions []string             `json:"compatible_versions"`
	AssetsResolved     *bool                `json:"assets_resolved"`
	Deprecated         *bool                `json:"deprecated"`
	Drifted            *bool                `json:"drifted"`
	UsedByCount        int                  `json:"used_by_count"`
}

type ModelVersionDetail struct {
	Ref                string         `json:"ref"`
	Version            string         `json:"version"`
	IsLatest           bool           `json:"is_latest"`
	Architecture       string         `json:"architecture"`
	Quantization       string         `json:"quantization"`
	Description        string         `json:"description"`
	Storage            StorageSummary `json:"storage"`
	StorageResolved    *bool          `json:"storage_resolved"`
	InUse              *bool          `json:"in_use"`
	RootPathResolved   bool           `json:"root_path_resolved"`
	UsedBy             []UsedByRef    `json:"used_by"`
	CompatibleProfiles []string       `json:"compatible_profiles"`
	CreatedAt          time.Time      `json:"created_at"`
}

// ModelDetail is the payload of GET /api/models/:name.
type ModelDetail struct {
	CatalogStatus
	ModelSummary
	LatestRef          string                  `json:"latest_ref"`
	Versions           []ModelVersionDetail    `json:"versions"`
	CompatibleProfiles []RuntimeProfileSummary `json:"compatible_profiles"`
	UsedBy             []UsedByRef             `json:"used_by"`
}
