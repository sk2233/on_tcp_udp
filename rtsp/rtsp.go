/*
@author: sk
@date: 2024/9/16
*/
package rtsp

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math/rand"
	"my_tcp/utils"
	"net"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	ServerName = "Sk/1.0"
	// 视频通道 trackID=0
	VideoTrack    = "trackID=0"
	RTPVideoPort  = 33333
	RTCPVideoPort = RTPVideoPort + 1 // 一般是设定差一
	// 音频通过 trackID=1
	AudioTrack    = "trackID=1"
	RTPAudioPort  = 44444
	RTCPAudioPort = RTPAudioPort + 1 // 一般是设定差一

	Fps = 25

	RTPMaxSize = 1400 // 可以使用 1500(MTU) - IP包头 - TCP 包头 - RTP包头 这里简单计算认为有 1400 可用的
)

type RTSPServer struct {
	Address string
}

func NewRTSPServer(address string) *RTSPServer {
	return &RTSPServer{Address: address}
}

func (s *RTSPServer) Listen() {
	listen, err := net.Listen("tcp", s.Address)
	utils.HandleErr(err)
	for {
		conn, err := listen.Accept()
		utils.HandleErr(err)
		go s.HandleConn(conn)
	}
}

type RTSPReq struct {
	Method  string
	Path    string
	Version string
	Header  map[string]string
}

func ParseReq(conn net.Conn) *RTSPReq {
	bs := make([]byte, 4096)
	l, err := conn.Read(bs)
	utils.HandleErr(err)
	fmt.Printf("======================Req======================\n%s", string(bs[:l])) // 测试输入

	lines := strings.Split(string(bs[:l]), "\r\n")
	header := make(map[string]string)
	for i := 1; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		items := strings.Split(line, ":")
		if len(items) != 2 {
			continue
		}
		header[strings.TrimSpace(items[0])] = strings.TrimSpace(items[1])
	}
	items := strings.Split(lines[0], " ")
	return &RTSPReq{
		Method:  strings.TrimSpace(items[0]),
		Path:    strings.TrimSpace(items[1]),
		Version: strings.TrimSpace(items[2]),
		Header:  header,
	}
}

type RTSPResp struct {
	Version string
	Code    int
	Msg     string
	Header  map[string]string
	Data    []byte
}

const (
	RandBase = "0123456789qwertyuiopasdfghjklzxcvbnm"
)

func genSession() string {
	buff := &bytes.Buffer{}
	l := len(RandBase)
	for i := 0; i < 16; i++ {
		buff.WriteByte(RandBase[rand.Intn(l)])
	}
	return buff.String()
}

type RTSPHelper struct {
	Session  string
	Run      bool
	LocalIP  string
	RemoteIP string
	// 视频远端信息
	VideoRTPPort   int
	VideoRTCPPort  int
	VideoData      []byte
	VideoDataIdx   int
	VideoSeq       uint16
	VideoTimestamp uint32
	VideoConn      *net.UDPConn
	// 音频远端信息
	AudioRTPPort   int
	AudioRTCPPort  int
	AudioData      []byte
	AudioDataIdx   int
	AudioSeq       uint16
	AudioTimestamp uint32
	AudioConn      *net.UDPConn
}

func (h *RTSPHelper) GetVideoConn() *net.UDPConn {
	if h.VideoConn == nil { // 必须懒加载 端口在  setup 才确定
		localAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", h.LocalIP, RTPVideoPort))
		utils.HandleErr(err)
		remoteAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", h.RemoteIP, h.VideoRTPPort))
		utils.HandleErr(err)
		h.VideoConn, err = net.DialUDP("udp", localAddr, remoteAddr)
		utils.HandleErr(err)
	}
	return h.VideoConn
}

func (h *RTSPHelper) GetAudioConn() *net.UDPConn {
	if h.AudioConn == nil { // 必须懒加载 端口在  setup 才确定
		localAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", h.LocalIP, RTPAudioPort))
		utils.HandleErr(err)
		remoteAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", h.RemoteIP, h.AudioRTPPort))
		utils.HandleErr(err)
		h.AudioConn, err = net.DialUDP("udp", localAddr, remoteAddr)
		utils.HandleErr(err)
	}
	return h.AudioConn
}

