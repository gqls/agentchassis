// FILE: internal/adapters/thunder/provision_idempotency_test.go
//
// Regression test for bugs_open/259 — resolve that number by SLUG,
// "..._one_provision_request_builds_several_billable_gpus".
//
// WHAT IT PROVES, and why it is shaped this way.
//
// The defect: the chassis retry driver re-executes dispatch_provision when its
// 600s await expires (coordinator.go retryExpiredAwaitedRequest, budget
// RetryVersion < 3), minting a FRESH request_id each time. One logical request
// therefore arrives at the adapter as up to four independent provision
// messages, alike in everything except request_id — and each one used to build
// a billable GPU. Measured on 2026-08-12: orchestration 8c5bf926, 4 awaited
// rows, 4 distinct request_ids, one correlation_id.
//
// So the test replays THAT shape: the same correlation, a different
// request_id, exactly as the retry driver produces it. A test that reused the
// request_id would pass against a broken build, because it would collide on an
// identifier the real retries never share.
//
// The assertion is on CREATES AT THE VENDOR, not on rows or errors: the vendor
// call is the thing that costs money, and it is the only fact that cannot be
// satisfied by bookkeeping that merely looks right.

package thunder

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"go.uber.org/zap"

	"github.com/gqls/agentchassis/internal/adapters/thunder/api"
	"github.com/gqls/agentchassis/internal/adapters/thunder/ssh"
	"github.com/gqls/agentchassis/internal/adapters/thunder/store"
)

// ── test doubles ────────────────────────────────────────────────────────

// countingThunderAPI records how many instances were actually created.
type countingThunderAPI struct {
	creates  int
	deletes  int
	nextID   int
	waitErr  error // if set, WaitForRunning fails — the real 258-defect-2 shape
	returnIP string

	specCalls  int                       // how many times the catalogue was consulted
	specsErr   error                     // if set, GetSpecs fails
	lastCreate api.CreateInstanceRequest // what we actually asked Thunder for
}

func (c *countingThunderAPI) CreateInstance(_ context.Context, req api.CreateInstanceRequest) (*api.CreateInstanceResponse, error) {
	c.creates++
	c.lastCreate = req
	c.nextID++
	return &api.CreateInstanceResponse{Identifier: c.nextID, UUID: "uuid-stub"}, nil
}

func (c *countingThunderAPI) WaitForRunning(_ context.Context, id int, _ time.Duration) (*api.Instance, error) {
	if c.waitErr != nil {
		return nil, c.waitErr
	}
	return &api.Instance{IP: c.returnIP, Port: 22}, nil
}

func (c *countingThunderAPI) DeleteInstance(_ context.Context, _ int) error {
	c.deletes++
	return nil
}

// GetSpecs returns the REAL catalogue, measured live from GET /v1/specs on
// 2026-08-13 (single-GPU entries). Invented numbers would defeat the purpose:
// bugs_open/258 defect 1 is precisely a case where a plausible-looking constant
// was wrong, so the fixture has to be what Thunder actually publishes.
func (c *countingThunderAPI) GetSpecs(_ context.Context) (map[string]api.Spec, error) {
	c.specCalls++
	if c.specsErr != nil {
		return nil, c.specsErr
	}
	mk := func(mode string, opts ...int) api.Spec {
		return api.Spec{GPUCount: 1, Mode: mode, VCPUOptions: opts}
	}
	return map[string]api.Spec{
		"a100xl_x1":             mk("", 8, 12, 16),
		"a100xl_x1_prototyping": mk("prototyping", 8, 12, 16),
		"a100xl_x1_production":  mk("production", 15),
		"a6000_x1":              mk("", 6, 8),
		"a6000_x1_prototyping":  mk("prototyping", 6, 8),
		"h100_x1":               mk("", 4, 8, 12, 16),
		"h100_x1_prototyping":   mk("prototyping", 4, 8, 12, 16),
		"h100_x1_production":    mk("production", 15),
		"l40_x1":                mk("", 6, 8, 12),
		"l40_x1_prototyping":    mk("prototyping", 6, 8, 12),
		"l40_x1_production":     mk("production", 10),
	}, nil
}

