package service

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/suanova/medrouter/internal/config"
	"github.com/suanova/medrouter/internal/kube"
)

// datasetSnapshot is an immutable view of one refresh attempt.
type datasetSnapshot struct {
	list      *DatasetList
	datasets  []kube.Dataset
	fetchedAt time.Time
	expiresAt time.Time
	available bool
}

// DatasetCatalog caches raw Dataset CRs and projects them per request. It does
// not share state with ModelCatalog: a missing datasets CRD must not make the
// model library unavailable, and vice versa.
type DatasetCatalog struct {
	source  kube.Source
	ttl     time.Duration
	timeout time.Duration
	now     func() time.Time

	mu       sync.Mutex
	current  *datasetSnapshot
	lastGood *datasetSnapshot

	refreshMu sync.Mutex
	lastLog   string
}

func NewDatasetCatalog(cfg *config.Config, source kube.Source) *DatasetCatalog {
	ttl := cfg.DatasetLibraryCacheTTL
	if ttl < minCatalogTTL {
		ttl = defaultCatalogTTL
	}
	timeout := cfg.DatasetLibraryTimeout
	if timeout <= 0 {
		timeout = defaultCatalogTimeout
	}
	return &DatasetCatalog{
		source:  source,
		ttl:     ttl,
		timeout: timeout,
		now:     time.Now,
	}
}

// List returns the projected catalog, refreshing it when the TTL expires.
func (c *DatasetCatalog) List(ctx context.Context) *DatasetList {
	return c.snapshot(ctx).list
}

// Detail returns the expanded dataset and whether ref exists. When the catalog
// is unavailable the returned detail carries the availability status and found
// is false.
func (c *DatasetCatalog) Detail(ctx context.Context, ref string) (*DatasetDetail, bool) {
	snap := c.snapshot(ctx)
	if !snap.available {
		return &DatasetDetail{DatasetCatalogStatus: snap.list.DatasetCatalogStatus}, false
	}
	detail, ok := BuildDatasetDetail(snap.list, snap.datasets, ref)
	if !ok {
		return nil, false
	}
	return detail, true
}

func (c *DatasetCatalog) snapshot(ctx context.Context) *datasetSnapshot {
	now := c.now()
	c.mu.Lock()
	current := c.current
	c.mu.Unlock()
	if current != nil && now.Before(current.expiresAt) {
		return current
	}

	// One refresh at a time; readers that missed the double-check wait for the
	// result instead of stampeding the API.
	c.refreshMu.Lock()
	defer c.refreshMu.Unlock()

	now = c.now()
	c.mu.Lock()
	current = c.current
	c.mu.Unlock()
	if current != nil && now.Before(current.expiresAt) {
		return current
	}

	fresh := c.fetch(ctx)
	c.mu.Lock()
	defer c.mu.Unlock()

	if fresh.available {
		c.current = fresh
		c.lastGood = fresh
		c.logState("", "")
		return fresh
	}

	reason := fresh.list.Reason
	if c.lastGood != nil && now.Sub(c.lastGood.fetchedAt) <= c.staleWindow() {
		stale := *c.lastGood
		stale.list = copyDatasetList(c.lastGood.list)
		stale.list.Stale = true
		stale.list.Warning = datasetCatalogMessage(reason)
		stale.list.GeneratedAt = c.lastGood.list.GeneratedAt
		stale.available = true
		stale.expiresAt = now.Add(c.negativeTTL(reason))
		c.current = &stale
		c.logState(reason, datasetCatalogMessage(reason))
		return &stale
	}

	c.current = fresh
	c.logState(reason, datasetCatalogMessage(reason))
	return fresh
}

func (c *DatasetCatalog) fetch(ctx context.Context) *datasetSnapshot {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	datasets, err := c.source.ListDatasets(ctx)

	now := c.now()
	if err != nil {
		reason, retryable := statusFromError(err)
		return &datasetSnapshot{
			list: &DatasetList{
				DatasetCatalogStatus: DatasetCatalogStatus{
					Available:   false,
					Reason:      reason,
					Message:     datasetCatalogMessage(reason),
					Retryable:   retryable,
					GeneratedAt: now.UTC(),
				},
				Datasets: []DatasetSummary{},
			},
			fetchedAt: now,
			expiresAt: now.Add(c.negativeTTL(reason)),
		}
	}

	return &datasetSnapshot{
		list:      BuildDatasetList(datasets, now),
		datasets:  datasets,
		fetchedAt: now,
		expiresAt: now.Add(c.ttl),
		available: true,
	}
}

func (c *DatasetCatalog) staleWindow() time.Duration {
	window := 5 * c.ttl
	if window < time.Minute {
		window = time.Minute
	}
	return window
}

func (c *DatasetCatalog) negativeTTL(reason string) time.Duration {
	if reason == CatalogCRDMissing {
		return crdMissingCatalogTTL
	}
	if c.ttl < negativeCatalogTTL {
		return c.ttl
	}
	return negativeCatalogTTL
}

func (c *DatasetCatalog) logState(reason, message string) {
	key := reason + "|" + message
	if key == c.lastLog {
		return
	}
	c.lastLog = key
	if reason == "" {
		log.Printf("dataset catalog: available")
		return
	}
	log.Printf("WARN dataset catalog unavailable reason=%s message=%q", reason, message)
}

func copyDatasetList(in *DatasetList) *DatasetList {
	if in == nil {
		return &DatasetList{Datasets: []DatasetSummary{}}
	}
	out := *in
	out.Datasets = append([]DatasetSummary(nil), in.Datasets...)
	return &out
}
