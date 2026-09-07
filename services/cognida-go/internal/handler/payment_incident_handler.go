package handler

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"cognida/internal/handler/middleware"
	"cognida/internal/model/incident"
)

type PaymentIncidentHandler struct{ repo incident.Repository }

type paymentIncidentDetail struct {
	Incident *incident.PaymentIncident `json:"incident"`
	Timeline []*incident.TimelineEvent `json:"timeline"`
}

func NewPaymentIncidentHandler(repo incident.Repository) *PaymentIncidentHandler {
	return &PaymentIncidentHandler{repo: repo}
}

func (h *PaymentIncidentHandler) Create(c *gin.Context) {
	var req struct {
		OrderID               string          `json:"order_id"`
		Channel               string          `json:"channel"`
		IncidentType          string          `json:"incident_type" binding:"required"`
		Priority              string          `json:"priority"`
		AssessmentRequestID   string          `json:"assessment_request_id"`
		RecommendationVersion string          `json:"recommendation_version"`
		EvidenceSnapshot      json.RawMessage `json:"evidence_snapshot"`
	}
	if !BindJSON(c, &req) {
		return
	}
	v, err := incident.NewPaymentIncident("pi-"+uuid.NewString(), GetTenantID(c), GetUserID(c), req.IncidentType)
	if err != nil {
		BadRequest(c, err.Error())
		return
	}
	if len(req.EvidenceSnapshot) > 64*1024 || (len(req.EvidenceSnapshot) > 0 && !json.Valid(req.EvidenceSnapshot)) {
		BadRequest(c, "evidence_snapshot must be valid JSON and no larger than 64 KiB")
		return
	}
	v.OrderID, v.Channel, v.AssessmentRequestID = strings.TrimSpace(req.OrderID), strings.TrimSpace(req.Channel), strings.TrimSpace(req.AssessmentRequestID)
	v.RecommendationVersion, v.EvidenceSnapshot = strings.TrimSpace(req.RecommendationVersion), string(req.EvidenceSnapshot)
	if req.Priority != "" {
		v.Priority = strings.TrimSpace(req.Priority)
	}
	if err := h.repo.Create(c.Request.Context(), v); err != nil {
		InternalError(c, err.Error())
		return
	}
	if err := h.createEvent(c, v, "created", "案件已创建", "", "", ""); err != nil {
		InternalError(c, err.Error())
		return
	}
	middleware.SetAuditSnapshot(c, map[string]interface{}{"type": "payment_incident_created", "incident_id": v.ID, "assessment_request_id": v.AssessmentRequestID})
	Created(c, v)
}

func (h *PaymentIncidentHandler) Get(c *gin.Context) {
	v, err := h.repo.FindByID(c.Request.Context(), c.Param("id"), GetTenantID(c))
	if err != nil {
		NotFound(c, "案件不存在")
		return
	}
	events, err := h.repo.ListEvents(c.Request.Context(), v.ID, v.TenantID)
	if err != nil {
		InternalError(c, err.Error())
		return
	}
	OK(c, paymentIncidentDetail{Incident: v, Timeline: events})
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
	if err := h.createEvent(c, v, "status_changed", "", from, v.Status, ""); err != nil {
		InternalError(c, err.Error())
		return
	}
	middleware.SetAuditSnapshot(c, map[string]interface{}{"type": "payment_incident_transition", "incident_id": v.ID, "from": from, "to": v.Status})
	OK(c, v)
}

