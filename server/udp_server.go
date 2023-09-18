package udp_server

import (
	"encoding/json"
	"fmt"
	"github.com/Eznopot/udp"
	"log"
	"net"
	"sync"
)

var instance *net.PacketConn
var addrs []*net.Addr
var isServerClose = false
var logger func(string)
var once sync.Once

// The function `serverPacketSystemHandler` handles different types of data received from clients and
// performs corresponding actions.
//
// Args:
//
//	data (string): The `data` parameter is a string that represents the type of packet received from
//
// the client. It can have two possible values: "handshake" or "close".
//
//	addr: The `addr` parameter is a pointer to a `net.Addr` object. It represents the network address
//
// of the client that sent the packet.
//
// Returns:
//
//	an integer value.
func serverPacketSystemHandler(data string, addr *net.Addr) int {
	switch data := data; data {
	case "handshake":
		if addrs == nil {
			var tmp []*net.Addr
			tmp = append(tmp, addr)
			addrs = tmp
			return 0
		}
		addrs = (append(addrs, addr))
		return 0
	case "close":
		for i, tmpAddr := range addrs {
			if (*addr).String() == (*tmpAddr).String() {
				addrs = append(addrs[:i], addrs[i+1:]...)
				break
			}
		}
		return 0
	default:
		return 0
	}
}

// The SetLogger function sets the logger function to be used for logging messages.
//
// Args:
//
//	loggerFunc: The loggerFunc parameter is a function that takes a string as input and does not
//
// return anything.
func SetLogger(loggerFunc func(string)) {
	logger = loggerFunc
}

// The function `listener` reads packets from a socket and calls a handler function to process the
// packets, while also handling system packets separately.
//
// Args:
//
//	wg: The parameter `wg` is of type `*sync.WaitGroup`. It is used to synchronize the goroutines and
//
// wait for them to finish before exiting the function.
//
//	handler: The `handler` parameter is a function that takes three arguments:
func listener(wg *sync.WaitGroup, handler func(net.PacketConn, *net.Addr, udp.Packet)) {
	defer wg.Done()
	for !isServerClose {
		packet, addr, err := readFromSocket()
		if isServerClose {
			break
		} else if err != nil {
			println("error on socket", err.Error())
			continue
		}
		if logger != nil {
			res, err := json.Marshal(packet)
			if err != nil {
				return
			}
			logger(string(res))
		}
		if packet.Type == "system" {
			if serverPacketSystemHandler(packet.Data, addr) == 1 {
				return
			}
		}
		handler(*instance, addr, packet)
	}
}

// The function `readFromSocket` reads data from a socket, unmarshals it into a UDP packet, and returns
// the packet, the address it was received from, and any error that occurred.
func readFromSocket() (udp.Packet, *net.Addr, error) {
	var packet udp.Packet
	buf := make([]byte, 2048)
	len, addr, err := (*instance).ReadFrom(buf)
	if isServerClose {
		return packet, nil, nil
	} else if err != nil {
		fmt.Println("One reading buff", err.Error())
		return packet, nil, err
	}
	buf = buf[:len]
	json.Unmarshal(buf, &packet)
	return packet, &addr, nil
}

// The function `CreateServer` creates a UDP server on the specified port and waits for the first
// connection before starting a listener goroutine.
//
// Args:
//
//	port (string): The `port` parameter is a string that represents the port number on which the
//
// server will listen for incoming UDP packets.
//
//	handler: The handler parameter is a function that takes three arguments:
func CreateServer(port string, handler func(net.PacketConn, *net.Addr, udp.Packet)) *sync.WaitGroup {
	var wg sync.WaitGroup
	once.Do(func() {
		udpServer, err := net.ListenPacket("udp", ":"+port)
		if err != nil {
			log.Fatal(err)
		}
		instance = &udpServer
		wg.Add(1)
		go listener(&wg, handler)
	})
	return &wg
}

// The function sends a packet with a specified type and data to all clients connected to a UDP server.
//
// Args:
//
//	str (string): The "str" parameter is a string that represents the data you want to send to all
//
// clients. It could be any information or message that you want to transmit.
//
//	packetType (string): The `packetType` parameter is a string that represents the type of packet
//
// being sent. It could be used to differentiate between different types of messages or data being sent
// to the clients.
func SendToAllClient(str, packetType string) {
	if instance == nil {
		log.Fatal("instance of UDP server is null")
		return
	}
	toSend := udp.Packet{
		Type: packetType,
		Data: str,
	}

	res, err := json.Marshal(toSend)
	if err != nil {
		log.Fatal("error on json:", err.Error())
	}
	for _, addr := range addrs {
		(*instance).WriteTo(res, *addr)
	}
}

// The function sends a UDP packet to all clients except for the client specified by the given address.
//
// Args:
//
//	str (string): The `str` parameter is a string that represents the data to be sent to the clients.
//
// It could be any information or message that you want to send.
//
//	packetType (string): The `packetType` parameter is a string that represents the type of the packet
//
// being sent. It could be any value that is meaningful in the context of your application, such as
// "message", "data", "request", etc.
//
//	addr: The `addr` parameter is a pointer to a `net.Addr` object. It represents the address of a
//
// client that should be excluded from the list of clients to which the packet should be sent.
func SendToAllExcludingItselfClient(str, packetType string, addr *net.Addr) {
	if instance == nil {
		log.Fatal("instance of UDP server is null")
		return
	}
	toSend := udp.Packet{
		Type: packetType,
		Data: str,
	}

	res, err := json.Marshal(toSend)
	if err != nil {
		log.Fatal("error on json:", err.Error())
	}
	for _, elemAddr := range addrs {
		if addr != elemAddr {
			(*instance).WriteTo(res, *addr)
		}
	}
}

// The function sends a packet to a client using a UDP server instance.
//
// Args:
//
//	str (string): The `str` parameter is a string that represents the data to be sent to the client.
//
// It could be any information or message that you want to send.
//
//	packetType (string): The `packetType` parameter is a string that represents the type of the packet
//
// being sent to the client. It could be any string value that you define to categorize the packet.
//
//	index (int): The "index" parameter is an integer that represents the index of the address in the
//
// "addrs" slice. It is used to determine the address to which the packet should be sent.
func SendToClient(str, packetType string, index int) {
	if instance == nil {
		log.Fatal("instance of UDP server is null")
		return
	}
	toSend := udp.Packet{
		Type: packetType,
		Data: str,
	}
	res, err := json.Marshal(toSend)
	if err != nil {
		log.Fatal("error on json:", err.Error())
	}
	(*instance).WriteTo(res, *(addrs[index]))
}

// The function GetAllClientInfo returns a list of strings containing the addresses of all clients.
//
// Returns:
//
//	a list of strings, which contains the string representation of each address in the `addrs` slice.
func GetAllClientInfo() []string {
	var list []string
	for _, addr := range addrs {
		list = append(list, (*addr).String())
	}
	return list
}

// The CloseServer function closes the server and sends a "close" message to all clients.
func CloseServer() {
	SendToAllClient("close", "system")
	isServerClose = true
	defer (*instance).Close()
}
