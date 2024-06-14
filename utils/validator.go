package utils

import (
	"fmt"
	"net/mail"
	"regexp"

	"github.com/go-playground/validator/v10"
	"github.com/rs/zerolog/log"
)

var (
	startsWithLetterRegex                   = regexp.MustCompile(`^[A-Za-z]`)
	isValidCharactersRegex                  = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)
	hasConsecutiveUnderscoresOrHyphensRegex = regexp.MustCompile(`(__|-{2,})`)
	endsWithUnderscoreOrHyphenRegex         = regexp.MustCompile(`[_-]$`)
	hasUppercaseRegex                       = regexp.MustCompile(`[A-Z]`)
	hasLowercaseRegex                       = regexp.MustCompile(`[a-z]`)
	hasNumberRegex                          = regexp.MustCompile(`[0-9]`)
	hasSpecialCharRegex                     = regexp.MustCompile(`[!@#$%^&*()]`)
	cleanNumberRegex                        = regexp.MustCompile(`[\s-().]`)
	hasTenDigitsRegex                       = regexp.MustCompile(`^\d{10}$`)
	isValidPhoneRegex                       = regexp.MustCompile(`^\+?1?\d{10}$`)
)

const (
	usernameMinLen = 4
	usernameMaxLen = 30
	passwordMinLen = 8
	passwordMaxLen = 64
	fullNameMinLen = 3
	fullNameMaxLen = 30
	emailMinLen    = 3
	emailMaxLen    = 254
)

var ValidUsername validator.Func = func(fl validator.FieldLevel) bool {
	if username, ok := fl.Field().Interface().(string); ok {
		return validateUsername(username)
	}
	return false
}

var ValidEmail validator.Func = func(fl validator.FieldLevel) bool {
	if email, ok := fl.Field().Interface().(string); ok {
		return validateEmail(email)
	}
	return false
}

var ValidPassword validator.Func = func(fl validator.FieldLevel) bool {
	if password, ok := fl.Field().Interface().(string); ok {
		return validatePassword(password)
	}
	return false
}

var ValidPhone validator.Func = func(fl validator.FieldLevel) bool {
	if phone, ok := fl.Field().Interface().(string); ok {
		return validatePhone(phone)
	}
	return false
}

func ValidateString(value string, min int, max int) error {
	if n := len(value); n < min || n > max {
		return fmt.Errorf("length must be between %d and %d characters", min, max)
	}
	return nil
}

func validateUsername(username string) bool {
	if len(username) < usernameMinLen || len(username) > usernameMaxLen {
		return false
	}
	if !startsWithLetterRegex.MatchString(username) {
		return false
	}
	if !isValidCharactersRegex.MatchString(username) {
		return false
	}
	if hasConsecutiveUnderscoresOrHyphensRegex.MatchString(username) {
		return false
	}
	if endsWithUnderscoreOrHyphenRegex.MatchString(username) {
		return false
	}
	return true
}

func validatePhone(phoneNumber string) bool {
	cleanNumber := cleanNumberRegex.ReplaceAllString(phoneNumber, "")
	return hasTenDigitsRegex.MatchString(cleanNumber)
}

func validatePassword(password string) bool {
	if len(password) <= passwordMinLen || len(password) >= passwordMaxLen {
		return false
	}
	if !hasUppercaseRegex.MatchString(password) || !hasLowercaseRegex.MatchString(password) || !hasNumberRegex.MatchString(password) || !hasSpecialCharRegex.MatchString(password) {
		return false
	}
	commonPatterns := []string{"123456", "password"}
	for _, pattern := range commonPatterns {
		if password == pattern {
			return false
		}
	}

	// FIXME: Check for uniqueness
	// You can add your own logic here to check if the password has been used before

	// Check for personal information
	// You can add your own logic here to check if the password contains personal information

	// Check for randomness
	// You can add your own logic here to check if the password is random

	return true
}

func validateEmail(value string) bool {
	if _, err := mail.ParseAddress(value); err != nil {
		log.Error().Msgf("invalid email: %s", err)
		return false
	}
	return true
}
