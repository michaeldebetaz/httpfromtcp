package main

import (
	"fmt"
	"log"
	"log/slog"
	"net"

	"httpfromtcp/internal/request"
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

		req, err := request.RequestFromReader(conn)
		if err != nil {
			slog.Error("Error reading request", "error", err, "remoteAddr", conn.RemoteAddr().String())
			conn.Close()
			continue
		}

		fmt.Println("Request line:")
		fmt.Printf("- Method: %s\n", req.RequestLine.Method)
		fmt.Printf("- Target: %s\n", req.RequestLine.RequestTarget)
		fmt.Printf("- Version: %s\n", req.RequestLine.HttpVersion)
		fmt.Println("Headers:")
		for k, v := range req.Headers {
			fmt.Printf("- %s: %s\n", k, v)
		}
		fmt.Println("Body:")
		fmt.Printf("%s", string(req.Body))

		slog.Info("Connection closed", "remoteAddr", conn.RemoteAddr().String())
	}
}
