package styles

import (
	"image/color"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/exp/charmtone"
)

// ThemeEntry describes a selectable UI theme.
type ThemeEntry struct {
	// Key is the stable identifier stored in the config under
	// options.tui.theme.
	Key string
	// Name is the human-readable display name shown in the theme picker.
	Name string
	// Build constructs the theme's styles.
	Build func() Styles
	// Swatches are representative accent colors rendered as preview
	// chips in the theme picker. They mirror the palette of Build.
	Swatches []color.Color
}

// ThemeKeyForProvider returns a stable identifier for the theme
// associated with the given provider ID. Providers that share a theme
// yield the same key, so callers can cheaply detect when switching
// providers would not actually change the active theme and skip the
// expensive style rebuild. This is the single source of truth for the
// provider-to-theme mapping; [ThemeForProvider] builds on it.
func ThemeKeyForProvider(providerID string) string {
	switch providerID {
	case "hyper":
		return "hyper"
	default:
		return "pantera"
	}
}

// ThemeForProvider returns the Styles associated with the given provider
// ID. Unknown or empty provider IDs yield the default Charmtone Pantera
// theme.
func ThemeForProvider(providerID string) Styles {
	s, _ := ThemeByKey(ThemeKeyForProvider(providerID))
	return s
}

// EffectiveThemeKey resolves which registered theme applies: a
// user-selected theme key (from options.tui.theme) wins as long as it is
// known; otherwise the theme follows the given provider ID. The returned
// key can be compared against a previously applied key to skip rebuilds.
func EffectiveThemeKey(configKey, providerID string) string {
	if configKey != "" {
		if _, ok := themeRegistry[configKey]; ok {
			return configKey
		}
	}
	return ThemeKeyForProvider(providerID)
}

// ThemeByKey returns the Styles registered under the given theme key.
// The boolean result reports whether the key is known.
func ThemeByKey(key string) (Styles, bool) {
	entry, ok := themeRegistry[key]
	if !ok {
		return CharmtonePantera(), false
	}
	return entry.Build(), true
}

// ThemeFromOptions resolves the active theme from the user-selected
// theme key (may be empty) and the fallback provider ID. It returns the
// built styles together with the effective theme key.
func ThemeFromOptions(configKey, providerID string) (Styles, string) {
	key := EffectiveThemeKey(configKey, providerID)
	s, _ := ThemeByKey(key)
	return s, key
}

// ThemeName returns the display name registered for the given theme
// key, falling back to the key itself when unknown.
func ThemeName(key string) string {
	if entry, ok := themeRegistry[key]; ok {
		return entry.Name
	}
	return key
}

// AvailableThemes returns the themes offered in the theme picker, in
// display order. Provider-derived aliases such as "hyper" are internal
// fallbacks and intentionally not listed.
func AvailableThemes() []ThemeEntry {
	return listedThemes
}

// listedThemes holds the selectable themes in display order; hyper is
// registered separately below as a provider-only alias.
var (
	listedThemes = []ThemeEntry{
		{
			Key:   "pantera",
			Name:  "Charmtone Pantera",
			Build: CharmtonePantera,
			Swatches: []color.Color{
				charmtone.Charple, charmtone.Bok, charmtone.Mustard,
				charmtone.Coral, charmtone.Pepper,
			},
		},
		{
			Key:   "catppuccin-mocha",
			Name:  "Catppuccin Mocha",
			Build: CatppuccinMocha,
			Swatches: []color.Color{
				cpMochaPrimary, cpMochaSuccess, cpMochaWarning,
				cpMochaError, cpMochaBg,
			},
		},
		{
			Key:   "gruvbox-dark",
			Name:  "Gruvbox Dark",
			Build: GruvboxDark,
			Swatches: []color.Color{
				gvPrimary, gvSuccess, gvWarning,
				gvError, gvBg,
			},
		},
		{
			Key:   "tokyonight",
			Name:  "Tokyo Night",
			Build: TokyoNight,
			Swatches: []color.Color{
				tnPrimary, tnSuccess, tnWarning,
				tnError, tnBg,
			},
		},
		{
			Key:   "nord",
			Name:  "Nord",
			Build: Nord,
			Swatches: []color.Color{
				ndPrimary, ndSuccess, ndWarning,
				ndError, ndBg,
			},
		},
		{
			Key:   "dracula",
			Name:  "Dracula",
			Build: Dracula,
			Swatches: []color.Color{
				dcPrimary, dcSuccess, dcWarning,
				dcError, dcBg,
			},
		},
		{
			Key:   "one-dark",
			Name:  "One Dark",
			Build: OneDark,
			Swatches: []color.Color{
				odPrimary, odSuccess, odWarning,
				odError, odBg,
			},
		},
		{
			Key:   "ayu-dark",
			Name:  "Ayu Dark",
			Build: AyuDark,
			Swatches: []color.Color{
				ayPrimary, aySuccess, ayWarning,
				ayError, ayBg,
			},
		},
		{
			Key:   "vesper",
			Name:  "Vesper",
			Build: Vesper,
			Swatches: []color.Color{
				vsPrimary, vsSuccess, vsWarning,
				vsError, vsBg,
			},
		},
		{
			Key:   "catppuccin-latte",
			Name:  "Catppuccin Latte",
			Build: CatppuccinLatte,
			Swatches: []color.Color{
				cpLattePrimary, cpLatteSuccess, cpLatteWarning,
				cpLatteError, cpLatteBg,
			},
		},
		{
			Key:   "solarized-light",
			Name:  "Solarized Light",
			Build: SolarizedLight,
			Swatches: []color.Color{
				slPrimary, slSuccess, slWarning,
				slError, slBg,
			},
		},
	}

	themeRegistry = func() map[string]ThemeEntry {
		m := make(map[string]ThemeEntry, len(listedThemes)+1)
		for _, t := range listedThemes {
			m[t.Key] = t
		}
		m["hyper"] = ThemeEntry{Key: "hyper", Name: "Hypercrush Obsidiana", Build: HypercrushObsidiana}
		return m
	}()
)

