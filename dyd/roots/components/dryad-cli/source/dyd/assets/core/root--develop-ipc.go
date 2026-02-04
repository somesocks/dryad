package core

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"strings"

	zlog "github.com/rs/zerolog/log"
)

type rootDevelopIPCHandlers struct {
	OnSave func() error
	OnStatus func() ([]rootDevelopStatusEntry, error)
	OnStop func() error
}

type rootDevelopIPCRequest struct {
	Cmd string `json:"cmd"`
}

type rootDevelopIPCResponse struct {
	Status  string                  `json:"status"`
	Entries []rootDevelopStatusEntry `json:"entries,omitempty"`
	Message string                  `json:"message,omitempty"`
}

type rootDevelopIPCServer struct {
	listener net.Listener
	socket   string
}

func rootDevelopIPC_start(socketPath string, handlers rootDevelopIPCHandlers) (*rootDevelopIPCServer, error) {
	if socketPath == "" {
		return nil, errors.New("rootDevelopIPC_start: empty socket path")
	}

	_ = os.Remove(socketPath)

	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		return nil, err
	}

	server := &rootDevelopIPCServer{
		listener: listener,
		socket:   socketPath,
	}

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				if errors.Is(err, net.ErrClosed) {
					return
				}
				zlog.Error().Err(err).Msg("root develop ipc accept error")
				return
			}
			go rootDevelopIPC_handle(conn, handlers)
		}
	}()

	return server, nil
}

func (s *rootDevelopIPCServer) Close() error {
	if s == nil {
		return nil
	}
	if s.listener != nil {
		_ = s.listener.Close()
	}
	if s.socket != "" {
		_ = os.Remove(s.socket)
	}
	return nil
}

func rootDevelopIPC_handle(conn net.Conn, handlers rootDevelopIPCHandlers) {
	defer conn.Close()

	reader := bufio.NewReader(conn)
	payload, err := rootDevelopIPC_readMessage(reader)
	if err != nil {
		_ = rootDevelopIPC_writeMessage(conn, rootDevelopIPCResponse{
			Status:  "error",
			Message: err.Error(),
		})
		return
	}

	var req rootDevelopIPCRequest
	if err := json.Unmarshal(payload, &req); err != nil {
		_ = rootDevelopIPC_writeMessage(conn, rootDevelopIPCResponse{
			Status:  "error",
			Message: err.Error(),
		})
		return
	}

	switch strings.TrimSpace(req.Cmd) {
	case "save":
		if handlers.OnSave != nil {
			if err := handlers.OnSave(); err != nil {
				_ = rootDevelopIPC_writeMessage(conn, rootDevelopIPCResponse{
					Status:  "error",
					Message: err.Error(),
				})
				return
			}
		}
		_ = rootDevelopIPC_writeMessage(conn, rootDevelopIPCResponse{
			Status: "ok",
		})
	case "status":
		if handlers.OnStatus != nil {
			entries, err := handlers.OnStatus()
			if err != nil {
				_ = rootDevelopIPC_writeMessage(conn, rootDevelopIPCResponse{
					Status:  "error",
					Message: err.Error(),
				})
				return
			}
			_ = rootDevelopIPC_writeMessage(conn, rootDevelopIPCResponse{
				Status:  "ok",
				Entries: entries,
			})
			return
		}
		_ = rootDevelopIPC_writeMessage(conn, rootDevelopIPCResponse{
			Status:  "error",
			Message: "status handler not available",
		})
	case "stop":
		if handlers.OnStop != nil {
			if err := handlers.OnStop(); err != nil {
				_ = rootDevelopIPC_writeMessage(conn, rootDevelopIPCResponse{
					Status:  "error",
					Message: err.Error(),
				})
				return
			}
		}
		_ = rootDevelopIPC_writeMessage(conn, rootDevelopIPCResponse{
			Status: "ok",
		})
	default:
		_ = rootDevelopIPC_writeMessage(conn, rootDevelopIPCResponse{
			Status:  "error",
			Message: "unknown command",
		})
	}
}

func rootDevelopIPC_readMessage(reader *bufio.Reader) ([]byte, error) {
	var contentLength int
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return nil, err
		}

		line = strings.TrimRight(line, "\r\n")
		if line == "" {
			break
		}

		if strings.HasPrefix(strings.ToLower(line), "content-length:") {
			_, err := fmt.Sscanf(line, "Content-Length: %d", &contentLength)
			if err != nil {
				return nil, err
			}
		}
	}

	if contentLength <= 0 {
		return nil, errors.New("missing Content-Length")
	}

	payload := make([]byte, contentLength)
	_, err := io.ReadFull(reader, payload)
	if err != nil {
		return nil, err
	}

	return payload, nil
}

func rootDevelopIPC_writeMessage(conn net.Conn, res rootDevelopIPCResponse) error {
	payload, err := json.Marshal(res)
	if err != nil {
		return err
	}
	header := fmt.Sprintf("Content-Length: %d\r\n\r\n", len(payload))
	if _, err := conn.Write([]byte(header)); err != nil {
		return err
	}
	_, err = conn.Write(payload)
	return err
}
