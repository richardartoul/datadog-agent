package traces

import (
	"fmt"
	"io"

	"github.com/DataDog/datadog-agent/pkg/trace/pb"
	"github.com/gogo/protobuf/proto"
)

var _ Span = &EagerSpan{}

type EagerSpan struct {
	Span pb.Span
}

func NewEagerSpan(span pb.Span) Span {
	return &EagerSpan{
		Span: span,
	}
}

func (e *EagerSpan) TraceID() uint64 {
	return e.Span.TraceID
}

func (e *EagerSpan) SetTraceID(x uint64) {
	e.Span.TraceID = x
}

func (e *EagerSpan) SpanID() uint64 {
	return e.Span.SpanID
}

func (e *EagerSpan) SetSpanID(x uint64) {
	e.Span.SpanID = x
}

func (e *EagerSpan) UnsafeService() string {
	// This operation is actually safe in this implementation, but callers should behave like its not.
	return e.Span.Service
}

func (e *EagerSpan) SetService(s string) {
	e.Span.Service = s
}

func (e *EagerSpan) UnsafeName() string {
	// This operation is actually safe in this implementation, but callers should behave like its not.
	return e.Span.Name
}

func (e *EagerSpan) SetName(s string) {
	e.Span.Name = s
}

func (e *EagerSpan) UnsafeResource() string {
	// This operation is actually safe in this implementation, but callers should behave like its not.
	return e.Span.Resource
}

func (e *EagerSpan) SetResource(s string) {
	e.Span.Resource = s
}

func (e *EagerSpan) Duration() int64 {
	return e.Span.Duration
}

func (e *EagerSpan) SetDuration(d int64) {
	e.Span.Duration = d
}

func (e *EagerSpan) ParentID() uint64 {
	return e.Span.ParentID
}

func (e *EagerSpan) SetParentID(id uint64) {
	e.Span.ParentID = id
}

func (e *EagerSpan) Start() int64 {
	return e.Span.Start
}

func (e *EagerSpan) SetStart(d int64) {
	e.Span.Start = d
}

func (e *EagerSpan) UnsafeType() string {
	// This operation is actually safe in this implementation, but callers should behave like its not.
	return e.Span.Type
}

func (e *EagerSpan) SetType(s string) {
	e.Span.Type = s
}

func (e *EagerSpan) Error() int32 {
	// This operation is actually safe in this implementation, but callers should behave like its not.
	return e.Span.Error
}

func (e *EagerSpan) SetError(x int32) {
	e.Span.Error = x
}

func (e *EagerSpan) GetMetaUnsafe(s string) (string, bool) {
	v, ok := e.Span.Meta[s]
	return v, ok
}

func (e *EagerSpan) SetMeta(k, v string) {
	e.Span.Meta[k] = v
}

func (e *EagerSpan) GetMetric(s string) (float64, bool) {
	v, ok := e.Span.Metrics[s]
	return v, ok
}

func (e *EagerSpan) SetMetric(k string, v float64) {
	e.Span.Metrics[k] = v
}

func (e *EagerSpan) MsgSize() int {
	return e.Span.Size()
}

func (e *EagerSpan) WriteProto(w io.Writer) error {
	// TODO: This is inefficient and allocates.
	marshaled, err := proto.Marshal(&e.Span)
	if err != nil {
		return fmt.Errorf("EagerSpan: WriteProto: error marshaling span: %v", err)
	}

	if _, err := w.Write(marshaled); err != nil {
		return fmt.Errorf("EagerSpan: WriteProto: error writing marshaled span: %v", err)
	}

	return nil
}

func (e *EagerSpan) DebugString() string {
	return e.Span.String()
}

// WriteAsAPITraces(w io.Writer) error
// 	WriteAsSpans(w io.Writer) error
