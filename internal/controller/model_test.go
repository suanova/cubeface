package controller

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/cubeface/cubeface/internal/config"
	"github.com/cubeface/cubeface/internal/kube"
	"github.com/cubeface/cubeface/internal/service"
)

type emptySource struct{}

func (emptySource) ListModelVersions(context.Context) ([]kube.ModelVersion, error) {
	return []kube.ModelVersion{}, nil
}
func (emptySource) ListRuntimeProfiles(context.Context) ([]kube.RuntimeProfile, error) {
	return []kube.RuntimeProfile{}, nil
}

func modelRouter(source kube.Source) *gin.Engine {
	cfg := &config.Config{ModelLibraryCacheTTL: time.Minute, ModelLibraryTimeout: time.Second}
	controller := NewModelController(service.NewModelCatalog(cfg, source))
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.GET("/models", controller.List)
	engine.GET("/models/:name", controller.Get)
	return engine
}

func doRequest(t *testing.T, engine *gin.Engine, path string) (int, map[string]any) {
	t.Helper()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	engine.ServeHTTP(rec, req)
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode %s: %v (%s)", path, err, rec.Body.String())
	}
	return rec.Code, body
}

func TestModelListUnavailableEnvelope(t *testing.T) {
	engine := modelRouter(kube.DisabledSource)

	status, body := doRequest(t, engine, "/models")
	if status != http.StatusOK || body["success"] != true {
		t.Fatalf("status=%d body=%v", status, body)
	}
	data := body["data"].(map[string]any)
	if data["available"] != false || data["reason"] != "DISABLED" {
		t.Fatalf("data = %v", data)
	}
	if models, ok := data["models"].([]any); !ok || len(models) != 0 {
		t.Fatalf("models = %v", data["models"])
	}
}

func TestModelDetailUnavailableEnvelope(t *testing.T) {
	engine := modelRouter(kube.DisabledSource)
	status, body := doRequest(t, engine, "/models/anything")
	if status != http.StatusOK || body["success"] != true {
		t.Fatalf("status=%d body=%v", status, body)
	}
	data := body["data"].(map[string]any)
	if data["available"] != false || data["reason"] != "DISABLED" {
		t.Fatalf("data = %v", data)
	}
}

func TestModelDetailNotFound(t *testing.T) {
	engine := modelRouter(emptySource{})
	status, body := doRequest(t, engine, "/models/missing")
	if status != http.StatusNotFound || body["success"] != false || body["code"] != "MODEL_NOT_FOUND" {
		t.Fatalf("status=%d body=%v", status, body)
	}
}

func TestModelListEmptyIsAvailable(t *testing.T) {
	engine := modelRouter(emptySource{})
	status, body := doRequest(t, engine, "/models")
	if status != http.StatusOK {
		t.Fatalf("status = %d", status)
	}
	data := body["data"].(map[string]any)
	if data["available"] != true || data["total"].(float64) != 0 {
		t.Fatalf("data = %v", data)
	}
}
