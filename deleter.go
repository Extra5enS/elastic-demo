package main

import (
	"crypto/tls"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
)

func main() {
	log.Println("I will delete your database")
	log.Println(strings.Repeat("-", 37))
	// Create Client
	cfg := elasticsearch.Config{
		Addresses: []string{
			"http://localhost:9200",
			"http://localhost:9201",
		},
		Transport: &http.Transport{
			MaxIdleConnsPerHost:   10,
			ResponseHeaderTimeout: 5 * time.Second,
			DialContext:           (&net.Dialer{Timeout: 5 * time.Second}).DialContext,
			TLSClientConfig: &tls.Config{
				MaxVersion:         tls.VersionTLS11,
				InsecureSkipVerify: true,
			},
		},
	}
	es, err := elasticsearch.NewClient(cfg)

	if err != nil {
		log.Fatalf("Error creating the client: %s", err)
	} else {
		//log.Println(es.Info())
	}

	res, err := es.Indices.Delete(
		[]string{"test"},
		es.Indices.Delete.WithMasterTimeout(2*time.Second),
	)
	if err != nil {
		log.Fatalf("mapping delete error: %s", err)
	} else {
		if res.Status() == "404" {
			log.Println(`File doesn't exist or already deleted`);
		} else {
			log.Println(`File is deleted`);
		}
	}
	res.Body.Close()
	log.Println(strings.Repeat("-", 37))
}
