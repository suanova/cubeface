package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/cubeface/cubeface/internal/common"
	"github.com/cubeface/cubeface/internal/service"
)

type DatasetController struct {
	catalog *service.DatasetCatalog
}

func NewDatasetController(catalog *service.DatasetCatalog) *DatasetController {
	return &DatasetController{catalog: catalog}
}

func (h *DatasetController) List(c *gin.Context) {
	common.Success(c, h.catalog.List(c.Request.Context()))
}

func (h *DatasetController) Get(c *gin.Context) {
	detail, found := h.catalog.Detail(c.Request.Context(), c.Param("name"))
	if detail != nil && !detail.Available {
		// A cluster-side failure is not a missing dataset: keep the 200 envelope
		// so the page can show a retryable unavailable state.
		common.Success(c, detail)
		return
	}
	if !found {
		common.Fail(c, common.NewAPIErrorStatus(http.StatusNotFound, "DATASET_NOT_FOUND", "Dataset not found"))
		return
	}
	common.Success(c, detail)
}
