package agentbase

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	segkafka "github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

// bugs_open/158 item 1 — owner ruling 2026-08-03, option (b).
//
// Added in answer to the council's editquality seat (corr f13212d4, medium):
// the submission invoked 158's landmine — that adoption tests must live in the
// CALLER's package because FailedUndeliverable is a return value a caller can
// ignore — and then tested only the coordinator. This is the agentbase half.
//
// What makes this site the worst of the three: the produce error was never
// CAPTURED at all, so an error travelling up the spawn chain could vanish at any
// hop with no record anywhere, and the message being dropped is itself an error
// report.

type forwardProducer struct {
	maxBytes int
	accepted [][]byte
	refused  int
}

func (p *forwardProducer) Produce(ctx context.Context, topic string, headers map[string]string, key, value []byte) error {
	if len(value) > p.maxBytes {
		p.refused++
		return segkafka.MessageSizeTooLarge
	}
	p.accepted = append(p.accepted, value)
	return nil
}

func (p *forwardProducer) ProduceWithValidation(ctx context.Context, topic string, headers map[string]string, key, value []byte) error {
	return p.Produce(ctx, topic, headers, key, value)
}

func (p *forwardProducer) Close() error { return nil }

func spawnedAgent(prod *forwardProducer) *Agent {
	return &Agent{
		AgentType:     "child-agent",
		ParentAgentID: "parent-1",
		spawned:       true,
		producer:      prod,
		logger:        zap.NewNop(),
	}
}

// TestErrorForwardDegradesRatherThanVanishing — an error too large to forward
// intact must still reach the parent as an error. Before the fix the produce
// error was not even looked at, so the whole chain ended silently at this hop.
func TestErrorForwardDegradesRatherThanVanishing(t *testing.T) {
	t.Setenv("PARENT_RESPONSES_TOPIC", "system.agent.parent.responses")

	prod := &forwardProducer{maxBytes: 2048}
	a := spawnedAgent(prod)

	big := make([]byte, 8192)
	for i := range big {
		big[i] = 'x'
	}
	msg := segkafka.Message{
		Key:     []byte("corr-158"),
		Value:   big,
		Headers: errHeaders("corr-158"),
	}
	a.processMessage(msg, "request")

	if prod.refused == 0 {
		t.Fatalf("test is vacuous — the oversized forward was not refused, so the degrade branch never ran")
	}
	if len(prod.accepted) == 0 {
		t.Fatalf("the parent was told NOTHING: %d refused, no degraded form sent — "+
			"the error chain ended here silently, which is the defect", prod.refused)
	}

	var got map[string]interface{}
	if err := json.Unmarshal(prod.accepted[len(prod.accepted)-1], &got); err != nil {
		t.Fatalf("the degraded forward must be valid JSON the parent can read: %v", err)
	}
	// The signal that must survive: this is still an ERROR, and it says the
	// payload was dropped rather than passing a stub off as the original.
	if got["success"] != false {
		t.Errorf("the degraded forward must still read as a failure; got %v", got)
	}
	blob := string(prod.accepted[len(prod.accepted)-1])
	if !strings.Contains(blob, "ERROR_FORWARD_TOO_LARGE") {
		t.Errorf("the degraded forward must name WHY it is not the original; got: %.300s", blob)
	}
}

// TestErrorForwardUnchangedWhenItFits is the negative control: a normal-sized
// error forward must go through byte-for-byte as before, or every spawned
// agent's error path changes shape.
func TestErrorForwardUnchangedWhenItFits(t *testing.T) {
	t.Setenv("PARENT_RESPONSES_TOPIC", "system.agent.parent.responses")

	prod := &forwardProducer{maxBytes: 1 << 20}
	a := spawnedAgent(prod)

	original := []byte(`{"success":false,"error":{"code":"UPSTREAM","message":"the real error"}}`)
	a.processMessage(segkafka.Message{Key: []byte("corr-ok"), Value: original, Headers: errHeaders("corr-ok")}, "request")

	if prod.refused != 0 {
		t.Fatalf("nothing should have been refused at 1MB; refused=%d", prod.refused)
	}
	if len(prod.accepted) != 1 {
		t.Fatalf("exactly one forward expected, got %d", len(prod.accepted))
	}
	if string(prod.accepted[0]) != string(original) {
		t.Errorf("a forward that FITS must be passed through unchanged\n got: %s\nwant: %s",
			prod.accepted[0], original)
	}
	if strings.Contains(string(prod.accepted[0]), "ERROR_FORWARD_TOO_LARGE") {
		t.Errorf("a deliverable forward must not be marked as degraded")
	}
}

// errHeaders is the minimum that routes a message down the "error REQUEST — pass
// it up to the parent" branch: is_error true, and a message type that is not a
// response (a response is deliberately routed to the coordinator instead).
func errHeaders(corr string) []segkafka.Header {
	return []segkafka.Header{
		{Key: "is_error", Value: []byte("true")},
		{Key: "correlation_id", Value: []byte(corr)},
		{Key: "from_agent_type", Value: []byte("upstream-agent")},
		{Key: "error_chain", Value: []byte("upstream-agent")},
	}
}
