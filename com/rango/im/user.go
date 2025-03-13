package im

import (
	"fmt"
	"net"
	"strings"
)

type User struct {
	Name string
	Addr string
	C    chan string
	conn net.Conn

	server *Server
}

//创建一个用户的API

func NewUser(conn net.Conn, server *Server) *User {
	userAddr := conn.RemoteAddr().String()
	user := &User{
		Name:   userAddr,
		Addr:   userAddr,
		C:      make(chan string),
		conn:   conn,
		server: server,
	}
	//启动监听当前user channel消息的goroutine
	go user.ListenMessage()
	return user
}

// Online 用户上线
func (user *User) Online() {
	user.server.mapLock.Lock()
	user.server.OnlineMap[user.Name] = user
	user.server.mapLock.Unlock()

	user.server.BroadCast(user, "已上线")
}

// Offline 用户下线
func (user *User) Offline() {
	user.server.mapLock.Lock()
	delete(user.server.OnlineMap, user.Name)
	user.server.mapLock.Unlock()

	user.server.BroadCast(user, "已下线")

}

func (user *User) SendMessage(message string) {
	user.conn.Write([]byte(message + "\n"))
}

// DoMessage 处理消息
func (user *User) DoMessage(msg string) {
	if msg == "who" {
		//查询当前在线用户都有哪些
		user.server.mapLock.Lock()
		for _, ou := range user.server.OnlineMap {
			ouMsg := fmt.Sprintf("[%s] %s : 在线...", ou.Addr, ou.Name)
			user.SendMessage(ouMsg)
		}
		user.server.mapLock.Unlock()
	} else if len(msg) > 7 && msg[:7] == "rename|" {
		//消息格式: rename|张三
		newName := strings.Split(msg, "|")[1]

		//判断name是否存在
		_, ok := user.server.OnlineMap[newName]
		if ok {
			user.SendMessage("当前用户名已被使用，请重新输入\n")
		} else {
			user.server.mapLock.Lock()

			delete(user.server.OnlineMap, user.Name)
			user.server.OnlineMap[newName] = user

			user.server.mapLock.Unlock()

			user.Name = newName
			user.SendMessage("您已更新用户名" + newName + "\n")
		}
	} else {
		user.server.BroadCast(user, msg)
	}
}

// ListenMessage 监听当前User channel的方法，一旦有消息，就直接发送给对端客户端
func (user *User) ListenMessage() {
	for {
		msg, ok := <-user.C
		if !ok {
			break
		}
		_, err := user.conn.Write([]byte(msg + "\n"))
		if err != nil {
			fmt.Printf("write to addr[%v] err, user name: %v", user.Addr, user.Name)
			continue
		}
	}
}
