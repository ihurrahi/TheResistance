package users

import (
    "log"
    "strconv"
    "net/url"
    "net/http"
    "net/http/cookiejar"
    "database/sql"
)

const (
    USERNAME_KEY = "username"
    PASSWORD_KEY = "password"
    REPEAT_PASSWORD_KEY = "repeatPassword"
    COOKIE_NAME = "RC"
    
    CREDENTIALS_QUERY = "select user_id from users where username = ? and password = ?"
    LOOKUP_BY_USERNAME_QUERY = "select user_id from users where username = ?"
    LOOKUP_BY_USERID_QUERY = "select username from users where user_id = ?"
    PERSIST_USER_QUERY = "insert into users (`username`, `password`) values (?, ?)"
)

type User struct {
    Username string
    UserId int
}

var cookieJar *cookiejar.Jar
var cookieToUserIdMap map[string]int
var userLog *log.Logger
var myURL *url.URL
var db *sql.DB

func lookupUserById(id int) *User {
    var user *User
    var username string
    err := db.QueryRow(LOOKUP_BY_USERID_QUERY, id).Scan(&username)
    switch {
    case err == sql.ErrNoRows:
        userLog.Println("Warning: No user found for id: " + strconv.Itoa(id))
        user = nil
    case err != nil:
        userLog.Fatalln("Error while looking up user: " + err.Error())
    default:
        user = new(User)
        user.UserId = id
        user.Username = username
        userLog.Println("Found a User! " + user.Username)
    }
    
    return user
}

func lookupUserByUsername(username string) *User {
    var user *User
    var id int
    err := db.QueryRow(LOOKUP_BY_USERNAME_QUERY, username).Scan(&id)
    switch {
    case err == sql.ErrNoRows:
        userLog.Println("Warning: No user found for id: " + strconv.Itoa(id))
        user = nil
    case err != nil:
        userLog.Fatalln("Error while looking up user: " + err.Error())
    default:
        user = new(User)
        user.UserId = id
        user.Username = username
        userLog.Println("Found a User! " + user.Username)
    }
    
    return user
}

func persistUser(username string, password string) error {
    _, err := db.Exec(PERSIST_USER_QUERY, username, password)
    return err
}

func validateUserCredentials(user string, pass string) (int, bool) {
    var id int
    err := db.QueryRow(CREDENTIALS_QUERY, user, pass).Scan(&id)
    switch {
    case err == sql.ErrNoRows:
        userLog.Println("Login failed for username: " + user + " using password: " + pass)
        return 0, false
    case err != nil:
        userLog.Fatalln("Error while looking up user: " + err.Error())
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
        userLog.Println("Error persisting user: " + err.Error())
        return true, "Error creating user"
    }
    
    return false, ""
}

func ValidateUserCookie(request *http.Request) (*User, bool) {
    requestCookies := request.Cookies()
    if len(requestCookies) == 0 {
        return nil, false
    }
    
    cookie := requestCookies[0]
    cookies := cookieJar.Cookies(myURL)
    for i := 0; i < len(cookies); i++ {
        if cookies[i].String() == cookie.String() {
            user := lookupUserById(cookieToUserIdMap[cookie.Value])
            return user, true
        }
    }
    return nil, false
}

func ValidateUser(request *http.Request) (*http.Cookie, bool) {
    if len(request.Form) > 0 {
        username := request.FormValue(USERNAME_KEY)
        password := request.FormValue(PASSWORD_KEY)
        userLog.Printf("will validate user:%v with password:%v", username, password)
        id, validUser := validateUserCredentials(username, password)
        if validUser {
            cookie := generateNewCookie(username)
            storeCookie(id, cookie)
            cookies := cookieJar.Cookies(myURL)
            cookies = append(cookies, cookie)
            userLog.Printf("cookie created: %v", cookie)
            cookieJar.SetCookies(myURL, cookies)
            return cookie, true
        }
    }
    return nil, false
}

func storeCookie(id int, cookie *http.Cookie) {
    cookieToUserIdMap[cookie.Value] = id
}

func generateNewCookie(username string) *http.Cookie {
    cookie := &http.Cookie{Name: COOKIE_NAME, Value: username}
    userLog.Println("Creating a new cookie " + cookie.String())
    return cookie
}

func Initialize(logger *log.Logger, database *sql.DB) {
    var err error
    cookieJar, err = cookiejar.New(nil)
    if err != nil {
        panic(err)
    }
    
    cookieToUserIdMap = make(map[string]int)
    
    userLog = logger
    db = database
    
    myURL = &url.URL{Scheme: "http"}
    
}