package llm

import (
	"net"
	"net/http"
	"time"
)

var (
	dialTimeout     = 1 * time.Second
	keepAlive       = 30 * time.Second
	responseTimeout = 30 * time.Second
)

var httpTransport = http.Transport{
	DialContext: (&net.Dialer{
		Timeout:   dialTimeout,
		KeepAlive: keepAlive,
	}).DialContext,
}

var httpClient = http.Client{
	Transport: &httpTransport,
	Timeout:   responseTimeout,
}
