package main 

import (
    "html/template"
    "net/http"
    "path/filepath"
    "io"
    "os"
    "log"
    "fmt"
    "resistance/users"
)

const (
    TEMPLATE_PATH = "src/resistance/templates"
    INDEX_TEMPLATE = "index.html"
    LOGIN_TEMPLATE = "login.html"
    SIGNUP_TEMPLATE = "signup.html"
    HOME_TEMPLATE = "home.html"
)

var accessLogger *log.Logger

func indexHandler(writer http.ResponseWriter, request *http.Request) {
    accessLogger.Println(request.URL.Path + " was requested")
    renderTemplate(writer, INDEX_TEMPLATE)
}

func loginHandler(writer http.ResponseWriter, request *http.Request) {
    accessLogger.Println(request.URL.Path + " was requested")
    err := request.ParseForm()
    if err != nil {
        accessLogger.Println("Error parsing form values")
    } else {
        cookie, validUser := users.ValidateUser(request)
        for i := 0; i < len(request.Cookies()); i++ {
            accessLogger.Printf("cookie from request : %v", request.Cookies()[i])
        }
        accessLogger.Printf("validUser: %v", validUser)
        if validUser {
            if cookie != nil {
                accessLogger.Println("cookie was created" + cookie.String())
                http.SetCookie(writer, cookie)
            }
            http.Redirect(writer, request, "/home.html", 302)
        } else {
            renderTemplate(writer, LOGIN_TEMPLATE)
        }
    }
}

func signupHandler(writer http.ResponseWriter, request *http.Request) {
    accessLogger.Println(request.URL.Path + " was requested")
    renderTemplate(writer, SIGNUP_TEMPLATE)
}

func homeHandler(writer http.ResponseWriter, request *http.Request) {
    renderTemplate(writer, HOME_TEMPLATE)
}

func renderTemplate(writer io.Writer, name string) {
    filePath := filepath.Join(TEMPLATE_PATH, name)
    templates := template.Must(template.ParseFiles(filePath))
    templates.Execute(writer, nil)
}

func faviconHandler(writer http.ResponseWriter, request *http.Request) {
    // no-op
}

func main() {
    logFile, err := os.OpenFile("logs/accessLog.log", os.O_RDWR|os.O_APPEND, 0666)
    if err != nil {
        fmt.Fprintln(os.Stderr, "Error accessing log file... Abort!")
        return
    }
    defer logFile.Close()
    accessLogger = log.New(logFile, "", log.Ldate|log.Ltime|log.Lshortfile)
    
    users.InitializeCookieJar()
    
    http.HandleFunc("/", indexHandler)
    http.HandleFunc("/favicon.ico", faviconHandler)
    http.HandleFunc("/login.html", loginHandler)
    http.HandleFunc("/signup.html", signupHandler)
    http.HandleFunc("/home.html", homeHandler)
    http.ListenAndServe(":8080", nil)
}

