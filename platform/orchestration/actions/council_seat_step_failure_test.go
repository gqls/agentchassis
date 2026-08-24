package actions

import "testing"

// bugs_open/243. A council seat whose STEP FAILED and a seat the relevance filter
// SKIPPED both leave the seat's result field absent — and they must not be counted
// the same way. An `abstained` seat is a considered non-applicability and does not
// gate; an `unreadable` seat is an opinion we were owed and lost, and it downgrades
// an approval to revise. Conflating them is the silent-approval hazard that makes
// the config half of the 243 fix unsafe on its own, so both arms are pinned here.
//
// These test reviewStepFailed directly rather than the whole action, because that
// helper IS the discrimination — everything downstream of it is the pre-existing
// unreadable machinery, already covered by TestUnreadableSeatCannotApprove.
func TestReviewStepFailedDiscriminatesLostSeatFromSkippedSeat(t *testing.T) {
	stepErrors := map[string]interface{}{
		"review_editquality": map[string]interface{}{
			"message": "step review_editquality failed: AI endpoint unavailable: " +
				"API request failed with status 400: You have reached your specified API usage limits",
			"at": "2026-08-24T13:00:00Z",
		},
	}
	collected := map[string]interface{}{"__step_errors": stepErrors}

	t.Run("seat whose step failed is a LOST opinion", func(t *testing.T) {
		failed, why := reviewStepFailed(collected, "review_editquality.result")
		if !failed {
			t.Fatal("a seat whose step routed to error_step must be reported as failed — " +
				"counting it as an abstention lets a lost opinion read as a considered non-objection")
		}
		if why == "" {
			t.Error("the recorded message must come back, so the loss is attributable in the log rather than merely counted")
		}
	})

	t.Run("seat skipped by the relevance filter is NOT a lost opinion", func(t *testing.T) {
		// This is the arm that makes the test discriminate. review_architecture is
		// absent from __step_errors because nothing went wrong — the panel simply
		// did not select it. If this returned true, every gated-off seat in every
		// round would block approvals, which is the opposite failure.
		if failed, _ := reviewStepFailed(collected, "review_architecture.result"); failed {
			t.Fatal("a seat absent from __step_errors was skipped, not lost — it must stay an abstention")
		}
	})

	t.Run("fails CLOSED when the coordinator has not written the key", func(t *testing.T) {
		// An older coordinator, or a state written before the 243 change shipped.
		// The safe direction is today's behaviour (abstention), NOT unreadable:
		// this is what makes the Go half safe to ship ahead of the config half.
		if failed, _ := reviewStepFailed(map[string]interface{}{}, "review_editquality.result"); failed {
			t.Error("with no __step_errors the helper must report false, so behaviour is unchanged until a writer exists")
		}
		if failed, _ := reviewStepFailed(map[string]interface{}{"__step_errors": map[string]interface{}{}}, "review_editquality.result"); failed {
			t.Error("an empty __step_errors map must report false")
		}
		// A malformed value must not panic and must not manufacture a failure.
		if failed, _ := reviewStepFailed(map[string]interface{}{"__step_errors": "not-a-map"}, "review_editquality.result"); failed {
			t.Error("a non-map __step_errors must report false rather than being coerced into a seat loss")
		}
	})

	t.Run("step name is the field's first segment, not a trimmed .result suffix", func(t *testing.T) {
		// Guards the seat→step mapping against a seat configured with a different
		// output field: splitting on the first "." keeps working, trimming
		// ".result" would silently stop matching and the seat would go back to
		// counting as an abstention.
		if failed, _ := reviewStepFailed(collected, "review_editquality.something_else"); !failed {
			t.Error("the step name must be derived from the field's first segment")
		}
		if failed, _ := reviewStepFailed(collected, "review_editquality"); !failed {
			t.Error("a bare step name with no dotted suffix must still resolve")
		}
	})
}
