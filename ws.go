package main

import (
    "golang.org/x/net/websocket"
    "log"
	"strings"
	"github.com/hpcloud/tail"
)
const (
	buffer_length = 512
)

var ws_dir string

func ws_set_dir(dir string){
	ws_dir = dir
}

func monitor(ws *websocket.Conn,tail *tail.Tail){//监控ws是否断开，若断开则停止tail
	msg := make([]byte, buffer_length)
	for {
		_,err := ws.Read(msg)
		if err!=nil&&err.Error()=="EOF"{
			tail.Stop()
			tail.Cleanup()
			return 
		}
	}
}

func echoHandler(ws *websocket.Conn) {
	defer ws.Close()
	msg := make([]byte,buffer_length)
	msg_length, err := ws.Read(msg)//将websocket收到的消息读到msg中
	if err != nil {
		log.Println(err)
		return 
	}
	//msg中file_path与key_word用一个空格分离
	var space_pos int = -1
	for i:=1;i<msg_length;i++{
		if msg[i]==' '	{
			space_pos = i  //找到空格的位置
			break
		}
	}
	if space_pos == -1{
		log.Println("illegal message")
		return 
	}
	file_path := ws_dir+string(msg[:space_pos])
	key_word := string(msg[space_pos+1:msg_length])
	update,er := tail.TailFile(file_path,tail.Config{
	Follow:true,
	ReOpen:true})//用tail对文件进行追踪
	if er!=nil{
		log.Fatal(er)
		return 
	}
	defer func() {
		update.Stop()
		update.Cleanup()
	}()
	go monitor(ws,update)
	for line:= range update.Lines{
		if strings.Contains(line.Text,key_word){//如果一行中包含关键字，则将该行输出到客户端
			_,errr := ws.Write([]byte(line.Text))
			if errr!=nil {
				log.Println(errr)
				break
			}
		}
	}
}