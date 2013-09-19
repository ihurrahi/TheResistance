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
    "strconv"
    "database/sql"
    _ "github.com/go-sql-driver/mysql"
    "resistance/users"
    "resistance/game"
)

const (
    RESISTANCE_LOG_PATH = "logs/resistance.log"
    USER_LOG_PATH = "logs/userLog.log"
    GAME_LOG_PATH = "logs/gameLog.log"
    TEMPLATE_PATH = "src/resistance/templates"
    INDEX_TEMPLATE = "index.html"
    LOGIN_TEMPLATE = "login.html"
    SIGNUP_TEMPLATE = "signup.html"
    HOME_TEMPLATE = "home.html"
    CREATE_GAME_TEMPLATE = "create.html"
    LOBBY_TEMPLATE = "lobby.html"
    HISTORY_TEMPLATE = "history.html"
    GAME_TEMPLATE = "game.html"
)

var resistanceLogger *log.Logger

func faviconHandler(writer http.ResponseWriter, request *http.Request) {
    // no-op
}

func indexHandler(writer http.ResponseWriter, request *http.Request) {
    resistanceLogger.Println(request.URL.Path + " was requested")
    
    // If this person has a valid cookie, send them to their homepage
    _, validUser := users.ValidateUserCookie(request)
    if validUser {
        resistanceLogger.Println("Valid User, redirecting to /home.html")
        http.Redirect(writer, request, "/home.html", 302)
    }
    
    renderTemplate(writer, INDEX_TEMPLATE, make(map[string]string))
}

func loginHandler(writer http.ResponseWriter, request *http.Request) {
    resistanceLogger.Println(request.URL.Path + " was requested")
    
    // If this person has a valid cookie, send them to their homepage instead
    _, validUser := users.ValidateUserCookie(request)
    if validUser {
        resistanceLogger.Println("Valid User, redirecting to /home.html")
        http.Redirect(writer, request, "/home.html", 302)
    }
    
    err := request.ParseForm()
    if err != nil {
        resistanceLogger.Println("Error parsing form values")
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
    resistanceLogger.Println(request.URL.Path + " was requested")
    
    err := request.ParseForm()
    if err != nil {
        resistanceLogger.Println("Error parsing form values")
    } else if len(request.Form) > 0 {
        hasSignUpError, errorMessage := users.UserSignUp(request)
        if hasSignUpError {
            signUpErrors := make(map[string]string)
            signUpErrors["Error"] = errorMessage
            renderTemplate(writer, SIGNUP_TEMPLATE, signUpErrors)
            return
        } else {
            // TODO: redirect to login page with success message
            http.Redirect(writer, request, "/login.html", 302)
        }
    }
    
    renderTemplate(writer, SIGNUP_TEMPLATE, make(map[string]string))
}

func homeHandler(writer http.ResponseWriter, request *http.Request) {
    resistanceLogger.Println(request.URL.Path + " was requested")
    
    user := requiresLogin(writer, request)
    
    renderTemplate(writer, HOME_TEMPLATE, user)
}

func createGameHandler(writer http.ResponseWriter, request *http.Request) {
    resistanceLogger.Println(request.URL.Path + " was requested")
    
    user := requiresLogin(writer, request)
    
    err := request.ParseForm()
    if err != nil {
        resistanceLogger.Println("Error parsing form values")
    } else if len(request.Form) > 0 {
        gameId, err := game.CreateGame(request)
        if err == nil {
            http.Redirect(writer, request, "/game.html?gameId=" + strconv.FormatInt(gameId, 10), 302)
        } else {
            renderTemplate(writer, CREATE_GAME_TEMPLATE, user)
        }
    }
    
    renderTemplate(writer, CREATE_GAME_TEMPLATE, user)
}

func lobbyHandler(writer http.ResponseWriter, request *http.Request) {
    resistanceLogger.Println(request.URL.Path + " was requested")
    
    user := requiresLogin(writer, request)
    
    renderTemplate(writer, LOBBY_TEMPLATE, user)
}

func historyHandler(writer http.ResponseWriter, request *http.Request) {
    resistanceLogger.Println(request.URL.Path + " was requested")
    
    user := requiresLogin(writer, request)
    
    renderTemplate(writer, HISTORY_TEMPLATE, user)
}

func gameHandler(writer http.ResponseWriter, request *http.Request) {
    resistanceLogger.Println(request.URL.Path + " was requested")
    
    _ = requiresLogin(writer, request)
    
    err := request.ParseForm()
    if err != nil {
        resistanceLogger.Println(err.Error())
    }
    gameInfo := make(map[string]string)
    gameInfo["GameId"] = request.FormValue("gameId")
    renderTemplate(writer, GAME_TEMPLATE, gameInfo)
}

func renderTemplate(writer io.Writer, name string, parameters interface{}) {
    filePath := filepath.Join(TEMPLATE_PATH, name)
    templates := template.Must(template.ParseFiles(filePath))
    templates.Execute(writer, parameters)
}

func requiresLogin(writer http.ResponseWriter, request *http.Request) *users.User {
    // If this person has an invalid cookie, send them to the login page instead
    user, validUser := users.ValidateUserCookie(request)
    if !validUser {
        resistanceLogger.Println("Invalid User, redirecting to /login.html")
        http.Redirect(writer, request, "/login.html", 302)
    }
    return user
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
    
    resistanceLogger, err = createLogger(RESISTANCE_LOG_PATH)
    if err != nil {
        fmt.Println(err)
        return
    }
    
    userLogger, err := createLogger(USER_LOG_PATH)
    if err != nil {
        fmt.Println(err)
        return
    }
    
    gameLogger, err := createLogger(GAME_LOG_PATH)
    if err != nil {
        fmt.Println(err)
        return
    }
    
    db, err := sql.Open("mysql", "resistance:resistance@unix(/var/run/mysql/mysql.sock)/resistance")
    if err != nil {
        resistanceLogger.Fatalln(err.Error())
    }
    err = db.Ping()
    if err != nil {
        resistanceLogger.Fatalln(err.Error())
    }
    resistanceLogger.Println("Database connection successful")
    
    users.Initialize(userLogger, db)
    game.Initialize(gameLogger, db)
    
    http.HandleFunc("/", indexHandler)
    http.HandleFunc("/favicon.ico", faviconHandler)
    http.HandleFunc("/login.html", loginHandler)
    http.HandleFunc("/signup.html", signupHandler)
    http.HandleFunc("/home.html", homeHandler)
    http.HandleFunc("/create.html", createGameHandler)
    http.HandleFunc("/lobby.html", lobbyHandler)
    http.HandleFunc("/history.html", historyHandler)
    http.HandleFunc("/game.html", gameHandler)
    http.Handle("/socket.io.js", http.FileServer(http.Dir("src/github.com/justinfx/go-socket.io/bin/www/vendor/socket.io-client")))
    
    resistanceLogger.Println("Starting TheResistance")
    
    http.ListenAndServe(":8080", nil)
}
