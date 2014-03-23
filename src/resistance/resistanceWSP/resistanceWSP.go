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

type AcceptUser struct {
	AcceptUser bool
	GameId     int
	UserId     int
}

var (
	DISCONNECT_MESSAGE                                    = []byte("DISCONNECT")
	allListeners           map[*socketio.Conn]*zmq.Socket = make(map[*socketio.Conn]*zmq.Socket)
	connectionSyncChannels map[*socketio.Conn]chan []byte = make(map[*socketio.Conn]chan []byte)
	refreshChan                                           = make(chan bool)
	deleteFinishChan                                      = make(chan bool)
)

// subscribeConnection is meant to run in the background. Waits for a message
// from ZMQ (passed through the channel from receiveZmqMessages) and forwards it
// to the frontend. Also waits for a message from socketio in case a player
// disconnects.
func subscribeConnection(socket *socketio.Conn) {
	messageChannel := connectionSyncChannels[socket]

	for {
		message := <-messageChannel
		if bytes.Equal(DISCONNECT_MESSAGE, message) {
			utils.LogMessage("Disconnect message received.", utils.RWSP_LOG_PATH)
			return
		} else {
			socket.Send(message)
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
		for connection, zmqSocket := range allListeners {
			pollItems = append(pollItems, zmq.PollItem{Socket: zmqSocket, Events: zmq.POLLIN})
			allSockets[zmqSocket] = connection
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
						case connectionSyncChannels[allSockets[pollItem.Socket]] <- rest:
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
			utils.LogMessage("No connections found", utils.RWSP_LOG_PATH)
			time.Sleep(time.Second * 3)
		}
	}
}

// isAcceptUser checks if the reply from the ZMQ socket indicated
// that the user was ok, so the proxy can start a listener on the SUBSCRIBE
// socket for that user. Returns (acceptUser, gameId, userId)
func isAcceptUser(zmqReply []byte) (bool, int, int) {
	var acceptUser AcceptUser
	err := json.Unmarshal(zmqReply, &acceptUser)
	if err != nil {
		return false, 0, 0
	}
	return acceptUser.AcceptUser, acceptUser.GameId, acceptUser.UserId
}

// handleMessage handles a message from the frontend. Basically forwards it
// through ZMQ to the game backend, waits for a reply, then forwards the
// reply to the frontend. If this was a player connect message, and the
// user is accepted, we should start a listener to the SUBSCRIBE socket.
func handleMessage(msg socketio.Message, socket *socketio.Conn, context *zmq.Context) {
	zmqSocket, _ := context.NewSocket(zmq.REQ)
	defer zmqSocket.Close()

	zmqSocket.Connect("tcp://localhost:" + utils.GAME_REP_REQ_PORT)
	utils.LogMessage("WSP connected to port "+utils.GAME_REP_REQ_PORT, utils.RWSP_LOG_PATH)

	zmqSocket.Send([]byte(msg.Data()), 0)
	utils.LogMessage("Sending to game backend", utils.RWSP_LOG_PATH)
	utils.LogMessage(msg.Data(), utils.RWSP_LOG_PATH)

	reply, _ := zmqSocket.Recv(0)
	utils.LogMessage("Reply received", utils.RWSP_LOG_PATH)
	utils.LogMessage(string(reply), utils.RWSP_LOG_PATH)

	accept, gameId, userId := isAcceptUser(reply)
	if accept {
		utils.LogMessage("User accepted: "+strconv.Itoa(userId), utils.RWSP_LOG_PATH)

		// Create the channel to which to communicate with the subscribeConnection go routine
		messageChannel := make(chan []byte)
		connectionSyncChannels[socket] = messageChannel

		// Create the zmq socket to use to subscribe to the appropriate game id
		subSocket, _ := context.NewSocket(zmq.SUB)
		subSocket.Connect("tcp://localhost:" + utils.GAME_PUB_SUB_PORT)
		subSocket.SetSockOptString(zmq.SUBSCRIBE, strconv.Itoa(gameId))
		utils.LogMessage("SUBCRIBER connected to port "+utils.GAME_REP_REQ_PORT+" with filter "+strconv.Itoa(gameId), utils.RWSP_LOG_PATH)
		allListeners[socket] = subSocket

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
		utils.LogMessage("Disconnect!", utils.RWSP_LOG_PATH)

		// Close the subscribe zmq socket.
		if allListeners[c] != nil {
			// Need to synchronize with the receiveZmqMessages go routine
			// because we can't close the socket until we stop polling it.
			go func() {
				<-refreshChan
				allListeners[c].Close()
				delete(allListeners, c)
				deleteFinishChan <- true
			}()
		}

		// Close the channel used to communicate between the zmq sockets
		// and the subscribeConnection go routine.
		if connectionSyncChannels[c] != nil {
			connectionSyncChannels[c] <- DISCONNECT_MESSAGE
			close(connectionSyncChannels[c])
			delete(connectionSyncChannels, c)
		}

		utils.LogMessage("Deleting connection from WSP", utils.RWSP_LOG_PATH)
		// TODO: send message to backend on disconnect
	})

	sio.OnMessage(func(c *socketio.Conn, msg socketio.Message) {
		utils.LogMessage(c.String()+msg.Data(), utils.RWSP_LOG_PATH)
		go handleMessage(msg, c, context)
	})

	go receiveZmqMessages()

	// Start server
	mux := sio.ServeMux()

	if err := http.ListenAndServe(":"+utils.WSP_PORT, mux); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}
