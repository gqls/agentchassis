// FILE: platform/orchestration/actions/send_followup_email_action.go
//
// The follow-up email: the second and last thing we send a customer after
// delivery. It repeats where the hosting instructions live and offers the
// confirm link again, and it is SUPPRESSED by the confirm button — which is the
// whole point of it existing (bugs_open/477).
//
// WHY THIS IS NOT A CONFIG CHANGE TO delivery-email-sender. That action goes
// through delivery.Claim, which stamps the handover and returns
// ErrAlreadyDelivered for anything already stamped. Every site a follow-up
// targets is by definition already handed over, so the existing action refuses,
// by design, exactly the population this one exists for. The two are different
// claims on different columns and they must stay that way: making the delivery
// send skippable by config would remove the guard that stops a retry
// double-emailing a customer.
//
// WHAT IS DELIBERATELY NOT HERE, following send_delivery_email_action.go:
//   - No copy. subject/body_template live in the STEP CONFIG (DB,
//     owner-editable without a roll). This file only refuses to send one that
//     names a link this dispatch cannot produce.
//   - No interval. followup_after_days is REQUIRED config with no default,
//     because the interval is the owner's decision ("a week or so" is not a
//     number) and a compiled-in guess would quietly become the answer.
//   - No hosting steps. The email carries {{instructions_link}} and never the
//     instructions themselves — the rule from bugs_open/475: anything that can
//     go out of date lives on the page, never in a copy the customer already
//     holds. An email is exactly such a copy.
//   - No SMTP secrets. mailer.FromEnv reads DELIVERY_SMTP_* from the pod env.
//
// ⚠ THE RECIPIENT COMES FROM build_queue.direction->>'customer_email', NEVER
// FROM sites.email (corrected 2026-08-31 in 651's header, bugs_open/420: since
// the contract split, sites.email is the PUBLISHED contact only and is
// legitimately NULL on a post-420 site). This action does not read either — the
// address arrives as input_data, and it is the scheduled task's pre_query that
// must take it from the right place. Migration 775 does.
//
// Workflow config example (the delivery-followup-sender seed, sql_for_agents/775):
//
//   "send_followup": {
//     "action": "send_followup_email",
//     "config": {
//       "site_id":             "input_data.site_id",
//       "customer_email":      "input_data.customer_email",
//       "live_site_url":       "input_data.live_site_url",
//       "links_host":          "links.webdesign.uk",
//       "instructions_url":    "https://webdesign.uk/your-site",
//       "followup_after_days": 7,
//       "subject":             "Your website files, and where the instructions live",
//       "body_template":       "... {{live_site}} ... {{instructions_link}} ... {{confirm_link}}"
//     },
//     "output_field": "followup_email",
//     "next_step": "complete"
//   }
//
// Registration (registry.go): "send_followup_email" -> SendFollowupEmailAction.

package actions

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/gqls/agentchassis/platform/delivery"
	"github.com/gqls/agentchassis/platform/mailer"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
)

var SendFollowupEmailInputSpec = datahelpers.ActionInputSpec{
	Required: []string{"site_id", "customer_email", "live_site_url"},
	Optional: []string{},
	Defaults: map[string]interface{}{},
}

