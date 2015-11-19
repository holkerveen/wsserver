// Simple server for peer-to-peer signaling
//
// This server has been written specifically for use as a signaling
// channel to set up WebRTC connections. It is set up in a full mesh
// topology; that is: a message sent to the server by a client is
// forwarded to all other connected peers in the same server
//
// Connect procedure is as follows: A master htmldocument page
// requests a uniqe channel ID and displays is on the page. It will
// connect to said channel. Any number of client documents can use
// the same channel id to connect as well.
//
// After that, signaling can start. Any message sent by any client
// will be pushed to all other clients.
package wssrv

import (
	"net/http"
	"golang.org/x/net/websocket"
	"fmt"
	"math/rand"
	"log"
	"io"
)

// Request describes the structure of requests sent by the clients
type Request struct {
	Command string `json:"cmd"`
	Channel string `json:"channel"`
	Data string `json:"data"`
}

// RequestChannelIdResponse is a response type for the channelId
// request
type RequestChannelIdResponse struct {
	ChannelId string `json:"cid"`
}

// Variable containing all channels and associated connections
var connections = make(map[string][]*websocket.Conn)

// letters define the set of characters used to generate a channel
// id.
const letters = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
const generateChannelIdLength = 4
const generateMaxTries = 20

// generateChannelId attempts to generate a new channel id.
// It generates a random string of the available charachters, and
// tries again if the generated channel id has already been used.
func generateChannelId()(string,bool) {
	id := make([]byte,generateChannelIdLength)
	for try:=0; try<generateMaxTries; try++ {
		for i := range id {
			id[i] = letters[rand.Intn(len(letters))]
		}
		if _,exists:=connections[string(id)]; !exists {
			return string(id),false
		}
	}
	return "",true
}

// EchoServer contains the main server loop
func WsHandler(ws *websocket.Conn) {
	log.Printf("%v connected to server",ws.Request().RemoteAddr)
	addr := ws.Request().RemoteAddr

	// Cleanup
	defer func() {
		// TODO: remove conn from channel
		if err:=ws.Close(); err != nil {
			log.Panicf("%v cleanup could not close connecteion: %v",addr,err.Error())
		}
	}()

	var data Request
	for {
		err := websocket.JSON.Receive(ws,&data)
		switch {
		case err == io.EOF:
				log.Printf("%v disconnected",addr)
				return
		default:
				panic(err.Error())
		case err == nil:
		}

		switch data.Command {
		case "":
			fmt.Printf("empty message\n")
		case "requestChannelId":
			log.Printf("%v requestChannelId",addr)
			channelId, err := generateChannelId()
			if err {
				log.Panicf("%v could not generate channel id",addr)
				return
			}
			connections[channelId] = []*websocket.Conn{}
			response := RequestChannelIdResponse{
				ChannelId:channelId,
			}
			websocket.JSON.Send(ws,response)
		case "connectChannel":
			log.Printf("%v connectChannel",addr)
			// TODO: disconnect from previous channel
			connections[data.Channel] = append(connections[data.Channel],ws)
		case "send":
			log.Printf("%v send",addr)
			/* Iterate all connections */
			for _,conn := range connections[data.Channel] {
				if conn != ws {
					websocket.JSON.Send(conn,data)
				}
			}
		default:
			log.Panicf("%v Unhandled message\n%v",addr,data)
		}
	}
}


// main is the program entry point. Execution starts here.
func main() {
	fmt.Printf("wssrv start\n")

	http.Handle("/", websocket.Handler(WsHandler))
	err := http.ListenAndServe(":8000",nil)
	if err != nil {
		panic("ListenAndServe: "+err.Error())
	}
}
