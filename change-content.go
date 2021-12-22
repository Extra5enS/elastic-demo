
package main

import (
	"bytes"
	"os"
	"bufio"
	"context"
	"crypto/tls"
	"encoding/json"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"
	"fmt"

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


	log.Println("Enter ID of doc you want to change")
	log.Println(strings.Repeat("-", 37))
	var docID string
	fmt.Scanf("%s\n", &docID)
	// Find _doc by id
	query := map[string]interface{} {
		""
	}

	log.Println(strings.Repeat("-", 37))
	log.Println("Enter <field> : <value>")
	log.Println(strings.Repeat("-", 37))

	request := make(map[string]string)
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		// simple_query = [<field_name>, <:>, <field_value>]
		string_query := strings.Fields(scanner.Text())
		// Build the request body.
		if len(string_query) != 3 || (string_query[1] != ":" && string_query[1] != "=") {
			log.Println("Wrong query")
			continue
		}
		request[string_query[0]] = string_query[2]
	}
	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(request); err != nil {
		log.Fatalf("Error encoding request: %s", err)
	}

	// Set up the request object.
	req := esapi.IndexRequest{
		Index:      "test",
		DocumentID: docID,
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
		log.Printf("[%s] Error indexing document ID=%s", res.Status(), docID)
	} else {
		// Deserialize the response into a map.
		var r map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
			log.Printf("Error parsing the response body: %s", err)
		} else {
			// Print the response status and indexed document version.
			log.Printf("{ID=%s} [%s] %s; version=%d", docID, res.Status(), r["result"], int(r["_version"].(float64)))
		}
	}
}
