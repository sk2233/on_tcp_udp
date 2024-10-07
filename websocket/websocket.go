/*
@author: sk
@date: 2024/9/15
*/
package websocket

import (
	"bytes"
	"crypto/sha1"
	"encoding/base64"
	"io"
	"my_tcp/utils"
	"net"
)

const (
	OpText = 0x01
)

type Frame struct {
	Flag1, Flag2 uint8
	MaskKey      []byte
	Data         []byte
}

func (f *Frame) GetData() []byte {
	if len(f.MaskKey) == 0 {
		return f.Data
	}

	res := make([]byte, len(f.Data))
	for i := 0; i < len(res); i++ { // 客服端发的数据会mask一下不发送明文 就是异或一下 服务端不需要
		res[i] = f.Data[i] ^ f.MaskKey[i%4]
	}
	return res
}

func (f *Frame) GetOp() uint8 {
	return f.Flag1 & 0b0000_1111
}

func (f *Frame) HasMask() bool {
	return f.Flag2&0b1000_0000 > 0
}

func (f *Frame) GetDataLen() uint32 {
	return uint32(f.Flag2 & 0b0111_1111) // 这里实际是变长编码，这里假定数据量较小不会触发后面的编码规则
}

type WebSocket struct {
	Conn net.Conn
}

// websocket 统一约定必须是这个
var keyUID = []byte("258EAFA5-E914-47DA-95CA-C5AB0DC85B11")

func GenAcceptKey(key string) string {
	s := sha1.New()
	s.Write([]byte(key)) // 按约定对请求的 key 进行加密返回
	s.Write(keyUID)
	return base64.StdEncoding.EncodeToString(s.Sum(nil))
}

func (s *WebSocket) Upgrade(req *HttpReq) {
	resp := NewHttpResp(101, "Switching Protocols")
	resp.Header["Upgrade"] = "websocket"
	resp.Header["Connection"] = "Upgrade"
	resp.Header["Sec-WebSocket-Accept"] = GenAcceptKey(req.Header["Sec-WebSocket-Key"])
	WriteResp(s.Conn, resp)
}

func ReadU8(reader io.Reader) uint8 {
	bs := make([]byte, 1)
	_, err := reader.Read(bs)
	utils.HandleErr(err)
	return bs[0]
}

func ReadBytes(reader io.Reader, l uint32) []byte {
	bs := make([]byte, l)
	_, err := reader.Read(bs)
	utils.HandleErr(err)
	return bs
}

func (s *WebSocket) Read() ([]byte, uint8) {
	frame := &Frame{
		Flag1: ReadU8(s.Conn),
		Flag2: ReadU8(s.Conn),
	}
	if frame.HasMask() {
		frame.MaskKey = ReadBytes(s.Conn, 4)
	}
	l := frame.GetDataLen()
	frame.Data = ReadBytes(s.Conn, l)
	return frame.GetData(), frame.GetOp()
}

func (s *WebSocket) Write(op uint8, data []byte) {
	frame := &Frame{
		Flag1: op | 0b1000_0000,
		Flag2: uint8(len(data)) | 0x0,
		Data:  data,
	}
	WriteFrame(s.Conn, frame)
}

func WriteFrame(conn net.Conn, frame *Frame) {
	buff := &bytes.Buffer{}
	buff.Write([]byte{frame.Flag1, frame.Flag2})
	buff.Write(frame.Data)
	_, err := conn.Write(buff.Bytes())
	utils.HandleErr(err)
}

func NewWebSocket(conn net.Conn) *WebSocket {
	return &WebSocket{Conn: conn}
}
