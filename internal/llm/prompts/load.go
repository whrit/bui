package prompts

import (
	"bytes"
	_ "embed"
	"text/template"
)

type PRCreateVars struct {
	Repo      string
	Base      string
	Head      string
	CommitLog string
	DiffStat  string
	UserNotes string
}

type PRReviewVars struct {
	Title              string
	Author             string
	Base               string
	Head               string
	DescriptionOrSummary string
	SelectedDiff       string
}

type PRSummaryVars struct {
	Title       string
	Description string
	CommitLog   string
	DiffStat    string
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

func RenderPRCreate(v PRCreateVars) (system string, user string, err error) {
	u, err := render(prCreateUser, v)
	return prCreateSystem, u, err
}

func RenderPRReview(v PRReviewVars) (system string, user string, err error) {
	u, err := render(prReviewUser, v)
	return prReviewSystem, u, err
}

func RenderPRSummary(v PRSummaryVars) (system string, user string, err error) {
	u, err := render(prSummaryUser, v)
	return prSummarySystem, u, err
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
