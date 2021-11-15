all:
	docker network create elastic
	docker pull docker.elastic.co/elasticsearch/elasticsearch:7.15.1
	docker run --name es01-test --net elastic -p 9200:9200 -p 9300:9300 -e "discovery.type=single-node" docker.elastic.co/elasticsearch/elasticsearch:7.15.1

# clear
