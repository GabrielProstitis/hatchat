package connection

import (
	"log"
	"net"
	"time"
)

const (
	ReadTimeout   time.Duration = 50 * time.Millisecond
	WriteTimeout  time.Duration = 50 * time.Millisecond
	RawBufferSize int           = 1024
)

// Connection struct
type Connection struct {
	net.Conn
	Id int
}

// non blocking
// returns a packet, and a bool (if true, connection is alive, else, received an EOF)
func (c *Connection) RecvPacket() (*Packet, bool) {
	c.SetReadDeadline(time.Now().Add(ReadTimeout))
	return c.BlockingRecvPacket()
}

func (c *Connection) BlockingRecvPacket() (*Packet, bool) {
	var bRead int
	var errRead error

	rawDataBuffer := make([]byte, RawBufferSize)

	if bRead, errRead = c.Read(rawDataBuffer); errRead != nil {
		if net_err, ok := errRead.(net.Error); ok && net_err.Timeout() {
			return nil, true
		} else if errRead.Error() == "EOF" {
			return nil, false
		} else {
			log.Fatalf("error: %s", errRead)
			return nil, false
		}
	}

	// resize the buffer
	rawDataBuffer = rawDataBuffer[:bRead]

	packet := new(Packet)
	packet.SenderId = c.Id
	packet.Id = 0 // TO-DO id generator

	if err := packet.parseRawData(rawDataBuffer); err != nil {
		log.Printf("packet parsing error: %s", err)
		// the connection is alive but the packet sent is not parsable
		return nil, true
	}

	return packet, true
}

func (c *Connection) SendPacket(packet *Packet) error {
	c.SetWriteDeadline(time.Now().Add(WriteTimeout))

	raw, err := packet.convertToRawData()

	if err != nil {
		return err
	}

	if _, err := c.Write(raw); err != nil {
		return err
	}

	return nil
}
