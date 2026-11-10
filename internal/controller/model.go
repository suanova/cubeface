package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/cubeface/cubeface/internal/common"
	"github.com/cubeface/cubeface/internal/service"
)

type ModelController struct {
	catalog *service.ModelCatalog
}

func NewModelController(catalog *service.ModelCatalog) *ModelController {
	return &ModelController{catalog: catalog}
}

func (h *ModelController) List(c *gin.Context) {
	common.Success(c, h.catalog.List(c.Request.Context()))
}

func (h *ModelController) Get(c *gin.Context) {
	detail, found := h.catalog.Detail(c.Request.Context(), c.Param("name"))
	if detail != nil && !detail.Available {
		// A cluster-side failure is not a missing model: keep the 200 envelope
		// so the page can show a retryable unavailable state.
		common.Success(c, detail)
		return
	}
	if !found {
		common.Fail(c, common.NewAPIErrorStatus(http.StatusNotFound, "MODEL_NOT_FOUND", "Model not found"))
		return
	}
	common.Success(c, detail)
}
