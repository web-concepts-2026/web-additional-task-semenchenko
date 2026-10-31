# HTTP Server — Request Examples

Server: `http://localhost:8080`

### Run the server

```bash
go run main.go
```

### 1. GET — Serve static files

```bash
curl http://localhost:8080/
```

```bash
curl http://localhost:8080/style.css
```

### 2. GET — Query parameters

```bash
curl "http://localhost:8080/page?id=5"
```

### 3. HTTP Headers

```bash
curl http://localhost:8080/ \
  -H "X-Test: hello" \
  -H "Accept: text/html"
```

### 4. POST — JSON request body

```bash
curl -X POST http://localhost:8080/api/echo \
  -H "Content-Type: application/json" \
  -d '{"grade":"67","passed":true}'
```