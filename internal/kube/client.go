package kube

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	serviceAccountDir   = "/var/run/secrets/kubernetes.io/serviceaccount"
	serviceAccountToken = serviceAccountDir + "/token"
	serviceAccountCA    = serviceAccountDir + "/ca.crt"

	modelVersionPath   = "/apis/ai.cubestack.io/v1alpha1/modelversions"
	runtimeProfilePath = "/apis/ai.cubestack.io/v1alpha1/inferenceruntimeprofiles"

	pageLimit   = 500
	maxPages    = 20
	maxBodySize = 8 << 20
)

// ErrNoCluster means no in-cluster service account and no configured API URL.
var ErrNoCluster = errors.New("no Kubernetes API is available")

// StatusError is a classified API failure. Reason is one of the stable codes
// the frontend translates; Error() stays safe to log and never contains a
// response body or credentials.
type StatusError struct {
	Status    int
	Reason    string
	Retryable bool
	detail    string
}

func (e *StatusError) Error() string {
	if e.detail == "" {
		return fmt.Sprintf("kubernetes API: %s", e.Reason)
	}
	return fmt.Sprintf("kubernetes API: %s: %s", e.Reason, e.detail)
}

// Source is the read-only view of the CubeStack CRDs used by the catalog.
type Source interface {
	ListModelVersions(ctx context.Context) ([]ModelVersion, error)
	ListRuntimeProfiles(ctx context.Context) ([]RuntimeProfile, error)
}

// Client lists cluster-scoped CubeStack custom resources over the Kubernetes
// REST API using only the standard library.
type Client struct {
	baseURL    string
	httpClient *http.Client
	token      string
	tokenFile  string
}

// NewInClusterClient builds a client from the pod's service account. It
// returns ErrNoCluster when the process is not running in a pod.
func NewInClusterClient(timeout time.Duration) (*Client, error) {
	host := strings.TrimSpace(os.Getenv("KUBERNETES_SERVICE_HOST"))
	port := strings.TrimSpace(os.Getenv("KUBERNETES_SERVICE_PORT"))
	if host == "" || port == "" {
		return nil, ErrNoCluster
	}
	if _, err := os.Stat(serviceAccountToken); err != nil {
		return nil, ErrNoCluster
	}

	caPEM, err := os.ReadFile(serviceAccountCA)
	if err != nil {
		return nil, fmt.Errorf("read service account CA: %w", err)
	}
	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(caPEM) {
		return nil, fmt.Errorf("service account CA is not valid PEM")
	}

	client := &Client{
		baseURL:   "https://" + net.JoinHostPort(host, port),
		tokenFile: serviceAccountToken,
	}
	client.httpClient = &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs:    pool,
				MinVersion: tls.VersionTLS12,
				ServerName: host,
			},
			Proxy: http.ProxyFromEnvironment,
		},
	}
	return client, nil
}

// NewAPIClient builds a client for an explicit API base URL, for example
// http://127.0.0.1:8001 from `kubectl proxy`. No credentials are attached; a
// proxy is expected to handle authentication.
func NewAPIClient(rawURL string, timeout time.Duration) (*Client, error) {
	parsed, err := url.Parse(strings.TrimRight(strings.TrimSpace(rawURL), "/"))
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return nil, fmt.Errorf("MODEL_LIBRARY_API_URL must be an absolute URL")
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return nil, fmt.Errorf("MODEL_LIBRARY_API_URL must use http or https")
	}
	return &Client{
		baseURL: parsed.String(),
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}, nil
}

func (c *Client) ListModelVersions(ctx context.Context) ([]ModelVersion, error) {
	return listAll[ModelVersion](ctx, c, modelVersionPath)
}

func (c *Client) ListRuntimeProfiles(ctx context.Context) ([]RuntimeProfile, error) {
	return listAll[RuntimeProfile](ctx, c, runtimeProfilePath)
}

func listAll[T any](ctx context.Context, c *Client, path string) ([]T, error) {
	var out []T
	cont := ""
	for page := 0; page < maxPages; page++ {
		query := url.Values{"limit": {fmt.Sprintf("%d", pageLimit)}}
		if cont != "" {
			query.Set("continue", cont)
		}
		var body rawList[T]
		if err := c.getJSON(ctx, path+"?"+query.Encode(), &body); err != nil {
			return nil, err
		}
		out = append(out, body.Items...)
		if body.Metadata.Continue == "" {
			return out, nil
		}
		if body.Metadata.Continue == cont {
			return nil, &StatusError{Status: 0, Reason: "API_ERROR", Retryable: true, detail: "pagination did not advance"}
		}
		cont = body.Metadata.Continue
	}
	return nil, &StatusError{Status: 0, Reason: "API_ERROR", Retryable: true, detail: "too many pages"}
}

func (c *Client) getJSON(ctx context.Context, path string, into any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return &StatusError{Status: 0, Reason: "API_ERROR", Retryable: false, detail: err.Error()}
	}
	req.Header.Set("Accept", "application/json")
	if token, err := c.bearerToken(); err != nil {
		return err
	} else if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return &StatusError{Status: 0, Reason: "TIMEOUT", Retryable: true}
		}
		if errors.Is(ctx.Err(), context.Canceled) {
			return &StatusError{Status: 0, Reason: "TIMEOUT", Retryable: true}
		}
		return &StatusError{Status: 0, Reason: "API_ERROR", Retryable: true, detail: "request failed"}
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 4<<10))
		return classifyStatus(resp.StatusCode)
	}

	decoder := json.NewDecoder(io.LimitReader(resp.Body, maxBodySize))
	if err := decoder.Decode(into); err != nil {
		return &StatusError{Status: resp.StatusCode, Reason: "API_ERROR", Retryable: true, detail: "invalid response"}
	}
	return nil
}

func classifyStatus(status int) *StatusError {
	switch status {
	case http.StatusNotFound:
		return &StatusError{Status: status, Reason: "CRD_MISSING", Retryable: true}
	case http.StatusForbidden:
		return &StatusError{Status: status, Reason: "FORBIDDEN", Retryable: true}
	case http.StatusUnauthorized:
		return &StatusError{Status: status, Reason: "UNAUTHORIZED", Retryable: true}
	default:
		return &StatusError{Status: status, Reason: "API_ERROR", Retryable: status >= 500}
	}
}

func (c *Client) bearerToken() (string, error) {
	if c.token != "" {
		return c.token, nil
	}
	if c.tokenFile == "" {
		return "", nil
	}
	data, err := os.ReadFile(filepath.Clean(c.tokenFile))
	if err != nil {
		return "", &StatusError{Status: 0, Reason: "UNAUTHORIZED", Retryable: true, detail: "service account token is unreadable"}
	}
	return strings.TrimSpace(string(data)), nil
}
