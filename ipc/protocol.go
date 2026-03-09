package ipc

import "fmt"

const (
	CmdReload   = "reload"
	CmdStatus   = "status"
	CmdShutdown = "shutdown"
)

type Request struct {
	Cmd  string `json:"cmd"`
	Data string `json:"data,omitempty"`
}

type DomainStatus struct {
	Domain string `json:"domain"`
	Target string `json:"target"`
	Active bool   `json:"active"`
}

type Response struct {
	OK      bool           `json:"ok"`
	Error   string         `json:"error,omitempty"`
	Domains []DomainStatus `json:"domains,omitempty"`
}

func (r *Response) Err() error {
	if !r.OK {
		return fmt.Errorf("daemon error: %s", r.Error)
	}
	return nil
}
