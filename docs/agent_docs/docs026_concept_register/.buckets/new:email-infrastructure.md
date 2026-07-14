
<!-- SOURCE: U04_idea_uk.md -->
### Operator email identity: leopardess.uk + deterministic per-site addresses + email spec aspect
- **category:** NEW:email-infrastructure
- **status-signal:** partial
- **status-evidence:** "Status: design, not yet implemented in the chassis. idea.uk… carries these values in its env"; the identity scheme is live for idea.uk (idea-uk@leopardess.uk), the aspect/provisioner are design-only.
- **what:** One neutral operator domain (leopardess.uk — also given a one-page identity site) fronts all sites' transactional/support mail. Per-site address = deterministic encoding (lowercase, dots→dashes, @operator_domain), resolved by matching against the known-domain set, never by reversing; collisions detected at assignment and stored overrides win. A new `site_specs` aspect `email` (no DDL) carries per-site identity/status/provider; a future `email-provisioner` agent flips provisioned=false→true (same provision-and-write-back shape as model-trainer/Thunder). Refined 2026-06-06: prefer a **specific forwarder per published site** over a server catch-all (no backscatter; only forward addresses that exist).
- **sources:** idea.uk/EMAIL_identity_in_site_spec(5).md; idea.uk/idea_uk_architecture_and_deployment(6).md (2026-06-05 correction); idea.uk/leopardess_uk_index.html
- **relations:** transactional sending realities; site-spec aspect model (021); feasibility-recheck promotion mechanism.
- **verify-later:** site_specs DISTINCT aspect for 'email' (expect absent); 021 doc aspect list.

<!-- SOURCE: U04_idea_uk.md -->
### Transactional email sending realities (587-only, relay filtering, SES + per-domain DKIM)
- **category:** NEW:email-infrastructure
- **status-signal:** deployed
- **status-evidence:** "DECISIVE: MailChannels blocks leopardess.uk DIRECT outbound too → must leave MailChannels" (2026-06-11); SES live in production same day; EMAIL doc header codifies the lesson for the future provisioner.
- **what:** Hard-won operational truths now standing framework guidance: cloud boxes can't use outbound SMTP 25/465 (Hetzner leaves only 587 submission open — the cPanel UI advertising 465 misleads); Go's smtp.SendMail does STARTTLS not implicit-TLS, so a 465 path needs a tls.Dial branch; shared-host relays (Clook→MailChannels) content-filter legitimate transactional mail (a `From:`-like line + raw JSON in a body triggered "Spam Content"); therefore transactional sending needs a **dedicated sender (AWS SES eu-west-2) with the operator domain's own DKIM**, bodies kept clean, and the mailer async/bounded so a hung send can't freeze the request path. Gotcha: SES SMTP_USER is the AKIA access-key-id, not the IAM user name (535s otherwise). Chronology: Clook both-ways → catch-all/Default-Address fixes → MailChannels blocks → SES.
- **sources:** idea.uk/idea_uk_architecture_and_deployment(6).md (2026-06-05/06/10/11 updates); idea.uk/EMAIL_identity_in_site_spec(5).md (2026-06-11 header + operational note); idea.uk/running_notes(63).md (email checkpoints)
- **relations:** operator email identity; deliverable quality standards (clean bodies).
- **verify-later:** /etc/idea/idea.env SMTP block (on the box); smtpSend in service.go.

<!-- SOURCE: U04_idea_uk.md -->
### Operator email identity: leopardess.uk + deterministic per-site addresses + email spec aspect
- **category:** NEW:email-infrastructure
- **status-signal:** partial
- **status-evidence:** "Status: design, not yet implemented in the chassis. idea.uk… carries these values in its env"; the identity scheme is live for idea.uk (idea-uk@leopardess.uk), the aspect/provisioner are design-only.
- **what:** One neutral operator domain (leopardess.uk — also given a one-page identity site) fronts all sites' transactional/support mail. Per-site address = deterministic encoding (lowercase, dots→dashes, @operator_domain), resolved by matching against the known-domain set, never by reversing; collisions detected at assignment and stored overrides win. A new `site_specs` aspect `email` (no DDL) carries per-site identity/status/provider; a future `email-provisioner` agent flips provisioned=false→true (same provision-and-write-back shape as model-trainer/Thunder). Refined 2026-06-06: prefer a **specific forwarder per published site** over a server catch-all (no backscatter; only forward addresses that exist).
- **sources:** idea.uk/EMAIL_identity_in_site_spec(5).md; idea.uk/idea_uk_architecture_and_deployment(6).md (2026-06-05 correction); idea.uk/leopardess_uk_index.html
- **relations:** transactional sending realities; site-spec aspect model (021); feasibility-recheck promotion mechanism.
- **verify-later:** site_specs DISTINCT aspect for 'email' (expect absent); 021 doc aspect list.

<!-- SOURCE: U04_idea_uk.md -->
### Transactional email sending realities (587-only, relay filtering, SES + per-domain DKIM)
- **category:** NEW:email-infrastructure
- **status-signal:** deployed
- **status-evidence:** "DECISIVE: MailChannels blocks leopardess.uk DIRECT outbound too → must leave MailChannels" (2026-06-11); SES live in production same day; EMAIL doc header codifies the lesson for the future provisioner.
- **what:** Hard-won operational truths now standing framework guidance: cloud boxes can't use outbound SMTP 25/465 (Hetzner leaves only 587 submission open — the cPanel UI advertising 465 misleads); Go's smtp.SendMail does STARTTLS not implicit-TLS, so a 465 path needs a tls.Dial branch; shared-host relays (Clook→MailChannels) content-filter legitimate transactional mail (a `From:`-like line + raw JSON in a body triggered "Spam Content"); therefore transactional sending needs a **dedicated sender (AWS SES eu-west-2) with the operator domain's own DKIM**, bodies kept clean, and the mailer async/bounded so a hung send can't freeze the request path. Gotcha: SES SMTP_USER is the AKIA access-key-id, not the IAM user name (535s otherwise). Chronology: Clook both-ways → catch-all/Default-Address fixes → MailChannels blocks → SES.
- **sources:** idea.uk/idea_uk_architecture_and_deployment(6).md (2026-06-05/06/10/11 updates); idea.uk/EMAIL_identity_in_site_spec(5).md (2026-06-11 header + operational note); idea.uk/running_notes(63).md (email checkpoints)
- **relations:** operator email identity; deliverable quality standards (clean bodies).
- **verify-later:** /etc/idea/idea.env SMTP block (on the box); smtpSend in service.go.
