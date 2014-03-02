package main

import (
	"encoding/json"
	zmq "github.com/alecthomas/gozmq"
	"html/template"
	"io"
	"net/http"
	"path/filepath"
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

var zmqContext *zmq.Context

func faviconHandler(writer http.ResponseWriter, request *http.Request) {
	// no-op
}

func indexHandler(writer http.ResponseWriter, request *http.Request) {
	utils.LogMessage(request.URL.Path+" was requested", utils.RHTTP_LOG_PATH)

	// If this person has a valid cookie, send them to their homepage
	user := users.ValidateUserCookie(request.Cookies())
	if user.IsValidUser() {
		utils.LogMessage("Valid User, redirecting to /home.html", utils.RHTTP_LOG_PATH)
		http.Redirect(writer, request, "/home.html", 302)
	}

	renderTemplate(writer, INDEX_TEMPLATE, make(map[string]string))
}

func loginHandler(writer http.ResponseWriter, request *http.Request) {
	utils.LogMessage(request.URL.Path+" was requested", utils.RHTTP_LOG_PATH)

	// If this person has a valid cookie, send them to their homepage instead
	user := users.ValidateUserCookie(request.Cookies())
	if user.IsValidUser() {
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
		data := make(map[string]interface{})
		data["title"] = request.FormValue(TITLE_KEY)
		data["hostId"] = request.FormValue(HOST_ID_KEY)
		cookie, err := request.Cookie(users.COOKIE_NAME)
		if err == nil {
			data["userCookie"] = cookie.Name + "=" + cookie.Value
		}
		data["gameId"] = "-1"
		parsedReply := sendToGameBackend("createGame", data)
		newGameId, ok := parsedReply["gameId"].(float64)
		//newGame := game.NewGame(request.FormValue(TITLE_KEY), request.FormValue(HOST_ID_KEY))
		if ok && int(newGameId) > 0 {
			http.Redirect(writer, request, "/game.html?gameId="+strconv.Itoa(int(newGameId)), 302)
		} else {
			renderTemplate(writer, CREATE_GAME_TEMPLATE, user)
		}
	}

	renderTemplate(writer, CREATE_GAME_TEMPLATE, user)
}

func lobbyHandler(writer http.ResponseWriter, request *http.Request) {
	utils.LogMessage(request.URL.Path+" was requested", utils.RHTTP_LOG_PATH)

	user := requiresLogin(writer, request)

	if user.IsValidUser() {
		gameInfo := sendToGameBackend("getAllGames", make(map[string]interface{}))
		renderTemplate(writer, LOBBY_TEMPLATE, gameInfo)
	}
}

func historyHandler(writer http.ResponseWriter, request *http.Request) {
	utils.LogMessage(request.URL.Path+" was requested", utils.RHTTP_LOG_PATH)

	user := requiresLogin(writer, request)

	renderTemplate(writer, HISTORY_TEMPLATE, user)
}

func gameHandler(writer http.ResponseWriter, request *http.Request) {
	utils.LogMessage(request.URL.Path+" was requested", utils.RHTTP_LOG_PATH)

	user := requiresLogin(writer, request)

	if user.IsValidUser() {
		err := request.ParseForm()
		if err != nil {
			utils.LogMessage(err.Error(), utils.RHTTP_LOG_PATH)
		} else if len(request.Form) > 0 {
			data := make(map[string]interface{})
			data["gameId"] = request.FormValue("gameId")
			cookie, err := request.Cookie(users.COOKIE_NAME)
			if err == nil {
				data["userCookie"] = cookie.Name + "=" + cookie.Value
			}
			gameInfo := sendToGameBackend("isValidGame", data)
			if gameInfo["error"] == nil {
				utils.LogMessage(gameInfo["GameTitle"].(string), utils.RESISTANCE_LOG_PATH)
				renderTemplate(writer, GAME_TEMPLATE, gameInfo)
			} else {
				// TODO: how do i redirect to home and pass in an error message?
				writer.Write([]byte(gameInfo["error"].(string)))
			}
		} else {
			http.Redirect(writer, request, "/home.html", 302)
		}
	}
}

func renderTemplate(writer io.Writer, name string, parameters interface{}) {
	filePath := filepath.Join(TEMPLATE_PATH, name)
	templates := template.Must(template.ParseFiles(filePath))
	templates.Execute(writer, parameters)
}

func requiresLogin(writer http.ResponseWriter, request *http.Request) *users.User {
	// If this person has an invalid cookie, send them to the login page instead
	user := users.ValidateUserCookie(request.Cookies())
	if !user.IsValidUser() {
		utils.LogMessage("Invalid User, redirecting to /login.html", utils.RHTTP_LOG_PATH)
		http.Redirect(writer, request, "/login.html", 302)
	}
	return user
}

func sendToGameBackend(msg string, data map[string]interface{}) map[string]interface{} {
	zmqSocket, _ := zmqContext.NewSocket(zmq.REQ)
	defer zmqSocket.Close()

	zmqSocket.Connect("tcp://localhost:" + utils.GAME_REP_REQ_PORT)
	utils.LogMessage("HTTP connected to port "+utils.GAME_REP_REQ_PORT, utils.RHTTP_LOG_PATH)

	data["message"] = msg
	newData, _ := json.Marshal(data)
	zmqSocket.Send(newData, 0)
	utils.LogMessage("Sending to game backend", utils.RHTTP_LOG_PATH)
	utils.LogMessage(string(newData), utils.RHTTP_LOG_PATH)

	reply, _ := zmqSocket.Recv(0)
	utils.LogMessage("Reply received", utils.RHTTP_LOG_PATH)
	utils.LogMessage(string(reply), utils.RHTTP_LOG_PATH)

	var parsedReply map[string]interface{}
	json.Unmarshal(reply, &parsedReply)

	return parsedReply
}

func main() {
	zmqContext, _ = zmq.NewContext()
	defer zmqContext.Close()

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
