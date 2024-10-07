/*
@author: sk
@date: 2024/9/16
*/
package hls

import (
	"fmt"
	"my_tcp/utils"
	"os/exec"
	"strings"
	"testing"
)

func TestHLS(t *testing.T) { // 这里只测试了点播，直播的 m3u8文件是会动态变化的
	m3u8Url := "http://playertest.longtailvideo.com/adaptive/bipbop/gear4/prog_index.m3u8"
	savePath := utils.BasePath + "hls/prog"
	fmt.Println(m3u8Url, savePath)
	//DownloadHLS(m3u8Url, savePath)
	MergeHLS(savePath)
}

// ffmpeg -i concat:fileSequence0.ts|fileSequence1.ts -c copy output.mp4
func TestIt(t *testing.T) {
	// 定义需要合并的视频文件列表
	videos := []string{utils.BasePath + "hls/prog/fileSequence0.ts", utils.BasePath + "hls/prog/fileSequence1.ts"}
	// 拼接视频命令
	cmdArgs := []string{"-i"}
	// 添加每个视频文件到拼接命令中
	buff := &strings.Builder{}
	buff.WriteString("concat:")
	for _, video := range videos {
		buff.WriteString(video)
		buff.WriteString("|")
	}
	input := buff.String()
	// 输出文件名
	output := utils.BasePath + "hls/prog/output.mp4"
	// 添加输出文件到拼接命令中
	cmdArgs = []string{"-i", input[:len(input)-1], "-c", "copy", output}
	// 执行FFmpeg命令
	cmd := exec.Command("ffmpeg", cmdArgs...)
	err := cmd.Run()
	if err != nil {
		panic(err)
	}
}
