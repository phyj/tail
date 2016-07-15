package main

import (
    "golang.org/x/net/websocket"
    "fmt"
    "log"
    "net/http"
	"strings"
	"github.com/hpcloud/tail"
	"time"
)

const (
	gap = 2
	timeout = 10
	greet = ""
)

func monitor(ws *websocket.Conn,tail *tail.Tail,ti *time.Time){
	for{
		time.Sleep(gap*time.Second)//每次sleep了gap秒之后，检查websocket是否断开，是否等待文件更新过久
		_,err := ws.Write([]byte(greet))
		if err!=nil || time.Now().Sub(*ti)>timeout*time.Second {
			tail.Stop()
			tail.Cleanup()
			return
		}
		fmt.Println("watch now")
	}
}

func echoHandler(ws *websocket.Conn) {
	fmt.Println("one in")
    msg := make([]byte, 512)
    n, err := ws.Read(msg)//将websocket收到的消息读到msg中
    if err != nil {
        log.Println(err)
		return 
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
	update,er := tail.TailFile(path,tail.Config{
		Follow:true,
		ReOpen:true})//用tail对文件进行追踪
	if er!=nil{
		log.Fatal(er)
		return 
	}
	ti := time.Now()//用来记录目标文件最近一次被修改的时间
	go monitor(ws,update,&ti)
	for line:= range update.Lines{
		ti = time.Now()//文件有更新，更新ti
		if strings.Contains(line.Text,word){//如果一行中包含关键字，则将该行传回服务器
			_,errr := ws.Write([]byte(line.Text))
			if errr!=nil {
				update.Stop()
				update.Cleanup()
				break
			}
		}
	}
	fmt.Println("one out");
	defer ws.Close()
}

func main() {
	
    http.Handle("/echo", websocket.Handler(echoHandler))//指定websocket连接的处理方式，echo是指定的匹配模式
    http.Handle("/", http.FileServer(http.Dir(".")))//对于其他请求，我们根据所在目录文件系统的内容进行处理

    err := http.ListenAndServe(":8080", nil)//监听指定的TCP地址,nil表示使用默认的handler

    if err != nil {
        panic("ListenAndServe: " + err.Error())
    }
}