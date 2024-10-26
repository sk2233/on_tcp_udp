/*
@author: sk
@date: 2024/9/16
*/
package ssh

import (
	"fmt"
	"my_tcp/utils"
	"testing"

	"github.com/luke-park/ecdh25519"
)

// ssh -o StrictHostKeyChecking=no  sk@192.168.31.182  采用非严格校验

func TestSSH(t *testing.T) {
	conf := utils.GetConf()
	// 也要使用系统 ssh 的实现
	client := NewSSHClient(conf.SSHAddr)
	client.Login(conf.SSHUser, conf.SSHPasswd)

	data := client.Run("dir") // 这里链接的是 win
	fmt.Println(string(data))

	client.Close()
}

func TestEcdh25519(t *testing.T) {
	prv1, _ := ecdh25519.GenerateKey()
	prv2, _ := ecdh25519.GenerateKey()

	s1 := prv1.ComputeSecret(prv2.Public())
	s2 := prv2.ComputeSecret(prv1.Public())

	fmt.Println(s1, len(s1))
	fmt.Println(s2, len(s2))
}
