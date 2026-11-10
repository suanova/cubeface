package service

import "time"

// DatasetCatalogStatus is the availability envelope shared by the dataset list
// and detail responses. It intentionally has no profiles_available field: the
// dataset catalog does not depend on inference runtimes.
type DatasetCatalogStatus struct {
	Available   bool      `json:"available"`
	Reason      string    `json:"reason,omitempty"`
	Message     string    `json:"message,omitempty"`
	Retryable   bool      `json:"retryable"`
	Stale       bool      `json:"stale"`
	Warning     string    `json:"warning,omitempty"`
	GeneratedAt time.Time `json:"generated_at"`
}

// DatasetSummary is one dataset group (all versions sharing spec.dataset).
type DatasetSummary struct {
	Dataset           string    `json:"dataset"`
	DisplayName       string    `json:"display_name"`
	Publisher         string    `json:"publisher"`
	PublisherInitials string    `json:"publisher_initials"`
	Description       string    `json:"description"`
	Format            string    `json:"format"`
	Task              string    `json:"task"`
	License           string    `json:"license"`
	Tags              []string  `json:"tags"`
	Access            string    `json:"access"`
	Deidentification  string    `json:"deidentification"`
	Ethics            string    `json:"ethics"`
	SourceRepo        string    `json:"source_repo"`
	UsageBoundary     string    `json:"usage_boundary"`
	Size              string    `json:"size,omitempty"`
	RecordCount       *int64    `json:"record_count,omitempty"`
	StorageStrategy   string    `json:"storage_strategy"`
	LatestVersion     string    `json:"latest_version"`
	VersionCount      int       `json:"version_count"`
	Featured          bool      `json:"featured"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// DatasetList is the payload of GET /api/datasets.
type DatasetList struct {
	DatasetCatalogStatus
	Total        int              `json:"total"`
	VersionTotal int              `json:"version_total"`
	Datasets     []DatasetSummary `json:"datasets"`
}

type DatasetVersionDetail struct {
	Ref         string         `json:"ref"`
	Version     string         `json:"version"`
	IsLatest    bool           `json:"is_latest"`
	Format      string         `json:"format"`
	Size        string         `json:"size,omitempty"`
	RecordCount *int64         `json:"record_count,omitempty"`
	Description string         `json:"description"`
	Storage     StorageSummary `json:"storage"`
	CreatedAt   time.Time      `json:"created_at"`
}

// DatasetDetail is the payload of GET /api/datasets/:name.
type DatasetDetail struct {
	DatasetCatalogStatus
	DatasetSummary
	LatestRef string                 `json:"latest_ref"`
	Versions  []DatasetVersionDetail `json:"versions"`
}
