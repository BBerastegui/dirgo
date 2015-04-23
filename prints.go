package main

import (
	"fmt"
	"net/http"
	"strconv"
)

func print200(path string, contentLen string) {
	fmt.Printf("\r                                ")
	fmt.Println("\r\033[32m\033[1m[!]\033[0m 200 on: " + path + " - Size: " + contentLen)
	return
}

func print30x(response *http.Response, content []byte, path string, statusCode int) {
	if isDirectory(response, path) {
		printDirectory(path)
		// Call to httpRequest with following redirects
		response, content, _ := httpRequest(targetUrl, path, true)
		if isListable(content) {
			fmt.Println("\r    └[i] Directory: " + path + " is listable. It won't be added to queue.")
			found_dir = append(found_dir, path+"(LISTABLE)")
		} else {
			fmt.Println("\r    └[i] Directory: " + path + " ended with a " + strconv.Itoa(response.StatusCode) + ". It will be added to queue.")
			feed(task_queue, path+"/")
			found_dir = append(found_dir, path)
		}
	} else {
		fmt.Println("\r    [i] " + path + " is NOT a directory.")
		found_files = append(found_files, path+" - Size: "+strconv.Itoa(len(content)))
	}
	fmt.Printf("\r                                ")
	fmt.Println("\r\033[32m\033[1m[!]\033[0m Response.Statuscode == " + strconv.Itoa(response.StatusCode) + " on " + path + " - Size: " + strconv.Itoa(len(content)))
	return
}

func print403(path string, contentLen int) {
	fmt.Printf("\r                                ")
	fmt.Println("\r\033[33m\033[1m[!]\033[0m 403 on " + path + " - Size: " + strconv.Itoa(contentLen))
}

func print405(path string, contentLen int) {
	fmt.Printf("\r                                ")
	fmt.Println("\r\033[34m\033[1m[!]\033[0m 405 on " + path + " - Size: " + strconv.Itoa(contentLen))
}

func print404(path string) {
	fmt.Printf("\r                                ")
	fmt.Printf("\r %s", path)
	return
}

func printDirectory(path string) {
	fmt.Println("\r   [i] " + path + " is a directory.")
	return
}

func printListable(path string) {
	fmt.Println("\r\033[32m\033[1m[!]\033[0m Path: " + path + " is listable.")
	return
}
