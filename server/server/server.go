package server

import (
	"hatchat/server/accepter"
	"hatchat/server/client"
	"hatchat/server/connection"
	"hatchat/server/consts"
	"hatchat/server/dispatcher"
	"log"
	"sync"
)

//
// Server struct
//

type Server struct {
	//  the handler id is equals to the index of the handler in this slice
	clientHandlers []client.ClientHandler
	dispatcher     *dispatcher.Dispatcher
	accepter       *accepter.Accepter

	// this is the common channel between the dispatcher and the client handlers
	commonDC chan *connection.Packet
}

//
// Server methods
//

func (s *Server) Run(address string, client_handlers_sz int) {
	s.init(client_handlers_sz)

	// start accepting and handling new connections
	var serverWG sync.WaitGroup

	log.Printf("Spawning accepter")
	// ACCEPTER
	serverWG.Add(1)
	go func() {
		defer serverWG.Done()
		s.accepter.Run(address, s.clientHandlers)
	}()

	log.Printf("Spawning dispatcher")
	// DISPATCHER
	serverWG.Add(1)
	go func() {
		defer serverWG.Done()
		s.dispatcher.Run(s.clientHandlers, s.commonDC)
	}()

	log.Printf("Spawning client handlers")
	// CLIENT HANDLERS
	for i := range len(s.clientHandlers) {
		serverWG.Add(1)
		go func() {
			defer serverWG.Done()
			s.clientHandlers[i].Run(s.commonDC)
		}()
	}

	serverWG.Wait()
}

func (s *Server) init(client_handlers_sz int) {
	// logging
	log.SetPrefix("server: ")

	// creating class members
	s.accepter = new(accepter.Accepter)
	s.dispatcher = new(dispatcher.Dispatcher)
	s.clientHandlers = make([]client.ClientHandler, client_handlers_sz)

	// create the common channel between the
	s.commonDC = make(chan *connection.Packet, consts.PacketsChannelSize)

}
