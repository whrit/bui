package gh

import (
	"errors"
	"testing"
)

func TestGetReviews(t *testing.T) {
	tests := []struct {
		name       string
		prNumber   int
		mockOutput string
		mockErr    error
		wantCount  int
		wantErr    bool
	}{
		{
			name:     "get reviews successfully",
			prNumber: 42,
			mockOutput: `[
				{
					"id": 1,
					"author": {"login": "reviewer1"},
					"state": "APPROVED",
					"body": "LGTM!",
					"submittedAt": "2024-01-15T12:00:00Z"
				},
				{
					"id": 2,
					"author": {"login": "reviewer2"},
					"state": "CHANGES_REQUESTED",
					"body": "Please fix the typo on line 42",
					"submittedAt": "2024-01-15T14:00:00Z"
				}
			]`,
			wantCount: 2,
			wantErr:   false,
		},
		{
			name:       "no reviews on PR",
			prNumber:   43,
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
			prNumber:   44,
			mockOutput: "not valid json",
			mockErr:    nil,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := newMockExecutor()
			mock.On("pr reviews", []byte(tt.mockOutput), tt.mockErr)

			client := NewWithExecutor(mock)
			reviews, err := client.GetReviews(tt.prNumber)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetReviews() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && len(reviews) != tt.wantCount {
				t.Errorf("GetReviews() returned %d reviews, want %d", len(reviews), tt.wantCount)
			}
		})
	}
}

func TestGetReviews_ReviewStates(t *testing.T) {
	mock := newMockExecutor()
	mock.On("pr reviews", []byte(`[
		{"id": 1, "author": {"login": "user1"}, "state": "APPROVED", "body": "", "submittedAt": "2024-01-15T12:00:00Z"},
		{"id": 2, "author": {"login": "user2"}, "state": "CHANGES_REQUESTED", "body": "Fix it", "submittedAt": "2024-01-15T13:00:00Z"},
		{"id": 3, "author": {"login": "user3"}, "state": "COMMENTED", "body": "Nice!", "submittedAt": "2024-01-15T14:00:00Z"},
		{"id": 4, "author": {"login": "user4"}, "state": "PENDING", "body": "", "submittedAt": "2024-01-15T15:00:00Z"}
	]`), nil)

	client := NewWithExecutor(mock)
	reviews, err := client.GetReviews(42)
	if err != nil {
		t.Fatalf("GetReviews() unexpected error: %v", err)
	}

	expectedStates := []string{"APPROVED", "CHANGES_REQUESTED", "COMMENTED", "PENDING"}
	for i, review := range reviews {
		if review.State != expectedStates[i] {
			t.Errorf("Review %d state = %q, want %q", i, review.State, expectedStates[i])
		}
	}
}

func TestSubmitReview(t *testing.T) {
	tests := []struct {
		name     string
		prNumber int
		body     string
		event    ReviewEvent
		mockErr  error
		wantErr  bool
		wantArgs []string
	}{
		{
			name:     "submit approval",
			prNumber: 42,
			body:     "LGTM!",
			event:    ReviewApprove,
			mockErr:  nil,
			wantErr:  false,
			wantArgs: []string{"pr", "review", "42", "--approve", "--body", "LGTM!"},
		},
		{
			name:     "request changes",
			prNumber: 43,
			body:     "Please fix the tests",
			event:    ReviewRequestChanges,
			mockErr:  nil,
			wantErr:  false,
			wantArgs: []string{"pr", "review", "43", "--request-changes"},
		},
		{
			name:     "submit comment",
			prNumber: 44,
			body:     "Just a comment",
			event:    ReviewComment,
			mockErr:  nil,
			wantErr:  false,
			wantArgs: []string{"pr", "review", "44", "--comment"},
		},
		{
			name:     "empty body approval",
			prNumber: 45,
			body:     "",
			event:    ReviewApprove,
			mockErr:  nil,
			wantErr:  false,
			wantArgs: []string{"pr", "review", "45", "--approve"},
		},
		{
			name:     "review fails - PR not found",
			prNumber: 999,
			body:     "Review",
			event:    ReviewApprove,
			mockErr:  errors.New("GraphQL: Could not resolve to a PullRequest"),
			wantErr:  true,
		},
		{
			name:     "review fails - cannot review own PR",
			prNumber: 50,
			body:     "Self review",
			event:    ReviewApprove,
			mockErr:  errors.New("Pull request author cannot approve their own pull request"),
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := newMockExecutor()
			mock.On("pr review", []byte(""), tt.mockErr)

			client := NewWithExecutor(mock)
			err := client.SubmitReview(tt.prNumber, tt.body, tt.event)

			if (err != nil) != tt.wantErr {
				t.Errorf("SubmitReview() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && len(tt.wantArgs) > 0 {
				mock.AssertCalled(t, "gh", tt.wantArgs...)
			}
		})
	}
}

func TestReviewEvent_Flag(t *testing.T) {
	tests := []struct {
		event ReviewEvent
		want  string
	}{
		{ReviewApprove, "--approve"},
		{ReviewRequestChanges, "--request-changes"},
		{ReviewComment, "--comment"},
		{ReviewEvent("INVALID"), "--comment"}, // Default to comment for unknown
	}

	for _, tt := range tests {
		t.Run(string(tt.event), func(t *testing.T) {
			if got := tt.event.Flag(); got != tt.want {
				t.Errorf("ReviewEvent(%q).Flag() = %q, want %q", tt.event, got, tt.want)
			}
		})
	}
}
