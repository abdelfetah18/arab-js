package api

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"os"

	"github.com/google/uuid"
	"github.com/sourcegraph/jsonrpc2"
)

type EmitArgs struct {
	SessionId string `json:"session_id"`
	FilePath  string `json:"file_path"`
}

type EchoArgs struct {
	Message string `json:"message"`
}

type readWriteCloser struct {
	io.Reader
	io.Writer
}

func (rwc readWriteCloser) Close() error {
	return nil
}

var sessions = map[string]Session{}

func StartServer() {
	log.SetOutput(os.Stderr)
	log.Println("Server is up and running (JSON-RPC 2.0 over stdio)")

	handler := jsonrpc2.HandlerWithError(func(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (interface{}, error) {
		switch req.Method {
		case "createSession":
			sessionId := uuid.NewString()
			session := NewSession(sessionId)
			sessions[sessionId] = session
			return session, nil
		case "emit":
			var params EmitArgs
			if req.Params != nil {
				if err := json.Unmarshal(*req.Params, &params); err != nil {
					return nil, err
				}
			}

			if session, ok := sessions[params.SessionId]; ok {
				return session.EmitFile(params.FilePath), nil
			}

			return nil, errors.New("sessions not found")
		case "echo":
			var params EchoArgs
			if req.Params != nil {
				if err := json.Unmarshal(*req.Params, &params); err != nil {
					return nil, err
				}
			}
			reply := "Echo: " + params.Message
			return reply, nil
		default:
			return nil, &jsonrpc2.Error{
				Code:    jsonrpc2.CodeMethodNotFound,
				Message: "method not found",
			}
		}
	})

	stream := jsonrpc2.NewBufferedStream(readWriteCloser{os.Stdin, os.Stdout}, jsonrpc2.VSCodeObjectCodec{})
	_ = jsonrpc2.NewConn(context.Background(), stream, handler)
	select {} // block forever
}
