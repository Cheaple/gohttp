package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"mymodule/utils"
)

var (
	debug			bool = false
	port			int = 8080
	mediaType		= [...]string{"text/html", "text/plain", "image/gif", "image/jpeg", "image/jpeg", "text/css"}
	fileType		= [...]string{"html", "txt", "gif", "jpeg", "jpg", "css"}
	workspace string
	maxWorkers		= 10
	workerSem		= make(chan struct{}, maxWorkers)
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
		log.Printf("Error listening on port %s: %s", address, err.Error())
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
	defer func() { <-workerSem }()
	log.Println("Handling a new connection from ", conn.RemoteAddr().String())

	request, err := utils.ParseRequest(conn)
	if err != nil {
		fmt.Println("Error parsing connection:", err)
		return
	}

	responseWriter := utils.NewConnResponseWriter(conn)

	// Check the request method
	if request.Method == "GET" {
		log.Println("Handing a GET request")
		handleGET(responseWriter, request)
	} else if request.Method == "POST" {
		log.Println("Handing a POST request")
		handlePOST(responseWriter, request)
	} else {
		responseWriter.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// Handle GET request
func handleGET(w *utils.ConnResponseWriter, r *http.Request) {
	
}

// Handle POST request
func handlePOST(w *utils.ConnResponseWriter, r *http.Request) {
	// Check if the request method is POST
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// Check if the content type is "multipart/form-data"
	contentType := r.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "multipart/form-data") {
		w.WriteHeader(http.StatusUnsupportedMediaType)
		w.WriteText(fmt.Sprintf("Unsupported Content-Type: %s", contentType))
		return
	}

	// Parse multipart data
	err := r.ParseMultipartForm(16 << 20)
	if err != nil {
		log.Println("Error parsing form:", err)
		return
	}
	for _, h := range r.MultipartForm.File["file"] {
		// Read next file of multipart data
		file, _ := h.Open()
		if err != nil {
			log.Println("Error retriving file in MultipartForm:", err)
			return
		}
		defer file.Close()
		log.Println("File Uploaded: %+v", h.Header)

		// Create a target file locally 
		dst, err := os.Create(workspace + "/" + h.Filename)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.WriteText("Error creating destination file")
			return
		}
		defer dst.Close()
	
		// Store uploaded data
		_, err = io.Copy(dst, file)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.WriteText("Error copying file")
			return
		}
		log.Printf("Received %s file content type: %s", h.Header.Get("Content-Type"), h.Filename)
	}
	
	w.WriteHeader(http.StatusOK)
	w.WriteText("Files uploaded successfully")
}

func typeMatch(typeList [6]string, target string) (bool, string) {
	for i, str := range typeList {
		if str == target {
			return true, mediaType[i]
		}
	}
	return false, ""
}


