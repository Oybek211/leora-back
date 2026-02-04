package localization

import (
	"context"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
)

var db *sqlx.DB

func SetDB(conn *sqlx.DB) {
	db = conn
}

func ResolveLanguage(c *fiber.Ctx) string {
	if c == nil {
		return ""
	}
	if lang := strings.TrimSpace(c.Get("X-Lang")); lang != "" {
		return normalizeLang(lang)
	}
	if header := strings.TrimSpace(c.Get("Accept-Language")); header != "" {
		parts := strings.Split(header, ",")
		if len(parts) > 0 {
			return normalizeLang(parts[0])
		}
	}
	return ""
}

func normalizeLang(raw string) string {
	raw = strings.ToLower(strings.TrimSpace(raw))
	if raw == "" {
		return ""
	}
	if strings.Contains(raw, ";") {
		raw = strings.Split(raw, ";")[0]
	}
	if strings.Contains(raw, "-") {
		raw = strings.Split(raw, "-")[0]
	}
	return raw
}

func defaultLanguage(ctx context.Context) string {
	if db == nil {
		return ""
	}
	var code string
	_ = db.GetContext(ctx, &code, `SELECT code FROM meta_languages WHERE is_default = true AND is_active = true LIMIT 1`)
	return strings.TrimSpace(code)
}

func TranslateError(ctx context.Context, code, fallback, lang string) string {
	if db == nil {
		return fallback
	}
	resolvedLang := strings.TrimSpace(lang)
	if resolvedLang == "" {
		resolvedLang = defaultLanguage(ctx)
	}
	if resolvedLang == "" || code == "" {
		return fallback
	}
	var message string
	err := db.GetContext(ctx, &message, `
		SELECT message FROM error_translations
		WHERE code = $1 AND lang_code = $2
		LIMIT 1
	`, code, resolvedLang)
	if err != nil || strings.TrimSpace(message) == "" {
		return fallback
	}
	return message
}