// css builds a color from a "#rrggbb" literal.
func css(s string) color.Color { return lipgloss.Color(s) }

// CharmtonePantera returns the Charmtone dark theme. It's the default style
// for the UI.
func CharmtonePantera() Styles {
	s := quickStyle(quickStyleOpts{
		primary:   charmtone.Charple,
		secondary: charmtone.Dolly,
		accent:    charmtone.Bok,
		keyword:   charmtone.Blush,

		fgBase:       charmtone.Sash,
		fgMoreSubtle: charmtone.Squid,
		fgSubtle:     charmtone.Smoke,
		fgMostSubtle: charmtone.Oyster,

		onPrimary: charmtone.Butter,

		bgBase:         charmtone.Pepper,
		bgLeastVisible: charmtone.BBQ,
		bgLessVisible:  charmtone.Char,
		bgMostVisible:  charmtone.Iron,

		separator: charmtone.Char,

		destructive:       charmtone.Coral,
		error:             charmtone.Sriracha,
		warningSubtle:     charmtone.Zest,
		warning:           charmtone.Mustard,
		attention:         charmtone.Tang,
		busy:              charmtone.Citron,
		info:              charmtone.Malibu,
		infoMoreSubtle:    charmtone.Sardine,
		infoMostSubtle:    charmtone.Damson,
		success:           charmtone.Julep,
		successMoreSubtle: charmtone.Bok,
		successMostSubtle: charmtone.Guac,

		// ANSI 16-color palette for remapping raw terminal output
		// (e.g. bang-mode shell commands) onto legible Charmtone colors.
		ansiBlack:   charmtone.BBQ,
		ansiRed:     charmtone.Coral,
		ansiGreen:   charmtone.Guac,
		ansiYellow:  charmtone.Mustard,
		ansiBlue:    charmtone.Charple,
		ansiMagenta: charmtone.Dolly,
		ansiCyan:    charmtone.Malibu,
		ansiWhite:   charmtone.Smoke,

		ansiBrightBlack:   charmtone.Iron,
		ansiBrightRed:     charmtone.Tuna,
		ansiBrightGreen:   charmtone.Julep,
		ansiBrightYellow:  charmtone.Zest,
		ansiBrightBlue:    charmtone.Guppy,
		ansiBrightMagenta: charmtone.Blush,
		ansiBrightCyan:    charmtone.Sardine,
		ansiBrightWhite:   charmtone.Salt,
	})

	// Bang > prompt overrides - use Salt/Hazy/Larple colors.
	s.Editor.PromptBangArrowFocused = s.Editor.PromptBangArrowFocused.
		Foreground(charmtone.Salt)
	s.Editor.PromptBangDotsFocused = s.Editor.PromptBangDotsFocused.
		Foreground(charmtone.Hazy)
	s.Editor.PromptBangDotsBlurred = s.Editor.PromptBangDotsBlurred.
		Foreground(charmtone.Larple)

	// Shell bar/prompt overrides - use Charple/Iron/Hazy colors.
	s.Messages.ShellBarFocused = s.Messages.ShellBarFocused.
		BorderForeground(charmtone.Charple)
	s.Messages.ShellBarBlurred = s.Messages.ShellBarBlurred.
		BorderForeground(charmtone.Iron)
	s.Messages.ShellPrompt = s.Messages.ShellPrompt.
		Foreground(charmtone.Hazy)
	s.Messages.ShellPromptBlurred = s.Messages.ShellPromptBlurred.
		Foreground(charmtone.Hazy)

	return s
}

