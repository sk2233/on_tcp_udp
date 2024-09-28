/*
@author: sk
@date: 2024/9/16
*/
package ftp

import (
	"fmt"
	"io"
	"my_tcp/utils"
	"net"
	"strconv"
	"strings"
)

func ParseLines(conn net.Conn) []string {
	bs := make([]byte, 4096)
	l, err := conn.Read(bs)
	utils.HandleErr(err)
	items := strings.Split(string(bs[:l]), "\r\n")
	res := make([]string, 0)
	for _, item := range items {
		item = strings.TrimSpace(item)
		if len(item) > 0 {
			res = append(res, item)
		}
	}
	return res
}

type FTPClient struct {
	Address string
	Conn    net.Conn
}

func (c *FTPClient) Login(username string, password string) {
	var err error
	c.Conn, err = net.Dial("tcp", c.Address)
	utils.HandleErr(err)
	c.PrintInfo() // 建立连接后，服务端请求新用户
	c.SendLine("USER %s", username)
	c.PrintInfo() // 发送完用户信息还需要密码信息
	c.SendLine("PASS %s", password)
	c.PrintInfo()
}

func (c *FTPClient) PrintInfo() {
	lines := ParseLines(c.Conn)
	fmt.Println(lines) // 没啥重要信息，打印下吧
}

func (c *FTPClient) SendLine(lineFmt string, args ...any) { // 不用再加 \r\n了
	_, err := c.Conn.Write([]byte(fmt.Sprintf(lineFmt, args...) + "\r\n"))
	utils.HandleErr(err)
}

func (c *FTPClient) OpenDataConn() net.Conn {
	c.SendLine("PASV")          // 采用主动模式，询问待连接的端口
	lines := ParseLines(c.Conn) // 解析返回的端口进行连接
	l := strings.Index(lines[0], "(")
	r := strings.LastIndex(lines[0], ")")
	items := strings.Split(lines[0][l+1:r], ",")
	hi, err := strconv.ParseInt(items[4], 10, 64)
	utils.HandleErr(err)
	lo, err := strconv.ParseInt(items[5], 10, 64)
	utils.HandleErr(err)
	addr := fmt.Sprintf("%s.%s.%s.%s:%d", items[0], items[1], items[2], items[3], hi*256+lo)
	conn, err := net.Dial("tcp", addr) // 建立数据连接
	utils.HandleErr(err)
	return conn
}

func (c *FTPClient) List(path string) []string {
	conn := c.OpenDataConn()
	defer conn.Close()
	c.SendLine("LIST %s", path) // 发送指令
	c.PrintInfo()               // 等待数据端口数据准备完毕
	lines := ParseLines(conn)
	c.PrintInfo() // 传输完毕还会再添加传输完毕指令
	return lines
}

func (c *FTPClient) Quit() {
	c.SendLine("QUIT")
	c.PrintInfo()
}

func (c *FTPClient) Del(path string) {
	c.SendLine("DELE %s", path) // 不需要使用数据端口的，不需要开数据端口
	c.PrintInfo()
}

func (c *FTPClient) Get(path string) io.Reader {
	conn := c.OpenDataConn()
	c.SendLine("RETR %s", path)
	c.PrintInfo() // 等待数据端口准备完毕
	out := conn   // 直接返回底层流，外面负责关闭
	c.PrintInfo() //
	return out
}

func (c *FTPClient) Sto(path string, in io.Reader) {
	conn := c.OpenDataConn()
	defer conn.Close()
	c.SendLine("STOR %s", path)
	c.PrintInfo()
	_, err := io.Copy(conn, in)
	utils.HandleErr(err)
	//c.PrintInfo()
}

func NewFTPClient(address string) *FTPClient {
	return &FTPClient{Address: address}
}
