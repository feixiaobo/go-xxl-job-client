package transport

import (
	"net/http"
	"strconv"
	"strings"
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
	byteArray = append(byteArray, strings.ToUpper(p.Proto)...)
	byteArray = append(byteArray, " "...)
	byteArray = append(byteArray, p.StatusMsg...)
	byteArray = append(byteArray, "\n"...)
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

type HttpRequestPkg struct {
	Header     map[string]string
	ContentLen int32
	Body       []byte
	MethodName string
}

func ParseHttpRequestPkg(content string) (*HttpRequestPkg, bool, []byte) {
	res := &HttpRequestPkg{}

	headerMap := make(map[string]string)
	pos := strings.Index(content, "\r\n\r\n")
	if pos > 0 {
		front := strings.Split(content[:pos], "\r\n")
		if len(front) > 0 {
			method := strings.Split(front[0], " ")[1][1:]
			res.MethodName = method

			for _, s := range front {
				ss := strings.Split(s, ": ")
				if len(ss) > 1 {
					headerMap[ss[0]] = ss[1]
				}
			}
		}

		length := 0
		lens, ok := headerMap["Content-Length"]
		if ok {
			l, err := strconv.Atoi(lens)
			if err == nil {
				length = l
			}
		}
		body := []byte(content[pos+4:])
		if len(body) < length {
			return nil, false, []byte(content)
		}
		res.Body = body
	}
	if len(headerMap) > 0 {
		res.Header = headerMap
	}
	return res, true, nil
}