func (h *RTSPHelper) RunAudioRTP() {
	bs, err := os.ReadFile("/Users/bytedance/Documents/go/my_tcp/rtsp/test.aac")
	utils.HandleErr(err)
	h.AudioData = bs // 为了方便索引先全部放到内存里
	for {
		time.Sleep(time.Second / (44100 / 1024))
		if !h.Run {
			continue
		}

		frame := h.ReadAudioFrame()
		conn := h.GetAudioConn()
		h.SendAudioFrame(conn, frame)
	}
}

func (h *RTSPHelper) SendAudioFrame(conn *net.UDPConn, frame []byte) {
	hdr := NewRTPHeader(h.AudioSeq, h.AudioTimestamp, RTPTypeAcc)
	data := make([]byte, 4)
	data[0] = 0x00
	data[1] = 0x10
	frameLen := len(frame)
	data[2] = byte((frameLen & 0x1FE0) >> 5) // 高 8 位
	data[3] = byte((frameLen & 0x1F) << 3)   // 低 5 位
	data = append(data, frame...)
	h.SendRTPHdrAndData(conn, hdr, data)
	h.AudioSeq++
	h.AudioTimestamp += 44100 / (44100 / 1024) // 一帧数据是 1024 个采样
}

func (h *RTSPHelper) ReadAudioFrame() []byte {
	if !IsValidAcc(h.AudioData[h.AudioDataIdx:]) {
		panic("err acc file")
	}
	frameLen := GetAccFrameLen(h.AudioData[h.AudioDataIdx:])
	start := h.AudioDataIdx + 7
	h.AudioDataIdx += frameLen
	return h.AudioData[start:h.AudioDataIdx]
}

func GetAccFrameLen(data []byte) int {
	return (int(data[3]&0x03) << 11) | (int(data[4]) << 3) | (int(data[5]&0xE0) >> 5)
}

func IsValidAcc(data []byte) bool {
	return data[0] == 0xFF && (data[1]&0xF0) == 0xF0
}

func (h *RTSPHelper) RunVideoRTP() {
	bs, err := os.ReadFile("/Users/bytedance/Documents/go/my_tcp/rtsp/test.h264")
	utils.HandleErr(err)
	h.VideoData = bs // 为了方便索引先全部放到内存里
	for {
		time.Sleep(time.Second / Fps)
		if !h.Run {
			continue
		}

		frame := h.ReadVideoFrame()
		conn := h.GetVideoConn()
		h.SendVideoFrame(conn, frame)
	}
}

func (h *RTSPHelper) ReadVideoFrame() []byte {
	// 每帧数据都是以 0 0 1 或 0 0 0 1 开始 以其为首尾读取一帧数据
	if IsCode3(h.VideoData[h.VideoDataIdx:]) {
		h.VideoDataIdx += 3
	} else if IsCode4(h.VideoData[h.VideoDataIdx:]) {
		h.VideoDataIdx += 4
	} else {
		panic(fmt.Sprintf("err h264 file"))
	}

	start := h.VideoDataIdx
	for h.VideoDataIdx < len(h.VideoData)-3 {
		if IsCode3(h.VideoData[h.VideoDataIdx:]) || IsCode4(h.VideoData[h.VideoDataIdx:]) {
			return h.VideoData[start:h.VideoDataIdx]
		}
		h.VideoDataIdx++
	}
	panic(fmt.Sprintf("play over"))
}

func IsCode4(data []byte) bool {
	return data[0] == 0 && data[1] == 0 && data[2] == 0 && data[3] == 1
}

func IsCode3(data []byte) bool {
	return data[0] == 0 && data[1] == 0 && data[2] == 1
}

type RTPHeader struct {
	Flag1     uint8
	Flag2     uint8
	Seq       uint16
	Timestamp uint32
	SSID      uint32
}

const (
	RTPTypeH264 = 96
	RTPTypeAcc  = 97
)

func NewRTPHeader(seq uint16, timestamp uint32, rtpType uint8) *RTPHeader {
	return &RTPHeader{
		Flag1:     0b10_0_0_0000,         // 版本 pad extension csID  只有版本选 2，其他都是 0
		Flag2:     0b1000_0000 | rtpType, // 指定负载类型
		Seq:       seq,
		Timestamp: timestamp,
		SSID:      0x22334455, // 用于区分多个连接源的这里，不需要区分统一写死
	}
}

