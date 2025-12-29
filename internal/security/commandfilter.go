package security

import (
	"regexp"
	"strings"
)

// DangerousCommand represents a blocked command pattern
type DangerousCommand struct {
	Pattern     *regexp.Regexp
	Description string
}

// CommandFilter provides command validation and filtering
type CommandFilter struct {
	blockedCommands []DangerousCommand
	allowedCommands []string
}

// NewCommandFilter creates a new command filter with default rules
func NewCommandFilter() *CommandFilter {
	cf := &CommandFilter{
		blockedCommands: []DangerousCommand{
			// File system destruction
			{regexp.MustCompile(`\brm\s+(-rf?|--recursive|--force)\s+/`), "Recursive delete from root"},
			{regexp.MustCompile(`\bdd\s+`), "Direct disk access (dd command)"},
			{regexp.MustCompile(`\bmkfs`), "Filesystem creation"},
			{regexp.MustCompile(`\bfdisk`), "Disk partitioning"},
			{regexp.MustCompile(`\bparted`), "Disk partitioning"},

			// System manipulation
			{regexp.MustCompile(`\breboot\b`), "System reboot"},
			{regexp.MustCompile(`\bshutdown\b`), "System shutdown"},
			{regexp.MustCompile(`\bhalt\b`), "System halt"},
			{regexp.MustCompile(`\bpoweroff\b`), "System poweroff"},
			{regexp.MustCompile(`\binit\s+[06]`), "System reboot/halt via init"},

			// Privilege escalation
			{regexp.MustCompile(`\bsudo\b`), "Sudo execution"},
			{regexp.MustCompile(`\bsu\s`), "Switch user"},

			// Network attacks
			{regexp.MustCompile(`\b(nmap|masscan)\b`), "Network scanning"},
			{regexp.MustCompile(`\bmetasploit\b`), "Metasploit framework"},

			// Process manipulation
			{regexp.MustCompile(`\bkill\s+-9\s+1\b`), "Kill init process"},
			{regexp.MustCompile(`\bkillall\s+-9`), "Force kill all processes"},

			// Fork bombs and resource exhaustion
			{regexp.MustCompile(`:\(\)\{.*:\|:.*\};:`), "Fork bomb"},
			{regexp.MustCompile(`\bwhile\s+true.*do.*done`), "Infinite loop"},

			// Escape attempts
			{regexp.MustCompile(`\bchroot\b`), "Chroot escape attempt"},
			{regexp.MustCompile(`\bdocker\s+(run|exec)\b`), "Docker escape attempt"},
			{regexp.MustCompile(`\bnsenter\b`), "Namespace escape"},

			// Suspicious encoding
			{regexp.MustCompile(`base64.*-d.*\|.*sh`), "Base64 decode and execute"},
			{regexp.MustCompile(`echo.*\|.*base64.*-d`), "Encoded command execution"},
		},

		// Allowed commands for CKS practice
		allowedCommands: []string{
			"kubectl", "k", "get", "describe", "logs", "exec", "apply", "create",
			"delete", "edit", "explain", "api-resources", "api-versions",
			"config", "cluster-info", "top", "drain", "cordon", "uncordon",
			"taint", "label", "annotate", "scale", "autoscale", "rollout",
			"set", "patch", "replace", "wait", "auth", "certificate",
			"ls", "cd", "pwd", "cat", "less", "more", "head", "tail",
			"grep", "find", "which", "echo", "printf", "wc", "sort",
			"awk", "sed", "cut", "tr", "uniq", "diff", "tee",
			"mkdir", "touch", "cp", "mv", "chmod", "chown",
			"curl", "wget", "ping", "netstat", "ss", "ip", "route",
			"ps", "top", "htop", "free", "df", "du", "uptime",
			"date", "cal", "history", "clear", "reset", "exit",
			"vim", "vi", "nano", "emacs",
			"git", "make", "gcc", "python", "python3", "node", "npm",
			"jq", "yq", "yaml", "json",
			// CKS-specific tools
			"falco", "trivy", "kube-bench", "kubesec", "opa", "conftest",
			"crictl", "ctr", "nerdctl",
			"openssl", "ssh-keygen", "gpg",
			"apparmor_parser", "aa-status", "seccomp",
		},
	}

	return cf
}

// ValidateCommand checks if a command is safe to execute
func (cf *CommandFilter) ValidateCommand(cmd string) (bool, string) {
	// Trim whitespace
	cmd = strings.TrimSpace(cmd)

	// Allow empty commands
	if cmd == "" {
		return true, ""
	}

	// Check against blocked patterns
	for _, blocked := range cf.blockedCommands {
		if blocked.Pattern.MatchString(cmd) {
			return false, "Blocked: " + blocked.Description
		}
	}

	// Additional heuristics for suspicious patterns
	if strings.Count(cmd, "|") > 5 {
		return false, "Blocked: Excessive command chaining"
	}

	if strings.Count(cmd, ";") > 5 {
		return false, "Blocked: Excessive command chaining"
	}

	if len(cmd) > 1000 {
		return false, "Blocked: Command too long"
	}

	// Check for suspicious character sequences
	suspiciousPatterns := []string{
		"../../..", // Directory traversal
		"/dev/null", // Redirecting to /dev/null (suspicious if many times)
		"&>/dev/null", // Background execution hiding output
	}

	suspiciousCount := 0
	for _, pattern := range suspiciousPatterns {
		if strings.Contains(cmd, pattern) {
			suspiciousCount++
		}
	}

	if suspiciousCount > 2 {
		return false, "Blocked: Suspicious pattern detected"
	}

	return true, ""
}

// SanitizeInput removes potentially dangerous characters
func (cf *CommandFilter) SanitizeInput(input string) string {
	// Remove null bytes (security)
	input = strings.ReplaceAll(input, "\x00", "")

	// Don't trim whitespace - breaks normal terminal input!
	// Spaces and newlines are necessary for shell commands

	return input
}

// IsCommandAllowed checks if command starts with an allowed command
// This is a soft check, not enforced, just for logging/metrics
func (cf *CommandFilter) IsCommandAllowed(cmd string) bool {
	cmd = strings.TrimSpace(cmd)
	if cmd == "" {
		return true
	}

	// Extract first word
	parts := strings.Fields(cmd)
	if len(parts) == 0 {
		return true
	}

	firstCmd := parts[0]

	// Check if it's in allowed list
	for _, allowed := range cf.allowedCommands {
		if firstCmd == allowed {
			return true
		}
	}

	// Also allow common bash built-ins
	builtins := []string{"cd", "pwd", "echo", "export", "source", "alias", "unalias"}
	for _, builtin := range builtins {
		if firstCmd == builtin {
			return true
		}
	}

	return false
}
