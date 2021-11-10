package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
)

func main() {

	log.Println("I will fill your database with some data")
	log.Println(strings.Repeat("-", 37))
	// Create default Client
	es, err := elasticsearch.NewDefaultClient()
	if err != nil {
		log.Fatalf("Error creating the client: %s", err)
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
				body = body + fmt.Sprintf(`"%s":"%s"`, col_names[i], field)
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
					log.Printf("[%s] %s; version=%d", res.Status(), r["result"], int(r["_version"].(float64)))
				}
			}
		}(i, line, col_names)
		i++
	}
	wg.Wait()

	log.Println(strings.Repeat("-", 37))

}
