package handler

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"cognida/internal/handler/middleware"
	"cognida/internal/model/incident"
)

type PaymentIncidentHandler struct{ repo incident.Repository }

func NewPaymentIncidentHandler(repo incident.Repository) *PaymentIncidentHandler {
	return &PaymentIncidentHandler{repo: repo}
}

func (h *PaymentIncidentHandler) Create(c *gin.Context) {
	var req struct {
		OrderID             string `json:"order_id"`
		Channel             string `json:"channel"`
		IncidentType        string `json:"incident_type" binding:"required"`
		Priority            string `json:"priority"`
		AssessmentRequestID string `json:"assessment_request_id"`
	}
	if !BindJSON(c, &req) {
		return
	}
	v, err := incident.NewPaymentIncident("pi-"+uuid.NewString(), GetTenantID(c), GetUserID(c), req.IncidentType)
	if err != nil {
		BadRequest(c, err.Error())
		return
	}
	v.OrderID, v.Channel, v.AssessmentRequestID = strings.TrimSpace(req.OrderID), strings.TrimSpace(req.Channel), strings.TrimSpace(req.AssessmentRequestID)
	if req.Priority != "" {
		v.Priority = strings.TrimSpace(req.Priority)
	}
	if err := h.repo.Create(c.Request.Context(), v); err != nil {
		InternalError(c, err.Error())
		return
	}
	middleware.SetAuditSnapshot(c, map[string]interface{}{"type": "payment_incident_created", "incident_id": v.ID, "assessment_request_id": v.AssessmentRequestID})
	Created(c, v)
}

func (h *PaymentIncidentHandler) List(c *gin.Context) {
	page, size := GetPageParams(c)
	items, total, err := h.repo.List(c.Request.Context(), GetTenantID(c), page, size)
	if err != nil {
		InternalError(c, err.Error())
		return
	}
	PageSuccessJSON(c, total, items, page, size)
}

func (h *PaymentIncidentHandler) Transition(c *gin.Context) {
	var req struct {
		Status incident.Status `json:"status" binding:"required"`
	}
	if !BindJSON(c, &req) {
		return
	}
	v, err := h.repo.FindByID(c.Request.Context(), c.Param("id"), GetTenantID(c))
	if err != nil {
		NotFound(c, err.Error())
		return
	}
	from := v.Status
	if err := v.TransitionTo(req.Status); err != nil {
		BadRequest(c, err.Error())
		return
	}
	if err := h.repo.Update(c.Request.Context(), v); err != nil {
		InternalError(c, err.Error())
		return
	}
	middleware.SetAuditSnapshot(c, map[string]interface{}{"type": "payment_incident_transition", "incident_id": v.ID, "from": from, "to": v.Status})
	OK(c, v)
}
