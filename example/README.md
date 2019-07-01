# ginopentracing

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A simple implementation of the api gateway, server1, server2 and server3 examples.

To build:

```shell
go build api_gateway.go
go build server1.go
go build server2.go
go build server3.go
```

To run, open four terminals, and execute the following:

```shell
./api_gateway
./server1
./server2
./server3
```

To test:

```shell
curl -X POST http://localhost:8000/service1 -v
curl -X POST http://localhost:8000/service2 -v
```

Header information is printed to stdout. You should see headers propagated from service to service.

On the API gateway:

```
Incoming Headers
User-Agent: [curl/7.47.0]
Accept: [*/*]
```

On service1:

```
Incoming headers
User-Agent: [Go-http-client/1.1]
Content-Length: [0]
X-B3-Sampled: [1]
X-B3-Spanid: [65025274cfd25c6b]
X-B3-Traceid: [65025274cfd25c6b]
Accept-Encoding: [gzip]
```

On service3:

```
Incoming Headers
X-B3-Spanid: [aa66150e951c54c]
X-B3-Traceid: [10386b198f22ca04]
Accept-Encoding: [gzip]
User-Agent: [Go-http-client/1.1]
Content-Length: [0]
X-B3-Parentspanid: [10386b198f22ca04]
X-B3-Sampled: [1]
```
