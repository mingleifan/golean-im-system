package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"
)

type Client struct {
	ServerIp   string
	ServerPort int
	Name       string
	conn       net.Conn
	mode       int
}

func NewClient(serverIp string, serverPort int) *Client {
	//创建客户端对象
	client := &Client{
		ServerIp:   serverIp,
		ServerPort: serverPort,
		mode:       999,
	}
	//链接server
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", serverIp, serverPort))
	if err != nil {
		fmt.Println("net.Dial error:", err)
		return nil
	}
	client.conn = conn
	//返回对象
	return client
}

func (client *Client) UpdateName() bool {
	fmt.Println(">>>>请输入用户名<<<<")
	fmt.Scanln(&client.Name)

	sendMsg := fmt.Sprintf("rename|%s\n", client.Name)

	_, err := client.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("conn.Write err:", err)
		return false
	}
	return true
}

// 处理server响应的信息，直接显示到标准输出
func (client *Client) handleServerResp() {
	//一旦client.conn有数据，就直接copy到stdout标准输出上，拥挤阻塞监听
	io.Copy(os.Stdout, client.conn)
}

func (client *Client) menu() bool {
	var selectedMode int

	fmt.Println("1.公聊模式")
	fmt.Println("2.私聊模式")
	fmt.Println("3.更新用户名")
	fmt.Println("0.退出")

	fmt.Scanln(&selectedMode)

	if selectedMode >= 0 && selectedMode <= 3 {
		client.mode = selectedMode
		return true
	} else {
		fmt.Println(">>>>请输入合法范围内的数字<<<<")
		return false
	}
}

func (client *Client) Run() {
	for client.mode != 0 {
		for client.menu() != true {

		}
		//根据不同的模式处理不同的业务
		switch client.mode {
		case 1:
			//公聊模式
			break
		case 2:
			//私聊模式
			break
		case 3:
			//更新用户名
			client.UpdateName()
			break
		case 0:
			//退出
			break
		}
	}
	fmt.Println("退出客户端")
}

var serverIp string
var serverPort int

// ./im-client -ip 127.0.0.1 -port 8888
func init() {
	flag.StringVar(&serverIp, "ip", "127.0.0.1", "设置服务器IP地址(默认是127.0.0.1)")
	flag.IntVar(&serverPort, "port", 8888, "设置服务器端口(默认是8888)")
}

func main() {
	//命令行解析
	flag.Parse()

	client := NewClient(serverIp, serverPort)
	if client == nil {
		fmt.Println(">>>>>>>>>>>连接服务器失败<<<<<<<<<<<")
		return
	}

	//单独开启一个goroutine，处理server回执的消息
	go client.handleServerResp()

	fmt.Println(">>>>>>>>>>>连接服务器成功<<<<<<<<<<<")

	//启动客户端的业务
	client.Run()
}
