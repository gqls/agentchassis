# 226 — chrome rebuild silently discards hand-patched content; nothing warns, nothing records what was lost

**Filed 2026-08-08 by the oufe rerender-safety lane, at the council's direction**
(trail `5c18ccaa`, round-2 gating objection from the `bug_historian` seat: the
STY-052/053 fix "re-armours a symptom rather than closing the mechanism").

**090 substitution, stated per the owner ruling of 2026-07-31:** this file did
not go through the diagnosis loop. The substitute is that the mechanism is not
a hypothesis — it is *documented platform behaviour* (`sql_for_agents/268`'s
header predicted the exact loss in writing eleven days before it happened;
`bugs_open/117`'s handoff lists "hand-patched chrome fleet-wide is one
legitimate rebuild from reset" as a standing constraint), and the damage was
measured first-hand twice this session on one site (below), artefact and wire
both. What this file asserts beyond that record is only the *absence* of a
guard, verified by reading `renderAndStoreSiteComponent` — the render replaces
`rendered_html` outright with no comparison against what it is replacing.

## The mechanism

`site_components.rendered_html` is a stored artefact (117). Any content put
there by hand — a `replace()` migration, a psql edit — is invisible to the
template+config path that regenerates it. `renderAndStoreSiteComponent`
overwrites the stored HTML with the fresh render, unconditionally, on every
legitimate rebuild (force refresh, repoint, link-policy re-mark, the 117
staleness wave). **There is no diff, no warning, no record: the platform
cannot tell "I am reproducing this artefact" from "I am destroying content
only this artefact holds."**

## The damage, measured (both on oufe.com, one rebuild, 2026-07-31 19:21Z)

1. The footer honesty note (fallibility disclosure, mig 268's protected
   object): deleted from the store and — after page reassembly — from the
   wire. Unnoticed for eight days; found by the 117 lane canary-hunting.
2. FIX_2026-07-26's header CTA rewrite: the wire served
   `<a href="/contact.html" class="header-cta">Get Started</a>` again, on a
   site whose brief forbids implying a purchase. Unnoticed for eight days;
   found only because this lane went looking for OTHER artefact-only patches
   after restoring the note.

Finding (2) is the argument that this is a class, not an incident: the first
loss was known and still nobody thought to ask "what else was in that
artefact"; a third hand-patch on any of the 16/15 sites sharing these
components dies the same way, silently, at the next wave.

## What already exists (do not rebuild these)

- **Config carriage (STY-050/051/052/053)** — a rebuild *reproduces* declared
  content. The correct destination for content that should exist; four
  consumers now, worked examples `SQL_2026-08-02d` and `sql_for_agents/339`.
  It protects only what someone has already declared.
- **069 locks (`site_components.locked_at`/`lock_type`)** — a rebuild *refuses*
  and files `lock_blocked_change`. Correct for "never touch this"; but an
  unlock loses the content, and locked slots are invisible to the 117
  staleness check by design. (Answering the bug_historian's "why not locks"
  directly: for oufe the content SHOULD evolve with the site — a reproduction
  beats a freeze, which is why 339 chose carriage. Locks remain right for
  content with no data path.)
- **117 render_inputs fingerprint (IMP-052)** — detects *input* drift. A hand
  patch changes the ARTEFACT, not the inputs, so the fingerprint neither
  detects it nor protects it; it is in fact the thing that now schedules the
  rebuild that will destroy it.

None of the three makes an **undeclared** hand-patch loud. That is the gap.

## Fix candidates, ranked by what closes the door

1. **Divergence check at overwrite time** (closes the door): in
   `renderAndStoreSiteComponent`, before replacing `rendered_html`, re-render
   with the row's *stamped* inputs (117 stores them) and compare to the stored
   HTML. Byte-equal → the store is machine-made, overwrite freely. Divergent →
   the artefact holds something the pipeline did not put there: file a work
   item naming the diff (or at minimum WARN with the lost bytes logged), then
   proceed or hold per policy. Needs the render to be deterministic given
   stamped inputs — which is precisely the property the 117 fingerprint work
   validated ("fingerprint deterministic" in its pre-proposal validation).
   Rows with no stamp (pre-117 renders, hand edits since) can't distinguish —
   which converges to correct as the wave stamps the fleet.
2. **Loss ledger only** (records, doesn't prevent): keep 1's comparison but
   only archive the outgoing HTML (a `site_components_history` row or
   `doc_notes`) whenever it differs from the incoming render. Cheap, no
   behaviour change, turns every future silent loss into a recoverable one.
   (git-for-artefacts, the same argument as the memory-dir snapshot hook.)
3. **Convention only** (closes nothing): "never hand-patch chrome; use
   STY-050 carriage or a 069 lock." Already the written rule — mig 268 obeyed
   it knowingly and the loss happened anyway, because the rule cannot reach
   the session that has not read it. A comment is not a control (owner ruling
   2026-08-02).

Candidate 1 is a platform change on a shared path (`render_site_components`)
— council gate, and probably the architecture seat's opinion on whether the
divergence branch defaults to WARN or HOLD. Candidate 2 could ship inside 1
as its first slice.

## How to verify a fix

Hand-patch a throwaway string into a test site's footer `rendered_html`,
trigger `refresh_site_components`, and require: (fix 1) a work item / WARN
naming the divergence before the overwrite; (fix 2) the outgoing HTML
recoverable afterwards. Negative control: an unpatched slot must rebuild with
no warning and no ledger row.

## Relations

`bugs_open/117` (stored-artefact mechanism + the stamp that makes candidate 1
computable) · `bugs_closed/058`, `069` (locks — the refuse-shape) ·
`bugs_open/146` (the oufe trap this class shaped) · STY-050..053 (the
reproduce-shape) · `sql_for_agents/268` + `339` (the incident pair: the
warning, then the loss, then the durable fix for the two known patches) ·
council trail `5c18ccaa` round 2 (the objection this file answers).
