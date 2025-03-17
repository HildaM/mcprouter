package api

import (
	"fmt"
	"net/http"

	"github.com/chatmcp/mcprouter/service/mcpserver"
	"github.com/chatmcp/mcprouter/service/sse"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// SSE is a handler for the sse endpoint
func SSE(c echo.Context) error {
	ctx := sse.GetSSEContext(c)
	if ctx == nil {
		return c.String(http.StatusInternalServerError, "Failed to get SSE context")
	}

	req := c.Request()

	key := c.Param("key")
	if key == "" {
		return c.String(http.StatusBadRequest, "Key is required")
	}

	command := mcpserver.GetCommand(key)
	fmt.Println("[debug] command: ", command)
	if command == "" {
		return c.String(http.StatusBadRequest, "Server command not found")
	}

	writer, err := sse.NewSSEWriter(c)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	// store session
	sessionID := uuid.New().String()
	session := sse.NewSSESession(writer, key, command)
	ctx.StoreSession(sessionID, session)
	defer ctx.DeleteSession(sessionID)

	go func() {
		for {
			select {
			// todo: get notification and send to session.messages channel
			case <-session.Done():
				return
			case <-req.Context().Done():
				return
			}
		}
	}()

	// response to client with endpoint url
	messagesUrl := fmt.Sprintf("/messages?sessionid=%s", sessionID)
	writer.SendEventData("endpoint", messagesUrl)

	// listen to messages
	for {
		select {
		case message := <-session.Messages():
			// fmt.Printf("sse send message: %s\n", message)
			writer.SendMessage(message)
		case <-req.Context().Done():
			fmt.Println("sse request done")
			// session.Close()
			return nil
		}
	}
}
