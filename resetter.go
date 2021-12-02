package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
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

	reindexCnf := make(map[string]interface{})
	reindexCnf["source"] = make(map[string]interface{})
	reindexCnf["source"].(map[string]interface{})["index"] = "test"
	reindexCnf["dest"] = make(map[string]interface{})
	reindexCnf["dest"].(map[string]interface{})["index"] = "test-new"

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(reindexCnf); err != nil {
		log.Fatalf("Error encoding reindexCnf: %s", err)
	}

	res, err := es.Reindex(
		strings.NewReader(buf.String()),
		es.Reindex.WithPretty(),
	)

	if err != nil {
		log.Fatalf("reindex error : %s", err)
	} else {
		log.Println(res)
	}
	res.Body.Close()
	res, err = es.Indices.Delete(
		[]string{"test"},
	)
	res.Body.Close()
	// change back wit hnew settings
	reindexCnf["source"].(map[string]interface{})["index"] = "test-new"
	reindexCnf["dest"].(map[string]interface{})["index"] = "test"

	var buf2 bytes.Buffer
	if err := json.NewEncoder(&buf2).Encode(reindexCnf); err != nil {
		log.Fatalf("Error encoding reindexCnf: %s", err)
	}

	res, err = es.Reindex(
		strings.NewReader(buf2.String()),
		es.Reindex.WithPretty(),
	)

	if err != nil {
		log.Fatalf("reindex error : %s", err)
	} else {
		log.Println(res)
	}

	res, err = es.Indices.Delete(
		[]string{"test-new"},
	)

	log.Println(strings.Repeat("-", 37))
}
