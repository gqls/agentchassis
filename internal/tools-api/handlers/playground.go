// FILE: internal/tools-api/handlers/playground.go
//
// The playground: finetuning.uk's public demo chat against a self-hosted
// open-weight model (finetuning_uk_service PLAN Phase P; owner decision
// 2026-09-03: a public demo on the in-cluster Ollama, plus booked GPU hours
// later). This is the demo half. It is deliberately stateless — no session
// row, no transcript — because the visitor's browser holds the conversation
// and sends it back each turn, capped by MaxMessages. The booked-hours half
// will add a /session that routes to a provisioned box; this handler's request
// shape gains an optional session_id then, additively.
//
// Why it streams: the demo runs on CPU at ~14 tokens/s (measured 2026-09-03,
// NOTES). A 150-token reply is ~11 s. Delivered as one JSON body that is a
// stall; delivered as server-sent events it is a reply appearing at reading
// speed. So the response is text/event-stream: `event: token` per fragment,
// `event: done` with the model's own counts, `event: error` if the model
// stops mid-reply. Errors BEFORE the first token are ordinary JSON errors
// (the status code can still be set); after it, the status is already 200
// and the error travels in-band.
//
// Why the model server is not aiservice.OllamaClient: that client wraps
// /api/generate non-streaming for the chassis's own use. The demo needs
// /api/chat with a message list and a streaming reader, and the whole point
// of the route is to hand tokens on as they arrive.
package handlers

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
	"github.com/gqls/agentchassis/internal/tools-api/config"
	"github.com/gqls/agentchassis/internal/tools-api/httperr"
)

const (
	// PlaygroundMaxMessages bounds the transcript the browser may send back.
	PlaygroundMaxMessages = 12
	// PlaygroundMaxMessageRunes bounds one message. The body cap (config) is
	// the byte-level backstop; this is the per-message limit a visitor sees.
	PlaygroundMaxMessageRunes = 1000
	// playgroundCallTimeout covers model load (2.4 s cold, measured) plus a
	// full-length reply at CPU speed with headroom for a busy node.
	playgroundCallTimeout = 90 * time.Second
	// playgroundScanBuf is the largest single NDJSON line accepted from the
	// model server; a token fragment is tens of bytes.
	playgroundScanBuf = 1 << 20
)

// PlaygroundSystemPrompt tells the demo model what it is, in the site's own
// register. Fixed text on purpose: the demo is not LLM-voiced and cannot
// drift from what the page says about it.
const PlaygroundSystemPrompt = "You are a small language model that has been fine-tuned on a business's own documents. " +
	"You are running as a free demonstration on finetuning.uk. Answer plainly and briefly, in British English. " +
	"If you are not sure of something, say so."

type playgroundMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type playgroundRequest struct {
	Messages []playgroundMessage `json:"messages"`
}

// ollamaChatRequest is the subset of Ollama's /api/chat request the demo uses.
type ollamaChatRequest struct {
	Model    string                 `json:"model"`
	Messages []playgroundMessage    `json:"messages"`
	Stream   bool                   `json:"stream"`
	Options  map[string]interface{} `json:"options"`
}

// ollamaChatChunk is one NDJSON line of Ollama's streamed /api/chat reply.
type ollamaChatChunk struct {
	Message struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	} `json:"message"`
	Done         bool   `json:"done"`
	DoneReason   string `json:"done_reason"`
	EvalCount    int    `json:"eval_count"`
	EvalDuration int64  `json:"eval_duration"` // nanoseconds
	LoadDuration int64  `json:"load_duration"` // nanoseconds
	Error        string `json:"error"`
}

// validatePlaygroundRequest is the whole input contract, in one place so the
// test can enumerate it. Returns a visitor-facing message on failure.
func validatePlaygroundRequest(req *playgroundRequest) string {
	if len(req.Messages) == 0 {
		return "messages is required"
	}
	if len(req.Messages) > PlaygroundMaxMessages {
		return fmt.Sprintf("at most %d messages per request", PlaygroundMaxMessages)
	}
	for i := range req.Messages {
		m := &req.Messages[i]
		if m.Role != "user" && m.Role != "assistant" {
			return "each message role must be user or assistant"
		}
		m.Content = strings.TrimSpace(m.Content)
		if m.Content == "" {
			return "each message needs some text"
		}
		if utf8.RuneCountInString(m.Content) > PlaygroundMaxMessageRunes {
			return fmt.Sprintf("each message must be at most %d characters", PlaygroundMaxMessageRunes)
		}
	}
	if req.Messages[len(req.Messages)-1].Role != "user" {
		return "the last message must be from the user"
	}
	return ""
}

