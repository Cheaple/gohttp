package main

import (
	"bufio"
	"fmt"
	"flag"
	"io"
	"log"
	"net"
	"net/http"
	"time"
)

var maxRequests = 10  // max number of concurrent HTTP requests
var requestSem = make(chan struct{}, maxRequests)

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
		log.Println("Receiving a new request ...")
		go httpClientHandler(client)
	}
}

func httpClientHandler(clientConn net.Conn) {
	defer clientConn.Close()
	requestSem <- struct{}{}  // wait for other coroutines
	defer func() { <-requestSem }()
	log.Println("Handling a new request")

	// Parse the request
	request, err := http.ReadRequest(
		bufio.NewReader(clientConn))
	if err != nil {
		log.Println("Error parsing request:", err)
		return
	}

	// TODO: check validity of the request
	// Check the request method
	if request.Method == "GET" {
		// Forward the GET request to the remote server
		serverURL := request.URL
		log.Println("Forwarding a HTTP GET request to target server: ", serverURL.Host)
		response, err := http.Get(serverURL.String())
		if err != nil {
			log.Println("Error forwarding request to target server:", err)
			return
		}
		defer response.Body.Close()

		// Copy the remote server's response back to the client
		log.Println("Forwarding a HTTP GET response from target server: ", serverURL.Host)
		forwardGetRsp(response , clientConn)

	} else {
		// Return a "501 Not Implemented" response for non-GET requests
		log.Println("Error forwarding: not implemented method", request.Method)
		returnNotImplement(clientConn)
		return
	}
	time.Sleep(5 * time.Second)  // to test max concurrency
}

func forwardGetRsp(r *http.Response, conn net.Conn) {
	defer r.Body.Close()

	// Set the status code and headers for the new response
	_, _ = conn.Write([]byte(fmt.Sprintf("HTTP/1.1 %d %s\r\n", r.StatusCode, http.StatusText(r.StatusCode))))
	for key, values := range r.Header {
		for _, value := range values {
			_, _ = conn.Write([]byte(fmt.Sprintf("%s: %s\r\n", key, value)))
		}
	}
	_, _ = conn.Write([]byte("\r\n"))

	// Copy the response body to the connection's output stream
	_, err := io.Copy(conn, r.Body)
	if err != nil {
		fmt.Println("Error copying response body to connection:", err)
	}
}

func returnNotImplement(conn net.Conn) {
	var buf string
	buf = "HTTP/1.1 501 Method Not Implemented\r\n"
	_, _ = conn.Write([]byte(buf))
	buf = "Content-Type: text/html; charset=utf-8\r\n"
	_, _ = conn.Write([]byte(buf))
	buf = "Connection: close\r\n"
	_, _ = conn.Write([]byte(buf))
	buf = "\r\n"
	_, _ = conn.Write([]byte(buf))
	buf = "<HTML><HEAD><TITLE>\r\nMethod Not Implemented\r\n"
	_, _ = conn.Write([]byte(buf))
	buf = "</TITLE></HEAD>\r\n"
	_, _ = conn.Write([]byte(buf))
	buf = "<BODY><P>\r\nHTTP request method not supported.\r\n"
	_, _ = conn.Write([]byte(buf))
	buf = "</BODY></HTML>\r\n"
	_, _ = conn.Write([]byte(buf))
}
