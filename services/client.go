package src

import (
	"log"
	"net/http"
	"net/http/cookiejar"
	"time"

	"golang.org/x/net/publicsuffix"
)

// return the client carrying the token
func NewClient() *http.Client {
	cookiejar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	if err != nil {
		log.Fatal(err)
	}

	client := &http.Client{
		Jar:     cookiejar,
		Timeout: time.Second * 30,
	}
	return client
}
