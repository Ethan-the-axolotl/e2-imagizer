package main

import (
	"errors"
	"flag"
	"fmt"
	"image"
	"image/draw"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/nfnt/resize"
)

func main() {
	var host string
	var port string
	flag.StringVar(&host, "host", "localhost", "host to listen on for the webserver")
	flag.StringVar(&port, "port", "8000", "port to listen on for the webserver")
	flag.Parse()
	http.HandleFunc("/", serve)
	bind := fmt.Sprintf("%s:%s", host, port)
	fmt.Printf("listening on %s...\n", bind)
	err := http.ListenAndServe(bind, nil)
	if err != nil {
		log.Panic(err)
	}
}

func serve(response http.ResponseWriter, request *http.Request) {
	target := request.URL.Query().Get("q")
	u, err := url.Parse(target) // Parse the raw URL value we were given into somthing we can work with
	if err != nil {
		http.Error(response, err.Error(), 400)
		return
	}

	if !u.IsAbs() { // Is our URL absolute or not?
		u.Scheme = "http"
	} else { // If our URL is absolute, make sure the protocol is http(s)
		if strings.HasPrefix(u.Scheme, "http") {
			u.Scheme = "http"
		}
	}
	targetFormatted := u.String() // Turn our type URL back into a nice easy string, and store it in a variable
	fetchedResponse, err := http.Get(targetFormatted)
	if err != nil {
		http.Error(response, err.Error(), 400)
		return
	}
	fetchedBody, err := ioutil.ReadAll(fetchedResponse.Body) // Read the response into another variable
	if err != nil {
		http.Error(response, err.Error(), 400)
		return
	}
	detectedContentType := http.DetectContentType(fetchedBody)
	parsedContentType, err := parseContentType(detectedContentType)
	if err != nil {
		http.Error(response, err.Error(), 500)
		return
	}
	if parsedContentType.Type == "image" {
		if parsedContentType.Subtype == "jpeg" || parsedContentType.Subtype == "png" {
			serialized, err := serializeImage(strings.NewReader(string(fetchedBody)))
			if err != nil {
				http.Error(response, err.Error(), 500)
				return
			}
			fmt.Fprint(response, serialized)
		}
	}
}

func serializeImage(reader io.Reader) (string, error) {
	var output string

	img, _, err := image.Decode(reader)
	if err != nil {
		return "", err
	}
	m := resize.Thumbnail(512, 512, img, resize.Lanczos3)
	rect := m.Bounds()
	rgba := image.NewRGBA(rect)
	draw.Draw(rgba, rect, m, rect.Min, draw.Src)
	output += strconv.Itoa(rect.Max.X) + " " + strconv.Itoa(rect.Max.Y) + "|"
	for e := 0; e <= rect.Max.Y-1; e++ {
		for i := 0; i <= rect.Max.X-1; i++ {
			offset := rgba.PixOffset(i, e)
			values := [3]int{int(rgba.Pix[offset]), int(rgba.Pix[offset+1]), int(rgba.Pix[offset+2])}
			output += strconv.Itoa(values[0]) + "," + strconv.Itoa(values[1]) + "," + strconv.Itoa(values[2]) + " "
		}
	}
	return output, nil
}

type contentType struct { // The contentType type holds easily usable information that is normally held as a string for indentifying MIME type and character encoding along with other information
	Type       string            // The first part of the MIME type (eg. "text")
	Subtype    string            // The second part of the MIME type (eg. "html")
	Parameters map[string]string // Any extra information (eg. "charset=utf8") represeted as a map
}

func parseContentType(rawcontype string) (*contentType, error) { // Parse a MIME string into a contentType struct
	rawcontype = strings.ToLower(rawcontype)
	var conType contentType
	conType.Parameters = make(map[string]string)
	splitcontype := strings.Split(rawcontype, " ")
	splitcontype[0] = strings.Replace(splitcontype[0], ";", "", -1)
	mimetype := strings.Split(splitcontype[0], "/")
	if len(mimetype) <= 1 {
		return new(contentType), errors.New("contype: malformed content-type MIME type provided")
	}
	if len(splitcontype) > 1 {
		params := strings.Split(splitcontype[1], ";")
		for it := range params {
			splitparams := strings.Split(params[it], "=")
			conType.Parameters[splitparams[0]] = splitparams[1]
		}
	}
	conType.Type = mimetype[0]
	conType.Subtype = mimetype[1]
	return &conType, nil
}
