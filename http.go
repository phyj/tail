package main

import (
	"encoding/json"
    "log"
	"fmt"
    "net/http"
	"strings"
	"github.com/hpcloud/tail"
	"strconv"
)

var	http_dir string

func http_set_dir(dir string){//设置所在目录
	http_dir = dir
}

func hello(w http.ResponseWriter,req *http.Request){
	log.Println("a http request")
	err := req.ParseForm()//对query_string进行解包,query_string的格式为file=x&limit=y&word=z
	if(err!=nil){
		log.Println(err.Error())
		return 
	}
	limit,errr := strconv.Atoi(req.Form["limit"][0])//limit为输出的最大行数
	if(limit<=0){
		return 
	}
	if(errr!=nil){
		log.Println(errr.Error())
		return 
	}
	file_path := http_dir+req.Form["file"][0]//file为要tail的文件的路径
	key_word := req.Form["word"][0]
	lines := make([]string,limit)//lines存储符合条件的行
	//fmt.Println("file=%s",file)
	update,er := tail.TailFile(file_path,tail.Config{
	MustExist:true})
	if(er!=nil){
		log.Println(er.Error())
		return 
	}
	defer func() {
		update.Stop()
		update.Cleanup()
	}()
	var line_cnt int = 0//统计符合条件的行的数量
	var pos int = 0//指向下一个用来存储的位置
	for line:= range update.Lines{
		if	strings.Contains(line.Text,key_word){
			line_cnt++
			lines[pos] = line.Text
			pos++
			if pos==limit{//循环存储，节省空间
				pos = 0
			}
		}
	}
	if line_cnt>=limit {
		result := make([]string,limit)
		for i:=0;i<limit;i++{
			result[i] = lines[pos]
			pos++
			if pos==limit{
				pos = 0
			}
		}
		js,error := json.Marshal(result)
		if error==nil{
			fmt.Fprintf(w,"{\"lines\":"+string(js)+"}")
		}
	}else{
		result := make([]string,line_cnt)
		for i:=0;i<pos;i++{
			result[i] = lines[i]
		}
		js,error := json.Marshal(result)
		if error==nil{
			fmt.Fprintf(w,"{\"lines\":"+string(js)+"}")
		}
	}
}