// UpdateDetails records human ownership and conclusion without granting any
// payment execution capability.
func (h *PaymentIncidentHandler) UpdateDetails(c *gin.Context) {
	var req struct {
		AssigneeID *int64  `json:"assignee_id"`
		Conclusion *string `json:"conclusion"`
	}
	if !BindJSON(c, &req) {
		return
	}
	if req.AssigneeID == nil && req.Conclusion == nil {
		BadRequest(c, "assignee_id or conclusion is required")
		return
	}
	if req.AssigneeID != nil && *req.AssigneeID <= 0 {
		BadRequest(c, "assignee_id must be positive")
		return
	}
	v, err := h.repo.FindByID(c.Request.Context(), c.Param("id"), GetTenantID(c))
	if err != nil {
		NotFound(c, "案件不存在")
		return
	}
	if req.AssigneeID != nil {
		v.AssigneeID = req.AssigneeID
	}
	if req.Conclusion != nil {
		v.Conclusion = strings.TrimSpace(*req.Conclusion)
	}
	v.UpdatedAt = time.Now()
	if err := h.repo.Update(c.Request.Context(), v); err != nil {
		InternalError(c, err.Error())
		return
	}
	if req.AssigneeID != nil {
		if err := h.createEvent(c, v, "assignee_changed", "", "", "", ""); err != nil {
			InternalError(c, err.Error())
			return
		}
	}
	if req.Conclusion != nil {
		if err := h.createEvent(c, v, "conclusion_recorded", v.Conclusion, "", "", ""); err != nil {
			InternalError(c, err.Error())
			return
		}
	}
	middleware.SetAuditSnapshot(c, map[string]interface{}{"type": "payment_incident_updated", "incident_id": v.ID})
	OK(c, v)
}

func (h *PaymentIncidentHandler) AddTimelineNote(c *gin.Context) {
	var req struct {
		Content       string `json:"content" binding:"required"`
		AttachmentRef string `json:"attachment_ref"`
	}
	if !BindJSON(c, &req) {
		return
	}
	v, err := h.repo.FindByID(c.Request.Context(), c.Param("id"), GetTenantID(c))
	if err != nil {
		NotFound(c, "案件不存在")
		return
	}
	content, attachment := strings.TrimSpace(req.Content), strings.TrimSpace(req.AttachmentRef)
	if content == "" || len(content) > 8000 || len(attachment) > 512 {
		BadRequest(c, "note content or attachment reference is invalid")
		return
	}
	eventType := "note_added"
	if attachment != "" {
		eventType = "attachment_referenced"
	}
	if err := h.createEvent(c, v, eventType, content, "", "", attachment); err != nil {
		InternalError(c, err.Error())
		return
	}
	middleware.SetAuditSnapshot(c, map[string]interface{}{"type": "payment_incident_timeline_note", "incident_id": v.ID, "event_type": eventType})
	Created(c, map[string]string{"message": "timeline event recorded"})
}

// GetSimulatedOrder, GetSimulatedPayments and GetSimulatedCallbackLogs expose
// local verification fixtures only. They are intentionally GET-only and each
// repository lookup includes the authenticated tenant boundary.
func (h *PaymentIncidentHandler) GetSimulatedOrder(c *gin.Context) {
	order, err := h.repo.FindSimulatedOrder(c.Request.Context(), c.Param("orderID"), GetTenantID(c))
	if err != nil {
		NotFound(c, "模拟订单不存在")
		return
	}
	OK(c, order)
}

func (h *PaymentIncidentHandler) GetSimulatedPayments(c *gin.Context) {
	orderID, tenantID := c.Param("orderID"), GetTenantID(c)
	if _, err := h.repo.FindSimulatedOrder(c.Request.Context(), orderID, tenantID); err != nil {
		NotFound(c, "模拟订单不存在")
		return
	}
	items, err := h.repo.ListSimulatedPayments(c.Request.Context(), orderID, tenantID)
	if err != nil {
		InternalError(c, err.Error())
		return
	}
	OK(c, map[string]interface{}{"order_id": orderID, "items": items})
}

func (h *PaymentIncidentHandler) GetSimulatedCallbackLogs(c *gin.Context) {
	orderID, tenantID := c.Param("orderID"), GetTenantID(c)
	if _, err := h.repo.FindSimulatedOrder(c.Request.Context(), orderID, tenantID); err != nil {
		NotFound(c, "模拟订单不存在")
		return
	}
	items, err := h.repo.ListSimulatedCallbackLogs(c.Request.Context(), orderID, tenantID)
	if err != nil {
		InternalError(c, err.Error())
		return
	}
	OK(c, map[string]interface{}{"order_id": orderID, "items": items})
}

