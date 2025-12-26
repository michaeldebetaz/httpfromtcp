package response

import (
	"fmt"
	"io"
	"strconv"

	"httpfromtcp/internal/headers"
)

type StatusCode int

const (
	OK                  StatusCode = 200
	BadRequest          StatusCode = 400
	InternalServerError StatusCode = 500
)

var statusReasonPhrase = map[StatusCode]string{
	200: "OK",
	400: "Bad Request",
	500: "Internal Server Error",
}

func WriteStatusLine(w io.Writer, statusCode StatusCode) error {
	statusLine := fmt.Sprintf("HTTP/1.1 %d %s \r\n", statusCode, statusReasonPhrase[statusCode])

	if _, err := w.Write([]byte(statusLine)); err != nil {
		return fmt.Errorf("error writing status line: %w", err)
	}

	return nil
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	h := headers.NewHeaders()
	h["content-length"] = strconv.Itoa(contentLen)
	h["connection"] = "close"
	h["content-type"] = "text/plain"

	return h
}

func WriteHeaders(w io.Writer, headers headers.Headers) error {
	for k, v := range headers {
		if _, err := w.Write([]byte(k + ": " + v + "\r\n")); err != nil {
			return fmt.Errorf("error writing headers: %w", err)
		}
	}

	if _, err := w.Write([]byte("\r\n")); err != nil {
		return fmt.Errorf("error writing headers: %w", err)
	}

	return nil
}
