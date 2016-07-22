package main

import (
    "golang.org/x/net/websocket"
    "net/http"
	"os"
)


func main() {
	if len(os.Args)>1 {
		ws_set_dir(os.Args[1])
		http_set_dir(os.Args[1])
	}
	http.HandleFunc("/tail",hello)
    http.Handle("/echo", websocket.Handler(echoHandler))//指定websocket连接的处理方式，echo是指定的匹配模式
    http.Handle("/", http.FileServer(http.Dir(".")))//对于其他请求，我们根据所在目录文件系统的内容进行处理

    err := http.ListenAndServe(":8080", nil)//监听指定的TCP地址,nil表示使用默认的handler

    if err != nil {
        panic("ListenAndServe: " + err.Error())
    }
}