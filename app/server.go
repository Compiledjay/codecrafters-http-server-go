package main

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"fmt"
	"net"
	"os"
	"strings"
)

const (
	// response type
	Response200 = "HTTP/1.1 200 OK\r\n"
	Response201 = "HTTP/1.1 201 Created\r\n"
	Response400 = "HTTP/1.1 400 Bad Request\r\n"
	Response404 = "HTTP/1.1 404 Not Found\r\n"
	// headers
	ContentTextPlain      = "Content-Type: text/plain\r\n"
	ContentAppOctetStream = "Content-Type: application/octet-stream\r\n"
	ContentLength         = "Content-Length: "
	ContentEncoding       = "Content-Encoding: "
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
	headers := make(map[string]string)
	for _, header := range postReq {
		if header == "" {
			break
		}

		headerField := strings.SplitN(header, ":", 2)
		name := strings.Trim(headerField[0], " ")
		value := strings.Trim(headerField[1], " ")
		headers[strings.ToLower(name)] = value
	}
	body := lines[len(lines)-1]

	return &httpRequest{
		reqType,
		path,
		headers,
		body,
	}, nil
}

func createGetResponse(r *httpRequest) string {
	var response string
	switch p := strings.Split(r.path, "/")[1]; p {
	case "":
		response = fmt.Sprintf("%s\r\n", Response200)
	case "echo":
		splitReq := strings.SplitN(r.path, "/", 3)
		headers := ""
		gzipCompressed := false
		if v, ok := r.headers["accept-encoding"]; ok {
			headerValues := strings.Split(v, ", ")
			for _, s := range headerValues {
				switch s {
				case "gzip":
					headers += ContentEncoding + s + "\r\n"
					gzipCompressed = true
				}
			}
		}

		body := splitReq[2]
		if gzipCompressed {
			var buf bytes.Buffer
			gzipWriter := gzip.NewWriter(&buf)
			gzipWriter.Write([]byte(body))
			gzipWriter.Close()
			body = buf.String()
		}

		headers += fmt.Sprintf("%s%s%d\r\n\r\n", ContentTextPlain, ContentLength, len(body))
		response = fmt.Sprintf("%s%s%s", Response200, headers, body)
	case "user-agent":
		body := r.headers["user-agent"]
		body = strings.Trim(body, "\r\n")
		headers := fmt.Sprintf("%s%s%d\r\n\r\n", ContentTextPlain, ContentLength, len(body))
		response = fmt.Sprintf("%s%s%s", Response200, headers, body)
	case "files":
		if len(os.Args) < 3 && os.Args[1] != "--directory" {
			fmt.Println("invalid arguments, usage: ./server.sh --directory <directory>")
			response = fmt.Sprintf("%s\r\n", Response400)
			break
		}

		dir := os.Args[2]
		fileName := strings.Split(r.path, "/")[2]
		file, err := os.Open(fmt.Sprintf("%s/%s", dir, fileName))
		if err != nil {
			fmt.Println("error opening file in directory: ", err.Error())
			response = fmt.Sprintf("%s\r\n", Response404)
			break
		}
		defer file.Close()

		body := ""
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			body += scanner.Text()
		}
		headers := fmt.Sprintf("%s%s%d\r\n\r\n", ContentAppOctetStream, ContentLength, len(body))
		response = fmt.Sprintf("%s%s%s", Response200, headers, body)
	default:
		fmt.Println("unrecognized path: ", r.path)
		response = fmt.Sprintf("%s\r\n", Response404)
	}

	return response
}

func createPostResponse(r *httpRequest) string {
	var response string
	switch p := strings.Split(r.path, "/")[1]; p {
	case "files":
		if len(os.Args) < 3 && os.Args[1] != "--directory" {
			fmt.Println("invalid arguments, usage: ./server.sh --directory <directory>")
			response = fmt.Sprintf("%s\r\n", Response400)
			break
		}

		dir := os.Args[2]
		fileName := strings.Split(r.path, "/")[2]
		file, err := os.Create(fmt.Sprintf("%s/%s", dir, fileName))
		if err != nil {
			fmt.Println("error creating file in directory: ", err.Error())
			response = fmt.Sprintf("%s\r\n", Response404)
			break
		}
		defer file.Close()

		_, err = file.WriteString(r.body)
		if err != nil {
			fmt.Println("error writing to file: ", err.Error())
			response = fmt.Sprintf("%s\r\n", Response400)
			break
		}
		headers := fmt.Sprintf("%s%s%d\r\n\r\n", ContentAppOctetStream, ContentLength, len(r.body))
		response = fmt.Sprintf("%s%s%s", Response201, headers, r.body)
	default:
		fmt.Println("unrecognized path: ", r.path)
		response = fmt.Sprintf("%s\r\n", Response404)
	}

	return response
}

func manageConnection(conn net.Conn) {
	bufSize := 1024
	buffer := make([]byte, bufSize)
	data := ""
	for {
		n, err := conn.Read(buffer)
		if err != nil {
			fmt.Println("Error reading from connection: ", err.Error())
			os.Exit(1)
		}

		data += string(buffer[:n])
		if n < bufSize {
			break
		}
	}

	request, err := parseRequest(data)
	if err != nil {
		fmt.Println("Error parsing request: ", err.Error())
	}

	var response string
	switch request.reqType {
	case "GET":
		response = createGetResponse(request)
	case "POST":
		response = createPostResponse(request)
	default:
		fmt.Println("Invalid request: ", request.reqType)
		os.Exit(1)
	}

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
