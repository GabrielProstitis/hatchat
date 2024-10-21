package main

import (
	"fmt"
	"hatchat/client/client"
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

	client := new(client.Client)
	client.Run(fmt.Sprintf("%s:%d", address, port))
}
