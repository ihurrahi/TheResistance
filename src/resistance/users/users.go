package users

import (
    "net/url"
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

func ValidateUserCookie(url *url.URL, cookies []*http.Cookie) bool {
    if len(cookies) == 0 {
        return false
    }

    cookieJarCookies := cookieJar.Cookies(url)
    var cookieJarCookieMap = make(map[string]bool)
    for i := 0; i < len(cookieJarCookies); i++ {
        cookieJarCookieMap[cookieJarCookies[i].String()] = true
    }
    
    for i := 0; i < len(cookies); i++ {
        if !cookieJarCookieMap[cookies[i].String()] {
            return false
        }
    }
    return true
}

func ValidateUser(request *http.Request) (*http.Cookie, bool) {
    if len(request.Cookies()) > 0 {
        validUser := ValidateUserCookie(request.URL, request.Cookies())
        if validUser {
            return nil, true
        }
    }
    request.ParseForm()
    if len(request.Form) > 0 {
        username := request.FormValue(USERNAME_KEY)
        password := request.FormValue(PASSWORD_KEY)
        validUser := ValidateUserCredentials(username, password)
        if validUser {
            cookie := generateNewCookie(username)
            cookies := make([]*http.Cookie, 1)
            cookies[0] = cookie
            cookieJar.SetCookies(request.URL, cookies)
            return cookie, true
        }
    }
    return nil, false
}

func generateNewCookie(username string) *http.Cookie {
    return &http.Cookie{Value: username}
}

func InitializeCookieJar() {
    jar, err := cookiejar.New(nil)
    if err != nil {
        panic(err)
    } else {
        cookieJar = jar
    }
}