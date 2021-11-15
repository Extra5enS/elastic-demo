# elastic-demo
In this repository you can find demo programs for working with Elasticsearch server.

## Cluster start
First you need to raise the dockers server, this can be done using a makefile.

### Cluster with 3 nodes 
For cluster with 3 nodes we will uses ```docker-compose```, it's already written in Makefile, so you can use it. 
```
sudo make
```
To shutdown it also use make
```
sudo make clear
```
### Single node Cluster
For single node cluster we create network and up docker by hand. So you can use next expation:
```
sudo make docker-single-node
```
To shutdown it also use make
```
sudo make clear-single-node
```
## filler.go
Next, you need to fill the server with data. In this case, the server has a single data stream with the index "test". You can use this commend to do it.
```
go run filler.go
```
Here we use Dynamic Mapping, which defines fields and their types when added. It is important that he remembers the type of the field and when trying to change it, he may give an incorrect result or error.

## simple-search.go
The first example shows the easiest way to use search. You can **match** the lines with values that you need. This requests calculate  **_score** for every matched line, so the result is sorted begins with the most similar result.

Use comment to look at the result
```
go run simple-search.go
```

## filter-demo.go
Next example shows abilities  of **filter**. You can use **temp** to match lines. In this case **_score** isn't calculated so the result isn't sorted. Next you can use **range** to find lines this fields in some range. Special  words here are: **lt** (<), **gt** (>), **e** (=). **e** being written after **lt** and **gt** to create **lte** (<=) and **gte** (>=).

Use comment to look at the result
```
go run filter-demo.go
```
Always understand for what type you are filtering, if you are not sure, you can use keyword **format**

## sort-demo.go
Last example shows how we can **sort** results. After filter call result isn't sort, so we can fix this problem. **sort** define the order of sorts rolls. Result will be sorted by first condition. If the is equal values, requests will use next condition and so on. You can use keywords **asc** or **desc** to set ascending order or descending order.

Use comment to look at the result
```
go run sort-demo.go
```
**sort** throws an error if it has nothing to sort, that is, if nothing was found as a result of the search or filtering

## Links

[General documentation](https://www.elastic.co/guide/index.html)

[Golang ElasticSearch githab](https://github.com/elastic/go-elasticsearch) 

[Golang ElasticSearch pcg.go.dev](https://pkg.go.dev/github.com/elastic/go-easticsearch)



## Goodluck 
