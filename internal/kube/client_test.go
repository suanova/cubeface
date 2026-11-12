package kube

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestListFollowsContinueAndSendsToken(t *testing.T) {
	var calls int32
	var authHeaders []string
	var queryValues []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != modelVersionPath {
			t.Errorf("path = %s", r.URL.Path)
		}
		authHeaders = append(authHeaders, r.Header.Get("Authorization"))
		queryValues = append(queryValues, r.URL.RawQuery)
		n := atomic.AddInt32(&calls, 1)
		if n == 1 {
			_, _ = w.Write([]byte(`{"metadata":{"continue":"next"},"items":[{"metadata":{"name":"a"},"spec":{"model":"m","version":"v1"}}]}`))
			return
		}
		_, _ = w.Write([]byte(`{"metadata":{},"items":[{"metadata":{"name":"b"},"spec":{"model":"m","version":"v2"}}]}`))
	}))
	defer server.Close()

	client := &Client{baseURL: server.URL, httpClient: server.Client(), token: "secret-token"}
	items, err := client.ListModelVersions(context.Background())
	if err != nil {
		t.Fatalf("ListModelVersions: %v", err)
	}
	if len(items) != 2 || items[0].Metadata.Name != "a" || items[1].Metadata.Name != "b" {
		t.Fatalf("items = %#v", items)
	}
	if calls != 2 {
		t.Fatalf("calls = %d, want 2", calls)
	}
	for _, got := range authHeaders {
		if got != "Bearer secret-token" {
			t.Fatalf("Authorization = %q", got)
		}
	}
	if !strings.Contains(queryValues[0], "limit=500") || strings.Contains(queryValues[0], "continue=") {
		t.Fatalf("first query = %q", queryValues[0])
	}
	if !strings.Contains(queryValues[1], "continue=next") {
		t.Fatalf("second query = %q", queryValues[1])
	}
}

func TestListClassifiesHTTPStatus(t *testing.T) {
	cases := []struct {
		status int
		reason string
	}{
		{http.StatusNotFound, "CRD_MISSING"},
		{http.StatusForbidden, "FORBIDDEN"},
		{http.StatusUnauthorized, "UNAUTHORIZED"},
		{http.StatusInternalServerError, "API_ERROR"},
	}
	for _, tc := range cases {
		t.Run(tc.reason, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(tc.status)
				_, _ = w.Write([]byte(`{"kind":"Status","message":"do not leak me"}`))
			}))
			defer server.Close()

			client, err := NewAPIClient(server.URL, time.Second)
			if err != nil {
				t.Fatal(err)
			}
			_, err = client.ListRuntimeProfiles(context.Background())
			var statusErr *StatusError
			if !errors.As(err, &statusErr) {
				t.Fatalf("err = %v, want StatusError", err)
			}
			if statusErr.Reason != tc.reason {
				t.Fatalf("reason = %s, want %s", statusErr.Reason, tc.reason)
			}
			if strings.Contains(err.Error(), "do not leak me") {
				t.Fatalf("error leaked response body: %v", err)
			}
		})
	}
}

func TestListTimesOut(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		_, _ = w.Write([]byte(`{"items":[]}`))
	}))
	defer server.Close()

	client, err := NewAPIClient(server.URL, 20*time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}
	_, err = client.ListModelVersions(context.Background())
	var statusErr *StatusError
	if !errors.As(err, &statusErr) || statusErr.Reason != "TIMEOUT" {
		t.Fatalf("err = %v, want TIMEOUT", err)
	}
}

func TestListDatasetsDecodesSpec(t *testing.T) {
	var gotPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_, _ = w.Write([]byte(`{"metadata":{},"items":[{
			"metadata":{"name":"demo-v1","creationTimestamp":"2026-11-01T08:00:00Z",
				"annotations":{"medrouter.io/display-name":"Demo"}},
			"spec":{"dataset":"demo","version":"v1","format":"Parquet",
				"size":"1.2Gi","recordCount":120000,
				"storage":{"strategy":"S3","s3":{"uri":"s3://bucket/demo","credentialsRef":{"name":"creds"}}}}
		}]}`))
	}))
	defer server.Close()

	client, err := NewAPIClient(server.URL, time.Second)
	if err != nil {
		t.Fatal(err)
	}
	items, err := client.ListDatasets(context.Background())
	if err != nil {
		t.Fatalf("ListDatasets: %v", err)
	}
	if gotPath != datasetPath {
		t.Fatalf("path = %s, want %s", gotPath, datasetPath)
	}
	if len(items) != 1 {
		t.Fatalf("items = %#v", items)
	}
	ds := items[0]
	if ds.Spec.Dataset != "demo" || ds.Spec.Version != "v1" || ds.Spec.Format != "Parquet" {
		t.Fatalf("spec = %#v", ds.Spec)
	}
	// resource.Quantity decodes as its canonical string; 1.2Gi stays 1288490188800
	// bytes when serialized by the apiserver, but any non-empty size is enough
	// for the catalog to display.
	if ds.Spec.Size == "" || ds.Spec.RecordCount == nil || *ds.Spec.RecordCount != 120000 {
		t.Fatalf("size/record count = %q/%v", ds.Spec.Size, ds.Spec.RecordCount)
	}
	if ds.Spec.Storage.Strategy != "S3" || ds.Spec.Storage.S3 == nil || ds.Spec.Storage.S3.URI == "" {
		t.Fatalf("storage = %#v", ds.Spec.Storage)
	}
}

func TestNewInClusterClientWithoutEnvironment(t *testing.T) {
	t.Setenv("KUBERNETES_SERVICE_HOST", "")
	t.Setenv("KUBERNETES_SERVICE_PORT", "")
	_, err := NewInClusterClient(time.Second)
	if !errors.Is(err, ErrNoCluster) {
		t.Fatalf("err = %v, want ErrNoCluster", err)
	}
}

func TestNewAPIClientRejectsBadURL(t *testing.T) {
	for _, raw := range []string{"", "127.0.0.1:8001", "ftp://example.com"} {
		if _, err := NewAPIClient(raw, time.Second); err == nil {
			t.Fatalf("NewAPIClient(%q) accepted an invalid URL", raw)
		}
	}
}
