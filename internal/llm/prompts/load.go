package prompts

import (
	"bytes"
	_ "embed"
	"regexp"
	"strings"
	"text/template"
)

// =============================================================================
// Prompt Injection Sanitization
// =============================================================================

// Common prompt injection patterns to detect and neutralize
var injectionPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)ignore\s+(all\s+)?(previous|prior|above)\s+(instructions?|prompts?|text)`),
	regexp.MustCompile(`(?i)disregard\s+(all\s+)?(previous|prior|above)`),
	regexp.MustCompile(`(?i)</?system>`),
	regexp.MustCompile(`(?i)</?assistant>`),
	regexp.MustCompile(`(?i)</?user>`),
	regexp.MustCompile(`(?i)</?human>`),
	regexp.MustCompile(`(?i)\[INST\]|\[/INST\]`),
	regexp.MustCompile(`(?i)###\s*(system|user|assistant|instruction)`),
}

// =============================================================================
// Secret Redaction
// =============================================================================

// secretPatterns detects common secret patterns in diffs
var secretPatterns = []*regexp.Regexp{
	// API Keys (generic patterns)
	regexp.MustCompile(`(?i)(api[_-]?key|apikey)\s*[:=]\s*["']?[A-Za-z0-9_\-]{20,}["']?`),
	regexp.MustCompile(`(?i)(secret[_-]?key|secretkey)\s*[:=]\s*["']?[A-Za-z0-9_\-]{20,}["']?`),
	regexp.MustCompile(`(?i)(access[_-]?token|accesstoken)\s*[:=]\s*["']?[A-Za-z0-9_\-]{20,}["']?`),

	// AWS keys
	regexp.MustCompile(`(?i)AKIA[0-9A-Z]{16}`),
	regexp.MustCompile(`(?i)(aws[_-]?secret[_-]?access[_-]?key)\s*[:=]\s*["']?[A-Za-z0-9/+=]{40}["']?`),

	// GitHub tokens
	regexp.MustCompile(`ghp_[A-Za-z0-9]{36}`),
	regexp.MustCompile(`gho_[A-Za-z0-9]{36}`),
	regexp.MustCompile(`ghu_[A-Za-z0-9]{36}`),
	regexp.MustCompile(`ghs_[A-Za-z0-9]{36}`),
	regexp.MustCompile(`ghr_[A-Za-z0-9]{36}`),

	// Generic tokens
	regexp.MustCompile(`(?i)(bearer|token)\s+[A-Za-z0-9_\-\.]{20,}`),

	// Passwords in config
	regexp.MustCompile(`(?i)(password|passwd|pwd)\s*[:=]\s*["'][^"']{8,}["']`),

	// Private keys
	regexp.MustCompile(`-----BEGIN\s+(RSA\s+)?PRIVATE\s+KEY-----`),
	regexp.MustCompile(`-----BEGIN\s+OPENSSH\s+PRIVATE\s+KEY-----`),

	// Database URLs with credentials
	regexp.MustCompile(`(?i)(postgres|mysql|mongodb)://[^:]+:[^@]+@`),
}

// RedactSecrets removes potential secrets from content before sending to LLM.
// This is applied to diffs and other code content.
func RedactSecrets(content string) string {
	result := content
	for _, pattern := range secretPatterns {
		result = pattern.ReplaceAllString(result, "[REDACTED]")
	}
	return result
}

// =============================================================================
// Input Sanitization
// =============================================================================

// sanitizeUserInput removes or neutralizes potential prompt injection patterns
// from user-provided content before including it in prompts.
func sanitizeUserInput(input string) string {
	result := input

	// Neutralize known injection patterns by adding a prefix that breaks them
	for _, pattern := range injectionPatterns {
		result = pattern.ReplaceAllStringFunc(result, func(match string) string {
			// Prefix with [USER_CONTENT] to make it clear this is user input
			return "[text: " + match + "]"
		})
	}

	// Remove null bytes and other control characters that could cause issues
	result = strings.Map(func(r rune) rune {
		if r == 0 || (r < 32 && r != '\n' && r != '\r' && r != '\t') {
			return -1 // Remove the character
		}
		return r
	}, result)

	return result
}

