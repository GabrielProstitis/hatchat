package client

import (
	"encoding/binary"
	"hatchat/server/connection"
	"hatchat/server/consts"
	"log"
	"sync"
)

// ClientHandler struct
type ClientHandler struct {
	// public
	Connections   []*connection.Connection
	AcceptChannel chan *connection.Connection

	DispatcherChannelSend chan *connection.Packet // packets to send to the dispatcher
	DispatcherChannelRecv chan *connection.Packet // packets to receive from the dispatcher

	// private
	connectionsMu sync.Mutex
}

//
// ClientHandler methods
//

func (ch *ClientHandler) init(dispatcherChannelSend chan *connection.Packet) {
	ch.AcceptChannel = make(chan *connection.Connection, 100)

	ch.DispatcherChannelSend = dispatcherChannelSend
	ch.DispatcherChannelRecv = make(chan *connection.Packet, consts.PacketsChannelSize)
}

// Returns the idx of the connection in the ClinetHandler.Connection slice
func (ch *ClientHandler) FindConnection(targetId int) int {
	ch.connectionsMu.Lock()
	defer ch.connectionsMu.Unlock()

	for idx := 0; idx < len(ch.Connections); idx++ {
		if ch.Connections[idx].Id == targetId {
			return idx
		}
	}

	return -1
}

func (ch *ClientHandler) Run(dispatcherChannelSend chan *connection.Packet) {
	// handle logins
	// handle registrarions
	// check for new messages in different connections and send the message to the right handler and connection
	ch.init(dispatcherChannelSend)

	var chWG sync.WaitGroup

	chWG.Add(3)

	go func() {
		defer chWG.Done()
		ch.receiveNewConnections()
	}()

	go func() {
		defer chWG.Done()
		ch.receivePackets()
	}()

	go func() {
		defer chWG.Done()
		ch.sendPackets()
	}()

	chWG.Wait()
}

func (ch *ClientHandler) sendNewConnectionPacket(conn *connection.Connection) {
	// Send the personal connection id to the new client
	packet := new(connection.Packet)
	packet.Type = connection.PacketNewConnection
	packet.ReceiverId = conn.Id
	packet.Contents = make([]byte, 8)
	binary.LittleEndian.PutUint64(packet.Contents, uint64(conn.Id))
	ch.DispatcherChannelSend <- packet
}

func (ch *ClientHandler) receiveNewConnections() {
	for conn := range ch.AcceptChannel {
		ch.connectionsMu.Lock()
		ch.Connections = append(ch.Connections, conn)
		ch.sendNewConnectionPacket(conn)
		ch.connectionsMu.Unlock()
	}
}

func (ch *ClientHandler) removeConnection(idx int) {
	log.Printf("removing connection %d", ch.Connections[idx].Id)
	ch.connectionsMu.Lock()
	ch.Connections = append(ch.Connections[:idx], ch.Connections[idx+1:]...)
	ch.connectionsMu.Unlock()
}

// The function receives incoming packets and sends them to the dispatcher
func (ch *ClientHandler) receivePackets() {
	for {
		for idx := 0; idx < len(ch.Connections); idx++ {
			ch.connectionsMu.Lock()
			packet, isAlive := ch.Connections[idx].RecvPacket()
			ch.connectionsMu.Unlock()

			if packet == nil {
				if !isAlive {
					// handle EOF by deleting the connection
					ch.removeConnection(idx)
				}
			} else {
				if packet.Type == connection.PacketIn {
					ch.DispatcherChannelSend <- packet
				}
			}

		}
	}
}

// The function send the packets received from the dispatcher to the destination connection

func (ch *ClientHandler) sendPackets() {
	for packet := range ch.DispatcherChannelRecv {
		log.Printf("packet receiver id: %d, packet contents size: %d", packet.ReceiverId, len(packet.Contents))

		ch.connectionsMu.Lock()
		if err := ch.Connections[packet.TargetConnectionIdx].SendPacket(packet); err != nil {
			if err.Error() == "Cannot send PacketIn packets" || err.Error() == "Cannot send PacketBroadcast packets" {
				log.Println(err)
			} else {
				ch.removeConnection(packet.TargetConnectionIdx)
			}
		}
		ch.connectionsMu.Unlock()
	}
}

