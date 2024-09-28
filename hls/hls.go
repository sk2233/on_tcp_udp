/*
@author: sk
@date: 2024/9/16
*/
package hls

import (
	"fmt"
	"io"
	"my_tcp/utils"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

type TS struct {
	Duration float64
	Desc     string
	Name     string
}

type M3U8 struct {
	Duration float64
	Sequence int
	Ts       []*TS
}

func ParseM3U8(bs []byte) *M3U8 {
	lines := strings.Split(string(bs), "\n")
	res := &M3U8{}
	index := 0
	var err error
	for index < len(lines) {
		if !strings.HasPrefix(lines[index], "#EXT") {
			continue
		}
		line := lines[index]
		switch {
		case strings.HasPrefix(line, "#EXTM3U"): // 开始标记
		case strings.HasPrefix(line, "#EXT-X-TARGETDURATION"): // 每个片段持续时间
			i := strings.Index(line, ":")
			res.Duration, err = strconv.ParseFloat(line[i+1:], 64)
			utils.HandleErr(err)
		case strings.HasPrefix(line, "#EXT-X-MEDIA-SEQUENCE"): // 序号
			i := strings.Index(line, ":")
			res.Sequence, err = strconv.Atoi(line[i+1:])
			utils.HandleErr(err)
		case strings.HasPrefix(line, "#EXTINF"): // 片段标记
			i := strings.Index(line, ":")
			items := strings.Split(line[i+1:], ",")
			duration, err := strconv.ParseFloat(items[0], 64)
			utils.HandleErr(err)
			index++
			res.Ts = append(res.Ts, &TS{
				Duration: duration,
				Desc:     items[1],
				Name:     lines[index],
			})
		case strings.HasPrefix(line, "#EXT-X-ENDLIST"): // 结束标记
			break
		}
		index++
	}
	return res
}

func DownloadHLS(url string, savePath string) {
	err := os.MkdirAll(savePath, os.ModePerm)
	utils.HandleErr(err)

	resp, err := http.Get(url)
	utils.HandleErr(err)
	bs, err := io.ReadAll(resp.Body)
	utils.HandleErr(err)
	resp.Body.Close()
	err = os.WriteFile(filepath.Join(savePath, "index.m3u8"), bs, os.ModePerm)
	utils.HandleErr(err)

	m3u8 := ParseM3U8(bs)
	index := strings.LastIndex(url, "/")
	baseUrl := url[:index]
	fmt.Printf("download m3u8 index item duration = %v , ts count = %v\n", m3u8.Duration, len(m3u8.Ts))
	for i, ts := range m3u8.Ts {
		resp, err = http.Get(baseUrl + "/" + ts.Name)
		utils.HandleErr(err)
		bs, err = io.ReadAll(resp.Body)
		utils.HandleErr(err)
		resp.Body.Close()
		err = os.WriteFile(filepath.Join(savePath, ts.Name), bs, os.ModePerm)
		utils.HandleErr(err)
		fmt.Printf("download ts %s %d/%d\n", ts.Name, i+1, len(m3u8.Ts))
	}
}

// ffmpeg -i "concat:input1.mpg|input2.mpg|input3.mpg" -c copy output.mpg
func MergeHLS(savePath string) {
	items, err := os.ReadDir(savePath)
	utils.HandleErr(err)
	buff := &strings.Builder{}
	buff.WriteString("concat:")
	for _, item := range items {
		if filepath.Ext(item.Name()) != ".ts" {
			continue
		}
		buff.WriteString(filepath.Join(savePath, item.Name()))
		buff.WriteRune('|')
	}
	input := buff.String()
	cmdArgs := []string{"-i", input[:len(input)-1], "-c", "copy", filepath.Join(savePath, "out.mp4")}
	fmt.Println("ffmpeg", strings.Join(cmdArgs, " "))
	err = exec.Command("ffmpeg", cmdArgs...).Run()
	utils.HandleErr(err)
}
