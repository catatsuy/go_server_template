package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"runtime"
	"syscall"
)

func main() {
	var (
		port = flag.Uint("port", 8080, "port to listen")
		sock = flag.String("sock", "", "unix domain socket to listen")
		fd   = flag.Uint("fd", 0, "file descriptor to listen and serve")
	)

	flag.Parse()

	sigchan := make(chan os.Signal)
	signal.Notify(sigchan, syscall.SIGTERM)
	signal.Notify(sigchan, syscall.SIGINT)

	var l net.Listener
	var err error

	if *fd == 0 && *sock == "" {
		// listening on port
		log.Println(fmt.Sprintf("listening on port %d", *port))
		l, err = net.ListenTCP("tcp", &net.TCPAddr{Port: int(*port)})
	}

	if *fd == 0 && *sock != "" {
		// listening on unix domain socket
		log.Println(fmt.Sprintf("listening on unix domain socket %s", *sock))
		ferr := os.Remove(*sock)
		if ferr != nil {
			if !os.IsNotExist(ferr) {
				panic(ferr.Error())
			}
		}
		l, err = net.Listen("unix", *sock)
		os.Chmod(*sock, 0777)
	} else if *sock == "" && *fd != 0 {
		// listening on file descriptor
		log.Println(fmt.Sprintf("listening on file descriptor %d", *fd))
		l, err = net.FileListener(os.NewFile(uintptr(*fd), ""))
	}

	if err != nil {
		panic(err.Error())
	}

	// http.HandleFunc("/", Action)

	// blocking profiler
	// cf: http://blog.livedoor.jp/sonots/archives/39879160.html
	runtime.SetBlockProfileRate(1)
	go func() {
		log.Println(http.Serve(l, nil))
	}()

	<-sigchan

	// remove socket file if listening on unix domain socket
	if *sock != "" {
		ferr := os.Remove(*sock)
		if ferr != nil {
			if !os.IsNotExist(ferr) {
				panic(ferr.Error())
			}
		}
	}
}
