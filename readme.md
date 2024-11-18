## Simple queue system

This project uses a very easy http based queue system to transfer files line by line

Requirements:
- Use only stdlib

This project is composed by a client and a server


### How to run queue server

    make server
    ./server --port=8080

You can check other parameters such as long-polling-timeout with `./server --help`

API is pretty simple:

    POST /push/<queue-name>
    GET /push/<queue-name>

Now you can push or pull data from the queue using utilities like curl

    curl -X POST -d "my fancy data" http://localhost:8080/push/myqueue
    curl http://localhost:8080/pull/myqueue

Pull implements a configurable long polling timeout to avoid flooding server with requests.  

### How to use client

    make client

Push a file line by line to the queue

    ./client write <url> <queue-name> <filename>

Read from the queue and write to a file

    ./client read <url> <queue-name> <filename>

Take into account that client does not know when the queue is empty, so we used a strategy to have maximum 3 retries. This might be configured in a future verison.

A small bash script is provided to generate some files with correlative data

    ./generate.sh 100 > my-file-with-100-lines.txt

### Ready to run example

    ./server --port=8080 --long-polling-timeout=5
    ./client read http://localhost:8080 myqueue mycopy.txt
    ./client write http://localhost:8080 myqueue myfile.txt

### Design decisions

Simple HTTP protocol has been selected as network protocol in favor of regular TCP for those reasons:

- Golang Stdlib provides a very good http client and server library which made implementation faster
- Lots of new features can be implemented easily in top of HTTP using Header and Path semantics
- Very easy to add authentication using (using for example a middleware)
- It is much more cheap to protect against DDOS attacks using HTTP rather than raw TCP. Mainly because existing WAF offerings such as Cloudflare.
- Easier to debug (curl, http clients, etc ...)
- Easier to implement clients in other languages

There was an internal debate wether websockets library was part of standard library (https://pkg.go.dev/golang.org/x/net/websocket), but since the official net package recommends another package we dediced to do not implement websocktes.

Websockets would have added less load to the server, because there would be just a single connection for each file, but then implementation complexity would have been increased too. And we are optimizing for speed here.

As an alternative to websockets to lower to number of connections to the server (specially when pulling data) a long polling strategy has been implemented.

Also, we can extend the functionality enabling a multiline body such as a batch operation. This batch operation might be also implemented when reading, but then we should be able to read unless there is at least a batch number of elements in the queue.

### Scalability

Currently, the queues are implemented in memory, that means that it would not be possible to run more than one server. To scale this simple server horizontally we might apply to strategies:

a) Shared memory: Using an external server such as redis and change the in memory queue for a redis based queue.
b) Path/Queue based sharding: using a custom HAProxy based we can use a path prefix or a sharding operation over queue name to distribute load among servers. Each queue should reside only in a server. That means that in case a server goes down, data is lost.

