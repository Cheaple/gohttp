# README

## Server

Run server using Docker image:

```shell
docker pull chenyule/http-server
docker run -d -p <port number>:8080 chenyule/http-server
```

Build & Run server from source code:

```shell
go build -o ./server/main ./server/server.go 
./server/main -p=<port number, default as 80>
```

Test GET command:

```shell
curl -X GET 16.170.214.245:8080/test.txt
curl -X GET 16.170.214.245:8080/no.txt		# 404 not found
curl -X GET 16.170.214.245:8080/test		# 500 invalid file type
```

or [16.170.214.245:8080/test.jpg ](http://16.170.214.245:8080/test.jpg) (download an image)

Test POST command: 

+ Use Postman

Test concurrency:

```shell
for i in {1..10}; do curl -X GET 16.170.214.245:8080/test.txt; done
```

## Proxy

Run proxy locally:

```shell
go build -o ./proxy/main ./proxy/proxy.go 
./proxy/main -p=8999
```

Test GET command:

```shell
curl -X GET 16.170.214.245:8080/test.jpg -x localhost:8999
```

Test POST command:

```shell
curl -X POST 16.170.214.245:8080/test.jpg -x localhost:8999  # 501 Method not allowed
```

Test proxy concurrency:

```shell
for i in {1..10}; do curl -X GET 16.170.214.245:8080/test.txt -x localhost:8999; done
```

