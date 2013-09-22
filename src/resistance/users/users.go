package users

import (
    "strconv"
    "net/http"
    "database/sql"
    "resistance/utils"
)

const (
    USERNAME_KEY = "username"
    PASSWORD_KEY = "password"
    REPEAT_PASSWORD_KEY = "repeatPassword"
    COOKIE_NAME = "RC"
    
    CREDENTIALS_QUERY = "select user_id from users where username = ? and password = ?"
    LOOKUP_BY_USERNAME_QUERY = "select user_id from users where username = ?"
    LOOKUP_BY_USERID_QUERY = "select username from users where user_id = ?"
    LOOKUP_BY_COOKIE_QUERY = "select user_id, username from users where cookie = ?"
    PERSIST_USER_QUERY = "insert into users (`username`, `password`) values (?, ?)"
    PERSIST_COOKIE_QUERY = "update users set cookie = ? where user_id = ?"
)

type User struct {
    Username string
    UserId int
}

func (user *User) isValidUser() bool {
    return user.UserId > 0 && user.Username != ""
}

var UNKNOWN_USER = &User{Username:"", UserId:-1}

func lookupUserById(id int) *User {
    db, err := utils.ConnectToDB()
    if err != nil {
        return UNKNOWN_USER
    }

    user := UNKNOWN_USER
    var username string
    err = db.QueryRow(LOOKUP_BY_USERID_QUERY, id).Scan(&username)
    switch {
    case err == sql.ErrNoRows:
        utils.LogMessage("Warning: No user found for id: " + strconv.Itoa(id), utils.USER_LOG_PATH)
        user = UNKNOWN_USER
    case err != nil:
        utils.LogMessage("Error while looking up user: " + err.Error(), utils.USER_LOG_PATH)
    default:
        user = new(User)
        user.UserId = id
        user.Username = username
        utils.LogMessage("Found a User! " + user.Username, utils.USER_LOG_PATH)
    }
    
    return user
}

func lookupUserByUsername(username string) *User {
    db, err := utils.ConnectToDB()
    if err != nil {
        return UNKNOWN_USER
    }

    user := UNKNOWN_USER
    var id int
    err = db.QueryRow(LOOKUP_BY_USERNAME_QUERY, username).Scan(&id)
    switch {
    case err == sql.ErrNoRows:
        utils.LogMessage("Warning: No user found for id: " + strconv.Itoa(id), utils.USER_LOG_PATH)
        user = UNKNOWN_USER
    case err != nil:
        utils.LogMessage("Error while looking up user: " + err.Error(), utils.USER_LOG_PATH)
    default:
        user = new(User)
        user.UserId = id
        user.Username = username
        utils.LogMessage("Found a User! " + user.Username, utils.USER_LOG_PATH)
    }
    
    return user
}

func lookupUserByCookie(cookie *http.Cookie) *User {
    db, err := utils.ConnectToDB()
    if err != nil {
        return UNKNOWN_USER
    }

    user := UNKNOWN_USER
    var id int
    var username string
    err = db.QueryRow(LOOKUP_BY_COOKIE_QUERY, cookie.Value).Scan(&id, &username)
    switch {
    case err == sql.ErrNoRows:
        utils.LogMessage("Warning: No user found for id: " + strconv.Itoa(id), utils.USER_LOG_PATH)
        user = UNKNOWN_USER
    case err != nil:
        utils.LogMessage("Error while looking up user: " + err.Error(), utils.USER_LOG_PATH)
    default:
        user = new(User)
        user.UserId = id
        user.Username = username
        utils.LogMessage("Found a User! " + user.Username, utils.USER_LOG_PATH)
    }
    
    return user
}

func persistUser(username string, password string) error {
    db, err := utils.ConnectToDB()
    if err != nil {
        return err
    }

    _, err = db.Exec(PERSIST_USER_QUERY, username, password)
    return err
}

func validateUserCredentials(user string, pass string) (int, bool) {
    db, err := utils.ConnectToDB()
    if err != nil {
        return 0, false
    }
    
    var id int
    err = db.QueryRow(CREDENTIALS_QUERY, user, pass).Scan(&id)
    switch {
    case err == sql.ErrNoRows:
        utils.LogMessage("Login failed for username: " + user + " using password: " + pass, utils.USER_LOG_PATH)
        return 0, false
    case err != nil:
        utils.LogMessage("Error while looking up user: " + err.Error(), utils.USER_LOG_PATH)
    default:
    }
    return id, true
}

func UserSignUp(request *http.Request) (bool, string) {
    username := request.FormValue(USERNAME_KEY)
    user := lookupUserByUsername(username)
    if user != nil {
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
        utils.LogMessage("Error persisting user: " + err.Error(), utils.USER_LOG_PATH)
        return true, "Error creating user"
    }
    
    return false, ""
}

func ValidateUserCookie(requestCookies []*http.Cookie) (*User, bool) {
    if len(requestCookies) == 0 {
        return nil, false
    }
    
    cookie := requestCookies[0]
    user := lookupUserByCookie(cookie)
    if user.isValidUser() {
        return user, true
    }

    return nil, false
}

func ValidateUser(request *http.Request) (*http.Cookie, bool) {
    if len(request.Form) > 0 {
        username := request.FormValue(USERNAME_KEY)
        password := request.FormValue(PASSWORD_KEY)
        utils.LogMessage("will validate user:" + username + " with password:" + password, utils.USER_LOG_PATH)
        id, validUser := validateUserCredentials(username, password)
        if validUser {
            cookie := generateNewCookie(username)
            utils.LogMessage("cookie created: " + cookie.String(), utils.USER_LOG_PATH)
            err := storeCookie(id, cookie)
            if err != nil {
                utils.LogMessage("Error storing cookie" + err.Error(), utils.USER_LOG_PATH)
            }
            return cookie, true
        }
    }
    return nil, false
}

func storeCookie(id int, cookie *http.Cookie) error {
    db, err := utils.ConnectToDB()
    if err != nil {
        return err
    }
    
    _, err = db.Exec(PERSIST_COOKIE_QUERY, cookie.Value, id)
    return err
}

func generateNewCookie(username string) *http.Cookie {
    cookie := &http.Cookie{Name: COOKIE_NAME, Value: username}
    utils.LogMessage("Creating a new cookie " + cookie.String(), utils.USER_LOG_PATH)
    return cookie
}