package client

import (
	"encoding/binary"
	"fmt"
	"hatchat/server/connection"
	"log"
	"net"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

type Client struct {
	id      int
	conn    connection.Connection
	id2tab  map[int]*fyne.Container
	tabs    *container.AppTabs
	app     fyne.App
	window  fyne.Window
	idLabel *widget.Label
}

func (c *Client) init() {
	log.SetPrefix("client: ")

	// initialize graphical items
	c.app = app.New()
	c.window = c.app.NewWindow("hatchat")
	c.idLabel = widget.NewLabel("Waiting for your ID...")
	c.tabs = container.NewAppTabs(
		container.NewTabItem("Home", c.idLabel),
	)
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
				c.idLabel.SetText(fmt.Sprintf("ID: %d", c.id))
			case connection.PacketOut:
				log.Printf("%d> %s\n", packet.SenderId, packet.Contents)

				if _, ok := c.id2tab[packet.SenderId]; !ok {
					c.id2tab[c.id] = container.New(layout.NewGridLayout(1))
					c.tabs.Append(container.NewTabItem(fmt.Sprintf("%d", c.id), c.id2tab[c.id]))
				} else {
				}

			default:
				log.Println("recieved malformed packet")
			}
		}
	}
}

func (c *Client) sendMessages() {
	// to implement

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

	// set contents
	c.tabs.SetTabLocation(container.TabLocationLeading)
	c.window.SetContent(c.tabs)
	c.window.ShowAndRun()

	clientwg.Wait()
}
