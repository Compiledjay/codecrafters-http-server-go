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
	Response200 = "HTTP/1.1 200 OK\r\n"
	Response404 = "HTTP/1.1 404 Not Found\r\n"
	// headers
	ContentTextPlain = "Content-Type: text/plain\r\n"
	ContentLength    = "Content-Length: "
)

type httpRequest struct {
	reqType string
	path    string
	headers map[string]string
	body    string
}

func parseRequest(req string) (*httpRequest, error) {
	lines := strings.Split(req, "\r\n")
	reqLine := strings.Split(lines[0], " ")

	reqType := reqLine[0]
	path := reqLine[1]

	postReq := lines[1:]
	mapHeaders := make(map[string]string)
	for _, header := range postReq {
		if header == "" {
			break
		}

		headerField := strings.SplitN(header, ":", 2)
		name := strings.Trim(headerField[0], " ")
		value := strings.Trim(headerField[1], " ")
		mapHeaders[strings.ToLower(name)] = value + "\r\n"
	}
	body := lines[len(lines)-1]

	return &httpRequest{
		reqType,
		path,
		mapHeaders,
		body,
	}, nil
}

func createResponse(r *httpRequest) string {
	var validRequest = func(path string, allowedRequests []string) bool {
		if path == "/" {
			return true
		}

		splitPath := strings.Split(path, "/")
		partsPath := slices.DeleteFunc(splitPath, func(s string) bool {
			return s == ""
		})

		p := partsPath[0]
		for _, v := range allowedRequests {
			if strings.EqualFold(p, v) {
				return true
			}
		}
		return false
	}
	allowedRequests := []string{"echo", "user-agent"}

	if !validRequest(r.path, allowedRequests) {
		return fmt.Sprintf("%s\r\n", Response404)
	} else if r.path == "/" {
		return fmt.Sprintf("%s\r\n", Response200)
	} else if r.path == "/user-agent" {
		body := r.headers["user-agent"]
		body = strings.Trim(body, "\r\n")
		return fmt.Sprintf("%s%s%s%d\r\n\r\n%s", Response200, ContentTextPlain, ContentLength, len(body), body)
	} else {
		echo := strings.Split(r.path, "/")
		body := echo[len(echo)-1]
		return fmt.Sprintf("%s%s%s%d\r\n\r\n%s", Response200, ContentTextPlain, ContentLength, len(body), body)
	}
}

func manageConnection(conn net.Conn) {
	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil {
		fmt.Println("Error reading from connection: ", err.Error())
		os.Exit(1)
	}

	request, err := parseRequest(string(buffer[:n]))
	if err != nil {
		fmt.Println("Error parsing request: ", err.Error())
	}

	response := createResponse(request)
	fmt.Println(response)
	fmt.Println([]byte(response))
	conn.Write([]byte(response))
	conn.Close()
}

func main() {
	fmt.Println("Logs from your program will appear here!")

	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221", err.Error())
		os.Exit(1)
	}
	defer l.Close()

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}
		go manageConnection(conn)
	}
}
