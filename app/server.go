package main

import (
	"fmt"
	"net"
	"os"
	"slices"
	"strings"
)

const (
	// response type
	httpOk       = "HTTP/1.1 200 OK\r\n"
	httpNotFound = "HTTP/1.1 404 Not Found\r\n"
	// headers
	headerCTypeTextPlain = "Content-Type: text/plain\r\n"
	headerCTypeLength    = "Content-Length: "
)

func main() {
	fmt.Println("Logs from your program will appear here!")

	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221", err.Error())
		os.Exit(1)
	}

	conn, err := l.Accept()
	if err != nil {
		fmt.Println("Error accepting connection: ", err.Error())
		os.Exit(1)
	}

	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil {
		fmt.Println("Error reading from connection: ", err.Error())
		os.Exit(1)
	}

	request := string(buffer[:n])
	req_segments := strings.Split(request, "\r\n")
	req_segments = slices.DeleteFunc(req_segments, func(s string) bool {
		return s == ""
	})

	path := strings.Split(req_segments[0], " ")[1]
	if path == "/" {
		_, err = conn.Write([]byte(httpOk + "\r\n"))
	} else if path[0:6] == "/echo/" {
		body := path[6:]
		_, err = conn.Write([]byte(fmt.Sprintf("%s%s%s%d\r\n\r\n%s", httpOk, headerCTypeTextPlain, headerCTypeLength, len(body), body)))
	} else {
		_, err = conn.Write([]byte(httpNotFound + "\r\n"))
	}
	if err != nil {
		fmt.Println("Error writing to connection: ", err.Error())
		os.Exit(1)
	}
}
