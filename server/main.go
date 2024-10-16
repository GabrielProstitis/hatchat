package main

import (
	"fmt"
	"hatchat/server/server"
	"os"
	"strconv"
)

func main() {
	if len(os.Args) != 3 {
		panic("bad args")
	}

	address := os.Args[1]
	port, err := strconv.Atoi(os.Args[2])

	if err != nil {
		panic("bad port")
	}

	server := new(server.Server)
	server.Run(fmt.Sprintf("%s:%d", address, port), 2)
}
