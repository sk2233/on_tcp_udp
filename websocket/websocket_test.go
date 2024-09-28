/*
@author: sk
@date: 2024/9/15
*/
package websocket

import (
	"my_tcp/utils"
	"net"
	"os"
	"strconv"
	"testing"
)

func TestWebSocket(t *testing.T) {
	server := NewHttpServer("127.0.0.1:8080")
	server.RegisterHandler("/", GoHome)
	server.RegisterHandler("/echo", Echo)
	server.Listen()
}

func Echo(req *HttpReq, conn net.Conn) {
	socket := NewWebSocket(conn)
	socket.Upgrade(req) // 必须先调用，进行协议升级
	for {
		data, op := socket.Read()
		socket.Write(op, data)
	}
}

func GoHome(_ *HttpReq, conn net.Conn) {
	resp := NewHttpResp(200, "OK")
	bs, err := os.ReadFile(utils.BasePath + "websocket/home.html")
	utils.HandleErr(err)
	resp.Header["Content-Length"] = strconv.Itoa(len(bs))
	resp.Header["Content-Type"] = "text/html; charset=utf-8"
	resp.Data = bs
	WriteResp(conn, resp)
	conn.Close()
}
