package api

import (
	"fmt"
	"net/http"

	"github.com/chatmcp/mcprouter/service/jsonrpc"
	"github.com/chatmcp/mcprouter/service/mcpclient"
	"github.com/chatmcp/mcprouter/service/sse"
	"github.com/labstack/echo/v4"
)

// Messages is a handler for the messages endpoint
func Messages(c echo.Context) error {
	ctx := sse.GetSSEContext(c)
	if ctx == nil {
		return c.String(http.StatusInternalServerError, "Failed to get SSE context")
	}

	sessionID := ctx.QueryParam("sessionid")
	if sessionID == "" {
		return ctx.JSONRPCError(jsonrpc.ErrorInvalidParams, nil)
	}

	session := ctx.GetSession(sessionID)
	if session == nil {
		return ctx.JSONRPCError(jsonrpc.ErrorInvalidParams, nil)
	}

	request, err := ctx.GetJSONRPCRequest()
	if err != nil {
		return ctx.JSONRPCError(jsonrpc.ErrorParseError, nil)
	}
	fmt.Println("[debug] Messages request: ", request)

	client := session.Client()

	if request.Method == "initialize" && client == nil {
		command := session.Command()
		_client, err := mcpclient.NewStdioClient(command)
		if err != nil {
			fmt.Printf("connect to mcp server failed: %v\n", err)
			return ctx.JSONRPCError(jsonrpc.ErrorProxyError, request.ID)
		}
		session.SetClient(_client)
		ctx.StoreSession(sessionID, session)

		client = _client
	}

	if client == nil {
		return nil
	}

	response := sse.ForwardRequest(client, request)

	if response != nil {
		session.SendMessage(response.String())
	}

	// notification message
	return ctx.JSONRPCResponse(response)
}
