package main

import (
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"http-scratch/internal/request"
	"http-scratch/internal/response"
	"http-scratch/internal/server"
)

const port = 42069

func resp400() []byte {
	return []byte(`<html>
  <head>
    <title>400 Bad Request</title>
  </head>
  <body>
    <h1>Bad Request</h1>
    <p>Your request honestly kinda sucked.</p>
  </body>
</html>`)
}

func resp500() []byte {
	return []byte(`<html>
  <head>
    <title>500 Internal Server Error</title>
  </head>
  <body>
    <h1>Internal Server Error</h1>
    <p>Okay, you know what? This one is on me.</p>
  </body>
</html>`)
}

func resp200() []byte {
	return []byte(`<html>
  <head>
    <title>200 OK</title>
  </head>
  <body>
    <h1>Success!</h1>
    <p>Your request was an absolute banger.</p>
  </body>
</html>`)
}

func main() {
	s, err := server.Serve(port, func(w *response.Writer, req *request.Request) {
		body := resp200()
		status := response.StatusOK

		switch req.RequestLine.RequestTarget {
		case "/yourproblem":
			body = resp400()
			status = response.StatusBadRequest
		case "/myproblem":
			body = resp500()
			status = response.StatusInternalServerError
		}

		h := response.GetDefaultHeaders(0)
		h.Replace("content-length", strconv.Itoa(len(body)))
		h.Replace("content-type", "text/html")
		w.WriteStatusLine(status)
		w.WriteHeaders(h)
		w.WriteBody(body)
	})
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer s.Close()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}
