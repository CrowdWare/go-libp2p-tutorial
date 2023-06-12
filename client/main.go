package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/network"
	peerstore "github.com/libp2p/go-libp2p/core/peer"
	multiaddr "github.com/multiformats/go-multiaddr"
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
	log.Println("Client started")
	if len(os.Args) < 2 {
		fmt.Println("Usage: ./client <server address>")
		return
	}

	// start a libp2p node that listens on a random local TCP port,
	// but without running the built-in ping protocol
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

	addr, err := multiaddr.NewMultiaddr(os.Args[1])
	if err != nil {
		panic(err)
	}
	peer, err := peerstore.AddrInfoFromP2pAddr(addr)
	if err != nil {
		panic(err)
	}
	if err := node.Connect(context.Background(), *peer); err != nil {
		panic(err)
	}

	// Send a string message to the remote peer
	stream, err := node.NewStream(context.Background(), peer.ID, myProtocol)
	if err != nil {
		panic(err)
	}
	writer := bufio.NewWriter(stream)
	_, err = writer.WriteString("Hello from sender!\n")
	if err != nil {
		panic(err)
	}
	err = writer.Flush()
	if err != nil {
		panic(err)
	}
	stream.Close()

	// shut the node down
	if err := node.Close(); err != nil {
		panic(err)
	}
}
