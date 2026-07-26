# 063 — hallucinated-email check FAILS OPEN when the site has no registered contact email

**Filed 2026-07-24** (relojistas thread, at a council seat's insistence — bug_historian on
corr `320878ca` objected to this being "an open question in running notes" rather than a
tracked defect. It was right). **Status: OPEN.** Fleet-wide, `validate_page_content` check 5.

## Live damage that motivated it

2026-07-24 07:56: a full LLM rebuild of the relojistas.com homepage fabricated a contact
email — `mailto:relojistas@contactforsales.com` — inside `info-card-grid`. `contactforsales.com`
is not the site's domain; relojistas has **no** contact email at all (its `contacto` page's
`needs_section_data` item "Business contact email address" is still open). The build passed
validation and deployed. The fabricated address served on the live homepage until hand-repair
(~4h). A visitor mailing it reaches whoever owns `contactforsales.com`.

(The rebuild itself was triggered by a separately-fixed mis-routed emitter — see the
2026-07-24 running-notes entry in `traffic_probe/` — but the emitter only determined how
often the LLM ran, not why its fabrication was accepted. This file is about the acceptance.)

## Mechanism — one conditional

`validate_page_content.go` `validateEmails` (:617-660):

```go
officialEmail := loadSiteContactEmail(ctx, db, siteID, logger)
for _, email := range emails {
    if officialEmail != "" && email == strings.ToLower(officialEmail) { continue }
    if isPlaceholderEmail(email) { /* blocker */ continue }
    if officialEmail != "" {           // ← THE FAIL-OPEN
        /* error: email doesn't match site contact */
    }
}
```

When `officialEmail == ""` — the site has no registered contact address — a non-placeholder
email falls through **both** branches and is accepted. So:

- a site **with** a contact email is protected against invented ones (mismatch → error);
- a site **without** one — where *every* email is by definition invented — has **no
  protection at all** except the placeholder-pattern list, which a plausible-looking
  fabrication like `relojistas@contactforsales.com` sails past.

The protection is inversely correlated with the need. This is the same fail-open-on-missing-
config family as `bugs_open/026`'s schema-dialect fail-open and the claims layer's
opt-in-by-presence gate — absent configuration silently disables the check instead of
tightening it.

## Fix candidate (small)

In `validateEmails`, when `officialEmail == ""`, flag ANY non-placeholder assertion-context
email as `invalid_email` (severity `error`, same as the mismatch case), with a description
saying the site has no registered contact address so no email may be asserted. One added
`else` branch. The existing assertion-context extraction (text nodes + `mailto:` only,
fixed 2026-07-14) already prevents the placeholder-attribute false-positive class returning.

