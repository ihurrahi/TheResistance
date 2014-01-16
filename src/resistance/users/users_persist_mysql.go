package users

import (
	"database/sql"
	"net/http"
	"resistance/utils"
	"strconv"
)

const (
	PERSIST_USER_QUERY       = "insert into users (`username`, `password`) values (?, ?)"
	PERSIST_COOKIE_QUERY     = "update users set cookie = ? where user_id = ?"
	CREDENTIALS_QUERY        = "select user_id from users where username = ? and password = ?"
	LOOKUP_BY_USERNAME_QUERY = "select user_id from users where username = ?"
	LOOKUP_BY_USERID_QUERY   = "select username from users where user_id = ?"
	LOOKUP_BY_COOKIE_QUERY   = "select user_id, username from users where cookie = ?"
)

var db *sql.DB

func init() {
	db = utils.ConnectToDB()
}

// lookupUserById looks up the user in the DB based on the given id.
func LookupUserById(id int) *User {
	user := UNKNOWN_USER
	var username string
	err := db.QueryRow(LOOKUP_BY_USERID_QUERY, id).Scan(&username)
	switch {
	case err == sql.ErrNoRows:
		utils.LogMessage("Warning: No user found for id: "+strconv.Itoa(id), utils.USER_LOG_PATH)
		user = UNKNOWN_USER
	case err != nil:
		utils.LogMessage("Error while looking up user: "+err.Error(), utils.USER_LOG_PATH)
	default:
		user = new(User)
		user.UserId = id
		user.Username = username
		utils.LogMessage("Found a User! "+user.Username, utils.USER_LOG_PATH)
	}

	return user
}

// lookupUserByUsername looks up the user in the DB based on the given username.
func lookupUserByUsername(username string) *User {
	user := UNKNOWN_USER
	var id int
	err := db.QueryRow(LOOKUP_BY_USERNAME_QUERY, username).Scan(&id)
	switch {
	case err == sql.ErrNoRows:
		utils.LogMessage("Warning: No user found for id: "+strconv.Itoa(id), utils.USER_LOG_PATH)
		user = UNKNOWN_USER
	case err != nil:
		utils.LogMessage("Error while looking up user: "+err.Error(), utils.USER_LOG_PATH)
	default:
		user = new(User)
		user.UserId = id
		user.Username = username
		utils.LogMessage("Found a User! "+user.Username, utils.USER_LOG_PATH)
	}

	return user
}

// lookupUserByCookie looks up the user in the DB based on the given cookie.
func lookupUserByCookie(cookie *http.Cookie) *User {
	user := UNKNOWN_USER
	var id int
	var username string
	err := db.QueryRow(LOOKUP_BY_COOKIE_QUERY, cookie.Value).Scan(&id, &username)
	switch {
	case err == sql.ErrNoRows:
		utils.LogMessage("Warning: No user found for id: "+strconv.Itoa(id), utils.USER_LOG_PATH)
		user = UNKNOWN_USER
	case err != nil:
		utils.LogMessage("Error while looking up user: "+err.Error(), utils.USER_LOG_PATH)
	default:
		user = new(User)
		user.UserId = id
		user.Username = username
		utils.LogMessage("Found a User! "+user.Username, utils.USER_LOG_PATH)
	}

	return user
}

// persistUser stores the user in the DB, effectively completing registration
// of a user.
func persistUser(username string, password string) error {
	_, err := db.Exec(PERSIST_USER_QUERY, username, password)
	return err
}

// validateUserCredentials validates the given username and password combination.
func validateUserCredentials(user string, pass string) (int, bool) {
	var id int
	err := db.QueryRow(CREDENTIALS_QUERY, user, pass).Scan(&id)
	switch {
	case err == sql.ErrNoRows:
		utils.LogMessage("Login failed for username: "+user+" using password: "+pass, utils.USER_LOG_PATH)
		return 0, false
	case err != nil:
		utils.LogMessage("Error while looking up user: "+err.Error(), utils.USER_LOG_PATH)
	default:
	}
	return id, true
}

// storeCookie stores the cookie in the DB for the given user id.
func storeCookie(id int, cookie *http.Cookie) error {
	_, err := db.Exec(PERSIST_COOKIE_QUERY, cookie.Value, id)
	return err
}
