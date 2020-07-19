package traces

import (
	"bytes"
	"testing"

	"github.com/DataDog/datadog-agent/pkg/trace/pb"
	"github.com/gogo/protobuf/proto"
	"github.com/stretchr/testify/require"
)

// TODO: Add tests for meta/metrics iteration.

func TestLazySpanUnmarshal(t *testing.T) {
	span := newTestSpan()
	marshaled, err := proto.Marshal(&span)
	require.NoError(t, err)

	lazy, err := NewLazySpan(marshaled, nil, nil)
	require.NoError(t, err)

	require.Equal(t, lazy.TraceID(), span.TraceID)
	require.Equal(t, lazy.SpanID(), span.SpanID)
	require.Equal(t, lazy.ParentID(), span.ParentID)
	require.Equal(t, lazy.UnsafeType(), span.Type)
	require.Equal(t, lazy.UnsafeService(), span.Service)
	require.Equal(t, lazy.UnsafeName(), span.Name)
	require.Equal(t, lazy.UnsafeResource(), span.Resource)
	require.Equal(t, lazy.Start(), span.Start)
	require.Equal(t, lazy.Duration(), span.Duration)

	for k, v := range span.Meta {
		lazyV, ok := lazy.GetMetaUnsafe(k)
		require.True(t, ok)
		require.Equal(t, v, lazyV)
	}

	for k, v := range span.Metrics {
		lazyV, ok := lazy.GetMetric(k)
		require.True(t, ok)
		require.Equal(t, v, lazyV)
	}
}

func TestLazySpanRoundTrip(t *testing.T) {
	span := newTestSpan()
	marshaled, err := proto.Marshal(&span)
	require.NoError(t, err)

	lazy, err := NewLazySpan(marshaled, nil, nil)
	require.NoError(t, err)

	buf := bytes.NewBuffer(nil)
	require.NoError(t, lazy.WriteProto(buf))

	roundTripped := pb.Span{}
	roundTripped.Unmarshal(buf.Bytes())

	require.Equal(t, span, roundTripped)
}

// TODO: Assert on mutations *before* serialization and after serialization/deserialization.
func TestLazySpanRoundTripMutate(t *testing.T) {
	span := newTestSpan()
	marshaled, err := proto.Marshal(&span)
	require.NoError(t, err)

	lazy, err := NewLazySpan(marshaled, nil, nil)
	require.NoError(t, err)

	// Top-level field mutations.
	lazy.SetTraceID(999)
	lazy.SetSpanID(999)
	lazy.SetParentID(999)
	lazy.SetStart(999)
	lazy.SetDuration(999)
	lazy.SetError(999)
	lazy.SetType("new_type")
	lazy.SetService("new_service")
	lazy.SetName("new_name")
	lazy.SetResource("new_resource")

	// Meta map mutations.
	// Mutate existing key.
	lazy.SetMeta("http.host", "new_host")
	// Add new key/value.
	lazy.SetMeta("http.new_field", "new_field")

	// Metrics map mutations.
	// Mutate existing key.
	lazy.SetMetric("http.monitor", 999.0)
	// Add new key/value.
	lazy.SetMetric("http.new_field", 999.0)

	buf := bytes.NewBuffer(nil)
	require.NoError(t, lazy.WriteProto(buf))

	mutated := pb.Span{}
	mutated.Unmarshal(buf.Bytes())

	require.Equal(t, uint64(999), mutated.TraceID)
	require.Equal(t, uint64(999), mutated.SpanID)
	require.Equal(t, uint64(999), mutated.ParentID)
	require.Equal(t, int64(999), mutated.Start)
	require.Equal(t, int64(999), mutated.Duration)
	require.Equal(t, int32(999), mutated.Error)
	require.Equal(t, "new_type", mutated.Type)
	require.Equal(t, "new_service", mutated.Service)
	require.Equal(t, "new_name", mutated.Name)
	require.Equal(t, "new_resource", mutated.Resource)

	require.Equal(t, map[string]string{
		"http.host":      "new_host",
		"http.port":      "8080",
		"http.new_field": "new_field",
	}, mutated.Meta)

	require.Equal(t, map[string]float64{
		"http.monitor":   999.0,
		"http.duration":  127.3,
		"http.new_field": 999.0,
	}, mutated.Metrics)

}

func newTestSpan() pb.Span {
	return pb.Span{
		TraceID:  42,
		SpanID:   52,
		ParentID: 42,
		Type:     "web",
		Service:  "fennel_IS amazing!",
		Name:     "something &&<@# that should be a metric!",
		Resource: "NOT touched because it is going to be hashed",
		Start:    9223372036854775807,
		Duration: 9223372036854775807,
		Meta: map[string]string{
			"http.host": "192.168.0.1",
			"http.port": "8080",
		},
		Metrics: map[string]float64{
			"http.monitor":  41.99,
			"http.duration": 127.3,
		},
	}
}