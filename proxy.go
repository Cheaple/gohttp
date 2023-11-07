package main

import (
	"fmt"
	"flag"
	"bufio"
	"net"
	"net/http"
	"log"
)

func HttpClientHandler(clientConn net.Conn) {
	defer clientConn.Close()

	// Parse the request
	request, err := http.ReadRequest(
		bufio.NewReader(clientConn))
	if err != nil {
		log.Println("Error parsing request:", err)
		return
	}

	// Check the request method
	if request.Method != "GET" {
		// Return a "501 Not Implemented" response for non-GET requests
		log.Println("Error Not Implemented: method", request.Method)
		notImplement(clientConn)
		return
	}

	// Forward the GET request to the remote server
	serverURL := request.URL
	targetConn, err := net.Dial("tcp", serverURL.Host)
	if err != nil {
		log.Println("Error connecting to target server:", err)
		return
	}
	defer targetConn.Close()
	log.Println("Forwarding a request to target server: ", serverURL)
	forwardData(clientConn, targetConn)

	// Copy the remote server's response back to the client
	log.Println("Forwarding a response from target server: ", serverURL)
	forwardData(targetConn, clientConn)
}

func forwardData(src, dest net.Conn) {
	buffer := make([]byte, 4096)
	for {
		n, err := src.Read(buffer)
		if err != nil {
			log.Println("Error reading from source:", err)
			return
		}
		_, err = dest.Write(buffer[:n])
		if err != nil {
			log.Println("Error forwarding to target:", err)
			return
		}
	}
}

func notImplement(conn net.Conn) {
	var buf string
	buf = "HTTP/1.1 501 Method Not Implemented\r\n"
	_, _ = conn.Write([]byte(buf))
	buf = "Content-Type: text/html; charset=utf-8\r\n"
	_, _ = conn.Write([]byte(buf))
	buf = "Connection: close\r\n"
	_, _ = conn.Write([]byte(buf))
	buf = "\r\n"
	_, _ = conn.Write([]byte(buf))
	buf = "<HTML><HEAD><TITLE>Method Not Implemented\r\n"
	_, _ = conn.Write([]byte(buf))
	buf = "</TITLE></HEAD>\r\n"
	_, _ = conn.Write([]byte(buf))
	buf = "<BODY><P>HTTP request method not supported.\r\n"
	_, _ = conn.Write([]byte(buf))
	buf = "</BODY></HTML>\r\n"
	_, _ = conn.Write([]byte(buf))
}

func main() {
	var port int
	flag.IntVar(&port, "port", 8080, "Port to listen on")
	flag.IntVar(&port, "p", 8080, "Port to listen on")
	flag.Parse()

	// Create a proxy server listening on, listen on the port specified from the command line
	address := fmt.Sprintf(":%d", port)
	ln, err := net.Listen("tcp", address)
	if err != nil {
		log.Println("Error starting proxy server:", err)
		return
	}
	defer ln.Close()

	log.Printf("Proxy server is listening on %d\n", port)

	// Accept and handle incoming client connections
	for {
		client, err := ln.Accept()
		if err != nil {
			log.Println("Error accepting client connection:", err)
			continue
		}
		log.Println("Receiving a new request")
		go HttpClientHandler(client)
	}
}
