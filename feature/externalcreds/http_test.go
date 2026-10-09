package externalcreds

import (
	"net/http"
	"net/url"
	"testing"
)

func TestMetadataTransportOwnsProxyPolicy(t *testing.T) {
	original := http.DefaultTransport.(*http.Transport).Clone()
	original.Proxy = func(*http.Request) (*url.URL, error) { return url.Parse("http://proxy.invalid") }
	client := &http.Client{Transport: original}
	s, err := configure(Options{HTTPClient: client}, true)
	if err != nil {
		t.Fatal(err)
	}
	owned := s.client.(*http.Client).Transport.(*http.Transport)
	if owned == original || owned.Proxy != nil || original.Proxy == nil || client.CheckRedirect != nil {
		t.Fatal("metadata proxy isolation or caller ownership failed")
	}
}
