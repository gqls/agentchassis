// FILE: platform/orchestration/actions/flatten_presign_results_action.go
//
// flatten_presign_results: a PURE connector between the launcher's checkpoint
// presign LOOP and assemble_upload_manifest.
//
// The chassis loop_complete step emits {iterations, results, count}, where each
// results[i] is a per-iteration map shaped like:
//   {"iteration": i, "<substep_output_field>": <adapter reply>, "original_item": <loop item>, "name": ...}
// assemble_upload_manifest instead wants two FLAT, ordered, same-length lists
// (checkpoint_urls[i] paired with checkpoint_keys[i]). This action does only that
// reshape, so assemble stays loop-agnostic — and keeps working unchanged if the
// presign loop is ever replaced by a batch presign that already returns flat lists.
//
// For each results[i] it digs (via ExtractNestedFieldString, which absorbs the
// adapter reply's one-level ".response" wrapper):
//   url_field (default "ckpt_presign.presigned_url") -> checkpoint_urls[i]
//   key_field (default "ckpt_presign.key")           -> checkpoint_keys[i]
// An element missing either value is an error — a silent skip would drop a
// checkpoint or misalign the key/url pairing.
//
// Inputs (the launcher declares a config dot-path, like ssh_exec_launch's
// scripts_url/dataset_url; local steps don't get input_mapping resolution):
//   presign_results (required) — the loop step's results array, e.g. "ckpt_presigns.results".
// Config:
//   url_field (string, default "ckpt_presign.presigned_url") — path within each element.
//   key_field (string, default "ckpt_presign.key")           — path within each element.
//
// Output (consumed by assemble_upload_manifest):
//   { "checkpoint_urls": [...], "checkpoint_keys": [...], "count": N }

package actions

import (
	"context"
	"fmt"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

const (
	defaultPresignURLField = "ckpt_presign.presigned_url"
	defaultPresignKeyField = "ckpt_presign.key"
)

var flattenPresignResultsSpec = datahelpers.ActionInputSpec{
	Required: []string{"presign_results"},
}

func init() {
	datahelpers.RegisterActionInputSpec("flatten_presign_results", flattenPresignResultsSpec)
}

// FlattenPresignResultsAction reshapes loop_complete results into flat url/key lists.
func FlattenPresignResultsAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(zap.String("action", "flatten_presign_results"))

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	inputs, err := datahelpers.ExtractActionInputs(
		params.CollectedData, params.StepConfig.Config, flattenPresignResultsSpec, logger,
	)
	if err != nil {
		return nil, fmt.Errorf("flatten_presign_results: %w", err)
	}

	rawResults := inputs.GetRaw("presign_results")
	results, ok := rawResults.([]interface{})
	if !ok {
		return nil, fmt.Errorf("flatten_presign_results: presign_results is %T, want a list (the loop step's .results)", rawResults)
	}

	urlField := datahelpers.GetStringField(params.StepConfig.Config, "url_field", defaultPresignURLField)
	keyField := datahelpers.GetStringField(params.StepConfig.Config, "key_field", defaultPresignKeyField)

	urls := make([]string, 0, len(results))
	keys := make([]string, 0, len(results))
	for i, elem := range results {
		em, ok := elem.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("flatten_presign_results: result %d is %T, want an object", i, elem)
		}
		url := datahelpers.ExtractNestedFieldString(em, urlField)
		key := datahelpers.ExtractNestedFieldString(em, keyField)
		if url == "" || key == "" {
			return nil, fmt.Errorf(
				"flatten_presign_results: result %d missing url or key (url_field=%q key_field=%q) — got url=%q key=%q",
				i, urlField, keyField, url, key)
		}
		urls = append(urls, url)
		keys = append(keys, key)
	}

	logger.Info("flattened presign results", zap.Int("count", len(urls)))

	return map[string]interface{}{
		"checkpoint_urls": urls,
		"checkpoint_keys": keys,
		"count":           len(urls),
	}, nil
}
