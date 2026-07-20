// FILE: platform/aiservice/stop_signal_test.go
//
// Every provider client must decode its own stop/finish signal and fail LOUD on
// a completion that was cut or declined. This file is the regression net for
// bugs_open/008 (stop_reason undecoded) and bugs_open/019 (a truncated reviewer
// voiding a whole council round), plus the CI guard the bug-historian raised
// twice during 008's council review and the owner accepted (008 §8, D1/D2):
// nothing structural stopped a FUTURE third adapter from shipping without the
// guard, so TestEveryProviderDecodesItsStopSignal fails CI if one does.
//
// The clients are constructed directly rather than through their New* functions:
// AnthropicClient hardcodes api.anthropic.com, so the only seam is its
// httpClient field, and an in-package test can reach it. That is deliberate —
// routing these tests through a config map would test the config plumbing, not
// the response decoding, which is where both bugs lived.
package aiservice

import (
	"bytes"
	"context"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
)

// stubTransport returns one canned response for any request.
type stubTransport struct {
	body   string
	status int
}

func (s stubTransport) RoundTrip(_ *http.Request) (*http.Response, error) {
	status := s.status
	if status == 0 {
		status = http.StatusOK
	}
	return &http.Response{
		StatusCode: status,
		Body:       io.NopCloser(bytes.NewBufferString(s.body)),
		Header:     make(http.Header),
	}, nil
}

func anthropicWith(body string) *AnthropicClient {
	return &AnthropicClient{
		apiKey:     "test-key",
		model:      "claude-sonnet-5",
		httpClient: &http.Client{Transport: stubTransport{body: body}},
	}
}

func ollamaWith(body string) *OllamaClient {
	return &OllamaClient{
		apiURL:     "http://stub",
		model:      "llama3",
		httpClient: &http.Client{Transport: stubTransport{body: body}},
	}
}

func TestAnthropicTruncationIsLoudAndKeepsPartial(t *testing.T) {
	const body = `{
		"content": [{"type": "text", "text": "the model got this far and"}],
		"stop_reason": "max_tokens",
		"usage": {"input_tokens": 10, "output_tokens": 8000}
	}`

	got, err := anthropicWith(body).GenerateText(context.Background(), "prompt", nil)
	if err == nil {
		t.Fatal("a completion cut at max_tokens must return an error, not a successful partial")
	}

	te, ok := IsTruncated(err)
	if !ok {
		t.Fatalf("want a *TruncatedError callers can detect by type, got %T: %v", err, err)
	}
	if te.Partial != "the model got this far and" {
		t.Errorf("partial destroyed at the transport layer — this is bugs_open/019's root cause; got %q", te.Partial)
	}
	if got != te.Partial {
		t.Errorf("returned text %q should match the carried partial %q", got, te.Partial)
	}
	if te.OutputTokens != 8000 {
		t.Errorf("OutputTokens = %d, want 8000", te.OutputTokens)
	}
	if te.Provider != "anthropic" {
		t.Errorf("Provider = %q, want anthropic", te.Provider)
	}
}

func TestAnthropicRefusalNamesItsCause(t *testing.T) {
	// The production shape from 2026-07-18: stop_reason=refusal, one non-text
	// block. It surfaced as "no text content in response (had 1 blocks)" and
	// killed an experience-planner council round.
	const body = `{
		"content": [{"type": "thinking", "text": ""}],
		"stop_reason": "refusal",
		"usage": {"input_tokens": 10, "output_tokens": 2}
	}`

	_, err := anthropicWith(body).GenerateText(context.Background(), "prompt", nil)
	if err == nil {
		t.Fatal("a refusal must be an error")
	}

	re, ok := IsRefusal(err)
	if !ok {
		t.Fatalf("want a *RefusalError callers can detect by type, got %T: %v", err, err)
	}
	if re.Reason != "stop_reason=refusal" {
		t.Errorf("Reason = %q, want the provider signal verbatim", re.Reason)
	}
	if len(re.Blocks) != 1 || re.Blocks[0] != "thinking" {
		t.Errorf("Blocks = %v, want [thinking] — the block TYPE is the diagnostic tell, not the count", re.Blocks)
	}
	if re.Model != "claude-sonnet-5" {
		t.Errorf("Model = %q, want the model that declined", re.Model)
	}

	// The message must send the reader to the prompt, not the parser.
	msg := err.Error()
	for _, want := range []string{"declined", "stop_reason=refusal", "thinking", "claude-sonnet-5"} {
		if !strings.Contains(msg, want) {
			t.Errorf("error message %q must contain %q", msg, want)
		}
	}
	// A refusal must NOT be mistaken for a truncation — they have opposite fixes.
	if _, isTrunc := IsTruncated(err); isTrunc {
		t.Error("a refusal must not classify as a truncation: one is fixed by raising a cap, the other by changing the prompt")
	}
}

