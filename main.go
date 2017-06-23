package main

import (
	"flag"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"net/http"
	"sync"
)

var softCacheSize int // TODO: Implement cache clearing
var imageSizeMax int
var requests requestHashTable

type requestHashTable struct {
	table map[uint64]image.RGBA
	mux   sync.RWMutex
}

func main() {
	var host string
	var port string
	flag.StringVar(&host, "host", "localhost", "host to listen on for the webserver")
	flag.StringVar(&port, "port", "8000", "port to listen on for the webserver")
	flag.IntVar(&softCacheSize, "softcache", 16, "the maximum allowed number of images in the cache")
	flag.IntVar(&imageSizeMax, "imagemax", 512, "the maximum square size of the image (will be automatically downsampled if larger)")
	flag.Parse()

	requests.table = make(map[uint64]image.RGBA)

	http.HandleFunc("/initalize/", initalize)
	http.HandleFunc("/serialize/", serialize)
	bind := fmt.Sprintf("%s:%s", host, port)
	fmt.Printf("listening on %s...\n", bind)
	log.Panic(http.ListenAndServe(bind, nil))
}
