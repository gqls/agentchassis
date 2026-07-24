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
