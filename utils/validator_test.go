package utils

import (
	"testing"
)

// Test data
var usernameTests = []struct {
	username string
	expected bool
}{
	{"validUser", true},
	{"valid_user", true},
	{"Valid-User1", true},
	{"inv@lidUser", false},
	{"Invalid__User", false},
	{"Invalid--User", false},
	{"InvalidUser-", false},
	{"1InvalidUser", false},
	{"abc", false},
	{"aVeryLongUsernameThatExceedsTheMaximumAllowedLength", false},
}

var passwordTests = []struct {
	password string
	expected bool
}{
	{"Passw0rd!", true},
	{"P@ssw0rd1", true},
	{"password", false},
	{"PASSWORD", false},
	{"12345678", false},
	{"Short1!", false},
	{"ThisPasswordIsWayTooLongToBeConsideredValidEvenThoughItMeetsOtherCriteria1!", false},
	{"P@ssw0rd123", true},
	{"password123", false},
	{"Passw0rd", false},
}

var emailTests = []struct {
	email    string
	expected bool
}{
	{"valid@example.com", true},
	{"invalid-email", false},
	{"another.invalid.email@com", true},
	{"valid_email+alias@example.com", true},
	{"@missingusername.com", false},
	{"missingdomain@", false},
	{"missingatsymbol.com", false},
	{"valid@subdomain.example.com", true},
}

var phoneTests = []struct {
	phone    string
	expected bool
}{
	{"123-456-7890", true},
	{"(123) 456-7890", true},
	{"1234567890", true},
	{"+11234567890", false},
	{"123-456-789", false},
	{"123-45-67890", true},
	{"abcdefghij", false},
	{"", false},
	{"123.456.7890", true},
}

func TestValidateUsername(t *testing.T) {
	for _, test := range usernameTests {
		result := validateUsername(test.username)
		if result != test.expected {
			t.Errorf("validateUsername(%q) = %v; want %v", test.username, result, test.expected)
		}
	}
}

func TestValidatePassword(t *testing.T) {
	for _, test := range passwordTests {
		result := validatePassword(test.password)
		if result != test.expected {
			t.Errorf("validatePassword(%q) = %v; want %v", test.password, result, test.expected)
		}
	}
}

func TestValidateEmail(t *testing.T) {
	for _, test := range emailTests {
		result := validateEmail(test.email)
		if result != test.expected {
			t.Errorf("validateEmail(%q) = %v; want %v", test.email, result, test.expected)
		}
	}
}

func TestValidatePhone(t *testing.T) {
	for _, test := range phoneTests {
		result := validatePhone(test.phone)
		if result != test.expected {
			t.Errorf("validatePhone(%q) = %v; want %v", test.phone, result, test.expected)
		}
	}
}
