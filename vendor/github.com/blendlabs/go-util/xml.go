package util

import (
	"bytes"
	"encoding/xml"
	"io"
	"regexp"
)

var (
	cdataPrefix = []byte("<![CDATA[")
	cdataSuffix = []byte("]]>")
	cdataRe     = regexp.MustCompile("<!\\[CDATA\\[(.*?)\\]\\]>")
)

// XMLCharsetReader is a reader for a given charset (i.e. non utf-8 charsets).
type XMLCharsetReader func(charset string, input io.Reader) (io.Reader, error)

// EncodeCDATA writes a data blob to a cdata tag.
func EncodeCDATA(data []byte) []byte {
	return bytes.Join([][]byte{cdataPrefix, data, cdataSuffix}, []byte{})
}

// DecodeCDATA decodes a cdata tag to a byte array.
func DecodeCDATA(cdata []byte) []byte {
	matches := cdataRe.FindAllSubmatch(cdata, 1)
	if len(matches) == 0 {
		return cdata
	}

	return matches[0][1]
}

// DeserializeXML unmarshals xml to an object.
func DeserializeXML(object interface{}, body string) error {
	return DeserializeXMLFromReader(object, bytes.NewBufferString(body))
}

// DeserializeXMLFromReader unmarshals xml to an object from a reader
func DeserializeXMLFromReader(object interface{}, reader io.Reader) error {
	decoder := xml.NewDecoder(reader)
	return decoder.Decode(object)
}

// DeserializeXMLFromReaderWithCharsetReader uses a charset reader to deserialize xml.
func DeserializeXMLFromReaderWithCharsetReader(object interface{}, body io.Reader, charsetReader XMLCharsetReader) error {
	decoder := xml.NewDecoder(body)
	decoder.CharsetReader = charsetReader
	return decoder.Decode(object)
}

// SerializeXML marshals an object to xml.
func SerializeXML(object interface{}) string {
	b, _ := xml.Marshal(object)
	return string(b)
}

// SerializeXMLToReader marshals an object to a reader.
func SerializeXMLToReader(object interface{}) io.Reader {
	b, _ := xml.Marshal(object)
	return bytes.NewBufferString(string(b))
}
