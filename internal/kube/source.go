package kube

import "context"

// UnavailableSource fails every list with a fixed classified reason. It is used
// when the feature is disabled or no cluster is configured, so the service can
// use one code path and still return a well-formed unavailable response.
type UnavailableSource struct {
	Reason    string
	Retryable bool
}

func (s UnavailableSource) ListModelVersions(context.Context) ([]ModelVersion, error) {
	return nil, &StatusError{Reason: s.Reason, Retryable: s.Retryable}
}

func (s UnavailableSource) ListRuntimeProfiles(context.Context) ([]RuntimeProfile, error) {
	return nil, &StatusError{Reason: s.Reason, Retryable: s.Retryable}
}

// DisabledSource reports MODEL_LIBRARY_ENABLED=false.
var DisabledSource Source = UnavailableSource{Reason: "DISABLED", Retryable: false}

// NoClusterSource reports that neither an in-cluster service account nor
// MODEL_LIBRARY_API_URL was available.
var NoClusterSource Source = UnavailableSource{Reason: "NO_CLUSTER", Retryable: true}
