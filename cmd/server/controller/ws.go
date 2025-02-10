package controller

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"

	"github.com/agalitsyn/activity/cmd/server/renderer"
	"github.com/agalitsyn/activity/internal/model"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type WebSocketController struct {
	*renderer.HTMLRenderer

	agentStorage model.AgentRepository
	browserConns sync.Map
}

func NewWebsocketController(
	r *renderer.HTMLRenderer,
	agentStorage model.AgentRepository,
) *WebSocketController {
	return &WebSocketController{
		HTMLRenderer: r,
		agentStorage: agentStorage,
	}
}

func (s *WebSocketController) HandleAgent(w http.ResponseWriter, r *http.Request) {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		http.Error(w, "invalid remote address", http.StatusInternalServerError)
		return
	}

	agent, err := s.agentStorage.FetchAgent(host)
	if errors.Is(err, model.ErrAgentNotFound) {
		agent = model.Agent{ID: host}
		if err = s.agentStorage.CreateAgent(agent); err != nil {
			http.Error(w, "connection error", http.StatusInternalServerError)
			return
		}
	} else if err != nil {
		http.Error(w, "connection error", http.StatusForbidden)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("ERROR failed to upgrade to ws:", err)
		return
	}
	defer func() {
		if err := s.agentStorage.DeleteAgent(host); err != nil {
			log.Println("ERROR disconnect:", err)
		}

		// Send updates to all browser connections after new agent disconnected
		s.broadcastAgentList()

		conn.Close()
	}()

	if err := s.writeWelcomeMsg(conn); err != nil {
		log.Println("ERROR write welcome message:", err)
		return
	}
	// Send updates to all browser connections after new agent connected
	s.broadcastAgentList()

	// keep connection
	for {
		_, rawMessage, err := conn.ReadMessage()
		if err != nil {
			log.Println("ERROR read ws:", err)
			break
		}
		log.Printf("DEBUG received: %s:", rawMessage)

		var message model.Message
		if err := json.Unmarshal(rawMessage, &message); err != nil {
			log.Println("ERROR unmarshal message:", err)
			continue
		}

		activeApp := ""
		activeAppContext := ""
		for _, app := range message.Apps {
			if app.IsActive {
				activeApp = app.Name
				if app.Context != nil {
					for k, v := range app.Context {
						activeAppContext += fmt.Sprintf("%s: %s\n", k, v)
					}
				}
				break
			}
		}
		agent.ActiveApp = activeApp
		agent.ActiveAppContext = activeAppContext

		if err := s.agentStorage.UpdateAgent(agent); err != nil {
			log.Println("ERROR update agent:", err)
		}
	}
}

func (s *WebSocketController) broadcastAgentList() {
	agents, err := s.agentStorage.FetchAgents()
	if err != nil {
		log.Println("ERROR fetch agents:", err)
		return
	}

	s.browserConns.Range(func(key, value interface{}) bool {
		conn, ok := value.(*websocket.Conn)
		if !ok {
			panic(fmt.Errorf("could not get browser ws connection by key: %s", key))
		}

		if err := s.writeAgentsListMsg(conn, agents); err != nil {
			log.Println("ERROR write agents list message:", err)
		}

		return true
	})
}

func (s *WebSocketController) StartPeriodicBroadcast(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)

	go func() {
		defer ticker.Stop()

		for {
			select {
			case t := <-ticker.C:
				log.Println("DEBUG run broadcast", t.Format(time.RFC3339))
				s.broadcastAgentList()
			case <-ctx.Done():
				return
			}
		}
	}()
}

func (s *WebSocketController) HandleBrowser(w http.ResponseWriter, r *http.Request) {
	agents, err := s.agentStorage.FetchAgents()
	if err != nil {
		log.Println("ERROR fetch agents:", err)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("ERROR failed to upgrade to ws:", err)
		return
	}
	defer conn.Close()

	connID := uuid.New().String()
	s.browserConns.Store(connID, conn)
	defer s.browserConns.Delete(connID)

	if err := s.writeWelcomeMsg(conn); err != nil {
		log.Println("ERROR write welcome message:", err)
		return
	}

	if err := s.writeAgentsListMsg(conn, agents); err != nil {
		log.Println("ERROR write agents list message:", err)
		return
	}

	// keep connection
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("ERROR %v", err)
			}
			break
		}
	}
}

func (s *WebSocketController) writeWelcomeMsg(conn *websocket.Conn) error {
	return conn.WriteMessage(websocket.TextMessage, []byte("activity server connected"))
}

func (s *WebSocketController) writeAgentsListMsg(conn *websocket.Conn, agents []model.Agent) error {
	var rows strings.Builder
	for i, agent := range agents {
		rows.WriteString(fmt.Sprintf(`
        <tr>
            <td>%d</td>
            <td>%s</td>
			<td>%s</td>
			<td>%s</td>
        </tr>`, i+1, agent.ID, agent.ActiveApp, agent.ActiveAppContext))
	}
	if len(agents) == 0 {
		rows.WriteString(`<tr><td colspan="99" class="text-center">No clients</td></tr>`)
	}

	msg := fmt.Sprintf(`<tbody id="client-list" hx-swap-oob="morphdom">%s</tbody>`, rows.String())
	return conn.WriteMessage(websocket.TextMessage, []byte(msg))
}
