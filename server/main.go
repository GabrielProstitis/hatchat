package main

import (
	"hatchat/server/server"
)

func main() {
	server := new(server.Server)
	server.Run(":8080", 2)
}
