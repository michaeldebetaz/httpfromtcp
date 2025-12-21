package headers

import (
	"bytes"
	"fmt"
	"slices"
	"strings"
	"unicode"
)

type Headers map[string]string

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	const CRLF = "\r\n"

	idx := bytes.Index(data, []byte(CRLF))

	if idx < 0 {
		return 0, false, nil
	}

	if idx == 0 {
		return 2, true, nil
	}

	read := idx + len(CRLF)

	line := data[:idx]

	sepIdx := bytes.Index(line, []byte(":"))

	if sepIdx < 0 {
		err := fmt.Errorf("No semicolon found")
		return 0, false, err
	}

	if sepIdx == 0 {
		err := fmt.Errorf("line starts with semicolon")
		return 0, false, err
	}

	if unicode.IsSpace(rune(line[sepIdx-1])) {
		err := fmt.Errorf("field name contains trailing whitespace")
		return 0, false, err
	}

	fieldName := bytes.TrimSpace(line[:sepIdx])

	if len(fieldName) < 1 {
		err := fmt.Errorf("field name is empty")
		return 0, false, err
	}

	for _, b := range fieldName {
		if !isValidFieldNameChar(rune(b)) {
			err := fmt.Errorf("field name contains invalid character '%s'", string(b))
			return 0, false, err
		}
	}

	key := strings.ToLower(string(fieldName))

	if sepIdx > len(line)-1 {
		err := fmt.Errorf("No data after semicolon")
		return 0, false, err
	}

	fieldValue := string(bytes.TrimSpace(line[sepIdx+1:]))

	if value, ok := h[key]; !ok {
		h[key] = fieldValue
	} else {
		h[key] = value + ", " + fieldValue
	}

	return read, false, nil
}

func NewHeaders() Headers {
	headers := make(Headers)
	return headers
}

func isValidFieldNameChar(r rune) bool {
	validChars := []rune{'!', '#', '$', '%', '&', '\'', '*', '+', '-', '.', '^', '_', '`', '|', '~'}
	if unicode.IsUpper(r) || unicode.IsLower(r) || unicode.IsDigit(r) || slices.Contains(validChars, r) {
		return true
	}

	return false
}
