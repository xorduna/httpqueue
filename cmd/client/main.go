package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

const MAX_RETRIES = 3

func main() {
	if len(os.Args) != 5 {
		fmt.Println("Usage: client <write/read> <server> <queue-name> <filename>")
		os.Exit(1)
	}

	mode := os.Args[1]
	server := os.Args[2]
	queueName := os.Args[3]
	filename := os.Args[4]

	switch mode {
	case "write":
		if err := writeMode(server, queueName, filename); err != nil {
			fmt.Fprintf(os.Stderr, "Error in write mode: %v\n", err)
			os.Exit(1)
		}
	case "read":
		if err := readMode(server, queueName, filename); err != nil {
			fmt.Fprintf(os.Stderr, "Error in read mode: %v\n", err)
			os.Exit(1)
		}
	default:
		fmt.Fprintf(os.Stderr, "Invalid mode: %s. Use 'write' or 'read'\n", mode)
		os.Exit(1)
	}
}

func writeMode(server, queueName, filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("cannot open file: %v", err)
	}
	defer file.Close()

	pushURL := fmt.Sprintf("%s/queues/%s", server, queueName)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		resp, err := http.Post(pushURL, "text/plain", strings.NewReader(line))
		if err != nil {
			return fmt.Errorf("error making request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("server returned status %d", resp.StatusCode)
		}

		log.Printf("SENT DATA\n")
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading file: %v", err)
	}

	return nil
}

func readMode(server, queueName, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("cannot create file: %v", err)
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	defer writer.Flush()

	pullURL := fmt.Sprintf("%s/queues/%s", server, queueName)
	timeouts := 0

	for timeouts < MAX_RETRIES {

		resp, err := http.Get(pullURL)
		if err != nil {
			return fmt.Errorf("error making request: %v", err)
		}

		switch resp.StatusCode {
		case http.StatusOK:
			// Llegim el missatge
			body, err := io.ReadAll(resp.Body)
			resp.Body.Close()
			if err != nil {
				return fmt.Errorf("error reading response: %v", err)
			}

			log.Printf("RECEIVED DATA\n")

			if _, err := writer.WriteString(string(body) + "\n"); err != nil {
				return fmt.Errorf("error writing to file: %v", err)
			}

			// We flush at each line just in case something wrong happens
			writer.Flush()

			// Reset timeout
			timeouts = 0

		case http.StatusNoContent:
			// Timeout, increment counter
			timeouts++
			resp.Body.Close()

			if timeouts < MAX_RETRIES {
				log.Printf("Timeout! %d retries already ...\n", timeouts)
			} else {
				log.Printf("Done waiting, %d total retries\n", timeouts)
			}

		default:
			resp.Body.Close()
			return fmt.Errorf("server returned unexpected status %d", resp.StatusCode)
		}
	}

	return nil
}
