/*
@author: sk
@date: 2024/9/16
*/
package smtp

import (
	"my_tcp/utils"
	"testing"
)

func TestSMTP(t *testing.T) {
	conf := utils.GetConf()
	client := NewSMTPClient("smtp.qq.com", 587)
	client.Login(conf.SMTPEmail, conf.SMTPAuth) // 必须先登录
	client.Send(conf.SMTPEmail, "测试邮件1", "测试内容1")
	client.Send(conf.SMTPEmail, "测试邮件2", "测试内容2")
	client.Quit()
}
