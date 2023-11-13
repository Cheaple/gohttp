package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"mymodule/utils"
)

var (
	debug			bool = false
	port			int = 8080
	mediaTypeList	= [...]string{"text/html", "text/plain", "image/gif", "image/jpeg", "image/jpeg", "text/css"}
	fileTypeList	= [...]string{"html", "txt", "gif", "jpeg", "jpg", "css"}
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
		log.Println("Error parsing connection:", err)
		return
	}
	responseWriter := utils.NewConnResponseWriter(conn)

	// Check the request method
	if request.Method == http.MethodGet {
		log.Println("Handing a GET request")
		handleGET(responseWriter, request)
	} else if request.Method == http.MethodPost {
		log.Println("Handing a POST request")
		handlePOST(responseWriter, request)
	} else {
		log.Printf("Handing a %s request, not implemented!", request.Method)
		responseWriter.WriteHeader(http.StatusNotImplemented)
		responseWriter.WriteText("Method Not Implemented")
	}
}

// Handle GET request
func handleGET(w *utils.ConnResponseWriter, r *http.Request) {
	filename := path.Base(r.URL.String())
	filename_list := strings.Split(filename, ".")
	if len(filename_list) <= 1 {
		w.WriteHeader(http.StatusBadRequest)
		w.WriteText("No specified file type")
		return
	}
	valid, contentType := isValidType(filename_list[len(filename_list)-1])
	if !valid {
		log.Println("Error opening file: invalid file type")
		w.WriteHeader(http.StatusBadRequest)
		w.WriteText("Invalid file type")
		return
	}

	// Open file
	targetPath := workspace + "/" + filename
	file, err := os.Open(targetPath)
	if err != nil {
		log.Println("Error opening file:", err)
		w.WriteHeader(http.StatusFound)
		w.WriteText("File not found")
		return
	}
	defer file.Close()

	// Get file info
	fileInfo, err := file.Stat()
	if err != nil {
		log.Println("Error fetching file info:", err)
		return
	}
	buffer := make([]byte, fileInfo.Size())
	_, err = file.Read(buffer)

	// Set Headers
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileInfo.Name()))
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Length", fmt.Sprintf("%d", fileInfo.Size()))

	// Send response
	w.WriteHeader(http.StatusOK)
	w.Write(buffer)
	log.Println("Success sending file:", targetPath)
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
		log.Println("File Uploaded: %+v", h.Header)
		// log.Printf("Received %s file content type: %s", h.Header.Get("Content-Type"), h.Filename)
	}
	w.WriteHeader(http.StatusOK)
	w.WriteText("Files uploaded successfully")
}


// Check whether the target file type is valid
func isValidType(target string) (bool, string) {
	for i, str := range fileTypeList {
		if str == target {
			return true, mediaTypeList[i]
		}
	}
	return false, ""
}


