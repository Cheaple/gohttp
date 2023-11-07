package main

import (
	"fmt"
	"flag"
	"bufio"
	"io"
	"net"
	"net/http"
	// "os"
	// "strconv"
)

func HttpClientHandler(client net.Conn) {
	defer client.Close()

	// Parse the request
	request, err := http.ReadRequest(
		bufio.NewReader(client))
	if err != nil {
		fmt.Println("Error parsing request:", err)
		return
	}

	// Check the request method
	if request.Method != "GET" {
		// Return a "501 Not Implemented" response for non-GET requests
		fmt.Println("Error Not Implemented: method", request.Method)
		client.Write([]byte("HTTP/1.1 501 Not Implemented\r\n"))
		client.Write([]byte("Content-Length: 0\r\n\r\n"))
		return
	}

	// Forward the GET request to the remote server
	serverURL := request.URL
	targetConn, err := net.Dial("tcp", serverURL.Host)
	if err != nil {
		fmt.Println("Error connecting to target server:", err)
		return
	}
	defer targetConn.Close()
	forwardData(client, targetConn)

	// Copy the remote server's response back to the client
	client.Write([]byte("HTTP/1.1 200 OK\r\n"))
	client.Write([]byte("Content-Length: "))
	io.Copy(client, targetConn)
}

func forwardData(src, dest net.Conn) {
	buffer := make([]byte, 4096)
	for {
		n, err := src.Read(buffer)
		if err != nil {
			fmt.Println("Error forwarding request to remote server:", err)
			return
		}
		_, err = dest.Write(buffer[:n])
		if err != nil {
			return
		}
	}
}
func main() {
	var (
		port = flag.Int("port", 8080, "Port to listen on")
	)
	flag.Parse()

	// Create a proxy server listening on, listen on the port specified from the command line
	address := fmt.Sprintf(":%d", *port)
	ln, err := net.Listen("tcp", address)
	if err != nil {
		fmt.Println("Error starting proxy server:", err)
		return
	}
	defer ln.Close()

	fmt.Printf("Proxy server is listening on %d\n", *port)

	// Accept and handle incoming client connections
	for {
		client, err := ln.Accept()
		if err != nil {
			fmt.Println("Error accepting client connection:", err)
			continue
		}
		go HttpClientHandler(client)
	}
}
