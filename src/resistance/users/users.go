package users

import (
    "net/http"
    "net/http/cookiejar"
)

const (
    USERNAME_KEY = "username"
    PASSWORD_KEY = "password"
)

var cookieJar *cookiejar.Jar = nil

func ValidateUserCredentials(username string, password string) bool {
    return true
}

func ValidateUserCookie(cookies []*http.Cookie) bool {
    return true
}

func ValidateUser(request *http.Request) (*http.Cookie, bool) {
    if len(request.Cookies()) > 0 {
        validUser := ValidateUserCookie(request.Cookies())
        if validUser {
            return nil, true
        }
    } else if len(request.Form) > 0 {
        validUser := ValidateUserCredentials(request.Form.Get(USERNAME_KEY), request.Form.Get(PASSWORD_KEY))
        if validUser {
            cookie := generateNewCookie()
            cookies := make([]*http.Cookie, 1)
            cookies[0] = cookie
            cookieJar.SetCookies(request.URL, cookies)
            return cookie, true
        }
    }
    return nil, false
}

func generateNewCookie() *http.Cookie {
    return nil
}

func InitializeCookieJar() {
    jar, err := cookiejar.New(nil)
    if err != nil {
        panic(err)
    } else {
        cookieJar = jar
    }
}