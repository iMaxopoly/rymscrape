package main

import (
	"errors"
	"time"

	"github.com/uber-go/zap"
	"github.com/valyala/fasthttp"
)

// requestGet gets html content from remote http server using fasthttp package
func requestGet(link string, timeout uint, wantsLocation bool, siteSig string) (body []byte, header fasthttp.ResponseHeader, err error) {
	redirCount := 0
redir:
	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()

	req.SetRequestURI(link)
	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/54.0.2840.71 Safari/537.36")
	req.Header.Add("Accept-Language", "en-US,en;q=0.8")
	req.Header.Add("Upgrade-Insecure-Requests", "1")
	req.Header.Add("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	req.Header.SetMethod("GET")

	err = fasthttp.DoTimeout(req, resp, time.Duration(timeout)*time.Second)
	if err != nil {
		return []byte{}, fasthttp.ResponseHeader{}, err
	}

	body = resp.Body()
	header = resp.Header
	locationBytes := header.Peek("Location")
	locationString := string(locationBytes)

	fasthttp.ReleaseRequest(req)
	fasthttp.ReleaseResponse(resp)

	if wantsLocation {
		if locationString != "" {
			return locationBytes, fasthttp.ResponseHeader{}, nil
		} else {
			return locationBytes, fasthttp.ResponseHeader{}, errors.New("Location was empty")
		}
	}

	if locationString != "" && locationString != link && redirCount < 15 {
		logger.Debug("redir anticipated", zap.String("link", link), zap.String("location", locationString))
		link = locationString
		redirCount++
		goto redir
	} else if locationString != "" && locationString != link && redirCount >= 15 {
		return []byte{}, fasthttp.ResponseHeader{}, errors.New("Tons of redirect")
	}

	return body, header, nil
}
