/*
@author: sk
@date: 2024/9/16
*/
package ftp

import (
	"my_tcp/utils"
	"os"
	"testing"
)

func TestFTP(t *testing.T) {
	conf := utils.GetConf()
	client := NewFTPClient(conf.FTPAddr)
	client.Login(conf.FTPUser, conf.FTPPasswd)

	//items := client.List("./")
	//for _, item := range items {
	//	fmt.Println(item)
	//}
	//client.Del("new.jpeg")
	//out := client.Get("20240123-184937.jpeg")
	//open, err := os.Create(utils.BasePath + "ftp/temp.jpeg")
	//utils.HandleErr(err)
	//_, err = io.Copy(open, out)
	//utils.HandleErr(err)
	in, err := os.Open(utils.BasePath + "ftp/temp.jpeg")
	utils.HandleErr(err)
	client.Sto("new.jpeg", in)

	client.Quit()
}
