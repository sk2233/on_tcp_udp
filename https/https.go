/*
@author: sk
@date: 2024/9/16
*/
package https

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"my_tcp/utils"
	"net"
	"strings"
)

type HttpsReq struct {
	Method string
	Path   string
	Header map[string]string
}

type HttpsHandler func(req *HttpsReq, conn net.Conn)

type HttpsServer struct {
	Addr     string
	CertFile string
	KeyFile  string
	Handlers map[string]HttpsHandler
}

func NewHttpsServer(addr string, certFile string, keyFile string) *HttpsServer {
	return &HttpsServer{Addr: addr, CertFile: certFile, KeyFile: keyFile, Handlers: make(map[string]HttpsHandler)}
}

func (h *HttpsServer) RegisterHandler(path string, handler HttpsHandler) {
	h.Handlers[path] = handler
}

func (h *HttpsServer) Listen() {
	pair, err := tls.LoadX509KeyPair(h.CertFile, h.KeyFile) // 加载证书并使用 tls 的服务端
	utils.HandleErr(err)
	conf := &tls.Config{Certificates: []tls.Certificate{pair}}
	listen, err := tls.Listen("tcp", h.Addr, conf)
	utils.HandleErr(err)
	for {
		conn, err := listen.Accept()
		utils.HandleErr(err)
		go h.HandleConn(conn)
	}
}

func (h *HttpsServer) HandleConn(conn net.Conn) {
	bs := make([]byte, 4096)
	l, err := conn.Read(bs)
	if err != nil { // 可能存在证书问题不能直接报错
		fmt.Println(err)
		conn.Close()
		return
	}

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
	req := &HttpsReq{
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

type HttpsResp struct {
	Code   int
	Msg    string
	Header map[string]string
	Data   []byte
}

func NewHttpsResp(code int, msg string) *HttpsResp {
	return &HttpsResp{Code: code, Msg: msg, Header: make(map[string]string)}
}

func WriteResp(conn net.Conn, resp *HttpsResp) {
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

//type HttpsHandler func()
//
//type HttpsServer struct {
//	Address  string
//	CertFile string
//	KeyFile  string
//	Handlers map[string]HttpsHandler
//}
//
//func (h *HttpsServer) RegisterHandler(path string, handler HttpsHandler) {
//	h.Handlers[path] = handler
//}
//
//func (h *HttpsServer) Listen() {
//	listen, err := net.Listen("tcp", h.Address)
//	utils.HandleErr(err)
//	for {
//		conn, err := listen.Accept()
//		utils.HandleErr(err)
//		go h.HandleConn(conn)
//	}
//}
//
//type TLSFrame struct {
//	Type    uint8
//	Version uint16
//	Len     uint16
//	Data    []byte
//}
//
//func ReadBytes(reader io.Reader, len0 uint32) []byte {
//	bs := make([]byte, len0)
//	_, err := reader.Read(bs)
//	utils.HandleErr(err)
//	return bs
//}
//
//func ReadU16(reader io.Reader) uint16 {
//	bs := make([]byte, 2)
//	_, err := reader.Read(bs)
//	utils.HandleErr(err)
//	return binary.LittleEndian.Uint16(bs)
//}
//
//func ReadU8(reader io.Reader) uint8 {
//	bs := make([]byte, 1)
//	_, err := reader.Read(bs)
//	utils.HandleErr(err)
//	return bs[0]
//}
//
//func ParseFrame(conn net.Conn) []*TLSFrame {
//	bs := make([]byte, 4096) // 有多帧数据先都读取出来
//	l, err := conn.Read(bs)
//	utils.HandleErr(err)
//
//	reader := bytes.NewReader(bs[:l])
//	res := make([]*TLSFrame, 0)
//	for reader.Len() > 0 {
//		frame := &TLSFrame{}
//		frame.Type = ReadU8(reader)
//		frame.Version = ReadU16(reader)
//		frame.Len = ReadU16(reader)
//		frame.Data = ReadBytes(reader, uint32(frame.Len))
//		res = append(res, frame)
//	}
//	return res
//}
//
//type HttpsData struct { // 一些中间过程数据
//	Version   uint16
//	Random    []byte
//	SessionID []byte
//}
//
//func (h *HttpsServer) HandleClientHello(conn net.Conn) *HttpsData {
//	frames := ParseFrame(conn)
//	reader := bytes.NewReader(frames[0].Data) // 只看第一帧的
//	ReadBytes(reader, 4)
//	data := &HttpsData{}
//	data.Version = ReadU16(reader)
//	data.Random = ReadBytes(reader, 32)
//	l := ReadU8(reader)
//	data.SessionID = ReadBytes(reader, uint32(l))
//	return data
//}
//
//func (h *HttpsServer) HandleConn(conn net.Conn) {
//	data := h.HandleClientHello(conn)
//	h.SendServerHello(conn, data)
//}
//
//func (h *HttpsServer) SendServerHello(conn net.Conn, data *HttpsData) {
//
//}
//
//func NewHttpsServer(addr string, certFile string, keyFile string) *HttpsServer {
//	return &HttpsServer{Address: addr, CertFile: certFile, KeyFile: keyFile, Handlers: make(map[string]HttpsHandler)}
//}
