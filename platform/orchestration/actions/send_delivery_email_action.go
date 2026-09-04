// FILE: platform/orchestration/actions/send_delivery_email_action.go
//
// The delivery email: claim the delivery (owner review gate + at-most-once
// stamp + link minting via platform/delivery.Claim), fill the operator-editable
// template, send through platform/mailer. The LAST step of Phase 4's chain.
//
// What is deliberately NOT here:
//   - No copy. The template lives in the STEP CONFIG (DB, owner-editable
//     without a roll); this file only refuses to send one that still contains
//     an unfilled placeholder.
//   - No retry. delivery.Claim is once-only BY THE STAMP: a retry of a FAILED
//     send must be a deliberate operator act on a fresh dispatch, and Claim
//     will refuse it (ErrAlreadyDelivered) so the operator must consciously
//     re-mint (documented in the seed header) rather than double-send.
//   - No SMTP secrets. mailer.FromEnv reads DELIVERY_SMTP_* from the pod env;
//     the password arrives by secretKeyRef and is never a value in config.
//
// Workflow config example (the delivery-email-sender seed, sql_for_agents/651):
//
//   "send_email": {
//     "action": "send_delivery_email",
//     "config": {
//       "site_id":            "input_data.site_id",
//       "customer_email":     "input_data.customer_email",
//       "live_site_url":      "input_data.live_site_url",
//       "zip_presigned_url":  "zip_result.presigned_url",
//       "zip_presign_minutes":"zip_result.expiry_minutes",
//       "links_host":         "links.webdesign.uk",
//       "subject":            "Your website is ready",
//       "body_template":      "Your site is live now at {{live_site}} ..."
//     },
//     "output_field": "delivery_email",
//     "next_step": "complete"
//   }
//
// Registration (registry.go): "send_delivery_email" -> SendDeliveryEmailAction.

package actions

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/gqls/agentchassis/platform/delivery"
	"github.com/gqls/agentchassis/platform/mailer"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
)

var SendDeliveryEmailInputSpec = datahelpers.ActionInputSpec{
	Required: []string{"site_id", "customer_email", "live_site_url"},
	Optional: []string{"zip_presigned_url", "zip_presign_minutes"},
	Defaults: map[string]interface{}{},
}

// newDeliverySender is the seam a test replaces. Production reads
// DELIVERY_SMTP_{HOST,PORT,USER,PASS,FROM,...} — see mailer.FromEnv — which
// arrive on the pod via env + secretKeyRef, so a missing secret fails HERE,
// loudly, before any claim is made.
var newDeliverySender = func() (mailer.Sender, error) {
	cfg, err := mailer.FromEnv("DELIVERY_SMTP")
	if err != nil {
		return nil, err
	}
	return mailer.New(cfg)
}

