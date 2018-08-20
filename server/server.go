package main

import (
	"fmt"
	"net"
	"strings"
	"time"
)

//用户信息
type Client struct {
	C    chan string
	Name string
	Addr string
}

//用于存储在线用户
var onlineMap = make(map[string]Client)

var message = make(chan string)

var isQuit = make(chan bool)
var hasData = make(chan bool)

func MakeMsg(cli Client, msg string) (str string) {
	str = "[" + cli.Addr + "]" + cli.Name + ":" + msg
	return
}

func Manager() {
	for {
		msg := <-message
		for _, cli := range onlineMap {
			cli.C <- msg
		}
	}
}

func HandleConn(conn net.Conn) {
	defer conn.Close()

	clientAddr := conn.RemoteAddr().String()

	cli := Client{
		make(chan string),
		clientAddr,
		clientAddr,
	}

	onlineMap[clientAddr] = cli

	go func() {
		buf := make([]byte, 2048)
		for {
			n, err := conn.Read(buf)
			if n == 0 { //对方断开或者出问题
				isQuit <- true
				fmt.Printf("conn read msg from client failed,err:%v\n", err)
				return
			}
			msg := string(buf[:n])

			if len(msg) == 3 && msg == "who" {
				conn.Write([]byte("user list:\n"))
				for _, tmp := range onlineMap {
					msg = tmp.Addr + ":" + tmp.Name + "\n"
					conn.Write([]byte(msg))

				}
			} else if len(msg) >= 8 && msg[:6] == "rename" {
				name := strings.Split(msg, "|")[1]
				cli.Name = name
				onlineMap[clientAddr] = cli
				conn.Write([]byte("rename ok\n"))
			} else {
				message <- MakeMsg(cli, msg)
			}
			hasData <- true
		}
	}()

	go WriteMsgToClient(cli, conn)

	message <- MakeMsg(cli, "login")

	cli.C <- MakeMsg(cli, "i am here")

	for {
		select {
		case <-isQuit:
			delete(onlineMap, clientAddr)
			message <- MakeMsg(cli, "log out")
		case <-hasData:

		case <-time.After(30 * time.Second):
			delete(onlineMap, clientAddr)
			message <- MakeMsg(cli, "time out leave out!")
		}
	}
}

func WriteMsgToClient(cli Client, conn net.Conn) {
	for msg := range cli.C {
		conn.Write([]byte(msg + "\n"))
	}
}

func main() {
	listen, err := net.Listen("tcp", "localhost:8080")
	if err != nil {
		fmt.Printf("net.Listen err:%v\n", err)
		return
	}
	defer listen.Close()

	go Manager()

	for {
		conn, err := listen.Accept()
		if err != nil {
			fmt.Printf("listen.Accept err:%v\n", err)
			continue
		}

		go HandleConn(conn)

	}
}
