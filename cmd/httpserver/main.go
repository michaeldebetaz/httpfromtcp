package main

import (
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"

	"httpfromtcp/internal/request"
	"httpfromtcp/internal/response"
	"httpfromtcp/internal/server"
)

const port = 42069

func main() {
	server, err := server.Serve(port, func(w io.Writer, req *request.Request) *server.HandlerError {
		switch req.RequestLine.RequestTarget {
		case "/yourproblem":
			return handlerError(response.BadRequest, "Your problem is not my problem\n")
		case "/myproblem":
			return handlerError(response.InternalServerError, "Woopsie, my bad\n")
		}

		w.Write([]byte("All good, frfr\n"))

		return nil
	})
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer server.Close()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}

func handlerError(statusCode response.StatusCode, msg string) *server.HandlerError {
	return &server.HandlerError{StatusCode: statusCode, Message: msg}
}