func TestAnthropicNoTextBlockNamesStopReasonAndBlocks(t *testing.T) {
	// Not a refusal and not a cut — the genuine fallback. It must still say what
	// the response actually contained; "had 1 blocks" is what wasted the time.
	const body = `{
		"content": [{"type": "tool_use", "text": ""}],
		"stop_reason": "tool_use",
		"usage": {"input_tokens": 10, "output_tokens": 5}
	}`

	_, err := anthropicWith(body).GenerateText(context.Background(), "prompt", nil)
	if err == nil {
		t.Fatal("a response with no text block must error")
	}
	msg := err.Error()
	for _, want := range []string{"tool_use", "stop_reason"} {
		if !strings.Contains(msg, want) {
			t.Errorf("error message %q must contain %q", msg, want)
		}
	}
}

func TestAnthropicNormalCompletionSucceeds(t *testing.T) {
	const body = `{
		"content": [{"type": "text", "text": "a complete answer"}],
		"stop_reason": "end_turn",
		"usage": {"input_tokens": 10, "output_tokens": 4}
	}`

	got, err := anthropicWith(body).GenerateText(context.Background(), "prompt", nil)
	if err != nil {
		t.Fatalf("a normal completion must not error: %v", err)
	}
	if got != "a complete answer" {
		t.Errorf("got %q, want %q", got, "a complete answer")
	}
}

func TestAnthropicWritesUsageBackToOptions(t *testing.T) {
	const body = `{
		"content": [{"type": "text", "text": "ok"}],
		"stop_reason": "end_turn",
		"usage": {"input_tokens": 123, "output_tokens": 456}
	}`

	options := map[string]interface{}{}
	if _, err := anthropicWith(body).GenerateText(context.Background(), "prompt", options); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// llm_call_log's truncation detection joins on these; if they stop being
	// written the "output_tokens == max_tokens means CUT" rule goes blind.
	if options["__usage_input_tokens"] != 123 || options["__usage_output_tokens"] != 456 {
		t.Errorf("usage not written back for logging: %v", options)
	}
}

func TestOllamaTruncationIsLoudAndKeepsPartial(t *testing.T) {
	const body = `{
		"message": {"content": "cut off right"},
		"model": "llama3",
		"done_reason": "length",
		"prompt_eval_count": 10,
		"eval_count": 2048
	}`

	got, err := ollamaWith(body).GenerateText(context.Background(), "prompt", nil)
	if err == nil {
		t.Fatal("done_reason=length is a truncation and must error")
	}
	te, ok := IsTruncated(err)
	if !ok {
		t.Fatalf("want a *TruncatedError, got %T: %v", err, err)
	}
	if te.Partial != "cut off right" || got != "cut off right" {
		t.Errorf("partial not preserved: carried %q, returned %q", te.Partial, got)
	}
	if te.Provider != "ollama" {
		t.Errorf("Provider = %q, want ollama — a mixed-provider deployment must be able to tell where a cut came from", te.Provider)
	}
}

func TestOllamaNormalCompletionSucceeds(t *testing.T) {
	const body = `{
		"message": {"content": "a complete answer"},
		"model": "llama3",
		"done_reason": "stop",
		"prompt_eval_count": 10,
		"eval_count": 4
	}`

	got, err := ollamaWith(body).GenerateText(context.Background(), "prompt", nil)
	if err != nil {
		t.Fatalf("a normal completion must not error: %v", err)
	}
	if got != "a complete answer" {
		t.Errorf("got %q, want %q", got, "a complete answer")
	}
}

// TestEveryProviderDecodesItsStopSignal is the CI guard (008 §8, owner decision
// D1/D2). The bug-historian's residual objection was that fixing the two known
// adapters leaves nothing preventing a THIRD from shipping without the guard —
// production would find it, as it did twice already.
//
// So this asserts the property structurally rather than per-adapter: every type
// in this package with a GenerateText method must reference TruncatedError in
// its own file. A new provider that decodes no stop signal fails here.
func TestEveryProviderDecodesItsStopSignal(t *testing.T) {
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, ".", func(fi os.FileInfo) bool {
		return !strings.HasSuffix(fi.Name(), "_test.go")
	}, 0)
	if err != nil {
		t.Fatalf("parsing package: %v", err)
	}

	// receiver type name -> file that declares its GenerateText
	generators := map[string]string{}
	for _, pkg := range pkgs {
		for path, file := range pkg.Files {
			for _, decl := range file.Decls {
				fn, ok := decl.(*ast.FuncDecl)
				if !ok || fn.Name.Name != "GenerateText" || fn.Recv == nil || len(fn.Recv.List) == 0 {
					continue
				}
				generators[receiverTypeName(fn.Recv.List[0].Type)] = path
			}
		}
	}

	if len(generators) < 2 {
		t.Fatalf("expected at least the anthropic and ollama clients, found %d: %v", len(generators), generators)
	}

	for recv, path := range generators {
		src, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("reading %s: %v", path, err)
		}
		if !bytes.Contains(src, []byte("TruncatedError")) {
			t.Errorf(""+
				"provider %s (%s) implements GenerateText but never references TruncatedError.\n"+
				"Every provider client must decode its own stop/finish signal and fail loud on a cut\n"+
				"completion, carrying the partial — see platform/aiservice/truncation.go and\n"+
				"bugs_open/008. A silent truncation is indistinguishable from a complete answer at\n"+
				"every layer above this one.", recv, path)
		}
	}
}

func receiverTypeName(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.StarExpr:
		return receiverTypeName(t.X)
	case *ast.Ident:
		return t.Name
	}
	return ""
}