func (h *RTSPHelper) SendVideoFrame(conn *net.UDPConn, frame []byte) {
	naluType := frame[0]
	if len(frame) > RTPMaxSize { // 分片传输
		l := RTPMaxSize - 2                              // 分片的话前两位要用于标记分片，不能用于数据传输
		for start := 1; start < len(frame); start += l { // 第一位可以不要了
			rtpHdr := NewRTPHeader(h.VideoSeq, h.VideoTimestamp, RTPTypeH264)
			data := make([]byte, 2) // 前两位标记分片
			data[0] = (naluType & 0x60) | 28
			data[1] = naluType & 0x1F
			if start == 1 {
				data[1] |= 0x80 // 开始标记
			} else if start+l >= len(frame) {
				data[1] |= 0x40 // 结束标记
			}
			end := min(start+l, len(frame))
			data = append(data, frame[start:end]...)
			h.SendRTPHdrAndData(conn, rtpHdr, data)
			h.VideoSeq++ // 每发一个包 seq 就要 ++
		}
	} else { // 直接传输
		rtpHdr := NewRTPHeader(h.VideoSeq, h.VideoTimestamp, RTPTypeH264)
		h.SendRTPHdrAndData(conn, rtpHdr, frame)
		h.VideoSeq++ // 每发一个包 seq 就要 ++
	} // Timestamp 每帧增加一个
	if (naluType&0x1F) != 7 && (naluType&0x1F) != 8 { // 如果是SPS、PPS就不需要加时间戳
		h.VideoTimestamp += 90000 / Fps // 90000 是 h264频率
	}
}

func (h *RTSPHelper) SendRTPHdrAndData(conn *net.UDPConn, hdr *RTPHeader, frame []byte) {
	buff := &bytes.Buffer{}
	err := binary.Write(buff, binary.BigEndian, hdr)
	utils.HandleErr(err)
	buff.Write(frame)

	_, err = conn.Write(buff.Bytes())
	utils.HandleErr(err)
}

func (s *RTSPServer) HandleConn(conn net.Conn) {
	localAddr := conn.LocalAddr().(*net.TCPAddr)
	remoteAddr := conn.RemoteAddr().(*net.TCPAddr)
	helper := &RTSPHelper{
		Session:        genSession(),
		Run:            false,
		LocalIP:        localAddr.IP.String(),
		RemoteIP:       remoteAddr.IP.String(),
		VideoDataIdx:   0,
		VideoSeq:       0,
		VideoTimestamp: 0,
		AudioDataIdx:   0,
		AudioSeq:       0,
		AudioTimestamp: 0,
	} // 本身是控制携程，还需要再打开对应数据携程  这里需要两个
	go helper.RunVideoRTP()
	go helper.RunAudioRTP()
	for {
		req := ParseReq(conn)
		var resp *RTSPResp
		switch req.Method {
		case "OPTIONS":
			resp = s.HandleOption(req)
		case "DESCRIBE":
			resp = s.HandleDescribe(req)
		case "SETUP":
			resp = s.HandleSetup(req, helper) // 在 Setup后的操作就需要传递 session 来区分客服端了，也可以一开始就传递
		case "PLAY":
			resp = s.HandlePlay(req, helper)
		default:
			panic(fmt.Sprintf("unknown method: %s", req.Method))
		}
		WriteResp(conn, resp)
	}
}

func WriteResp(conn net.Conn, resp *RTSPResp) {
	buff := &bytes.Buffer{}
	buff.WriteString(fmt.Sprintf("%s %d %s\r\n", resp.Version, resp.Code, resp.Msg))
	for k, v := range resp.Header {
		buff.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
	}
	buff.WriteString("\r\n")
	if len(resp.Data) > 0 {
		buff.Write(resp.Data)
	}
	fmt.Printf("======================Resp======================\n%s", buff.String()) // 测试输出
	_, err := conn.Write(buff.Bytes())
	utils.HandleErr(err)
}

func (s *RTSPServer) HandleOption(req *RTSPReq) *RTSPResp {
	header := make(map[string]string)
	header["Content-length"] = "0"                   // 没有 Data
	header["CSeq"] = req.Header["CSeq"]              // 必须保持一致
	header["Public"] = "OPTIONS,DESCRIBE,SETUP,PLAY" // 支持的参数
	header["Server"] = ServerName
	return &RTSPResp{
		Version: req.Version,
		Code:    200,
		Msg:     "OK",
		Header:  header,
	}
}

