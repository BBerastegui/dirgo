package main

import (
	"errors"
	"io/ioutil"
	"net/http"
	"regexp"
)

func isListable(content []byte) bool {
	rListable := regexp.MustCompile(".*Parent Directory.*|.*Directory listing.*|.*Up To .*|.*Al directorio pri.*|.*<title>Index of .*")
	if len(rListable.FindString(string(content))) > 0 {
		return true
	} else {
		return false
	}
}

func isDirectory(response *http.Response, path string) bool {
	rDir := regexp.MustCompile(".*" + path + "/")
	if len(rDir.FindString(string(response.Header["Location"][0]))) > 0 {
		// TODO
		// Fix location URL-encoded !!!
		// Example: http://testaspnet.vulnweb.com/jscripts/tiny_mce
		return true
	} else {
		return false
	}
}

func httpRequest(url string, path string, followRedirect bool) (response *http.Response, content []byte, err error) {
	client := &http.Client{}
	// If its requested not to follow redirects
	if !followRedirect {
		client = &http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return errors.New("Don't redirect!")
			},
		}
	} else {
		client = &http.Client{}
	}
	// ---
	// Perform HTTP request
	req, err := http.NewRequest("GET", url+path, nil)
	response, err = client.Do(req)

	// TODO
	/*if err != nil {
		if response.StatusCode == 301 {
			fmt.Println("\n[i] Got 301 on " + path)
			content = []byte("")
			return response, content, err
		} else {
			// Not a 301, re-perform request with follow redirects.
			//log.Printf("HTTP request failed.")
			client = &http.Client{}
			req, err := http.NewRequest("GET", url+path, nil)
			if err != nil {
				fmt.Println("\n Error performing request:")
			}
			response, err = client.Do(req)
			if err != nil {
				fmt.Println("\n Error doing request:")
			}
		}
	}
	*/
	defer response.Body.Close()
	content, err = ioutil.ReadAll(response.Body)
	if err != nil {
		// Return. Error on reading content.
		return response, content, err
	}
	// Return. Everything went OK
	return response, content, err

	/*
		if err != nil {
			// Return. Error when performing request.
			return response, content, err
		} else {
			defer response.Body.Close()
			content, err = ioutil.ReadAll(response.Body)
			if err != nil {
				// Return. Error on reading content.
				return response, content, err
			}
			// Return. Everything went OK
			return response, content, err
		}
	*/
}