type stubSecretManager struct{}

func (stubSecretManager) CreateKeypairSecret(_ context.Context, uuid string, _ *ssh.Keypair) (string, error) {
	return "secret-" + uuid, nil
}
func (stubSecretManager) DeleteKeypairSecret(_ context.Context, _ string) error { return nil }

// ── DB expectation helpers ──────────────────────────────────────────────

// expectGates primes the two reads every Execute makes before it can reach
// the vendor: thunder_config and thunder_provision_check.
func expectGates(mock sqlmock.Sqlmock) {
	mock.ExpectQuery(`FROM thunder_config`).WillReturnRows(
		sqlmock.NewRows([]string{
			"daily_cap_usd", "max_concurrent_instances", "default_hard_uptime_hours",
			"default_hourly_rate_usd", "estimated_new_run_cost_usd", "is_paused", "pause_reason",
			"provision_wait_timeout_seconds",
		}).AddRow(100.0, 4, 18, 0.43, 5.0, false, nil, 540))

	mock.ExpectQuery(`FROM thunder_provision_check`).WillReturnRows(
		sqlmock.NewRows([]string{
			"can_provision", "denial_reason", "daily_cap_usd",
			"max_concurrent_instances", "total_24h_spend", "active_count",
		}).AddRow(true, nil, 100.0, 4, 0.0, 0))
}

// expectClaimWon primes TakeProvisionClaim's INSERT ... RETURNING for the
// winning attempt: xmax = 0 means the row was INSERTed, not updated.
func expectClaimWon(mock sqlmock.Sqlmock) {
	mock.ExpectQuery(`INSERT INTO thunder_provision_claims`).WillReturnRows(
		sqlmock.NewRows([]string{
			"correlation_id", "attempts", "status",
			"thunder_instance_id", "provisioning_id", "last_error", "inserted",
		}).AddRow("corr-A", 1, "claimed", nil, nil, nil, true))
}

// expectClaimLost primes the SAME statement for a losing attempt: the
// ON CONFLICT arm fired, attempts was bumped, and xmax != 0 so inserted=false.
// This is exactly what Postgres returns for the retry driver's second message.
func expectClaimLost(mock sqlmock.Sqlmock, attempts int, status string) {
	mock.ExpectQuery(`INSERT INTO thunder_provision_claims`).WillReturnRows(
		sqlmock.NewRows([]string{
			"correlation_id", "attempts", "status",
			"thunder_instance_id", "provisioning_id", "last_error", "inserted",
		}).AddRow("corr-A", attempts, status, "7", nil, nil, false))
}

func newTestAction(t *testing.T, tapi thunderAPI) (*ProvisionAction, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New(
		sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp),
		sqlmock.MonitorPingsOption(false),
	)
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	mock.MatchExpectationsInOrder(false)

	a := NewProvisionAction(tapi, stubSecretManager{}, db, zap.NewNop())
	a.pollInterval = time.Millisecond
	a.waitTimeout = 50 * time.Millisecond
	a.dbInsertBackoffs = []time.Duration{time.Millisecond, time.Millisecond, time.Millisecond}
	return a, mock
}

// ── the regression ──────────────────────────────────────────────────────

