package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/network"
	peerstore "github.com/libp2p/go-libp2p/core/peer"
)

func handleStream(stream network.Stream) {
	defer stream.Close()

	reader := bufio.NewReader(stream)
	message, err := reader.ReadString('\n')
	if err != nil {
		if err != io.EOF {
			fmt.Println("Error reading message:", err)
		}
		return
	}
	message = strings.TrimSuffix(message, "\n")
	fmt.Println("Received message:", message)
}

const myProtocol = "/myprotocol/1.0.0"

func main() {
	log.Println("Server started")

	node, err := libp2p.New(
		libp2p.ListenAddrStrings("/ip4/0.0.0.0/tcp/0"),
		libp2p.Ping(false),
	)
	if err != nil {
		panic(err)
	}

	// Set a stream handler for your custom protocol
	node.SetStreamHandler(myProtocol, handleStream)

	// print the node's PeerInfo in multiaddr format
	peerInfo := peerstore.AddrInfo{
		ID:    node.ID(),
		Addrs: node.Addrs(),
	}
	addrs, err := peerstore.AddrInfoToP2pAddrs(&peerInfo)
	if err != nil {
		panic(err)
	}
	fmt.Println("libp2p node address:", addrs[0])

	// wait for a SIGINT or SIGTERM signal
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch
	fmt.Println("Received signal, shutting down...")

	// shut the node down
	if err := node.Close(); err != nil {
		panic(err)
	}
}
