package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"sync"
	"time"
)

func random(min, max int) int {
	rand.Seed(time.Now().UnixNano())
	return rand.Intn(max-min) + min
}

// Channel of simultaneous tasks
var simul = make(chan string, 100)
var task_queue = make(chan string)

// Array of strings with finished tasks
var finished_tasks []string

// Array with directories found
var found_dir []string
var found_files []string

// Array with pending directories
var pending_dir []string

// Last word used
var lastword string

var wg sync.WaitGroup

var task_recursive string
var recursive bool

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
	recursive = false
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

	// Re-fill tasks
	if len(task_recursive) > 0 {
		recursive = true
	} else {
		close(task_queue)
		fmt.Println("FEED FINISHED")
	}
}

func consume(task_queue chan string, simul chan string) {
	for path := range task_queue {
		simul <- path
		go scan(url, path, simul)
	}
	fmt.Println("DONE")
}

func scan(url string, path string, simul chan string) {
	wg.Add(1)
	// Foo demo func.
	response, content, err := httpRequest(url, path, true)
	if err != nil {
		log.Fatal("[Request error] %s", err)
	}
	switch {
	case response.StatusCode == 404:
		fmt.Printf("\r                                ")
		fmt.Printf("\r %s", path)
	case response.StatusCode == 200:
		if listable(content) {
			fmt.Println("Path: " + path + " is listable.")
		} else {
			fmt.Printf("\r                                ")
			fmt.Println("\rFound: " + path)
		}
	case response.StatusCode >= 300 && response.StatusCode <= 399:
		fmt.Printf("\r                                ")
		fmt.Println("\rresponse.Statuscode == 30X on " + path)
	case response.StatusCode == 403:
		fmt.Printf("\r                                ")
		fmt.Println("\r403 on " + path)
	}
	finished_tasks = append(finished_tasks, <-simul)
	defer wg.Done()
}
