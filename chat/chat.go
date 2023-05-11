package chat

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"os"

	libp2p "github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"

	multiaddr "github.com/multiformats/go-multiaddr"
)

func HandleStream(s network.Stream) {
	log.Println("Send your message: ")

	rw := bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))

	go ReadData(rw)
	go WriteData(rw)
}
func ReadData(rw *bufio.ReadWriter){
	for {
		str, err := rw.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading from buffer")
			return
		}
		if str == "exit\n" {
			fmt.Println("Exiting...")
			return
		}
		fmt.Printf("\x1b[32m %s \x1b[0m>", str)
	}
}
func WriteData(rw *bufio.ReadWriter ){
	stdReader := bufio.NewReader(os.Stdin)
	for  {
		fmt.Print(">")
		sendData, err := stdReader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading from stdin")
			return
		}
		rw.WriteString(fmt.Sprintf("%s\n", sendData))
		rw.Flush()
	}
}
func MakeHost(port int, randomness io.Reader) (host.Host, error) {
	prvKey, _, err := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, randomness)
	if err != nil {
		return nil, err
	}

	sourceMultiAddr, _ := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", port))

	return libp2p.New(
		libp2p.ListenAddrs(sourceMultiAddr),
		libp2p.Identity(prvKey),
	)
}
func StartPeer(ctx context.Context, h host.Host, streamHandler network.StreamHandler){
	h.SetStreamHandler("/chat/1.0.0", streamHandler)

	var port string
	for _, la := range h.Network().ListenAddresses() {
		if p, err := la.ValueForProtocol(multiaddr.P_TCP); err == nil {
			port = p
			break
		}
	}
	if port == "" {
		panic("was not able to find actual local port")
	}
	log.Printf("Run './chat -d /ipv4/127.0.0.1/tcp/%v/p2p/%s' on another console.\n", port, h.ID().Pretty())
	log.Printf("\n[*] Your Multiaddress Is: /ip4/127.0.0.1")
}
func StartPeerAndConnect(ctx context.Context, h host.Host, destination string)(*bufio.ReadWriter, error){
	log.Println("These are your muktiaddress: ")
	
	for _, la := range h.Addrs() {
		log.Printf(" - %v\n", la)
	}
	log.Println()

	destinationAddr, err := multiaddr.NewMultiaddr(destination)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	targetInfo, err := peer.AddrInfoFromP2pAddr(destinationAddr)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	h.Peerstore().AddAddrs(targetInfo.ID, targetInfo.Addrs, peerstore.PermanentAddrTTL)

	s, err := h.NewStream(ctx, targetInfo.ID, "/chat/1.0.0")
	if err != nil {
		log.Println(err)
		return nil, err
	}
	log.Println("Connected to ", targetInfo.ID)

	rw := bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))

	return rw, nil
}