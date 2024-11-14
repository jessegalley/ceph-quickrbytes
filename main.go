// ceph-quickrbytes
//
package main

import (
	"fmt"
	// "io/fs"
	"os"
	"path/filepath"
	"sync"
	"syscall"
)

type Work struct {
	path string
}

type Result struct {
	path  string
	bytes string
	err   error
}

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <directory>\n", os.Args[0])
		os.Exit(1)
	}

	baseDir := os.Args[1]

	workChan := make(chan Work, 100)
	resultChan := make(chan Result, 100)
	var wg sync.WaitGroup

	// Start worker goroutines
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go worker(workChan, resultChan, &wg)
	}

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	go func() {
		entries, err := os.ReadDir(baseDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading directory: %v\n", err)
			close(workChan)
			return
		}

		for _, entry := range entries {
			if entry.IsDir() {
				workChan <- Work{path: filepath.Join(baseDir, entry.Name())}
			}
		}
		close(workChan)
	}()

	fmt.Printf("directory\tbytes\n")

	for result := range resultChan {
		if result.err != nil {
			fmt.Fprintf(os.Stderr, "Error processing %s: %v\n", result.path, result.err)
			continue
		}
		fmt.Printf("%s\t%s\n", result.path, result.bytes)
	}
}

func worker(workChan <-chan Work, resultChan chan<- Result, wg *sync.WaitGroup) {
	defer wg.Done()

	for work := range workChan {
		bytes, err := getXattr(work.path, "ceph.dir.rbytes")
		resultChan <- Result{
			path:  work.path,
			bytes: string(bytes),
			err:   err,
		}
	}
}

func getXattr(path string, attr string) ([]byte, error) {
	size, err := syscall.Getxattr(path, attr, nil)
	if err != nil {
		return nil, err
	}

	data := make([]byte, size)
	_, err = syscall.Getxattr(path, attr, data)
	if err != nil {
		return nil, err
	}

	return data, nil
}
