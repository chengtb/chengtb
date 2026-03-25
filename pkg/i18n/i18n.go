// Package i18n provides lightweight internationalisation support for API error
// messages.  Locale resolution is driven by the HTTP Accept-Language header,
// and messages are looked up from a built-in catalog.
//
// Supported locales: "en" (default), "zh".
package i18n

import (
	"strings"

	apperrors "github.com/chengtb/chengtb/pkg/errors"
)

// catalog maps locale → error code → human-readable message.
var catalog = map[string]map[apperrors.Code]string{
	"en": {
		apperrors.CodeUnknown:        "unknown error",
		apperrors.CodeNotFound:       "resource not found",
		apperrors.CodeInvalidArg:     "invalid argument",
		apperrors.CodeInternal:       "internal server error",
		apperrors.CodeConflict:       "resource conflict",
		apperrors.CodeUnauthenticated: "unauthenticated",
	},
	"zh": {
		apperrors.CodeUnknown:        "未知错误",
		apperrors.CodeNotFound:       "资源不存在",
		apperrors.CodeInvalidArg:     "请求参数无效",
		apperrors.CodeInternal:       "服务器内部错误",
		apperrors.CodeConflict:       "资源冲突",
		apperrors.CodeUnauthenticated: "未认证",
	},
}

const defaultLocale = "en"

// Translate returns the localised message for the given error code and locale.
// If the locale or code has no entry, the English fallback is returned.
func Translate(code apperrors.Code, locale string) string {
	if msgs, ok := catalog[locale]; ok {
		if msg, ok := msgs[code]; ok {
			return msg
		}
	}
	// fall back to English
	if msgs, ok := catalog[defaultLocale]; ok {
		if msg, ok := msgs[code]; ok {
			return msg
		}
	}
	return "error"
}

// ParseLocale extracts the best-matching supported locale from an
// Accept-Language header value (e.g. "zh-CN,zh;q=0.9,en;q=0.8").
// Returns "en" when no supported locale is found.
func ParseLocale(acceptLang string) string {
	if acceptLang == "" {
		return defaultLocale
	}
	for _, part := range strings.Split(acceptLang, ",") {
		// each part is like "zh-CN" or "zh-CN;q=0.9"
		tag := strings.TrimSpace(strings.SplitN(part, ";", 2)[0])
		tag = strings.ToLower(strings.TrimSpace(tag))
		switch {
		case strings.HasPrefix(tag, "zh"):
			return "zh"
		case strings.HasPrefix(tag, "en"):
			return "en"
		}
	}
	return defaultLocale
}
