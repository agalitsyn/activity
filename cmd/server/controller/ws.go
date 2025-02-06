package controller

import (
	"errors"
	"log"
	"net"
	"net/http"

	"github.com/gorilla/websocket"

	"github.com/agalitsyn/activity/internal/model"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type WebSocketContoller struct {
	clientStorage model.ClientRepository
}

func NewWebsocketController(clientStorage model.ClientRepository) *WebSocketContoller {
	return &WebSocketContoller{
		clientStorage: clientStorage,
	}
}

func (s *WebSocketContoller) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		http.Error(w, "invalid remote address", http.StatusInternalServerError)
		return
	}

	_, err = s.clientStorage.FetchClient(host)
	if errors.Is(err, model.ErrClientNotFound) {
		client := model.Client{ID: host}
		if err = s.clientStorage.CreateClient(client); err != nil {
			log.Println("ERROR client connection:", err)
			http.Error(w, "client connection error", http.StatusInternalServerError)
			return
		}
	} else if err != nil {
		http.Error(w, "client connection error", http.StatusForbidden)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("ERROR: failed to upgrade to WebSocket:", err)
		return
	}
	defer func() {
		conn.Close()

		if err := s.clientStorage.DeleteClient(host); err != nil {
			log.Println("ERROR client disconnection:", err)
		}
	}()

	welcomeMessage := []byte("activity server connected")
	if err := conn.WriteMessage(websocket.TextMessage, welcomeMessage); err != nil {
		log.Println("ERROR write ws:", err)
		return
	}

	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			log.Println("ERROR read ws:", err)
			break
		}
		log.Printf("DEBUG received: %s", message)

		if err := conn.WriteMessage(messageType, message); err != nil {
			log.Println("ERROR write ws:", err)
			break
		}
	}
}
