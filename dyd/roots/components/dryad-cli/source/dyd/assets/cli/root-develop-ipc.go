package cli

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
)

type rootDevelopIPCRequest struct {
	Cmd string `json:"cmd"`
}

type rootDevelopIPCResponse struct {
	Status    string   `json:"status"`
	Changed   []string `json:"changed,omitempty"`
	Conflicts []string `json:"conflicts,omitempty"`
	Message   string   `json:"message,omitempty"`
}

func rootDevelopIPC_send(socketPath string, cmd string) (rootDevelopIPCResponse, error) {
	var res rootDevelopIPCResponse

	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		return res, err
	}
	defer conn.Close()

	req := rootDevelopIPCRequest{Cmd: cmd}
	if err := rootDevelopIPC_writeMessage(conn, req); err != nil {
		return res, err
	}

	reader := bufio.NewReader(conn)
	payload, err := rootDevelopIPC_readMessage(reader)
	if err != nil {
		return res, err
	}

	if err := json.Unmarshal(payload, &res); err != nil {
		return res, err
	}

	return res, nil
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

func rootDevelopIPC_writeMessage(conn net.Conn, req rootDevelopIPCRequest) error {
	payload, err := json.Marshal(req)
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
