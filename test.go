package main

import (
    "golang.org/x/net/websocket"
    "fmt"
    "log"
    "net/http"
	"strings"
	"github.com/hpcloud/tail"
)

func echoHandler(ws *websocket.Conn) {
    msg := make([]byte, 512)
    n, err := ws.Read(msg)//将websocket收到的消息读到msg中
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Receive: %s,len=%d\n", msg[:n],n)//在命令行打印收到的消息和长度
	var p int
	for i:=2;i<n;i++{
		if msg[i]==' '	{
			p = i  //找到空格的位置
			break
		}
	}
	path:= string(msg[:p])//空格前的部分是文件名
	word:= string(msg[p+1:n])//空格后的部分是关键字
	update,_ := tail.TailFile(path,tail.Config{
		Follow:true,
		ReOpen:true})//用tail对文件进行追踪
	for line:= range update.Lines{
		if strings.Contains(line.Text,word){//如果一行中包含关键字，则将该行传回服务器
			ws.Write([]byte(line.Text))
		}
	}
}

func main() {
	
    http.Handle("/echo", websocket.Handler(echoHandler))//指定websocket连接的处理方式，echo是指定的匹配模式
    http.Handle("/", http.FileServer(http.Dir(".")))//对于其他请求，我们根据所在目录文件系统的内容进行处理

    err := http.ListenAndServe(":8080", nil)//监听指定的TCP地址,nil表示使用默认的handler

    if err != nil {
        panic("ListenAndServe: " + err.Error())
    }
}