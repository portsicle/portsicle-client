package client

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/shamaton/msgpack/v2"
)

// Message struct to match the server's message structure
type Message struct {
	Method   string      `json:"method"`
	Path     string      `json:"path"`
	Headers  http.Header `json:"headers"`
	Body     string      `json:"body"`
	Response *Response   `json:"response,omitempty"`
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

	serverURL := "wss://portsicle.koyeb.app/ws"
	conn, _, err := websocket.DefaultDialer.Dial(serverURL, nil)
	if err != nil {
		log.Fatalf("Failed to connect to remote server: %v", err)
	}
	log.Println("Connected to remote server.")

	done := make(chan struct{})

	// Initial read deadline
	err = conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	if err != nil {
		log.Printf("Error setting initial read deadline: %v", err)
		return
	}

	// Setup pong handler to reset the read deadline
	conn.SetPongHandler(func(string) error {
		log.Println("server: pong")
		err := conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		if err != nil {
			log.Printf("Error setting read deadline: %v", err)
			return err
		}
		return nil
	})

	// this Keepalive ping will avoid the 1006 abnormal connection closure: EOF
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				log.Println("client: ping")
				err := conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(10*time.Second))
				if err != nil {
					log.Printf("Error sending ping: %v", err)
					return
				}
			}
		}
	}()

	go func() {
		defer close(done)

		// First message from server will be the public URL aka session ID
		_, messageBytes, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Error reading session ID: %v", err)
			return
		}
		sessionID := strings.TrimPrefix(string(messageBytes), "Session Id: ")
		log.Printf("Your public url: https://portsicle.koyeb.app/%s", sessionID)

		for {
			_, messageBytes, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
					log.Println("Disconnectd from remote server.")
				} else {
					log.Println("Error reading message:", err)
				}
				return
			}

			// Reset read deadline after successful read
			err = conn.SetReadDeadline(time.Now().Add(60 * time.Second))
			if err != nil {
				log.Printf("Error resetting read deadline: %v", err)
				return
			}

			var message Message
			if err := msgpack.Unmarshal(messageBytes, &message); err != nil {
				log.Printf("Received invaid response from server: %s", string(messageBytes))
				continue
			}

			// log.Printf("Received message: %+v", message)

			// Make request to local server
			client := &http.Client{
				Timeout: 10 * time.Second,
			}
			req, err := http.NewRequest(message.Method, fmt.Sprintf("http://localhost:%s%s", port, message.Path), bytes.NewBufferString(message.Body))
			if err != nil {
				log.Printf("Could not create request for local server: %v", err)
				continue
			}

			// Attach headers from the message
			for k, v := range message.Headers {
				req.Header.Add(k, v[0])
			}

			resp, err := client.Do(req)
			if err != nil {
				log.Printf("Failed to connect to the local server. Is it running?")

				// assuming that the user has stopped the local server but still trying to access the remote url
				// As of now, I've decided to stop the client and close the connection in such case.
				conn.Close()
				os.Exit(1)
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
			responseBytes, err := msgpack.Marshal(message)
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
	log.Println("Closing connection...")

	// inform remote server about client conection closure
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
	log.Println("Thanks for using Portsicle! Have a nice day :)")
}
