package gh

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// GetChecks retrieves all CI/status checks for a pull request.
func (c *Client) GetChecks(prNumber int) ([]Check, error) {
	if err := validatePRNumber(prNumber); err != nil {
		return nil, err
	}

	args := []string{
		"pr", "checks",
		strconv.Itoa(prNumber),
		"--json", strings.Join(checkFields, ","),
	}

	output, err := c.executor.Run(ghCommand, args...)
	if err != nil {
		return nil, NewPRError(prNumber, "get checks", err)
	}

	var checks []Check
	if err := json.Unmarshal(output, &checks); err != nil {
		return nil, fmt.Errorf("failed to parse checks response: %w", err)
	}

	return checks, nil
}