type PRCreateVars struct {
	Repo      string
	Base      string
	Head      string
	CommitLog string
	DiffStat  string
	UserNotes string
}

type PRReviewVars struct {
	Title                string
	Author               string
	Base                 string
	Head                 string
	DescriptionOrSummary string
	SelectedDiff         string
}

type PRSummaryVars struct {
	Title       string
	Description string
	CommitLog   string
	DiffStat    string
}

// PRReviewCommentVars holds the variables for generating a review comment.
type PRReviewCommentVars struct {
	ReviewType      string // "approve", "request_changes", or "comment"
	PRTitle         string
	PRNumber        int
	Diff            string // Truncated diff for context
	ExistingComment string // User's draft to refine, if any
}

//go:embed pr_create_system.txt
var prCreateSystem string

//go:embed pr_create_user.tmpl
var prCreateUser string

//go:embed pr_review_system.txt
var prReviewSystem string

//go:embed pr_review_user.tmpl
var prReviewUser string

//go:embed pr_summary_system.txt
var prSummarySystem string

//go:embed pr_summary_user.tmpl
var prSummaryUser string

//go:embed pr_review_comment_system.txt
var prReviewCommentSystem string

//go:embed pr_review_comment_user.tmpl
var prReviewCommentUser string

func RenderPRCreate(v PRCreateVars) (system string, user string, err error) {
	// Sanitize user-provided content and redact secrets from diffs
	sanitized := PRCreateVars{
		Repo:      sanitizeUserInput(v.Repo),
		Base:      sanitizeUserInput(v.Base),
		Head:      sanitizeUserInput(v.Head),
		CommitLog: sanitizeUserInput(v.CommitLog),
		DiffStat:  RedactSecrets(sanitizeUserInput(v.DiffStat)),
		UserNotes: sanitizeUserInput(v.UserNotes),
	}
	u, err := render(prCreateUser, sanitized)
	return prCreateSystem, u, err
}

func RenderPRReview(v PRReviewVars) (system string, user string, err error) {
	// Sanitize user-provided content and redact secrets from diffs
	sanitized := PRReviewVars{
		Title:                sanitizeUserInput(v.Title),
		Author:               sanitizeUserInput(v.Author),
		Base:                 sanitizeUserInput(v.Base),
		Head:                 sanitizeUserInput(v.Head),
		DescriptionOrSummary: sanitizeUserInput(v.DescriptionOrSummary),
		SelectedDiff:         RedactSecrets(sanitizeUserInput(v.SelectedDiff)),
	}
	u, err := render(prReviewUser, sanitized)
	return prReviewSystem, u, err
}

func RenderPRSummary(v PRSummaryVars) (system string, user string, err error) {
	// Sanitize user-provided content and redact secrets from diffs
	sanitized := PRSummaryVars{
		Title:       sanitizeUserInput(v.Title),
		Description: sanitizeUserInput(v.Description),
		CommitLog:   sanitizeUserInput(v.CommitLog),
		DiffStat:    RedactSecrets(sanitizeUserInput(v.DiffStat)),
	}
	u, err := render(prSummaryUser, sanitized)
	return prSummarySystem, u, err
}

// RenderPRReviewComment renders a prompt for generating a PR review comment.
func RenderPRReviewComment(v PRReviewCommentVars) (system string, user string, err error) {
	// Sanitize user-provided content and redact secrets from diffs
	sanitized := PRReviewCommentVars{
		ReviewType:      v.ReviewType, // Enum value, not user input
		PRTitle:         sanitizeUserInput(v.PRTitle),
		PRNumber:        v.PRNumber, // Numeric, not user input
		Diff:            RedactSecrets(sanitizeUserInput(v.Diff)),
		ExistingComment: sanitizeUserInput(v.ExistingComment),
	}
	u, err := render(prReviewCommentUser, sanitized)
	return prReviewCommentSystem, u, err
}

func render(tmpl string, data any) (string, error) {
	t, err := template.New("p").Option("missingkey=zero").Parse(tmpl)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}
