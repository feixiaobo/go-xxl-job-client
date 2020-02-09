package transport

import (
	"net/http"
	"strconv"
)

type HttpResponsePkg struct {
	Proto      string
	StatusCode int
	StatusMsg  string
	Header     map[string]string
	Content    []byte
}

func NewHttpResponsePkg(code int, content []byte) *HttpResponsePkg {
	headerMap := make(map[string]string)
	if content == nil {
		headerMap["content-length"] = "0"
	} else {
		headerMap["content-length"] = strconv.Itoa(len(content))
	}
	headerMap["connection"] = "keep-alive"

	res := &HttpResponsePkg{
		Proto:      "HTTP/1.1",
		StatusCode: code,
		StatusMsg:  strconv.Itoa(code) + " " + http.StatusText(code),
		Content:    content,
		Header:     headerMap,
	}
	return res
}

func (p *HttpResponsePkg) Decoder() []byte {
	var (
		byteArray []byte
	)
	byteArray = append(byteArray, p.Proto...)
	byteArray = append(byteArray, " "...)
	byteArray = append(byteArray, p.StatusMsg...)
	byteArray = append(byteArray, "\n"...)
	byteArray = append(byteArray, " "...)
	if len(p.Header) > 0 {
		for k, v := range p.Header {
			c := k + ": " + v
			byteArray = append(byteArray, c...)
			byteArray = append(byteArray, "\n"...)
		}
	}

	byteArray = append(byteArray, "\n"...)
	byteArray = append(byteArray, p.Content...)
	return byteArray
}
