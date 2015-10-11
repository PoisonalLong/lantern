package measured

import (
	"fmt"
	"net"
	"net/http"
	"runtime"
	"testing"
	"time"

	"github.com/getlantern/testify/assert"
)

type MockReporter struct {
	s []*Stats
}

func (nr *MockReporter) Submit(s *Stats) error {
	nr.s = append(nr.s, s)
	return nil
}

func TestReportError(t *testing.T) {
	nr := MockReporter{}
	AddReporter(&nr)
	Start()
	defer Stop()
	runtime.Gosched()
	d := Dialer(net.Dial, "localhost:9000")
	_, _ = d("tcp", "localhost:9999")
	_, _ = d("tcp", "localhost:9998")
	time.Sleep(100 * time.Millisecond)
	if assert.Equal(t, 2, len(nr.s)) {
		assert.Equal(t, "errors", nr.s[0].Type, "should report correct server")
		assert.Equal(t, "localhost:9000", nr.s[0].Tags["server"], "should report correct server")
		assert.Equal(t, "connection refused", nr.s[0].Tags["error"], "should report connection reset")
		assert.Equal(t, 1, nr.s[0].Fields["value"], "should report connection reset")

		assert.Equal(t, "errors", nr.s[1].Type, "should report correct server")
		assert.Equal(t, "localhost:9000", nr.s[1].Tags["server"], "should report correct server")
		assert.Equal(t, "connection refused", nr.s[1].Tags["error"], "should report connection reset")
		assert.Equal(t, 1, nr.s[1].Fields["value"], "should report connection reset")
	}
}

func TestDefaultTags(t *testing.T) {
	SetDefaults(map[string]string{"app": "test-app"})
	nr := MockReporter{}
	AddReporter(&nr)
	Start()
	defer Stop()
	runtime.Gosched()
	reportError("test-server", fmt.Errorf("test-error"), "dial-phase")
	time.Sleep(100 * time.Millisecond)
	if assert.Equal(t, 1, len(nr.s)) {
		assert.Equal(t, "test-app", nr.s[0].Tags["app"], "should report default tags")
	}
}

func TestListener(t *testing.T) {
	l, err := net.Listen("tcp", ":0")
	if assert.NoError(t, err, "Listen should not fail") {
		ml := Listener(l)
		s := http.Server{Handler: http.NotFoundHandler()}
		go s.Serve(ml)
	}
	_, _ = http.Get("http://" + l.Addr().String())
}
