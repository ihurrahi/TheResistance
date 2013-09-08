package main 

import (
    "html/template"
    "net/http"
    "path/filepath"
    "io"
    "os"
    "log"
    "fmt"
    "errors"
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
    
    // If this person has a valid cookie, send them to their homepage
    _, validUser := users.ValidateUserCookie(request)
    if validUser {
        accessLogger.Println("Valid User, redirecting to /home.html")
        http.Redirect(writer, request, "/home.html", 302)
    }
    
    renderTemplate(writer, INDEX_TEMPLATE, make(map[string]string))
}

func loginHandler(writer http.ResponseWriter, request *http.Request) {
    accessLogger.Println(request.URL.Path + " was requested")
    
    // If this person has a valid cookie, send them to their homepage instead
    _, validUser := users.ValidateUserCookie(request)
    if validUser {
        accessLogger.Println("Valid User, redirecting to /home.html")
        http.Redirect(writer, request, "/home.html", 302)
    }
    
    err := request.ParseForm()
    if err != nil {
        accessLogger.Println("Error parsing form values")
    } else if len(request.Form) > 0 {
        cookie, validUser := users.ValidateUser(request)
        if validUser {
            http.SetCookie(writer, cookie)
            http.Redirect(writer, request, "/home.html", 302)
        } else {
            invalidUser := make(map[string]string)
            invalidUser["Error"] = "Username and password did not match."
            renderTemplate(writer, LOGIN_TEMPLATE, invalidUser)
            return
        }
    }
    
    renderTemplate(writer, LOGIN_TEMPLATE, make(map[string]string))
}

func signupHandler(writer http.ResponseWriter, request *http.Request) {
    accessLogger.Println(request.URL.Path + " was requested")
    renderTemplate(writer, SIGNUP_TEMPLATE, make(map[string]string))
}

func homeHandler(writer http.ResponseWriter, request *http.Request) {
    accessLogger.Println(request.URL.Path + " was requested")
    
    // If this person has an invalid cookie, send them to the login page instead
    user, validUser := users.ValidateUserCookie(request)
    if !validUser {
        accessLogger.Println("Invalid User, redirecting to /login.html")
        http.Redirect(writer, request, "/login.html", 302)
    }
    
    renderTemplate(writer, HOME_TEMPLATE, user)
}

func renderTemplate(writer io.Writer, name string, parameters interface{}) {
    filePath := filepath.Join(TEMPLATE_PATH, name)
    templates := template.Must(template.ParseFiles(filePath))
    templates.Execute(writer, parameters)
}

func faviconHandler(writer http.ResponseWriter, request *http.Request) {
    // no-op
}

func createLogger(filename string) (*log.Logger, error) {
    logFile, err := os.OpenFile(filename, os.O_RDWR|os.O_APPEND, 0666)
    if err != nil {
        logFile, err = os.Create(filename)
        if err != nil {
            return nil, errors.New("Error accessing access log file... Abort!") 
        }
    }
    logger := log.New(logFile, "", log.Ldate|log.Ltime|log.Lshortfile)
    return logger, nil
}

func main() {
    var err error
    accessLogger, err = createLogger("logs/accessLog.log")
    if err != nil {
        fmt.Println(err)
        return
    }
    accessLogger.Println("Starting TheResistance")
    
    userLogger, err := createLogger("logs/userLog.log")
    if err != nil {
        fmt.Println(err)
        return
    }
    users.Initialize(userLogger)
    
    http.HandleFunc("/", indexHandler)
    http.HandleFunc("/favicon.ico", faviconHandler)
    http.HandleFunc("/login.html", loginHandler)
    http.HandleFunc("/signup.html", signupHandler)
    http.HandleFunc("/home.html", homeHandler)
    http.ListenAndServe(":8080", nil)
}
