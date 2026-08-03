// FILE: platform/aiservice/vision_test.go
//
// Wire-shape tests for the VisionCapable seam (vigilant_designer lane, A1.2).
// Each asserts what actually went on the wire — the request body — not merely
// that a call returned; a provider that silently dropped the images would pass
// any response-only test while critiquing pages it never saw.
package aiservice

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"testing"
)

// anthropicVisionWith mirrors stop_signal_test.go's anthropicWith but returns
// the capturing transport too — these tests assert the request BODY.
func anthropicVisionWith(body string) (*AnthropicClient, *capturingTransport) {
	tr := &capturingTransport{body: body}
	return &AnthropicClient{
		apiKey:     "test-key",
		model:      "claude-test",
		httpClient: &http.Client{Transport: tr},
	}, tr
}

const anthropicOKBody = `{
	"content": [{"type": "text", "text": "a critique"}],
	"stop_reason": "end_turn",
	"usage": {"input_tokens": 12, "output_tokens": 5}
}`

// Compile-time capability roster: the two critic-trial providers have eyes,
// and the assertion is here so removing an implementation fails the build,
// not a live critique run.
var (
	_ VisionCapable = (*AnthropicClient)(nil)
	_ VisionCapable = (*GeminiClient)(nil)
)

func TestOllamaIsNotVisionCapable(t *testing.T) {
	// The action treats a non-VisionCapable provider as a configuration error.
	// This pins that ollama stays on that path — if someone implements vision
	// for it, this test makes them say so deliberately.
	var svc AIService = &OllamaClient{}
	if _, ok := svc.(VisionCapable); ok {
		t.Fatal("OllamaClient unexpectedly claims VisionCapable — update the action's capability story deliberately")
	}
}

func TestAnthropicGenerateWithImagesWireShape(t *testing.T) {
	c, tr := anthropicVisionWith(anthropicOKBody)

	imgA := []byte("png-a")
	imgB := []byte("png-b")
	out, err := c.GenerateWithImages(context.Background(), "critique these pages",
		[]ImageInput{
			{MediaType: "image/png", Data: imgA},
			{MediaType: "image/png", Data: imgB},
		}, map[string]interface{}{"max_tokens": 512})
	if err != nil {
		t.Fatalf("GenerateWithImages: %v", err)
	}
	if out != "a critique" {
		t.Fatalf("response text lost: %q", out)
	}

	var sent struct {
		Messages []struct {
			Role    string `json:"role"`
			Content []struct {
				Type   string `json:"type"`
				Text   string `json:"text"`
				Source struct {
					Type      string `json:"type"`
					MediaType string `json:"media_type"`
					Data      string `json:"data"`
				} `json:"source"`
			} `json:"content"`
		} `json:"messages"`
	}
	if err := json.Unmarshal(tr.captured, &sent); err != nil {
		t.Fatalf("request body was not valid JSON: %v", err)
	}
	if len(sent.Messages) != 1 || sent.Messages[0].Role != "user" {
		t.Fatalf("want one user message, got %+v", sent.Messages)
	}
	blocks := sent.Messages[0].Content
	if len(blocks) != 3 {
		t.Fatalf("want image,image,text blocks, got %d", len(blocks))
	}
	if blocks[0].Type != "image" || blocks[1].Type != "image" || blocks[2].Type != "text" {
		t.Fatalf("block order wrong: %s,%s,%s", blocks[0].Type, blocks[1].Type, blocks[2].Type)
	}
	if blocks[0].Source.Type != "base64" || blocks[0].Source.MediaType != "image/png" {
		t.Fatalf("image source malformed: %+v", blocks[0].Source)
	}
	if blocks[0].Source.Data != base64.StdEncoding.EncodeToString(imgA) ||
		blocks[1].Source.Data != base64.StdEncoding.EncodeToString(imgB) {
		t.Fatal("image bytes did not survive to the wire in order")
	}
	if blocks[2].Text != "critique these pages" {
		t.Fatalf("prompt text lost: %q", blocks[2].Text)
	}
}

