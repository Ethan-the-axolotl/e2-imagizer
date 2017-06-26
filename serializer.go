package main

import (
	"fmt"
	"image/color"
	"math"
	"net/http"
	"strconv"
	"time"
)

func serialize(response http.ResponseWriter, request *http.Request) {
	hashString := request.URL.Query().Get("hash")
	chunkPowerString := request.URL.Query().Get("chunkpwr")
	chunkSegmentString := request.URL.Query().Get("segment")

	if hashString == "" || chunkPowerString == "" || chunkSegmentString == "" {
		http.Error(response, "hash, chunkpwr, and segment parameters cannot be empty.", 400)
		return
	}

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
	} else if chunkDimensions < 0 || int(chunkDimensions) > imageSizeMax*imageSizeMax {
		http.Error(response, "chunkpwr must be greater than 0 and smaller than the maximum image size.", 400)
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

	maximumAxisChunks := imageSizeMax / int(chunkDimensions)                                      // Once again, we can assume no remainder or rounding because these are both powers of 2. This variable is the maximum chunks on the X or Y axes
	chunksOver := math.Trunc(float64(chunkSegment) / float64(maximumAxisChunks))                  // Number of chunks over the maximum chunks in an axis
	segmentXMax := int((chunkSegment+1)*chunkDimensions) - int(chunkDimensions*int64(chunksOver)) // The X coordinate of the outer bound of the segment
	segmentYMax := int(chunksOver)*int(chunkDimensions) + int(chunkDimensions)                    // The Y coordinate of the outer bound of the segment
	pixelsInChunk := int(chunkDimensions * chunkDimensions)                                       // The number of pixels in a chunk

	x := int(chunkSegment*chunkDimensions) - int(chunkDimensions*int64(chunksOver)) // The current starting X, which will change in the loop
	y := int(chunksOver) * int(chunkDimensions)                                     // The current starting Y, which will change in the loop

	// Get the information to serialize
	cache.mux.RLock()
	table := cache.table[hash] // Take a memory hit to stop ourselves from blocking other goroutines while serializing
	cache.mux.RUnlock()
	image := &table.image

	// If we're completely out of bounds of the image than we don't need to do any other work than this
	if x >= image.Bounds().Max.X || y >= image.Bounds().Max.Y {
		fmt.Fprint(response, "eeeeeeeee")
		return
	}

	fmt.Println("X:", x, "Y:", y, "Chunks Over:", chunksOver, "X Maximum:", segmentXMax, "Y Maximum:", segmentYMax, "Pixels in a Chunk:", pixelsInChunk, "Image Max Bounds:", table.image.Bounds().Max)

	// Loop through all the pixels and serialize them
	for i := 0; i <= pixelsInChunk; i++ {
		if x > segmentXMax || x > table.image.Bounds().Max.X { // Keep ourselves within the X bounds of the chunkSegment *and* the image
			fmt.Fprint(response, "nnnnnnnnn")                                              // End of X line
			x = int(chunkSegment*chunkDimensions) - int(chunkDimensions*int64(chunksOver)) // Move X to origin
			y++
		} else if y > segmentYMax || y > table.image.Bounds().Max.Y { // Keep ourselves within the Y bounds of the chunkSegment *and* the image
			break
		} else {
			var curColor color.RGBA
			curColor = image.At(x, y).(color.RGBA) // We can do this type assertion because color.RGBA satisfies color.Color
			if curColor.A == 0 {
				curColor.R, curColor.G, curColor.B = 0, 0, 0 // Make transparent pixels black
			}
			// 0 pad the serialized RGB values up for up to 3 digits and write them to the HTTP response
			fmt.Fprintf(response, "%03d%03d%03d", curColor.R, curColor.G, curColor.R)
		}
		x++ // We are reading along the X-Axis so always iterate X
	}
	fmt.Fprint(response, "eeeeeeeee") // End of Y line and segment
}
