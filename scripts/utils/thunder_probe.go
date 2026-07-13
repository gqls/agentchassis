// thunder_probe.go — standalone, no deps beyond stdlib.
//
// Purpose: discover Thunder's real API contract WITHOUT guessing or
// redeploying the adapter. The CLI hit GET /v1/specs before creating —
// that endpoint almost certainly enumerates valid GPU types. We fetch it,
// print it, and optionally attempt a create with a chosen gpu_type so we
// can iterate the value locally against the real API until we get a 201.
//
// Usage:
//   export THUNDER_COMPUTE_API_KEY=<token>   # already in your shell env? then skip
//   go run thunder_probe.go                  # just dump /specs and /instances/list
//   go run thunder_probe.go -create -gpu a100xl   # also try a create (then DELETE it)
//
// The -create path deletes the instance immediately after, so cost is ~$0.
// If create returns non-201, nothing is provisioned and we see the exact error.

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

const baseURL = "https://api.thundercompute.com:8443/v1"

func main() {
	var (
		doCreate = flag.Bool("create", false, "attempt a create (auto-deletes after)")
		gpu      = flag.String("gpu", "a100", "gpu_type value to try on create")
		vcpus    = flag.Int("vcpus", 4, "cpu_cores")
		disk     = flag.Int("disk", 100, "disk_size_gb")
		template = flag.String("template", "base", "template")
		mode     = flag.String("mode", "prototyping", "mode")
		numGPUs  = flag.Int("num-gpus", 1, "num_gpus")
	)
	flag.Parse()

	token := os.Getenv("THUNDER_COMPUTE_API_KEY")
	if token == "" {
		fmt.Fprintln(os.Stderr, "ERROR: THUNDER_COMPUTE_API_KEY not set in env")
		os.Exit(1)
	}

	ctx := context.Background()

	// 1. GET /specs — the endpoint the CLI hit before create. Likely the
	//    authoritative list of valid GPU types, vcpu options, disk ranges.
	fmt.Println("==================== GET /specs ====================")
	dump(ctx, token, http.MethodGet, "/specs", nil)

	// 2. GET /instances/list — confirms the list response shape we already
	//    saw via tnr, and proves auth works on this token.
	fmt.Println("\n==================== GET /instances/list ====================")
	dump(ctx, token, http.MethodGet, "/instances/list", nil)

	// 3. Optional create probe.
	if *doCreate {
		fmt.Printf("\n==================== POST /instances/create (gpu_type=%q) ====================\n", *gpu)
		body := map[string]interface{}{
			"gpu_type":     *gpu,
			"num_gpus":     *numGPUs,
			"cpu_cores":    *vcpus,
			"disk_size_gb": *disk,
			"mode":         *mode,
			"template":     *template,
		}
		respBytes := dump(ctx, token, http.MethodPost, "/instances/create", body)

		// If create succeeded, parse out an id and delete immediately.
		var created map[string]interface{}
		if err := json.Unmarshal(respBytes, &created); err == nil {
			id := firstNonEmpty(created, "identifier", "id", "uuid")
			if id != "" {
				fmt.Printf("\n>>> create returned id=%q — deleting immediately to avoid cost\n", id)
				time.Sleep(1 * time.Second)
				dump(ctx, token, http.MethodPost, "/instances/"+id+"/delete", nil)
			} else {
				fmt.Println("\n>>> WARNING: create response had no recognisable id field — CHECK MANUALLY with `tnr status` and delete if a real instance exists")
			}
		}
	}
}

// dump performs the request, prints status + pretty body, returns raw body.
func dump(ctx context.Context, token, method, path string, body interface{}) []byte {
	var reader io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		reader = bytes.NewReader(b)
		fmt.Printf("REQUEST BODY: %s\n", string(b))
	}
	req, err := http.NewRequestWithContext(ctx, method, baseURL+path, reader)
	if err != nil {
		fmt.Printf("build request error: %v\n", err)
		return nil
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("http error: %v\n", err)
		return nil
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)
	fmt.Printf("STATUS: %d\n", resp.StatusCode)

	// Pretty-print if it's JSON, else raw.
	var pretty bytes.Buffer
	if json.Indent(&pretty, raw, "", "  ") == nil {
		fmt.Printf("RESPONSE:\n%s\n", pretty.String())
	} else {
		fmt.Printf("RESPONSE (raw): %s\n", string(raw))
	}
	return raw
}

func firstNonEmpty(m map[string]interface{}, keys ...string) string {
	for _, k := range keys {
		if v, ok := m[k]; ok {
			switch t := v.(type) {
			case string:
				if t != "" {
					return t
				}
			case float64:
				return fmt.Sprintf("%d", int(t))
			}
		}
	}
	return ""
}
