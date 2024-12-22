package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

// Message struct to match the server's message structure
type Message struct {
	Method   string      `json:"method"`
	Path     string      `json:"path"`
	Headers  http.Header `json:"headers"`
	Body     string      `json:"body"`
	Response *Response   `json:"response,omitempty"` // Added Response field
}

// Response struct to capture HTTP response details
type Response struct {
	StatusCode int         `json:"statusCode"`
	Headers    http.Header `json:"headers"`
	Body       string      `json:"body"`
}

func HandleClient(port string) {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	serverURL := "ws://horrible-maritsa-attorney-fa65d70c.koyeb.app/ws"
	conn, _, err := websocket.DefaultDialer.Dial(serverURL, nil)
	if err != nil {
		log.Fatalf("Failed to connect to server: %v", err)
	}
	log.Println("Client Connected to WebSocket server!")

	done := make(chan struct{})

	go func() {
		defer close(done)

		// First message will be the session ID
		_, messageBytes, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Error reading session ID: %v", err)
			return
		}
		sessionID := strings.TrimPrefix(string(messageBytes), "Session Id: ")
		log.Printf("Received public url: https://horrible-maritsa-attorney-fa65d70c.koyeb.app/%s", sessionID)

		for {
			_, messageBytes, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
					log.Println("Connection closed normally")
				} else {
					log.Println("Error reading message:", err)
				}
				return
			}

			var message Message
			if err := json.Unmarshal(messageBytes, &message); err != nil {
				log.Printf("Received non-JSON message: %s", string(messageBytes))
				continue
			}

			log.Printf("Received message: %+v", message)

			// Make request to local server
			client := &http.Client{
				Timeout: 10 * time.Second,
			}
			req, err := http.NewRequest(message.Method, fmt.Sprintf("http://localhost:%s%s", port, message.Path), bytes.NewBufferString(message.Body))
			if err != nil {
				log.Printf("Could not create request: %v", err)
				continue
			}

			// Add headers from the message
			for k, v := range message.Headers {
				req.Header.Add(k, v[0])
			}

			resp, err := client.Do(req)
			if err != nil {
				log.Printf("Cannot send request: %v", err)
				continue
			}

			// Read response body
			body, err := io.ReadAll(resp.Body)
			resp.Body.Close()
			if err != nil {
				log.Printf("Error reading response body: %v", err)
				continue
			}

			// Create response message
			message.Response = &Response{
				StatusCode: resp.StatusCode,
				Headers:    resp.Header,
				Body:       string(body),
			}

			// Send response back through WebSocket
			responseBytes, err := json.Marshal(message)
			if err != nil {
				log.Printf("Error marshaling response: %v", err)
				continue
			}

			if err := conn.WriteMessage(websocket.TextMessage, responseBytes); err != nil {
				log.Printf("Error sending response through WebSocket: %v", err)
				continue
			}
		}
	}()

	<-interrupt
	log.Println("Received interrupt signal. Closing connection...")

	err = conn.WriteControl(
		websocket.CloseMessage,
		websocket.FormatCloseMessage(websocket.CloseNormalClosure, "Client closing connection"),
		time.Now().Add(time.Second),
	)
	if err != nil {
		log.Println("Error sending close message:", err)
	}

	<-done
	conn.Close()
	log.Println("Connection closed cleanly")
}
