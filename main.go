package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

func main() {
	f, err := os.Open("messages.txt")
	if err != nil {
		log.Fatalf("Error opening file: %v", err)
	}

	for line := range getLinesChannel(f) {
		fmt.Printf("read: %s\n", line)
	}
}

func getLinesChannel(f io.ReadCloser) <-chan string {
	ch := make(chan string)
	buf := make([]byte, 8)
	currentLine := ""
	go func() {
		for {
			n, err := f.Read(buf)
			if err != nil {
				if errors.Is(err, io.EOF) {
					if len(currentLine) > 0 {
						ch <- currentLine
					}
					f.Close()
					close(ch)
					break
				} else {
					log.Fatalf("Error reading file: %v", err)
				}
			}

			parts := strings.Split(string(buf[:n]), "\n")

			if len(parts) == 1 {
				currentLine += parts[0]
			} else if len(parts) == 2 {
				currentLine += parts[0]
				ch <- currentLine
				currentLine = parts[1]
			} else {
				log.Fatal("Error: two or more newlines in the buffer")
			}
		}
	}()

	return ch
}
