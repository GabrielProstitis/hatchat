package accepter

import (
	"hatchat/server/client"
	"hatchat/server/connection"
	"log"
	"math"
	"net"
)

type Accepter struct {
	// private
	listener net.Listener

	// when called it will generate a new valid connection id
	connectiondIdGenerator func() int

	// reference to the list of client handlers
	clientHandlers []client.ClientHandler
}

// empty for now
func (acc *Accepter) init(clientHandlers []client.ClientHandler) {
	acc.connectiondIdGenerator = acc.createConnectionIdGenerator()
	acc.clientHandlers = clientHandlers
}

// accepts new connections
func (acc *Accepter) Run(address string, clientHandlers []client.ClientHandler) {
	acc.init(clientHandlers)

	acc.listen(address)

	for {
		conn, handlerId := acc.getConnection()
		log.Printf("new connection (id: %d) @ handler[%d]", conn.Id, handlerId)
		acc.clientHandlers[handlerId].AcceptChannel <- conn
	}
}

// start listening on the specified address
func (acc *Accepter) listen(address string) {
	ls, err := net.Listen("tcp", address)

	if err != nil {
		panic(err)
	}

	acc.listener = ls
}

// the function will accept new connections and assign the connection to a client handler
func (acc *Accepter) getConnection() (*connection.Connection, int) {
	// accept new connection
	c, err := acc.listener.Accept()

	if err != nil {
		panic(err)
	}

	conn := new(connection.Connection)
	// assign connection to less busy handler
	conn.Conn = c
	conn.Id = acc.connectiondIdGenerator()
	handlerId := acc.getAvailableHandler()

	return conn, handlerId
}

// client id generator
func (acc Accepter) createConnectionIdGenerator() func() int {
	id := 0

	return func() int {
		id++
		return id
	}
}

// client handler id finder
func (acc *Accepter) getAvailableHandler() int {
	var r, ml int

	ml = math.MaxInt
	for idx := 0; idx < len(acc.clientHandlers); idx++ {
		if l := len(acc.clientHandlers[idx].Connections); l < ml {
			r = idx
			ml = l
		}
	}

	return r
}
