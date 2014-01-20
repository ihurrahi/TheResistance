package users

import (
	"net/http"
	"resistance/utils"
)

const (
	USERNAME_KEY        = "username"
	PASSWORD_KEY        = "password"
	REPEAT_PASSWORD_KEY = "repeatPassword"
	COOKIE_NAME         = "RC"
)

const (
	USERS_TABLE           = "users"
	USERS_ID_COLUMN       = "user_id"
	USERS_USERNAME_COLUMN = "username"
)

type User struct {
	Username string
	UserId   int
}

var UNKNOWN_USER = &User{Username: "", UserId: -1}

// isValidUser determines whether the user object is valid.
func (user *User) IsValidUser() bool {
	return user.UserId > 0 && user.Username != ""
}

// UserSignUp signs up the user given in the request and performs basic
// validation on the fields. Returns if there is an error and the
// corresponding error message
// TODO: method should just use form values instead of being passed entire http request.
func UserSignUp(request *http.Request) (bool, string) {
	username := request.FormValue(USERNAME_KEY)
	user := lookupUserByUsername(username)
	if user.IsValidUser() {
		// If this is a valid user (not the UNKNOWN user), then the user already exists
		return true, "Username " + username + " already exists!"
	}

	if len(username) < 3 {
		return true, "Username must be at least 3 characters long."
	}

	if len(username) > 10 {
		return true, "Username must be at most 10 characters long."
	}

	password := request.FormValue(PASSWORD_KEY)
	repeatPassword := request.FormValue(REPEAT_PASSWORD_KEY)
	if password != repeatPassword {
		return true, "Passwords do not match"
	}

	if len(password) < 3 {
		return true, "Password must be at least 3 characters long."
	}

	if len(password) > 30 {
		return true, "Password must be at most 30 characters long."
	}

	err := persistUser(username, password)
	if err != nil {
		utils.LogMessage("Error persisting user: "+err.Error(), utils.USER_LOG_PATH)
		return true, "Error creating user"
	}

	return false, ""
}

// ValidateUserCookie validates a user given the cookies from a request.
func ValidateUserCookie(requestCookies []*http.Cookie) *User {
	if len(requestCookies) == 0 {
		return UNKNOWN_USER
	}

	cookie := requestCookies[0]
	user := lookupUserByCookie(cookie)
	if user.IsValidUser() {
		return user
	}

	return UNKNOWN_USER
}

// ValidateUser is the entry point for the login handler. It validates the
// user credentials (it assumes cookie validation failed already) and if
// they provide valid credentials, then creates and cookie and returns it.
func ValidateUser(request *http.Request) (*http.Cookie, bool) {
	if len(request.Form) > 0 {
		username := request.FormValue(USERNAME_KEY)
		password := request.FormValue(PASSWORD_KEY)
		utils.LogMessage("will validate user:"+username+" with password:"+password, utils.USER_LOG_PATH)
		id, validUser := validateUserCredentials(username, password)
		if validUser {
			cookie := generateNewCookie(username)
			utils.LogMessage("cookie created: "+cookie.String(), utils.USER_LOG_PATH)
			err := storeCookie(id, cookie)
			if err != nil {
				utils.LogMessage("Error storing cookie"+err.Error(), utils.USER_LOG_PATH)
			}
			return cookie, true
		}
	}
	return nil, false
}

// generatenewCookie creates new cookies for users without one.
// TODO: have a better cookie generation strategy than using the username -_-
func generateNewCookie(username string) *http.Cookie {
	cookie := &http.Cookie{Name: COOKIE_NAME, Value: username}
	utils.LogMessage("Creating a new cookie "+cookie.String(), utils.USER_LOG_PATH)
	return cookie
}
