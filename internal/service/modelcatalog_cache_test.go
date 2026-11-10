package service

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/cubeface/cubeface/internal/config"
	"github.com/cubeface/cubeface/internal/kube"
)

type fakeSource struct {
	mu        sync.Mutex
	mvs       []kube.ModelVersion
	profiles  []kube.RuntimeProfile
	datasets  []kube.Dataset
	mvErr     error
	profErr   error
	dsErr     error
	mvCalls   int
	profCalls int
	dsCalls   int
}

func (f *fakeSource) ListModelVersions(context.Context) ([]kube.ModelVersion, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.mvCalls++
	return f.mvs, f.mvErr
}

func (f *fakeSource) ListRuntimeProfiles(context.Context) ([]kube.RuntimeProfile, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.profCalls++
	return f.profiles, f.profErr
}

func (f *fakeSource) ListDatasets(context.Context) ([]kube.Dataset, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.dsCalls++
	return f.datasets, f.dsErr
}

func (f *fakeSource) calls() (int, int) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.mvCalls, f.profCalls
}

func newTestCatalog(source kube.Source, clock *time.Time) *ModelCatalog {
	cfg := &config.Config{
		ModelLibraryCacheTTL: 30 * time.Second,
		ModelLibraryTimeout:  2 * time.Second,
	}
	catalog := NewModelCatalog(cfg, source)
	catalog.now = func() time.Time { return *clock }
	return catalog
}

func TestCatalogCachesWithinTTL(t *testing.T) {
	now := time.Date(2026, 11, 10, 9, 0, 0, 0, time.UTC)
	source := &fakeSource{mvs: []kube.ModelVersion{testModelVersion("m-v1", "m", "v1", "a", "b", now, nil)}}
	catalog := newTestCatalog(source, &now)

	first := catalog.List(context.Background())
	second := catalog.List(context.Background())
	if !first.Available || !second.Available {
		t.Fatalf("list unavailable: %#v", first)
	}
	if mv, _ := source.calls(); mv != 1 {
		t.Fatalf("calls within TTL = %d, want 1", mv)
	}

	now = now.Add(31 * time.Second)
	catalog.List(context.Background())
	if mv, _ := source.calls(); mv != 2 {
		t.Fatalf("calls after TTL = %d, want 2", mv)
	}
}

func TestCatalogNegativeCache(t *testing.T) {
	now := time.Date(2026, 11, 10, 9, 0, 0, 0, time.UTC)
	source := &fakeSource{mvErr: &kube.StatusError{Reason: "FORBIDDEN", Retryable: true}}
	catalog := newTestCatalog(source, &now)

	list := catalog.List(context.Background())
	if list.Available || list.Reason != "FORBIDDEN" || list.Retryable != true || list.Models == nil {
		t.Fatalf("list = %#v", list)
	}
	catalog.List(context.Background())
	if mv, _ := source.calls(); mv != 1 {
		t.Fatalf("calls in negative cache = %d, want 1", mv)
	}

	now = now.Add(11 * time.Second)
	catalog.List(context.Background())
	if mv, _ := source.calls(); mv != 2 {
		t.Fatalf("calls after negative TTL = %d, want 2", mv)
	}
}

func TestCatalogServesStaleLastGoodData(t *testing.T) {
	now := time.Date(2026, 11, 10, 9, 0, 0, 0, time.UTC)
	source := &fakeSource{mvs: []kube.ModelVersion{testModelVersion("m-v1", "m", "v1", "a", "b", now, nil)}}
	catalog := newTestCatalog(source, &now)

	if list := catalog.List(context.Background()); !list.Available || list.Total != 1 {
		t.Fatalf("initial list = %#v", list)
	}

	now = now.Add(40 * time.Second) // past TTL, inside the stale window
	source.mu.Lock()
	source.mvErr = &kube.StatusError{Reason: "TIMEOUT", Retryable: true}
	source.mu.Unlock()

	stale := catalog.List(context.Background())
	if !stale.Available || !stale.Stale || stale.Total != 1 || stale.Warning == "" {
		t.Fatalf("stale list = %#v", stale)
	}
	if mv, _ := source.calls(); mv != 2 {
		t.Fatalf("mv calls = %d, want 2", mv)
	}

	// The stale snapshot is itself cached briefly; no stampede on every request.
	catalog.List(context.Background())
	if mv, _ := source.calls(); mv != 2 {
		t.Fatalf("mv calls after stale cache = %d, want 2", mv)
	}

	now = now.Add(11 * time.Second)
	catalog.List(context.Background())
	if mv, _ := source.calls(); mv != 3 {
		t.Fatalf("mv calls after stale cache expiry = %d, want 3", mv)
	}
}

func TestCatalogProfileFailureKeepsModels(t *testing.T) {
	now := time.Date(2026, 11, 10, 9, 0, 0, 0, time.UTC)
	source := &fakeSource{
		mvs:     []kube.ModelVersion{testModelVersion("m-v1", "m", "v1", "a", "b", now, nil)},
		profErr: &kube.StatusError{Reason: "FORBIDDEN", Retryable: true},
	}
	catalog := newTestCatalog(source, &now)

	list := catalog.List(context.Background())
	if !list.Available || list.ProfilesAvailable || list.Total != 1 || list.Warning == "" {
		t.Fatalf("list = %#v", list)
	}
	if list.Models[0].CompatibleProfileCount != 0 {
		t.Fatalf("compatible count = %d", list.Models[0].CompatibleProfileCount)
	}
}

func TestCatalogDisabled(t *testing.T) {
	now := time.Now()
	catalog := newTestCatalog(kube.DisabledSource, &now)
	list := catalog.List(context.Background())
	if list.Available || list.Reason != "DISABLED" || list.Retryable {
		t.Fatalf("list = %#v", list)
	}

	detail, found := catalog.Detail(context.Background(), "anything")
	if found || detail == nil || detail.Available || detail.Reason != "DISABLED" {
		t.Fatalf("detail = %#v found=%v", detail, found)
	}
}
