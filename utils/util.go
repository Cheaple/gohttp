package utils

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"log"
)

// ConnResponseWriter implements the http.ResponseWriter interface for net.Conn.
type ConnResponseWriter struct {
	conn   net.Conn
	header http.Header
	status int
}


// NewConnResponseWriter creates a new ConnResponseWriter.
func NewConnResponseWriter(conn net.Conn) *ConnResponseWriter {
	return &ConnResponseWriter{
		conn:   conn,
		header: make(http.Header),
		status: http.StatusOK,
	}
}


// Header returns the header map.
func (w *ConnResponseWriter) Header() http.Header {
	return w.header
}


// Write writes the data to the connection.
func (w *ConnResponseWriter) Write(data []byte) (int, error) {
	if w.status == 0 {
		w.WriteHeader(http.StatusOK)
	}
	return w.conn.Write(data)
}

func (w *ConnResponseWriter) WriteText(data string) (int, error) {
	if w.status == 0 {
		w.WriteHeader(http.StatusOK)
	}
	return w.conn.Write([]byte(data))
}

// WriteHeader writes the status code to the connection.
func (w *ConnResponseWriter) WriteHeader(statusCode int) {
	if w.status == 0 {
		w.status = statusCode
		statusText := http.StatusText(statusCode)
		fmt.Fprintf(w.conn, "HTTP/1.1 %d %s\r\n", statusCode, statusText)
		w.header.Write(w.conn)
		fmt.Fprintf(w.conn, "\r\n")
	}
}


// parse HTTP connection into a http.Request object
func ParseRequest(conn net.Conn) (*http.Request, error) {
	reader := bufio.NewReader(conn)
	log.Println("00")
	// Read the request line
	requestLine, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}
	log.Println("111")
	parts := strings.Fields(requestLine)
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid request line: %s", requestLine)
	}
	reqURL, err := url.Parse(parts[1])
	if err != nil {
		return nil, fmt.Errorf("error parsing URL: %v", err)
	}
	log.Println("222")
	req := &http.Request{
		Method: parts[0],
		URL: reqURL,
		Proto: "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
	}
	log.Println("333")

	// Read request header
	req.Header = make(http.Header)
	for {
		line, err := reader.ReadString('\n')
		if err != nil || line == "\r\n" {
			break
		}

		headerParts := strings.SplitN(line, ":", 2)
		if len(headerParts) == 2 {
			req.Header.Add(strings.TrimSpace(headerParts[0]), strings.TrimSpace(headerParts[1]))
		}
	}

	// Read request body
	contentLengthStr := req.Header.Get("Content-Length")
	if contentLengthStr != "" {
		var contentLength int
		fmt.Sscanf(contentLengthStr, "%d", &contentLength)

		// Read the request body
		bodyBytes := make([]byte, contentLength)
		_, err := io.ReadFull(reader, bodyBytes)

		if err != nil {
			return nil, fmt.Errorf("error reading request body: %v", err)
		}
		req.Body = io.NopCloser(bytes.NewReader(bodyBytes))
	}
	log.Println("999")
	return req, nil
}