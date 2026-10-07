package sdktest_test

import (
	"context"
	"errors"
	"fmt"
	"github.com/rambow-cloud/alicloud-go-sdk-x/sdktest"
	"io"
	"net/http"
	"sync"
	"testing"
	"time"
)

func TestScriptCopyChecksExhaustion(t *testing.T) {
	header := http.Header{"X-Test": []string{"original"}}
	tr := sdktest.NewTransport(sdktest.Step{Header: header, Body: "ok", Check: func(r *http.Request) error {
		if r.URL.Host != "example.invalid" {
			t.Error("host")
		}
		return nil
	}})
	header.Set("X-Test", "changed")
	req, _ := http.NewRequest("GET", "https://example.invalid", nil)
	res, err := tr.RoundTrip(req)
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(res.Body)
	res.Body.Close()
	if string(body) != "ok" || res.Header.Get("X-Test") != "original" {
		t.Fatal("response not copied")
	}
	if _, err = tr.RoundTrip(req); err == nil || tr.Calls() != 2 {
		t.Fatal("exhaustion must fail")
	}
	sentinel := errors.New("check failed")
	tr = sdktest.NewTransport(sdktest.Step{Check: func(*http.Request) error { return sentinel }})
	_, err = tr.RoundTrip(req)
	if !errors.Is(err, sentinel) {
		t.Fatal(err)
	}
}
func TestConcurrentClockAndCanceledSleep(t *testing.T) {
	c := sdktest.NewClock(time.Unix(0, 0))
	var wg sync.WaitGroup
	for range 20 {
		wg.Go(func() { c.Advance(time.Second) })
	}
	wg.Wait()
	if !c.Now().Equal(time.Unix(20, 0)) {
		t.Fatal(c.Now())
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := c.Sleep(ctx, time.Hour); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if !c.Now().Equal(time.Unix(20, 0)) {
		t.Fatal("canceled sleep advanced")
	}
}
func ExampleNewTransport() {
	tr := sdktest.NewTransport(sdktest.Step{Body: `{"RequestId":"offline"}`})
	req, _ := http.NewRequest("GET", "https://example.invalid", nil)
	res, _ := tr.RoundTrip(req)
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)
	fmt.Println(string(body))
	// Output: {"RequestId":"offline"}
}
