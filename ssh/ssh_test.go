/*
@author: sk
@date: 2024/9/16
*/
package ssh

import (
	"fmt"
	"my_tcp/utils"
	"testing"
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
