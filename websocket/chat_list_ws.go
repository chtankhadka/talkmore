package websocket

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"my-work/config"
	"my-work/controllers"
	"my-work/models"
	"my-work/utils"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var chatClients = make(map[string]map[*websocket.Conn]bool) // user_id → WebSocket connections
var chatMutex sync.Mutex

// WebSocket upgrader
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all connections (Modify if needed)
	},
}

// FetchInitialChats retrieves and sends a paginated list of chat data for a user

func HandleChatListWebSocket(app *config.AppConfig) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		userID := "67db5c78675256d16903e454"

		ws, err := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
		if err != nil {
			log.Printf("WebSocket upgrade error for user %s: %v", userID, err)
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "WebSocket upgrade failed"})
			return
		}

		chatMutex.Lock()
		if chatClients[userID] == nil {
			chatClients[userID] = make(map[*websocket.Conn]bool)
		}
		chatClients[userID][ws] = true
		chatMutex.Unlock()

		log.Printf("New WebSocket connection for user %s", userID)

		defer func() {
			chatMutex.Lock()
			delete(chatClients[userID], ws)
			if len(chatClients[userID]) == 0 {
				delete(chatClients, userID)
			}
			chatMutex.Unlock()
			ws.Close()
			log.Printf("WebSocket connection closed for user %s", userID)
		}()

		// Start watching for changes
		go utils.WatchChatsCollection(app, userID, ws)

		// Keep connection alive
		for {
			if err := ws.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				return
			}
			time.Sleep(30 * time.Second)
		}
	}
}
func HandleMessageListWebSocket(app *config.AppConfig) gin.HandlerFunc {
	return func(ctx *gin.Context) {

		mctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		clientToken, tokenError := controllers.GetMyToken(ctx)
		if tokenError != "" {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": tokenError})
			ctx.Abort()
			return
		}
		userDetails, idError := controllers.GetMyId(mctx, app, clientToken)
		if idError != "" {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": idError})
			ctx.Abort()
			return
		}
		userID := userDetails.UserID

		ws, err := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
		if err != nil {
			log.Printf("WebSocket upgrade error for user %s: %v", userID, err)
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "WebSocket upgrade failed"})
			return
		}

		chatMutex.Lock()
		if chatClients[userID] == nil {
			chatClients[userID] = make(map[*websocket.Conn]bool)
		}
		chatClients[userID][ws] = true
		chatMutex.Unlock()

		log.Printf("New WebSocket connection for user %s", userID)

		done := make(chan struct{})

		defer func() {
			chatMutex.Lock()
			delete(chatClients[userID], ws)
			if len(chatClients[userID]) == 0 {
				delete(chatClients, userID)
			}
			chatMutex.Unlock()
			close(done) // Signal goroutine to stop
			ws.Close()
			log.Printf("WebSocket connection closed for user %s", userID)
		}()

		go utils.WatchMessagesCollection(app, userID, ws, done)

		for {
			_, message, err := ws.ReadMessage()
			if err != nil {
				if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
					log.Printf("WebSocket closed normally by client for user %s", userID)
				} else if websocket.IsUnexpectedCloseError(err) {
					log.Printf("WebSocket closed unexpectedly for user %s: %v", userID, err)
				} else {
					log.Printf("WebSocket read error for user %s: %v", userID, err)
				}
				return
			}
			log.Printf("Received message from user %s: %s", userID, string(message))
			var messageDetails models.Message
			if err := json.Unmarshal(message, &messageDetails); err != nil {
				log.Printf("Error decoding JSON message from user %s: %v", userDetails.UserID, err)
				continue // Skip invalid messages
			}
			go utils.HandleClientMessage(app, *userDetails, messageDetails)
		}
	}
}
