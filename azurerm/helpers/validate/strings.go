package validate

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

// NoEmptyStrings validates that the string is not just whitespace characters (equal to [\r\n\t\f\v ])
func NoEmptyStrings(i interface{}, k string) ([]string, []error) {
	v, ok := i.(string)
	if !ok {
		return nil, []error{fmt.Errorf("expected type of %q to be string", k)}
	}

	if strings.TrimSpace(v) == "" {
		return nil, []error{fmt.Errorf("%q must not be empty", k)}
	}

	return nil, nil
}

// StringAtLeast returns a SchemaValidateFunc which tests if the provided value
// is of type string and has length of at least `min
func StringAtLeast(min int) schema.SchemaValidateFunc {
	return func(i interface{}, k string) (s []string, es []error) {
		v, ok := i.(string)
		if !ok {
			es = append(es, fmt.Errorf("expected type of %s to be string", k))
			return
		}
		if len(v) < min {
			es = append(es, fmt.Errorf("%s must be at least %d characters - got %s", k, min, v))
		}
		return
	}
}
