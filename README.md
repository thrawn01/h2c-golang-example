## HTTP/2 Cleartext (H2C) golang example

Since the internet failed me, and the only workable example of a H2C client I
can find was in the actual go code test suite I'm going to lay out what I
discovered about H2C support in golang here.

First is that the standard golang code supports HTTP2 but does not
directly support H2C. H2C support only exists in the `golang.org/x/net/http2/h2c`
package. You can make your HTTP server H2C capable by wrapping your handler or
mux with `h2c.NewHandler()` like so.

```go
h2s := &http2.Server{}

handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "Hello, %v, http: %v", r.URL.Path, r.TLS == nil)
})

server := &http.Server{
    Addr:    "0.0.0.0:1010",
    Handler: h2c.NewHandler(handler, h2s),
}

fmt.Printf("Listening [0.0.0.0:1010]...\n")
checkErr(server.ListenAndServe(), "while listening")
```
The above code allows the server to support
[H2C upgrade](https://httpwg.org/specs/rfc7540.html#discover-http) and
[H2C prior knowledge](https://httpwg.org/specs/rfc7540.html#known-http) along with
 standard HTTP/2 and HTTP/1.1 that golang natively supports.

If you don't care about supporting HTTP/1.1 then you can run this code which only
supports H2C prior knowledge.

```go
server := http2.Server{}

l, err := net.Listen("tcp", "0.0.0.0:1010")
checkErr(err, "while listening")

fmt.Printf("Listening [0.0.0.0:1010]...\n")
for {
    conn, err := l.Accept()
    checkErr(err, "during accept")

    server.ServeConn(conn, &http2.ServeConnOpts{
        Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            fmt.Fprintf(w, "Hello, %v, http: %v", r.URL.Path, r.TLS == nil)
        }),
    })
}
```

Once you have a running server you can test your server by installing `curl-openssl`.

```
$ brew install curl-openssl

# Add curl-openssl to the front of your path
$ export PATH="/usr/local/opt/curl-openssl/bin:$PATH"
```

You can now use curl to test your H2C enabled server like so.

##### Connect via HTTP1.1 then upgrade to HTTP/2 (H2C)
```
curl -v --http2 http://localhost:1010
*   Trying ::1:1010...
* TCP_NODELAY set
* Connected to localhost (::1) port 1010 (#0)
> GET / HTTP/1.1
> Host: localhost:1010
> User-Agent: curl/7.65.0
> Accept: */*
> Connection: Upgrade, HTTP2-Settings
> Upgrade: h2c
> HTTP2-Settings: AAMAAABkAARAAAAAAAIAAAAA
>
* Mark bundle as not supporting multiuse
< HTTP/1.1 101 Switching Protocols
< Connection: Upgrade
< Upgrade: h2c
* Received 101
* Using HTTP2, server supports multi-use
* Connection state changed (HTTP/2 confirmed)
* Copying HTTP/2 data in stream buffer to connection buffer after upgrade: len=0
* Connection state changed (MAX_CONCURRENT_STREAMS == 250)!
< HTTP/2 200
< content-type: text/plain; charset=utf-8
< content-length: 20
< date: Wed, 05 Jun 2019 19:01:40 GMT
<
* Connection #0 to host localhost left intact
Hello, /, http: true
```

##### Connect via HTTP/2 (H2C)
```
curl -v --http2-prior-knowledge http://localhost:1010
*   Trying ::1:1010...
* TCP_NODELAY set
* Connected to localhost (::1) port 1010 (#0)
* Using HTTP2, server supports multi-use
* Connection state changed (HTTP/2 confirmed)
* Copying HTTP/2 data in stream buffer to connection buffer after upgrade: len=0
* Using Stream ID: 1 (easy handle 0x7fdab8007000)
> GET / HTTP/2
> Host: localhost:1010
> User-Agent: curl/7.65.0
> Accept: */*
>
* Connection state changed (MAX_CONCURRENT_STREAMS == 250)!
< HTTP/2 200
< content-type: text/plain; charset=utf-8
< content-length: 20
< date: Wed, 05 Jun 2019 19:00:43 GMT
<
* Connection #0 to host localhost left intact
Hello, /, http: true
```

Now, Remember when I said that the golang standard library does not support H2C?
While that is technically correct there is a workaround to get the golang
standard http2 client to connect to an H2C enabled server.


To do so you have to override `DialTLS` and set the super secret `AllowHTTP` flag.

```go
client := http.Client{
    Transport: &http2.Transport{
        // So http2.Transport doesn't complain the URL scheme isn't 'https'
        AllowHTTP: true,
        // Pretend we are dialing a TLS endpoint. (Note, we ignore the passed tls.Config)
        DialTLSContext: func(ctx context.Context, network, addr string, cfg *tls.Config) (net.Conn, error) {
            var d net.Dialer
            return d.DialContext(ctx, network, addr)
        },
    },
}

resp, _ := client.Get(url)
fmt.Printf("Client Proto: %d\n", resp.ProtoMajor)
```

Although this all looks a little wonky it actually works really well and
performs nicely in production environments.
