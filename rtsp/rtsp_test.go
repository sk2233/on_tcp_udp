/*
@author: sk
@date: 2024/9/16
*/
package rtsp

import "testing"

// 解封装，方便推流
// ffmpeg -i test.mp4 -codec copy -bsf: h264_mp4toannexb -f h264 test.h264
// 获取 acc音频流
// ffmpeg -i test.mp4 -vn -acodec aac test.aac
// 拉流指令
// ffplay -i rtsp://127.0.0.1:8554/test

func TestRTSP(t *testing.T) {
	server := NewRTSPServer("127.0.0.1:8554")
	server.Listen()
}
