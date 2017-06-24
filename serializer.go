package main

import (
	"fmt"
	"math"
	"net/http"
	"strconv"
	"time"
)

func serialize(response http.ResponseWriter, request *http.Request) {
	hashString := request.URL.Query().Get("hash")
	chunkPowerString := request.URL.Query().Get("chunkpwr")
	chunkSegmentString := request.URL.Query().Get("segment")

	hash, err := strconv.ParseUint(hashString, 10, 64) // The hash number is a uint64 (base 10)
	if err != nil {
		http.Error(response, err.Error(), 400)
		return
	}

	cache.mux.RLock()
	if _, exists := cache.table[hash]; !exists {
		cache.mux.RUnlock()
		http.Error(response, "Specified hash isn't cached.", 400)
		return
	}
	cache.mux.RUnlock()

	chunkDimensions, err := strconv.ParseInt(chunkPowerString, 10, 0)
	if err != nil {
		http.Error(response, err.Error(), 400)
		return
	} else if chunkDimensions < 0 {
		http.Error(response, "chunkpwr must be greater than 0.", 400)
		return
	}
	chunkDimensions = int64(math.Pow(2, float64(chunkDimensions))) // We are making all dimensions a power of 2 so that they are all divisible without remainder or rounding

	chunkSegment, err := strconv.ParseInt(chunkSegmentString, 10, 0) // The chunk segment number is a plain old int (base 10)
	if err != nil {
		http.Error(response, err.Error(), 400)
		return
	} else if chunkSegment > int64(imageSizeMax*imageSizeMax)/(chunkDimensions*chunkDimensions) || chunkSegment < 0 { // Check to see if the segment is greater than the maximum allowed segments or less than 0, we can assume no remainders or rounding because of powers of 2
		http.Error(response, "Chunk segment is not between the maximum for the specified chunk percentage and 0.", 400)
		return
	}

	// We haven't encountered any errors with the request, so update the last time accesed
	cache.mux.Lock()
	tmp := cache.table[hash]
	tmp.lastAccess = time.Now()
	cache.table[hash] = tmp
	cache.mux.Unlock()

	var x, y int
	maximumAxisChunks := imageSizeMax / int(chunkDimensions)                     // Once again, we can assume no remainder or rounding because these are both powers of 2. This variable is the maximum chunks on the X or Y axes
	chunksOver := math.Trunc(float64(chunkSegment) / float64(maximumAxisChunks)) // Number of chunks over the maximum chunks in an axis
	x = int(chunkSegment*chunkDimensions) - int(chunkSegment*int64(chunksOver))  // Get current X
	y = int(chunksOver) * int(chunkDimensions)                                   // Get current Y
	fmt.Println(x, y)

	// Serialize the requested segment of the image
	cache.mux.RLock()
	table := cache.table[hash] // Take a memory hit to stop ourselves from blocking other goroutines while serializing
	cache.mux.RUnlock()
	image := &table.image
	fmt.Println(int(chunkDimensions*chunkDimensions), int((chunkSegment+1)*chunkDimensions)-int(chunkSegment*int64(chunksOver)), int(chunksOver+1)*int(chunkDimensions)+int(chunkDimensions))
	for i := 0; i <= int(chunkDimensions*chunkDimensions); i++ {
		if x >= int((chunkSegment+1)*chunkDimensions)-int(chunkSegment*int64(chunksOver)) { // Keep ourselves within the X bounds of the chunkSegment
			fmt.Fprint(response, "nnnnnnnnn")                                           // End of X line
			x = int(chunkSegment*chunkDimensions) - int(chunkSegment*int64(chunksOver)) // Move X to origin
			y++
		} else if y >= int(chunksOver)*int(chunkDimensions)+int(chunkDimensions) { // Keep ourselves within the X bounds of the chunkSegment
			break
		} else {
			offset := image.PixOffset(x, y)
			if table.image.Pix[offset+3] != 0 {
				values := [3]int{int(table.image.Pix[offset]), int(table.image.Pix[offset+1]), int(table.image.Pix[offset+2])}
				for v := range values {
					fmt.Fprintf(response, "%03d", v) // 0 pad the serialized RGB values up for up to 3 digits and write them to the HTTP response
				}
			} else {
				fmt.Fprint(response, "000000000") // Make transparent pixels black
			}
		}
		x++
	}
	fmt.Fprint(response, "eeeeeeeee") // End of Y line and segment
}
