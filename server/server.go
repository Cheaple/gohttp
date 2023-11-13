package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"../utils"
)

var (
	debug			bool = false
	port			int = 8080
	mediaType		= [...]string{"text/html", "text/plain", "image/gif", "image/jpeg", "image/jpeg", "text/css"}
	fileType		= [...]string{"html", "txt", "gif", "jpeg", "jpg", "css"}
	workspace string
	maxWorkers		= 10
	var workerSem	= make(chan struct{}, maxWorkers)
)

func main() {
	_, src, _, ok := runtime.Caller(0)
	if !ok {
		log.Println("Cannot access path of workspace")
		return
	}
	workspace = filepath.Dir(src) + "/data"

	flag.IntVar(&port, "port", 8080, "Port to listen on")
	flag.IntVar(&port, "p", 8080, "Port to listen on")
	flag.Parse()
	
	// Create a proxy server listening on, listen on the port specified from the command line
	address := fmt.Sprintf(":%d", port)
	ln, err := net.Listen("tcp", address)
	if err != nil {
		log.Printf("Fatal error listening on port %s: %s", strport, err.Error())
		os.Exit(1)
	}
	defer ln.Close()

	log.Printf("Proxy server is listening on %d\n", port)

	for {
		conn, err := ln.Accept()
		if err != nil {
			continue
		}
		log.Println("Receiving a new request from", conn.RemoteAddr().String())
		go handleRequest(conn)
	}
}


//
// Handler for incoming request from HTTP connection
//
func handleRequest(conn net.Conn) {
	defer conn.Close()
	workerSem <- struct{}{}
	defer func() { <-seworkerSemm }()
	log.Println("Handling a new connection from ", clientConn.RemoteAddr().String())

	request, err := parseRequest(conn)
	if err != nil {
		fmt.Println("Error parsing connection:", err)
		return
	}

	// Check the request method
	if request.Method == "GET" {
		handle(request)
	} else if request.Method == "POST" {
		returnBadRequest(conn)
	} else {
		// Return a "501 Not Implemented" response for non-GET requests
		log.Println("Error forwarding: not implemented method", request.Method)
		returnNotImplement(clientConn)
		return
	}
}

func typeMatch(typeList [6]string, target string) (bool, string) {
	for i, str := range typeList {
		if str == target {
			return true, mediaType[i]
		}
	}
	return false, ""
}





func handlePOST(w http.ResponseWriter, r *http.Request) {
	// Check if the request method is POST
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	// Check if the content type is "image"
	contentType := r.Header.Get("Content-Type")
	if contentType != "image/jpeg" {
		// http.Error(w, "Unsupported media type", http.StatusUnsupportedMediaType)
		return
	}

	file, err := os.Create("received_image.jpg")
	if err != nil {
		// http.Error(w, "Error creating file", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	// Read the request body and write it to the local file
	_, err = io.Copy(file, r.Body)
	if err != nil {
		// http.Error(w, "Error reading request body", http.StatusInternalServerError)
		return
	}

	log.Println("Received image with content type:", contentType)
	log.Println("Image size:", len(file), "bytes")

	// You can handle the image data here, such as saving it to a file or processing it.

	// // Send a response
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Image received successfully"))
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
	buf = "<HTML><HEAD><TITLE>Method Not Implemented\r\n"
	_, _ = conn.Write([]byte(buf))
	buf = "</TITLE></HEAD>\r\n"
	_, _ = conn.Write([]byte(buf))
	buf = "<BODY><P>HTTP request method not supported.\r\n"
	_, _ = conn.Write([]byte(buf))
	buf = "</BODY></HTML>\r\n"
	_, _ = conn.Write([]byte(buf))
}

func returnNotFound(conn net.Conn) {
	var buf string

	buf = "HTTP/1.1 404 Not Found\r\n"
	_, _ = conn.Write([]byte(buf))
	buf = "Content-Type: text/html; charset=utf-8\r\n"
	_, _ = conn.Write([]byte(buf))
	buf = "Connection: close\r\n"
	_, _ = conn.Write([]byte(buf))
	buf = "\r\n"
	_, _ = conn.Write([]byte(buf))
	buf = "<HTML><HEAD><TITLE>404 Not Found\r\n"
	_, _ = conn.Write([]byte(buf))
	buf = "</TITLE></HEAD>\r\n"
	_, _ = conn.Write([]byte(buf))
	buf = "<BODY><P>404 not Found.\r\n"
	_, _ = conn.Write([]byte(buf))
	buf = "</BODY></HTML>\r\n"
	_, _ = conn.Write([]byte(buf))
}

func returnBadRequest(conn net.Conn) {
	var buf string

	buf = "HTTP/1.1 400 Bad Request\r\n"
	_, _ = conn.Write([]byte(buf))
	buf = "Content-Type: text/html; charset=utf-8\r\n"
	_, _ = conn.Write([]byte(buf))
	buf = "Connection: close\r\n"
	_, _ = conn.Write([]byte(buf))
	buf = "\r\n"
	_, _ = conn.Write([]byte(buf))
	buf = "<HTML><HEAD><TITLE>400 Bad Request\r\n"
	_, _ = conn.Write([]byte(buf))
	buf = "</TITLE></HEAD>\r\n"
	_, _ = conn.Write([]byte(buf))
	buf = "<BODY><P>400 Bad Request.\r\n"
	_, _ = conn.Write([]byte(buf))
	buf = "</BODY></HTML>\r\n"
	_, _ = conn.Write([]byte(buf))
}


