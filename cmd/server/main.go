package main

import (
	"fmt"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"net"
	"net/http"
	"os"
)

func checkErr(err error, msg string) {
	if err == nil {
		return
	}
	fmt.Printf("ERROR: %s: %s\n", msg, err)
	os.Exit(1)
}

func main() {
	H2CServerUpgrade()
	//H2CServerPrior()
}

// H2CServerUpgrade server supports "H2C upgrade" and "H2C prior knowledge" along with
// standard HTTP/2 and HTTP/1.1 that golang natively supports.
func H2CServerUpgrade() {
	h2s := &http2.Server{}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, %v, http: %v", r.URL.Path, r.TLS == nil)
	})

	server := &http.Server{
		Addr:    "0.0.0.0:1010",
		Handler: h2c.NewHandler(handler, h2s),
	}

	checkErr(http2.ConfigureServer(server, h2s), "during call to ConfigureServer()")

	fmt.Printf("Listening [0.0.0.0:1010]...\n")
	checkErr(server.ListenAndServe(), "while listening")
}

// H2CServerPrior server only supports "H2C prior knowledge".
// You can add standard HTTP/2 support by adding a TLS config.
func H2CServerPrior() {
	server := http2.Server{}

	l, err := net.Listen("tcp", "0.0.0.0:1010")
	checkErr(err, "while listening")

	fmt.Printf("Listening [0.0.0.0:1010]...\n")
	for {
		conn, err := l.Accept()
		checkErr(err, "during accept")

		server.ServeConn(conn, &http2.ServeConnOpts{
			Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				fmt.Printf("New Connection: %+v\n", r)
				fmt.Fprintf(w, "Hello, %v, http: %v", r.URL.Path, r.TLS == nil)
			}),
		})
		_ = conn.Close()
	}
}
