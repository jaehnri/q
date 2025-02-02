package transport

import (
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/miekg/dns"
	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/http3"
	log "github.com/sirupsen/logrus"
)

// HTTP makes a DNS query over HTTP(s)
func HTTP(
	m *dns.Msg, tlsConfig *tls.Config,
	server, userAgent, method string,
	timeout, handshakeTimeout time.Duration,
	h3, noPMTUD bool) (*dns.Msg, error) {
	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig:     tlsConfig,
			MaxConnsPerHost:     1,
			MaxIdleConns:        1,
			TLSHandshakeTimeout: handshakeTimeout,
			Proxy:               http.ProxyFromEnvironment,
		},
		Timeout: timeout,
	}
	if h3 {
		log.Debug("Using HTTP/3")
		httpClient.Transport = &http3.RoundTripper{
			TLSClientConfig: tlsConfig,
			QuicConfig: &quic.Config{
				HandshakeIdleTimeout:    handshakeTimeout,
				DisablePathMTUDiscovery: noPMTUD,
			},
		}
	}

	buf, err := m.Pack()
	if err != nil {
		return nil, fmt.Errorf("packing message: %w", err)
	}

	queryURL := server + "?dns=" + base64.RawURLEncoding.EncodeToString(buf)
	req, err := http.NewRequest(method, queryURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating http request to %s: %w", queryURL, err)
	}

	req.Header.Set("Accept", "application/dns-message")
	if userAgent != "" {
		log.Debugf("Setting User-Agent to %s", userAgent)
		req.Header.Set("User-Agent", userAgent)
	}

	log.Debugf("[http] sending %s request to %s", method, queryURL)
	resp, err := httpClient.Do(req)
	if resp != nil && resp.Body != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return nil, fmt.Errorf("requesting %s: %w", queryURL, err)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", queryURL, err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("got status code %d from %s", resp.StatusCode, queryURL)
	}

	response := dns.Msg{}
	err = response.Unpack(body)
	if err != nil {
		return nil, fmt.Errorf("unpacking DNS response from %s: %w", queryURL, err)
	}

	if response.Id != m.Id {
		err = dns.ErrId
	}

	return &response, err
}
