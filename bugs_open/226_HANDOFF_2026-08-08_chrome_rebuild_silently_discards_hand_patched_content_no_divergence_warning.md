# 226 — chrome rebuild silently discards hand-patched content; nothing warns, nothing records what was lost

> **STATUS: DONE IN SUBSTANCE 2026-08-09 — stays in `bugs_open/` per the owner
> ruling of 2026-08-06.** All three close criteria met (pod-verified v1.0.1270,
> re-verified v1.0.1274 post-roll; e2e protocol passed on dartsonline; both
> convergence routes observed, zero trigger errors ever). Fleet fingerprinted
> 54/60 by the owner-ordered restamp pass (rerender-chrome, mig 351/STY-055);
> the 6 unstamped slots are permanent-locked authored chrome, stamping on
> unlock. The guard is in production and has already refereed a real event:
> the fixer-vs-rebuild loop on dartsonline's header, OPEN item
> `…header:20b7c324b983` (needs_human_review — re-declare via STY-050 or
> dismiss; that decision is the queue's, not this lane's). Page-side sibling =
> `bugs_open/229`, OWNER CALL. Full trail: `bugfix_226_chrome_divergence/`.

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

> **CORRECTED 2026-08-08 (fixing lane, session `1eb2bf20`): candidate 1 as
> written below cannot be built.** "Re-render with the row's *stamped* inputs"
> assumes the stamp holds inputs; `render_inputs` is a jsonb map of md5
> **digests** of the input stores (`datahelpers/chrome_render_inputs.go` — "a
> map of NAMED per-input digests"), which detects input drift but cannot
> reproduce a render. This file's own "What already exists" section describes
> the fingerprint correctly; the candidate contradicts it two paragraphs later
> — the LANDMINES "fix candidate refuted by its own file" shape, caught by
> reading the helper before planning. **Shipped instead (same intent, stronger
> mechanism):** stamp `rendered_html_digest = md5(rendered_html)` in the SAME
> statement that stores the bytes; a mismatch at the next overwrite IS the
> hand-patch signal — no re-render, no determinism assumption. The archive
> half is a DB **trigger** (`trg_site_component_archive`, mig 344, live
> 2026-08-08) so the raw-psql writer class — this bug's origin — is covered,
> along with every other writer present and future; candidate 2's ledger ships
> inside it as `site_component_history`. Divergence WARNs + files
> `chrome_divergence_overwritten`, then proceeds (holding is the 069 locks'
> job). Register: STY-054. Plan + council corr `cffbfec4`:
> `docs/agent_docs/docs024_key_docs_latest/bugfix_226_chrome_divergence/`.
> Still true and unchanged from the candidate: unstamped rows (46/57 at fix
> time) cannot be classified and converge as the fleet re-renders.
>
> **COUNCIL APPROVED 2026-08-09 09:08Z — round 3 of trail `cffbfec4`** (r1
> REVISE guardian, r2 REVISE bug_historian, r3 approved with 2 advisories,
> none high). Standing advisories to carry: (a) the ledger read-back's
> `ErrNoRows` = "nothing lost" holds only while the 344 trigger's WHEN gate
> has no false negative — the same trust surface as the archive itself, filed
> in STY-054's open-review; (b) the page-side exposure stands until the OWNER
> CALL in `bugs_open/229` is taken.
>
> **CLOSE CRITERIA (this bug stays OPEN until all three, per the fixed-AND-live
> bar; the DB half alone is not the fix):**
> 1. **Pod-verified image** (the 153 discipline — a roll is not evidence).
>    **CORRECTED 2026-08-09 (round-2 `debug_historian` objection, confirmed by
>    measurement): "every replica" via the standard label selector is the
>    documented one-image-many-labels trap — 65 pods ran this image at check
>    time, most of them spawn pods.** The honest enumeration is by IMAGE at
>    the deployment level, then binary-grep the deployment that actually runs
>    `render_site_components` (the main `agent-chassis` deployment):
>    `kubectl get pods -o jsonpath='…image…' | grep agent-chassis | sort | uniq -c`
>    then, on each main-deployment pod:
>    `strings /app/agent-chassis | grep -c classifySiteComponentArtefact`
>    (expect >0) AND the negative control
>    `grep -c "hand-patched bytes are being overwritten"` — a round-1 log
>    string round 2 REMOVED (expect 0; the live string ends `were overwritten
>    and archived`).
>    **DONE 2026-08-09 for v1.0.1270**: all three chassis deployments
>    (agent-chassis 2/2, business-intel, vet-intel) at 1270; both main
>    replicas grepped 2 / 0 / 1. Residual: ~60 spawn pods still at 1269
>    (pre-roll spawns; they do not consume the render step and are reaped on
>    completion) — noted, not claimed.
> 2. **The verification protocol above run end-to-end**: hand-patch a test
>    site's footer artefact, rebuild, require the WARN + the
>    `chrome_divergence_overwritten` item + the ledger row; negative control:
>    a stamped, untouched slot rebuilds with no item and no archive row when
>    byte-identical. **Note the two-step**: all 57 rows are unstamped until
>    their first post-roll render, so the protocol needs rebuild → (stamp) →
>    hand-patch → rebuild → (hand_patched + item).
>    **DONE 2026-08-09 ~09:25Z on dartsonline.com** (the wave had already
>    stamped its 3 slots at 09:08:30Z, so step (a) was free). Evidence, all
>    by row identity not count: the psql patch itself drew a `machine_made` /
>    `psql` ledger row (raw-psql writer class proven visible); the forced
>    rebuild (orch `322b266e`, dispatched via kcat with PUBLISH_OK receipt,
>    consumed in ~5s) fired the WARN once on pod zhz2g (log line captured
>    live — chassis retention is seconds, followers were armed first),
>    archived the patched bytes as `hand_patched` (archived md5 == patched
>    md5 `2ed6dd06…`; `application_name` carries zhz2g's IP), filed
>    `chrome_divergence_overwritten:site_component:footer:2ed6dd067c5f`
>    (digest-fragment key as designed), and re-stamped (archive .907s →
>    item .931s: archive atomic with overwrite, item after RowsAffected).
>    Negative control (orch `453b2eb6`): all 3 slots demonstrably re-written
>    (updated_at bumped) byte-identical — no WARN, no new ledger row, no new
>    item. Probe item then CANCELLED with a note naming this protocol run
>    (a deliberate probe is not a queue item for a human). Full trail:
>    `bugfix_226_chrome_divergence/NOTES_chrome_divergence_guard.md`.
> 3. **The 117 wave's first pass observed**: archive rows present, zero
>    `trg_site_component_archive` errors in the wave's window. **Partial
>    evidence 2026-08-09 (pre-wave): four production overwrites (webdesign.uk
>    + leopardessconsulting.co.uk, header+footer each) were archived
>    unprompted on 08-08 evening, all `unstamped` as expected pre-roll, zero
>    trigger errors — the archive works on real traffic. The wave itself has
>    not fired (0/57 digests stamped).**
>    **CORRECTED 2026-08-09 13:10Z — READ THIS BEFORE THE PARAGRAPH BELOW IT,
>    WHICH OVERSTATES THE WAVE.** There is no fleet-wide wave and no scheduled
>    task that would run one. Convergence happens two ways, both unscheduled:
>    (i) a `needs_rerender`/`render_inputs_drift` item filed by a discovery
>    agent, IF something dispatches it; (ii) **any ordinary chrome render at
>    all** — `mortgagecalculator.co.uk` was stamped 3/3 at 11:26Z by a
>    `nav-updater` run with no drift item involved. As of 13:06Z: **6/57 slots,
>    2/19 sites** (dartsonline, mortgagecalculator), zero trigger errors since
>    the trigger went live, guardian ratio still 1:1.
>    **Route (i) is currently NOT draining.** The three drift items that
>    completed were filed by `design-discovery-agent` on the old
>    manually-dispatched path and were triaged in 4m47s / 4m17s / 4m40s. The two
>    newest — `robot-hands.com` 09:50Z and `idea.uk` 12:51Z, both `created_by`
>    `generic` — have **never been triaged** (3h18m and 17m at measurement).
>    They come from the new hourly `site-discovery-rotation-*` scheduled tasks
>    shipped 2026-08-09 by the `bugfix_230_discovery_driver` lane (mig 346),
>    which that lane's own handoff states are **observe-only**, with the drain
>    half left to `bugs_open/083` ("detected findings never reach a handler",
>    OPEN). So detection now recurs hourly; dispatch does not follow it.
>    **Consequence for this criterion: as written it has no completion event to
>    wait for.** Rewritten to something reachable — the criterion is met when a
>    site is observed converging by BOTH routes with zero
>    `trg_site_component_archive` errors, which route (ii) satisfied at 11:26Z
>    and route (i) satisfied at 09:08Z. The remaining 17 sites are then a
>    convergence *rate* question owned by 083/230, NOT a blocker on 226 — this
>    bug's mechanism is proven closed at every writer class. Whoever picks the
>    lane up should record the 51 unstamped slots as a known, shrinking window
>    (a hand-patch on an unstamped slot is still ARCHIVED at its destruction —
>    only the work item is missed) and stop treating a wave as pending.
>    The reasoning error that produced the paragraph below — counting arrivals
>    and calling it throughput — is logged in `WRONG_CALLS.md`.
>
>    **RESTAMP PASS EXECUTED 2026-08-09 13:52–13:54Z (owner instruction:
>    "please dispatch the rebuilds") — the unstamped window is now CLOSED
>    except where a human lock holds it open.** Mechanism: `rerender-chrome`
>    (STY-055, mig 351) — a stamp-only two-step agent seeded for this pass
>    because every existing chrome-rendering agent drags page fan-out or
>    template mutation along (measured in the migration header). 15 sites
>    dispatched (kcat receipts 15/15), 15 orchestrations COMPLETED, ~8s apart.
>    Result: **54/60 slots stamped and matching**; the 6 unstamped are the two
>    `permanent`-locked authored-chrome sites (loancalculator,
>    loanandmortgagecalculator) whose lock refusal is the 069 prevent-leg
>    working — they stamp when an owner lifts the locks. 10 slots' stored
>    bytes differed from the fresh render and were archived
>    (divergence `unstamped`, recoverable); 35 re-rendered byte-identical, no
>    archive spam. Zero trigger errors.
>    **The guard caught its first REAL production event during this window,
>    twice over**: dartsonline's header is being re-patched (+390 bytes, then
>    a 3486-byte variant) by `component-template-fixer` runs (13:38Z, 13:41Z
>    writes; ledger `hand_patched` archives at 11:59Z and 13:41Z), and every
>    rebuild wipes it — the 11:59Z `chrome_divergence_overwritten` item
>    (key `…header:20b7c324b983`, needs_human_review, left OPEN deliberately)
>    is the alarm saying exactly what the PLAN's consumers-told section
>    predicted for TRANSIENT chrome patches. Dedup holds further items while
>    it stays open. Side observation for its own lane: relojistas' fresh head
>    renders at 61KB (the shared Document Head component's theme-injection
>    slot inlines that site's 52KB theme CSS; store-only until pages
>    reassemble; 9.4KB predecessor archived).
>
>    ~~IN PROGRESS 2026-08-09 ~09:30Z: the wave HAS started~~ — discovery
>    (`needs_rerender`, reason `render_inputs_drift`) has completed 3 sites:
>    leopardessconsulting.co.uk 08-08 17:21Z and webdesign.co.uk 08-08 18:15Z
>    (both pre-roll ⇒ rebuilt unstamped, correctly), dartsonline.com 08-09
>    09:08Z (post-roll ⇒ 3/3 stamped, byte-identical renders, zero archive
>    rows — the WHEN gate's no-op path on real traffic). 3/57 stamped
>    fleet-wide; zero trigger errors since the trigger went live; guardian
>    accumulation check 1 item : 1 distinct patched digest (the probe's own).
>    NOTE the 08-08 archive rows (webdesign.uk 22:29Z, leopardess 23:43Z)
>    were OTHER rebuild traffic, not those wave items — the wave's two
>    pre-roll sites predate the trigger and archived nothing. Remaining: the
>    other 16 sites as discovery reaches them — same three queries, in the
>    NOTES runbook section.

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
