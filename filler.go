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

	file, err := os.Open("source/data.txt")
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
	for _, name := range col_names {
		createCnf["mappings"].(map[string]interface{})["properties"].(map[string]interface{})[name] = make(map[string]string)
		if name == "text" || name == "myID" {
			createCnf["mappings"].(map[string]interface{})["properties"].(map[string]interface{})[name].(map[string]string)["type"] = "integer"
		} else {
			createCnf["mappings"].(map[string]interface{})["properties"].(map[string]interface{})[name].(map[string]string)["type"] = "text"
		}
	}
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(createCnf); err != nil {
		log.Fatalf("Error encoding createCnf: %s", err)
	}

	res, err := es.Indices.Create(
		"test",
		es.Indices.Create.WithBody(strings.NewReader(buf.String())),
	)

	if err != nil {
		log.Fatalf("new mapping creation error: %s", err)
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
				DocumentID: strconv.Itoa(i),
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

/*
`{
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
			"number_of_shards": 2,
			"number_of_replicas": 1
		}
	}
}`*/
