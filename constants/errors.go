package constants

// var usernameString string = fmt.Sprintf("Invalid username. Username must be between %d and %d characters, it can also contain numbers and underscore", utils.UsernameMinLen, utils.UsernameMaxLen)
const (
	LoginError      = "id or password mismatch"
	InvalidEmail    = "invalid email format"
	InvalidPhone    = "invalid phone format"
	InvalidPassword = "invalid password format. Password must contain at least 1 lowercase, 1 uppercase, 1 special character and 1 digit"
	InvalidUsername = "Invalid username. Username must be between 4 and 30 characters, it can also contain numbers and underscore"
)
