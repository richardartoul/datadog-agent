// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-2020 Datadog, Inc.

package sampler

import (
	"hash/fnv"
	"sort"

	"github.com/DataDog/datadog-agent/pkg/trace/traces"
)

// Signature is a hash representation of trace or a service, used to identify
// simlar signatures.
type Signature uint64

// spanHash is the type of the hashes used during the computation of a signature
// Use FNV for hashing since it is super-cheap and we have no cryptographic needs
type spanHash uint32
type spanHashSlice []spanHash

func (p spanHashSlice) Len() int           { return len(p) }
func (p spanHashSlice) Less(i, j int) bool { return p[i] < p[j] }
func (p spanHashSlice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func sortHashes(hashes []spanHash)         { sort.Sort(spanHashSlice(hashes)) }

// computeSignatureWithRootAndEnv generates the signature of a trace knowing its root
// Signature based on the hash of (env, service, name, resource, is_error) for the root, plus the set of
// (env, service, name, is_error) of each span.
func computeSignatureWithRootAndEnv(trace traces.Trace, root traces.Span, env string) Signature {
	rootHash := computeSpanHash(root, env, true)
	spanHashes := make([]spanHash, 0, len(trace.Spans))

	for i := range trace.Spans {
		spanHashes = append(spanHashes, computeSpanHash(trace.Spans[i], env, false))
	}

	// Now sort, dedupe then merge all the hashes to build the signature
	sortHashes(spanHashes)

	last := spanHashes[0]
	traceHash := last ^ rootHash
	for i := 1; i < len(spanHashes); i++ {
		if spanHashes[i] != last {
			last = spanHashes[i]
			traceHash = spanHashes[i] ^ traceHash
		}
	}

	return Signature(traceHash)
}

// ServiceSignature represents a unique way to identify a service.
type ServiceSignature struct{ Name, Env string }

func (s *ServiceSignature) SafeForMap() ServiceSignature {
	safe := *s // Shallow clone.
	safe.Name = string([]byte(safe.Name))
	safe.Env = string([]byte(safe.Env))
	return safe
}

// Hash generates the signature of a trace with minimal information such as
// service and env, this is typically used by distributed sampling based on
// priority, and used as a key to store the desired rate for a given
// service,env tuple.
func (s ServiceSignature) Hash() Signature {
	h := fnv.New32a()
	h.Write([]byte(s.Name))
	h.Write([]byte{','})
	h.Write([]byte(s.Env))
	return Signature(h.Sum32())
}

func (s ServiceSignature) String() string {
	return "service:" + s.Name + ",env:" + s.Env
}

func computeSpanHash(span traces.Span, env string, withResource bool) spanHash {
	h := fnv.New32a()
	h.Write([]byte(env))
	h.Write([]byte(span.UnsafeService()))
	h.Write([]byte(span.UnsafeName()))
	h.Write([]byte{byte(span.Error())})
	if withResource {
		h.Write([]byte(span.UnsafeResource()))
	}

	// TODO: Fix me.
	// code, ok := traceutil.GetMeta(span, KeyHTTPStatusCode)
	// if ok {
	// 	h.Write([]byte(code))
	// }
	// typ, ok := traceutil.GetMeta(span, KeyErrorType)
	// if ok {
	// 	h.Write([]byte(typ))
	// }

	return spanHash(h.Sum32())
}
