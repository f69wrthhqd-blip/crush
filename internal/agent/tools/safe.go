package tools

import (
	"runtime"
	"slices"
	"strings"
)

// safeCommands are command prefixes the bash tool treats as read-only:
// they run without asking the user for permission. Anything that can
// spawn another command, mutate state, or be combined with trailing
// subcommands must not be on this list.
var safeCommands = []string{
	// Bash builtins and core utils
	"cal",
	"date",
	"df",
	"du",
	"echo",
	"free",
	"groups",
	"hostname",
	"id",
	"ls",
	"printenv",
	"ps",
	"pwd",
	"set",
	"top",
	"type",
	"uname",
	"unset",
	"uptime",
	"whatis",
	"whereis",
	"which",
	"whoami",

	// Git (read-only forms only; mutations like `git branch -D` or
	// `git tag -d` must hit the permission flow)
	"git blame",
	"git branch --list",
	"git branch --show-current",
	"git config --get",
	"git config --list",
	"git describe",
	"git diff",
	"git grep",
	"git log",
	"git ls-files",
	"git ls-remote",
	"git remote get-url",
	"git rev-parse",
	"git shortlog",
	"git show",
	"git status",
	"git tag --list",
}

// chainingMetacharacters are shell metacharacters that let one command
// string run extra commands or redirect I/O: chaining ( ; | & $( ) and
// backticks, including newlines) and redirection ( < and > ).
var chainingMetacharacters = []string{
	";",
	"|",
	"&",
	"$(",
	"`",
	">",
	"<",
	"\n",
	"\r",
}

// bannedGitFlags are git flags that turn read-only subcommands into program
// execution or file writes. --ext-diff and --textconv run external drivers
// from git config or .gitattributes, --open-files-in-pager starts a pager
// process, --upload-pack, --receive-pack, and --exec run a custom helper
// for remote queries, and --output writes to a file without ever using a
// shell redirection metacharacter.
var bannedGitFlags = []string{
	"--ext-diff",
	"--textconv",
	"--output",
	"--open-files-in-pager",
	"--upload-pack",
	"--receive-pack",
	"--exec",
}

// dangerousArgPrefixes maps an allowlisted command prefix to argument
// prefixes that must never reach it. They are matched case-sensitively
// against each whitespace-separated argument, so combined short forms such
// as -O<pager> or -u<exec> are caught too. git grep -O opens a pager
// (distinct from the harmless lowercase -o), and ls-remote -u<exec> runs a
// custom upload-pack helper; date -s sets the system clock.
var dangerousArgPrefixes = map[string][]string{
	"git grep":      {"-O"},
	"git ls-remote": {"-u"},
	"date":          {"-s", "--set"},
}

// dangerousArgPrefixesFold is like dangerousArgPrefixes but matched
// case-insensitively, for commands with case-insensitive flags. Windows
// ipconfig can drop or renew network interfaces and flush DNS caches.
var dangerousArgPrefixesFold = map[string][]string{
	"ipconfig": {
		"/release",
		"/renew",
		"/flushdns",
		"/registerdns",
		"/setclassid",
		"/setdnsservers",
	},
}

// noFlagArguments lists commands whose bare, non-flag arguments have side
// effects: a positional argument to hostname sets the system hostname.
var noFlagArguments = []string{
	"hostname",
}

// containsCommandChaining reports whether s contains shell metacharacters
// that enable command chaining, substitution, or redirection.
func containsCommandChaining(s string) bool {
	return slices.ContainsFunc(chainingMetacharacters, func(c string) bool {
		return strings.Contains(s, c)
	})
}

// isSafeReadOnlyCommand reports whether the bash tool may run cmd without
// asking the user for permission. The command must both match a read-only
// prefix (next character being end, space, or a dash), be free of chaining
// or redirection metacharacters, and carry no argument-level side effects
// for the matched command (see bannedGitFlags and dangerousArgPrefixes).
//
// Flags of commands without explicit rules are not audited, so an
// undocumented harmful flag would still be missed: audit any addition to
// safeCommands against its full flag surface.
func isSafeReadOnlyCommand(cmd string) bool {
	if containsCommandChaining(cmd) {
		return false
	}
	cmdLower := strings.ToLower(cmd)
	args := strings.Fields(cmd)
	if len(args) > 0 {
		args = args[1:]
	}
	for _, safe := range safeCommands {
		if !strings.HasPrefix(cmdLower, safe) {
			continue
		}
		if len(cmdLower) == len(safe) || cmdLower[len(safe)] == ' ' || cmdLower[len(safe)] == '-' {
			if !hasDangerousArgs(safe, args) {
				return true
			}
		}
	}
	return false
}

// hasDangerousArgs reports whether args (the whitespace-split arguments of
// a command matching the allowlisted prefix safe) would give that command
// side effects beyond reading: running external programs, writing files,
// or changing system state.
func hasDangerousArgs(safe string, args []string) bool {
	if strings.HasPrefix(safe, "git") {
		for _, arg := range args {
			if slices.Contains(bannedGitFlags, arg) {
				return true
			}
			for _, banned := range bannedGitFlags {
				if strings.HasPrefix(arg, banned+"=") {
					return true
				}
			}
		}
	}
	for _, banned := range dangerousArgPrefixes[safe] {
		for _, arg := range args {
			if strings.HasPrefix(arg, banned) {
				return true
			}
		}
	}
	for _, banned := range dangerousArgPrefixesFold[safe] {
		for _, arg := range args {
			if strings.HasPrefix(strings.ToLower(arg), banned) {
				return true
			}
		}
	}
	if slices.Contains(noFlagArguments, safe) {
		for _, arg := range args {
			if !strings.HasPrefix(arg, "-") {
				return true
			}
		}
	}
	return false
}

func init() {
	if runtime.GOOS == "windows" {
		safeCommands = append(
			safeCommands,
			// Windows-specific commands
			"ipconfig",
			"nslookup",
			"ping",
			"systeminfo",
			"tasklist",
			"where",
		)
	}
}
