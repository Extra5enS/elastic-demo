package main

import (
	"log"
	"strings"

	"github.com/elastic/go-elasticsearch/v8"
)

func main() {
	log.Println("I will fill your database with some data")
	log.Println(strings.Repeat("-", 37))
	// Create default Client
	es, err := elasticsearch.NewDefaultClient()
	if err != nil {
		log.Fatalf("Error creating the client: %s", err)
	}

	res, err := es.Indices.Delete(
		[]string{"test"},
	)
	if err != nil {
		log.Printf("mapping delete error: %s", err)
	} else {
		log.Println(res)
	}
}
