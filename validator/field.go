/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package validator

import "regexp"

const (
	Phone = iota
	Mail
	Unknown
)
const (
	emailPattern = `^[a-zA-Z0-9.!#$%&'*+/=?^_` + "`" + `{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$`
	phonePattern = `^1[0-9]{10}$`
)

func PhoneOrMail(input string) int {
	phoneMatch, _ := regexp.MatchString(phonePattern, input)
	if phoneMatch {
		return Phone
	} else {
		emailMatch, _ := regexp.MatchString(emailPattern, input)
		if emailMatch {
			return Mail
		}
	}
	return Unknown
}

func IsPhone(input string) bool {
	phoneMatch, _ := regexp.MatchString(phonePattern, input)
	return phoneMatch
}