// HypercrushObsidiana returns the Hypercrush dark theme.
func HypercrushObsidiana() Styles {
	return CharmtonePantera()
}

// Catppuccin Mocha palette, shared between the theme builder and its
// picker swatches. https://github.com/catppuccin/catppuccin
var (
	cpMochaPrimary = css("#cba6f7") // mauve
	cpMochaSuccess = css("#a6e3a1") // green
	cpMochaWarning = css("#f9e2af") // yellow
	cpMochaError   = css("#f38ba8") // red
	cpMochaBg      = css("#1e1e2e") // base
)

// CatppuccinMocha returns the Catppuccin Mocha dark theme.
func CatppuccinMocha() Styles {
	return quickStyle(quickStyleOpts{
		primary:   cpMochaPrimary,
		secondary: css("#89b4fa"), // blue
		accent:    css("#94e2d5"), // teal
		keyword:   css("#f5c2e7"), // pink

		fgBase:       css("#cdd6f4"), // text
		fgMoreSubtle: css("#7f849c"), // overlay1
		fgSubtle:     css("#a6adc8"), // subtext0
		fgMostSubtle: css("#6c7086"), // overlay0

		onPrimary: css("#11111b"), // crust

		bgBase:         cpMochaBg,
		bgLeastVisible: css("#181825"), // mantle
		bgLessVisible:  css("#262637"),
		bgMostVisible:  css("#313244"), // surface0

		separator: css("#313244"), // surface0

		destructive:       css("#eba0ac"), // maroon
		error:             cpMochaError,
		warningSubtle:     css("#fab387"), // peach
		warning:           cpMochaWarning,
		attention:         css("#fab387"), // peach
		busy:              css("#b4befe"), // lavender
		info:              css("#89dceb"), // sky
		infoMoreSubtle:    css("#74c7ec"), // sapphire
		infoMostSubtle:    css("#89b4fa"), // blue
		success:           cpMochaSuccess,
		successMoreSubtle: css("#94e2d5"), // teal
		successMostSubtle: css("#74c7ec"), // sapphire

		ansiBlack:   css("#45475a"), // surface1
		ansiRed:     css("#f38ba8"),
		ansiGreen:   css("#a6e3a1"),
		ansiYellow:  css("#f9e2af"),
		ansiBlue:    css("#89b4fa"),
		ansiMagenta: css("#f5c2e7"),
		ansiCyan:    css("#94e2d5"),
		ansiWhite:   css("#bac2de"), // subtext1

		ansiBrightBlack:   css("#585b70"), // surface2
		ansiBrightRed:     css("#f38ba8"),
		ansiBrightGreen:   css("#a6e3a1"),
		ansiBrightYellow:  css("#f9e2af"),
		ansiBrightBlue:    css("#89b4fa"),
		ansiBrightMagenta: css("#f5c2e7"),
		ansiBrightCyan:    css("#94e2d5"),
		ansiBrightWhite:   css("#cdd6f4"), // text
	})
}

// Gruvbox Dark palette. https://github.com/morhetz/gruvbox
var (
	gvPrimary = css("#fe8019") // bright orange
	gvSuccess = css("#b8bb26") // bright green
	gvWarning = css("#fabd2f") // bright yellow
	gvError   = css("#fb4934") // bright red
	gvBg      = css("#282828") // bg0
)

