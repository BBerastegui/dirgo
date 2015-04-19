package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"sync"
	"time"
)

func random(min, max int) int {
	rand.Seed(time.Now().UnixNano())
	return rand.Intn(max-min) + min
}

// Channel of simultaneous tasks
var simul = make(chan string, 10)
var task_queue = make(chan string)

// Array of strings with finished tasks
var finished_tasks []string

// Array with directories found
var found_dir []string
var found_files []string

// Array with pending directories
var pending_dir []string

// TODO
// Persistence variables
// Last word used
var lastword string

var wg sync.WaitGroup

var url string
var dict string

func main() {

	// FLAGS
	urlFlag := flag.String("u", "foo", "the url to test")
	dictFlag := flag.String("d", "foo", "the dictionary to use")
	flag.Parse()

	url = *urlFlag
	dict = *dictFlag
	// /FLAGS

	// ---
	go feed(task_queue, "")
	consume(task_queue, simul)
	wg.Wait()
	fmt.Println(len(finished_tasks), " finished tasks.")
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
		fmt.Println("\n\r [i] Waiting for tasks to end and close the task_queue.")
		wg.Wait()
		close(task_queue)
	}
	fmt.Println("\n\r [i] Task feed finished for prefix: " + prefix)
}

func consume(task_queue chan string, simul chan string) {
	for path := range task_queue {
		simul <- path
		go scan(url, path, simul)
	}
	// All tasks consumed
}

func scan(url string, path string, simul chan string) {
	wg.Add(1)
	response, content, _ := httpRequest(url, path, false)
	/*	if err != nil {
			log.Print("[Request error] %s", err)
		}
	*/
	switch {
	case response.StatusCode == 404:
		fmt.Printf("\r                                ")
		fmt.Printf("\r %s", path)
	case response.StatusCode == 200:
		if isListable(content) {
			fmt.Println("\r[!] Path: " + path + " is listable.")
		} else {
			fmt.Printf("\r                                ")
			fmt.Println("\r[!] 200 on: " + path + " - Size: " + strconv.Itoa(len(content)))
		}
	case response.StatusCode >= 300 && response.StatusCode <= 399:
		// TODO
		// Fix print race-condition problems
		fmt.Printf("\r                                ")
		fmt.Println("\r[!] Response.Statuscode == " + strconv.Itoa(response.StatusCode) + " on " + path + " - Size: " + strconv.Itoa(len(content)))
		if isDirectory(response, path) {
			fmt.Println("    [i] " + path + " is a directory.")
			// Call to httpRequest with following redirects
			response, content, _ := httpRequest(url, path, true)
			if isListable(content) {
				fmt.Println("\r    └[i] Directory: " + path + " is listable. It won't be added to queue.")
			} else {
				fmt.Println("\r    └[i] Directory: " + path + " ended with a " + strconv.Itoa(response.StatusCode) + " It will be added to queue.")
				feed(task_queue, path+"/")
			}
		} else {
			fmt.Println("    [i] " + path + " is NOT a directory.")
		}
	case response.StatusCode == 403:
		fmt.Printf("\r                                ")
		fmt.Println("\r[!] 403 on " + path + " - Size: " + strconv.Itoa(len(content)))
	case response.StatusCode == 405:
		fmt.Printf("\r                                ")
		fmt.Println("\r[!] 405 on " + path + " - Size: " + strconv.Itoa(len(content)))
	}

	finished_tasks = append(finished_tasks, <-simul)
	defer wg.Done()
}
