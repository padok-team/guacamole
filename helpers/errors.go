package helpers

import "github.com/padok-team/guacamole/data"

func HasError(checks []data.Check) bool {
	for _, check := range checks {
		if check.Status == "❌" {
			return true
		}
	}
	return false
}