// SendDeliveryEmailAction claims the delivery for one site and sends the email.
//
// ORDER, and why: the SENDER is constructed FIRST, before Claim, because Claim
// stamps the handover — after which this site can never be claimed again. A
// missing SMTP secret discovered after the stamp would strand the site in
// "handed over, never emailed", recoverable only by a deliberate operator
// re-mint. Constructed-then-unused is free; stamped-then-unsendable is not.
func SendDeliveryEmailAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(zap.String("action", "send_delivery_email"))

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}
	if params.DB == nil {
		return nil, fmt.Errorf("database connection required")
	}

	inputs, err := datahelpers.ExtractActionInputs(
		params.CollectedData, params.StepConfig.Config, SendDeliveryEmailInputSpec, logger)
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

	sender, err := newDeliverySender()
	if err != nil {
		// Before the stamp, on purpose — see the function comment.
		return nil, fmt.Errorf("delivery email sender unavailable (check DELIVERY_SMTP_* env + secret): %w", err)
	}

	// TEMPLATE-VS-LINKS, ALSO BEFORE THE STAMP. Every optional link's
	// availability is knowable from config alone, so a template that names a
	// link this dispatch cannot produce is refused before anything irreversible
	// happens. Without this, {{zip_link}} with no presign would be REPLACED BY
	// AN EMPTY STRING — a customer email reading "Your files: " with nothing
	// after it, which no post-fill scan can see because the fill succeeded.
	zipPresign := inputs.Get("zip_presigned_url")
	for _, l := range []struct{ placeholder, value, source string }{
		{"{{zip_link}}", zipPresign, "zip_presigned_url"},
		{"{{domain_rent_link}}", stringOr(config, "domain_rent_url"), "domain_rent_url config"},
		{"{{domain_buy_link}}", stringOr(config, "domain_buy_url"), "domain_buy_url config"},
		{"{{stripe_portal_link}}", stringOr(config, "stripe_portal_url"), "stripe_portal_url config"},
	} {
		if strings.Contains(bodyTemplate, l.placeholder) && l.value == "" {
			return nil, fmt.Errorf("body_template names %s but %s is empty: the email would carry a blank where a link should be. Nothing was stamped — fix the template or supply the link and re-dispatch", l.placeholder, l.source)
		}
	}

	linkCfg := delivery.LinkConfig{
		LinksHost:       linksHost,
		LiveSiteURL:     inputs.Get("live_site_url"),
		DomainRentURL:   stringOr(config, "domain_rent_url"),
		DomainBuyURL:    stringOr(config, "domain_buy_url"),
		StripePortalURL: stringOr(config, "stripe_portal_url"),
	}
	if presign := inputs.Get("zip_presigned_url"); presign != "" {
		mins := 60 * 24 * 7 // the SigV4 ceiling; overridden by the zip step's own figure when present
		if m := inputs.Get("zip_presign_minutes"); m != "" {
			if _, err := fmt.Sscanf(m, "%d", &mins); err != nil {
				return nil, fmt.Errorf("zip_presign_minutes %q is not a number: %w", m, err)
			}
		}
		linkCfg.ZipPresignedURL = presign
		linkCfg.ZipPresignExpiresAt = time.Now().Add(time.Duration(mins) * time.Minute)
	}

	prepared, err := delivery.Claim(ctx, params.DB, siteID, linkCfg, customerEmail, time.Now())
	if err != nil {
		// ErrNotReviewed and ErrAlreadyDelivered surface verbatim: the first is
		// the owner's gate doing its job, the second is the once-only stamp
		// refusing a double-send. Neither is retriable by machinery.
		return nil, fmt.Errorf("delivery claim refused: %w", err)
	}

	body := fillTemplate(bodyTemplate, prepared)
	// A surviving placeholder means the template names a link this claim did
	// not produce (e.g. {{zip_link}} with no presign supplied). Refusing beats
	// emailing a customer literal mustache — but note the handover IS stamped
	// by now: the operator re-mints deliberately after fixing the config, and
	// the error says so rather than leaving them to find out at the next Claim.
	if i := strings.Index(body, "{{"); i >= 0 {
		end := i + 40
		if end > len(body) {
			end = len(body)
		}
		return nil, fmt.Errorf("body_template still contains %q after filling: the template names a link this claim did not produce. NOTE: the handover is now stamped; after fixing the template, re-dispatch and expect ErrAlreadyDelivered — recovery is the operator re-mint recipe in the 651 seed header", body[i:end])
	}

	msg := mailer.Message{To: []string{customerEmail}, Subject: subject, Text: body}
	if err := sender.Send(ctx, msg); err != nil {
		// The stamp stands (loud-and-recoverable beats silent-and-duplicated,
		// see delivery.Claim); the work item fails and a human sees it.
		return nil, fmt.Errorf("delivery email send failed (handover IS stamped; recovery = the operator re-mint recipe in the 651 seed header): %w", err)
	}

	logger.Info("delivery email sent",
		zap.String("site_id", siteID.String()),
		zap.Bool("zip_link", prepared.Links.ZipDownload != ""))
	return map[string]interface{}{
		"sent":            true,
		"to":              customerEmail,
		"zip_link":        prepared.Links.ZipDownload != "",
		"advertised_days": prepared.AdvertisedWindowDays,
	}, nil
}

// fillTemplate substitutes the closed placeholder vocabulary. Closed on
// purpose: a template author cannot invent a placeholder this code silently
// leaves standing — the {{ scan above catches both typos and inventions.
//
// ⚠ THIS VOCABULARY HAS A SECOND CALLER, AND ADDING TO IT SILENTLY BREAKS THAT
// CALLER'S GUARD. send_followup_email_action.go (bugs_open/477) reuses this
// function and carries its OWN pre-claim list mirroring the placeholders below,
// so it can refuse a template naming a link it cannot produce. That list is not
// derived from this one — it is a hand-kept copy.
//
// So: ADD A PLACEHOLDER HERE AND YOU MUST ADD IT THERE. Otherwise the new
// placeholder reaches a FOLLOW-UP email as an empty string — a customer reading
// "The instructions are here: " with nothing after it — and it is silent,
// because the fill succeeded and the post-fill `{{` scan finds nothing. That is
// precisely the failure the guard in this file exists to prevent, arriving
// through the door you did not change.
//
// This cross-reference is here rather than only in the follow-up file because a
// comment in the OTHER file protects nobody reading this one — three council
// seats made that point about a different duplicate on 2026-09-04 and they were
// right.
func fillTemplate(tpl string, p delivery.Prepared) string {
	r := strings.NewReplacer(
		"{{live_site}}", p.Links.LiveSite,
		"{{confirm_link}}", p.Links.ConfirmTransfer,
		"{{zip_link}}", p.Links.ZipDownload,
		"{{domain_rent_link}}", p.Links.DomainRent,
		"{{domain_buy_link}}", p.Links.DomainBuy,
		"{{stripe_portal_link}}", p.Links.StripePortal,
		"{{days}}", fmt.Sprintf("%d", p.AdvertisedWindowDays),
	)
	return r.Replace(tpl)
}

func stringOr(config map[string]interface{}, key string) string {
	v, _ := config[key].(string)
	return v
}
