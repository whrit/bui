package gh

import (
	"testing"
)

func TestNew(t *testing.T) {
	client := New()
	if client == nil {
		t.Fatal("New() returned nil")
	}
	if client.executor == nil {
		t.Error("New() client has nil executor")
	}
}

func TestNewWithExecutor(t *testing.T) {
	mock := newMockExecutor()
	client := NewWithExecutor(mock)

	if client == nil {
		t.Fatal("NewWithExecutor() returned nil")
	}
	if client.executor != mock {
		t.Error("NewWithExecutor() did not set provided executor")
	}
}

func TestMergeMethod_String(t *testing.T) {
	tests := []struct {
		method MergeMethod
		want   string
	}{
		{MergeMethodMerge, "merge"},
		{MergeMethodSquash, "squash"},
		{MergeMethodRebase, "rebase"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.method.String(); got != tt.want {
				t.Errorf("MergeMethod.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestMergeMethod_Flag(t *testing.T) {
	tests := []struct {
		method MergeMethod
		want   string
	}{
		{MergeMethodMerge, "--merge"},
		{MergeMethodSquash, "--squash"},
		{MergeMethodRebase, "--rebase"},
	}

	for _, tt := range tests {
		t.Run(string(tt.method), func(t *testing.T) {
			if got := tt.method.Flag(); got != tt.want {
				t.Errorf("MergeMethod.Flag() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestPRFields(t *testing.T) {
	// Verify all required fields are present in prFields
	requiredFields := []string{
		"number", "title", "body", "state", "author",
		"headRefName", "baseRefName", "url", "createdAt", "updatedAt",
		"isDraft", "mergeable", "additions", "deletions", "labels", "reviewRequests",
	}

	for _, required := range requiredFields {
		found := false
		for _, field := range prFields {
			if field == required {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("prFields missing required field: %q", required)
		}
	}
}

func TestReviewFields(t *testing.T) {
	requiredFields := []string{"id", "author", "state", "body", "submittedAt"}

	for _, required := range requiredFields {
		found := false
		for _, field := range reviewFields {
			if field == required {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("reviewFields missing required field: %q", required)
		}
	}
}

func TestCheckFields(t *testing.T) {
	requiredFields := []string{"name", "state", "conclusion", "link"}

	for _, required := range requiredFields {
		found := false
		for _, field := range checkFields {
			if field == required {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("checkFields missing required field: %q", required)
		}
	}
}

func TestMockExecutor_RecordsCalls(t *testing.T) {
	mock := newMockExecutor()
	mock.On("test", []byte("output"), nil)

	_, _ = mock.Run("gh", "test", "arg1", "arg2")

	if len(mock.calls) != 1 {
		t.Fatalf("Expected 1 call, got %d", len(mock.calls))
	}

	call := mock.calls[0]
	if call.name != "gh" {
		t.Errorf("Call name = %q, want %q", call.name, "gh")
	}
	if len(call.args) != 3 {
		t.Errorf("Call args count = %d, want 3", len(call.args))
	}
}

func TestMockExecutor_ReturnsConfiguredResponse(t *testing.T) {
	mock := newMockExecutor()
	expectedOutput := []byte("test output")
	mock.On("test", expectedOutput, nil)

	output, err := mock.Run("gh", "test")

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if string(output) != string(expectedOutput) {
		t.Errorf("Output = %q, want %q", string(output), string(expectedOutput))
	}
}

func TestMockExecutor_ReturnsError(t *testing.T) {
	mock := newMockExecutor()

	// No response configured, should return error
	_, err := mock.Run("gh", "unknown", "command")

	if err == nil {
		t.Error("Expected error for unconfigured command")
	}
}

// Test concurrency safety of Client
func TestClient_ConcurrentAccess(t *testing.T) {
	mock := newMockExecutor()
	mock.On("pr list", []byte(`[]`), nil)
	mock.On("pr view", []byte(`{"number": 1, "title": "Test", "body": "", "state": "OPEN", "author": {"login": "user"}, "headRefName": "head", "baseRefName": "main", "url": "", "createdAt": "2024-01-01T00:00:00Z", "updatedAt": "2024-01-01T00:00:00Z", "isDraft": false, "mergeable": "UNKNOWN", "additions": 0, "deletions": 0, "labels": [], "reviewRequests": []}`), nil)
	mock.On("pr checks", []byte(`[]`), nil)
	mock.On("pr reviews", []byte(`[]`), nil)

	client := NewWithExecutor(mock)

	done := make(chan bool)

	// Run multiple goroutines concurrently
	for i := 0; i < 10; i++ {
		go func() {
			_, _ = client.ListPRs(PRFilters{State: "open"})
			done <- true
		}()
		go func() {
			_, _ = client.GetPR(1)
			done <- true
		}()
		go func() {
			_, _ = client.GetChecks(1)
			done <- true
		}()
		go func() {
			_, _ = client.GetReviews(1)
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 40; i++ {
		<-done
	}
}
