package main

import (
	"html/template"
	"io"
	"net/http"
	"path/filepath"
	"resistance/game"
	"resistance/users"
	"resistance/utils"
	"strconv"
)

const (
	TEMPLATE_PATH        = "src/resistance/frontend"
	INDEX_TEMPLATE       = "index.html"
	LOGIN_TEMPLATE       = "login.html"
	SIGNUP_TEMPLATE      = "signup.html"
	HOME_TEMPLATE        = "home.html"
	CREATE_GAME_TEMPLATE = "create.html"
	LOBBY_TEMPLATE       = "lobby.html"
	HISTORY_TEMPLATE     = "history.html"
	GAME_TEMPLATE        = "game.html"
)

const (
	TITLE_KEY   = "title"
	HOST_ID_KEY = "host"
)

func faviconHandler(writer http.ResponseWriter, request *http.Request) {
	// no-op
}

func indexHandler(writer http.ResponseWriter, request *http.Request) {
	utils.LogMessage(request.URL.Path+" was requested", utils.RHTTP_LOG_PATH)

	// If this person has a valid cookie, send them to their homepage
	_, validUser := users.ValidateUserCookie(request.Cookies())
	if validUser {
		utils.LogMessage("Valid User, redirecting to /home.html", utils.RHTTP_LOG_PATH)
		http.Redirect(writer, request, "/home.html", 302)
	}

	renderTemplate(writer, INDEX_TEMPLATE, make(map[string]string))
}

func loginHandler(writer http.ResponseWriter, request *http.Request) {
	utils.LogMessage(request.URL.Path+" was requested", utils.RHTTP_LOG_PATH)

	// If this person has a valid cookie, send them to their homepage instead
	_, validUser := users.ValidateUserCookie(request.Cookies())
	if validUser {
		utils.LogMessage("Valid User, redirecting to /home.html", utils.RHTTP_LOG_PATH)
		http.Redirect(writer, request, "/home.html", 302)
	}

	err := request.ParseForm()
	if err != nil {
		utils.LogMessage("Error parsing form values", utils.RHTTP_LOG_PATH)
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
	utils.LogMessage(request.URL.Path+" was requested", utils.RHTTP_LOG_PATH)

	err := request.ParseForm()
	if err != nil {
		utils.LogMessage("Error parsing form values", utils.RHTTP_LOG_PATH)
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
	utils.LogMessage(request.URL.Path+" was requested", utils.RHTTP_LOG_PATH)

	user := requiresLogin(writer, request)

	renderTemplate(writer, HOME_TEMPLATE, user)
}

func createGameHandler(writer http.ResponseWriter, request *http.Request) {
	utils.LogMessage(request.URL.Path+" was requested", utils.RHTTP_LOG_PATH)

	user := requiresLogin(writer, request)

	err := request.ParseForm()
	if err != nil {
		utils.LogMessage("Error parsing form values", utils.RHTTP_LOG_PATH)
	} else if len(request.Form) > 0 {
		newGame := game.NewGame(request.FormValue(TITLE_KEY), request.FormValue(HOST_ID_KEY))
		if newGame != nil {
			gameId := newGame.GameId
			http.Redirect(writer, request, "/game.html?gameId="+strconv.Itoa(gameId), 302)
		} else {
			renderTemplate(writer, CREATE_GAME_TEMPLATE, user)
		}
	}

	renderTemplate(writer, CREATE_GAME_TEMPLATE, user)
}

func lobbyHandler(writer http.ResponseWriter, request *http.Request) {
	utils.LogMessage(request.URL.Path+" was requested", utils.RHTTP_LOG_PATH)

	user := requiresLogin(writer, request)

	renderTemplate(writer, LOBBY_TEMPLATE, user)
}

func historyHandler(writer http.ResponseWriter, request *http.Request) {
	utils.LogMessage(request.URL.Path+" was requested", utils.RHTTP_LOG_PATH)

	user := requiresLogin(writer, request)

	renderTemplate(writer, HISTORY_TEMPLATE, user)
}

func gameHandler(writer http.ResponseWriter, request *http.Request) {
	utils.LogMessage(request.URL.Path+" was requested", utils.RHTTP_LOG_PATH)

	user := requiresLogin(writer, request)

	err := request.ParseForm()
	if err != nil {
		utils.LogMessage(err.Error(), utils.RHTTP_LOG_PATH)
	} else if len(request.Form) > 0 {
		gameInfo, err := game.IsValidGame(request.FormValue("gameId"), user)
		if err == nil {
			renderTemplate(writer, GAME_TEMPLATE, gameInfo)
		} else {
			// TODO: how do i redirect to home and pass in an error message?
			writer.Write([]byte(err.Error()))
		}
	} else {
		http.Redirect(writer, request, "/home.html", 302)
	}
}

func renderTemplate(writer io.Writer, name string, parameters interface{}) {
	filePath := filepath.Join(TEMPLATE_PATH, name)
	templates := template.Must(template.ParseFiles(filePath))
	templates.Execute(writer, parameters)
}

func requiresLogin(writer http.ResponseWriter, request *http.Request) *users.User {
	// If this person has an invalid cookie, send them to the login page instead
	user, validUser := users.ValidateUserCookie(request.Cookies())
	if !validUser {
		utils.LogMessage("Invalid User, redirecting to /login.html", utils.RHTTP_LOG_PATH)
		http.Redirect(writer, request, "/login.html", 302)
	}
	return user
}

func main() {

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
	http.Handle("/game.js", http.FileServer(http.Dir("src/resistance/frontend")))
	http.Handle("/game.css", http.FileServer(http.Dir("src/resistance/frontend")))

	utils.LogMessage("Starting TheResistance HTTP Server...", utils.RHTTP_LOG_PATH)

	http.ListenAndServe(":8080", nil)
}
