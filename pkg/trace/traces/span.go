package traces

import "io"

type MetaIterFunc func(k, v string) bool

type MetricIterFunc func(k string, v float64) bool

type Span interface {
	TraceID() uint64
	SetTraceID(x uint64)

	SpanID() uint64
	SetSpanID(x uint64)

	UnsafeService() string
	SetService(s string)

	UnsafeName() string
	SetName(s string)

	UnsafeResource() string
	SetResource(s string)

	Duration() int64
	SetDuration(d int64)

	ParentID() uint64
	SetParentID(x uint64)

	Start() int64
	SetStart(x int64)

	UnsafeType() string
	SetType(s string)

	Error() int32
	SetError(x int32)

	GetMetaUnsafe(s string) (string, bool)
	SetMeta(k, v string)
	ForEachMetaUnsafe(MetaIterFunc)

	GetMetric(s string) (float64, bool)
	SetMetric(k string, v float64)
	ForEachMetricUnsafe(MetricIterFunc)

	MsgSize() int

	WriteProto(w io.Writer) error

	DebugString() string
}