// PlaygroundChatHandler answers one turn from the demo model as an SSE stream.
func PlaygroundChatHandler(cfg *config.PlaygroundConfig, client *http.Client) gin.HandlerFunc {
	if client == nil {
		client = &http.Client{}
	}
	return func(c *gin.Context) {
		var req playgroundRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			httperr.JSONError(c, http.StatusBadRequest, "invalid request body")
			return
		}
		if msg := validatePlaygroundRequest(&req); msg != "" {
			httperr.JSONError(c, http.StatusBadRequest, msg)
			return
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), playgroundCallTimeout)
		defer cancel()

		upstream := ollamaChatRequest{
			Model:    cfg.Model,
			Messages: append([]playgroundMessage{{Role: "system", Content: PlaygroundSystemPrompt}}, req.Messages...),
			Stream:   true,
			Options: map[string]interface{}{
				"num_predict": cfg.MaxTokens,
				"num_ctx":     cfg.NumCtx,
			},
		}
		body, err := json.Marshal(upstream)
		if err != nil {
			httperr.JSONError(c, http.StatusInternalServerError, "could not build the model request")
			return
		}
		httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, cfg.OllamaURL+"/api/chat", bytes.NewReader(body))
		if err != nil {
			httperr.JSONError(c, http.StatusInternalServerError, "could not build the model request")
			return
		}
		httpReq.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(httpReq)
		if err != nil {
			logPlaygroundFailure("chat", "connect", err)
			httperr.JSONError(c, http.StatusServiceUnavailable, "the demo model is not available right now")
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			snippet, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
			logPlaygroundFailure("chat", "upstream_status", fmt.Errorf("status %d: %s", resp.StatusCode, strings.TrimSpace(string(snippet))))
			httperr.JSONError(c, http.StatusServiceUnavailable, "the demo model is not available right now")
			return
		}

		// From here the reply is a stream: status 200 is committed on the first
		// write, so later failures are reported in-band as `event: error`.
		h := c.Writer.Header()
		h.Set("Content-Type", "text/event-stream")
		h.Set("Cache-Control", "no-cache")
		h.Set("X-Accel-Buffering", "no")
		c.Status(http.StatusOK)

		sc := bufio.NewScanner(resp.Body)
		sc.Buffer(make([]byte, 0, 64*1024), playgroundScanBuf)
		sentAny := false
		for sc.Scan() {
			line := bytes.TrimSpace(sc.Bytes())
			if len(line) == 0 {
				continue
			}
			var chunk ollamaChatChunk
			if err := json.Unmarshal(line, &chunk); err != nil {
				logPlaygroundFailure("chat", "decode_chunk", err)
				writeSSE(c, "error", map[string]string{"message": "the demo model sent something unreadable"})
				return
			}
			if chunk.Error != "" {
				logPlaygroundFailure("chat", "upstream_error", errors.New(chunk.Error))
				writeSSE(c, "error", map[string]string{"message": "the demo model stopped"})
				return
			}
			if chunk.Message.Content != "" {
				writeSSE(c, "token", chunk.Message.Content)
				sentAny = true
			}
			if chunk.Done {
				writeSSE(c, "done", map[string]interface{}{
					"eval_count":       chunk.EvalCount,
					"eval_duration_ms": chunk.EvalDuration / int64(time.Millisecond),
					"load_duration_ms": chunk.LoadDuration / int64(time.Millisecond),
					"done_reason":      chunk.DoneReason,
				})
				return
			}
		}
		// The stream ended without a done line: a dropped connection or a
		// timeout. Say so rather than leaving the browser waiting.
		if err := sc.Err(); err != nil {
			logPlaygroundFailure("chat", "stream", err)
		}
		msg := "the demo model stopped before finishing"
		if !sentAny {
			msg = "the demo model did not answer"
		}
		writeSSE(c, "error", map[string]string{"message": msg})
	}
}

// writeSSE emits one server-sent event and flushes it, so a token reaches the
// browser as it arrives rather than when the buffer fills.
func writeSSE(c *gin.Context, event string, data interface{}) {
	payload, err := json.Marshal(data)
	if err != nil {
		return
	}
	fmt.Fprintf(c.Writer, "event: %s\ndata: %s\n\n", event, payload)
	c.Writer.Flush()
}

func logPlaygroundFailure(endpoint, stage string, err error) {
	log.Printf("playground %s: %s: %v", endpoint, stage, err)
}
