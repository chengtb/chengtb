package api

import (
	"net/http"

	apperrors "github.com/chengtb/chengtb/pkg/errors"
	"github.com/chengtb/chengtb/pkg/i18n"
	"github.com/chengtb/chengtb/internal/infrastructure/transport"
	"github.com/chengtb/chengtb/internal/interfaces/dto"
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

// respondError writes a localised JSON error response.
//
//   - AppErrors are translated using their numeric code.
//   - All other errors (e.g. Gin binding/validation failures) are treated as
//     CodeInvalidArg and mapped to HTTP 400.
func respondError(c *gin.Context, err error) {
	lang := localeFromCtx(c)

	var appErr *apperrors.AppError
	if apperrors.As(err, &appErr) {
		c.JSON(apperrors.HTTPStatus(appErr), dto.ErrorResponse{
			Code:  int(appErr.Code),
			Error: i18n.Translate(appErr.Code, lang),
		})
		return
	}

	// binding / validation errors
	c.JSON(http.StatusBadRequest, dto.ErrorResponse{
		Code:  int(apperrors.CodeInvalidArg),
		Error: i18n.Translate(apperrors.CodeInvalidArg, lang),
	})
}