Risk to check before shipping: sites that legitimately publish third-party emails in
editorial content (a directory site listing businesses' addresses?). Assertion-context
extraction means any email in body text triggers — survey
`SELECT DISTINCT s.domain FROM ...` for live pages carrying non-official emails first; if
legitimate cases exist, severity `warning`+review rather than `error` for the no-official-
email branch may be the right floor.

## How to verify

Reproduction: a site with `sites.email` NULL/empty, a page whose content carries a
plausible fabricated email in a text node or mailto. Today: passes validation. Fixed:
`invalid_email` issue raised, build routed to review. Then re-run the 07-24 case: the
relojistas homepage rebuild would have been held, not deployed.

## Related

- `bugs_open/026` — schema-dialect fail-open: same "missing config disables the gate" family.
- `traffic_probe/relojistas_rebuild_running_notes.md` 2026-07-24 — the incident record.
- `016b §9` "spec.reason does not make needs_page scoped" — the emitter half of the incident.

---

## FIX COMMITTED 2026-07-24 (`fb3d5f5ea`) — stays OPEN, inert until an image roll

**Change** (bugfix-063 session): `validateEmails` split into a thin DB-load wrapper + pure
`checkEmails(html, officialEmail)` (the `checkDomainContamination` testability shape, 055),
and the missing `else` branch added: when `officialEmail == ""`, ANY non-placeholder
assertion-context email → `invalid_email`, severity `error` (routes the build to review,
same as the mismatch case; not a blocker). Placeholder classification keeps precedence;
mismatch branch unchanged. Five contract tests in
`validate_page_content_email_test.go`, all green against `git archive HEAD` + the two
changed files overlaid (the shared tree had another session's WIP in the same package).

**The risk check this file asked for was run before shipping** (fleet survey, live DB,
2026-07-24, regex over `page_components.rendered_html` on active pages vs the same
five-source COALESCE `loadSiteContactEmail` uses):

- Non-official emails serving on live pages fleet-wide: **3**, ALL on sites that HAVE an
  official email (finetuning.uk `finetuning@` vs official `finetune@` — a near-miss variant;
  idea.uk `idea-uk@leopardess.uk`; robot-hands.com `jane@company.com`). These are
  pre-existing mismatch escapes, not the 063 class.
- Sites with NO registered contact email: **20 of 31**. Zero of their live pages assert any
  email → **the new branch flags nothing on today's fleet; false-positive exposure at ship
  time is zero.** No directory-style legitimate-third-party-email site exists, so severity
  `error` was chosen over the `warning` fallback contemplated above.

**Council:** submission corr `7080124b-716f-45ac-8d42-f24465228b4b` (2026-07-24, this
session). **R1 = REVISE** (6 approve, 2 object, 8 abstained; gating objection:
prior_art_librarian, high) — a submission-authoring defect, not a fix defect: I sketched
the edits as *proposals* while the rationale said "committed as fb3d5f5ea", which reads as
the dormant-machinery pattern. Same final-state-sketch lesson the relojistas thread logged
this week; WRONG_CALLS row added. The two answerable checks R1 surfaced, both run:
- **guardian (medium): consumer breadth** — `validate_page_content` has THREE active
  consumers (`page-build-handler`/`validate_content`, `content-reviewer`/`validate_content`,
  `tool-recreation-handler`/`validate_tool`). All three already carry `invalid_email`/error
  via the mismatch branch — no new severity class anywhere; the new branch widens the
  triggering inputs symmetrically, with a measured-zero triggering population at ship.
  Config-flag opt-out rejected: it would recreate the 026/063 fail-open family.
- **prior_art (medium): test absence** — `git grep -l validateEmails fb3d5f5ea^ --
  '*_test.go'` → 0 files; the pre-existing trio is claims/contamination/meta.

**R2 resubmitted** same day with final-state sketches (verbatim `git show fb3d5f5ea`
hunks) + all checks attached (run `5f884438`). **R2 = APPROVED** (2026-07-24 21:04 UTC,
12 reviewers, 4 abstained, 0 unreadable; verified in `diagnosis_artifacts`
`metadata->>'decision'` for corr `7080124b`). Trailer `Council-Reviewed:
7080124b-716f-45ac-8d42-f24465228b4b` carried on the follow-up docs commit recording
this verdict (forward-only repo — the fix commit `fb3d5f5ea` predates the verdict and
is never amended; the 098 report joins on the trailer wherever it appears).

**Follow-up worth its own pass (compliance seat, R1):** 063 is the second member of the
fail-open-on-missing-config family in validation code (with 026) — audit the OTHER
`validate_page_content` checks for the same missing-else shape. Not this fix's scope.

**Premise shift to note:** relojistas NOW has `sites.email = 'relojistas@contactforsales.com'`
— the very address this file calls fabricated — `[OBSERVED 2026-07-24, source unknown]`; it
was NOT set by the hand-repair (the running notes show the repair *removed* the mailto).
Three other sites (finetuning.uk, idea.uk, robot-hands.com) have official emails at
`contactforsales.com`, so it appears to be the owner's for-sale-contact domain and the
registration likely legitimises the address after the fact. That softens the "reaches
whoever owns contactforsales.com" damage line above, but does not touch the mechanism: the
fabrication deployed while the site had no registered email, and 20 sites remain in that
state.

**Close criteria** (fixed AND live, per the /bugs_closed/ bar):
1. Image roll carrying `fb3d5f5ea`; pod-grep a string this change CREATED
   (`no registered contact address`) + a positive control.
2. **Behavioural, the failing branch**: induce a fabricated non-placeholder email on a
   no-email site's page build → expect `invalid_email` error and the build routed to
   review, not deployed. (A green happy path + pod-grep proves deployment, not detection.)

---

## CLOSED 2026-07-26 — fixed AND live on v1.0.1165, both branches induced

**Criterion 1 — deployment.** Pod `agent-chassis-f4d46c88d-p6wqc`, image `v1.0.1165`:
`strings /app/agent-chassis | grep -c "no registered contact address"` → **1** (a string
this change CREATED); positive control `placeholder_email` → **1**.

**Criterion 2 — the failing branch, INDUCED** via the scratch one-step probe harness
(`durable_write_guard/RUNBOOK` — reused as designed): scratch agent_definition
`scratch-063-email-probe`, sole step `validate_page_content` reading
`html_field`/`site_id` from `input_data.*`, fired with the 091 kcat `orchestrate`
envelope. Expected verdicts were written down BEFORE firing (target site's no-email
status re-verified with the code's exact five-source COALESCE).

- **Probe A — must FLAG** (corr `345b113e`, 2026-07-26 13:14 UTC): gamesdesign.co.uk
  (`e33263f4`, no email in any of the five sources), HTML asserting a fabricated
  `hello@gamesdesign-enquiries.co.uk` (text node + mailto). Result: step `validate`
  FAILED — `content validation failed: 0 blockers, 1 errors` — routed to
  `complete_error`; `agent_error_log` detail row (`CONTENT_VALIDATION_BLOCKER_DETAIL`,
  site_id = gamesdesign) carried exactly
  `{"type":"invalid_email","severity":"error","value":"hello@gamesdesign-enquiries.co.uk",
  "description":"Email 'hello@gamesdesign-enquiries.co.uk' asserted but the site has no
  registered contact address — no email may be published"}`. On the pre-fix binary this
  exact input fell through both branches and passed.
- **Probe B — must PASS** (corr `8f8c1d7c`): dartsonline.com + its own official
  `darts@contactforsales.com` → `valid=true, issue_count=0, checked_emails=2`, clean
  route to `complete`. No false positive on the legitimate case.

**All probe rows were deleted after reading** (1 scratch agent_definitions, 2
orchestration_states + 13 audit rows, 2 agent_error_log rows; leak check 0/0/0/0 —
else the immune sweep triages a deliberate probe as a real failure). This section is
the preserved evidence.

**Scope of what was proven:** the probes prove the DETECTION branch live (error issue +
step failure) and the non-flagging of a legitimate address. The routed-to-review
consequence is the consumers' standing handling of a failed validate step — the path
the mismatch class already exercises in production; not re-proven here.

**Fleet shift since the 07-24 survey** (re-grounded live 2026-07-26): the owner has
since registered `<site>@contactforsales.com` addresses across the deployed fleet —
the no-email class is now **1 deployed site** (gamesdesign.co.uk) + 19 internal
pool/system sites, down from 20/31. The fix still matters: every new site starts
email-less, which is exactly the state relojistas was in on 07-24.

**Residue: none on this mechanism.** The compliance follow-up (audit the OTHER
`validate_page_content` checks for the same missing-else fail-open shape) stays open
as a follow-up recorded above — it is not this bug.
