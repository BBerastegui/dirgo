package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"
)

var red = "\\033[0;31m"

// Channel of simultaneous tasks
var simul = make(chan string, 10)
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
	urlFlag := flag.String("u", "foo", "the url to test")
	dictFlag := flag.String("d", "foo", "the dictionary to use")
	flag.Parse()
	var err error
	targetUrl, err = formatUrl(*urlFlag)
	if err != nil {
		log.Println("[E] Error" + err.Error())
	}
	dict = *dictFlag
	// /FLAGS

	// BANNER
	fmt.Println("[/!\\] Starting bruteforcing with dict: \n\t" + dict + "\n   On site: " + targetUrl + "\n")
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
	var errRedirect = errors.New("no_redirect")
	if err != nil && err != errRedirect {
		log.Println("[Request error] %s", err)
		os.Exit(1)
	}
	switch {
	case response.StatusCode == 404:
		fmt.Printf("\r                                ")
		fmt.Printf("\r %s", path)
	case response.StatusCode == 200:
		if isListable(content) {
			fmt.Println("\r\033[32m\033[1m[!]\033[0m Path: " + path + " is listable.")
			// Add directory to the found list
			found_dir = append(found_dir, path+" (LISTABLE)")
		} else {

			fmt.Printf("\r                                ")
			fmt.Println("\r\033[32m\033[1m[!]\033[0m 200 on: " + path + " - Size: " + strconv.Itoa(len(content)))
			found_files = append(found_files, path+" - Size: "+strconv.Itoa(len(content)))
		}
	case response.StatusCode >= 300 && response.StatusCode <= 399:
		// TODO
		// Fix print race-condition problems
		fmt.Printf("\r                                ")
		fmt.Println("\r\033[32m\033[1m[!]\033[0m Response.Statuscode == " + strconv.Itoa(response.StatusCode) + " on " + path + " - Size: " + strconv.Itoa(len(content)))
		if isDirectory(response, path) {
			fmt.Println("\r   [i] " + path + " is a directory.")
			// Call to httpRequest with following redirects
			response, content, _ := httpRequest(targetUrl, path, true)
			if isListable(content) {
				fmt.Println("\r    └[i] Directory: " + path + " is listable. It won't be added to queue.")
				found_dir = append(found_dir, path+"(LISTABLE)")
			} else {
				fmt.Println("\r    └[i] Directory: " + path + " ended with a " + strconv.Itoa(response.StatusCode) + " It will be added to queue.")
				feed(task_queue, path+"/")
				found_dir = append(found_dir, path)
			}
		} else {
			fmt.Println("\r    [i] " + path + " is NOT a directory.")
			found_files = append(found_files, path+" - Size: "+strconv.Itoa(len(content)))
		}
	case response.StatusCode == 403:
		fmt.Printf("\r                                ")
		fmt.Println("\r\033[33m\033[1m[!]\033[0m 403 on " + path + " - Size: " + strconv.Itoa(len(content)))
		found_files = append(found_files, path+" - Size: "+strconv.Itoa(len(content)))
	case response.StatusCode == 405:
		fmt.Printf("\r                                ")
		fmt.Println("\r\033[34m\033[1m[!]\033[0m 405 on " + path + " - Size: " + strconv.Itoa(len(content)))
		found_files = append(found_files, path+" - Size: "+strconv.Itoa(len(content)))
	}

	finished_tasks = append(finished_tasks, <-simul)
	defer wg.Done()
}