// GruvboxDark returns the Gruvbox dark theme.
func GruvboxDark() Styles {
	return quickStyle(quickStyleOpts{
		primary:   gvPrimary,
		secondary: gvWarning,
		accent:    css("#8ec07c"), // bright aqua
		keyword:   css("#d3869b"), // bright purple

		fgBase:       css("#ebdbb2"), // fg1
		fgMoreSubtle: css("#928374"), // gray
		fgSubtle:     css("#bdae93"), // fg3
		fgMostSubtle: css("#7c6f64"), // bg4

		onPrimary: css("#1d2021"), // bg0_h

		bgBase:         gvBg,
		bgLeastVisible: css("#1d2021"), // bg0_h
		bgLessVisible:  css("#32302f"), // bg0_s
		bgMostVisible:  css("#504945"), // bg2

		separator: css("#3c3836"), // bg1

		destructive:       gvError,
		error:             css("#cc241d"), // neutral red
		warningSubtle:     css("#d79921"), // neutral yellow
		warning:           gvWarning,
		attention:         gvPrimary,
		busy:              gvSuccess,
		info:              css("#83a598"), // bright blue
		infoMoreSubtle:    css("#458588"), // neutral blue
		infoMostSubtle:    css("#076678"), // faded blue
		success:           gvSuccess,
		successMoreSubtle: css("#8ec07c"), // bright aqua
		successMostSubtle: css("#98971a"), // neutral green

		ansiBlack:   css("#3c3836"), // bg1
		ansiRed:     css("#fb4934"),
		ansiGreen:   css("#b8bb26"),
		ansiYellow:  css("#fabd2f"),
		ansiBlue:    css("#83a598"),
		ansiMagenta: css("#d3869b"),
		ansiCyan:    css("#8ec07c"),
		ansiWhite:   css("#bdae93"), // fg3

		ansiBrightBlack:   css("#928374"), // gray
		ansiBrightRed:     css("#fb4934"),
		ansiBrightGreen:   css("#b8bb26"),
		ansiBrightYellow:  css("#fabd2f"),
		ansiBrightBlue:    css("#83a598"),
		ansiBrightMagenta: css("#d3869b"),
		ansiBrightCyan:    css("#8ec07c"),
		ansiBrightWhite:   css("#fbf1c7"), // fg0
	})
}

// Tokyo Night palette. https://github.com/folke/tokyonight.nvim
var (
	tnPrimary = css("#7aa2f7") // blue
	tnSuccess = css("#9ece6a") // green
	tnWarning = css("#e0af68") // yellow
	tnError   = css("#f7768e") // red
	tnBg      = css("#1a1b26") // bg
)

// TokyoNight returns the Tokyo Night dark theme.
func TokyoNight() Styles {
	return quickStyle(quickStyleOpts{
		primary:   tnPrimary,
		secondary: css("#bb9af7"), // purple
		accent:    css("#73daca"), // green1
		keyword:   css("#ff007c"), // pink

		fgBase:       css("#c0caf5"), // fg
		fgMoreSubtle: css("#565f89"), // comment
		fgSubtle:     css("#a9b1d6"), // fg_dark
		fgMostSubtle: css("#414868"), // dark3

		onPrimary: css("#16161e"), // bg_dark

		bgBase:         tnBg,
		bgLeastVisible: css("#16161e"), // bg_dark
		bgLessVisible:  css("#1f2335"), // bg_highlight
		bgMostVisible:  css("#292e42"), // terminal black

		separator: css("#292e42"),

		destructive:       tnError,
		error:             css("#db4b4b"), // red1
		warningSubtle:     css("#ff9e64"), // orange
		warning:           tnWarning,
		attention:         css("#ff9e64"), // orange
		busy:              tnWarning,
		info:              css("#7dcfff"), // cyan
		infoMoreSubtle:    css("#89ddff"), // blue5
		infoMostSubtle:    css("#0db9d7"), // blue2
		success:           tnSuccess,
		successMoreSubtle: css("#73daca"), // green1
		successMostSubtle: css("#41a6b5"), // green2

		ansiBlack:   css("#414868"), // dark3
		ansiRed:     css("#f7768e"),
		ansiGreen:   css("#9ece6a"),
		ansiYellow:  css("#e0af68"),
		ansiBlue:    css("#7aa2f7"),
		ansiMagenta: css("#bb9af7"),
		ansiCyan:    css("#7dcfff"),
		ansiWhite:   css("#a9b1d6"), // fg_dark

		ansiBrightBlack:   css("#565f89"), // comment
		ansiBrightRed:     css("#ff007c"), // pink
		ansiBrightGreen:   css("#73daca"),
		ansiBrightYellow:  css("#e0af68"),
		ansiBrightBlue:    css("#7aa2f7"),
		ansiBrightMagenta: css("#bb9af7"),
		ansiBrightCyan:    css("#7dcfff"),
		ansiBrightWhite:   css("#c0caf5"), // fg
	})
}

