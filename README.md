# elastic-demo
In this repository you can find demo programs for working with Elasticsearch server.

## Server start
First you need to raise the dockers server, this can be done using a makefile.
```
sudo make default-server-start
```

## filler.go
Next, you need to fill the server with data. In this case, the server has a single data stream with the index "test". You can use make to do it.
```
> make all
```
Or you can use golang program.
```
> go run filler.go
```
*The first two actions can be combined with a single make call.*

## simple-search.go
The Fisrt example shows the easiest way to use search. You can **match** the lines with values that you need. This requests calculate  **_score** for every matched line, so the result is sorted begins with the most similar result.

Use comment to look at the result
```
> go run simple-search.go
```

## filter-demo.go
Next example shows abilitis of **filter**. You can use **temp** to match lines. In this case **_score** isn't calculated so the result isn't sorted. Next you can use **range** to find lines this fileds in some range. Speciale words here are: **lt**(<), **gt**(>), **e**(=). **e** being written after **lt** and **gt** to create **lte**(<=) and **gte** (>=).

Use comment to look at the result
```
> go run filter-demo.go
```

## sort-demo.go
Last example shows how we can **srote** results. After **filter** call result isn't sort, so we can fix this problem. **sort** define the order of sort's rolls. Result will be sorted by first condiotion. If the is equle values, requets will use next condiotion and so on. You can use keywords **asc** or **desc** to set ascending order or descending order.

Use comment to look at the result
```
> go run sort-demo.go
```

## Goodluck 