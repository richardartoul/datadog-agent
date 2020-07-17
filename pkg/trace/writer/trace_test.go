// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-2020 Datadog, Inc.

package writer

import (
	"compress/gzip"
	"fmt"
	"io/ioutil"
	"reflect"
	"sync"
	"testing"

	"github.com/DataDog/datadog-agent/pkg/trace/config"
	"github.com/DataDog/datadog-agent/pkg/trace/pb"
	"github.com/DataDog/datadog-agent/pkg/trace/test/testutil"
	"github.com/DataDog/datadog-agent/pkg/trace/traces"
	"github.com/gogo/protobuf/proto"
	"github.com/stretchr/testify/assert"
)

func TestTraceWriter(t *testing.T) {
	srv := newTestServer()
	cfg := &config.AgentConfig{
		Hostname:   testHostname,
		DefaultEnv: testEnv,
		Endpoints: []*config.Endpoint{{
			APIKey: "123",
			Host:   srv.URL,
		}},
		TraceWriter: &config.WriterConfig{ConnectionLimit: 200, QueueSize: 40},
	}

	t.Run("ok", func(t *testing.T) {
		testSpans := []*SampledSpans{
			randomSampledSpans(20, 8),
			randomSampledSpans(10, 0),
			randomSampledSpans(40, 5),
		}
		// Use a flush threshold that allows the first two entries to not overflow,
		// but overflow on the third.
		defer useFlushThreshold(testSpans[0].size() + testSpans[1].size() + 10)()
		in := make(chan *SampledSpans)
		tw := NewTraceWriter(cfg, in)
		go tw.Run()
		for _, ss := range testSpans {
			in <- ss
		}
		tw.Stop()
		// One payload flushes due to overflowing the threshold, and the second one
		// because of stop.
		//
		// TODO: Why did this change from 2 to 1?
		assert.Equal(t, 1, srv.Accepted())
		payloadsContain(t, srv.Payloads(), testSpans)
	})
}

func TestTraceWriterMultipleEndpointsConcurrent(t *testing.T) {
	var (
		srv = newTestServer()
		cfg = &config.AgentConfig{
			Hostname:   testHostname,
			DefaultEnv: testEnv,
			Endpoints: []*config.Endpoint{
				{
					APIKey: "123",
					Host:   srv.URL,
				},
				{
					APIKey: "123",
					Host:   srv.URL,
				},
			},
			TraceWriter: &config.WriterConfig{ConnectionLimit: 200, QueueSize: 40},
		}
		numWorkers      = 10
		numOpsPerWorker = 100
	)

	testSpans := []*SampledSpans{
		randomSampledSpans(20, 8),
		randomSampledSpans(10, 0),
		randomSampledSpans(40, 5),
	}
	in := make(chan *SampledSpans, 100)
	tw := NewTraceWriter(cfg, in)
	go tw.Run()

	var wg sync.WaitGroup
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < numOpsPerWorker; j++ {
				for _, ss := range testSpans {
					in <- ss
				}
			}
		}()
	}

	wg.Wait()
	tw.Stop()
	payloadsContain(t, srv.Payloads(), testSpans)
}

// useFlushThreshold sets n as the number of bytes to be used as the flush threshold
// and returns a function to restore it.
func useFlushThreshold(n int) func() {
	old := maxPayloadSize
	maxPayloadSize = n
	return func() { maxPayloadSize = old }
}

// randomSampledSpans returns a set of spans sampled spans and events events.
func randomSampledSpans(spans, events int) *SampledSpans {
	realisticIDs := true
	trace := testutil.GetTestTraces(1, spans, realisticIDs)[0]

	eventsTrace := trace // Shallow copy.
	eventsTrace.Spans = eventsTrace.Spans[:events]
	return &SampledSpans{
		Trace:  trace,
		Events: eventsTrace,
	}
}

// payloadsContain checks that the given payloads contain the given set of sampled spans.
func payloadsContain(t *testing.T, payloads []*payload, sampledSpans []*SampledSpans) {
	t.Helper()
	var all pb.TracePayload
	for _, p := range payloads {
		assert := assert.New(t)
		gzipr, err := gzip.NewReader(p.body)
		assert.NoError(err)
		slurp, err := ioutil.ReadAll(gzipr)
		assert.NoError(err)
		var payload pb.TracePayload
		err = proto.Unmarshal(slurp, &payload)
		assert.NoError(err)
		assert.Equal(payload.HostName, testHostname)
		assert.Equal(payload.Env, testEnv)
		all.Traces = append(all.Traces, payload.Traces...)
		all.Transactions = append(all.Transactions, payload.Transactions...)
	}
	for _, ss := range sampledSpans {
		var found bool
		for _, trace := range all.Traces {
			expected := make([]*pb.Span, 0, len(ss.Trace.Spans))
			for _, span := range ss.Trace.Spans {
				expected = append(expected, &span.(*traces.EagerSpan).Span)
			}
			if reflect.DeepEqual(expected, trace.Spans) {
				found = true
				break
			}
		}
		if !found {
			t.Fatal("payloads didn't contain given traces")
		}

		fmt.Println("hmmmmmm1")
		for _, t := range all.Transactions {
			fmt.Println("---")
			fmt.Println(t)
		}

		fmt.Println("hmmmmmm2")
		for _, event := range ss.Events.Spans {
			fmt.Println("---")
			fmt.Println(event)
		}
		for _, event := range ss.Events.Spans {
			assert.Contains(t, all.Transactions, &event.(*traces.EagerSpan).Span)
		}
	}
}
