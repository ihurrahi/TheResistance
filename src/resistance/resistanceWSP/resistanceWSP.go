package main

import (
	"bytes"
	"encoding/json"
	zmq "github.com/alecthomas/gozmq"
	"github.com/justinfx/go-socket.io/socketio"
	"log"
	"net/http"
	"resistance/utils"
	"strconv"
	"time"
)

const (
	MESSAGE_KEY               = "message"
	PLAYER_DISCONNECT_MESSAGE = "playerDisconnect"
	USER_COOKIE_KEY           = "userCookie"
	GAME_ID_KEY               = "gameId"
)

type AcceptUser struct {
	AcceptUser bool
	GameId     int
	UserId     int
	UserCookie string
}

type UserInformation struct {
	Socket      *zmq.Socket
	SyncChannel chan []byte
	Cookie      string
	GameId      string
}

var (
	DISCONNECT_MESSAGE                                     = []byte("DISCONNECT")
	userInfos          map[*socketio.Conn]*UserInformation = make(map[*socketio.Conn]*UserInformation)
	refreshChan                                            = make(chan bool)
	deleteFinishChan                                       = make(chan bool)
)

// subscribeConnection is meant to run in the background. Waits for a message
// from ZMQ (passed through the channel from receiveZmqMessages) and forwards it
// to the frontend. Also waits for a message from socketio in case a player
// disconnects.
func subscribeConnection(socket *socketio.Conn) {
	messageChannel := userInfos[socket].SyncChannel

	for {
		message, more := <-messageChannel
		if more {
			if bytes.Equal(DISCONNECT_MESSAGE, message) {
				utils.LogMessage("Disconnect message received.", utils.RWSP_LOG_PATH)
				return
			} else {
				socket.Send(message)
			}
		} else {
			return
		}
	}
}

// receiveZmqMessages is meant to be run in the background. A single go routine
// running this method will go through all the zmq sockets and try and receive
// from them by polling.
func receiveZmqMessages() {
	for {
		pollItems := make([]zmq.PollItem, 0)
		allSockets := make(map[*zmq.Socket]*socketio.Conn)
		for connection, userInfo := range userInfos {
			pollItems = append(pollItems, zmq.PollItem{Socket: userInfo.Socket, Events: zmq.POLLIN})
			allSockets[userInfo.Socket] = connection
		}

		if len(pollItems) > 0 {
			_, _ = zmq.Poll(pollItems, time.Second*3)
			for _, pollItem := range pollItems {
				// Check all items to see if we receive anything
				if pollItem.REvents == zmq.POLLIN {
					multiPartMessage, err1 := pollItem.Socket.RecvMultipart(0)
					if err1 == nil {
						game := multiPartMessage[0]
						rest := multiPartMessage[1]
						utils.LogMessage("got message on sub socket for game "+string(game)+":"+string(rest), utils.RWSP_LOG_PATH)
						// In case we already closed the channel, we don't want to block on this
						// If the channel is closed, the message can't go anywhere anyways.
						select {
						case userInfos[allSockets[pollItem.Socket]].SyncChannel <- rest:
						default:
						}
					}
				}
			}
			// Allow any disconnections to go through, so we don't error out
			// during polling, since we might keep polling closed sockets
			continueWaitForRefresh := true
			for continueWaitForRefresh {
				select {
				case refreshChan <- true:
					<-deleteFinishChan
				default:
					continueWaitForRefresh = false
				}
			}
		} else {
			time.Sleep(time.Second * 3)
		}
	}
}

// isAcceptUser checks if the reply from the ZMQ socket indicated
// that the user was ok, so the proxy can start a listener on the SUBSCRIBE
// socket for that user. Returns (acceptUser, gameId, userId)
func isAcceptUser(zmqReply []byte) *AcceptUser {
	var acceptUser AcceptUser
	err := json.Unmarshal(zmqReply, &acceptUser)
	if err != nil {
		return &AcceptUser{false, 0, 0, ""}
	}
	return &acceptUser
}

// sendMessageToBackend performs the actual sending of the message
// over ZMQ using the given context
func sendMessageToBackend(msg string, context *zmq.Context) []byte {
	zmqSocket, _ := context.NewSocket(zmq.REQ)
	defer zmqSocket.Close()

	zmqSocket.Connect("tcp://localhost:" + utils.GAME_REP_REQ_PORT)
	utils.LogMessage("WSP connected to port "+utils.GAME_REP_REQ_PORT, utils.RWSP_LOG_PATH)

	zmqSocket.Send([]byte(msg), 0)
	utils.LogMessage("Sending to game backend", utils.RWSP_LOG_PATH)
	utils.LogMessage(msg, utils.RWSP_LOG_PATH)

	reply, _ := zmqSocket.Recv(0)
	utils.LogMessage("Reply received", utils.RWSP_LOG_PATH)
	utils.LogMessage(string(reply), utils.RWSP_LOG_PATH)

	return reply
}

