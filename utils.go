package main

import (
	"crypto/tls"
	"errors"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
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
	rDir := regexp.MustCompile(".*" + regexp.QuoteMeta(path) + "/")
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

var errNoRedirect = errors.New("no_redirect")

func httpRequest(targetUrl string, path string, followRedirect bool) (response *http.Response, content []byte, err error) {

	// SET PROXY
	//fmt.Println("SET PROXY FORCED")
	//proxyUrl, err := url.Parse("http://localhost:8081")

	// DISABLING SSL CHECKS
	tr := &http.Transport{
		//Proxy:           http.ProxyURL(proxyUrl),
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	// Set timeout check
	timeout := time.Duration(60 * time.Second)
	// Create client with custom parameters
	client := &http.Client{Transport: tr, Timeout: timeout}

	// client := &http.Client{}
	// If its requested not to follow redirects
	if !followRedirect {
		client = &http.Client{
			Transport: tr,
			Timeout:   timeout,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				//var err error
				return errNoRedirect
			},
		}
	} else {
		client = &http.Client{Transport: tr, Timeout: timeout}
	}
	// ---

	// Perform HTTP request

	req, err := http.NewRequest("GET", targetUrl+path, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Windows x86) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/42.0.2311.135 Safari/537.36")
	response, err = client.Do(req)
	if err, ok := err.(net.Error); ok && err.Timeout() {
		return response, content, err
	}
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