func TestGeminiGenerateWithImagesWireShape(t *testing.T) {
	c, tr := geminiWith("gemini-test", geminiOKBody)

	img := []byte("png-bytes")
	_, err := c.GenerateWithImages(context.Background(), "critique this page",
		[]ImageInput{{MediaType: "image/png", Data: img}},
		map[string]interface{}{"max_tokens": 512})
	if err != nil {
		t.Fatalf("GenerateWithImages: %v", err)
	}

	var sent struct {
		Contents []struct {
			Role  string `json:"role"`
			Parts []struct {
				Text       string `json:"text"`
				InlineData struct {
					MimeType string `json:"mime_type"`
					Data     string `json:"data"`
				} `json:"inline_data"`
			} `json:"parts"`
		} `json:"contents"`
	}
	if err := json.Unmarshal(tr.captured, &sent); err != nil {
		t.Fatalf("request body was not valid JSON: %v", err)
	}
	if len(sent.Contents) != 1 || len(sent.Contents[0].Parts) != 2 {
		t.Fatalf("want one turn with inline_data+text parts, got %+v", sent.Contents)
	}
	p := sent.Contents[0].Parts
	if p[0].InlineData.MimeType != "image/png" ||
		p[0].InlineData.Data != base64.StdEncoding.EncodeToString(img) {
		t.Fatalf("inline_data malformed: %+v", p[0].InlineData)
	}
	if p[1].Text != "critique this page" {
		t.Fatalf("prompt text lost: %q", p[1].Text)
	}
}

// The two tests below answer council fee9d810's gating objection
// (llm_reliability, high): "fails loud on truncation" was called an
// unsubstantiated completeness claim because no edit added stop_reason
// decoding. The decoding was not added because it already exists — the shared
// generate() arm each provider routes through (stop_reason=max_tokens /
// finishReason=MAX_TOKENS → *TruncatedError, shipped for bugs_open/019 in
// a3b606798) — and GenerateWithImages rides it by construction, both entry
// points delegating to the same generate(). The objection's other premise
// ("truncated calls have output_tokens=NULL, so a token-count heuristic can't
// work") is right and is WHY the mechanism is the provider's own stop signal,
// not a count. These tests pin the arm THROUGH the vision entry point, so the
// claim is measured rather than argued. The stale MDL-038 register entry that
// fed the objection ("GenerateText never decodes stop_reason", frozen
// 2026-07-17) is corrected in this commit.
func TestAnthropicVisionTruncationIsLoudAndKeepsPartial(t *testing.T) {
	c, _ := anthropicVisionWith(`{
		"content": [{"type": "text", "text": "partial crit"}],
		"stop_reason": "max_tokens",
		"usage": {"input_tokens": 12, "output_tokens": 512}
	}`)
	out, err := c.GenerateWithImages(context.Background(), "critique these pages",
		[]ImageInput{{MediaType: "image/png", Data: []byte("png")}},
		map[string]interface{}{"max_tokens": 512})
	if err == nil {
		t.Fatal("a vision completion cut at max_tokens must return an error")
	}
	te, ok := IsTruncated(err)
	if !ok {
		t.Fatalf("want *TruncatedError through the vision path, got %T: %v", err, err)
	}
	if te.Partial != "partial crit" {
		t.Fatalf("partial lost through the vision path: %q", te.Partial)
	}
	if out != "partial crit" {
		t.Fatalf("GenerateWithImages must surface the partial alongside the error, got %q", out)
	}
}

func TestGeminiVisionTruncationIsLoudAndKeepsPartial(t *testing.T) {
	const body = `{
		"candidates": [{
			"content": {"parts": [{"text": "partial crit"}]},
			"finishReason": "MAX_TOKENS"
		}],
		"usageMetadata": {"promptTokenCount": 900, "candidatesTokenCount": 12, "totalTokenCount": 912}
	}`
	c, _ := geminiWith("gemini-test", body)
	_, err := c.GenerateWithImages(context.Background(), "critique this page",
		[]ImageInput{{MediaType: "image/png", Data: []byte("png")}},
		map[string]interface{}{"max_tokens": 512})
	if err == nil {
		t.Fatal("a vision completion cut at MAX_TOKENS must return an error")
	}
	te, ok := IsTruncated(err)
	if !ok {
		t.Fatalf("want *TruncatedError through the vision path, got %T: %v", err, err)
	}
	if te.Provider != "gemini" {
		t.Fatalf("Provider = %q, want gemini", te.Provider)
	}
	if te.Partial != "partial crit" {
		t.Fatalf("partial lost through the vision path: %q", te.Partial)
	}
}

// The refactor moved GenerateText onto the shared generate path — this pins
// that a plain text call still sends STRING content (not a one-element block
// array), so no wire behaviour changed for every existing caller.
func TestAnthropicGenerateTextStillSendsStringContent(t *testing.T) {
	c, tr := anthropicVisionWith(anthropicOKBody)
	if _, err := c.GenerateText(context.Background(), "plain prompt", nil); err != nil {
		t.Fatalf("GenerateText: %v", err)
	}
	var sent struct {
		Messages []struct {
			Content interface{} `json:"content"`
		} `json:"messages"`
	}
	if err := json.Unmarshal(tr.captured, &sent); err != nil {
		t.Fatalf("bad body: %v", err)
	}
	if _, isString := sent.Messages[0].Content.(string); !isString {
		t.Fatalf("plain prompt no longer sent as string content: %T", sent.Messages[0].Content)
	}
}
