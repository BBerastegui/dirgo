package main

import (
	"crypto/tls"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strings"
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

func formatUrl(urlToParse string) (string, error) {
	u, err := url.Parse(urlToParse)
	if err != nil {
		return "", err
	}
	if u.Scheme == "" {
		u.Scheme = "http"
	}
	return strings.TrimRight(u.String(), "/") + "/", nil
}

func httpRequest(targetUrl string, path string, followRedirect bool) (response *http.Response, content []byte, err error) {

	// DISABLING SSL CHECKS

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	//	client := &http.Client{}
	// If its requested not to follow redirects
	if !followRedirect {
		client = &http.Client{
			Transport: tr,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				//var err error
				return errors.New("no_redirect")
			},
		}
	} else {
		client = &http.Client{Transport: tr}
	}
	// ---

	// Perform HTTP request

	req, err := http.NewRequest("GET", targetUrl+path, nil)
	response, err = client.Do(req)
	if err != nil {
		return response, content, err
	}
	defer response.Body.Close()
	content, err = ioutil.ReadAll(response.Body)
	if err != nil {
		// Return. Error on reading content.
		// We don't really care.
		return response, content, err
	}
	// Return. Everything went OK
	return response, content, err
}
