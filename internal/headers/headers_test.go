package headers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidSingleHeader(t *testing.T) {
	headers := NewHeaders()
	require.NotNil(t, headers)
	data := []byte("Host: localhost:42069\r\n\r\n")
	n, done, err := headers.Parse(data)
	require.NoError(t, err)
	assert.Equal(t, "localhost:42069", headers["host"])
	assert.Equal(t, 23, n)
	assert.False(t, done)
}

func TestValidMultipleHeaders(t *testing.T) {
	headers := NewHeaders()
	require.NotNil(t, headers)
	n, done, err := headers.Parse([]byte("Set-Person: lane-loves-go\r\n\r\n"))
	require.NoError(t, err)
	assert.Equal(t, 27, n)
	assert.False(t, done)
	n, done, err = headers.Parse([]byte("Set-Person: prime-loves-zig\r\n\r\n"))
	require.NoError(t, err)
	assert.Equal(t, 29, n)
	n, done, err = headers.Parse([]byte("Set-Person: tj-loves-ocaml\r\n\r\n"))
	require.NoError(t, err)
	assert.Equal(t, 28, n)
	assert.Equal(t, "lane-loves-go, prime-loves-zig, tj-loves-ocaml", headers["set-person"])
}

func TestInvalidSpacingHeader(t *testing.T) {
	headers := NewHeaders()
	data := []byte("       Host : localhost:42069       \r\n\r\n")
	n, done, err := headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)
}

func TestUpperCaseHeaderKey(t *testing.T) {
	headers := NewHeaders()
	require.NotNil(t, headers)
	data := []byte("HOST: localhost:42069\r\n\r\n")
	n, done, err := headers.Parse(data)
	require.NoError(t, err)
	assert.Equal(t, "localhost:42069", headers["host"])
	assert.Equal(t, 23, n)
	assert.False(t, done)
}

func TestInvalidCharacterInHeaderKey(t *testing.T) {
	headers := NewHeaders()
	data := []byte("H©st: localhost:42069\r\n\r\n")
	n, done, err := headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)
}
