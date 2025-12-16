package gh

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// GetReviews retrieves all reviews for a pull request.
func (c *Client) GetReviews(prNumber int) ([]Review, error) {
	args := []string{
		"pr", "reviews",
		strconv.Itoa(prNumber),
		"--json", strings.Join(reviewFields, ","),
	}

	output, err := c.executor.Run(ghCommand, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get reviews for PR #%d: %w", prNumber, err)
	}

	var reviews []Review
	if err := json.Unmarshal(output, &reviews); err != nil {
		return nil, fmt.Errorf("failed to parse reviews response: %w", err)
	}

	return reviews, nil
}

// SubmitReview submits a review on a pull request.
func (c *Client) SubmitReview(prNumber int, body string, event ReviewEvent) error {
	args := []string{
		"pr", "review",
		strconv.Itoa(prNumber),
		event.Flag(),
	}

	// Only add body if not empty
	if body != "" {
		args = append(args, "--body", body)
	}

	_, err := c.executor.Run(ghCommand, args...)
	if err != nil {
		return fmt.Errorf("failed to submit review on PR #%d: %w", prNumber, err)
	}

	return nil
}
