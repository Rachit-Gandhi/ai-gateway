package httpserver

import (
	"context"
	"net"
	"net/http"
	"sync"
	"sync/atomic"

	"github.com/ferro-labs/ai-gateway/internal/metrics"
)

type connContextKey struct{}

// ConnMetadata holds per-connection observability data stored in the request context.
type ConnMetadata struct {
	ID         uint64
	LocalAddr  string
	RemoteAddr string
}

type connTracker struct {
	nextID atomic.Uint64

	mu     sync.Mutex
	states map[net.Conn]http.ConnState
}

func newConnTracker() *connTracker {
	return &connTracker{
		states: make(map[net.Conn]http.ConnState),
	}
}

// ConnContext attaches a ConnMetadata to the context for each new connection.
func (t *connTracker) ConnContext(ctx context.Context, conn net.Conn) context.Context {
	meta := ConnMetadata{
		ID:         t.nextID.Add(1),
		LocalAddr:  conn.LocalAddr().String(),
		RemoteAddr: conn.RemoteAddr().String(),
	}
	return context.WithValue(ctx, connContextKey{}, meta)
}

// ConnState records state transitions and updates Prometheus gauges/counters.
func (t *connTracker) ConnState(conn net.Conn, state http.ConnState) {
	t.observe(conn, state)
}

func (t *connTracker) observe(conn net.Conn, state http.ConnState) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if prev, ok := t.states[conn]; ok {
		decrementConnectionGauge(prev)
	}

	switch state {
	case http.StateActive, http.StateIdle:
		incrementConnectionGauge(state)
		metrics.ServerConnectionTransitionsTotal.WithLabelValues(connStateLabel(state)).Inc()
		t.states[conn] = state
	case http.StateClosed, http.StateHijacked:
		metrics.ServerConnectionTransitionsTotal.WithLabelValues(connStateLabel(state)).Inc()
		delete(t.states, conn)
	default:
		metrics.ServerConnectionTransitionsTotal.WithLabelValues(connStateLabel(state)).Inc()
		t.states[conn] = state
	}
}

// ConnMetadataFromContext extracts connection metadata stored by ConnContext.
func ConnMetadataFromContext(ctx context.Context) (ConnMetadata, bool) {
	meta, ok := ctx.Value(connContextKey{}).(ConnMetadata)
	return meta, ok
}

func incrementConnectionGauge(state http.ConnState) {
	label, ok := connectionGaugeLabel(state)
	if !ok {
		return
	}
	metrics.ServerConnectionsCurrent.WithLabelValues(label).Inc()
}

func decrementConnectionGauge(state http.ConnState) {
	label, ok := connectionGaugeLabel(state)
	if !ok {
		return
	}
	metrics.ServerConnectionsCurrent.WithLabelValues(label).Dec()
}

func connectionGaugeLabel(state http.ConnState) (string, bool) {
	switch state {
	case http.StateActive:
		return "active", true
	case http.StateIdle:
		return "idle", true
	default:
		return "", false
	}
}

func connStateLabel(state http.ConnState) string {
	switch state {
	case http.StateNew:
		return "new"
	case http.StateActive:
		return "active"
	case http.StateIdle:
		return "idle"
	case http.StateHijacked:
		return "hijacked"
	case http.StateClosed:
		return "closed"
	default:
		return "unknown"
	}
}
