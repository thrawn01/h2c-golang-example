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

const url = "http://localhost:1010/_groupcache/"

func RoundTripExample() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	checkErr(err, "during new request")

	tr := &http2.Transport{
		AllowHTTP: true,
		DialTLSContext: func(ctx context.Context, network, addr string, cfg *tls.Config) (net.Conn, error) {
			var d net.Dialer
			return d.DialContext(ctx, network, addr)
		},
	}

	resp, err := tr.RoundTrip(req)
	checkErr(err, "during call to RoundTrip()")

	fmt.Printf("RoundTrip Proto: %d\n", resp.ProtoMajor)
}

func HttpClientExample() {
	client := http.Client{
		Transport: &http2.Transport{
			AllowHTTP: true,
			DialTLSContext: func(ctx context.Context, network, addr string, cfg *tls.Config) (net.Conn, error) {
				var d net.Dialer
				return d.DialContext(ctx, network, addr)
			},
		},
	}

	resp, err := client.Get(url)
	checkErr(err, "during get")

	fmt.Printf("Client Proto: %d\n", resp.ProtoMajor)
}
