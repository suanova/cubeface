package controller

import (
	"net/http"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/suanova/medrouter/internal/config"
	"github.com/suanova/medrouter/internal/kube"
	"github.com/suanova/medrouter/internal/service"
)

func datasetRouter(source kube.Source) *gin.Engine {
	cfg := &config.Config{DatasetLibraryCacheTTL: time.Minute, DatasetLibraryTimeout: time.Second}
	controller := NewDatasetController(service.NewDatasetCatalog(cfg, source))
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.GET("/datasets", controller.List)
	engine.GET("/datasets/:name", controller.Get)
	return engine
}

func TestDatasetListUnavailableEnvelope(t *testing.T) {
	engine := datasetRouter(kube.UnavailableSource{Reason: "DISABLED"})

	status, body := doRequest(t, engine, "/datasets")
	if status != http.StatusOK || body["success"] != true {
		t.Fatalf("status=%d body=%v", status, body)
	}
	data := body["data"].(map[string]any)
	if data["available"] != false || data["reason"] != "DISABLED" {
		t.Fatalf("data = %v", data)
	}
	if datasets, ok := data["datasets"].([]any); !ok || len(datasets) != 0 {
		t.Fatalf("datasets = %v", data["datasets"])
	}
}

func TestDatasetDetailUnavailableEnvelope(t *testing.T) {
	engine := datasetRouter(kube.UnavailableSource{Reason: "FORBIDDEN"})
	status, body := doRequest(t, engine, "/datasets/anything")
	if status != http.StatusOK || body["success"] != true {
		t.Fatalf("status=%d body=%v", status, body)
	}
	data := body["data"].(map[string]any)
	if data["available"] != false || data["reason"] != "FORBIDDEN" {
		t.Fatalf("data = %v", data)
	}
}

func TestDatasetDetailNotFound(t *testing.T) {
	engine := datasetRouter(emptySource{})
	status, body := doRequest(t, engine, "/datasets/missing")
	if status != http.StatusNotFound || body["success"] != false || body["code"] != "DATASET_NOT_FOUND" {
		t.Fatalf("status=%d body=%v", status, body)
	}
}

func TestDatasetListEmptyIsAvailable(t *testing.T) {
	engine := datasetRouter(emptySource{})
	status, body := doRequest(t, engine, "/datasets")
	if status != http.StatusOK {
		t.Fatalf("status = %d", status)
	}
	data := body["data"].(map[string]any)
	if data["available"] != true || data["total"].(float64) != 0 {
		t.Fatalf("data = %v", data)
	}
}
