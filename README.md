# tail
用法：
go build test.go
在命令行输入 test.exe log.txt key_word,会实时tail名为log.txt的文件中所有包含key_word的行，用浏览器打开localhost:8080/blank.html即可查看

用浏览器打开http://localhost:8080/tail?limit=1000&file=out.go  会tail名为out.go的文件的前1000行，limit=后接输出的最大行数，file=后接要tail的文件相对路径；不会实时更新新的内容
