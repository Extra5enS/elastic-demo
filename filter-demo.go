package main

import (
	"bufio"
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
)

func main() {
	log.SetFlags(0)

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

	// Search for the indexed documents
	//
	// Build the request body.
	var buf bytes.Buffer
	log.Println(strings.Repeat("-", 37))
	log.Println("Enter you query step by step")
	log.Println(strings.Repeat("-", 37))

	query := NewQuery()

	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		log.Fatalf("Error encoding query: %s", err)
	}

	// Perform the search request.
	res, err := es.Search(
		es.Search.WithContext(context.Background()),
		es.Search.WithIndex("test"),
		es.Search.WithBody(&buf),
		es.Search.WithTrackTotalHits(true),
		es.Search.WithPretty(),
	)
	if err != nil {
		log.Fatalf("Error getting response: %s", err)
	}
	defer res.Body.Close()
	printData(res)
	log.Println(strings.Repeat("=", 37))
}

/*
	Look at data in res
*/
func printData(res *esapi.Response) int {
	if res.IsError() {
		var e map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&e); err != nil {
			log.Fatalf("Error parsing the response body: %s", err)
		} else {
			// Print the response status and error information.
			log.Fatalf("[%s] %s: %s",
				res.Status(),
				e["error"].(map[string]interface{})["type"],
				e["error"].(map[string]interface{})["reason"],
			)
		}
	}

	var r map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
		log.Fatalf("Error parsing the response body: %s", err)
	}
	// Print the response status, number of results, and request duration.
	log.Printf(
		"[%s] %d hits; took: %dms",
		res.Status(),
		int(r["hits"].(map[string]interface{})["total"].(map[string]interface{})["value"].(float64)),
		int(r["took"].(float64)),
	)
	// Print the ID and document source for each hit.
	for _, hit := range r["hits"].(map[string]interface{})["hits"].([]interface{}) {
		log.Printf(" * ID=%s, %s", hit.(map[string]interface{})["_id"], hit.(map[string]interface{})["_source"])
	}

	return int(r["hits"].(map[string]interface{})["total"].(map[string]interface{})["value"].(float64))
}

/*
	Scan and create query
*/
func NewQuery() map[string]interface{} {
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string][]interface{}{
				"filter": []interface{}{},
			},
		},
	}

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		// simple_query = [<field_name>, <operatoin>, <field_value>]
		string_query := strings.Fields(scanner.Text())
		if len(string_query) != 3 {
			log.Println("Uncorrect query, please, try to write something another")
			continue
		}
		// simple_query uses to add new small query to "filter"
		simple_query := make(map[string]interface{})
		// place main information for filter about single field
		field_setting := make(map[string]interface{})
		var err error
		err = nil
		switch string_query[1] {
		case "=":
			switch {
			case isAccepterRegexp(string_query[2]):
				// it's regex, so we must ecode it in vertion for Elasticsearch
				var regexpbuf bytes.Buffer
				if err := NewRegexEncode(&regexpbuf).Encode(strings.NewReader(string_query[2])); err != nil {
					log.Fatalf("Error regex encoder: %s", err)
				}
				field_setting[string_query[0]] = regexpbuf.String()
				simple_query["regexp"] = field_setting
			default:
				// it's commone value of field
				simple_query["term"] = make(map[string]interface{})
				field_setting[string_query[0]] = string_query[2]
				simple_query["term"] = field_setting
			}
		default: // for "range" query
			range_setting := make(map[string]string)
			switch string_query[1] {
			case "<":
				range_setting["lt"] = string_query[2]
			case "<=":
				range_setting["lte"] = string_query[2]
			case ">":
				range_setting["gt"] = string_query[2]
			case ">=":
				range_setting["gte"] = string_query[2]
			default:
				err = fmt.Errorf(`Unknowen operation : "%s"`, string_query[1])
			}
			field_setting[string_query[0]] = range_setting
			simple_query["range"] = field_setting
		}
		if err != nil {
			log.Println(err.Error())
			err = nil
		} else {
			query["query"].(map[string]interface{})["bool"].(map[string][]interface{})["filter"] =
				append(query["query"].(map[string]interface{})["bool"].(map[string][]interface{})["filter"], simple_query)
		}
	}
	//log.Printf("%s", query)
	return query
}