// Nord palette. https://www.nordtheme.com/docs/colors-and-palettes
var (
	ndPrimary = css("#88c0d0") // nord8
	ndSuccess = css("#a3be8c") // nord14
	ndWarning = css("#ebcb8b") // nord13
	ndError   = css("#bf616a") // nord11
	ndBg      = css("#2e3440") // nord0
)

// Nord returns the Nord dark theme.
func Nord() Styles {
	return quickStyle(quickStyleOpts{
		primary:   ndPrimary,
		secondary: css("#81a1c1"), // nord9
		accent:    ndSuccess,
		keyword:   css("#b48ead"), // nord15

		fgBase:       css("#d8dee9"), // nord4
		fgMoreSubtle: css("#7b88a3"),
		fgSubtle:     css("#a5b0c7"),
		fgMostSubtle: css("#4c566a"), // nord3

		onPrimary: ndBg,

		bgBase:         ndBg,
		bgLeastVisible: css("#2b303b"),
		bgLessVisible:  css("#3b4252"), // nord1
		bgMostVisible:  css("#434c5e"), // nord2

		separator: css("#3b4252"), // nord1

		destructive:       css("#d08770"), // nord12
		error:             ndError,
		warningSubtle:     css("#d8b96a"),
		warning:           ndWarning,
		attention:         css("#d08770"), // nord12
		busy:              css("#8fbcbb"), // nord7
		info:              css("#81a1c1"), // nord9
		infoMoreSubtle:    ndPrimary,
		infoMostSubtle:    css("#5e81ac"), // nord10
		success:           ndSuccess,
		successMoreSubtle: css("#8fbcbb"), // nord7
		successMostSubtle: css("#7b9663"),

		ansiBlack:   css("#3b4252"), // nord1
		ansiRed:     css("#bf616a"),
		ansiGreen:   css("#a3be8c"),
		ansiYellow:  css("#ebcb8b"),
		ansiBlue:    css("#81a1c1"),
		ansiMagenta: css("#b48ead"),
		ansiCyan:    css("#8fbcbb"),
		ansiWhite:   css("#d8dee9"), // nord4

		ansiBrightBlack:   css("#4c566a"), // nord3
		ansiBrightRed:     css("#bf616a"),
		ansiBrightGreen:   css("#a3be8c"),
		ansiBrightYellow:  css("#ebcb8b"),
		ansiBrightBlue:    css("#81a1c1"),
		ansiBrightMagenta: css("#b48ead"),
		ansiBrightCyan:    css("#8fbcbb"),
		ansiBrightWhite:   css("#eceff4"), // nord6
	})
}

// Dracula palette. https://draculatheme.com
var (
	dcPrimary = css("#bd93f9") // purple
	dcSuccess = css("#50fa7b") // green
	dcWarning = css("#f1fa8c") // yellow
	dcError   = css("#ff5555") // red
	dcBg      = css("#282a36") // background
)

// Dracula returns the Dracula dark theme.
func Dracula() Styles {
	return quickStyle(quickStyleOpts{
		primary:   dcPrimary,
		secondary: css("#ff79c6"), // pink
		accent:    dcSuccess,
		keyword:   css("#ff79c6"), // pink

		fgBase:       css("#f8f8f2"), // foreground
		fgMoreSubtle: css("#8f96b8"),
		fgSubtle:     css("#6272a4"), // comment
		fgMostSubtle: css("#44475a"), // selection

		onPrimary: dcBg,

		bgBase:         dcBg,
		bgLeastVisible: css("#21222c"), // background darker
		bgLessVisible:  css("#343746"),
		bgMostVisible:  css("#44475a"), // selection

		separator: css("#44475a"), // selection

		destructive:       dcError,
		error:             dcError,
		warningSubtle:     css("#ffb86c"), // orange
		warning:           dcWarning,
		attention:         css("#ffb86c"), // orange
		busy:              dcWarning,
		info:              css("#8be9fd"), // cyan
		infoMoreSubtle:    css("#a0e9f8"),
		infoMostSubtle:    css("#6272a4"), // comment
		success:           dcSuccess,
		successMoreSubtle: css("#87f5a1"),
		successMostSubtle: css("#2fbf5f"),

		ansiBlack:   css("#21222c"),
		ansiRed:     css("#ff5555"),
		ansiGreen:   css("#50fa7b"),
		ansiYellow:  css("#f1fa8c"),
		ansiBlue:    css("#bd93f9"),
		ansiMagenta: css("#ff79c6"),
		ansiCyan:    css("#8be9fd"),
		ansiWhite:   css("#f8f8f2"),

		ansiBrightBlack:   css("#44475a"), // selection
		ansiBrightRed:     css("#ff6e6e"),
		ansiBrightGreen:   css("#69ff94"),
		ansiBrightYellow:  css("#ffffa5"),
		ansiBrightBlue:    css("#d6acff"),
		ansiBrightMagenta: css("#ff92df"),
		ansiBrightCyan:    css("#a4ffff"),
		ansiBrightWhite:   css("#ffffff"),
	})
}

