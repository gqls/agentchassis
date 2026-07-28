// FILE: platform/orchestration/actions/assemble_upload_manifest_action.go
//
// assemble_upload_manifest: a PURE (no DB, no network) launcher helper that
// builds /workspace/upload_manifest.json's CONTENT from the presigned URLs and
// the keys computed by compute_checkpoint_keys. It is the last new launcher step
// before the launch ssh_exec; the manifest is written onto the VM by that step
// (echo '<manifest_b64>' | base64 -d > /workspace/upload_manifest.json), so this
// action also returns a base64 form to dodge shell-quoting of the URL-laden JSON.
//
// 02_train's --upload-manifest consumes exactly this shape:
//   { "run_id": "<training_run_id>",
//     "checkpoints": [ {"index":0,"key":"...","url":"<PUT>"}, ... ],   // by save index
//     "final":  {"key":"...","url":"<PUT>"},
//     "resume": {"key":"...","url":"<GET>","index":N} }                // only if resume_* given
//
// checkpoint_keys[i] is paired with checkpoint_urls[i] (same order); index = i.
// This pairing is the contract with compute_checkpoint_keys (keys) and the
// presign step (urls): both MUST be the same ordered set. A length mismatch is
// an error — a silently short manifest would drop checkpoints.
//
// Inputs (the launcher workflow declares config dot-paths to the producing steps,
// exactly like ssh_exec_launch declares scripts_url/dataset_url; local steps do
// not get input_mapping resolution, so the values are reached via config paths):
//   training_run_id (required) — manifest run_id.
//   checkpoint_keys (optional list) — from compute_checkpoint_keys.checkpoint_keys
//   checkpoint_urls (optional list) — the presign step's ordered presigned_urls
//   final_key       (required) — compute_checkpoint_keys.final_key
//   final_url       (required) — the final presign's presigned_url
//   resume_key, resume_url (optional), resume_index (optional int) — resume launch only.
//
// Output:
//   { "manifest_json": "<json string>",
//     "manifest_b64":  "<base64 of manifest_json>",
//     "checkpoint_count": <len> }

package actions

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

// assembleUploadManifestSpec is the input contract (registered for validation and
// passed to ExtractActionInputs).
var assembleUploadManifestSpec = datahelpers.ActionInputSpec{
	CheckConfig: true,
	Required:    []string{"training_run_id", "final_key", "final_url"},
	Optional:    []string{"checkpoint_keys", "checkpoint_urls", "resume_key", "resume_url", "resume_index"},
}

func init() {
	datahelpers.RegisterActionInputSpec("assemble_upload_manifest", assembleUploadManifestSpec)
}

// AssembleUploadManifestAction builds the upload_manifest.json content (+ base64).
func AssembleUploadManifestAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(zap.String("action", "assemble_upload_manifest"))

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	inputs, err := datahelpers.ExtractActionInputs(
		params.CollectedData, params.StepConfig.Config, assembleUploadManifestSpec, logger,
	)
	if err != nil {
		return nil, fmt.Errorf("assemble_upload_manifest: %w", err)
	}

	runID := inputs.Get("training_run_id")
	finalKey := inputs.Get("final_key")
	finalURL := inputs.Get("final_url")
	if runID == "" || finalKey == "" || finalURL == "" {
		return nil, fmt.Errorf("assemble_upload_manifest: training_run_id, final_key, and final_url are all required")
	}

	// Reuse the existing list coercion. It drops any non-string element silently,
	// but the length check below turns any such drop into a loud error rather than
	// a misaligned key/url pairing.
	keys := datahelpers.ExtractStringListHelper(inputs.GetRaw("checkpoint_keys"))
	urls := datahelpers.ExtractStringListHelper(inputs.GetRaw("checkpoint_urls"))
	if len(keys) != len(urls) {
		return nil, fmt.Errorf(
			"assemble_upload_manifest: checkpoint_keys (%d) and checkpoint_urls (%d) length mismatch — both must be the same ordered set",
			len(keys), len(urls))
	}

	checkpoints := make([]map[string]interface{}, 0, len(keys))
	for i := range keys {
		checkpoints = append(checkpoints, map[string]interface{}{
			"index": i,
			"key":   keys[i],
			"url":   urls[i],
		})
	}

	manifest := map[string]interface{}{
		"run_id":      runID,
		"checkpoints": checkpoints,
		"final": map[string]interface{}{
			"key": finalKey,
			"url": finalURL,
		},
	}

	// Resume is present only on a resume launch.
	resumeURL := inputs.Get("resume_url")
	if resumeURL != "" {
		resume := map[string]interface{}{"url": resumeURL}
		if rk := inputs.Get("resume_key"); rk != "" {
			resume["key"] = rk
		}
		if inputs.Has("resume_index") {
			resume["index"] = inputs.GetInt("resume_index", 0)
		}
		manifest["resume"] = resume
	}

	jsonBytes, err := json.Marshal(manifest)
	if err != nil {
		return nil, fmt.Errorf("assemble_upload_manifest: marshal manifest: %w", err)
	}
	b64 := base64.StdEncoding.EncodeToString(jsonBytes)

	logger.Info("assembled upload manifest",
		zap.String("training_run_id", runID),
		zap.Int("checkpoint_count", len(checkpoints)),
		zap.Bool("resume", resumeURL != ""),
		zap.Int("manifest_bytes", len(jsonBytes)),
	)

	return map[string]interface{}{
		"manifest_json":    string(jsonBytes),
		"manifest_b64":     b64,
		"checkpoint_count": len(checkpoints),
	}, nil
}
