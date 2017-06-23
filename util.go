package main

import (
	"errors"
	"strings"
)

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
