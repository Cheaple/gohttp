package main

import (
	"fmt"
	"bufio"
	"io"
	"net"
	"net/http"
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
		client.Write([]byte("HTTP/1.1 501 Not Implemented\r\n"))
		client.Write([]byte("Content-Length: 0\r\n\r\n"))
		return
	}

	// Forward the GET request to the remote server
	serverURL := request.URL
	response, err := http.Get(serverURL.String())
	if err != nil {
		fmt.Println("Error forwarding request to remote server:", err)
		return
	}
	defer response.Body.Close()

	// Copy the remote server's response back to the client
	client.Write([]byte("HTTP/1.1 200 OK\r\n"))
	client.Write([]byte("Content-Length: "))
	client.Write([]byte(fmt.Sprintf("%d\r\n\r\n", response.ContentLength)))
	io.Copy(client, response.Body)
}

func main() {
	// Create a proxy server listening on :8080
	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Println("Error starting proxy server:", err)
		return
	}
	defer ln.Close()

	fmt.Println("Proxy server is listening on :8080")

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