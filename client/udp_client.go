package udp_client

import (
	"encoding/json"
	"go_mutateur/src/udp"
	"log"
	"net"
	"sync"
)

var conn *net.UDPConn
var isConnClose = false
var once sync.Once

// The function `CreateConnection` creates a new UDP connection and sends a handshake message.
//
// Args:
//   address (string): The "address" parameter is the IP address or hostname of the server you want to
// connect to. The "port" parameter is the port number on which the server is listening for UDP
// connections.
//   port (string): The "port" parameter is the port number on which the UDP connection will be
// established. It is a string representing the port number.
func CreateConnection(address, port string) *net.UDPConn {
	once.Do(func() {
		udpServer, err := net.ResolveUDPAddr("udp", address+":"+port)

		if err != nil {
			log.Fatal("ResolveUDPAddr failed:", err.Error())
		}

		connection, err := net.DialUDP("udp", nil, udpServer)
		if err != nil {
			log.Fatal("Listen failed:", err.Error())
		}
		conn = connection
		//send first message for handshake
		handshake := udp.Packet{
			Data: "handshake",
			Type: "system",
		}
		bytes, err := json.Marshal(handshake)
		if err != nil {
			log.Fatal("serialize failed:", err.Error())
		}
		_, err = conn.Write(bytes)
		if err != nil {
			log.Fatal("Write data failed:", err.Error())
		}
	})
	return conn
}

// The function `clientPacketSystemHandler` handles different types of client packets and returns a
// value indicating whether the connection should be closed or not.
//
// Args:
//   data (string): The parameter "data" is a string that represents the packet received from the
// client.
//
// Returns:
//   The function `clientPacketSystemHandler` returns an integer value.
func clientPacketSystemHandler(data string) int {
	switch data := data; data {
	case "close":
		CloseConnection()
		return 0
	default:
		return 0
	}
}

// The function `readFromConn` reads data from a connection, unmarshals it into a UDP packet, and
// returns the packet along with any errors encountered.
func readFromConn() (udp.Packet, error) {
	var packet udp.Packet
	received := make([]byte, 2048)
	len, err := conn.Read(received)
	if isConnClose {
		return packet, nil
	} else if err != nil {
		println("Read data failed:", err.Error())
		return packet, err
	}
	received = received[:len]
	err = json.Unmarshal(received, &packet)
	if err != nil {
		log.Fatal("error on json:", err.Error())
		return packet, err
	}
	return packet, nil
}

// The Receive function continuously reads packets from a connection and passes them to a handler
// function until the connection is closed.
//
// Args:
//   wg: The "wg" parameter is a pointer to a sync.WaitGroup. It is used to synchronize the completion
// of multiple goroutines.
//   handler: The "handler" parameter is a function that takes a single argument of type "udp.Packet".
// This function will be called with each received packet that is not of type "system".
func Receive(wg *sync.WaitGroup, handler func(udp.Packet)) {
	for !isConnClose {
		packet, err := readFromConn()
		if isConnClose {
			break
		} else if err != nil {
			continue
		}
		if packet.Type == "system" {
			if clientPacketSystemHandler(packet.Data) == 1 {
				return
			}
		}
		handler(packet)
	}
	if wg != nil {
		wg.Done()
	}
}

// The CloseConnection function sends a close packet over UDP and closes the connection.
func CloseConnection() {
	packet := udp.Packet{
		Data: "close",
		Type: "system",
	}
	bytes, err := json.Marshal(packet)
	if err != nil {
		log.Fatal("error on json:", err.Error())
	}
	conn.Write(bytes)
	isConnClose = true
	defer conn.Close()
}
