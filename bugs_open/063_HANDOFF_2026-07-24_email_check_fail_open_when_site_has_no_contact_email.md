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
session) — verdict recorded below when it lands.

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
