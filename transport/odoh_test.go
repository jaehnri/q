package transport

import (
	"testing"

	"github.com/miekg/dns"
	"github.com/stretchr/testify/assert"
)

func TestTransportBuildURL(t *testing.T) {
	// Test with no query params
	u := buildURL("https://www.example.com", "")
	assert.Equal(t, "https://www.example.com", u.String())

	// Test with query params
	u = buildURL("https://www.example.com", "?foo=bar&baz=qux")
	assert.Equal(t, "https://www.example.com/%3Ffoo=bar&baz=qux", u.String())

	// Test with HTTP
	//goland:noinspection HttpUrlsUsage
	u = buildURL("http://www.example.com", "")
	assert.Equal(t, "http://www.example.com", u.String())
}

// TODO: Enable test
//func TestTransportODoH(t *testing.T) {
//	msg := dns.Msg{}
//	msg.RecursionDesired = true
//	msg.Question = []dns.Question{{
//		Name:   "example.com.",
//		Qtype:  dns.StringToType["A"],
//		Qclass: dns.ClassINET,
//	}}
//
//	reply, err := ODoH(msg, "odoh.cloudflare-dns.com", "odoh.crypto.sx")
//	assert.Nil(t, err)
//	assert.Greater(t, len(reply.Answer), 0)
//}

func TestTransportODoHInvalidTarget(t *testing.T) {
	msg := dns.Msg{}
	msg.RecursionDesired = true
	msg.Question = []dns.Question{{
		Name:   "example.com.",
		Qtype:  dns.StringToType["A"],
		Qclass: dns.ClassINET,
	}}

	_, err := ODoH(msg, "example.com", "odoh.crypto.sx")
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "Invalid serialized ObliviousDoHConfig")
}

func TestTransportODoHInvalidProxy(t *testing.T) {
	msg := dns.Msg{}
	msg.RecursionDesired = true
	msg.Question = []dns.Question{{
		Name:   "example.com.",
		Qtype:  dns.StringToType["A"],
		Qclass: dns.ClassINET,
	}}

	_, err := ODoH(msg, "odoh.cloudflare-dns.com", "example.com")
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "responded with an invalid Content-Type header")
}
