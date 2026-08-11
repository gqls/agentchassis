// FILE: platform/orchestration/actions/deployed_image_read_audit.go
//
// The three readers of a deploy_image_asset result used to be silent when the
// result carried no URL: an `ok`-guarded map access with no else branch, at
// v3_site_actions.go (hero and logo) and assemble_from_library.go
// (extractLogoURLFromParams). A page shipping with no hero and no logo looked
// identical to a page that never wanted one, for five weeks — bugs_open/236,
// commission item 2 (owner ruling 2026-08-10).
//
// OBSERVABILITY ONLY. This file does not widen what the readers accept. Teaching
// them to also read response.data.file_path is bugs_open/236 §4 candidate 2, it
// encodes the merged await shape at three call sites (the
// unified_extractor.go:200 pattern the RFC_012 census already flagged), and it is
// item 1's design decision, which is reserved to the owner.
//
// WHY A DURABLE ROW AND NOT ONLY A Warn. The commission asked for a Warn. A Warn
// alone cannot be the evidence this is for, because both places it would live are
// short-lived:
//
//   - the pod log: these readers run in agent-chassis, whose own startup
//     "build provenance" line was measured absent from --tail=3000 hours after a
//     roll (2026-08-11);
//   - the orchestration row: hero_deployed/logo_deployed exist only while the
//     state is AWAITING_RESPONSES — because they ARE the awaited responses — and
//     database-cleanup prunes that status after FOUR hours (live scheduled_tasks
//     row, hourly; the repo seed says 24h and disagrees). The three observations
//     of this bug (0 of 1,667 → 2 each → 0) are one 4-hour window opening and
//     closing, not the bug coming and going.
//
// agent_error_log is documented as the one sink that outlives an awaited step
// (agenterrors.go:20-24, log_action_error.go:14-18), so the finding goes there.
//
// THE DEMAND GATE IS THE LOAD-BEARING PART. BuildRenderContextAction runs for
// every page build and most pages never deploy a hero or a logo, so an else on
// the outer guard would file a row for every page of every site. The container
// key's PRESENCE is the demand signal: collected_data carries hero_deployed only
// because a deploy_hero_image step ran. So absence is silent, and only
// present-but-unusable is reported. The no-op case is tested, not assumed.
//
// ⚠ STATED BLIND SPOT: this is silent if some mechanism removes the container key
// ENTIRELY, because then nothing local says a deploy ran. The observed shape
// (bugs_open/236 §1) has the key present carrying
// response/response_status/response_received_at, and coordinator.go:2719-2748
// preserves-then-adds, so key-present is the shape to catch today. Whether the
// key-absent variant ever occurs is [UNMEASURED].
package actions

