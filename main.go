package main

import (
	"fmt"
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
		fmt.Println("Litening...")
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
}

func parseRequest(data []byte) *Request {
	firstLine := ""
	index := 0

	for data[index] != '\n' {
		firstLine += string(data[index])
		index++
	}

	chunks := strings.Split(firstLine, " ")

	method := chunks[0]
	path := chunks[1]

	request := &Request{
		Method: method,
		Path:   path,
	}
	line := ""
	for index < len(data) {
		if data[index] == ' ' {
			continue
		}
		line += string(data[index])
		if data[index] == '\n' {
			chunks := strings.Split(line, ":")
			fmt.Println(chunks)
			// request.Header[chunks[0]] = chunks[1]
			line = ""
		}
		index++
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
	fmt.Println("--- begin ----")
	fmt.Println(string(tmp))
	fmt.Println("--- end ----")

	request := parseRequest(tmp)

	fmt.Println(request)

	var response string
	if request.Path == "/" {
		response = "HTTP/1.1 200 OK\n" + "Content-Type: text/html;\n\n"
		response += handleHomePath()
	} else if request.Path == "/upload" && request.Method == "POST" {
		handleUpload()
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

func handleUpload() {
	fmt.Println("handleUpload")
}
