package sshrpc

import (
	"fmt"
	"net/rpc"

	"golang.org/x/crypto/ssh"
)

type sshrpcSession struct {
	*ssh.Session
}

func (s sshrpcSession) Read(p []byte) (n int, err error) {
	pipe, err := s.StdoutPipe()
	if err != nil {
		return 0, err
	}
	return pipe.Read(p)
}

func (s sshrpcSession) Write(p []byte) (n int, err error) {
	pipe, err := s.StdinPipe()
	if err != nil {
		return 0, err
	}
	return pipe.Write(p)
}

// Client represents an RPC client using an SSH backed connection.
type Client struct {
	*rpc.Client
	Config    *ssh.ClientConfig
	Subsystem string
}

// NewClient returns a new Client to handle RPC requests.
func NewClient() *Client {

	config := &ssh.ClientConfig{
		User: "sshrpc",
		Auth: []ssh.AuthMethod{
			ssh.Password("sshrpc"),
		},
	}

	return &Client{nil, config, "sshrpc"}

}

// Connect starts a client connection to the given SSH/RPC server.
func (c *Client) Connect(address string) {

	sshClient, err := ssh.Dial("tcp", address, c.Config)
	if err != nil {
		panic("Failed to dial: " + err.Error())
	}

	// Each ClientConn can support multiple interactive sessions,
	// represented by a Session.
	sshSession, err := sshClient.NewSession()
	if err != nil {
		panic("Failed to create session: " + err.Error())
	}
	//defer sshSession.Close()

	err = sshSession.RequestSubsystem(c.Subsystem)
	if err != nil {
		fmt.Println("Unable to start subsystem:", err.Error())
	}

	session := sshrpcSession{sshSession}
	c.Client = rpc.NewClient(session)

	return
}
