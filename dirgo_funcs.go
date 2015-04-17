package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
)

func listable(content []byte) bool {
	rListable := regexp.MustCompile(".*Parent Directory.*|.*Directory listing.*|.*Up To .*|.*Al directorio pri.*")
	if len(rListable.FindString(string(content))) > 0 {
		return true
	} else {
		return false
	}
}

func noRedirect(req *http.Request, via []*http.Request) error {
	return errors.New("Don't redirect!")
}

func httpRequest(url string, path string, followRedirect bool) (response *http.Response, content []byte, err error) {

	client := &http.Client{}
	// If its requested not to follow redirects
	if followRedirect {
		client = &http.Client{
			CheckRedirect: noRedirect,
		}
	} else {
		client = &http.Client{}
	}
	// ---
	// Perform HTTP request
	//response, err = http.Get(url + path)
	req, err := http.NewRequest("GET", url+path, nil)
	response, err = client.Do(req)

	// TODO
	if err != nil {
		if response.StatusCode == 301 {
			fmt.Println("got redirect")
		} else {
			log.Fatal("HTTP request failed.")
		}
	}

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
