package controller

import (
	"github.com/gin-gonic/gin"

	"github.com/suanova/medrouter/internal/common"
	"github.com/suanova/medrouter/internal/model"
	"github.com/suanova/medrouter/internal/service"
)

type VerificationController struct {
	verification *service.VerificationService
}

func NewVerificationController(verification *service.VerificationService) *VerificationController {
	return &VerificationController{verification: verification}
}

type sendCodeRequest struct {
	Email   string `json:"email"`
	Purpose string `json:"purpose"`
}

func (h *VerificationController) Send(c *gin.Context) {
	var req sendCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, common.NewAPIError("INVALID_REQUEST", "Invalid request body"))
		return
	}
	purpose := req.Purpose
	if purpose == "" {
		purpose = model.PurposeRegister
	}
	if apiErr := h.verification.Send(req.Email, purpose, c.ClientIP()); apiErr != nil {
		common.Fail(c, apiErr)
		return
	}
	common.Success(c, gin.H{"sent": true})
}
