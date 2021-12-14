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
	log.Println("I will fill your database with some data")
	log.Println(strings.Repeat("-", 37))
	// Create default Client

	var (
		number_of_replicas, numbre_of_shards int
	)
	fmt.Printf("Write numbre_of_shards, that you want to set: ")
	fmt.Scanf("%d\n", &numbre_of_shards)
	fmt.Printf("Ok, you want to set: %d, interesting\n", numbre_of_shards)
	log.Println(strings.Repeat("-", 37))
	fmt.Printf("Write numbre_of_replicas, that you want to set: ")
	fmt.Scanf("%d\n", &number_of_replicas)
	fmt.Printf("Ok, you want to set: %d, interesting\n", number_of_replicas)
	log.Println(strings.Repeat("-", 37))

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

	createCnf := make(map[string]interface{})
	createCnf["index_patterns"] = make([]string, 1)
	createCnf["index_patterns"].([]string)[0] = "test*"

	createCnf["settings"] = make(map[string]interface{})
	createCnf["settings"].(map[string]interface{})["index"] = make(map[string]int)
	createCnf["settings"].(map[string]interface{})["index"].(map[string]int)["number_of_shards"] = numbre_of_shards
	createCnf["settings"].(map[string]interface{})["index"].(map[string]int)["number_of_replicas"] = number_of_replicas
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(createCnf); err != nil {
		log.Fatalf("Error encoding createCnf: %s", err)
	}

	putt_res, err := es.Indices.PutTemplate(
		"test-template",
		strings.NewReader(buf.String()),
		es.Indices.PutTemplate.WithOrder(1),
		//es.Indices.PutTemplate.WithCreate(true),
	)

	if err != nil {
		log.Fatalf("Putting new template error: %s", err)
	} else {
		log.Println(putt_res)
	}
	log.Println(strings.Repeat("-", 37))

	defer putt_res.Body.Close()

	gett_res, err := es.Indices.GetTemplate(
		es.Indices.GetTemplate.WithName("test-template"),
		es.Indices.GetTemplate.WithPretty(),
	)

	if err != nil {
		log.Fatalf("Getting template error: %s", err)
	} else {
		log.Print(gett_res)
	}
	gett_res.Body.Close()
	log.Println(strings.Repeat("-", 37))

}
