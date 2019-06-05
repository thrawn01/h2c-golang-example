package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"golang.org/x/net/http2"
	"net"
	"net/http"
	"os"
	"time"
)

func checkErr(err error, msg string) {
	if err == nil {
		return
	}
	fmt.Printf("ERROR: %s: %s\n", msg, err)
	os.Exit(1)
}

func main() {
	HttpClientExample()
	//RoundTripExample()
}

const url = "http://localhost:9030/_groupcache/"

func RoundTripExample() {
	req, err := http.NewRequest("GET", url, nil)
	checkErr(err, "during new request")

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	tr := &http2.Transport{
		AllowHTTP: true,
		DialTLS: func(network, addr string, cfg *tls.Config) (net.Conn, error) {
			return net.Dial(network, addr)
		},
	}

	req.WithContext(ctx)
	resp, err := tr.RoundTrip(req)
	checkErr(err, "during roundtrip")

	fmt.Printf("RoundTrip Proto: %d\n", resp.ProtoMajor)
}

func HttpClientExample() {
	client := http.Client{
		Transport: &http2.Transport{
			AllowHTTP: true,
			DialTLS: func(network, addr string, cfg *tls.Config) (net.Conn, error) {
				return net.Dial(network, addr)
			},
		},
	}

	resp, err := client.Get(url)
	checkErr(err, "during get")

	fmt.Printf("Client Proto: %d\n", resp.ProtoMajor)
}
