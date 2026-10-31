package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
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
	defer conn.Close()

	reader := bufio.NewReader(conn)

	// Беремо перший рядок запиту, типу "GET /page?id=5 HTTP/1.1"
	startLine, err := reader.ReadString('\n')
	if err != nil {
		return
	}
	startLine = strings.TrimSpace(startLine)
	fmt.Println("\n--> REQUEST:", startLine)

	// Ділимо його по пробілах на три частини
	parts := strings.Split(startLine, " ")
	if len(parts) < 3 {
		// запит порожній або некоректний
		sendResponse(conn, 400, "Bad Request", "text/plain; charset=utf-8", "Bad request line")
		return
	}
	method := parts[0]
	fullPath := parts[1]

	// Відділяємо шлях від квері
	path := fullPath
	rawQuery := ""
	if strings.Contains(fullPath, "?") {
		pieces := strings.SplitN(fullPath, "?", 2)
		path = pieces[0]
		rawQuery = pieces[1]
	}
	query := parseQuery(rawQuery)

	// http headers
	headers := make(map[string]string)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return
		}
		line = strings.TrimSpace(line)
		if line == "" {
			break
		}
		fmt.Println("    header:", line)

		kv := strings.SplitN(line, ":", 2)
		if len(kv) == 2 {
			name := strings.ToLower(strings.TrimSpace(kv[0]))
			headers[name] = strings.TrimSpace(kv[1])
		}
	}

	// http body
	body := ""
	if lengthText, ok := headers["content-length"]; ok {
		length, err := strconv.Atoi(lengthText)
		if err != nil || length < 0 {
			sendResponse(conn, 400, "Bad Request", "text/plain; charset=utf-8", "Bad Content-Length")
			return
		}
		buf := make([]byte, length)
		_, err = io.ReadFull(reader, buf)
		if err != nil {
			return
		}
		body = string(buf)
		fmt.Println("    body:", body)
	}

	// http response
	if method == "POST" && path == "/api/echo" {
		if !strings.Contains(headers["content-type"], "application/json") {
			sendResponse(conn, 415, "Unsupported Media Type", "application/json",
				`{"error": "Content-Type must be application/json"}`)
			return
		}

		var data map[string]any
		err := json.Unmarshal([]byte(body), &data)
		if err != nil {
			sendResponse(conn, 400, "Bad Request", "application/json",
				`{"error": "Invalid JSON"}`)
			return
		}

		answer := map[string]any{
			"status":   "ok",
			"received": data,
		}
		answerBytes, _ := json.Marshal(answer)
		sendResponse(conn, 200, "OK", "application/json", string(answerBytes))

	} else if method == "GET" && path == "/page" {
		id, hasID := query["id"]
		if !hasID {
			sendResponse(conn, 400, "Bad Request", "text/plain; charset=utf-8", "Parameter id is required")
			return
		}
		page := fmt.Sprintf(`
			<!DOCTYPE html>
			<html lang="en">
			<head>
  				<meta charset="utf-8">
				<title>My Go Server</title>
  				<link rel="stylesheet" href="/style.css">
			</head>
			<body>
    			<h1>Page #%s</h1>
			</body>
			</html>`, id)

		sendResponse(conn, 200, "OK", "text/html; charset=utf-8", page)

	} else if method == "GET" {
		serveFile(conn, path)

	} else {
		// Усі інші методи не підтримуються
		sendResponse(conn, 405, "Method Not Allowed", "text/plain; charset=utf-8", "Method not allowed")
	}
}

// serveFile шукає файл у папці static за шляхом із запиту
// і відправляє його клієнту
func serveFile(conn net.Conn, path string) {
	if path == "/" {
		path = "/index.html"
	}

	// забороняємо ".." щоб не вийти за межі папки static
	if strings.Contains(path, "..") {
		sendResponse(conn, 404, "Not Found", "text/plain; charset=utf-8", "File not found")
		return
	}

	// Читаємо файл з диска: "/style.css" → "static/style.css"
	content, err := os.ReadFile("static" + path)
	if err != nil {
		sendResponse(conn, 404, "Not Found", "text/plain; charset=utf-8", "File not found")
		return
	}

	// Тип вмісту визначаємо за закінченням назви файлу
	contentType := "application/octet-stream"
	if strings.HasSuffix(path, ".html") {
		contentType = "text/html; charset=utf-8"
	} else if strings.HasSuffix(path, ".css") {
		contentType = "text/css; charset=utf-8"
	} else if strings.HasSuffix(path, ".js") {
		contentType = "application/javascript; charset=utf-8"
	} else if strings.HasSuffix(path, ".txt") {
		contentType = "text/plain; charset=utf-8"
	}

	sendResponse(conn, 200, "OK", contentType, string(content))
}

// sendResponse формує http response
func sendResponse(conn net.Conn, status int, statusText string, contentType string, body string) {
	fmt.Printf("<-- RESPONSE: %d %s | %s\n", status, statusText, contentType)
	// Довгі відповіді (наприклад, файли) не друкуємо, лише їхній розмір
	if len(body) > 200 {
		fmt.Println("    body:", len(body), "bytes")
	} else {
		fmt.Println("    body:", body)
	}

	fmt.Fprintf(conn, "HTTP/1.1 %d %s\r\n"+
		"Content-Type: %s\r\n"+
		"Content-Length: %d\r\n"+
		"Connection: close\r\n"+
		"\r\n"+
		"%s", status, statusText, contentType, len(body), body)
}

// parseQuery перетворює query-рядок на map:
// "grade=67&passed=true" → {"grade": "67", "passed": "true"}
func parseQuery(raw string) map[string]string {
	result := make(map[string]string)
	if raw == "" {
		return result
	}
	for pair := range strings.SplitSeq(raw, "&") {
		key, value, ok := strings.Cut(pair, "=")
		if ok {
			result[key] = value
		}
	}
	return result
}
