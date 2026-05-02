package utils

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

type LengthRule struct {
	Field    string
	Value    string
	Min      int
	Max      int
	Required bool
}

func ValidateStringLength(rule LengthRule) error {
	value := strings.TrimSpace(rule.Value)
	length := utf8.RuneCountInString(value)

	if rule.Required && length == 0 {
		return fmt.Errorf("%s is required", rule.Field)
	}

	if !rule.Required && length == 0 {
		return nil
	}

	if rule.Min > 0 && length < rule.Min {
		return fmt.Errorf("%s must be at least %d characters long", rule.Field, rule.Min)
	}

	if rule.Max > 0 && length > rule.Max {
		return fmt.Errorf("%s must be at most %d characters long", rule.Field, rule.Max)
	}

	return nil
}
