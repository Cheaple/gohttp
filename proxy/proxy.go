package main

import (
	"bufio"
	"fmt"
	"flag"
	"io"
	"log"
	"net"
	"net/http"
	"mymodule/utils"
	// "time"
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
		log.Println("Receiving a new request from", client.RemoteAddr().String())
		go httpClientHandler(client)
	}
}

//
// Handler for incoming request from HTTP client
//
func httpClientHandler(clientConn net.Conn) {
	defer clientConn.Close()
	requestSem <- struct{}{}  // wait for other coroutines
	defer func() { <-requestSem }()
	log.Println("Handling a new connection from ", clientConn.RemoteAddr().String())

	// Parse the request
	request, err := http.ReadRequest(bufio.NewReader(clientConn))
	if err != nil {
		log.Println("Error parsing request:", err)
		return
	}

	responseWriter := utils.NewConnResponseWriter(clientConn)

	// TODO: check validity of the request
	// Check the request method
	if request.Method == http.MethodGet {
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
		log.Println("Forwarding a HTTP GET response from target server:", serverURL.Host)
		forwardGetRsp(responseWriter, response)

	} else {
		// Return a "501 Not Implemented" response for non-GET requests
		log.Printf("Handing a %s request, not implemented!", request.Method)
		responseWriter.WriteHeader(http.StatusNotImplemented)
		responseWriter.WriteText("Proxy Method Not Implemented")
		return
	}

	// to test max concurrency
	// clientConn.Close()
	// time.Sleep(10 * time.Second)
	// <-requestSem
}

//
// Forward GET response from server to client
//
func forwardGetRsp(w *utils.ConnResponseWriter, r *http.Response) {
	defer r.Body.Close()

	// Set headers for the new response
	for key, values := range r.Header {
		for _, value := range values {
			w.Header().Set(key, value)
		}
	}

	// Copy the response body to the connection's output stream
	w.WriteHeader(r.StatusCode)
	_, err := io.Copy(w, r.Body)
	if err != nil {
		fmt.Println("Error copying response body to connection:", err)
	}
}
