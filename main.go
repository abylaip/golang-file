package main

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"log"
	"net"
	"os"
	"strings"
)

func main() {
	arguments := os.Args
	if len(arguments) == 1 {
		fmt.Println("Provide a port")
		return
	}

	PORT := fmt.Sprintf(":%v", arguments[1])
	server, err := net.Listen("tcp4", PORT)
	if err != nil {
		log.Fatal(err)
	}
	defer server.Close()

	for {
		fmt.Println("Listening...")
		client, err := server.Accept()
		if err != nil {
			fmt.Println(err)
			return
		}

		go handleConnection(client)
	}
}

// Hardware: wifi adapter
// Driver: reads data from wifi
// OS: has cache 1kb

// func readAll() []byte {
// 	fullData := make([]byte, 8*1024*1024)
// }

type Request struct {
	Method string
	Path   string

	Header map[string]string
	Body   []byte
}

func ParseRequest(data []byte) *Request {
	var firstLine strings.Builder
	index := 0

	for data[index] != '\n' {
		firstLine.WriteByte(data[index])
		index++
	}

	chunks := strings.Split(firstLine.String(), " ")

	method := chunks[0]
	path := chunks[1]

	request := &Request{
		Method: method,
		Path:   path,
		Header: make(map[string]string),
		Body:   nil,
	}
	var line strings.Builder
	for index < len(data) {
		if index+1 < len(data) && (data[index] == '\n' && data[index+1] == '\n') {
			break
		}
		if data[index] == '\n' {
			chunks := strings.SplitN(line.String(), ":", 2)
			if len(chunks) == 2 {
				key := strings.TrimSpace(chunks[0])
				value := strings.TrimSpace(chunks[1])
				request.Header[key] = value
			}
			line.Reset()
		} else {
			line.WriteByte(data[index])
		}
		index++
	}

	if method == "POST" {
		if _, ok := request.Header["Content-Length"]; ok {
			body := data[index:]
			request.Body = bytes.Clone(body[:len(body)])
		}
	}

	return request
}

func handleConnection(client net.Conn) {
	fmt.Printf("Serving: %v\n", client.RemoteAddr().String())
	tmp := make([]byte, 8*1024*1024)
	defer client.Close()
	n, err := client.Read(tmp)
	if err != nil {
		if err != io.EOF {
			fmt.Println("read error")
		}
		return
	}
	fmt.Printf("size = %d", n)
	// fmt.Println("--- begin ----")
	// fmt.Println(string(tmp))
	// fmt.Println("--- end ----")

	request := ParseRequest(tmp[:n])
	var response string
	if request.Path == "/" {
		response = "HTTP/1.1 200 OK\n" + "Content-Type: text/html;\n\n"
		response += handleHomePath()
	} else if request.Path == "/upload" && request.Method == "POST" {
		handleUpload(request)
		response = "HTTP/1.1 201 Created\n" + "Content-Type: text/html;\n\nImage successfuly created"
	}

	client.Write([]byte(response))
}

func handleHomePath() string {
	data, err := os.ReadFile("public/index.html")
	if err == nil {
		return string(data)
	} else {
		fmt.Println(err)
		return "error reading file"
	}
}

func handleUpload(request *Request) {
	fmt.Println(request.Path, request.Method, request.Body)
	reader := bytes.NewReader(request.Body)

	img, _, err := image.Decode(reader)
	if err != nil {
		log.Println("error decoding image")
		return
	}

	outFile, err := os.Create("output.jpg")
	if err != nil {
		log.Println("error creating file")
		return
	}
	defer outFile.Close()

	err = jpeg.Encode(outFile, img, nil)
	if err != nil {
		log.Println("error encoding new image")
		return
	}
}
