package main

import (
	"bufio"
	"fmt"
	"log/slog"
	"net"
	"os"
)

func main() {
	udpAddr, err := net.ResolveUDPAddr("udp", "localhost:42069")
	if err != nil {
		slog.Error("Failed to resolve UDP address", "error", err)
	}

	conn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		slog.Error("Failed to dial UDP", "error", err)
	}
	defer conn.Close()

	reader := bufio.NewReader(os.Stdin)

	slog.Info("UDP sender started. Type messages to send to UDP server.")

	for {
		fmt.Print("> ")
		text, err := reader.ReadString('\n')
		if err != nil {
			slog.Error("Failed to read from stdin", "error", err)
		}

		_, err = conn.Write([]byte(text))
		if err != nil {
			slog.Error("Failed to write to UDP", "error", err)
		}
	}
}