import (
	"context"
	"fmt"
	"sort"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

// deployedImageMissCode is the queryable name of the finding. One code for all
// three sites: container_key in the context separates hero from logo, and action
// separates the readers, so a single code keeps the standing query one line.
const deployedImageMissCode = "DEPLOYED_IMAGE_RESULT_MISSING_URL"

// mapKeysSorted returns v's keys when v is a string-keyed map, sorted, and nil
// otherwise.
//
// KEYS, NEVER VALUES — deliberately. The keys are the whole diagnostic: a
// container holding response/response_status/response_received_at and no
// image_url is recognisably the bugs_open/236 shape at a glance. The values
// carry base64 image payloads and full response bodies, which is why the
// commission asked for "its keys — not the whole value, which can be large".
// Sorted so two rows of the same shape compare equal by eye and by string.
func mapKeysSorted(v interface{}) []string {
	m, ok := v.(map[string]interface{})
	if !ok {
		return nil
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// deployedImageAuditSiteID resolves the site this finding belongs to, for the
// row's site_id column only. Two rungs because the two calling actions carry it
// differently: assemble_from_library's params have site_record, while
// BuildRenderContextAction is configured with a site_id_field that conventionally
// points into input_data. Empty when neither answers — agenterrors.Write
// NULLIFs it, and the run join (orchestration_id) is inherited regardless, so an
// unresolved site_id costs precision, never the row.
func deployedImageAuditSiteID(params ActionParams) string {
	if id := extractSiteIDFromParams(params); id != uuid.Nil {
		return id.String()
	}
	return datahelpers.ExtractNestedFieldString(params.CollectedData, "input_data.site_id")
}

// deployedImageURL reads wantedKey out of the deploy_image_asset result that a
// step left under containerKey, and — when the container is PRESENT but carries
// no usable URL — records that durably before returning "".
//
// action names the reading site and becomes the row's action column.
// fallbackKey is the plain collected_data key the caller's own fallback path
// reads (hero_url / logo_url); it is used ONLY to annotate the finding, never to
// satisfy the read. That annotation is the discriminator bugs_open/236 §5 needs:
// deploy_image_asset_action.go:404-415 has written that sibling key since
// 2026-02-23, so "the container lost the key" and "the URL is lost everywhere"
// are different failures, and the row now says which one happened.
//
// The claim this makes is strictly about the CONTAINER. It does not assert the
// page shipped without an image — the caller's fallback may still supply one,
// which is exactly what fallback_sibling_present records.
func deployedImageURL(ctx context.Context, params ActionParams, containerKey, wantedKey, fallbackKey, action string) string {
	raw, present := params.CollectedData[containerKey]
	if !present {
		// NO DEMAND. Nothing deployed an image for this page, so there is
		// nothing to report and no row to file. This branch is the reason the
		// change is quiet on the fleet, and it is covered by a test.
		return ""
	}

	container, isMap := raw.(map[string]interface{})
	if isMap {
		if url, ok := container[wantedKey].(string); ok && url != "" {
			return url
		}
	}

	keysHeld := mapKeysSorted(raw)
	fallbackPresent := datahelpers.ExtractNestedFieldString(params.CollectedData, fallbackKey) != ""

	// The Warn is the immediate signal and the post-roll pod-grep handle (a log
	// string is a real literal, unlike a doc comment). The row below is what
	// still exists tomorrow.
	params.Logger.Warn("deploy_image_asset result carries no usable URL — image not set from it",
		zap.String("container_key", containerKey),
		zap.String("wanted_key", wantedKey),
		zap.Strings("keys_present", keysHeld),
		zap.String("container_type", fmt.Sprintf("%T", raw)),
		zap.Bool("fallback_sibling_present", fallbackPresent),
		zap.String("bug", "bugs_open/236"))

	message := fmt.Sprintf(
		"%s: collected_data[%q] is present but carries no usable %q (it held: %v) — this image was not set from the deploy result",
		action, containerKey, wantedKey, keysHeld)

	landed := LogActionError(ctx, params, deployedImageAuditSiteID(params), extractDomainFromParams(params),
		action, deployedImageMissCode, "warning", message,
		map[string]interface{}{
			"container_key":  containerKey,
			"wanted_key":     wantedKey,
			"keys_present":   keysHeld,
			"container_type": fmt.Sprintf("%T", raw),
			// true => the deploy's own sibling key survived and the caller's
			// fallback will still find a URL, so the page is probably fine and
			// the loss is confined to the container. false => the URL is gone
			// from both places and the page renders without this image.
			"fallback_sibling_present": fallbackPresent,
			"fallback_key":             fallbackKey,
			"remedy":                   "bugs_open/236: the deploy_image_asset result reaches this reader without the image_url/output_path/size_bytes the action assigned to it. Root cause OPEN (§5 — the obvious merge-overwrite theory is REFUTED there). This row is the evidence: keys_present is the container's shape at the moment of the miss, and fallback_sibling_present says whether the URL survived anywhere. Do NOT fix by teaching this reader the merged shape — that is §4 candidate 2, refused as item 1's owner decision",
		}, params.Logger)

	if !landed {
		// A lost row must never read as a recorded one. This is the only copy of
		// the finding once the orchestration state is pruned (4 hours), so losing
		// it is worth an Error rather than a silent best-effort shrug.
		params.Logger.Error("failed to record the deploy_image_asset URL miss durably — this finding now exists only in this log line, which does not outlive the pod's log buffer",
			zap.String("container_key", containerKey),
			zap.String("error_code", deployedImageMissCode))
	}

	return ""
}
