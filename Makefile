all:
	sysctl -w vm.max_map_count=262144
	docker-compose up

docker-single-node:
	docker network create elastic
	docker pull docker.elastic.co/elasticsearch/elasticsearch:7.15.1
	docker run --name es01-test --net elastic -p 9200:9200 -p 9300:9300 -e "discovery.type=single-node" docker.elastic.co/elasticsearch/elasticsearch:7.15.1

clear:
	docker-compose down

clear-single-node:
	docker network rm elastic
	docker rm es01-test

# write some scripts
show-filter:
	echo "Scripts will fullfit db with data from `data.txt` and try to use filter with your data"
	echo "1) Let's fullfit bd"
	cat source/data.txt 
	go run deleter.go 
	go run filler.go 
	echo "2) Fillter will select your data for range or temp and try to find"
	go run filter-demo.go regexpencoder.go
