package main 

import (
    "html/template"
    "net/http"
    "path/filepath"
    "io"
    "resistance/logger"
    "resistance/users"
)

const (
    TEMPLATE_PATH = "src/resistance/templates"
    INDEX_TEMPLATE = "index.html"
    LOGIN_TEMPLATE = "login.html"
    SIGNUP_TEMPLATE = "signup.html"
    HOME_TEMPLATE = "home.html"
)

var accessLogger *logger.Logger = nil

func indexHandler(writer http.ResponseWriter, request *http.Request) {
    accessLogger.LogMessage(request.URL.Path + " was requested")
    renderTemplate(writer, INDEX_TEMPLATE)
}

func loginHandler(writer http.ResponseWriter, request *http.Request) {
    accessLogger.LogMessage(request.URL.Path + " was requested")
    err := request.ParseForm()
    if err != nil {
        accessLogger.LogMessage("Error parsing form values")
    } else {
        cookie, validUser := users.ValidateUser(request)
        if validUser {
            if cookie != nil {
                writer.Header().Set("Set-Cookie", cookie.String())
            }
            http.Redirect(writer, request, "/home.html", 302)
        } else {
            renderTemplate(writer, LOGIN_TEMPLATE)
        }
    }
}

func signupHandler(writer http.ResponseWriter, request *http.Request) {
    accessLogger.LogMessage(request.URL.Path + " was requested")
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
    accessLogger = logger.InitializeLogger("accessLog.log")
    users.InitializeCookieJar()
    
    http.HandleFunc("/", indexHandler)
    http.HandleFunc("/favicon.ico", faviconHandler)
    http.HandleFunc("/login.html", loginHandler)
    http.HandleFunc("/signup.html", signupHandler)
    http.HandleFunc("/home.html", homeHandler)
    http.ListenAndServe(":8080", nil)
}

