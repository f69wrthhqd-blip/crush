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

// containsCommandChaining reports whether s contains shell metacharacters
// that enable command chaining, substitution, or redirection.
func containsCommandChaining(s string) bool {
	return slices.ContainsFunc(chainingMetacharacters, func(c string) bool {
		return strings.Contains(s, c)
	})
}

// isSafeReadOnlyCommand reports whether the bash tool may run cmd without
// asking the user for permission. The command must both match a read-only
// prefix (next character being end, space, or a dash) and be free of
// chaining or redirection metacharacters.
func isSafeReadOnlyCommand(cmd string) bool {
	if containsCommandChaining(cmd) {
		return false
	}
	cmdLower := strings.ToLower(cmd)
	for _, safe := range safeCommands {
		if strings.HasPrefix(cmdLower, safe) {
			if len(cmdLower) == len(safe) || cmdLower[len(safe)] == ' ' || cmdLower[len(safe)] == '-' {
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
