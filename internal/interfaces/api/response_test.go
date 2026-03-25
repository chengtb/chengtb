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

func parseErrorResponse(t *testing.T, w *httptest.ResponseRecorder) dto.ErrorResponse {
	t.Helper()
	var resp dto.ErrorResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	return resp
}

func TestRespondError_AppError_English(t *testing.T) {
	c, w := newTestContext("en")
	err := apperrors.New(apperrors.CodeNotFound, "user not found")
	respondErrorHelper(c, err)

	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", w.Code, http.StatusNotFound)
	}
	resp := parseErrorResponse(t, w)
	if resp.Error != "resource not found" {
		t.Errorf("error = %q, want %q", resp.Error, "resource not found")
	}
	if resp.Code != int(apperrors.CodeNotFound) {
		t.Errorf("code = %d, want %d", resp.Code, apperrors.CodeNotFound)
	}
}

func TestRespondError_AppError_Chinese(t *testing.T) {
	c, w := newTestContext("zh")
	err := apperrors.New(apperrors.CodeNotFound, "user not found")
	respondErrorHelper(c, err)

	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", w.Code, http.StatusNotFound)
	}
	resp := parseErrorResponse(t, w)
	if resp.Error != "资源不存在" {
		t.Errorf("error = %q, want %q", resp.Error, "资源不存在")
	}
}

func TestRespondError_BindingError_English(t *testing.T) {
	c, w := newTestContext("en")
	// simulate a binding/validation error (non-AppError)
	respondErrorHelper(c, fmt.Errorf("Key: 'CreateUserRequest.Email' validation failed"))

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
	resp := parseErrorResponse(t, w)
	if resp.Error != "invalid argument" {
		t.Errorf("error = %q, want %q", resp.Error, "invalid argument")
	}
	if resp.Code != int(apperrors.CodeInvalidArg) {
		t.Errorf("code = %d, want %d", resp.Code, apperrors.CodeInvalidArg)
	}
}

func TestRespondError_BindingError_Chinese(t *testing.T) {
	c, w := newTestContext("zh")
	respondErrorHelper(c, fmt.Errorf("validation failed"))

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
	resp := parseErrorResponse(t, w)
	if resp.Error != "请求参数无效" {
		t.Errorf("error = %q, want %q", resp.Error, "请求参数无效")
	}
}

func TestRespondError_Conflict_Chinese(t *testing.T) {
	c, w := newTestContext("zh")
	err := apperrors.New(apperrors.CodeConflict, "email already in use")
	respondErrorHelper(c, err)

	if w.Code != http.StatusConflict {
		t.Errorf("status = %d, want %d", w.Code, http.StatusConflict)
	}
	resp := parseErrorResponse(t, w)
	if resp.Error != "资源冲突" {
		t.Errorf("error = %q, want %q", resp.Error, "资源冲突")
	}
}

func TestRespondError_NoLocaleKey_DefaultsToEnglish(t *testing.T) {
	// deliberately no locale set
	c, w := newTestContext("")

	err := apperrors.New(apperrors.CodeInternal, "db down")
	respondErrorHelper(c, err)

	resp := parseErrorResponse(t, w)
	if resp.Error != "internal server error" {
		t.Errorf("error = %q, want %q", resp.Error, "internal server error")
	}
}

