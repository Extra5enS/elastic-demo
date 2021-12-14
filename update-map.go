package main

import (
	"bufio"
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
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

func main() {
	log.Println("I will fill your database with some data")
	log.Println(strings.Repeat("-", 37))
	// Create default Client
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

	file, err := os.Open("source/extra_data.txt")
	if err != nil {
		log.Fatalf("failed opening file: %s", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)

	if !scanner.Scan() {
		log.Fatalf("Empty file")
	}
	col_names := strings.Fields(scanner.Text())

	createCnf := make(map[string]interface{})
	createCnf["mappings"] = make(map[string]interface{})
	createCnf["mappings"].(map[string]interface{})["properties"] = make(map[string]interface{})
	createCnf["mappings"].(map[string]interface{})["properties"].(map[string]interface{})["arg_num"] = make(map[string]interface{})
	createCnf["mappings"].(map[string]interface{})["properties"].(map[string]interface{})["arg_num"].(map[string]interface{})["type"] = "long"

	var buf bytes.Buffer
	// we shouldn't write "mappings : {...}"
	if err := json.NewEncoder(&buf).Encode(createCnf["mappings"]); err != nil {
		log.Fatalf("Error encoding createCnf: %s", err)
	}
	// on main doc write wrong func args!!!!
	// https://pkg.go.dev/github.com/elastic/go-elasticsearch@v0.0.0/esapi#IndicesPutMapping
	// It hasn't WithIndex(...string) method

	res, err := es.Indices.PutMapping(
		[]string{"test"},
		strings.NewReader(buf.String()),
	)
	if err != nil {
		log.Fatalf("update mapping creation error: %s", err)
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

			var body bytes.Buffer
			request := make(map[string]string)
			// Build the request body.
			for i, field := range strings.Fields(line) {
				request[col_names[i]] = field
			}

			if err := json.NewEncoder(&body).Encode(request); err != nil {
				log.Fatalf("Error encoding request: %s", err)
			}
			// Set up the request object.
			req := esapi.IndexRequest{
				Index:      "test",
				DocumentID: strconv.Itoa(50 + i),
				Body:       strings.NewReader(body.String()),
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
					log.Printf("{i=%2d} [%s] %s; version=%d", i, res.Status(), r["result"], int(r["_version"].(float64)))
				}
			}
		}(i, line, col_names)
		i++
	}
	wg.Wait()

	log.Println(strings.Repeat("-", 37))

}
