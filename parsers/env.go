package parsers

import (
	"errors"
	"fmt"
	"os"
	"regexp"
)

var envRegex = regexp.MustCompile(`\${([^}:]+)(?::([^}]*))?}`)

// ParseEnvironment parses the content of a configuration file and replaces
// placeholders for environment variables with the value from the OS, or
// uses the default-value if the environment variable isn't set.
//
// ParseEnvironment looks for the following patterns:
//
//	${MY_ENV}
//	${MY_ENV:defaultValue}
//
// If a default value is not provided and an environment variable is not set an
// error will be returned.
func ParseEnvironment(content []byte) ([]byte, error) {
	var errs []error

	env := envRegex.ReplaceAllFunc(content, func(match []byte) []byte {
		parts := envRegex.FindSubmatch(match)
		name := parts[1]
		defaultValue := parts[2]

		val, exists := os.LookupEnv(string(name))
		if exists {
			return []byte(val)
		}

		if defaultValue != nil && len(defaultValue) > 0 {
			return defaultValue
		}

		errs = append(errs, fmt.Errorf("parse env: %s not set", name))
		return match
	})

	return env, errors.Join(errs...)
}
