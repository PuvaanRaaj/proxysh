package ipc

import (
	"encoding/json"
	"fmt"
	"net"
	"time"
)

type Client struct {
	socketPath string
}

func NewClient(socketPath string) *Client {
	return &Client{socketPath: socketPath}
}

func (c *Client) Send(req *Request) (*Response, error) {
	conn, err := net.DialTimeout("unix", c.socketPath, 3*time.Second)
	if err != nil {
		return nil, fmt.Errorf("daemon not running (could not connect to %s): %w", c.socketPath, err)
	}
	defer conn.Close()

	if err := json.NewEncoder(conn).Encode(req); err != nil {
		return nil, fmt.Errorf("send: %w", err)
	}

	var resp Response
	if err := json.NewDecoder(conn).Decode(&resp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	return &resp, nil
}

func (c *Client) Reload() error {
	resp, err := c.Send(&Request{Cmd: CmdReload})
	if err != nil {
		return err
	}
	return resp.Err()
}

func (c *Client) Status() (*Response, error) {
	return c.Send(&Request{Cmd: CmdStatus})
}

func (c *Client) Shutdown() error {
	resp, err := c.Send(&Request{Cmd: CmdShutdown})
	if err != nil {
		return err
	}
	return resp.Err()
}
