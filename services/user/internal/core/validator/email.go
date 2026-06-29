package validator

import "regexp"

var emailRegex = regexp.MustCompile(`(?i)^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,64}$`)

func IsValidEmail(email string) bool {
	if len(email) > 320 {
		return false
	}
	return emailRegex.MatchString(email)
}
