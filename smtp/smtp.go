/*
@author: sk
@date: 2024/9/21
*/
package smtp

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"mime"
	"mime/quotedprintable"
	"my_tcp/utils"
	"net"
	"strconv"
	"strings"
	"time"
)

type SMTPClient struct {
	Host     string
	Port     int
	UserMail string
	Feature  []string
	Conn     net.Conn
}

type SMTPResp struct {
	Code int
	Msg  string
}

type SMTPMulResp struct {
	Code int
	Msgs []string
}

func ReadSMTP(conn net.Conn) *SMTPResp {
	bs := make([]byte, 4096)
	l, err := conn.Read(bs)
	utils.HandleErr(err)
	line := string(bs[:l])
	index := strings.Index(line, " ")
	code, err := strconv.Atoi(line[:index])
	utils.HandleErr(err)
	return &SMTPResp{Code: code, Msg: strings.TrimSpace(line[index+1:])}
}

func ReadSMTPMul(conn net.Conn) *SMTPMulResp {
	bs := make([]byte, 4096)
	l, err := conn.Read(bs)
	utils.HandleErr(err)
	lines := strings.Split(string(bs[:l]), "\r\n")
	msgs := make([]string, 0)
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if len(line) == 0 {
			continue
		}
		msgs = append(msgs, line[4:]) // 移除前缀
	} // 250 是正常的 code 这里暂定 250
	return &SMTPMulResp{Code: 250, Msgs: msgs}
}

func PrintSMTPResp(conn net.Conn) {
	resp := ReadSMTP(conn)
	fmt.Printf("===============%d===============\n%s\n", resp.Code, resp.Msg)
}

func SendCmd(conn net.Conn, cmdFmt string, cmdArgs ...any) {
	cmd := fmt.Sprintf(cmdFmt+"\r\n", cmdArgs...)
	_, err := conn.Write([]byte(cmd))
	utils.HandleErr(err)
}

func (c *SMTPClient) Login(userMail string, authCode string) {
	c.UserMail = userMail
	// 建立链接
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", c.Host, c.Port))
	utils.HandleErr(err)
	PrintSMTPResp(conn)
	// ehlo 表明身份&获取服务端属性
	SendCmd(conn, "EHLO localhost")
	resp := ReadSMTPMul(conn)
	c.Feature = resp.Msgs
	fmt.Printf("===============%d===============\n", resp.Code)
	for _, msg := range resp.Msgs {
		fmt.Println(msg)
	}
	// 建立 tls 链接，一般需要看服务端属性选择建立，但是基本没有服务端会允许明文传输，这里直接认为需要加密
	SendCmd(conn, "STARTTLS") // 标记开始转换连接
	PrintSMTPResp(conn)
	conn = tls.Client(conn, &tls.Config{ServerName: c.Host})
	// 进行认证
	authStr := fmt.Sprintf("\x00%s\x00%s", userMail, authCode)
	auth := base64.StdEncoding.EncodeToString([]byte(authStr))
	SendCmd(conn, "AUTH PLAIN %s", auth) // PLAIN 是服务端支持的一种认证方式
	PrintSMTPResp(conn)
	c.Conn = conn
}

func (c *SMTPClient) Send(toEmail string, title string, content string) {
	// 标记发送人与接收人
	SendCmd(c.Conn, "MAIL FROM:<%s> BODY=8BITMIME", c.UserMail) // BODY=8BITMIME 是根据服务端特性来的
	PrintSMTPResp(c.Conn)
	SendCmd(c.Conn, "RCPT TO:<%s>", toEmail)
	PrintSMTPResp(c.Conn)
	// 发送 data 指令，传输数据
	SendCmd(c.Conn, "DATA")
	PrintSMTPResp(c.Conn)
	// 拼接传输内容
	buff := &bytes.Buffer{}
	// 设置基本属性
	buff.WriteString("Mime-Version: 1.0\r\n")
	buff.WriteString("Content-Transfer-Encoding: quoted-printable\r\n")
	buff.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
	buff.WriteString(fmt.Sprintf("Date: %s\r\n", time.Now().Format(time.RFC1123Z)))
	buff.WriteString(fmt.Sprintf("From: %s\r\n", c.UserMail))
	buff.WriteString(fmt.Sprintf("To: %s\r\n", toEmail))
	// 中文进行编码
	encode := mime.QEncoding
	title = encode.Encode("UTF-8", title)
	buff.WriteString(fmt.Sprintf("Subject: %s\r\n", title))
	buff.WriteString("\r\n")
	writer := quotedprintable.NewWriter(buff)
	_, err := writer.Write([]byte(content))
	writer.Close() // 立即写入
	utils.HandleErr(err)
	buff.WriteString("\r\n.\r\n") // 使用单独的一行 . 作为数据的结束
	_, err = c.Conn.Write(buff.Bytes())
	utils.HandleErr(err)
	PrintSMTPResp(c.Conn)
}

func (c *SMTPClient) Quit() {
	SendCmd(c.Conn, "QUIT")
	PrintSMTPResp(c.Conn)
	c.Conn.Close()
}

func NewSMTPClient(host string, port int) *SMTPClient {
	return &SMTPClient{Host: host, Port: port}
}
