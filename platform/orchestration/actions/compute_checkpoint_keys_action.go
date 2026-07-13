// FILE: platform/orchestration/actions/compute_checkpoint_keys_action.go
//
// compute_checkpoint_keys: a PURE (no DB, no network) launcher helper that
// produces the list of B2 object keys the training VM will PUT its periodic
// checkpoints to, plus the single final-adapter key. It is the first new step in
// the Phase 5 launcher's upload path: its keys are presigned (each key -> a
// write-only PUT URL) and then folded, with those URLs, into
// /workspace/upload_manifest.json by assemble_upload_manifest.
//
// Why keys are minted cluster-side: the VM holds no B2 credentials, so it cannot
// choose where to write. The launcher (trusted) decides the keys; the adapter
// (credential boundary) signs them; 02_train's CheckpointUploader PUTs its Nth
// save to checkpoints[N].
//
// Count (K):
//   - If BOTH max_steps and save_steps are provided (>0):
//       K = ceil(max_steps / save_steps) + buffer.
//     This is the precise count once the launcher can compute max_steps
//     (epochs * ceil(n_examples / (batch*grad_accum))).
//   - Otherwise: K = checkpoint_count (a config fallback, default 64). max_steps
//     is NOT currently known cluster-side (the launcher has dataset_uri but not
//     the example count), so the fallback is the working default for now.
//     Over-provisioning is cheap — presigning is local HMAC and unused URLs just
//     expire; 02_train only uses as many entries as it performs saves.
//   K is clamped to [1, 512] as a sanity bound.
//
// Keys (overridable via config; defaults match the adapter's existing artefact
// convention so the final adapter lands where prepare_artefact_url/readers expect):
//   checkpoint: finetuning/checkpoints/<training_run_id>/ckpt-<index>.tar.gz  (index 0..K-1)
//   final:      finetuning/artefacts/<training_run_id>/adapter.tar.gz
//
// Inputs:
//   training_run_id  (required) — from input_data (threaded by model-trainer).
//   max_steps        (optional) — config dot-path or input_data; 0 if unknown.
//   save_steps       (optional) — config dot-path or input_data; 0 if unknown.
// Config literals (read directly; ExtractActionInputs does not resolve bare literals):
//   checkpoint_count        (int, default 64)  — fallback K when max_steps unknown.
//   buffer                  (int, default 4)   — extra keys beyond the computed count.
//   checkpoint_key_template (string, optional) — must contain %s (run_id) then %d (index).
//   final_key_template      (string, optional) — must contain %s (run_id).
//
// Output (consumed by the presign step + assemble_upload_manifest):
//   { "checkpoint_keys": ["finetuning/checkpoints/<run>/ckpt-0.tar.gz", ...],
//     "final_key": "finetuning/artefacts/<run>/adapter.tar.gz",
//     "checkpoint_count": K }

package actions

import (
	"context"
	"fmt"
	"strings"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

const (
	defaultCheckpointCount     = 64
	defaultCheckpointKeyBuffer = 4
	minCheckpointCount         = 1
	maxCheckpointCount         = 512
	defaultCheckpointKeyTmpl   = "finetuning/checkpoints/%s/ckpt-%d.tar.gz"
	defaultFinalAdapterKeyTmpl = "finetuning/artefacts/%s/adapter.tar.gz"
)

// computeCheckpointKeysSpec is the input contract (registered for validation and
// passed to ExtractActionInputs).
var computeCheckpointKeysSpec = datahelpers.ActionInputSpec{
	Required: []string{"training_run_id"},
	Optional: []string{"max_steps", "save_steps"},
}

func init() {
	datahelpers.RegisterActionInputSpec("compute_checkpoint_keys", computeCheckpointKeysSpec)
}

// ComputeCheckpointKeysAction builds the checkpoint key list + final key.
func ComputeCheckpointKeysAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(zap.String("action", "compute_checkpoint_keys"))

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	inputs, err := datahelpers.ExtractActionInputs(
		params.CollectedData, params.StepConfig.Config, computeCheckpointKeysSpec, logger,
	)
	if err != nil {
		return nil, fmt.Errorf("compute_checkpoint_keys: %w", err)
	}
	runID := inputs.Get("training_run_id")
	if runID == "" {
		return nil, fmt.Errorf("compute_checkpoint_keys: training_run_id is required")
	}

	// Optional precise-count ingredients (0 if not provided).
	maxSteps := inputs.GetInt("max_steps", 0)
	saveSteps := inputs.GetInt("save_steps", 0)

	// Config literals (read directly — ExtractActionInputs does not resolve bare literals).
	fallbackCount := datahelpers.GetIntField(params.StepConfig.Config, "checkpoint_count", defaultCheckpointCount)
	buffer := datahelpers.GetIntField(params.StepConfig.Config, "buffer", defaultCheckpointKeyBuffer)
	ckptTmpl := datahelpers.GetStringField(params.StepConfig.Config, "checkpoint_key_template", defaultCheckpointKeyTmpl)
	finalTmpl := datahelpers.GetStringField(params.StepConfig.Config, "final_key_template", defaultFinalAdapterKeyTmpl)

	if !strings.Contains(ckptTmpl, "%s") || !strings.Contains(ckptTmpl, "%d") {
		return nil, fmt.Errorf("compute_checkpoint_keys: checkpoint_key_template %q must contain %%s (run_id) and %%d (index)", ckptTmpl)
	}
	if !strings.Contains(finalTmpl, "%s") {
		return nil, fmt.Errorf("compute_checkpoint_keys: final_key_template %q must contain %%s (run_id)", finalTmpl)
	}

	// Decide K.
	var k int
	if maxSteps > 0 && saveSteps > 0 {
		k = (maxSteps+saveSteps-1)/saveSteps + buffer // ceil(max/save) + buffer
		logger.Info("computed checkpoint count from max_steps/save_steps",
			zap.Int("max_steps", maxSteps), zap.Int("save_steps", saveSteps),
			zap.Int("buffer", buffer), zap.Int("k", k))
	} else {
		k = fallbackCount
		logger.Info("max_steps/save_steps not both provided — using fallback checkpoint_count",
			zap.Int("checkpoint_count", k))
	}
	if k < minCheckpointCount {
		k = minCheckpointCount
	}
	if k > maxCheckpointCount {
		logger.Warn("clamping checkpoint count to max",
			zap.Int("requested", k), zap.Int("max", maxCheckpointCount))
		k = maxCheckpointCount
	}

	keys := make([]string, 0, k)
	for i := 0; i < k; i++ {
		keys = append(keys, fmt.Sprintf(ckptTmpl, runID, i))
	}
	finalKey := fmt.Sprintf(finalTmpl, runID)

	logger.Info("built checkpoint key list",
		zap.String("training_run_id", runID),
		zap.Int("checkpoint_count", k),
		zap.String("final_key", finalKey),
	)

	return map[string]interface{}{
		"checkpoint_keys":  keys,
		"final_key":        finalKey,
		"checkpoint_count": k,
	}, nil
}