// One Dark palette. https://github.com/atom/one-dark-syntax
var (
	odPrimary = css("#61afef") // blue
	odSuccess = css("#98c379") // green
	odWarning = css("#e5c07b") // yellow
	odError   = css("#e06c75") // red
	odBg      = css("#282c34") // background
)

// OneDark returns the One Dark theme.
func OneDark() Styles {
	return quickStyle(quickStyleOpts{
		primary:   odPrimary,
		secondary: css("#c678dd"), // purple
		accent:    odSuccess,
		keyword:   css("#c678dd"), // purple

		fgBase:       css("#d7dae0"),
		fgMoreSubtle: css("#7f8798"),
		fgSubtle:     css("#abb2bf"), // foreground
		fgMostSubtle: css("#5c6370"), // comment

		onPrimary: odBg,

		bgBase:         odBg,
		bgLeastVisible: css("#21252b"),
		bgLessVisible:  css("#323842"),
		bgMostVisible:  css("#3e4451"), // selection

		separator: css("#3e4451"), // selection

		destructive:       odError,
		error:             css("#be5046"), // deep red
		warningSubtle:     css("#d19a66"), // orange
		warning:           odWarning,
		attention:         css("#d19a66"), // orange
		busy:              odWarning,
		info:              css("#56b6c2"), // cyan
		infoMoreSubtle:    css("#74c3d1"),
		infoMostSubtle:    css("#3f96ad"),
		success:           odSuccess,
		successMoreSubtle: css("#b3e88c"),
		successMostSubtle: css("#7ba05b"),

		ansiBlack:   css("#3e4451"), // selection
		ansiRed:     css("#e06c75"),
		ansiGreen:   css("#98c379"),
		ansiYellow:  css("#e5c07b"),
		ansiBlue:    css("#61afef"),
		ansiMagenta: css("#c678dd"),
		ansiCyan:    css("#56b6c2"),
		ansiWhite:   css("#abb2bf"), // foreground

		ansiBrightBlack:   css("#5c6370"), // comment
		ansiBrightRed:     css("#e06c75"),
		ansiBrightGreen:   css("#98c379"),
		ansiBrightYellow:  css("#e5c07b"),
		ansiBrightBlue:    css("#61afef"),
		ansiBrightMagenta: css("#c678dd"),
		ansiBrightCyan:    css("#56b6c2"),
		ansiBrightWhite:   css("#d7dae0"),
	})
}

// Ayu Dark palette. https://github.com/ayu-theme/ayu-colors
var (
	ayPrimary = css("#ff8f40") // accent orange
	aySuccess = css("#aad94c") // string green
	ayWarning = css("#e6b450") // yellow
	ayError   = css("#f07178") // tag red
	ayBg      = css("#0b0e14") // background
)

// AyuDark returns the Ayu dark theme.
func AyuDark() Styles {
	return quickStyle(quickStyleOpts{
		primary:   ayPrimary,
		secondary: css("#39bae6"), // tag cyan
		accent:    aySuccess,
		keyword:   css("#d2a6ff"), // constant purple

		fgBase:       css("#bfbdb6"), // foreground
		fgMoreSubtle: css("#6c7380"),
		fgSubtle:     css("#8a9199"),
		fgMostSubtle: css("#464c56"),

		onPrimary: ayBg,

		bgBase:         ayBg,
		bgLeastVisible: css("#070a0e"),
		bgLessVisible:  css("#11151c"), // line
		bgMostVisible:  css("#1f2430"), // selection

		separator: css("#151a23"),

		destructive:       css("#f29668"), // operator orange
		error:             ayError,
		warningSubtle:     css("#f2d49b"),
		warning:           ayWarning,
		attention:         ayPrimary,
		busy:              aySuccess,
		info:              css("#39bae6"), // tag cyan
		infoMoreSubtle:    css("#59c2ff"), // light blue
		infoMostSubtle:    css("#2f7fa8"),
		success:           aySuccess,
		successMoreSubtle: css("#c2f78c"),
		successMostSubtle: css("#7cb335"),

		ansiBlack:   css("#11151c"), // line
		ansiRed:     css("#f07178"),
		ansiGreen:   css("#aad94c"),
		ansiYellow:  css("#e6b450"),
		ansiBlue:    css("#59c2ff"),
		ansiMagenta: css("#d2a6ff"),
		ansiCyan:    css("#95e6cb"),
		ansiWhite:   css("#bfbdb6"), // foreground

		ansiBrightBlack:   css("#1f2430"), // selection
		ansiBrightRed:     css("#ff3333"), // error
		ansiBrightGreen:   css("#aad94c"),
		ansiBrightYellow:  css("#ffb454"), // func
		ansiBrightBlue:    css("#73b8ff"), // special
		ansiBrightMagenta: css("#d2a6ff"),
		ansiBrightCyan:    css("#95e6cb"),
		ansiBrightWhite:   css("#f8f8f2"),
	})
}

