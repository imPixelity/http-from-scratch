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
					fmt.Printf("read: %s\n", line)
				}
				line = nil
			}
		}

		if err != nil {
			if len(line) > 0 {
				fmt.Printf("read: %s\n", line)
			}
			if errors.Is(err, io.EOF) {
				break
			}
			break
		}
	}
}
