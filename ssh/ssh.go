/*
@author: sk
@date: 2024/9/28
*/
package ssh

import (
	"my_tcp/utils"
	"net"
	"time"

	"golang.org/x/crypto/ssh"
)

type SSHClient struct {
	Client *ssh.Client
	Addr   string
}

func (c *SSHClient) Login(user string, passwd string) {
	client, err := ssh.Dial("tcp", c.Addr, &ssh.ClientConfig{
		User:    user,
		Auth:    []ssh.AuthMethod{ssh.Password(passwd)},
		Timeout: 5 * time.Second,
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil // 允许任何主机的连接
		},
	})
	utils.HandleErr(err)
	c.Client = client
}

func (c *SSHClient) Run(cmd string) []byte {
	session, err := c.Client.NewSession()
	utils.HandleErr(err)
	defer session.Close()

	data, err := session.CombinedOutput(cmd)
	utils.HandleErr(err)
	return data
}

func (c *SSHClient) Close() {
	c.Client.Close()
}

func NewSSHClient(addr string) *SSHClient {
	return &SSHClient{Addr: addr}
}
