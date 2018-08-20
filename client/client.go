package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
)

func processConn(conn net.Conn) {
	var input string
	defer conn.Close()
	for {
		reader := bufio.NewReader(os.Stdin)
		data, _, _ := reader.ReadLine()
		input = string(data)

		_, err := conn.Write([]byte(input))
		if err != nil {
			fmt.Printf("send message to server failed,err:%v\n", err)
			return
		}
	}
}

func main() {
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		fmt.Printf("conn server failed,err:%v\n", err)
		return
	}
	defer conn.Close()

	go processConn(conn)

	buf := make([]byte, 1024)
	for {
		num, err := conn.Read(buf)
		if err != nil {
			fmt.Printf("received message from server failed,err:%v\n", err)
			return
		}
		fmt.Print(string(buf[:num]))
	}
}
