/*
@author: sk
@date: 2024/9/15
*/
package websocket

import (
	"bytes"
	"fmt"
	"my_tcp/utils"
	"net"
	"strings"
)

type HttpReq struct {
	Method string
	Path   string
	Header map[string]string
}

type HttpResp struct {
	Code   int
	Msg    string
	Header map[string]string
	Data   []byte
}

func NewHttpResp(code int, msg string) *HttpResp {
	return &HttpResp{Code: code, Msg: msg, Header: make(map[string]string)}
}

type HttpHandler func(req *HttpReq, conn net.Conn)

type HttpServer struct {
	Address  string
	Handlers map[string]HttpHandler
}

func (h *HttpServer) RegisterHandler(path string, handler HttpHandler) {
	h.Handlers[path] = handler
}

func (h *HttpServer) Listen() {
	listen, err := net.Listen("tcp", h.Address)
	utils.HandleErr(err)
	for {
		conn, err := listen.Accept()
		utils.HandleErr(err)
		go h.HandleConn(conn)
	}
}

func (h *HttpServer) HandleConn(conn net.Conn) {
	bs := make([]byte, 4096)
	l, err := conn.Read(bs)
	utils.HandleErr(err)

	lines := strings.Split(string(bs[:l]), "\r\n")
	items := strings.Split(lines[0], " ")
	method := strings.TrimSpace(items[0])
	path := strings.TrimSpace(items[1])
	header := make(map[string]string)
	for i := 1; i < len(lines); i++ {
		items = strings.Split(lines[i], ":")
		if len(items) == 2 {
			header[strings.TrimSpace(items[0])] = strings.TrimSpace(items[1])
		}
	}
	req := &HttpReq{
		Method: method,
		Path:   path,
		Header: header,
	}

	if handler, ok := h.Handlers[req.Path]; ok {
		handler(req, conn)
	} else {
		conn.Close()
	}
}

func NewHttpServer(address string) *HttpServer {
	return &HttpServer{Address: address, Handlers: make(map[string]HttpHandler)}
}

func WriteResp(conn net.Conn, resp *HttpResp) {
	buff := &bytes.Buffer{}
	buff.WriteString(fmt.Sprintf("HTTP/1.1 %d %s\r\n", resp.Code, resp.Msg))
	for k, v := range resp.Header {
		buff.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
	}
	buff.WriteString("\r\n")
	buff.Write(resp.Data)

	_, err := conn.Write(buff.Bytes())
	utils.HandleErr(err)
}
