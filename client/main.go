package main

import (
	"hatchat/client/client"
)

func main() {
	client := new(client.Client)
	client.Run(":8080")
}
