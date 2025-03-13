package im

import (
	"fmt"
	"net"
	"sync"
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

	user := NewUser(conn)
	//用户上线，将用户加入到onlineMap中
	server.mapLock.Lock()
	server.OnlineMap[user.Name] = user
	server.mapLock.Unlock()

	//广播当前用户上线信息
	server.BroadCast(user, "已上线")

	//当前handler阻塞
	select {}
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
