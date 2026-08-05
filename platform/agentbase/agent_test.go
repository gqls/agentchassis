// FILE: platform/agentbase/agent_test.go
package agentbase

import (
	"context"
	"go/ast"
	"go/parser"
	"go/token"
	"testing"

	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/mock"
)

// MockKafkaConsumer for testing
type MockKafkaConsumer struct {
	mock.Mock
}

func (m *MockKafkaConsumer) FetchMessage(ctx context.Context) (kafka.Message, error) {
	args := m.Called(ctx)
	return args.Get(0).(kafka.Message), args.Error(1)
}

func (m *MockKafkaConsumer) CommitMessages(ctx context.Context, msgs ...kafka.Message) error {
	args := m.Called(ctx, msgs)
	return args.Error(0)
}

func (m *MockKafkaConsumer) Close() error {
	return nil
}

func TestAgentHandleMessage(t *testing.T) {
	// This test needs to be redesigned since handleMessage is private
	// and the Agent struct expects real Kafka connections
	// For now, we'll skip this test or make it integration-only
	t.Skip("Skipping unit test that requires real Kafka connections")
}

// messagingCallsIn returns every messaging.<Name> call made by real code in
// path, mapped to the functions that make it.
//
// It parses rather than greps, and that is the load-bearing choice. The AST is
// built with parser mode 0, so comments are not retained and a symbol named
// only in a comment CANNOT satisfy any assertion built on this. That is not
// hypothetical here: server.go:113 discusses `MatchedPermanentFailure` in prose,
// and a grep-based version of the guard below would have been green on that
// comment alone while the real call site had been reverted.
func messagingCallsIn(t *testing.T, path string) map[string][]string {
	t.Helper()

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, nil, 0)
	if err != nil {
		t.Fatalf("parse %s: %v", path, err)
	}

	found := map[string][]string{}
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Body == nil {
			continue
		}
		ast.Inspect(fn.Body, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			sel, ok := call.Fun.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			pkg, ok := sel.X.(*ast.Ident)
			if !ok || pkg.Name != "messaging" {
				return true
			}
			found[sel.Sel.Name] = append(found[sel.Sel.Name], fn.Name.Name)
			return true
		})
	}
	return found
}

// TestAgentbaseClassifiesThroughTheSharedPermanentClassifier is the lockstep
// half for bugs_closed/195: this layer must decide "permanent?" through
// messaging.MatchedPermanentFailure, never through the legacy substring helper.
//
// The drift it guards is real and silent: a typed WORKFLOW_INVALID would be
// dropped-and-recorded at one layer and retried at the other, decided purely by
// which path the message took — the class bugs_closed/034 closed.
//
// > **REPLACED 2026-08-05 (code-review F2).** The two tests that stood here
// > (`TestAgentbaseUsesSharedValidationNeedles`,
// > `TestAgentbaseUsesSharedPermanentClassifier`) called messaging functions
// > directly and touched NO agentbase symbol at all, so they were cross-package
// > duplicates of tests that already live in platform/messaging. Neither could
// > detect the drift its own comment claimed to guard: reverting agent.go:1192
// > to MatchedValidationNeedle left both green. The reviewer named the second;
// > the first was identical in shape and is gone for the same reason.
//
// This is a SOURCE-level guard, not a behavioural one — processMessage needs
// real Kafka and a real DB (see TestAgentHandleMessage above), so the call site
// is what can honestly be pinned here. The behaviour of the classifier itself
// is tested where it lives, in platform/messaging/validation_drop_test.go.
func TestAgentbaseClassifiesThroughTheSharedPermanentClassifier(t *testing.T) {
	calls := messagingCallsIn(t, "agent.go")

	if fns := calls["MatchedPermanentFailure"]; len(fns) == 0 {
		t.Error("agent.go no longer calls messaging.MatchedPermanentFailure anywhere — " +
			"the permanent/transient decision has left the shared seam (bugs_closed/195)")
	} else {
		t.Logf("MatchedPermanentFailure called in: %v", fns)
	}

	// The legacy substring helper misses the fleet's commonest permanent error
	// (WORKFLOW_INVALID — capital I, and "requires a topic" is not "is
	// required"). Classifying with it here is the exact revert that reinstates
	// the bug, so it is the mutation this test exists to catch.
	if fns := calls["MatchedValidationNeedle"]; len(fns) > 0 {
		t.Errorf("agent.go classifies with the legacy messaging.MatchedValidationNeedle in %v — "+
			"it misses typed WORKFLOW_INVALID, which is bugs_closed/195 reinstated", fns)
	}
}
