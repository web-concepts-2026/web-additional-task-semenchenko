package main

import (
	"fmt"
	"net"
	"strings"
)

func main() {
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println("Server is listening at http://localhost:8080")

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error:", err)
			continue
		}

		handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {

	buffer := make([]byte, 4096)
	n, _ := conn.Read(buffer)
	request := string(buffer[:n])

	fmt.Println("REQUEST")
	fmt.Println(request)
	fmt.Println("REQUEST END")

	// Беремо перший рядок запиту, типу "GET /page?id=5 HTTP/1.1"
	lines := strings.Split(request, "\r\n")
	startLine := lines[0]

	// Ділимо його по пробілах на три частини
	parts := strings.Split(startLine, " ")
	if len(parts) < 3 {
		return // запит порожній або некоректний
	}
	method := parts[0]   // метод як "GET"
	fullPath := parts[1] // шлях плюс кверя "/page?id=5"

	// Відділяємо шлях від квері
	path := fullPath
	rawQuery := ""
	if strings.Contains(fullPath, "?") {
		pieces := strings.SplitN(fullPath, "?", 2)
		path = pieces[0]     // "/page"
		rawQuery = pieces[1] // "id=5"
	}

	fmt.Println("method:", method, "| path:", path, "| query:", rawQuery)

	body := fmt.Sprintf("method=%s\npath=%s\nquery=%s\n", method, path, rawQuery)
	fmt.Fprintf(conn, "HTTP/1.1 200 OK\r\n"+
		"Content-Type: text/plain; charset=utf-8\r\n"+
		"Content-Length: %d\r\n"+
		"\r\n"+
		"%s", len(body), body)

	conn.Close()
}
