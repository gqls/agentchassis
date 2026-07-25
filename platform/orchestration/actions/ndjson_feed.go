// FILE: platform/orchestration/actions/ndjson_feed.go
//
// The shared transport half of every cluster⇄external-VM pull.
//
// The cluster is never called by the public: where an external box holds data
// we want, the cluster PULLS it over a shared-secret header (the one-way
// direction is the security property — intent_collector_actions.go established
// it, the gripper-dossier island reuses it). Two call sites now do this, and
// the fetch-and-stream half is identical in both: GET with X-Internal-Key,
// require 200, scan an NDJSON body line by line with a raised buffer.
//
// What differs per caller — how the checkpoint and endpoint are built, and
// what each line is persisted into — stays with the caller. Extracting only
// the identical half was the point: a second copy of the transport is the
// duplication-by-imitation shape the council's reuse seat objected to
// (correlation 7ed137d1), and one guard raised here (buffer size, status
// check, partial-stream semantics) now protects both callers.

package actions

import (
	"bufio"
	"context"
	"fmt"
	"net/http"
	"time"
)

// ndjsonScanBufMax is the per-line ceiling. A report request's spec is far
// smaller, but a single oversized line must not silently truncate the stream:
// bufio.Scanner stops at the offending line and surfaces it through Err(),
// which reaches the caller as a partialStreamError rather than a short read
// that looks like a complete feed.
const (
	ndjsonScanBufInit = 64 * 1024
	ndjsonScanBufMax  = 1024 * 1024
)

// partialStreamError marks a feed that ended mid-stream. The lines already
// delivered are valid and were already persisted, so callers treat this as
// partial success and resume from their own checkpoint on the next run —
// distinct from a transport or status failure, where nothing was delivered
// and the whole site attempt failed.
type partialStreamError struct{ err error }

func (e *partialStreamError) Error() string { return "stream ended early: " + e.err.Error() }
func (e *partialStreamError) Unwrap() error { return e.err }

// scanInternalNDJSONFeed GETs endpoint with the X-Internal-Key shared secret
// and delivers each NDJSON line to onLine.
//
// Returns a *partialStreamError if the body ended mid-stream (lines delivered
// so far stand), or a plain error if the request or status failed (nothing was
// delivered). onLine's slice is only valid for the duration of the call —
// bufio reuses the buffer, so a caller retaining it must copy.
func scanInternalNDJSONFeed(
	ctx context.Context,
	endpoint, internalKey string,
	timeout time.Duration,
	onLine func(line []byte),
) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return err
	}
	req.Header.Set("X-Internal-Key", internalKey)

	resp, err := (&http.Client{Timeout: timeout}).Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("feed returned %d", resp.StatusCode)
	}

	sc := bufio.NewScanner(resp.Body)
	sc.Buffer(make([]byte, 0, ndjsonScanBufInit), ndjsonScanBufMax)
	for sc.Scan() {
		onLine(sc.Bytes())
	}
	if scErr := sc.Err(); scErr != nil {
		return &partialStreamError{err: scErr}
	}
	return nil
}
