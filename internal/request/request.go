package request

import (
	"bytes"
	"fmt"
	"io"
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
	RequestLine RequestLine
	Headers     headers.Headers
	Body        []byte
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

func (r *Request) parse(data []byte) (int, error) {
	switch r.state {
	case initialized:
		requestLine, read, err := parseRequestLine(data)
		if err != nil {
			err := fmt.Errorf("failed to parse request line: %w", err)
			return read, err
		}

		if read > 0 {
			r.RequestLine = requestLine
			r.state = parsingHeaders
		}

		return read, nil

	case parsingHeaders:
		totalBytesParsed := 0

		for r.state != parsingBody {
			read, err := r.parseSingle(data[totalBytesParsed:])
			if err != nil {
				err := fmt.Errorf("failed to parse headers: %w", err)
				return read, err
			}

			if read == 0 {
				break
			}

			totalBytesParsed += read
		}

		return totalBytesParsed, nil

	case parsingBody:
		read := len(data)

		contentLengthStr := r.Headers.Get("Content-length")
		if contentLengthStr == "" {
			r.state = done
			return read, nil
		}

		contentLength, err := strconv.Atoi(contentLengthStr)
		if err != nil {
			err := fmt.Errorf("error converting '%s' to integer", contentLengthStr)
			return read, err
		}

		r.Body = append(r.Body, data...)
		bodyLength := len(r.Body)

		if read == 0 && bodyLength < contentLength {
			err := fmt.Errorf("body length (%d) is smaller than headers content-length (%d)", bodyLength, contentLength)
			return read, err
		}

		if bodyLength > contentLength {
			err := fmt.Errorf("body length (%d) is greater than headers content-length (%d)", bodyLength, contentLength)
			return read, err
		}

		if bodyLength == contentLength {
			r.state = done
		}

		return read, nil
	}

	return 0, fmt.Errorf("unknown parser state")
}

func (r *Request) parseSingle(data []byte) (int, error) {
	read, d, err := r.Headers.Parse(data)
	if err != nil {
		err := fmt.Errorf("failed to parse heaeder: %w", err)
		return 0, err
	}

	if d {
		r.state = parsingBody
		return read, nil
	}

	return read, nil
}

func RequestFromReader(r io.Reader) (*Request, error) {
	buf := make([]byte, bufferSize)
	readToIndex := 0

	request := &Request{
		state:   initialized,
		Headers: headers.NewHeaders(),
	}

	for request.state != done {
		if readToIndex == len(buf) {
			buf2 := make([]byte, len(buf)*2)
			copy(buf2, buf[:readToIndex])
			buf = buf2
		}

		n, err := r.Read(buf[readToIndex:])
		if err != nil {
			if err == io.EOF {
				_, err := request.parse(buf[:readToIndex])
				if err != nil {
					err := fmt.Errorf("failed to parse request: %w", err)
					return nil, err
				}
				request.state = done
				break
			} else {
				err := fmt.Errorf("failed to read from reader: %w", err)
				return nil, err
			}
		}

		readToIndex += n

		read, err := request.parse(buf[:readToIndex])
		if err != nil {
			err := fmt.Errorf("failed to parse request: %w", err)
			return nil, err
		}
		if read == 0 {
			continue
		}

		if read > 0 {
			copy(buf, buf[read:readToIndex])
			readToIndex -= read
		}

	}

	return request, nil
}

func parseRequestLine(data []byte) (RequestLine, int, error) {
	idx := bytes.Index(data, []byte("\r\n"))
	if idx < 0 {
		return RequestLine{}, 0, nil
	}

	read := idx + 2

	line := string(data[:idx])

	parts := strings.Split(line, " ")
	if len(parts) != 3 {
		err := fmt.Errorf("request line must contain exactly 3 parts")
		return RequestLine{}, read, err
	}

	method := parts[0]
	if len(method) == 0 {
		err := fmt.Errorf("method must not be empty")
		return RequestLine{}, read, err
	}

	for _, c := range method {
		if c < 'A' || 'Z' < c {
			err := fmt.Errorf("method must contain only capital letters")
			return RequestLine{}, read, err
		}
	}

	httpVersion := strings.TrimPrefix(parts[2], "HTTP/")
	if httpVersion != "1.1" {
		err := fmt.Errorf("unsupported HTTP version: %s", httpVersion)
		return RequestLine{}, read, err
	}

	requestTarget := parts[1]
	if len(requestTarget) == 0 {
		err := fmt.Errorf("request target must not be empty")
		return RequestLine{}, read, err
	}

	for _, c := range requestTarget {
		if unicode.IsSpace(c) {
			err := fmt.Errorf("request target must not contain whitespace")
			return RequestLine{}, read, err
		}
	}

	return RequestLine{
		Method:        method,
		RequestTarget: requestTarget,
		HttpVersion:   httpVersion,
	}, read, nil
}
