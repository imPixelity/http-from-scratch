package main

import (
	"errors"
	"fmt"
	"io"
	"os"
)

func main() {
	f, err := os.Open("messages.txt")
	if err != nil {
		panic(err)
	}

	for data := range getLinesChannel(f) {
		fmt.Printf("read: %s\n", data)
	}
}

func getLinesChannel(f io.ReadCloser) <-chan string {
	out := make(chan string)

	go func() {
		defer close(out)
		defer f.Close()

		buf := make([]byte, 8)
		var line []byte
		for {
			read, err := f.Read(buf)

			if read > 0 {
				for _, v := range buf[:read] {
					if v != '\n' {
						line = append(line, v)
						continue
					}
					if len(line) > 0 {
						out <- string(line)
					}
					line = nil
				}
			}

			if err != nil {
				if len(line) > 0 {
					out <- string(line)
				}
				if errors.Is(err, io.EOF) {
					break
				}
				break
			}
		}
	}()

	return out
}
