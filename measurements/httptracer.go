package measurements

import (
	"crypto/tls"
	"net/http/httptrace"
	"time"
)

type HTTPTracer struct {
	Trace             *httptrace.ClientTrace
	DNSStart          time.Time
	ConnStart         time.Time
	ConnDone          time.Time
	GotConn           time.Time
	StartResponding   time.Time
	TLSHandshakeStart time.Time
	TLSHandshakeDone  time.Time
	Finished          time.Time
}

func NewHTTPTracer() *HTTPTracer {
	tracer := &HTTPTracer{}
	trace := &httptrace.ClientTrace{
		DNSStart: func(_ httptrace.DNSStartInfo) { tracer.DNSStart = time.Now() },
		DNSDone:  func(_ httptrace.DNSDoneInfo) { tracer.ConnStart = time.Now() },
		ConnectStart: func(_, _ string) {
			if tracer.ConnStart.IsZero() {
				tracer.ConnStart = time.Now()
			}
		},
		ConnectDone:          func(_, _ string, _ error) { tracer.ConnDone = time.Now() },
		GotConn:              func(_ httptrace.GotConnInfo) { tracer.GotConn = time.Now() },
		GotFirstResponseByte: func() { tracer.StartResponding = time.Now() },
		TLSHandshakeStart:    func() { tracer.TLSHandshakeStart = time.Now() },
		TLSHandshakeDone:     func(_ tls.ConnectionState, _ error) { tracer.TLSHandshakeDone = time.Now() },
	}

	tracer.Trace = trace

	return tracer
}

func (t *HTTPTracer) Done() {
	if t != nil {
		t.Finished = time.Now()
	}
}

func (t *HTTPTracer) Total() time.Duration {
	return t.Finished.Sub(t.DNSStart)
}

func (t *HTTPTracer) Namelookup() time.Duration {
	return t.ConnStart.Sub(t.DNSStart)
}

func (t *HTTPTracer) Connect() time.Duration {
	return t.GotConn.Sub(t.DNSStart)
}

func (t *HTTPTracer) Pretransfer() time.Duration {
	return t.TLSHandshakeDone.Sub(t.DNSStart)
}

func (t *HTTPTracer) Starttransfer() time.Duration {
	return t.StartResponding.Sub(t.DNSStart)
}