// handleMessage handles a message from the frontend. Basically forwards it
// through ZMQ to the game backend, waits for a reply, then forwards the
// reply to the frontend. If this was a player connect message, and the
// user is accepted, we should start a listener to the SUBSCRIBE socket.
func handleMessage(msg socketio.Message, socket *socketio.Conn, context *zmq.Context) {
	reply := sendMessageToBackend(msg.Data(), context)

	acceptUser := isAcceptUser(reply)
	if acceptUser.AcceptUser {
		utils.LogMessage("User accepted: "+strconv.Itoa(acceptUser.UserId), utils.RWSP_LOG_PATH)

		gameId := strconv.Itoa(acceptUser.GameId)

		// Create the channel to which to communicate with the subscribeConnection go routine
		messageChannel := make(chan []byte)

		// Create the zmq socket to use to subscribe to the appropriate game id
		subSocket, _ := context.NewSocket(zmq.SUB)
		subSocket.Connect("tcp://localhost:" + utils.GAME_PUB_SUB_PORT)
		subSocket.SetSockOptString(zmq.SUBSCRIBE, gameId)
		utils.LogMessage("SUBCRIBER connected to port "+utils.GAME_REP_REQ_PORT+" with filter "+strconv.Itoa(acceptUser.GameId), utils.RWSP_LOG_PATH)

		// Keep in memory all the necessary information about the connection
		userInfos[socket] = &UserInformation{
			Socket:      subSocket,
			SyncChannel: messageChannel,
			Cookie:      acceptUser.UserCookie,
			GameId:      gameId}

		go subscribeConnection(socket)
	}

	socket.Send(reply)
	utils.LogMessage("Sent to frontend", utils.RWSP_LOG_PATH)
}

func main() {
	// Setup ZMQ
	context, _ := zmq.NewContext()
	defer context.Close()

	// Setup Socket.IO
	config := socketio.DefaultConfig
	config.Origins = []string{"*:80"}
	sio := socketio.NewSocketIO(&config)

	sio.OnConnect(func(c *socketio.Conn) {
	})

	sio.OnDisconnect(func(c *socketio.Conn) {
		utils.LogMessage("Disconnect received", utils.RWSP_LOG_PATH)

		if userInfos[c] != nil {
			// Close the subscribe zmq socket.
			if userInfos[c].Socket != nil {

				// Need to synchronize with the receiveZmqMessages go routine
				// because we can't close the socket until we stop polling it.
				go func(socket *zmq.Socket) {
					<-refreshChan
					socket.Close()
					deleteFinishChan <- true
				}(userInfos[c].Socket)

				utils.LogMessage("Removing ZMQ socket from state", utils.RWSP_LOG_PATH)
			}

			// Close the channel used to communicate between the zmq sockets
			// and the subscribeConnection go routine.
			if userInfos[c].SyncChannel != nil {
				go func(channel chan []byte) {
					<-refreshChan
					userInfos[c].SyncChannel <- DISCONNECT_MESSAGE
					close(channel)
					deleteFinishChan <- true
				}(userInfos[c].SyncChannel)
			}
			utils.LogMessage("Removing sync channel from state", utils.RWSP_LOG_PATH)

			if userInfos[c].Cookie != "" {
				rawMessage := make(map[string]interface{})
				rawMessage[MESSAGE_KEY] = PLAYER_DISCONNECT_MESSAGE
				rawMessage[USER_COOKIE_KEY] = userInfos[c].Cookie
				rawMessage[GAME_ID_KEY] = userInfos[c].GameId
				message, err := json.Marshal(rawMessage)
				if err == nil {
					go sendMessageToBackend(string(message), context)
				} else {
					utils.LogMessage("Could not send message to game backend:"+err.Error(), utils.RWSP_LOG_PATH)
				}
			}
			utils.LogMessage("Removing cookie from state", utils.RWSP_LOG_PATH)

			// Need to wait for a signal to refresh before really deleting it
			go func(connection *socketio.Conn) {
				<-refreshChan
				delete(userInfos, connection)
				deleteFinishChan <- true
			}(c)
		}

		utils.LogMessage("Finished deleting connection from WSP", utils.RWSP_LOG_PATH)
	})

	sio.OnMessage(func(c *socketio.Conn, msg socketio.Message) {
		utils.LogMessage("Received message for "+c.String()+" with data:"+msg.Data(), utils.RWSP_LOG_PATH)
		go handleMessage(msg, c, context)
	})

	go receiveZmqMessages()

	// Start server
	mux := sio.ServeMux()

	if err := http.ListenAndServe(":"+utils.WSP_PORT, mux); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}
