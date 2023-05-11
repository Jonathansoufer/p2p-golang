package main

import (
	"context"
	"crypto/rand"
	"flag"
	"fmt"
	"io"
	mrand "math/rand"
	"os"

	chat "github.com/Jonathansoufer/p2p-golang/chat"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sourcePort := flag.Int("sp", 0, "Source port number")
	dest := flag.String("d", "", "Destination multiaddr string")
	help := flag.Bool("h", false, "Display help")
	debug := flag.Bool("debug", false, "Debug your execution")
	flag.Parse()

	if *help {
		fmt.Println("This program demonstrates a simple p2p chat application using libp2p")
		fmt.Println()
		fmt.Println("Usage: Run './p2p-golang -sp <SOURCE_PORT>' where <SOURCE_PORT> can be any port number.")
		fmt.Println("Now run './p2p-golang -d <MULTIADDR>' where <MULTIADDR> is multiaddress of previous listener host.")
		fmt.Println()
		os.Exit(0)
	}

	var r io.Reader
	if *debug {
		r = mrand.New(mrand.NewSource(int64(*sourcePort)))
		fmt.Println("Debugging...")

	} else {
		r = rand.Reader
	}

	h, err := chat.MakeHost(*sourcePort, r)
	if err != nil {
		panic(err)
	}

	if *dest == "" {
		chat.StartPeer(ctx, h, chat.HandleStream)
	} else {
		rw, err := chat.StartPeerAndConnect(ctx, h, *dest)
		if err != nil {
			panic(err)
		}
		go chat.WriteData(rw)
		go chat.ReadData(rw)
	}

}