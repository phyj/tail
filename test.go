package main

import (
    "golang.org/x/net/websocket"
    "fmt"
    "log"
    "net/http"
	"strings"
	"github.com/hpcloud/tail"
	"time"
	"os"
	"strconv"
)

const (
	gap = 2
	timeout = 6000
	greet = ""
)

type hub struct {
	// Registered connections.
	connections map[*websocket.Conn]bool

	// Inbound messages from the connections.
	broadcast chan string

	// Register requests from the connections.
	register chan *websocket.Conn

	// Unregister requests from connections.
	unregister chan *websocket.Conn
}

var (
	set bool
	s_path string 
	s_word string
)
var h = hub{
	broadcast:   make(chan string),
	register:    make(chan *websocket.Conn),
	unregister:  make(chan *websocket.Conn),
	connections: make(map[*websocket.Conn]bool),
}

func (h *hub) run() {
	for {
		select {
		case c := <-h.register:
			h.connections[c] = true
		case c := <-h.unregister:
			delete(h.connections, c)
		case m := <-h.broadcast:
			for c := range h.connections {
				_,err := c.Write([]byte(m))
				if(err!=nil){
					delete(h.connections,c)
				}
			}
		}
	}
}

func watch() {//tail指定的文件，将文件中所有包含关键字的行广播到所有websocket连接
	update,er := tail.TailFile(s_path,tail.Config{
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
	for line:= range update.Lines{
		if strings.Contains(line.Text,s_word){//如果一行中包含关键字，则将该行输出到所有websocket连接
			h.broadcast <- line.Text
		}
	}
}
func monitor(ws *websocket.Conn,tail *tail.Tail,ti *time.Time){
	for{
		time.Sleep(gap*time.Second)//每次sleep了gap秒之后，检查websocket是否断开，是否等待文件更新过久
		_,err := ws.Write([]byte(greet))
		if err!=nil || time.Now().Sub(*ti)>timeout*time.Second {
			tail.Stop()
			tail.Cleanup()
			return
		}
		//fmt.Println("watch now")
	}
}

func echoHandler(ws *websocket.Conn) {
	fmt.Println("one in")
	h.register <- ws
	defer func() {
		fmt.Println("one out");
		h.unregister <- ws 
		ws.Close()
	}()
	if set==false{
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
		path := string(msg[:p])
		word := string(msg[p+1:n])
		update,er := tail.TailFile(path,tail.Config{
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
		ti := time.Now()//用来记录目标文件最近一次被修改的时间
		go monitor(ws,update,&ti)
		for line:= range update.Lines{
			ti = time.Now()//文件有更新，更新ti
			if strings.Contains(line.Text,word){//如果一行中包含关键字，则将该行输出到客户端
				_,errr := ws.Write([]byte(line.Text))
				if errr!=nil {
					update.Stop()
					update.Cleanup()
					break
				}
			}
		}
	}else{
		for{
			time.Sleep(gap*time.Second)//每次sleep了gap秒之后，检查websocket是否断开
			_,err := ws.Write([]byte(greet))
			if err!=nil {
				break
			}
		}
	}
}

func hello(w http.ResponseWriter,r *http.Request){
	err := r.ParseForm()
	if(err!=nil){
		fmt.Fprintf(w,err.Error())
		return 
	}
	/*fmt.Println(r.Form)
	fmt.Println("path",r.URL.Path)
	fmt.Println("scheme",r.URL.Scheme)
	fmt.Println(r.Form["url_long"])*/
	limit,errr := strconv.Atoi(r.Form["limit"][0])//limit为输出的最大行数
	if(errr!=nil){
		fmt.Fprintf(w,errr.Error())
		return 
	}
	file := r.Form["file"][0]//file为要tail的文件的路径
	update,er := tail.TailFile(file,tail.Config{
	MustExist:true})
	if(er!=nil){
		fmt.Fprintf(w,er.Error())
		return 
	}
	defer func() {
		update.Stop()
		update.Cleanup()
	}()
	var cnt int = 0
	for line:= range update.Lines{
		fmt.Fprintln(w,line.Text)
		cnt++
		if cnt>=limit{
			break
		}
	}
	/*fmt.Println("limit=%s,file=%s",limit,file)
	for k,v := range r.Form{
		fmt.Println("key:",k)
		fmt.Println("val:",strings.Join(v,""))
	}
	fmt.Fprintf(w,"hello\n")
	fmt.Fprintf(w,"world\n")*/
}

func main() {
	if len(os.Args)>1 {
		set = true
		s_path = os.Args[1];//第一个参数存的是路径
		if len(os.Args)>2{
			s_word = os.Args[2]//第二个参数存的是关键字
		} else {
			s_word = ""
		}
	} else{
		set = false
	}
	go h.run()//h维护所有websocket连接，用于广播
	if set==true {
		go watch()
	}
	http.HandleFunc("/tail",hello)
    http.Handle("/echo", websocket.Handler(echoHandler))//指定websocket连接的处理方式，echo是指定的匹配模式
    http.Handle("/", http.FileServer(http.Dir(".")))//对于其他请求，我们根据所在目录文件系统的内容进行处理

    err := http.ListenAndServe(":8080", nil)//监听指定的TCP地址,nil表示使用默认的handler

    if err != nil {
        panic("ListenAndServe: " + err.Error())
    }
}