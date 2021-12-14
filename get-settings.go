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
	// Lest's scan settings frin
	getset_res, err := es.Indices.GetSettings(
		es.Indices.GetSettings.WithIndex("test"),
		es.Indices.GetSettings.WithPretty(),
	)
	if err != nil {
		log.Fatalf("index get setting error: %s", err)
	} else {
		log.Println(getset_res)
	}
	getset_res.Body.Close()
	mapset_res, err := es.Indices.GetMapping(
		es.Indices.GetMapping.WithIndex("test"),
		es.Indices.GetMapping.WithPretty(),
	)
	if err != nil {
		log.Fatalf("index get setting error: %s", err)
	} else {
		log.Println(mapset_res)
	}
	mapset_res.Body.Close()
	log.Println(strings.Repeat("-", 37))
}
