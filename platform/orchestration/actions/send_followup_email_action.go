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
//   - No hosting steps. The email carries {{instructions_url}} and never the
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
//       "body_template":       "... {{live_site}} ... {{instructions_url}} ... {{confirm_link}}"
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

	// TEMPLATE-VS-LINKS, ALSO BEFORE THE CLAIM, inherited from the delivery
	// send.
	//
	// ⚠ THIS IS A LOCAL GUARD AND THERE IS NO SHARED ONE — stated explicitly
	// because a council seat asked whether this reinvents a platform mechanism
	// (016b §9 case 7: Go templates render a missing field as empty with NO
	// error under missingkey=zero, and only one call site was ever fixed). It
	// does not reuse one, because none exists: `grep -rn "missingkey"` over
	// platform/ and internal/ finds only COMMENTS describing the hazard, and the
	// only other placeholder-refusal in the estate is send_delivery_email's,
	// which this copies deliberately rather than shares. Note this seam is not
	// even text/template — it is a strings.Replacer over a closed vocabulary, so
	// a shared template guard would not cover it anyway. If a shared mechanism is
	// ever built, these two are its first two callers. Every link this email can carry is knowable from config alone, so a
	// template naming one this dispatch cannot produce is refused before
	// anything irreversible happens. Without it, {{zip_link}} would be replaced
	// by an EMPTY STRING and the customer would read "Your files: " with nothing
	// after it — which no post-fill scan can see, because the fill succeeded.
	//
	// The list is every placeholder fillTemplate knows MINUS the two this action
	// always produces ({{live_site}} from input, {{confirm_link}} from the mint
	// below). If fillTemplate's vocabulary grows, this list must grow with it:
	// the {{ scan after filling is the backstop that makes a miss loud rather
	// than silent, but it fires AFTER the claim, which is worse.
	instructionsURL := stringOr(config, "instructions_url")
	for _, l := range []struct{ placeholder, value, source string }{
		{"{{instructions_url}}", instructionsURL, "instructions_url config"},
		{"{{zip_link}}", "", "a zip presign, which a scheduled follow-up has no step to mint"},
		{"{{domain_rent_link}}", stringOr(config, "domain_rent_url"), "domain_rent_url config"},
		{"{{domain_buy_link}}", stringOr(config, "domain_buy_url"), "domain_buy_url config"},
		{"{{stripe_portal_link}}", stringOr(config, "stripe_portal_url"), "stripe_portal_url config"},
	} {
		if strings.Contains(bodyTemplate, l.placeholder) && l.value == "" {
			return nil, fmt.Errorf("body_template names %s but %s is empty: the email would carry a blank where a link should be. Nothing was claimed — fix the template or supply the link and re-dispatch", l.placeholder, l.source)
		}
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

	body := fillTemplate(
		strings.ReplaceAll(bodyTemplate, "{{instructions_url}}", instructionsURL),
		delivery.Prepared{
			Handover: prepared,
			Links: delivery.Links{
				LiveSite:        inputs.Get("live_site_url"),
				ConfirmTransfer: confirmLink,
			},
			AdvertisedWindowDays: delivery.AdvertisedLiveWindowDays,
		})
	if i := strings.Index(body, "{{"); i >= 0 {
		end := i + 40
		if end > len(body) {
			end = len(body)
		}
		return nil, fmt.Errorf("body_template still contains %q after filling: the template names a link this dispatch did not produce. NOTE: followup_sent_at IS stamped; this site will not be picked up again", body[i:end])
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
