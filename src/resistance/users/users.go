package users

import (
    "log"
    "strconv"
    "net/url"
    "net/http"
    "net/http/cookiejar"
    "database/sql"
    _ "github.com/go-sql-driver/mysql"
)

const (
    USERNAME_KEY = "username"
    PASSWORD_KEY = "password"
    COOKIE_NAME = "RC"
)

type User struct {
    Username string
    Id int
}

var cookieJar *cookiejar.Jar
var cookieToUserIdMap map[string]int
var userLog *log.Logger
var myURL *url.URL
var db *sql.DB

func lookupUser(id int) *User {
    var user *User
    var username string
    err := db.QueryRow("select username from users where id = ?", id).Scan(&username)
    switch {
    case err == sql.ErrNoRows:
        userLog.Println("Warning: No user found for id: " + strconv.Itoa(id))
        user = nil
    case err != nil:
        userLog.Fatalln("Error while looking up user: " + err.Error())
    default:
        user = new(User)
        user.Id = id
        user.Username = username
        userLog.Println("Found a User! " + user.Username)
    }
    
    return user
}

func validateUserCredentials(user string, pass string) (int, bool) {
    var id int
    err := db.QueryRow("select id from users where username = ? and password = ?", user, pass).Scan(&id)
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

func ValidateUserCookie(request *http.Request) (*User, bool) {
    requestCookies := request.Cookies()
    if len(requestCookies) == 0 {
        return nil, false
    }
    
    cookie := requestCookies[0]
    cookies := cookieJar.Cookies(myURL)
    for i := 0; i < len(cookies); i++ {
        if cookies[i].String() == cookie.String() {
            user := lookupUser(cookieToUserIdMap[cookie.Value])
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

func Initialize(logger *log.Logger) {
    var err error
    cookieJar, err = cookiejar.New(nil)
    if err != nil {
        panic(err)
    }
    
    cookieToUserIdMap = make(map[string]int)
    
    userLog = logger
    
    myURL = &url.URL{Scheme: "http"}
    
    db, err = sql.Open("mysql", "resistance:resistance@/resistance")
    if err != nil {
        panic(err)
    }
}