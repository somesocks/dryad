package core

import (
	"bufio"
	"errors"
	"net"
	"os"
	"strings"

	zlog "github.com/rs/zerolog/log"
)

type rootDevelopIPCHandlers struct {
	OnSave func() error
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
	line, err := reader.ReadString('\n')
	if err != nil {
		return
	}

	cmd := strings.TrimSpace(line)
	switch cmd {
	case "save":
		if handlers.OnSave != nil {
			if err := handlers.OnSave(); err != nil {
				_, _ = conn.Write([]byte("error\n"))
				return
			}
		}
		_, _ = conn.Write([]byte("ok\n"))
	default:
		_, _ = conn.Write([]byte("unknown\n"))
	}
}