// TestRetryDriverRedispatchBuildsOnlyOneInstance is the bug, replayed.
//
// Two provision messages, same correlation_id, DIFFERENT request_ids — the
// exact shape the chassis retry driver produces on await expiry. Before the
// fix this produced two billable GPUs; it must now produce one.
func TestRetryDriverRedispatchBuildsOnlyOneInstance(t *testing.T) {
	tapi := &countingThunderAPI{returnIP: "10.0.0.5"}
	action, mock := newTestAction(t, tapi)

	// Attempt 1 — the winning claim, running through to a successful insert.
	expectGates(mock)
	expectClaimWon(mock)
	mock.ExpectExec(`UPDATE thunder_provision_claims`).
		WillReturnResult(sqlmock.NewResult(0, 1)) // MarkClaimCreated
	mock.ExpectExec(`INSERT INTO thunder_instances`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Attempt 2 — the retry driver's re-dispatch. It must get no further than
	// the claim, so NO further gate/vendor/insert expectations are primed.
	expectGates(mock)
	expectClaimLost(mock, 2, "succeeded")

	base := ProvisionInstanceRequest{
		GPU:           "a6000",
		VCPUs:         6,
		CorrelationID: "corr-A",
	}

	first := base
	first.RequestID = "req-1"
	if _, err := action.Execute(context.Background(), first); err != nil {
		t.Fatalf("first provision should succeed, got: %v", err)
	}

	// Same correlation, fresh request_id — the retry driver's signature.
	second := base
	second.RequestID = "req-2-fresh-from-retry-driver"
	_, err := action.Execute(context.Background(), second)

	if err == nil {
		t.Fatal("second provision returned success — the duplicate was NOT refused")
	}
	if !errors.Is(err, ErrProvisionDuplicate) {
		t.Errorf("second provision must fail with ErrProvisionDuplicate so the adapter can mark it\n"+
			"unrecoverable; got a different error: %v", err)
	}

	// The assertion that matters: one request, one billable box.
	if tapi.creates != 1 {
		t.Errorf("CreateInstance called %d times for one logical request — this is bugs_open/259 "+
			"(each extra call is a billable GPU); want exactly 1", tapi.creates)
	}
}

// TestProvisionRefusedWithoutCorrelationID: no key means no dedup is possible,
// so the request must be refused rather than provisioned unguarded. Fails
// closed — this path spends money.
func TestProvisionRefusedWithoutCorrelationID(t *testing.T) {
	tapi := &countingThunderAPI{returnIP: "10.0.0.5"}
	action, mock := newTestAction(t, tapi)
	expectGates(mock)

	_, err := action.Execute(context.Background(), ProvisionInstanceRequest{
		GPU: "a6000", VCPUs: 6, CorrelationID: "",
	})
	if err == nil {
		t.Fatal("a provision with no correlation_id must be refused")
	}
	if tapi.creates != 0 {
		t.Errorf("an unkeyed request reached the vendor %d times; want 0", tapi.creates)
	}
}

// TestFailedProvisionKeepsClaimAndRecordsError pins the deliberate choice that
// a FAILED attempt keeps its claim.
//
// This is the money case: on 2026-08-12 every attempt failed (an a6000 does not
// boot inside the adapter's 5-minute waitTimeout — bugs_open/258 defect 2), the
// compensating cleanup deleted each box, and the retry driver came straight
// back. If a failure released the claim, that loop would still run and this fix
// would be theatre. So: failure marks the claim failed and HOLDS it.
func TestFailedProvisionKeepsClaimAndRecordsError(t *testing.T) {
	tapi := &countingThunderAPI{
		returnIP: "10.0.0.5",
		waitErr:  context.DeadlineExceeded, // the real 258-defect-2 failure
	}
	action, mock := newTestAction(t, tapi)

	expectGates(mock)
	expectClaimWon(mock)
	mock.ExpectExec(`UPDATE thunder_provision_claims`).
		WillReturnResult(sqlmock.NewResult(0, 1)) // MarkClaimCreated
	// The failure record. Asserting the statement is reached is the point:
	// without it a failed provision leaves nothing behind (258 defect 3).
	markFailed := mock.ExpectExec(`UPDATE thunder_provision_claims SET status = 'failed'`).
		WillReturnResult(sqlmock.NewResult(0, 1))
	_ = markFailed

	req := ProvisionInstanceRequest{GPU: "a6000", VCPUs: 6, CorrelationID: "corr-A", RequestID: "req-1"}
	if _, err := action.Execute(context.Background(), req); err == nil {
		t.Fatal("provision should have failed when WaitForRunning fails")
	}

	// The box that was built must have been cleaned up...
	if tapi.deletes != 1 {
		t.Errorf("compensating cleanup ran %d times; want 1 (a live box was left billing)", tapi.deletes)
	}
	// ...and the failure must be on record.
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("the failed provision did not leave its durable record: %v", err)
	}
}

