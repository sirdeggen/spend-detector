package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"runtime"
	"sync"
	"time"
)

const hashSize = 32

var target = []byte{
	0xe8, 0xc1, 0xcc, 0x5e, 0x0c, 0xc7, 0x17, 0x04,
	0xf5, 0x2a, 0xea, 0x09, 0xa5, 0x70, 0x8a, 0x4b,
	0x3d, 0x0b, 0x05, 0xa6, 0xa6, 0x60, 0xe5, 0x99,
	0x4e, 0x0f, 0x42, 0x53, 0x52, 0x78, 0x0e, 0x0d,
}

func searchInChunk(data []byte, globalOffset int64, target []byte, found chan int64, start time.Time) {
	dataLen := len(data)
	targetLen := len(target)
	
	for i := 0; i <= dataLen-targetLen; i++ {
		select {
		case <-found:
			return // early exit if found by another goroutine
		default:
			if bytes.Equal(data[i:i+targetLen], target) {
				fmt.Printf("*** FOUND at offset %d! ***\n", globalOffset+int64(i))
				fmt.Printf("Execution time: %s\n", time.Since(start))
				select {
				case found <- globalOffset + int64(i):
				default:
				}
				return
			}
		}
	}
}


func main() {
	start := time.Now()
	filename := "block1324542.bin"

	numWorkers := runtime.NumCPU()
	runtime.GOMAXPROCS(numWorkers)


	file, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	// Get file size
	fileInfo, err := file.Stat()
	if err != nil {
		panic(err)
	}
	fileSize := fileInfo.Size()
	fmt.Printf("File size: %d bytes (%.2f GB)\n", fileSize, float64(fileSize)/(1024*1024*1024))

	// Use 1MB chunks for streaming
	const chunkSize = 1024 * 1024
	// Overlap size to handle target spanning chunk boundaries
	overlapSize := hashSize - 1
	var wg sync.WaitGroup
	found := make(chan int64, 1)
	chunks := make(chan []byte, numWorkers*2) // buffered channel for chunks
	offsets := make(chan int64, numWorkers*2) // corresponding offsets

	// Start worker goroutines
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for {
				select {
				case <-found:
					return // early exit
				case chunk, ok := <-chunks:
					if !ok {
						return // channel closed
					}
					offset := <-offsets
					searchInChunk(chunk, offset, target, found, start)
				}
			}
		}(i)
	}

	// Read and distribute chunks with proper overlap
	go func() {
		defer close(chunks)
		defer close(offsets)
		
		var chunkStart int64 = 0
		
		for chunkStart < fileSize {
			// Calculate actual read size and position
			readStart := chunkStart
			readSize := int64(chunkSize)
			
			// Add overlap to the end (except for last chunk)
			if chunkStart + readSize < fileSize {
				readSize += int64(overlapSize)
			}
			
			// Don't read past end of file
			if readStart + readSize > fileSize {
				readSize = fileSize - readStart
			}
			
			buffer := make([]byte, readSize)
			n, err := file.ReadAt(buffer, readStart)
			if err != nil && err != io.EOF {
				return
			}
			buffer = buffer[:n]
			
			// Send chunk to workers
			select {
			case <-found:
				return // early exit
			case chunks <- buffer:
				offsets <- readStart
			}
			
			// Move to next chunk (no gap, no double-counting)
			chunkStart += int64(chunkSize)
		}
	}()

	wg.Wait()
	close(found)


}