# Simple HTTP Proxy in Go

A lightweight HTTP proxy server written in Go. Forwards requests, copies headers, streams responses, and logs activity using `slog`.

---

## Features

- Forward HTTP requests (method, URL, body, headers).
- Filter hop-by-hop headers (`Connection`, `Keep-Alive`, etc.).
- Stream responses back to the client.
- Structured logging of requests and errors.

---

## Usage

Run the server:

```bash
go run main.go
```

Default port: 8080

Test with curl:

```bash
curl -x localhost:8080 http://example.com
```

## Todo

- [x] Log duration and other metadata
- [x] Implement caching
- [x] Way to handle indefinitely growing cache
- [ ] Implement support for https
- [ ] Implement TTL with the new cache
