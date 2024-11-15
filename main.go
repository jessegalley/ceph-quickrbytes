package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"

	"github.com/spf13/pflag"
)

type Work struct {
	path string
}

type Result struct {
	path  string
	bytes string
	err   error
}

const (
	KB = 1024
	MB = 1024 * KB
	GB = 1024 * MB
)

func formatBytes(bytes string, unit string) (string, error) {
	bytesInt, err := strconv.ParseInt(strings.TrimSpace(bytes), 10, 64)
	if err != nil {
		return "", fmt.Errorf("error parsing bytes: %v", err)
	}

	switch strings.ToLower(unit) {
	case "kb":
		return fmt.Sprintf("%.2f KB", float64(bytesInt)/float64(KB)), nil
	case "mb":
		return fmt.Sprintf("%.2f MB", float64(bytesInt)/float64(MB)), nil
	case "gb":
		return fmt.Sprintf("%.2f GB", float64(bytesInt)/float64(GB)), nil
	default:
		return bytes, nil
	}
}

func main() {
	units := pflag.StringP("units", "u", "", "display units (kb, mb, gb)")
	pflag.Parse()

	args := pflag.Args()
	if len(args) != 1 {
		fmt.Fprintf(os.Stderr, "Usage: %s [--units kb|mb|gb] <parent_directory>\n", os.Args[0])
		os.Exit(1)
	}

	baseDir := args[0]

	workChan := make(chan Work, 100)
	resultChan := make(chan Result, 100)
	var wg sync.WaitGroup

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
			fmt.Fprintf(os.Stderr, "error reading directory: %v\n", err)
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
			fmt.Fprintf(os.Stderr, "error processing %s: %v\n", result.path, result.err)
			continue
		}

		formattedBytes := result.bytes
		if *units != "" {
			formatted, err := formatBytes(result.bytes, *units)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error formatting bytes for %s: %v\n", result.path, err)
				continue
			}
			formattedBytes = formatted
		}

		fmt.Printf("%s\t%s\n", result.path, formattedBytes)
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
