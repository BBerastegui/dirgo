package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
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

func main() {
	recursive = false
	go feed(task_queue, "")
	consume(task_queue, simul)
	wg.Wait()
	fmt.Println(len(finished_tasks), " finished tasks.")
}

// This function will feed the task channel
func feed(task_queue chan string, prefix string) {

	// First open file and for each line...
	file, err := os.Open(os.Args[2])
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
		go scan(path, simul)
	}
	fmt.Println("DONE")
}

func scan(path string, simul chan string) {
	wg.Add(1)
	// Foo demo func.
	// Perform HTTP request
	response, err := http.Get(os.Args[1] + path)
	if err != nil {
		// Something happened.
		log.Fatal("[Request error] %s", err)
		os.Exit(1)
	} else {
		defer response.Body.Close()
		content, err := ioutil.ReadAll(response.Body)
		if err != nil {
			log.Fatal("ERROR While reading content: %s", err)
		}
		// Handle HTTP status
		switch {
		case response.StatusCode == 404:
			fmt.Printf("\r                                ")
			fmt.Printf("\r %s", path)
		default:
			switch {
			case response.StatusCode == 200:
				if listable(content) {
					fmt.Println("Path: " + path + " is listable.")
				} else {
					fmt.Printf("\r                                ")
					fmt.Println("\rFound: " + path)
				}
			case response.StatusCode >= 300 && response.StatusCode <= 399:
				// TODO
				// If "xxx" is 301'ed, and then followed location is 404 == 404
				// If "xxx" is 301'ed, and then followed location is 200 == 200
				fmt.Printf("\r                                ")
				fmt.Println("\r30X on " + path)
			case response.StatusCode == 403:
				fmt.Printf("\r                                ")
				fmt.Println("\r403 on " + path)
			}
		}
	}
	//fmt.Println("Task ", path, " finished.")
	finished_tasks = append(finished_tasks, <-simul)
	wg.Done()
}
