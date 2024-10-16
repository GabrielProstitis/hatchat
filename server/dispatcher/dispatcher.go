package dispatcher

import (
	"hatchat/server/client"
	"hatchat/server/connection"
	"log"
)

type Dispatcher struct {
	channelRecv    chan *connection.Packet
	clientHandlers []client.ClientHandler
}

func (d *Dispatcher) init(clientHandler []client.ClientHandler, channelRecv chan *connection.Packet) {
	d.channelRecv = channelRecv
	d.clientHandlers = clientHandler
}

func (d *Dispatcher) Run(clientHandlers []client.ClientHandler, channelRecv chan *connection.Packet) {
	d.init(clientHandlers, channelRecv)

	for packet := range d.channelRecv {
		found := d.findHandler(packet)

		if !found {
			log.Printf("the connection with id %d was not found!", packet.ReceiverId)
		} else {
			switch packet.Type {
			case connection.PacketIn:
				// the input packets become output packets
				packetOut := packet
				packet.Id = 0 // TO-DO id generator
				packetOut.Type = connection.PacketOut
				d.clientHandlers[packet.TargetHandlerIdx].DispatcherChannelRecv <- packetOut
			case connection.PacketNewConnection:
				d.clientHandlers[packet.TargetHandlerIdx].DispatcherChannelRecv <- packet
			default:
				log.Printf("Trying to send invalid packet (packet id: %d)", packet.Id)
			}
		}
	}
}

// resolves the packet if a connection with id == packet.ReceiverId is found
func (d *Dispatcher) findHandler(packet *connection.Packet) bool {
	found := false
	for i := 0; i < len(d.clientHandlers) && !found; i++ {
		if idx := d.clientHandlers[i].FindConnection(packet.ReceiverId); idx != -1 {
			found = true

			// insert all the information received in the packet
			packet.TargetConnectionIdx = idx
			packet.TargetHandlerIdx = i
		}
	}
	return found
}
