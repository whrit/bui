package gh

import (
	"errors"
	"testing"
)

func TestGetChecks(t *testing.T) {
	tests := []struct {
		name       string
		prNumber   int
		mockOutput string
		mockErr    error
		wantCount  int
		wantErr    bool
	}{
		{
			name:     "get checks successfully - all passing",
			prNumber: 42,
			mockOutput: `[
				{
					"name": "build",
					"state": "completed",
					"conclusion": "success",
					"link": "https://github.com/owner/repo/actions/runs/123"
				},
				{
					"name": "test",
					"state": "completed",
					"conclusion": "success",
					"link": "https://github.com/owner/repo/actions/runs/124"
				},
				{
					"name": "lint",
					"state": "completed",
					"conclusion": "success",
					"link": "https://github.com/owner/repo/actions/runs/125"
				}
			]`,
			wantCount: 3,
			wantErr:   false,
		},
		{
			name:     "get checks with failures",
			prNumber: 43,
			mockOutput: `[
				{
					"name": "build",
					"state": "completed",
					"conclusion": "success",
					"link": "https://github.com/owner/repo/actions/runs/200"
				},
				{
					"name": "test",
					"state": "completed",
					"conclusion": "failure",
					"link": "https://github.com/owner/repo/actions/runs/201"
				}
			]`,
			wantCount: 2,
			wantErr:   false,
		},
		{
			name:     "get checks in progress",
			prNumber: 44,
			mockOutput: `[
				{
					"name": "build",
					"state": "in_progress",
					"conclusion": "",
					"link": "https://github.com/owner/repo/actions/runs/300"
				},
				{
					"name": "test",
					"state": "queued",
					"conclusion": "",
					"link": "https://github.com/owner/repo/actions/runs/301"
				}
			]`,
			wantCount: 2,
			wantErr:   false,
		},
		{
			name:       "no checks on PR",
			prNumber:   45,
			mockOutput: `[]`,
			wantCount:  0,
			wantErr:    false,
		},
		{
			name:       "PR not found",
			prNumber:   999,
			mockOutput: "",
			mockErr:    errors.New("GraphQL: Could not resolve to a PullRequest"),
			wantErr:    true,
		},
		{
			name:       "invalid JSON response",
			prNumber:   46,
			mockOutput: "invalid json data",
			mockErr:    nil,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := newMockExecutor()
			mock.On("pr checks", []byte(tt.mockOutput), tt.mockErr)

			client := NewWithExecutor(mock)
			checks, err := client.GetChecks(tt.prNumber)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetChecks() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && len(checks) != tt.wantCount {
				t.Errorf("GetChecks() returned %d checks, want %d", len(checks), tt.wantCount)
			}
		})
	}
}

func TestGetChecks_CheckConclusions(t *testing.T) {
	mock := newMockExecutor()
	mock.On("pr checks", []byte(`[
		{"name": "success-check", "state": "completed", "conclusion": "success", "link": ""},
		{"name": "failure-check", "state": "completed", "conclusion": "failure", "link": ""},
		{"name": "neutral-check", "state": "completed", "conclusion": "neutral", "link": ""},
		{"name": "cancelled-check", "state": "completed", "conclusion": "cancelled", "link": ""},
		{"name": "skipped-check", "state": "completed", "conclusion": "skipped", "link": ""},
		{"name": "timed-out-check", "state": "completed", "conclusion": "timed_out", "link": ""}
	]`), nil)

	client := NewWithExecutor(mock)
	checks, err := client.GetChecks(42)
	if err != nil {
		t.Fatalf("GetChecks() unexpected error: %v", err)
	}

	expectedConclusions := []string{"success", "failure", "neutral", "cancelled", "skipped", "timed_out"}
	for i, check := range checks {
		if check.Conclusion != expectedConclusions[i] {
			t.Errorf("Check %d conclusion = %q, want %q", i, check.Conclusion, expectedConclusions[i])
		}
	}
}

func TestGetChecks_CheckStates(t *testing.T) {
	mock := newMockExecutor()
	mock.On("pr checks", []byte(`[
		{"name": "queued-check", "state": "queued", "conclusion": "", "link": ""},
		{"name": "in-progress-check", "state": "in_progress", "conclusion": "", "link": ""},
		{"name": "completed-check", "state": "completed", "conclusion": "success", "link": ""}
	]`), nil)

	client := NewWithExecutor(mock)
	checks, err := client.GetChecks(42)
	if err != nil {
		t.Fatalf("GetChecks() unexpected error: %v", err)
	}

	expectedStates := []string{"queued", "in_progress", "completed"}
	for i, check := range checks {
		if check.Status != expectedStates[i] {
			t.Errorf("Check %d status = %q, want %q", i, check.Status, expectedStates[i])
		}
	}
}

func TestGetChecks_CommandArgs(t *testing.T) {
	mock := newMockExecutor()
	mock.On("pr checks", []byte(`[]`), nil)

	client := NewWithExecutor(mock)
	_, err := client.GetChecks(42)
	if err != nil {
		t.Fatalf("GetChecks() unexpected error: %v", err)
	}

	mock.AssertCalled(t, "gh", "pr", "checks", "42", "--json")
}
