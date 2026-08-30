package tools

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContainsCommandChaining(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"plain ls", "ls -la", false},
		{"plain echo", "echo hello world", false},
		{"plain pwd", "pwd", false},
		{"plain git status", "git status", false},
		{"ls with redirect", "ls > /tmp/out", true},
		{"ls with append redirect", "ls >> /tmp/out", true},
		{"ls with stdin redirect", "wc -l < /etc/passwd", true},
		{"ls with pipe", "ls | grep foo", true},
		{"ls with double ampersand", "ls && echo done", true},
		{"ls with semicolon", "ls; echo done", true},
		{"ls with pipe pipe", "ls || echo fail", true},
		{"ls with backticks", "ls `echo foo`", true},
		{"ls with subshell", "ls $(echo foo)", true},
		{"ls with background ampersand", "ls & echo done", true},
		{"rm -rf with && ls (rm first)", "rm -rf / && ls", true},
		{"redirect with ampersand gt", "ls &> /dev/null", true},
		{"redirect with gt ampersand", "ls >& /dev/null", true},
		{"newline separated commands", "ls\nrm -rf /tmp/x", true},
		{"carriage return separated", "ls\r\nrm -rf /tmp/x", true},
		{"kill with pipe", "kill 1234 | echo foo", true},
		{"git log", "git log --oneline", false},
		{"git log with pipe", "git log | head", true},
		{"empty string", "", false},
		{"dollar sign in argument", "echo $HOME", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := containsCommandChaining(tt.input)
			assert.Equal(t, tt.expected, got, "containsCommandChaining(%q)", tt.input)
		})
	}
}

// TestIsSafeReadOnlyCommand guards the bash tool's permission bypass:
// anything that can spawn a command, mutate state, or redirect output
// must fall through to the permission flow.
func TestIsSafeReadOnlyCommand(t *testing.T) {
	t.Parallel()

	safe := []string{
		"ls -la",
		"ls",
		"echo hello world",
		"pwd",
		"git status",
		"git log --oneline",
		"git diff HEAD~1",
		"git blame main.go",
		"git branch --list",
		"git branch --list -a",
		"git branch --show-current",
		"git tag --list",
		"git tag --list -n 5",
		"git remote get-url origin",
		"git config --get user.email",
		"git config --list",
		"ps aux",
		"df -h",
	}

	for _, cmd := range safe {
		t.Run("safe: "+cmd, func(t *testing.T) {
			t.Parallel()
			assert.True(t, isSafeReadOnlyCommand(cmd), "expected %q to be safe", cmd)
		})
	}

	unsafe := []string{
		// Spawns another command.
		"timeout 5 ls",
		"nohup sleep 100",
		"nice rm -rf /tmp/x",
		"time ls",
		"env FOO=bar ls",
		// Mutates processes.
		"kill 1234",
		"killall nginx",
		// Redirects, even from an otherwise safe prefix.
		"echo x > /tmp/file",
		"echo x >> ~/.zshrc",
		"ls > /tmp/out",
		// Chained or substituted commands.
		"ls && rm -rf /tmp/x",
		"ls & rm -rf /tmp/x",
		"echo `id`",
		"echo $(id)",
		"ls\nrm -rf /tmp/x",
		// Git mutations must not ride on read-only prefixes.
		"git branch -D main",
		"git branch newbranch",
		"git tag -d v1",
		"git tag v1",
		"git remote -v add evil https://example.com/x",
		"git remote remove origin",
		"git push origin main",
		// Plainly unsafe commands.
		"rm -rf /",
		"curl https://example.com",
		"cat /etc/passwd",
		"uppercase LS",
	}

	for _, cmd := range unsafe {
		t.Run("unsafe: "+cmd, func(t *testing.T) {
			t.Parallel()
			assert.False(t, isSafeReadOnlyCommand(cmd), "expected %q to be unsafe", cmd)
		})
	}
}
