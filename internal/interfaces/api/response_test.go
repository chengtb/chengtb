package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/chengtb/chengtb/internal/infrastructure/transport"
	"github.com/chengtb/chengtb/internal/interfaces/dto"
	apperrors "github.com/chengtb/chengtb/pkg/errors"
	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// newTestContext creates a gin.Context backed by a ResponseRecorder and stores
// the given locale under transport.LocaleKey (simulates localeMiddleware).
func newTestContext(locale string) (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	if locale != "" {
		c.Set(transport.LocaleKey, locale)
	}
	return c, w
}

// respondErrorHelper is a thin wrapper so tests can call respondError
// (which lives in response.go, same package).
func respondErrorHelper(c *gin.Context, err error) { respondError(c, err) }

func parseResponse(t *testing.T, w *httptest.ResponseRecorder) dto.Response {
	t.Helper()
	var resp dto.Response
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	return resp
}

// ---- respondError tests ----

func TestRespondError_AppError_English(t *testing.T) {
	c, w := newTestContext("en")
	err := apperrors.New(apperrors.CodeNotFound, "user not found")
	respondErrorHelper(c, err)

	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", w.Code, http.StatusNotFound)
	}
	resp := parseResponse(t, w)
	if resp.Code != int(apperrors.CodeNotFound) {
		t.Errorf("code = %d, want %d", resp.Code, apperrors.CodeNotFound)
	}
	if resp.Message != "resource not found" {
		t.Errorf("message = %q, want %q", resp.Message, "resource not found")
	}
	if resp.Data != nil {
		t.Errorf("data = %v, want nil", resp.Data)
	}
}

func TestRespondError_AppError_Chinese(t *testing.T) {
	c, w := newTestContext("zh")
	err := apperrors.New(apperrors.CodeNotFound, "user not found")
	respondErrorHelper(c, err)

	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", w.Code, http.StatusNotFound)
	}
	resp := parseResponse(t, w)
	if resp.Message != "资源不存在" {
		t.Errorf("message = %q, want %q", resp.Message, "资源不存在")
	}
}

func TestRespondError_BindingError_English(t *testing.T) {
	c, w := newTestContext("en")
	// simulate a binding/validation error (non-AppError)
	respondErrorHelper(c, fmt.Errorf("Key: 'CreateUserRequest.Email' validation failed"))

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
	resp := parseResponse(t, w)
	if resp.Code != int(apperrors.CodeInvalidArg) {
		t.Errorf("code = %d, want %d", resp.Code, apperrors.CodeInvalidArg)
	}
	if resp.Message != "invalid argument" {
		t.Errorf("message = %q, want %q", resp.Message, "invalid argument")
	}
}

func TestRespondError_BindingError_Chinese(t *testing.T) {
	c, w := newTestContext("zh")
	respondErrorHelper(c, fmt.Errorf("validation failed"))

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
	resp := parseResponse(t, w)
	if resp.Message != "请求参数无效" {
		t.Errorf("message = %q, want %q", resp.Message, "请求参数无效")
	}
}

func TestRespondError_Conflict_Chinese(t *testing.T) {
	c, w := newTestContext("zh")
	err := apperrors.New(apperrors.CodeConflict, "email already in use")
	respondErrorHelper(c, err)

	if w.Code != http.StatusConflict {
		t.Errorf("status = %d, want %d", w.Code, http.StatusConflict)
	}
	resp := parseResponse(t, w)
	if resp.Message != "资源冲突" {
		t.Errorf("message = %q, want %q", resp.Message, "资源冲突")
	}
}

func TestRespondError_NoLocaleKey_DefaultsToEnglish(t *testing.T) {
	// deliberately no locale set
	c, w := newTestContext("")

	err := apperrors.New(apperrors.CodeInternal, "db down")
	respondErrorHelper(c, err)

	resp := parseResponse(t, w)
	if resp.Message != "internal server error" {
		t.Errorf("message = %q, want %q", resp.Message, "internal server error")
	}
}

// ---- respondOK tests ----

func TestRespondOK_WithData(t *testing.T) {
	c, w := newTestContext("en")
	respondOK(c, http.StatusOK, map[string]string{"id": "1", "name": "Alice"})

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
	resp := parseResponse(t, w)
	if resp.Code != 0 {
		t.Errorf("code = %d, want 0", resp.Code)
	}
	if resp.Message != "ok" {
		t.Errorf("message = %q, want %q", resp.Message, "ok")
	}
	if resp.Data == nil {
		t.Error("data is nil, want non-nil payload")
	}
}

func TestRespondOK_Created(t *testing.T) {
	c, w := newTestContext("en")
	respondOK(c, http.StatusCreated, map[string]string{"id": "42"})

	if w.Code != http.StatusCreated {
		t.Errorf("status = %d, want %d", w.Code, http.StatusCreated)
	}
	resp := parseResponse(t, w)
	if resp.Code != 0 {
		t.Errorf("code = %d, want 0", resp.Code)
	}
}

func TestRespondOK_NilData(t *testing.T) {
	c, w := newTestContext("en")
	respondOK(c, http.StatusOK, nil)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
	resp := parseResponse(t, w)
	if resp.Code != 0 {
		t.Errorf("code = %d, want 0", resp.Code)
	}
	if resp.Message != "ok" {
		t.Errorf("message = %q, want %q", resp.Message, "ok")
	}
}
