package main

import (
	"fmt"
	"net"
	"strings"
	"time"
)

type Client struct {
	C    chan string //用户发送信息的管道
	Name string      //用户名
	Addr string      //网络地址
}

//保存在线用户  ClientAddr =====> Client
var onlineMap = make(map[string]Client)

//提示用户退出
var isQuit = make(chan bool)

//超时处理，对方是否有数据发送
var hasData = make(chan bool)

//消息管道
var message = make(chan string)

func Manager() {
	for {
		msg := <-message //没有消息的时候这里会阻塞
		//遍历onlineMap ,给每个成员发送消息
		for _, cli := range onlineMap {
			cli.C <- msg

		}

	}
}

func WriteMsgToClient(cli Client, conn net.Conn) {

	for msg := range cli.C {
		conn.Write([]byte(msg + "\n"))
	}
}

func MakeMsg(cli Client, msg string) (buf string) {
	buf = "[" + cli.Addr + "]" + cli.Name + ":" + msg
	return
}

//处理用户链接
func HandleConn(conn net.Conn) {
	defer conn.Close()
	//获取客户端网络地址
	ClientAddr := conn.RemoteAddr().String()
	//创建结构体，默认用户名和地址一样
	cli := Client{
		make(chan string),
		ClientAddr,
		ClientAddr,
	}
	//把接构体加入到onlineMap中
	onlineMap[ClientAddr] = cli
	//新开一个协程专门给当前用户发送消息
	go WriteMsgToClient(cli, conn)
	//广播某个人在线
	message <- MakeMsg(cli, "login")
	//提示我是谁
	cli.C <- MakeMsg(cli, "i am here")
	//阻塞防止退出

	//新建一个协程处理用户发送过来的数据
	go func() {
		buf := make([]byte, 2048)
		for {
			n, err := conn.Read(buf)
			if n == 0 { //对方断开或者出问题
				isQuit <- true
				fmt.Printf("conn read err:%v\n", err)
				return
			}
			msg := string(buf[:n])

			if len(msg) == 3 && msg == "who" {
				//遍历onlineMap给当前用户发送所有成员
				conn.Write([]byte("user list:\n"))
				for _, tmp := range onlineMap {
					msg = tmp.Addr + ":" + tmp.Name + "\n"
					conn.Write([]byte(msg))
				}
			} else if len(msg) >= 8 && msg[:6] == "rename" { //查看用户列表
				//rename|mike
				name := strings.Split(msg, "|")[1]
				cli.Name = name
				onlineMap[ClientAddr] = cli
				conn.Write([]byte("rename ok!\n"))

			} else {
				message <- MakeMsg(cli, msg)
			}

			hasData <- true

		}

	}()
	for {
		//通过select检测channel的流动
		select {
		case <-isQuit:
			delete(onlineMap, ClientAddr)
			message <- MakeMsg(cli, "login out") //广播谁下线了
			return
		case <-hasData:

		case <-time.After(30 * time.Second): //30s后没有动作就会被移除
			delete(onlineMap, ClientAddr)
			message <- MakeMsg(cli, "time out leave out!")
		}
	}
}

func main() {
	//监听
	listener, err := net.Listen("tcp", "localhost:8000")
	if err != nil {
		fmt.Println("net.Listen err = ", err)
		return
	}
	defer listener.Close()

	//转发消息，只要一有消息，就遍历onlineMap,给其中的每一个成员发送消息
	go Manager()

	//循环等待用户链接
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("listener.Accept err = ", err)
			continue
		}
		go HandleConn(conn) //处理用户链接
	}
}
