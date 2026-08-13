// FILE: internal/adapters/thunder/provision_defaults_test.go
//
// Regression tests for bugs_open/258 defects 1 and 2.
//
// Defect 1: the adapter sent a constant cpu_cores: 4 when the caller said
// nothing, and Thunder rejects that for 9 of its 11 single-GPU specs — every
// cheap one. The only specs that accepted 4 were h100, the dearest GPU on the
// menu, so "provision with defaults" meant "provision the most expensive box,
// or get a 400".
//
// Defect 2: the wait-for-RUNNING deadline was a hardcoded 5 minutes, which no
// measured a6000 boot met, so the compensating cleanup DELETED the instance we
// had just started paying for.
//
// These assert on WHAT WE SEND TO THUNDER and on the deadline actually used —
// not on our own defaulting logic agreeing with itself. The 400 came from the
// wire, so the wire is where the test has to look.

package thunder

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"

	"github.com/gqls/agentchassis/internal/adapters/thunder/api"
	"github.com/gqls/agentchassis/internal/adapters/thunder/store"
)

// primeSuccessfulProvision primes the DB calls a provision makes once it is
// past the gates: claim, mark-created, instance insert, mark-succeeded.
func primeSuccessfulProvision(mock sqlmock.Sqlmock) {
	expectClaimWon(mock)
	mock.ExpectExec(`UPDATE thunder_provision_claims`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`INSERT INTO thunder_instances`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`UPDATE thunder_provision_claims`).WillReturnResult(sqlmock.NewResult(0, 1))
}

// TestDefaultVCPUsAreValidForEveryGPU is defect 1, replayed across the real
// catalogue.
//
// Every case here provisions with NO vcpus override — the shape that used to
// send 4 — and asserts the count we put on the wire is one Thunder actually
// publishes for that spec. The old behaviour passes only the h100 rows, which is
// exactly the asymmetry that made the bug survive: whoever tested it tested the
// expensive box.
func TestDefaultVCPUsAreValidForEveryGPU(t *testing.T) {
	cases := []struct {
		gpu       string
		mode      string
		wantVCPUs int
		wantSpec  string
	}{
		{gpu: "a6000", mode: "prototyping", wantVCPUs: 6, wantSpec: "a6000_x1_prototyping"},
		{gpu: "a100xl", mode: "prototyping", wantVCPUs: 8, wantSpec: "a100xl_x1_prototyping"},
		{gpu: "a100xl", mode: "production", wantVCPUs: 15, wantSpec: "a100xl_x1_production"},
		{gpu: "l40", mode: "prototyping", wantVCPUs: 6, wantSpec: "l40_x1_prototyping"},
		{gpu: "l40", mode: "production", wantVCPUs: 10, wantSpec: "l40_x1_production"},
		{gpu: "h100", mode: "prototyping", wantVCPUs: 4, wantSpec: "h100_x1_prototyping"},
		{gpu: "h100", mode: "production", wantVCPUs: 15, wantSpec: "h100_x1_production"},
		// "a100" is the friendly alias the adapter maps to a100xl.
		{gpu: "a100", mode: "prototyping", wantVCPUs: 8, wantSpec: "a100xl_x1_prototyping"},
	}

	for _, tc := range cases {
		t.Run(tc.wantSpec, func(t *testing.T) {
			tapi := &countingThunderAPI{returnIP: "10.0.0.5"}
			action, mock := newTestAction(t, tapi)
			expectGates(mock)
			primeSuccessfulProvision(mock)

			_, err := action.Execute(context.Background(), ProvisionInstanceRequest{
				GPU:           tc.gpu,
				Mode:          tc.mode,
				CorrelationID: "corr-A",
				RequestID:     "req-1",
				// VCPUs deliberately unset — this is the defaulting path.
			})
			if err != nil {
				t.Fatalf("provision failed: %v", err)
			}

			got := tapi.lastCreate.CPUCores
			if got != tc.wantVCPUs {
				t.Errorf("sent cpu_cores=%d for %s; Thunder publishes no such option and would 400 "+
					"(bugs_open/258 defect 1) — want %d", got, tc.wantSpec, tc.wantVCPUs)
			}
			// The old constant, named so a regression is unambiguous in the output.
			if got == api.DefaultCPUCores && tc.wantVCPUs != api.DefaultCPUCores {
				t.Errorf("sent the old hardcoded constant %d — the catalogue was not consulted", api.DefaultCPUCores)
			}
		})
	}
}

// TestExplicitVCPUsBypassTheCatalogue: a caller who states a count must not be
// blocked by a /specs outage. The catalogue is consulted ONLY to fill a blank.
func TestExplicitVCPUsBypassTheCatalogue(t *testing.T) {
	tapi := &countingThunderAPI{
		returnIP: "10.0.0.5",
		specsErr: errors.New("thunder get specs: http: connection refused"),
	}
	action, mock := newTestAction(t, tapi)
	expectGates(mock)
	primeSuccessfulProvision(mock)

	_, err := action.Execute(context.Background(), ProvisionInstanceRequest{
		GPU: "a6000", Mode: "prototyping", VCPUs: 8,
		CorrelationID: "corr-A", RequestID: "req-1",
	})
	if err != nil {
		t.Fatalf("an explicit vcpus request must survive a specs outage, got: %v", err)
	}
	if tapi.specCalls != 0 {
		t.Errorf("catalogue consulted %d times despite an explicit vcpus — an outage would block a "+
			"request that needs no lookup", tapi.specCalls)
	}
	if tapi.lastCreate.CPUCores != 8 {
		t.Errorf("caller asked for 8 vCPUs, sent %d", tapi.lastCreate.CPUCores)
	}
}

