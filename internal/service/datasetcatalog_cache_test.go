package service

import (
	"context"
	"testing"
	"time"

	"github.com/suanova/medrouter/internal/config"
	"github.com/suanova/medrouter/internal/kube"
)

func newTestDatasetCatalog(source kube.Source, clock *time.Time) *DatasetCatalog {
	cfg := &config.Config{
		DatasetLibraryCacheTTL: 30 * time.Second,
		DatasetLibraryTimeout:  2 * time.Second,
	}
	catalog := NewDatasetCatalog(cfg, source)
	catalog.now = func() time.Time { return *clock }
	return catalog
}

func TestDatasetCatalogCachesWithinTTL(t *testing.T) {
	now := time.Date(2026, 11, 10, 9, 0, 0, 0, time.UTC)
	source := &fakeSource{datasets: []kube.Dataset{testDataset("ds-v1", "ds", "v1", "Parquet", now, nil)}}
	catalog := newTestDatasetCatalog(source, &now)

	first := catalog.List(context.Background())
	second := catalog.List(context.Background())
	if !first.Available || !second.Available {
		t.Fatalf("list unavailable: %#v", first)
	}
	if source.dsCalls != 1 {
		t.Fatalf("calls within TTL = %d, want 1", source.dsCalls)
	}

	now = now.Add(31 * time.Second)
	catalog.List(context.Background())
	if source.dsCalls != 2 {
		t.Fatalf("calls after TTL = %d, want 2", source.dsCalls)
	}
}

func TestDatasetCatalogNegativeCache(t *testing.T) {
	now := time.Date(2026, 11, 10, 9, 0, 0, 0, time.UTC)
	source := &fakeSource{dsErr: &kube.StatusError{Reason: "FORBIDDEN", Retryable: true}}
	catalog := newTestDatasetCatalog(source, &now)

	list := catalog.List(context.Background())
	if list.Available || list.Reason != "FORBIDDEN" || !list.Retryable || list.Datasets == nil {
		t.Fatalf("list = %#v", list)
	}
	catalog.List(context.Background())
	if source.dsCalls != 1 {
		t.Fatalf("calls in negative cache = %d, want 1", source.dsCalls)
	}

	now = now.Add(11 * time.Second)
	catalog.List(context.Background())
	if source.dsCalls != 2 {
		t.Fatalf("calls after negative TTL = %d, want 2", source.dsCalls)
	}
}

func TestDatasetCatalogCRDMissingUsesLongNegativeTTL(t *testing.T) {
	now := time.Date(2026, 11, 10, 9, 0, 0, 0, time.UTC)
	source := &fakeSource{dsErr: &kube.StatusError{Reason: "CRD_MISSING", Retryable: true}}
	catalog := newTestDatasetCatalog(source, &now)

	catalog.List(context.Background())
	now = now.Add(31 * time.Second)
	catalog.List(context.Background())
	if source.dsCalls != 1 {
		t.Fatalf("calls after regular TTL = %d, want 1", source.dsCalls)
	}
	now = now.Add(30 * time.Second)
	catalog.List(context.Background())
	if source.dsCalls != 2 {
		t.Fatalf("calls after CRD-missing TTL = %d, want 2", source.dsCalls)
	}
}

func TestDatasetCatalogServesStaleLastGoodData(t *testing.T) {
	now := time.Date(2026, 11, 10, 9, 0, 0, 0, time.UTC)
	source := &fakeSource{datasets: []kube.Dataset{testDataset("ds-v1", "ds", "v1", "Parquet", now, nil)}}
	catalog := newTestDatasetCatalog(source, &now)

	if list := catalog.List(context.Background()); !list.Available || list.Total != 1 {
		t.Fatalf("initial list = %#v", list)
	}

	now = now.Add(40 * time.Second) // past TTL, inside the stale window
	source.mu.Lock()
	source.dsErr = &kube.StatusError{Reason: "TIMEOUT", Retryable: true}
	source.mu.Unlock()

	stale := catalog.List(context.Background())
	if !stale.Available || !stale.Stale || stale.Total != 1 || stale.Warning == "" {
		t.Fatalf("stale list = %#v", stale)
	}
	if source.dsCalls != 2 {
		t.Fatalf("calls = %d, want 2", source.dsCalls)
	}

	catalog.List(context.Background())
	if source.dsCalls != 2 {
		t.Fatalf("calls after stale cache = %d, want 2", source.dsCalls)
	}

	now = now.Add(11 * time.Second)
	catalog.List(context.Background())
	if source.dsCalls != 3 {
		t.Fatalf("calls after stale cache expiry = %d, want 3", source.dsCalls)
	}
}

func TestDatasetCatalogDisabled(t *testing.T) {
	now := time.Now()
	catalog := newTestDatasetCatalog(kube.UnavailableSource{Reason: CatalogDisabled}, &now)
	list := catalog.List(context.Background())
	if list.Available || list.Reason != "DISABLED" || list.Retryable {
		t.Fatalf("list = %#v", list)
	}

	detail, found := catalog.Detail(context.Background(), "anything")
	if found || detail == nil || detail.Available || detail.Reason != "DISABLED" {
		t.Fatalf("detail = %#v found=%v", detail, found)
	}
}

func TestDatasetCatalogDetail(t *testing.T) {
	now := time.Date(2026, 11, 10, 9, 0, 0, 0, time.UTC)
	datasets, _ := testDatasetCatalog()
	source := &fakeSource{datasets: datasets}
	catalog := newTestDatasetCatalog(source, &now)

	detail, found := catalog.Detail(context.Background(), "demo-dataset")
	if !found || !detail.Available || detail.LatestVersion != "v1.1.0" || len(detail.Versions) != 2 {
		t.Fatalf("detail = %#v found=%v", detail, found)
	}

	missing, found := catalog.Detail(context.Background(), "missing")
	if found || missing != nil {
		t.Fatalf("missing detail = %#v found=%v", missing, found)
	}
}