// TestClaimStatementIsSingleRoundTrip guards the property the dedup rests on:
// the claim must be ONE statement that both inserts and counts. A read-then-
// insert pair would reintroduce the race — two retry messages arriving close
// together would both read "no claim" and both provision.
func TestClaimStatementIsSingleRoundTrip(t *testing.T) {
	// The regexp is the assertion: an INSERT that carries its own conflict
	// arm, rather than a SELECT followed by an INSERT.
	re := regexp.MustCompile(`(?is)INSERT INTO thunder_provision_claims.*ON CONFLICT \(correlation_id\) DO UPDATE.*RETURNING`)
	if !re.MatchString(store.ClaimInsertSQL) {
		t.Error("TakeProvisionClaim must claim and count in a single INSERT ... ON CONFLICT ... RETURNING; " +
			"a read-then-write pair reopens the race this table exists to close")
	}
}

// TestDuplicateRefusalIsAlwaysUnrecoverable pins the classification invariant
// the council's edit-quality and guardian seats both flagged as untested.
//
// Why it needs its own test: the idempotency test above counts vendor creates,
// and it would keep passing even if the duplicate refusal were answered
// `error_recoverable` — the claim would still refuse that one attempt. What
// would change is that the CHASSIS would retry, and the whole fix rests on the
// chassis treating this as terminal. So the response classification has to be
// pinned independently of the refusal.
//
// A wrapped error is used deliberately: Execute returns the sentinel wrapped in
// context (%w), so a classifier written with == instead of errors.Is would pass
// a bare-sentinel test and fail in production.
func TestDuplicateRefusalIsAlwaysUnrecoverable(t *testing.T) {
	wrapped := fmt.Errorf("%w: correlation corr-A already attempted a provision (attempt 2, status failed) — refusing to build a second billable instance",
		ErrProvisionDuplicate)

	errCode, status := classifyProvisionError(wrapped)

	if status != "error_unrecoverable" {
		t.Errorf("a duplicate refusal was classified %q — the chassis would RETRY it, and the retry is what builds the second billable GPU (bugs_open/259); want error_unrecoverable", status)
	}
	if errCode != "provision_duplicate" {
		t.Errorf("errCode = %q, want provision_duplicate so an operator can see WHY it refused rather than a generic failure", errCode)
	}
}

// TestProvisionErrorClassificationOrder guards the ordering itself. The
// duplicate arm must win over the infrastructure arm, which is the specific
// regression the extraction exists to make visible: a context.DeadlineExceeded
// wrapped alongside the duplicate sentinel must still classify as a duplicate,
// because isInfrastructureError would otherwise claim it and mark it retryable.
func TestProvisionErrorClassificationOrder(t *testing.T) {
	cases := []struct {
		name       string
		err        error
		wantCode   string
		wantStatus string
	}{
		{
			name:       "duplicate beats infrastructure",
			err:        fmt.Errorf("%w: and the ctx also died: %w", ErrProvisionDuplicate, context.DeadlineExceeded),
			wantCode:   "provision_duplicate",
			wantStatus: "error_unrecoverable",
		},
		{
			name:       "paused is a denial, not retryable",
			err:        errors.New("thunder provisioning paused: phase0 halt"),
			wantCode:   "provision_denied",
			wantStatus: "error_unrecoverable",
		},
		{
			name:       "a genuine transient stays retryable",
			err:        fmt.Errorf("wait for instance running: %w", context.DeadlineExceeded),
			wantCode:   "infrastructure_error",
			wantStatus: "error_recoverable",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			code, status := classifyProvisionError(tc.err)
			if code != tc.wantCode || status != tc.wantStatus {
				t.Errorf("got (%q, %q), want (%q, %q)", code, status, tc.wantCode, tc.wantStatus)
			}
		})
	}
}
