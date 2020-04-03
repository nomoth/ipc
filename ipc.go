package ipc

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"io"
	"net"
	"os"
	"os/exec"
	"strings"
)

type msgType uint32

const (
	RUN_COMMAND msgType = iota
	GET_WORKSPACES
	SUBSCRIBE
	GET_OUTPUTS
	GET_TREE
	GET_MARKS
	GET_BAR_CONFIG
	GET_VERSION
	GET_BINDING_MODES
	GET_CONFIG
	SEND_TICK
	SYNC
)

const (
	swayEnvVar="SWAYSOCK"
)

var (
	ErrNoSocketPath = errors.New("impossible to retrieve the sway socket path")
)

type Connection struct {
	net.Conn
}

func NewConnection() (*Connection, error) {
	socket, err := getSocketPath()
	if err != nil {
		return nil, err
	}
	conn, err := getConnection(socket)
	if err != nil {
		return nil, err
	}
	return &Connection{
		Conn: conn,
	}, nil
}

func (c *Connection) Run(cmd string) error {
	j, err := c.send(RUN_COMMAND, cmd)
	if err != nil {
		return err
	}
	var status []struct {
		Success bool
		Error   string
	}
	err = json.Unmarshal(j, &status)
	if err != nil {
		return err
	}
	for _, s := range status {
		if !s.Success {
			return errors.New(s.Error)
		}
	}
	return nil
}

func (c *Connection) send(t msgType, cmd string) ([]byte, error) {
	h := &struct {
		Magic  [6]byte
		Length uint32
		Type   msgType
	}{
		Magic:  [6]byte{'i', '3', '-', 'i', 'p', 'c'},
		Length: uint32(len(cmd)),
		Type:   t,
	}
	err := binary.Write(c, NativeByteOrder, h)
	if err != nil {
		return nil, err
	}
	_, err = c.Write([]byte(cmd))
	if err != nil {
		return nil, err
	}
	err = binary.Read(c, NativeByteOrder, h)
	if err != nil {
		return nil, err
	}
	b := bytes.Buffer{}
	_, err = io.CopyN(&b, c, int64(h.Length))
	if err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func getConnection(address string) (net.Conn, error) {
	c, err := net.Dial("unix", address)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func getSocketPath() (string, error) {
	path := os.Getenv(swayEnvVar)
	if path != "" {
		return path, nil
	}
	cmd := exec.Command("sway", "--get-socketpath")
	var buf bytes.Buffer
	cmd.Stdout = &buf
	err := cmd.Run()
	if err != nil {
		return "", ErrNoSocketPath
	}
	return strings.TrimSpace(buf.String()), nil
}
