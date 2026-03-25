package api

import (
	"net/http"

	"github.com/chengtb/chengtb/internal/infrastructure/transport"
	"github.com/chengtb/chengtb/internal/interfaces/dto"
	apperrors "github.com/chengtb/chengtb/pkg/errors"
	"github.com/chengtb/chengtb/pkg/i18n"
	"github.com/gin-gonic/gin"
)

// localeFromCtx retrieves the resolved locale stored by localeMiddleware.
// Falls back to "en" when the context key is absent.
func localeFromCtx(c *gin.Context) string {
	if v, exists := c.Get(transport.LocaleKey); exists {
		if s, ok := v.(string); ok && s != "" {
			return s
		}
	}
	return "en"
}

// respondOK writes a successful JSON response using the standard envelope.
// code is always 0, message is "ok", and data carries the payload.
func respondOK(c *gin.Context, httpStatus int, data any) {
	c.JSON(httpStatus, dto.Response{
		Code:    0,
		Message: "ok",
		Data:    data,
	})
}

// respondError writes a localised JSON error response using the standard envelope.
//
//   - AppErrors are translated using their numeric code.
//   - All other errors (e.g. Gin binding/validation failures) are treated as
//     CodeInvalidArg and mapped to HTTP 400.
func respondError(c *gin.Context, err error) {
	lang := localeFromCtx(c)

	var appErr *apperrors.AppError
	if apperrors.As(err, &appErr) {
		c.JSON(apperrors.HTTPStatus(appErr), dto.Response{
			Code:    int(appErr.Code),
			Message: i18n.Translate(appErr.Code, lang),
			Data:    nil,
		})
		return
	}

	// binding / validation errors
	c.JSON(http.StatusBadRequest, dto.Response{
		Code:    int(apperrors.CodeInvalidArg),
		Message: i18n.Translate(apperrors.CodeInvalidArg, lang),
		Data:    nil,
	})
}
