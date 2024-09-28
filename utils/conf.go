/*
@author: sk
@date: 2024/9/28
*/
package utils

import (
	"encoding/json"
	"os"
)

const (
	BasePath = "/Users/bytedance/Documents/go/my_tcp/"
)

type Conf struct {
	FTPAddr   string `json:"ftp_addr"`
	FTPUser   string `json:"ftp_user"`
	FTPPasswd string `json:"ftp_passwd"`

	SMTPEmail string `json:"smtp_email"`
	SMTPAuth  string `json:"smtp_auth"`

	SSHAddr   string `json:"ssh_addr"`
	SSHUser   string `json:"ssh_user"`
	SSHPasswd string `json:"ssh_passwd"`
}

func GetConf() *Conf {
	// 固定死，有些在 test 环境位置会丢
	bs, err := os.ReadFile(BasePath + "conf.json")
	HandleErr(err)

	conf := &Conf{}
	err = json.Unmarshal(bs, conf)
	HandleErr(err)
	return conf
}
