package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"strconv"
	"sync"
)

// Channel of simultaneous tasks
var simul chan string
var task_queue = make(chan string)
var wg sync.WaitGroup

// Array of strings with finished tasks
var finished_tasks []string

// Array with directories found
var found_dir []string
var found_files []string

// TODO
// Persistence variables
// Last word used
var lastword string

// Array with pending directories
var pending_dir []string

var targetUrl string
var dict string

func main() {
	// FLAGS
	urlFlag := flag.String("u", "localhost", "The url to test.")
	dictFlag := flag.String("d", "/tmp/dict.txt", "The dictionary to use.")
	//	delayFlag := flag.Int("delay", 0, "Set delay for requests.")
	threadsFlag := flag.Int("threads", 1, "Max number of concurrent HTTP requests.")
	flag.Parse()

	// Handle the flag data
	simul = make(chan string, *threadsFlag)
	var err error
	targetUrl, err = formatUrl(*urlFlag)
	if err != nil {
		log.Println("[E] Error" + err.Error())
	}
	dict = *dictFlag
	// /Handle
	// /FLAGS

	// BANNER
	fmt.Println("[/!\\] Starting bruteforcing with dict: \n\t" + dict + "\n   On site: " + targetUrl + "\n Threads: " + strconv.Itoa(*threadsFlag))
	// /BANNER

	// Run !
	go feed(task_queue, "")
	consume(task_queue, simul)
	wg.Wait()

	// Finished.
	fmt.Printf("\n[FINISHED]\n\n" + strconv.Itoa(len(finished_tasks)) + " finished tasks.")
	fmt.Println("\n[i] Directories found: ")
	fmt.Println(found_dir)
	fmt.Println("[i] Files found: ")
	fmt.Println(found_files)
}

// This function will feed the task channel
func feed(task_queue chan string, prefix string) {
	// First open file and for each line...
	file, err := os.Open(dict)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		task_queue <- prefix + scanner.Text()
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	// If it has no prefix (first time called)
	if len(prefix) == 0 {
		fmt.Println("\r--- [i] Waiting for tasks to end and close the task_queue.\n")
		wg.Wait()
		close(task_queue)
	}
	fmt.Println("\r--- [i] Task feed finished for prefix: " + prefix + "\n")
}

func consume(task_queue chan string, simul chan string) {
	for path := range task_queue {
		simul <- path
		go scan(targetUrl, path, simul)
	}
	// All tasks consumed
}

func scan(targetUrl string, path string, simul chan string) {
	wg.Add(1)
	response, content, err := httpRequest(targetUrl, path, false)
	if err != nil {
		urlerr, ok := err.(*url.Error)
		if !ok {
			panic("[/!\\] no *url.Error.")
		}
		if urlerr.Err != errNoRedirect {
			log.Printf("[Request error] %s", err)
			os.Exit(1)
		}
	}
	switch {
	case response.StatusCode == 404:
		print404(path)
	case response.StatusCode == 200:
		if isListable(content) {
			printListable(path)
			// Add directory to the found list
			found_dir = append(found_dir, path+" (LISTABLE)")
		} else {
			print200(path, strconv.Itoa(len(content)))
			found_files = append(found_files, path+" - Size: "+strconv.Itoa(len(content)))
		}
	case response.StatusCode >= 300 && response.StatusCode <= 399:
		print30x(response, content, path, response.StatusCode)
	case response.StatusCode == 403:
		print403(path, len(content))
		found_files = append(found_files, path+" - Size: "+strconv.Itoa(len(content)))
	case response.StatusCode == 405:
		print405(path, len(content))
		found_files = append(found_files, path+" - Size: "+strconv.Itoa(len(content)))
	}

	finished_tasks = append(finished_tasks, <-simul)
	defer wg.Done()
}