func getSDPData(req *RTSPReq) []byte { // 暂时只支持一个固定视频的推流 这里获取其信息
	buff := &bytes.Buffer{}
	buff.WriteString("v=0\r\n")
	buff.WriteString("o=- 16902775634812346707 1 IN IP4 127.0.0.1\r\n")
	buff.WriteString("t=0 0\r\n")
	buff.WriteString(fmt.Sprintf("a=control:%s\r\n", req.Path))
	// 视频信息
	buff.WriteString("m=video 0 RTP/AVP 96\r\n")
	buff.WriteString("a=rtpmap:96 H264/90000\r\n")
	buff.WriteString(fmt.Sprintf("a=control:%s/%s\r\n", req.Path, VideoTrack))
	// 音频信息
	buff.WriteString("m=audio 0 RTP/AVP 97\r\n")
	buff.WriteString("a=rtpmap:97 mpeg4-generic/44100/2\r\n")
	//  a=fmtp:96 streamtype=5; profile-level-id=15; mode=AAC-hbr; config=1190; SizeLength=13; IndexLength=3; IndexDeltaLength=3; Profile=1;
	buff.WriteString("a=fmtp:97 profile-level-id=1;mode=AAC-hbr;SizeLength=13;IndexLength=3;IndexDeltaLength=3;config=1210;\r\n")
	buff.WriteString(fmt.Sprintf("a=control:%s/%s\r\n", req.Path, AudioTrack))
	return buff.Bytes()
}

func (s *RTSPServer) HandleDescribe(req *RTSPReq) *RTSPResp {
	data := getSDPData(req)
	header := make(map[string]string)
	header["Content-type"] = req.Header["Accept"]
	header["Content-Base"] = req.Path
	header["Content-length"] = strconv.Itoa(len(data))
	header["CSeq"] = req.Header["CSeq"]
	header["Server"] = ServerName
	return &RTSPResp{
		Version: req.Version,
		Code:    200,
		Msg:     "OK",
		Header:  header,
		Data:    data,
	}
}

func (s *RTSPServer) HandleSetup(req *RTSPReq, helper *RTSPHelper) *RTSPResp {
	transport := req.Header["Transport"]
	reg, err := regexp.Compile("client_port=(\\d+)-(\\d+)")
	utils.HandleErr(err)
	items := reg.FindStringSubmatch(transport)
	rtpPort, err := strconv.Atoi(items[1])
	utils.HandleErr(err)
	rtcpPort, err := strconv.Atoi(items[2])
	utils.HandleErr(err)
	var localRTPPort, localRTCPPort int
	if strings.HasSuffix(req.Path, VideoTrack) {
		helper.VideoRTPPort = rtpPort
		helper.VideoRTCPPort = rtcpPort
		localRTPPort, localRTCPPort = RTPVideoPort, RTCPVideoPort
	} else if strings.HasSuffix(req.Path, AudioTrack) {
		helper.AudioRTPPort = rtpPort
		helper.AudioRTCPPort = rtcpPort
		localRTPPort, localRTCPPort = RTPAudioPort, RTCPAudioPort
	} else {
		panic(fmt.Sprintf("unknown req path %s", req.Path))
	}
	header := make(map[string]string)
	header["Transport"] = fmt.Sprintf("RTP/AVP/UDP;unicast;client_port=%d-%d;server_port=%d-%d",
		rtpPort, rtcpPort, localRTPPort, localRTCPPort)
	header["Session"] = helper.Session
	header["Content-length"] = "0"
	header["CSeq"] = req.Header["CSeq"]
	header["Server"] = ServerName
	return &RTSPResp{
		Version: req.Version,
		Code:    200,
		Msg:     "OK",
		Header:  header,
	}
}

func (s *RTSPServer) HandlePlay(req *RTSPReq, data *RTSPHelper) *RTSPResp {
	data.Run = true // 开始播放
	header := make(map[string]string)
	header["Range"] = req.Header["Range"]
	header["Content-length"] = "0"
	header["Session"] = data.Session
	header["CSeq"] = req.Header["CSeq"]
	header["Server"] = ServerName
	return &RTSPResp{
		Version: req.Version,
		Code:    200,
		Msg:     "OK",
		Header:  header,
	}
}
