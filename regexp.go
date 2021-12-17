package main

import (
	"bytes"
	"context"
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
	log.SetFlags(0)

	var r map[string]interface{}

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

	// change settings of claster
	var clusterBuf bytes.Buffer
	if err := json.NewEncoder(&clusterBuf).Encode(
		map[string]interface{}{
			"transient": map[string]interface{}{
				"search.allow_expensive_queries": true,
			},
		},
	); err != nil {
		log.Fatalf("Error encoding clusterBuf: %s", err)
	}
	cluserSet, err := es.Cluster.PutSettings(
		strings.NewReader(clusterBuf.String()),
		es.Cluster.PutSettings.WithPretty(),
	)
	if err != nil {
		log.Fatalf("Cluster put settings error : %s", err)
	} else {
		log.Println(cluserSet)
	}
	cluserSet, err = es.Cluster.GetSettings(
		es.Cluster.GetSettings.WithPretty(),
	)
	if err != nil {
		log.Fatalf("Cluster put settings error : %s", err)
	} else {
		log.Println(cluserSet)
	}
	defer cluserSet.Body.Close()

	log.Println(`Write regular expression that you want to search:`)

	var regexp string
	fmt.Scanf("%s", &regexp)
	log.Println(`Template is:`, regexp)

	var regexpbuf bytes.Buffer
	if err := NewRegexEncode(&regexpbuf).Encode(strings.NewReader(regexp)); err != nil {
		log.Fatalf("Error regex encoder: %s", err)
	}
	log.Println(`For Elasticsaerch:`, regexpbuf.String())

	var buf bytes.Buffer
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"filter": []interface{}{ // if we want to list params, we will use list of maps
					map[string]interface{}{
						"regexp": map[string]interface{}{
							"title": regexpbuf.String(),
						},
					},
				},
			},
		},
	}

	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		log.Fatalf("Error encoding query: %s", err)
	}

	// Perform the search request.
	res, err := es.Search(
		es.Search.WithContext(context.Background()),
		es.Search.WithIndex("test"),
		es.Search.WithBody(&buf),
		es.Search.WithSize(20),
		es.Search.WithTrackTotalHits(true),
		//es.Search.WithAnalyzeWildcard(true),
		es.Search.WithPretty(),
	)
	if err != nil {
		log.Fatalf("Error getting response: %s", err)
	}
	defer res.Body.Close()

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

	log.Println(strings.Repeat("=", 37))
}
