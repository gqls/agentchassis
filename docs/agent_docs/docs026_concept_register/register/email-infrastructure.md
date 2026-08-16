# Register — email-infrastructure

> **covers-through: 2026-07-13** · extraction freeze.
> Subsystems that shipped after this date may be absent from this file
> **entirely** — absence here is not evidence of absence in the platform. See `bugs_open/106`.

_Concept count retired 2026-08-09 — derived, not stored; run the drift pair in `000_concept_index.md`, or read `concept-register-drift-check`'s daily row (DOC-074). It said **2** and the file held **0**._ consolidated from 4 raw extractions (2 unique blocks, each appearing
twice due to exact whole-block duplication in the cluster input file) across
unit U04.

### EMAIL-001 — Operator email identity: leopardess.uk + deterministic per-site addresses + email spec aspect
- **status:** partial
- **status-evidence:** "Status: design, not yet implemented in the chassis. idea.uk… carries these values in its env"; the identity scheme is live for idea.uk (idea-uk@leopardess.uk), the aspect/provisioner are design-only.
- **what:** One neutral operator domain (leopardess.uk — also given a one-page identity site) fronts all sites' transactional/support mail. Per-site address = deterministic encoding (lowercase, dots→dashes, @operator_domain), resolved by matching against the known-domain set, never by reversing; collisions detected at assignment and stored overrides win. A new `site_specs` aspect `email` (no DDL) carries per-site identity/status/provider; a future `email-provisioner` agent flips provisioned=false→true (same provision-and-write-back shape as model-trainer/Thunder). Refined 2026-06-06: prefer a **specific forwarder per published site** over a server catch-all (no backscatter; only forward addresses that exist).
- **sources:** idea.uk/EMAIL_identity_in_site_spec(5).md; idea.uk/idea_uk_architecture_and_deployment(6).md (2026-06-05 correction); idea.uk/leopardess_uk_index.html
- **relations:** Transactional email sending realities (EMAIL-002); site-spec aspect model (021); feasibility-recheck promotion mechanism
- **verify-later:** site_specs DISTINCT aspect for 'email' (expect absent); 021 doc aspect list

### EMAIL-002 — Transactional email sending realities (587-only, relay filtering, SES + per-domain DKIM)
- **status:** deployed
- **status-evidence:** "DECISIVE: MailChannels blocks leopardess.uk DIRECT outbound too → must leave MailChannels" (2026-06-11); SES live in production same day; EMAIL doc header codifies the lesson for the future provisioner.
- **what:** Hard-won operational truths now standing framework guidance: ~~cloud boxes can't use outbound SMTP 25/465~~ **CORRECTED 2026-08-16 — this is HOST-SPECIFIC, not a property of cloud boxes; see the narrowing below** (Hetzner leaves only 587 submission open — the cPanel UI advertising 465 misleads); Go's smtp.SendMail does STARTTLS not implicit-TLS, so a 465 path needs a tls.Dial branch; shared-host relays (Clook→MailChannels) content-filter legitimate transactional mail (a `From:`-like line + raw JSON in a body triggered "Spam Content"); therefore transactional sending needs a **dedicated sender (AWS SES eu-west-2) with the operator domain's own DKIM**, bodies kept clean, and the mailer async/bounded so a hung send can't freeze the request path. Gotcha: SES SMTP_USER is the AKIA access-key-id, not the IAM user name (535s otherwise). Chronology: Clook both-ways → catch-all/Default-Address fixes → MailChannels blocks → SES.
- **NARROWING 2026-08-16 (gripper dossier lane), [MEASURED] on the island, one counterexample:** *"cloud boxes can't use outbound SMTP 25/465"* is **false as a general rule** — it is true of **Hetzner**, which is the box this entry was written from. Measured from the gripper island (`toolsapisuk.vs.mythic-beasts.com`, **Mythic Beasts**, IP 176.126.243.183) against `mail.contactforsales.com`: **465 is fully open** — TLS handshake completes, Exim answers `220 rs17.uk-noc.com ESMTP Exim 4.99.5`, `EHLO` returns `250`, and **`250-AUTH PLAIN LOGIN` is advertised to that IP**. 587 is open too (same Exim banner), so both paths work there. **The lesson that survives is the CHECK, not the conclusion:** never infer a box's outbound-465 reachability from another box (a dev-box AUTH success proves nothing about the deploy target), and never from the provider's cPanel screen — *measure from the host that will actually send*, with `openssl s_client -connect <host>:465` and read for the `220`. The Go `tls.Dial`-vs-STARTTLS half of this entry is unaffected and still correct; `platform/mailer.UsesImplicitTLS` implements exactly that fork on `port == "465"`. Also captured while there, useful for capacity: the relay advertises `SIZE 52428800` and `LIMITS MAILMAX=1000`.
- **sources:** idea.uk/idea_uk_architecture_and_deployment(6).md (2026-06-05/06/10/11 updates); idea.uk/EMAIL_identity_in_site_spec(5).md (2026-06-11 header + operational note); idea.uk/running_notes(63).md (email checkpoints)
- **relations:** operator email identity (EMAIL-001); deliverable quality standards (clean bodies)
- **verify-later:** /etc/idea/idea.env SMTP block (on the box); smtpSend in service.go
