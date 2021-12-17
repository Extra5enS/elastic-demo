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
		// simple_query = ["field_name", "operatoin", "field_value"]
		simple_string_query := strings.Fields(scanner.Text())
		simple_query := make(map[string]interface{})
		var err error
		err = nil
		switch {
		case simple_string_query[1] == "=":
			switch {
			case isAccepterRegexp(simple_string_query[2]):
				simple_query["regexp"] = make(map[string]interface{})
				var regexpbuf bytes.Buffer
				if err := NewRegexEncode(&regexpbuf).Encode(strings.NewReader(simple_string_query[2])); err != nil {
					log.Fatalf("Error regex encoder: %s", err)
				}
				simple_query["regexp"].(map[string]interface{})[simple_string_query[0]] = regexpbuf.String()

			default:
				simple_query["temp"] = make(map[string]interface{})
				simple_query["temp"].(map[string]interface{})[simple_string_query[0]] = simple_string_query[2]
			}
		default:
			simple_query["range"] = make(map[string]interface{})
			simple_query["range"].(map[string]interface{})[simple_string_query[0]] = make(map[string]string)
			switch simple_string_query[1] {
			case "<":
				simple_query["range"].(map[string]interface{})[simple_string_query[0]].(map[string]string)["lt"] = simple_string_query[2]
			case "<=":
				simple_query["range"].(map[string]interface{})[simple_string_query[0]].(map[string]string)["lte"] = simple_string_query[2]
			case ">":
				simple_query["range"].(map[string]interface{})[simple_string_query[0]].(map[string]string)["gt"] = simple_string_query[2]
			case ">=":
				simple_query["range"].(map[string]interface{})[simple_string_query[0]].(map[string]string)["gte"] = simple_string_query[2]
			default:
				err = fmt.Errorf(`Unknowen operation : "%s"`, simple_string_query[1])
			}
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

/*
map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"filter": []interface{}{ // if we want to list params, we will use list of maps
					map[string]interface{}{
						"range": map[string]interface{}{
							"myID": map[string]interface{}{
								"gte": "5",
							},
						},
					},
					map[string]interface{}{
						"range": map[string]interface{}{
							"myID": map[string]interface{}{
								"lte": "10",
							},
						},
					},
						map[string]interface{}{
							"term": map[string]interface{}{
								"title": "test",
							},
						},
					},
				},
			},
		}
*/
