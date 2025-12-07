package validate

import "fmt"

// Validator represents validation errors.
type Validator struct {
	errors []string
}

// New creates a new validator.
func New() *Validator {
	return &Validator{
		errors: make([]string, 0),
	}
}

// Required adds an error if the value is empty.
func (v *Validator) Required(field, value string) {
	if value == "" {
		v.errors = append(v.errors, fmt.Sprintf("%s is required", field))
	}
}

// HasErrors returns true if there are validation errors.
func (v *Validator) HasErrors() bool {
	return len(v.errors) > 0
}

// Error returns the validation error message.
func (v *Validator) Error() string {
	if len(v.errors) == 0 {
		return ""
	}
	msg := "validation errors: "
	for i, err := range v.errors {
		if i > 0 {
			msg += "; "
		}
		msg += err
	}
	return msg
}

