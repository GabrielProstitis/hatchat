package client

import (
	"encoding/binary"
	"hatchat/server/connection"
	"log"
	"net"
	"sync"

	"hatchat/client/client/gui"
)

type Client struct {
	id             int
	conn           connection.Connection
	messageChannel chan *gui.Message
	gui            *gui.Gui
}

func (c *Client) init() {
	log.SetPrefix("client: ")

	c.messageChannel = make(chan *gui.Message, 1)

	c.gui = new(gui.Gui)
	c.gui.Init("hatchat", c.messageChannel)
}

func (c *Client) connect(address string) error {
	conn, err := net.Dial("tcp", address)

	if err != nil {
		return err
	}

	c.conn.Conn = conn
	c.conn.Id = 0

	return nil
}

// this function recieves, decodes and store the incoming packets
func (c *Client) recvMessages() {
	var packet *connection.Packet = nil
	var isAlive bool = true

	for isAlive {
		packet, isAlive = c.conn.BlockingRecvPacket()

		if packet != nil && isAlive {
			switch packet.Type {
			case connection.PacketNewConnection:
				c.id = int(binary.LittleEndian.Uint64(packet.Contents))
				c.gui.UpdateId(c.id)
			case connection.PacketOut:
				message := new(gui.Message)
				message.Init(packet.SenderId, packet.Contents, false)
				c.gui.AddMessage(message)
			default:
				log.Println("recieved malformed packet")
			}
		}
	}
}

func (c *Client) sendMessages() {
	for {
		msg := <-c.messageChannel
		packet := new(connection.Packet)

		packet.Contents = msg.Contents
		packet.ReceiverId = msg.ClientId
		packet.SenderId = c.id
		packet.Type = connection.PacketIn

		c.conn.SendPacket(packet)

	}
}

func (c *Client) Run(address string) {
	c.init()

	if err := c.connect(address); err != nil {
		panic(err)
	}

	log.Println("connected")

	var clientwg sync.WaitGroup

	clientwg.Add(1)

	// create a go routine for incoming packets
	go func() {
		defer clientwg.Done()
		c.recvMessages()
	}()

	clientwg.Add(1)
	go func() {
		defer clientwg.Done()
		c.sendMessages()
	}()

	log.Println("Setting contents")

	c.gui.Run()

	clientwg.Wait()
}