// Vesper palette. https://github.com/rauno56/vesper
var (
	vsPrimary = css("#ffc799") // accent peach
	vsSuccess = css("#99ffe4") // green
	vsWarning = css("#ffe3a5") // pale yellow
	vsError   = css("#ff8080") // red
	vsBg      = css("#101010") // background
)

// Vesper returns the Vesper minimal dark theme.
func Vesper() Styles {
	return quickStyle(quickStyleOpts{
		primary:   vsPrimary,
		secondary: css("#cc99ff"), // purple
		accent:    vsSuccess,
		keyword:   css("#91b3e6"), // blue

		fgBase:       css("#cdcdcd"), // foreground
		fgMoreSubtle: css("#6f6f6f"), // comment
		fgSubtle:     css("#9a9a9a"),
		fgMostSubtle: css("#505050"),

		onPrimary: vsBg,

		bgBase:         vsBg,
		bgLeastVisible: css("#0b0b0b"),
		bgLessVisible:  css("#161616"),
		bgMostVisible:  css("#222222"), // selection

		separator: css("#181818"),

		destructive:       vsError,
		error:             vsError,
		warningSubtle:     vsWarning,
		warning:           vsPrimary,
		attention:         vsPrimary,
		busy:              vsWarning,
		info:              css("#91b3e6"), // blue
		infoMoreSubtle:    css("#b3c9f2"),
		infoMostSubtle:    css("#6f8cb8"),
		success:           vsSuccess,
		successMoreSubtle: css("#b3ffe9"),
		successMostSubtle: css("#6fbfab"),

		ansiBlack:   css("#101010"),
		ansiRed:     css("#ff8080"),
		ansiGreen:   css("#99ffe4"),
		ansiYellow:  css("#ffe3a5"),
		ansiBlue:    css("#91b3e6"),
		ansiMagenta: css("#cc99ff"),
		ansiCyan:    css("#99ffe4"),
		ansiWhite:   css("#cdcdcd"), // foreground

		ansiBrightBlack:   css("#222222"), // selection
		ansiBrightRed:     css("#ff8080"),
		ansiBrightGreen:   css("#99ffe4"),
		ansiBrightYellow:  css("#ffe3a5"),
		ansiBrightBlue:    css("#91b3e6"),
		ansiBrightMagenta: css("#cc99ff"),
		ansiBrightCyan:    css("#99ffe4"),
		ansiBrightWhite:   css("#ffffff"),
	})
}

// Catppuccin Latte palette (light). https://github.com/catppuccin/catppuccin
var (
	cpLattePrimary = css("#8839ef") // mauve
	cpLatteSuccess = css("#40a02b") // green
	cpLatteWarning = css("#df8e1d") // yellow
	cpLatteError   = css("#d20f39") // red
	cpLatteBg      = css("#eff1f5") // base
)

