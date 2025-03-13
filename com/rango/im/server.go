package im

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

type Server struct {
	Ip   string
	Port int

	// 在线用户列表
	OnlineMap map[string]*User
	mapLock   sync.RWMutex

	//消息广播的channel
	Message chan string
}

// NewServer 创建一个server的接口
func NewServer(ip string, port int) *Server {
	server := &Server{
		Ip:        ip,
		Port:      port,
		OnlineMap: make(map[string]*User),
		Message:   make(chan string),
	}
	return server
}

// ListenMessage 监听Message广播消息channel的goroutine，一旦有消息就发送给全部在线的user
func (server *Server) ListenMessage() {
	for {
		msg := <-server.Message
		//将msg发送给全部的在线User
		server.mapLock.Lock()
		for _, user := range server.OnlineMap {
			user.C <- msg
		}
		server.mapLock.Unlock()
	}
}

// BroadCast 广播消息
func (server *Server) BroadCast(user *User, msg string) {
	sendMsg := fmt.Sprintf("[%s] %s : %s", user.Addr, user.Name, msg)
	server.Message <- sendMsg
}

func (server *Server) Handler(conn net.Conn) {
	//...当前链接的业务
	fmt.Printf("链接建立成功, remote addr:%v \n", conn.RemoteAddr())

	user := NewUser(conn, server)
	//用户上线，将用户加入到onlineMap中
	user.Online()

	//监听用户是否活跃的channel
	isLive := make(chan bool)

	//接受客户端发送的消息
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := conn.Read(buf)
			if n == 0 {
				user.Offline()
				return
			}
			if err != nil && err != io.EOF {
				fmt.Println("Conn Read err:", err)
				return
			}
			//提取用户的消息（去除'\n'）
			msg := string(buf[:n-1])
			//将得到的消息进行广播
			user.DoMessage(msg)

			//用户的任意消息，代表当前用户是一个活跃的
			isLive <- true
		}
	}()

	//当前handler阻塞
	for {
		select {
		case <-isLive:
			//当前用户是活跃的，应该重置定时器
			//不做任何事情，为了激活select，更新下面的定时器
		case <-time.After(time.Second * 10):
			//已经超时
			//将当前的User强制关闭
			user.SendMessage("你被踢了...\n")
			//销毁用户的资源
			close(user.C)
			//关闭连接
			conn.Close()
			//推出当前Handler
			return
		}

	}
}

// Start 启动服务器的接口
func (server *Server) Start() {
	//socket listen
	listen, err := net.Listen("tcp", fmt.Sprintf("%s:%d", server.Ip, server.Port))
	if err != nil {
		fmt.Println("net.Listen error:", err)
		return
	}
	//close listen socket
	defer listen.Close()

	//启动监听Message的goroutine
	go server.ListenMessage()

	for {
		//accept
		conn, err := listen.Accept()
		if err != nil {
			fmt.Println("listener accept err:", err)
			continue
		}
		//do handler
		go server.Handler(conn)
	}

}
