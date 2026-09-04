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

	// TEMPLATE-VS-VOCABULARY, ALL OF IT BEFORE THE STAMP.
	//
	// This was a hand-written slice of four placeholders, mirrored by a SECOND
	// hand-written slice in send_followup_email_action.go, with a comment
	// asking the next author to keep them in step. It is now DERIVED from
	// delivery.Vocabulary — one declaration, from which both the fill and the
	// refusal are taken, so a token cannot exist in the filler while being
	// absent from the guard (bugs_open/475; council c8ed56d2).
	//
	// The failure it exists to prevent is unchanged and worth restating: a
	// named token with no value is replaced by an EMPTY STRING, so the customer
	// reads "Your files: " with nothing after it — and no post-fill scan can
	// see that, because the fill SUCCEEDED. Hence pre-stamp.
	//
	// It now also refuses a token the vocabulary does not know at all. The
	// template is CONFIG (live the moment a migration applies) while the
	// vocabulary is compiled in, so a template migrated ahead of its image used
	// to survive the fill and trip the post-fill scan below — which runs AFTER
	// delivery.Claim has stamped. See LANDMINES.md, "A `body_template` is
	// CONFIG and goes live on apply".
	zipPresign := inputs.Get("zip_presigned_url")
	fill := deliveryEmailFill(config, inputs.Get("live_site_url"), zipPresign)
	if err := fill.Check(bodyTemplate); err != nil {
		return nil, fmt.Errorf("delivery email refused before claim: %w", err)
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

	// The claim has now produced the tokens that could not exist before it.
	// Apply refuses if the claim did not in fact supply one, so the FromClaim
	// exemption in Check cannot become a blank in a customer's letter.
	filled, err := fill.Claimed(map[delivery.Token]string{
		delivery.TokenConfirmLink: prepared.Links.ConfirmTransfer,
	})
	if err != nil {
		return nil, fmt.Errorf("delivery email fill failed after claim (the handover IS stamped; recovery = the operator re-mint recipe in the 651 seed header): %w", err)
	}
	body, err := filled.Apply(bodyTemplate)
	if err != nil {
		return nil, fmt.Errorf("delivery email fill failed after claim (the handover IS stamped; recovery = the operator re-mint recipe in the 651 seed header): %w", err)
	}
	// A surviving placeholder means the template names a link this claim did
	// not produce (e.g. {{zip_link}} with no presign supplied). Refusing beats
	// emailing a customer literal mustache — but note the handover IS stamped
	// by now: the operator re-mints deliberately after fixing the config, and
	// the error says so rather than leaving them to find out at the next Claim.
	//
	// ⚠ KEEP THIS EVEN THOUGH Check NOW REFUSES UNKNOWN TOKENS PRE-CLAIM. The
	// two catch different things: Check asks "does the vocabulary KNOW this
	// token", this asks "did Apply actually SUBSTITUTE it". A declared token
	// that passes Check and still survives the fill is invisible to coverage by
	// construction, and it is the one that puts literal mustache in an inbox.
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

// deliveryEmailFill is this sender's declaration over delivery.Vocabulary.
//
// Extracted rather than built inline so a test can assert it COVERS the
// vocabulary (delivery.AssertCoversVocabulary). That assertion is the mechanism
// this whole change exists for: it is what turns a vocabulary addition red for
// a sender nobody remembered to teach. Inline, the helper would have had no
// caller and the coverage guarantee would have been decorative.
func deliveryEmailFill(config map[string]interface{}, liveSiteURL, zipPresign string) delivery.Fill {
	return delivery.Fill{
		// Produced by the claim, so unknowable when Check runs. Fill.Apply
		// refuses if the claim did not in fact supply it, which is what stops
		// this exemption becoming a blank in a customer's letter.
		delivery.TokenConfirmLink: {FromClaim: true},

		delivery.TokenLiveSite: {Value: liveSiteURL, Source: "live_site_url input"},
		// The same constant Claim uses (prepare.go sets AdvertisedWindowDays
		// from it), so reading it here is not a second source of truth.
		delivery.TokenDays: {Value: fmt.Sprintf("%d", delivery.AdvertisedLiveWindowDays)},

		delivery.TokenZipLink: {Value: zipPresign, Source: "zip_presigned_url"},
		// Config KEY is instructions_url; the PLACEHOLDER is
		// {{instructions_link}}. That mismatch is the estate's convention, not
		// a slip — see delivery.Vocabulary.
		delivery.TokenInstructions: {Value: stringOr(config, "instructions_url"), Source: "instructions_url config"},
		delivery.TokenDomainRent:   {Value: stringOr(config, "domain_rent_url"), Source: "domain_rent_url config"},
		delivery.TokenDomainBuy:    {Value: stringOr(config, "domain_buy_url"), Source: "domain_buy_url config"},
		delivery.TokenStripePortal: {Value: stringOr(config, "stripe_portal_url"), Source: "stripe_portal_url config"},
	}
}

func stringOr(config map[string]interface{}, key string) string {
	v, _ := config[key].(string)
	return v
}
