package signer

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	HeaderEopDate          = "eop-date"
	HeaderEopRequestID     = "ctyun-eop-request-id"
	HeaderEopAuthorization = "eop-authorization"
)

type OpenApiSigner interface {
	Sign() (http.Header, error)

	Setter
}

type Setter interface {
	SetRequestId(string) OpenApiSigner
	SetBody([]byte) OpenApiSigner
	AddHeader(header http.Header) OpenApiSigner
	SetHeader(header http.Header) OpenApiSigner
	SetParam(map[string][]string) OpenApiSigner
}

type openApiSigner struct {
	requestId string
	regionId  string
	body      []byte
	header    http.Header
	param     url.Values
	ak        string
	sk        string
}

func NewOpenApiSigner(ak, sk string) OpenApiSigner {
	return &openApiSigner{
		ak:     ak,
		sk:     sk,
		header: make(http.Header),
	}
}

func (s *openApiSigner) Sign() (http.Header, error) {
	eopDate := time.Now().Format("20060102T150405Z")

	hash := sha256.New()
	_, err := hash.Write(s.body)
	if err != nil {
		return nil, err
	}

	bodyDigest := hex.EncodeToString(hash.Sum(nil))
	headerStr := fmt.Sprintf("ctyun-eop-request-id:%s\neop-date:%s\n", s.requestId, eopDate)
	var queryStr = ""
	if s.param != nil {
		queryStr = s.param.Encode()
	}
	signatureStr := fmt.Sprintf("%s\n%s\n%s", headerStr, queryStr, bodyDigest)

	signDate := strings.Split(eopDate, "T")[0]

	kTime := HmacSha256(s.sk, eopDate)
	kAk := HmacSha256(kTime, s.ak)
	kDate := HmacSha256(kAk, signDate)

	signatureBase64 := base64.StdEncoding.EncodeToString([]byte(HmacSha256(kDate, signatureStr)))

	authorization := fmt.Sprintf("%s Headers=%s;%s Signature=%s", s.ak, HeaderEopRequestID, HeaderEopDate, signatureBase64)

	signHeader := make(http.Header)
	signHeader.Add(HeaderEopDate, eopDate)
	signHeader.Add(HeaderEopRequestID, s.requestId)
	signHeader.Add(HeaderEopAuthorization, authorization)

	return signHeader, nil
}

func (s *openApiSigner) SetRequestId(requestId string) OpenApiSigner {
	s.requestId = requestId
	return s
}

func (s *openApiSigner) SetBody(body []byte) OpenApiSigner {
	s.body = body
	return s
}

func (s *openApiSigner) AddHeader(header http.Header) OpenApiSigner {
	for k, v := range header {
		s.header[k] = v
	}
	return s
}

func (s *openApiSigner) SetHeader(header http.Header) OpenApiSigner {
	s.header = header
	return s
}

func (s *openApiSigner) SetParam(param map[string][]string) OpenApiSigner {
	s.param = param
	return s
}

func HmacSha256(key string, data string) string {
	hmac := hmac.New(sha256.New, []byte(key))
	hmac.Write([]byte(data))
	return string(hmac.Sum([]byte(nil)))
}