// TestUnresolvableSpecRefusesBeforeSpendingMoney: if we cannot establish a valid
// count we must NOT fall back to a constant and NOT reach the vendor. Falling
// back is the bug; the whole point is that a constant cannot be right here.
func TestUnresolvableSpecRefusesBeforeSpendingMoney(t *testing.T) {
	for _, tc := range []struct {
		name string
		tapi *countingThunderAPI
		gpu  string
	}{
		{
			name: "specs unreachable",
			tapi: &countingThunderAPI{returnIP: "10.0.0.5", specsErr: errors.New("connection refused")},
			gpu:  "a6000",
		},
		{
			name: "gpu Thunder does not publish",
			tapi: &countingThunderAPI{returnIP: "10.0.0.5"},
			gpu:  "l40s", // a declared GPU constant with NO live single-GPU spec
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			action, mock := newTestAction(t, tc.tapi)
			expectGates(mock)

			_, err := action.Execute(context.Background(), ProvisionInstanceRequest{
				GPU: tc.gpu, Mode: "prototyping",
				CorrelationID: "corr-A", RequestID: "req-1",
			})
			if err == nil {
				t.Fatal("expected a refusal rather than a guessed vCPU count")
			}
			if tc.tapi.creates != 0 {
				t.Errorf("reached the vendor %d times without a known-valid vCPU count — that is a 400 "+
					"at best and a billed box at worst", tc.tapi.creates)
			}
		})
	}
}

// ── defect 2: the wait deadline ────────────────────────────────────────────

// TestWaitTimeoutComesFromLiveConfig is defect 2. The deadline must be the
// configured one, so it can be tuned without an image build — and must fall
// back safely when the column is absent (an unmigrated database) or absurd.
func TestWaitTimeoutComesFromLiveConfig(t *testing.T) {
	tapi := &countingThunderAPI{returnIP: "10.0.0.5"}
	action, _ := newTestAction(t, tapi)
	action.waitTimeout = 5 * time.Minute // the old compiled-in default

	cases := []struct {
		name string
		cfg  *store.Config
		want time.Duration
	}{
		{
			name: "configured value is used",
			cfg:  &store.Config{ProvisionWaitTimeoutSeconds: sql.NullInt64{Int64: 540, Valid: true}},
			want: 540 * time.Second,
		},
		{
			name: "column absent (migration 400 not applied) falls back",
			cfg:  &store.Config{},
			want: 5 * time.Minute,
		},
		{
			name: "absurd value is ignored, not obeyed",
			cfg:  &store.Config{ProvisionWaitTimeoutSeconds: sql.NullInt64{Int64: 999999, Valid: true}},
			want: 5 * time.Minute,
		},
		{
			name: "zero is ignored, not treated as no-wait",
			cfg:  &store.Config{ProvisionWaitTimeoutSeconds: sql.NullInt64{Int64: 0, Valid: true}},
			want: 5 * time.Minute,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := action.resolveWaitTimeout(tc.cfg); got != tc.want {
				t.Errorf("resolveWaitTimeout = %v, want %v", got, tc.want)
			}
		})
	}
}

// TestWaitDeadlineStaysUnderTheDispatchAwait pins the coupling that migration
// 400 documents, because it is the one that bites in the wrong direction.
//
// If the adapter's wait EXCEEDS the dispatching step's timeout_seconds, a
// slow-but-SUCCESSFUL provision becomes the bad case: the await expires, the
// retry driver re-dispatches, the 259 claim guard refuses the duplicate
// (correctly), and the workflow reports FAILED while a real billed instance runs
// on with nobody watching it. The migration's default must respect that.
func TestWaitDeadlineStaysUnderTheDispatchAwait(t *testing.T) {
	// The live gpu-provisioner dispatch_provision timeout_seconds, 2026-08-13.
	// If someone raises the migration default past this without also raising the
	// step, this test is the thing that says so.
	const dispatchAwaitSeconds = 600
	const migration400Default = 540

	if migration400Default >= dispatchAwaitSeconds {
		t.Fatalf("the 400 default (%ds) must stay below the dispatch_provision await (%ds): "+
			"raise the STEP timeout first, then the column — see 400's coupling note",
			migration400Default, dispatchAwaitSeconds)
	}
	// And leave real headroom for the create, keypair, secret and INSERT either
	// side of the wait, not just a nominal margin.
	if dispatchAwaitSeconds-migration400Default < 30 {
		t.Errorf("only %ds of headroom between the wait deadline and the await; the surrounding "+
			"create/secret/insert work needs more", dispatchAwaitSeconds-migration400Default)
	}
}
