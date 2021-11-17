package main

import (
	"bufio"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
)

func isNumeric(s string) bool {
	_, err := strconv.ParseFloat(s, 64)
	return err == nil
}

func main() {

	log.Println("I will fill your database with some data")
	log.Println(strings.Repeat("-", 37))
	// Create default Client
	cfg := elasticsearch.Config{
		Addresses: []string{
			"http://localhost:9200",
			//"http://localhost:9201",
			//"http://localhost:9202",
		},
		Transport: &http.Transport{
			MaxIdleConnsPerHost:   10,
			ResponseHeaderTimeout: time.Second,
			DialContext:           (&net.Dialer{Timeout: time.Second}).DialContext,
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

	file, err := os.Open("source/data.txt")
	if err != nil {
		log.Fatalf("failed opening file: %s", err)
	}

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)

	if !scanner.Scan() {
		log.Fatalf("Empty file")
	}
	col_names := strings.Fields(scanner.Text())

	res, err := es.Indices.Create(
		"test",
		es.Indices.Create.WithBody(strings.NewReader(`{
			"mappings": {
				"properties": {
					"title": {
						"type": "text"
					},
					"text": {
						"type": "integer"
					},       
					"myID": {
						"type": "integer"
					},    
					"word": {
						"type": "text"
					}
				}
			},
			"settings": {
				"index": {
					"number_of_shards": 3,  
					"number_of_replicas": 0 
				}
			}
		}`,
		)), // "number_of_replicas" should be equil to 2, use 1 only for experiment
	)

	if err != nil {
		log.Printf("new mapping creation error: %s", err)
	} else {
		log.Println(res)
	}
	log.Println(strings.Repeat("-", 37))

	defer res.Body.Close()
	// Fill database
	i := 1
	var wg sync.WaitGroup
	for scanner.Scan() {
		wg.Add(1)
		line := scanner.Text()
		go func(i int, line string, col_names []string) {
			defer wg.Done()

			// Build the request body.
			body := `{`
			for i, field := range strings.Fields(line) {
				if i != 0 {
					body = body + `, `
				}
				if isNumeric(field) {
					body = body + fmt.Sprintf(`"%s":%s`, col_names[i], field)
				} else {
					body = body + fmt.Sprintf(`"%s":"%s"`, col_names[i], field)
				}
			}
			body = body + `}`

			// Set up the request object.
			req := esapi.IndexRequest{
				Index:      "test",
				DocumentID: strconv.Itoa(i),
				Body:       strings.NewReader(body),
				Refresh:    "true",
			}

			// Perform the request with the client.
			res, err := req.Do(context.Background(), es)
			if err != nil {
				log.Fatalf("Error getting response: %s", err)
			}
			defer res.Body.Close()

			if res.IsError() {
				log.Printf("[%s] Error indexing document ID=%d", res.Status(), i+1)
			} else {
				// Deserialize the response into a map.
				var r map[string]interface{}
				if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
					log.Printf("Error parsing the response body: %s", err)
				} else {
					// Print the response status and indexed document version.
					log.Printf("{i=%d} [%s] %s; version=%d", i, res.Status(), r["result"], int(r["_version"].(float64)))
				}
			}
		}(i, line, col_names)
		i++
	}
	wg.Wait()

	log.Println(strings.Repeat("-", 37))

}
