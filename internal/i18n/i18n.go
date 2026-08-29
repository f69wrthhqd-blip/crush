// Package i18n provides lightweight, key-based localization for the
// Crush TUI. It only covers user-facing display strings rendered by the
// interface; agent system prompts and tool templates are intentionally
// not localized.
package i18n

import (
	"embed"
	"encoding/json"
	"fmt"
	"sync"
)

//go:embed locales/*.json
var localesFS embed.FS

// Locale describes an available language.
type Locale struct {
	Code string // stable identifier used in config, e.g. "en", "zh-CN"
	Name string // display name shown in the language picker
}

// Available lists every bundled locale in display order.
var Available = []Locale{
	{Code: "en", Name: "English"},
	{Code: "zh-CN", Name: "简体中文"},
}

var (
	mu          sync.RWMutex
	currentCode = "en"
	catalogs    = map[string]map[string]string{}
	catalogOnce = map[string]*sync.Once{}
)

func onceFor(code string) *sync.Once {
	if o, ok := catalogOnce[code]; ok {
		return o
	}
	o := &sync.Once{}
	catalogOnce[code] = o
	return o
}

func loadCatalog(code string) map[string]string {
	catalogOnce[code] = onceFor(code)
	once := catalogOnce[code]
	once.Do(func() {
		data, err := localesFS.ReadFile("locales/" + code + ".json")
		if err != nil {
			return
		}
		var m map[string]string
		if err := json.Unmarshal(data, &m); err != nil {
			return
		}
		mu.Lock()
		catalogs[code] = m
		mu.Unlock()
	})
	mu.RLock()
	m := catalogs[code]
	mu.RUnlock()
	return m
}

// SetLocale switches the active UI language to the given code. Unknown
// codes fall back to English. It is safe for concurrent use.
func SetLocale(code string) {
	if !IsSupported(code) {
		code = "en"
	}
	mu.Lock()
	currentCode = code
	mu.Unlock()
}

// CurrentLocale returns the active locale code.
func CurrentLocale() string {
	mu.RLock()
	defer mu.RUnlock()
	return currentCode
}

// IsSupported reports whether a locale code is bundled.
func IsSupported(code string) bool {
	for _, l := range Available {
		if l.Code == code {
			return true
		}
	}
	return false
}

// T returns the translated string for key in the active locale. When the
// key is missing from the active catalog it falls back to English, then
// to the raw key so incomplete translations never render blank.
func T(key string, args ...any) string {
	mu.RLock()
	code := currentCode
	mu.RUnlock()

	if s, ok := lookup(code, key); ok {
		return format(s, args)
	}
	if code != "en" {
		if s, ok := lookup("en", key); ok {
			return format(s, args)
		}
	}
	return format(key, args)
}

func lookup(code, key string) (string, bool) {
	catalog := loadCatalog(code)
	if catalog == nil {
		return "", false
	}
	s, ok := catalog[key]
	return s, ok
}

func format(s string, args []any) string {
	if len(args) == 0 {
		return s
	}
	return fmt.Sprintf(s, args...)
}
