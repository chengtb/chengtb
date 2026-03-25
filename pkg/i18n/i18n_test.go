package i18n_test

import (
	"testing"

	apperrors "github.com/chengtb/chengtb/pkg/errors"
	"github.com/chengtb/chengtb/pkg/i18n"
)

func TestTranslate_English(t *testing.T) {
	cases := []struct {
		code apperrors.Code
		want string
	}{
		{apperrors.CodeNotFound, "resource not found"},
		{apperrors.CodeInvalidArg, "invalid argument"},
		{apperrors.CodeInternal, "internal server error"},
		{apperrors.CodeConflict, "resource conflict"},
		{apperrors.CodeUnauthenticated, "unauthenticated"},
		{apperrors.CodeUnknown, "unknown error"},
	}
	for _, tc := range cases {
		got := i18n.Translate(tc.code, "en")
		if got != tc.want {
			t.Errorf("Translate(%d, en) = %q, want %q", tc.code, got, tc.want)
		}
	}
}

func TestTranslate_Chinese(t *testing.T) {
	cases := []struct {
		code apperrors.Code
		want string
	}{
		{apperrors.CodeNotFound, "资源不存在"},
		{apperrors.CodeInvalidArg, "请求参数无效"},
		{apperrors.CodeInternal, "服务器内部错误"},
		{apperrors.CodeConflict, "资源冲突"},
		{apperrors.CodeUnauthenticated, "未认证"},
		{apperrors.CodeUnknown, "未知错误"},
	}
	for _, tc := range cases {
		got := i18n.Translate(tc.code, "zh")
		if got != tc.want {
			t.Errorf("Translate(%d, zh) = %q, want %q", tc.code, got, tc.want)
		}
	}
}

func TestTranslate_FallsBackToEnglish(t *testing.T) {
	got := i18n.Translate(apperrors.CodeNotFound, "fr")
	want := "resource not found"
	if got != want {
		t.Errorf("Translate(NotFound, fr) = %q, want %q", got, want)
	}
}

func TestTranslate_EmptyLocale(t *testing.T) {
	got := i18n.Translate(apperrors.CodeInternal, "")
	want := "internal server error"
	if got != want {
		t.Errorf("Translate(Internal, \"\") = %q, want %q", got, want)
	}
}

func TestParseLocale(t *testing.T) {
	cases := []struct {
		header string
		want   string
	}{
		{"", "en"},
		{"en", "en"},
		{"en-US", "en"},
		{"zh", "zh"},
		{"zh-CN", "zh"},
		{"zh-TW", "zh"},
		{"zh-CN,zh;q=0.9,en;q=0.8", "zh"},
		{"en-US,en;q=0.9,zh;q=0.8", "en"},
		{"fr-FR,fr;q=0.9", "en"},  // unsupported language → default
		{"fr,zh;q=0.8", "zh"},     // second entry is zh
	}
	for _, tc := range cases {
		got := i18n.ParseLocale(tc.header)
		if got != tc.want {
			t.Errorf("ParseLocale(%q) = %q, want %q", tc.header, got, tc.want)
		}
	}
}
