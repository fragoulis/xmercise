package company

import "strings"

// Violation describes one invalid command field.
type Violation struct {
	Field   string
	Message string
}

// ValidationError contains invalid command fields.
type ValidationError struct {
	Violations []Violation
}

// Error returns a compact validation error message.
func (e ValidationError) Error() string {
	if len(e.Violations) == 0 {
		return "company validation failed"
	}

	parts := make([]string, 0, len(e.Violations))
	for _, violation := range e.Violations {
		parts = append(parts, violation.Field+": "+violation.Message)
	}

	return "company validation failed: " + strings.Join(parts, ", ")
}
