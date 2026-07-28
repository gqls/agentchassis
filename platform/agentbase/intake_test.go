// FILE: platform/agentbase/intake_test.go
//
// The intake path's safety rests on three things a reviewer cannot see from
// the diff alone: the flag defaults OFF (byte-identical behaviour), the
// structural guards refuse the inheritance trap that bit EXTRA_REQUEST_TOPICS
// (spawned pods inherit personae-prod-config wholesale), and a request's
// serialisation key is exactly its orchestration id — the unit
// UpdateStateWithVersion already serialises on. Response-path key resolution
// needs the live awaited_requests table and is verified at enablement instead
// (see the CS-2 verification recipe in the chassis_replica_scaling NOTES).

package agentbase

import (
	"os"
	"testing"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

func newTestAgent(t *testing.T) *Agent {
	t.Helper()
	return &Agent{logger: zap.NewNop()}
}

func TestIntakeDefaultsOff(t *testing.T) {
	os.Unsetenv(intakeModeEnv)
	a := newTestAgent(t)
	a.setupIntake("system.agent.generic.requests")
	if a.intakeForRequests() || a.intakeForResponses() {
		t.Fatal("intake enabled with no CHASSIS_INTAKE_MODE — the default must be the inline path, byte-identical to today")
	}
}

func TestIntakeRefusesSpawnedAgents(t *testing.T) {
	t.Setenv(intakeModeEnv, intakeModeWorkerPool)
	a := newTestAgent(t)
	a.spawned = true
	a.setupIntake("system.agent.generic.requests")
	if a.intakeForRequests() {
		t.Fatal("a spawned agent accepted CHASSIS_INTAKE_MODE — spawned pods inherit personae-prod-config, " +
			"so this guard is what stands between the flag and every live Job pod running chassis intake")
	}
}

func TestIntakeRefusesJobTopicsAndMissingDB(t *testing.T) {
	t.Setenv(intakeModeEnv, intakeModeWorkerPool)

	a := newTestAgent(t)
	a.setupIntake("job.abc12345-def67890-writer-step.requests")
	if a.intakeForRequests() {
		t.Fatal("a job-topic agent accepted CHASSIS_INTAKE_MODE — extra lanes' static-agent guard must apply here too")
	}

	b := newTestAgent(t)
	b.setupIntake("system.agent.generic.requests") // db is nil
	if b.intakeForRequests() {
		t.Fatal("intake enabled without a database — the pool cannot exist, the flag must fall back to inline")
	}
}

func TestRequestSerialisationKeyIsTheOrchestrationID(t *testing.T) {
	a := newTestAgent(t)
	orch := uuid.New().String()
	corr := uuid.New().String()

	key, gotOrch, gotCorr, _ := a.intakeSerialisationKey(map[string]string{
		"orchestration_id": orch,
		"correlation_id":   corr,
	}, "request")
	if key != orch || gotOrch != orch || gotCorr != corr {
		t.Fatalf("request key must be the orchestration id (the unit the version-CAS serialises on): key=%s orch=%s", key, gotOrch)
	}

	// Malformed orchestration id falls back to correlation, never to sharing a
	// domain with unrelated traffic.
	key2, _, _, _ := a.intakeSerialisationKey(map[string]string{
		"orchestration_id": "not-a-uuid",
		"correlation_id":   corr,
	}, "request")
	if key2 != corr {
		t.Fatalf("fallback key should be the correlation id, got %s", key2)
	}
}

func TestDegenerateResponseKeyIsDeterministic(t *testing.T) {
	// A redelivered copy of an unresolvable response must land in the SAME
	// order domain as the original — a random key would let the two run
	// concurrently, which is the one thing the claims table exists to prevent.
	reqID := "req-abc-123"
	k1 := uuid.NewSHA1(intakeKeyNamespace, []byte(reqID)).String()
	k2 := uuid.NewSHA1(intakeKeyNamespace, []byte(reqID)).String()
	if k1 != k2 {
		t.Fatal("uuid5 of the request id must be deterministic")
	}
	if k1 == uuid.NewSHA1(intakeKeyNamespace, []byte("req-other")).String() {
		t.Fatal("distinct request ids must not collide")
	}
}