// SendFollowupEmailAction claims the follow-up for one site and sends it.
//
// ORDER, and why — the same order as the delivery send, for the same reason.
// The SENDER is constructed FIRST, before the claim, because the claim stamps
// followup_sent_at and that stamp deliberately stands even if the send then
// fails. A missing SMTP secret discovered after the stamp would consume this
// site's one follow-up and send nothing. Constructed-then-unused is free.
//
// The CLAIM is what makes this at-most-once, not the send and not the
// scheduler's selection: delivery.ClaimFollowup claims the row on
// `followup_sent_at IS NULL` and re-checks `transfer_confirmed_at IS NULL` in
// the same statement, so a customer who presses the confirm button between the
// pre_query and this dispatch is not emailed. That re-check is the first reader
// sites.transfer_confirmed_at has ever had.
func SendFollowupEmailAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(zap.String("action", "send_followup_email"))

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}
	if params.DB == nil {
		return nil, fmt.Errorf("database connection required")
	}

	inputs, err := datahelpers.ExtractActionInputs(
		params.CollectedData, params.StepConfig.Config, SendFollowupEmailInputSpec, logger)
	if err != nil {
		return nil, fmt.Errorf("input extraction failed: %w", err)
	}

	siteID, err := uuid.Parse(inputs.Get("site_id"))
	if err != nil {
		return nil, fmt.Errorf("invalid site_id %q: %w", inputs.Get("site_id"), err)
	}
	customerEmail := strings.TrimSpace(inputs.Get("customer_email"))
	if customerEmail == "" {
		return nil, fmt.Errorf("customer_email resolved empty")
	}

	config := params.StepConfig.Config
	subject, _ := config["subject"].(string)
	bodyTemplate, _ := config["body_template"].(string)
	linksHost, _ := config["links_host"].(string)
	if subject == "" || bodyTemplate == "" || linksHost == "" {
		return nil, fmt.Errorf("subject, body_template and links_host config are all required (the copy is CONFIG, deliberately: owner-editable without a roll)")
	}

	// REQUIRED, with no default. The owner said "a week or so"; that is not a
	// number, and a number invented here would become the answer without anyone
	// deciding it. Refusing is the honest behaviour until the config says.
	afterDays, err := followupAfterDays(config)
	if err != nil {
		return nil, err
	}

	sender, err := newDeliverySender()
	if err != nil {
		// Before the claim, on purpose — see the function comment.
		return nil, fmt.Errorf("follow-up email sender unavailable (check DELIVERY_SMTP_* env + secret): %w", err)
	}

	// TEMPLATE-VS-LINKS, ALSO BEFORE THE CLAIM, and now DERIVED rather than
	// mirrored. This declaration IS the guard: delivery.Fill.Check refuses
	// malformed, unknown, uncovered and unavailable tokens in one pass, before
	// ClaimFollowup. Adding a token to delivery.Vocabulary and not declaring it
	// here turns this sender's coverage test red rather than putting a blank in
	// a customer's letter.
	//
	// ⚠ THE CHECK MUST PRECEDE **ClaimFollowup**, NOT delivery.Claim. The two
	// senders have DIFFERENT irreversible statements, and a Check that ran after
	// this one would pass every assertion in it while having already burnt
	// followup_sent_at — the customer's single follow-up, spent on a refusal.
	// The order is asserted in this package's tests, not in the vocabulary.
	//
	// NAME NOTE: the placeholder is {{instructions_link}} while its config key is
	// instructions_url. That is the estate's convention, not a slip —
	// {{domain_rent_link}} comes from domain_rent_url, and two more like it.
	instructionsURL := stringOr(config, "instructions_url")
	fill := delivery.Fill{
		delivery.TokenLiveSite: {Value: inputs.Get("live_site_url"), Source: "live_site_url input"},

		// PRODUCED BY THE CLAIM, so legitimately empty when Check runs: this
		// sender mints its own confirm token AFTER ClaimFollowup (the delivery
		// sender's is minted inside delivery.Claim — same token, different
		// provenance, which is why provenance is the caller's business). Apply
		// refuses if Claimed never supplied it.
		delivery.TokenConfirmLink: {FromClaim: true, Source: "the confirm token minted after ClaimFollowup"},

		// NEVER, BY CONSTRUCTION — not "empty today". A scheduled follow-up has
		// no zip step and no presign to mint, so {{zip_link}} in a follow-up
		// template is an AUTHOR ERROR to catch at dispatch, not a value someone
		// might later supply. The sentence is the point: flattening this to
		// "empty => refuse" behaves identically today and loses the reason that
		// stops a later session wiring a presign into a scheduled sender.
		delivery.TokenZipLink: {NeverReason: "a scheduled follow-up has no zip step and no presign to mint"},

		delivery.TokenInstructions: {Value: instructionsURL, Source: "instructions_url config"},
		delivery.TokenDomainRent:   {Value: stringOr(config, "domain_rent_url"), Source: "domain_rent_url config"},
		delivery.TokenDomainBuy:    {Value: stringOr(config, "domain_buy_url"), Source: "domain_buy_url config"},
		delivery.TokenStripePortal: {Value: stringOr(config, "stripe_portal_url"), Source: "stripe_portal_url config"},
		delivery.TokenDays:         {Value: fmt.Sprintf("%d", delivery.AdvertisedLiveWindowDays), Source: "delivery.AdvertisedLiveWindowDays"},
	}
	if err := fill.Check(bodyTemplate); err != nil {
		return nil, fmt.Errorf("follow-up template refused (nothing was claimed, this site is still selectable): %w", err)
	}

	now := time.Now()
	handedOverBefore := now.Add(-time.Duration(afterDays) * 24 * time.Hour)

	prepared, err := delivery.ClaimFollowup(ctx, params.DB, siteID, handedOverBefore, now)
	if err != nil {
		// The three refusals below are the mechanism WORKING, not failures, and
		// they are the common case on a scheduled run: a site the customer has
		// already confirmed, one already followed up, one not yet due or past
		// its window. Returning them as errors would fill the queue with red
		// rows for correct behaviour, so they end the step quietly and say why.
		switch {
		case errors.Is(err, delivery.ErrFollowupSuppressed),
			errors.Is(err, delivery.ErrFollowupAlreadySent),
			errors.Is(err, delivery.ErrFollowupNotDue),
			errors.Is(err, delivery.ErrFollowupWindowClosed):
			logger.Info("follow-up not sent",
				zap.String("site_id", siteID.String()),
				zap.String("reason", err.Error()))
			return map[string]interface{}{
				"sent":   false,
				"reason": err.Error(),
			}, nil
		}
		return nil, fmt.Errorf("follow-up claim failed: %w", err)
	}

	// A FRESH confirm token, not the delivery email's. Token plaintext is never
	// stored — only its hash — so the original link cannot be reproduced, and
	// re-minting is the only way to offer the button again. It expires with the
	// live-link window the claim has just proven is still open, so this email
	// cannot carry a link that outlives the site.
	confirmToken, err := delivery.MintToken(ctx, params.DB, siteID,
		delivery.PurposeConfirmTransfer, prepared.LiveLinkExpiresAt,
		false /* singleUse: see prepare.go — a second press must not fail */, "followup-email")
	if err != nil {
		// The claim stands. Loud and recoverable beats silent and duplicated:
		// the work item fails, a human sees it, and this customer is not
		// re-chased by the next tick.
		return nil, fmt.Errorf("follow-up confirm token mint failed for site %s (followup_sent_at IS stamped; a re-send is a deliberate operator act, RUNBOOK \"stamped but never sent\"): %w", siteID, err)
	}

	confirmLink, err := delivery.ConfirmTokenURL(linksHost, confirmToken)
	if err != nil {
		return nil, fmt.Errorf("follow-up confirm link could not be built for site %s (followup_sent_at IS stamped; RUNBOOK \"stamped but never sent\"): %w", siteID, err)
	}

	// The claim produced the one token Check was allowed to let through empty.
	// Claimed refuses a token this sender never declared or did not mark
	// FromClaim, so the post-claim value cannot land anywhere Check did not
	// already reason about.
	filled, err := fill.Claimed(map[delivery.Token]string{delivery.TokenConfirmLink: confirmLink})
	if err != nil {
		return nil, fmt.Errorf("follow-up fill rejected the claim's outputs for site %s (followup_sent_at IS stamped): %w", siteID, err)
	}
	body, err := filled.Apply(bodyTemplate)
	if err != nil {
		// Apply is the last gate before a customer sees the text: it refuses a
		// claim-produced token that is still empty, which is the hole the
		// FromClaim exemption would otherwise open.
		return nil, fmt.Errorf("follow-up body could not be filled for site %s (followup_sent_at IS stamped; RUNBOOK \"stamped but never sent\"): %w", siteID, err)
	}
	// BACKSTOP, deliberately kept though Check now makes it near-unreachable: it
	// asks whether Apply actually SUBSTITUTED, where Check asks whether the
	// vocabulary KNOWS. A declared token that Apply somehow failed to replace is
	// invisible to Check by construction, and this is the last thing between
	// that and literal mustache in a customer's letter. Do not delete it as
	// redundant — it now guards a different failure from the one it was written
	// for.
	if i := strings.Index(body, "{{"); i >= 0 {
		end := i + 40
		if end > len(body) {
			end = len(body)
		}
		return nil, fmt.Errorf("body_template still contains %q after filling despite passing Check: Apply did not substitute a declared token. NOTE: followup_sent_at IS stamped; this site will not be picked up again", body[i:end])
	}

	msg := mailer.Message{To: []string{customerEmail}, Subject: subject, Text: body}
	if err := sender.Send(ctx, msg); err != nil {
		// The stamp stands, deliberately. For a chase email, not sending beats
		// sending twice — and the run is red, so a human decides.
		return nil, fmt.Errorf("follow-up email send failed for site %s (followup_sent_at IS stamped; a re-send is a deliberate operator act, RUNBOOK \"stamped but never sent\"): %w", siteID, err)
	}

	logger.Info("follow-up email sent",
		zap.String("site_id", siteID.String()),
		zap.Int("after_days", afterDays))
	return map[string]interface{}{
		"sent":       true,
		"to":         customerEmail,
		"after_days": afterDays,
	}, nil
}

