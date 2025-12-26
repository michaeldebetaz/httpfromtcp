package request

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strconv"
	"strings"
	"unicode"

	"httpfromtcp/internal/headers"
)

type requestState int

const (
	initialized requestState = iota
	parsingHeaders
	parsingBody
	done
)

const bufferSize int = 8

type Request struct {
	state       requestState
	RequestLine *RequestLine
	Headers     headers.Headers
	Body        []byte
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

func (r *Request) parse(data []byte) (int, error) {
	bytesParsed := 0

	for r.state != done {
		n, err := r.parseSingle(data[bytesParsed:])
		if err != nil {
			return 0, err
		}

		bytesParsed += n
		if n == 0 {
			break
		}
	}

	return bytesParsed, nil
}

func (r *Request) parseSingle(data []byte) (int, error) {
	switch r.state {
	case initialized:
		requestLine, parsed, err := parseRequestLine(data)
		if err != nil {
			err := fmt.Errorf("failed to parse request line: %w", err)
			return parsed, err
		}

		if parsed == 0 {
			return 0, nil
		}

		r.RequestLine = requestLine
		r.state = parsingHeaders

		return parsed, nil

	case parsingHeaders:
		parsed, done, err := r.Headers.Parse(data)
		if err != nil {
			return 0, err
		}

		if done {
			r.state = parsingBody
		}

		return parsed, nil

	case parsingBody:
		contentLengthStr, ok := r.Headers.Get("Content-length")
		if !ok {
			r.state = done
			return len(data), nil
		}

		contentLength, err := strconv.Atoi(contentLengthStr)
		if err != nil {
			return 0, fmt.Errorf("error converting '%s' to integer", contentLengthStr)
		}

		r.Body = append(r.Body, data...)
		bodyLength := len(r.Body)

		if len(data) == 0 && bodyLength < contentLength {
			err := fmt.Errorf("body length (%d) is smaller than headers content-length (%d)", bodyLength, contentLength)
			return 0, err
		}

		if bodyLength > contentLength {
			err := fmt.Errorf("body length (%d) is greater than headers content-length (%d)", bodyLength, contentLength)
			return 0, err
		}

		if bodyLength == contentLength {
			r.state = done
		}

		return len(data), nil
	}

	return 0, fmt.Errorf("unknown parser state")
}

func RequestFromReader(r io.Reader) (*Request, error) {
	buf := make([]byte, bufferSize)
	readToIndex := 0

	request := &Request{
		state:   initialized,
		Headers: headers.NewHeaders(),
		Body:    make([]byte, 0),
	}

	for request.state != done {
		slog.Info("Request", "state", request.state)

		if readToIndex >= len(buf) {
			newBuf := make([]byte, len(buf)*2)
			copy(newBuf, buf[:readToIndex])
			buf = newBuf
		}

		bytesRead, err := r.Read(buf[readToIndex:])
		if err != nil {
			if errors.Is(err, io.EOF) {
				if request.state != done {
					return nil, fmt.Errorf(
						"incomplete request, in state: %d, read n bytes on EOF: %d", request.state, bytesRead,
					)
				}
				break
			}
			return nil, err
		}
		readToIndex += bytesRead

		bytesParsed, err := request.parse(buf[:readToIndex])
		if err != nil {
			err := fmt.Errorf("failed to parse request: %w", err)
			return nil, err
		}

		copy(buf, buf[bytesParsed:])
		readToIndex -= bytesParsed
	}

	return request, nil
}

func parseRequestLine(data []byte) (*RequestLine, int, error) {
	idx := bytes.Index(data, []byte("\r\n"))
	if idx < 0 {
		return &RequestLine{}, 0, nil
	}

	read := idx + 2

	line := string(data[:idx])

	parts := strings.Split(line, " ")
	if len(parts) != 3 {
		err := fmt.Errorf("request line must contain exactly 3 parts")
		return &RequestLine{}, read, err
	}

	method := parts[0]
	if len(method) == 0 {
		err := fmt.Errorf("method must not be empty")
		return &RequestLine{}, read, err
	}

	for _, c := range method {
		if c < 'A' || 'Z' < c {
			err := fmt.Errorf("method must contain only capital letters")
			return &RequestLine{}, read, err
		}
	}

	httpVersion := strings.TrimPrefix(parts[2], "HTTP/")
	if httpVersion != "1.1" {
		err := fmt.Errorf("unsupported HTTP version: %s", httpVersion)
		return &RequestLine{}, read, err
	}

	requestTarget := parts[1]
	if len(requestTarget) == 0 {
		err := fmt.Errorf("request target must not be empty")
		return &RequestLine{}, read, err
	}

	for _, c := range requestTarget {
		if unicode.IsSpace(c) {
			err := fmt.Errorf("request target must not contain whitespace")
			return &RequestLine{}, read, err
		}
	}

	return &RequestLine{
		Method:        method,
		RequestTarget: requestTarget,
		HttpVersion:   httpVersion,
	}, read, nil
}