// CatppuccinLatte returns the Catppuccin Latte light theme.
func CatppuccinLatte() Styles {
	s := quickStyle(quickStyleOpts{
		primary:   cpLattePrimary,
		secondary: css("#1e66f5"), // blue
		accent:    css("#179299"), // teal
		keyword:   css("#ea76cb"), // pink

		fgBase:       css("#4c4f69"), // text
		fgMoreSubtle: css("#6c6f85"), // subtext0
		fgSubtle:     css("#5c5f77"), // subtext1
		fgMostSubtle: css("#9ca0b0"), // overlay0

		onPrimary: cpLatteBg,

		// On a light background, visibility comes from darker shades, so
		// the bg ramp runs from closest-to-base (least visible) to the
		// darkest surface (most visible).
		bgBase:         cpLatteBg,
		bgLeastVisible: css("#e6e9ef"), // mantle
		bgLessVisible:  css("#dce0e8"), // crust
		bgMostVisible:  css("#ccd0da"), // surface0

		separator: css("#ccd0da"), // surface0

		destructive:       css("#fe640b"), // peach
		error:             cpLatteError,
		warningSubtle:     css("#fe640b"), // peach
		warning:           cpLatteWarning,
		attention:         css("#fe640b"), // peach
		busy:              css("#879100"),
		info:              css("#04a5e5"), // sky
		infoMoreSubtle:    css("#209fb5"), // sapphire
		infoMostSubtle:    css("#7287fd"), // lavender
		success:           cpLatteSuccess,
		successMoreSubtle: css("#179299"), // teal
		successMostSubtle: css("#568b3c"),

		// The ANSI palette maps to the darker Latte tones so raw terminal
		// output stays legible on the light background.
		ansiBlack:   css("#5c5f77"), // subtext1
		ansiRed:     css("#d20f39"),
		ansiGreen:   css("#40a02b"),
		ansiYellow:  css("#df8e1d"),
		ansiBlue:    css("#1e66f5"),
		ansiMagenta: css("#ea76cb"),
		ansiCyan:    css("#179299"),
		ansiWhite:   css("#9ca0b0"), // overlay0

		ansiBrightBlack:   css("#6c6f85"), // subtext0
		ansiBrightRed:     css("#e64553"), // maroon
		ansiBrightGreen:   css("#40a02b"),
		ansiBrightYellow:  css("#df8e1d"),
		ansiBrightBlue:    css("#1e66f5"),
		ansiBrightMagenta: css("#ea76cb"),
		ansiBrightCyan:    css("#179299"),
		ansiBrightWhite:   css("#7c7f93"), // overlay2
	})

	// Bang > prompt overrides - keep the bang affordances legible on the
	// light background.
	s.Editor.PromptBangDotsBlurred = s.Editor.PromptBangDotsBlurred.
		Foreground(css("#9ca0b0")) // overlay0

	return s
}

// Solarized Light palette. https://ethanschoonover.com/solarized/
var (
	slPrimary = css("#268bd2") // blue
	slSuccess = css("#859900") // green
	slWarning = css("#b58900") // yellow
	slError   = css("#dc322f") // red
	slBg      = css("#fdf6e3") // base3
)

// SolarizedLight returns the Solarized light theme.
func SolarizedLight() Styles {
	return quickStyle(quickStyleOpts{
		primary:   slPrimary,
		secondary: css("#d33682"), // magenta
		accent:    css("#2aa198"), // cyan
		keyword:   css("#6c71c4"), // violet

		fgBase:       css("#586e75"), // base01
		fgMoreSubtle: css("#839496"), // base0
		fgSubtle:     css("#657b83"), // base00
		fgMostSubtle: css("#93a1a1"), // base1

		onPrimary: slBg,

		bgBase:         slBg,
		bgLeastVisible: css("#eee8d5"), // base2
		bgLessVisible:  css("#e4dcc8"),
		bgMostVisible:  css("#d6cdb4"),

		separator: css("#eee8d5"), // base2

		destructive:       css("#cb4b16"), // orange
		error:             slError,
		warningSubtle:     css("#cb4b16"), // orange
		warning:           slWarning,
		attention:         css("#cb4b16"), // orange
		busy:              slSuccess,
		info:              css("#2aa198"), // cyan
		infoMoreSubtle:    slPrimary,
		infoMostSubtle:    css("#6c71c4"), // violet
		success:           slSuccess,
		successMoreSubtle: css("#2aa198"), // cyan
		successMostSubtle: css("#5c7a00"),

		ansiBlack:   css("#657b83"), // base00
		ansiRed:     css("#dc322f"),
		ansiGreen:   css("#859900"),
		ansiYellow:  css("#b58900"),
		ansiBlue:    css("#268bd2"),
		ansiMagenta: css("#d33682"),
		ansiCyan:    css("#2aa198"),
		ansiWhite:   css("#93a1a1"), // base1

		ansiBrightBlack:   css("#586e75"), // base01
		ansiBrightRed:     css("#cb4b16"), // orange
		ansiBrightGreen:   css("#859900"),
		ansiBrightYellow:  css("#b58900"),
		ansiBrightBlue:    css("#268bd2"),
		ansiBrightMagenta: css("#d33682"),
		ansiBrightCyan:    css("#2aa198"),
		ansiBrightWhite:   css("#839496"), // base0
	})
}