// followupAfterDays reads the interval from config and refuses everything that
// is not a positive whole number of days. JSON from the DB arrives as float64,
// and an operator may reasonably write it as a string, so both are accepted —
// but a MISSING or zero value is refused rather than defaulted, because that is
// the owner's undecided question and code must not answer it silently.
func followupAfterDays(config map[string]interface{}) (int, error) {
	raw, ok := config["followup_after_days"]
	if !ok {
		return 0, fmt.Errorf("followup_after_days config is required and has no default: the interval is a decision, not a fallback")
	}
	var days int
	switch v := raw.(type) {
	case float64:
		days = int(v)
		if float64(days) != v {
			return 0, fmt.Errorf("followup_after_days %v is not a whole number of days", v)
		}
	case int:
		days = v
	case string:
		if _, err := fmt.Sscanf(strings.TrimSpace(v), "%d", &days); err != nil {
			return 0, fmt.Errorf("followup_after_days %q is not a number: %w", v, err)
		}
	default:
		return 0, fmt.Errorf("followup_after_days has unusable type %T", raw)
	}
	if days <= 0 {
		// Zero would follow up the instant a site is delivered, which is a
		// second delivery email rather than a follow-up.
		return 0, fmt.Errorf("followup_after_days must be at least 1, got %d", days)
	}
	return days, nil
}
