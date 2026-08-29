package i18n

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestKeyParity verifies every locale ships the same key set as English,
// so switching languages can never produce a blank string.
func TestKeyParity(t *testing.T) {
	var en, zh map[string]string
	require.NoError(t, json.Unmarshal(mustRead(t, "en"), &en))
	require.NoError(t, json.Unmarshal(mustRead(t, "zh-CN"), &zh))

	for k := range en {
		_, ok := zh[k]
		require.True(t, ok, "zh-CN is missing key %q", k)
	}
	for k := range zh {
		_, ok := en[k]
		require.True(t, ok, "zh-CN has extra key %q not in en", k)
	}
}

func mustRead(t *testing.T, code string) []byte {
	t.Helper()
	data, err := localesFS.ReadFile("locales/" + code + ".json")
	require.NoError(t, err)
	return data
}

func TestSetLocaleAndT(t *testing.T) {
	SetLocale("zh-CN")
	require.Equal(t, "zh-CN", CurrentLocale())
	require.Equal(t, "命令", T("commands.title"))
	require.Equal(t, "切换模型", T("commands.switch_model"))

	SetLocale("en")
	require.Equal(t, "Commands", T("commands.title"))
}

func TestTUnknownLocaleFallsBackToEnglish(t *testing.T) {
	SetLocale("xx")
	require.Equal(t, "en", CurrentLocale())
	require.Equal(t, "Commands", T("commands.title"))
}

func TestTMissingKeyFallsBackToEnglishThenRaw(t *testing.T) {
	SetLocale("zh-CN")
	// Present in zh-CN: returns the translation.
	require.Equal(t, "命令", T("commands.title"))

	// Missing everywhere: returns the raw key.
	require.Equal(t, "no.such.key", T("no.such.key"))
}

func TestTWithArgs(t *testing.T) {
	SetLocale("zh-CN")
	require.Equal(t, "创建了 3 个待办", T("chat.created_todos", 3))
	require.Equal(t, "…还有 5 个", T("sidebar.more", 5))
}

func TestPlaceholderConsistency(t *testing.T) {
	// Every value in zh-CN that contains format verbs must contain the
	// same set of verbs as its English counterpart, so fmt.Sprintf can
	// never panic with wrong argument counts.
	var en, zh map[string]string
	require.NoError(t, json.Unmarshal(mustRead(t, "en"), &en))
	require.NoError(t, json.Unmarshal(mustRead(t, "zh-CN"), &zh))

	verbs := func(s string) string {
		var b strings.Builder
		for i := 0; i < len(s); i++ {
			if s[i] == '%' && i+1 < len(s) {
				c := s[i+1]
				if c == '%' {
					i++
					continue
				}
				b.WriteByte(c)
			}
		}
		return b.String()
	}

	for k, enV := range en {
		zhV := zh[k]
		if verbs(enV) != verbs(zhV) {
			t.Errorf("key %q: verbs mismatch en=%q zh=%q", k, verbs(enV), verbs(zhV))
		}
	}
}

func TestAvailableLocales(t *testing.T) {
	require.Len(t, Available, 2)
	require.True(t, IsSupported("en"))
	require.True(t, IsSupported("zh-CN"))
	require.False(t, IsSupported("fr"))
}
