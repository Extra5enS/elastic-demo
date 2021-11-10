all: default-server-start golang-packages
	go run filler.go

golang-packages:
	go get -u github.com/elastic/go-elasticsearch 

default-server-start:
	docker network create elastic
	docker pull docker.elastic.co/elasticsearch/elasticsearch:7.15.1
	docker run --name es01-test --net elastic -p 9200:9200 -p 9300:9300 -e "discovery.type=single-node" docker.elastic.co/elasticsearch/elasticsearch:7.15.1
