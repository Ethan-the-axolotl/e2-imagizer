package main

import (
	"fmt"
	"net/http"
	"strconv"
	"time"
)

func serialize(response http.ResponseWriter, request *http.Request) {
	hashString := request.URL.Query().Get("hash")
	chunkPercentageString := request.URL.Query().Get("chunkpercent")
	segmentString := request.URL.Query().Get("segment")

	hash, err := strconv.ParseUint(hashString, 10, 64) // The hash number is a uint64 (base 10)
	if err != nil {
		http.Error(response, err.Error(), 400)
		return
	}

	chunkPercentage, err := strconv.ParseFloat(chunkPercentageString, 64) // The chunk size number is a plain old int (base 10)
	if err != nil {
		http.Error(response, err.Error(), 400)
		return
	} else if chunkPercentage > 1 || chunkPercentage < 0 {
		http.Error(response, "Chunk percentage must be between 1 and 0.", 400)
	}

	chunkDimensions := chunkPercentage * float64(imageSizeMax) // This won't be larger than the image because of the previous check

	segment, err := strconv.ParseInt(segmentString, 10, 0) // The chunk size number is a plain old int (base 10)
	if err != nil {
		http.Error(response, err.Error(), 400)
		return
	} else if float64(segment) > 1/chunkPercentage || segment < 0 { // Check to see if the segment is greater than the maximum allowed segments or less than 0
		http.Error(response, "Segment is not between the maximum for the specified chunk percentage and 0.", 400)
		return
	}

	// We haven't encountered any errors with the request, so update the last time accesed
	cache.mux.Lock()
	tmp := cache.table[hash]
	tmp.lastAccess = time.Now()
	cache.mux.Unlock()

	var x, y int
	x = chunkDimensions * (segment)

	// Serialize the requested segment of the image
	cache.mux.RLock()
	table := cache.table[hash]
	image := &table.image
	for i := 1; i <= 1; i++ { // FIXMEd
		if i > image.Bounds().Max.X-1 { // If we're outside the bounds of our resized image we're out
			fmt.Fprint(response, "nnnnnnnnn") // End of X line
			y++
		} else if y > image.Bounds().Max.Y-1 { // Make sure we're not out of Y axis bounds
			break
		} else {
			offset := image.PixOffset(i, y)
			values := [3]int{int(cache.table[hash].image.Pix[offset]), int(cache.table[hash].image.Pix[offset+1]), int(cache.table[hash].image.Pix[offset+2])}
			for v := range values {
				fmt.Fprintf(response, "%03d\n", v) // 0 pad the serialized RGB values up for up to 3 digits and write them to the HTTP response
			}
		}
	}
	cache.mux.RUnlock()
	fmt.Fprint(response, "eeeeeeeee") // End of Y line and segment
}
