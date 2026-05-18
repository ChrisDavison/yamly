package frontmatter

import (
	"reflect"

	"gopkg.in/yaml.v3"
)

// ParseValue parses a CLI string argument as a YAML scalar or sequence.
// "true" → bool, "42" → int, "draft" → string, "[a, b]" → []any.
func ParseValue(s string) (any, error) {
	var v any
	if err := yaml.Unmarshal([]byte(s), &v); err != nil {
		return nil, err
	}
	return v, nil
}

// Matches reports whether fieldValue satisfies queryValue.
// If fieldValue is a []any, it returns true if queryValue is a member.
// Otherwise it uses deep equality.
func Matches(fieldValue, queryValue any) bool {
	if arr, ok := fieldValue.([]any); ok {
		for _, item := range arr {
			if reflect.DeepEqual(item, queryValue) {
				return true
			}
		}
		return false
	}
	return reflect.DeepEqual(fieldValue, queryValue)
}
