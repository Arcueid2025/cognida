package middleware

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestBuildAuditLogIncludesExplicitSnapshotOnly(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/api/v1/knowledge/search", nil)
	c.Set("tenant_id", int64(7))
	c.Set("user_id", int64(9))
	SetAuditSnapshot(c, map[string]interface{}{"recommendation_version": "evidence-bound-v1"})

	entry := buildAuditLog(c, "/api/v1/knowledge/search", 12)
	if entry.Details == "" || entry.ResourceType != "" {
		t.Fatalf("unexpected audit entry: %+v", entry)
	}
	if want := `"recommendation_version":"evidence-bound-v1"`; !strings.Contains(entry.Details, want) {
		t.Fatalf("details should contain explicit snapshot, got %s", entry.Details)
	}
}

func TestBuildAuditLogDoesNotAddSnapshotByDefault(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/v1/knowledge-bases", nil)

	entry := buildAuditLog(c, "/api/v1/knowledge-bases", 1)
	if strings.Contains(entry.Details, `"snapshot"`) {
		t.Fatalf("ordinary request must not have snapshot: %s", entry.Details)
	}
}
