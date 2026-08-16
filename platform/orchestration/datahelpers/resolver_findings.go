// FILE: platform/orchestration/datahelpers/resolver_findings.go
//
// The persistence hook for the RFC_029 Phase 1 observation instrument.
//
// WHY THIS EXISTS (council run ae2a88a7 on the Phase 1 submission, gating
// objection from the reuse_agent seat, 2026-08-15 — and it was right): the two
// Phase 1 WARNs (`aggressive search: conflicting candidates` in
// findFieldRecursive, `aggressive search: explicit single-segment mapping
// bypassed` in ExtractActionInputs) are the WHOLE instrument the
// Phase-1→Phase-2 plan rests on — Phase 2 flips conflicts to refusal only on
// "zero conflict WARNs over a 48h+ window, or every observed pair explicitly
// mapped first" (RFC_029 §9 D2). As plain log lines they cannot be read after
// the fact: chassis pod log retention is measured at ~90 seconds (CTS-059), so
// a window that is only observable by tailing a pod live is not observable.
// The platform's remedy for exactly this already exists — agent_error_log via
// platform/orchestration/agenterrors — so the WARNs are ALSO persisted there.
//
// WHY A REGISTERED SINK RATHER THAN A DB HANDLE: findFieldRecursive and
// ExtractActionInputs carry no *sql.DB and no ctx, and threading one through
// every one of their ~115 call sites fleet-wide is not a change this lane will
// make. So the resolver reports each occurrence to a package-level recorder,
// nil by default — with none registered, behaviour is byte-identical to the
// log-only build (the default-OFF shape of the 2026-08-02 §2 ruling). The
// chassis registers one at startup, where the pool and pod identity live
// (platform/agentbase, initializeComponents); a binary that never registers
// stays log-only and says nothing false.
//
// EVERY occurrence is reported — no dedup, no sampling: frequency is the data.
// §9's disconfirmation clause ("a substantial population of conflict WARNs
// whose lucky winner is load-bearing") needs the population SIZE, and a
// deduped row cannot give it. The log lines stay: rows are the instrument, the
// lines are for live tailing.
//
// KNOWN LIMIT, by design: per-call identity (orchestration_id, step_name) is
// NOT reachable from the resolver without the threading above, so the rows
// carry pod-level attribution only (pod_name + agent_type, known at
// registration). Each row says so in its context so a reader is not left
// wondering whether the blank orchestration_id is a bug.
package datahelpers

import (
	"sync"

	"go.uber.org/zap"
)

// ErrorCodes the recorder rows carry (agent_error_log.error_code). SCREAMING
// SNAKE like the table's other codes; grep-able as a pair. The observation
// window's query is written against these two names — see RFC_029 §10.2 and
// the staged_component_build RUNBOOK.
const (
	// ResolverFindingConflictingCandidates: the whole-tree search found two or
	// more candidates for a field that DISAGREE. Phase 1 still resolves the
	// stable shallowest winner; Phase 2 will resolve nothing.
	ResolverFindingConflictingCandidates = "RESOLVER_CONFLICTING_CANDIDATES"
	// ResolverFindingMappingBypassed: a dotless single-segment config reference
	// was outvoted by the whole-tree search (the bugs_closed/213 §D class).
	// Observation only in Phase 1.
	ResolverFindingMappingBypassed = "RESOLVER_MAPPING_BYPASSED"
)

// resolverFindingIdentityScope is written into every row's context so the
// blank orchestration_id / step_name reads as a stated limit, not a defect.
const resolverFindingIdentityScope = "pod (per-run identity is not reachable from the resolver; attribute by pod_name + agent_type)"

// ResolverFinding is one occurrence of a Phase 1 observation. Message is the
// same text as the WARN line, so a row and a log line are joinable by eye.
type ResolverFinding struct {
	Code    string                 // one of the ResolverFinding* constants
	Field   string                 // the input field being resolved
	Message string                 // the WARN message, verbatim
	Context map[string]interface{} // the WARN's structured fields, plus identity_scope
}

// ResolverFindingRecorder receives every occurrence. It must be best-effort
// and must never change the resolver's disposition — a recorder that panics
// is recovered here, so it cannot; a recorder that blocks WILL slow the step
// (there is no goroutine in front of it), so keep it bounded (the chassis
// bridge writes synchronously under a 5s timeout, like the other
// agent_error_log recorders in agentbase).
type ResolverFindingRecorder func(ResolverFinding)

var (
	resolverFindingMu       sync.RWMutex
	resolverFindingRecorder ResolverFindingRecorder
)

// SetResolverFindingRecorder installs the sink. nil restores log-only. Called
// once at process startup by the chassis; tests install a fake and reset via
// t.Cleanup. Not the SQL — the datahelpers tests assert on a fake recorder,
// never on a mocked INSERT (the INSERT is agenterrors' contract, tested there).
func SetResolverFindingRecorder(r ResolverFindingRecorder) {
	resolverFindingMu.Lock()
	defer resolverFindingMu.Unlock()
	resolverFindingRecorder = r
}

// recordResolverFinding hands one occurrence to the installed recorder, if
// any. Recovers a recorder panic: the instrument must never take the resolver
// down with it. The identity_scope note is stamped here so no call site can
// forget it.
func recordResolverFinding(logger *zap.Logger, f ResolverFinding) {
	resolverFindingMu.RLock()
	r := resolverFindingRecorder
	resolverFindingMu.RUnlock()
	if r == nil {
		return
	}
	if f.Context == nil {
		f.Context = map[string]interface{}{}
	}
	f.Context["identity_scope"] = resolverFindingIdentityScope
	defer func() {
		if rec := recover(); rec != nil && logger != nil {
			logger.Warn("resolver finding recorder panicked — the resolver's answer stands, the row is lost",
				zap.String("code", f.Code),
				zap.String("field", f.Field),
				zap.Any("panic", rec))
		}
	}()
	r(f)
}
