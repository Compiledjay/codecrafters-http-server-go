package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

const (
	// response type
	Response200 = "HTTP/1.1 200 OK\r\n"
	Response404 = "HTTP/1.1 404 Not Found\r\n"
	// headers
	ContentTextPlain      = "Content-Type: text/plain\r\n"
	ContentAppOctetStream = "Content-Type: application/octet-stream\r\n"
	ContentLength         = "Content-Length: "
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
	response := ""
	switch p := strings.Split(r.path, "/")[1]; p {
	case "":
		response = fmt.Sprintf("%s\r\n", Response200)
	case "echo":
		splitReq := strings.SplitN(r.path, "/", 3)
		body := splitReq[2]
		response = fmt.Sprintf("%s%s%s%d\r\n\r\n%s", Response200, ContentTextPlain, ContentLength, len(body), body)
	case "user-agent":
		body := r.headers["user-agent"]
		body = strings.Trim(body, "\r\n")
		response = fmt.Sprintf("%s%s%s%d\r\n\r\n%s", Response200, ContentTextPlain, ContentLength, len(body), body)
	case "files":
		if len(os.Args) < 3 {
			response = fmt.Sprintf("%s\r\n", Response404)
			break
		}
		dir := os.Args[2]
		splitReq := strings.Split(r.path, "/")
		fileName := splitReq[2]
		file, err := os.Open(dir + fileName)
		if err != nil {
			response = fmt.Sprintf("%s\r\n", Response404)
			break
		}
		defer file.Close()

		body := ""
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			body += scanner.Text()
		}
		response = fmt.Sprintf("%s%s%s%d\r\n\r\n%s", Response200, ContentAppOctetStream, ContentLength, len(body), body)
	default:
		response = fmt.Sprintf("%s\r\n", Response404)
	}

	return response
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
	fmt.Println([]byte(response))
	fmt.Println(response)

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