func (h *PaymentIncidentHandler) CreateDispositionDraft(c *gin.Context) {
	var req struct {
		ActionType     string `json:"action_type" binding:"required"`
		IdempotencyKey string `json:"idempotency_key" binding:"required"`
	}
	if !BindJSON(c, &req) {
		return
	}
	v, err := h.repo.FindByID(c.Request.Context(), c.Param("id"), GetTenantID(c))
	if err != nil {
		NotFound(c, "案件不存在")
		return
	}
	key := strings.TrimSpace(req.IdempotencyKey)
	if len(key) < 8 || len(key) > 128 {
		BadRequest(c, "idempotency_key must be 8-128 characters")
		return
	}
	d := &incident.DispositionDraft{ID: "pd-" + uuid.NewString(), TenantID: v.TenantID, IncidentID: v.ID, CreatedBy: GetUserID(c), ActionType: strings.TrimSpace(req.ActionType), Status: "pending_approval", IdempotencyKey: key, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	if err := h.repo.CreateDraft(c.Request.Context(), d); err != nil {
		BadRequest(c, "该幂等键已存在")
		return
	}
	_ = h.createEvent(c, v, "disposition_draft_created", d.ActionType, "", "", "")
	Created(c, d)
}
func (h *PaymentIncidentHandler) DecideDispositionDraft(c *gin.Context) {
	var req struct {
		Approve bool `json:"approve"`
	}
	if !BindJSON(c, &req) {
		return
	}
	d, err := h.repo.FindDraft(c.Request.Context(), c.Param("draftID"), GetTenantID(c))
	if err != nil {
		NotFound(c, "处置草稿不存在")
		return
	}
	if d.Status != "pending_approval" {
		BadRequest(c, "草稿不在待审批状态")
		return
	}
	if d.CreatedBy == GetUserID(c) {
		Forbidden(c, "创建者不能审批自己的草稿")
		return
	}
	uid := GetUserID(c)
	d.ApprovedBy = &uid
	if req.Approve {
		d.Status = "approved"
	} else {
		d.Status = "rejected"
	}
	d.UpdatedAt = time.Now()
	if err := h.repo.UpdateDraft(c.Request.Context(), d); err != nil {
		InternalError(c, err.Error())
		return
	}
	OK(c, d)
}
func (h *PaymentIncidentHandler) ExecuteDispositionDraft(c *gin.Context) {
	d, err := h.repo.FindDraft(c.Request.Context(), c.Param("draftID"), GetTenantID(c))
	if err != nil {
		NotFound(c, "处置草稿不存在")
		return
	}
	if d.Status == "executed" {
		OK(c, d)
		return
	}
	if d.Status != "approved" {
		Forbidden(c, "草稿未经审批，不能执行")
		return
	}
	d.Status = "executed"
	d.ExecutionResult = "模拟处置已记录；未对订单、支付或回调系统执行任何操作。"
	d.UpdatedAt = time.Now()
	if err := h.repo.UpdateDraft(c.Request.Context(), d); err != nil {
		InternalError(c, err.Error())
		return
	}
	OK(c, d)
}

func (h *PaymentIncidentHandler) createEvent(c *gin.Context, v *incident.PaymentIncident, eventType, content string, from, to incident.Status, attachment string) error {
	return h.repo.CreateEvent(c.Request.Context(), &incident.TimelineEvent{ID: "pie-" + uuid.NewString(), IncidentID: v.ID, TenantID: v.TenantID, ActorID: GetUserID(c), EventType: eventType, Content: content, FromStatus: from, ToStatus: to, AttachmentRef: attachment, CreatedAt: time.Now()})
}
