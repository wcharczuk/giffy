package util

import (
	"bufio"
	"io"
	"io/ioutil"
	"os"

	"github.com/blendlabs/go-exception"
)

// ReadFile reads a file as a string.
func ReadFile(path string) string {
	bytes, _ := ioutil.ReadFile(path)
	return string(bytes)
}

// ReadChunkHandler is a receiver for a chunk of a file.
type ReadChunkHandler func(line []byte)

//ReadLineHandler is a receiver for a line of a file.
type ReadLineHandler func(line string)

// ReadFileByLines reads a file and calls the handler for each line.
func ReadFileByLines(filePath string, handler ReadLineHandler) error {
	if f, err := os.Open(filePath); err == nil {
		defer f.Close()

		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := scanner.Text()
			handler(line)
		}
	} else {
		return exception.Wrap(err)
	}
	return nil
}

// ReadFileByChunks reads a file in `chunkSize` pieces, dispatched to the handler.
func ReadFileByChunks(filePath string, chunkSize int, handler ReadChunkHandler) error {
	if f, err := os.Open(filePath); err == nil {
		defer f.Close()

		chunk := make([]byte, chunkSize)
		for {
			readBytes, err := f.Read(chunk)
			if err == io.EOF {
				break
			}
			readData := chunk[:readBytes]
			handler(readData)
		}
	} else {
		return exception.Wrap(err)
	}
	return nil
}
