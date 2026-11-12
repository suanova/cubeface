package kube

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"
)

// The types in this file decode only the fields MedRouter needs from the
// CubeStack CRDs. They are deliberately not the operator's generated types:
// the model library must not depend on k8s.io/apimachinery.

type ObjectMeta struct {
	Name              string            `json:"name"`
	Annotations       map[string]string `json:"annotations"`
	Labels            map[string]string `json:"labels"`
	CreationTimestamp time.Time         `json:"creationTimestamp"`
}

type ObjectRef struct {
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
}

type Condition struct {
	Type               string    `json:"type"`
	Status             string    `json:"status"`
	Reason             string    `json:"reason"`
	Message            string    `json:"message"`
	LastTransitionTime time.Time `json:"lastTransitionTime"`
}

type ModelVersion struct {
	Metadata ObjectMeta         `json:"metadata"`
	Spec     ModelVersionSpec   `json:"spec"`
	Status   ModelVersionStatus `json:"status"`
}

type ModelVersionSpec struct {
	Model        string            `json:"model"`
	Version      string            `json:"version"`
	Architecture string            `json:"architecture"`
	Quantization string            `json:"quantization"`
	Meta         map[string]string `json:"meta"`
	Storage      ModelStorage      `json:"storage"`
}

type ModelStorage struct {
	Strategy string           `json:"strategy"`
	HostPath *HostPathStorage `json:"hostPath,omitempty"`
	Dynamic  *DynamicStorage  `json:"dynamic,omitempty"`
	Static   *StaticStorage   `json:"static,omitempty"`
	S3       *S3Storage       `json:"s3,omitempty"`
}

type HostPathStorage struct {
	// Path is deliberately decoded so callers can report whether it is
	// configured, but it is never included in an API response.
	Path string `json:"path"`
}

type DynamicStorage struct {
	StorageClassName string `json:"storageClassName"`
	SubPath          string `json:"subPath"`
	Capacity         string `json:"capacity"`
}

type StaticStorage struct {
	StorageClassName string `json:"storageClassName"`
	Capacity         string `json:"capacity"`
}

type S3Storage struct {
	// URI and CredentialsRef are decoded only to report configured flags.
	URI            string            `json:"uri,omitempty"`
	CredentialsRef *S3CredentialsRef `json:"credentialsRef,omitempty"`
}

type S3CredentialsRef struct {
	Name string `json:"name"`
}

type ModelVersionStatus struct {
	UsedBy     []ObjectRef `json:"usedBy,omitempty"`
	RootPath   string      `json:"rootPath,omitempty"`
	Conditions []Condition `json:"conditions,omitempty"`
}

// Dataset decodes a CubeStack Dataset. The CRD is a registry entry and has no
// status subresource, so only metadata and spec are read.
type Dataset struct {
	Metadata ObjectMeta  `json:"metadata"`
	Spec     DatasetSpec `json:"spec"`
}

type DatasetSpec struct {
	Dataset     string            `json:"dataset"`
	Version     string            `json:"version"`
	Format      string            `json:"format"`
	Size        string            `json:"size,omitempty"`
	RecordCount *int64            `json:"recordCount,omitempty"`
	Meta        map[string]string `json:"meta,omitempty"`
	Storage     ModelStorage      `json:"storage"`
}

type RuntimeProfile struct {
	Metadata ObjectMeta         `json:"metadata"`
	Spec     RuntimeProfileSpec `json:"spec"`
	Status   ProfileStatus      `json:"status"`
}

type RuntimeProfileSpec struct {
	Accelerator       Accelerator       `json:"accelerator"`
	Engine            Engine            `json:"engine"`
	ModelRequirements ModelRequirements `json:"modelRequirements"`
	Roles             []Role            `json:"roles"`
}

type Accelerator struct {
	Vendor string   `json:"vendor"`
	Models []string `json:"models"`
}

type Engine struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type ModelRequirements struct {
	Architectures []string `json:"architectures"`
	Quantization  []string `json:"quantization"`
}

type Role struct {
	Name        string       `json:"name"`
	Accelerator *Accelerator `json:"accelerator,omitempty"`
	PodTemplate PodTemplate  `json:"podTemplate"`
}

type PodTemplate struct {
	Resources *PodResources `json:"resources,omitempty"`
}

type PodResources struct {
	GPUPerPod *IntOrString `json:"gpuPerPod,omitempty"`
}

type ProfileStatus struct {
	UsedBy     []ObjectRef `json:"usedBy,omitempty"`
	Conditions []Condition `json:"conditions,omitempty"`
}

// IntOrString mirrors the Kubernetes type for fields that may be an integer or
// a template string. Only a literal integer is usable for display.
type IntOrString struct {
	Int    *int
	String string
}

func (v *IntOrString) UnmarshalJSON(data []byte) error {
	var n int
	if err := json.Unmarshal(data, &n); err == nil {
		v.Int = &n
		v.String = strconv.Itoa(n)
		return nil
	}
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		v.String = s
		return nil
	}
	return fmt.Errorf("gpuPerPod must be a number or string")
}

func (v *IntOrString) Number() (int, bool) {
	if v == nil || v.Int == nil {
		return 0, false
	}
	return *v.Int, true
}

// List envelopes returned by the Kubernetes API.

type rawList[T any] struct {
	Metadata struct {
		Continue string `json:"continue"`
	} `json:"metadata"`
	Items []T `json:"items"`
}
