package main

import "golean-im-system/com/rango/im"

func main() {

	server := im.NewServer("127.0.0.1", 8888)
	server.Start()

}
