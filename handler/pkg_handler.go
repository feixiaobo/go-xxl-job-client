package handler

import (
	"bytes"
	"errors"
	"github.com/apache/dubbo-go-hessian2"
	"github.com/dubbogo/getty"
	"github.com/feixiaobo/go-xxl-job-client/transport"
	"strings"
)

type PkgHandlerRes struct {
	LastPkg     []byte
	LastSuccess bool
	Valid       bool
}

var (
	pkgRes = &PkgHandlerRes{
		LastPkg:     nil,
		LastSuccess: false,
		Valid:       false,
	}
	maxReadBufLen = 4 * 1024
)

type PackageHandler struct{}

func (h *PackageHandler) Read(ss getty.Session, data []byte) (interface{}, int, error) {
	var res []interface{}
	length := len(data)

	if pkgRes.Valid { //粘包
		var buffer bytes.Buffer
		buffer.Write(pkgRes.LastPkg)
		buffer.Write(data)
		data = buffer.Bytes()
		pkgRes.Valid = false
		pkgRes.LastPkg = nil
	}

	str := string(data[:]) //需要分包
	splitStr := "POST"
	strs := strings.Split(str, splitStr)
	splitLen := len(strs) - 1
	if splitLen >= 1 {
		for index, s := range strs {
			if index > 1 || (index == 1 && (!pkgRes.Valid || (pkgRes.Valid && !pkgRes.LastSuccess))) {
				pos := strings.Index(s, "\r\n\r\n") //去掉http头部
				success := false
				if pos != -1 {
					resbyte := []byte(s[pos+4:])
					obj, err := decoder(resbyte)
					if err == nil {
						res = append(res, obj)
						success = true
					}
				}
				unpacks(s, index, splitLen, length, success)
			}
		}
	}
	return res, length, nil
}

func (h *PackageHandler) Write(ss getty.Session, p interface{}) ([]byte, error) {
	pkg := p.(*transport.HttpResponsePkg)
	return pkg.Decoder(), nil
}

func unpacks(s string, index, splitLen, length int, success bool) {
	//位于最后一段字节且总字节长度超过最大长度时，最后一段字节可能是下一个包的前半部分
	if index == splitLen && length >= maxReadBufLen {
		var buffer bytes.Buffer
		buffer.WriteString("POST")
		buffer.WriteString(s)
		pkgRes.Valid = true
		pkgRes.LastSuccess = success
		pkgRes.LastPkg = buffer.Bytes()
	}
}

func decoder(buf []byte) (interface{}, error) {
	out := hessian.NewDecoder(buf)
	r, err := out.Decode()
	if err != nil {
		return nil, err
	} else if r == nil {
		return r, errors.New("decode hessian obj error")
	}
	return r, err
}
