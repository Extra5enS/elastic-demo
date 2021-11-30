package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
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
	log.Println(strings.Repeat("-", 37))

	// Close to change settings
	close_res, err := es.Indices.Close(
		[]string{"test"},
		es.Indices.Close.WithPretty(),
	)
	if err != nil {
		log.Fatalf("index close error: %s", err)
	} else {
		log.Println(close_res)
	}
	close_res.Body.Close()
	log.Println(strings.Repeat("-", 37))

	// Scan and Put new settings
	var number_of_replicas int
	fmt.Printf("Write numbre_of_replicas, that you want to set: ")
	fmt.Scanf("%d\n", &number_of_replicas)
	log.Println(strings.Repeat("-", 37))

	createCnf := make(map[string]interface{})
	createCnf["index"] = make(map[string]int)
	createCnf["index"].(map[string]int)["number_of_replicas"] = number_of_replicas
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(createCnf); err != nil {
		log.Fatalf("Error encoding createCnf: %s", err)
	}
	potset_res, err := es.Indices.PutSettings(
		strings.NewReader(buf.String()),
		es.Indices.PutSettings.WithIndex("test"),
		es.Indices.PutSettings.WithPretty(),
	)

	if err != nil {
		log.Fatalf("index close error: %s", err)
	} else {
		log.Println(potset_res)
	}
	potset_res.Body.Close()
	log.Println(strings.Repeat("-", 37))

	// Open to change settings
	open_res, err := es.Indices.Open(
		[]string{"test"},
		es.Indices.Open.WithPretty(),
	)
	if err != nil {
		log.Fatalf("indes open error: %s", err)
	} else {
		log.Println(open_res)
	}
	open_res.Body.Close()

	log.Println(strings.Repeat("-", 37))
}
