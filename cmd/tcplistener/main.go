package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net"
	"strings"
)

func main() {
	slog.Info("Starting TCP server on localhost:42069...")

	listener, err := net.Listen("tcp", "localhost:42069")
	if err != nil {
		fmt.Println("Error starting server:", err)
		log.Fatalf("Error starting server: %v", err)
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			log.Fatalf("Error accepting connection: %v", err)
		}

		slog.Info("Accepted connection", "remoteAddr", conn.RemoteAddr().String())

		for line := range getLinesChannel(conn) {
			fmt.Println(line)
		}

		slog.Info("Connection closed", "remoteAddr", conn.RemoteAddr().String())
	}
}

func getLinesChannel(rc io.ReadCloser) <-chan string {
	ch := make(chan string)
	buf := make([]byte, 4096)
	currentLine := ""

	go func() {
		for {
			n, err := rc.Read(buf)
			if err != nil {
				if errors.Is(err, io.EOF) {
					if len(currentLine) > 0 {
						ch <- currentLine
					}

					rc.Close()
					close(ch)

					break
				} else {
					log.Fatalf("Error reading file: %v", err)
				}
			}

			parts := strings.Split(string(buf[:n]), "\r\n")

			if len(parts) == 1 {
				currentLine += parts[0]
			}

			if len(parts) > 1 {
				lastIndex := len(parts) - 1

				for _, part := range parts[:lastIndex] {
					ch <- currentLine + part
				}

				currentLine = parts[lastIndex]
			}
		}
	}()

	return ch
}
