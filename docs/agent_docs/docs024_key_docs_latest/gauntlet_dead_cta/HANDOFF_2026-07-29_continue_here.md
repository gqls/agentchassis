# HANDOFF — gauntlet_dead_cta, 2026-07-29

**Supersedes `HANDOFF_2026-07-28_continue_here.md`** (whose §2c/§2d deltas are
now history — everything they promised is shipped). Keep the 07-26b file for
the API contract; keep the 07-28 file for the record of how the day unfolded.

## 0. Read this first

Multiple sessions share this repo and cluster. Commit per task by explicit
pathspec, forward-only. Re-verify anything load-bearing before planning around
it — this file was written the morning after a day on which the fleet rolled
**nine times**, the shared tree changed branch mid-session, and two bug numbers
(083, 131) each name two unrelated cases. Resolve bugs BY SLUG.

## 1. State in one paragraph

**The owner's usability audit (131 A–G, the gauntlet one) is fully engineered —
filed at breakfast 07-28, everything live by that night.** The page opens
sealed behind one door; pressing Enter the Gauntlet starts a real round and the
engine's answer reveals the question, in its own marked card, with the input
directly beneath; the steps carry a state-driven pecking order; a finished
round yields a shareable 1200×630 card drawn from the visitor's own verdict.
`bugs_closed/130` (the fleet's only two untimed HTTP clients) is closed with
both halves proven live. What remains is not engineering: **H — why argue here
rather than in a free chat window — is the open product question** (§7).

## 2. What is live (all verified on the served artefact / running binary)

| piece | state | proof |
|---|---|---|
| Sealed reveal (C) | LIVE | 22-check harness + 16 live checks; corr `824c7f1c`; commit `c2969cbff` |
| Question card + step ranking (E+F) | LIVE | 18-check full-round harness + 11 live; corr `7b0accf3`; commit `ba7222b9b` |
| Verdict share-card (G, owner pick) | LIVE | real 60KB 1200×630 PNG pulled off the click in-harness; corr `ba1666a7`; commit `dce85ccd8` |
| 130 timeout ceiling — chassis | LIVE on 1196 | ancestry pod-grep: 124 marker (`af0cde87d` ⊇ `a554bc914`) greps 1, control 0 — re-checked after the 22:37Z roll |
| 130 timeout ceiling — island | LIVE | binary sha256 `7196ca8b…` byte-identical local↔running + live round; owner ran the swap |
| 131-B clipped-overflow check | LIVE on 1196, **unwitnessed on the acceptance lane** | fix marker 1 / positive 1 / negative 0 on the 1196 pod (grep the BINARY — no `strings` in this image) |
| Island engine | `aqls/tools-api:v1.0.1193` | compose bak: `docker-compose.yml.bak-1193` |

Fleet context: v1.0.1196 rolled 22:37Z 07-28. A `render-audit-adapter` pod
(another workstream's, 1194) appeared the same evening.

## 3. Next actions, highest value first

1. **§7 — the H decision.** Owner's call; the options are laid out there.
   Everything else on this page is polish until it is answered.
2. ~~**Witness 131-B on the acceptance lane.**~~ **DONE 2026-07-29 12:30Z — 131-B
   is CLOSED.** The deployed adapter (1197) failed `mobile-fit@mobile` on
   `tool-loot-table-balancer` with an attributed culprit on a page whose
   `scrollWidth` is clean — the one shape only the new clause produces. Found by
   scanning all 94 deployed tool pages with the clause extracted from the Go
   source: 86 clean, 8 flagged, **1** on the new branch. Evidence in
   `bugs_open/131` § "B check-side — WITNESSED AND CLOSED"; pattern in 016b §9.
   **Two live consequences for whoever is next:** (a) `improve_tool`
   `e7ea0125` now owns a genuinely broken page — ~~if it cycles without fixing,
   that is `bugs_closed/010`'s guard, expect escalation at
   `fix_cycles_spent=2`~~ **CORRECTED same session: it will not cycle at all.**
   The judge hardcodes the item's status to `detected`, and the only promoter to
   `triaged` lives in a scheduled task disabled since 2026-05-02 — so the fixer
   never sees it. That is `bugs_open/083` **by slug**
   (`…detected_findings_never_reach_a_handler`, NOT the gauntlet-engine-503 083),
   which I contributed the measurement to: 7 of 7 `improve_tool` items since
   07-17 parked, pile now 250 fleet-wide. **The page stays broken and that is
   not this lane's to fix** — do not hand-promote the row; 083 is waiting on an
   owner call and `bugs_open/126` is why aiming a fixer by hand is risky.
   (b) 010's line *"tool-loot-table-balancer passes Tier 4 now"* is retired — it
   passed a check that could not see this defect.

   > **SUPERSEDED BY OWNER INSTRUCTION, 2026-07-29 ~17:05Z: "please run that
   > scheduled task once."** The "do not hand-promote" line above stands as
   > written — nothing was hand-promoted — but the question it was waiting on has
   > been answered by running the machinery instead. One firing of
   > `improvement-sweep` at gamesdesign.co.uk (script:
   > `scripts/run_improvement_sweep_once.sh`, orchestration `30692439`) promoted
   > **67** items in one statement, `e7ea0125` among them, and dispatch followed
   > on its own within a minute. **So the promoter is not broken, it is unrun** —
   > 083's premise demonstrated rather than argued, and contributed back there
   > along with the narrowing that matters: **not every agent inserts at
   > `detected`** (`rerender-pages` inserted 32 of 32 directly as `triaged`),
   > which is why the queue never looked starved and nobody noticed.
   >
   > **Two things the next session must not inherit wrongly.** (1) The sweep
   > picks its OWN site — `ORDER BY updated_at ASC LIMIT 1` — and filing a work
   > item bumps that column, so a fresh finding sorts LAST; the script takes the
   > site as an argument for exactly this reason. (2) The run surfaced a
   > **separate defect, `bugs_open/150`**: the loop ended at `complete_clean`
   > ("No issues found — site is clean") straight after promoting those 67, and
   > skipped its own dispatch branch, because `triage_detected_items` is a step
   > in three agents and the parent's copy runs last. Anyone firing the sweep to
   > clear a backlog will be told the opposite of what happened.
   >
   > Scale, so nobody repeats my understatement: the discovery half outnumbered
   > the promotion half ~60:1 on a site unswept since May — 1 detected item
   > became 64 in three minutes, and the fixers are still working the queue
   > hours later, including LLM copy rewrites on live pages.
3. **og:image** — vonc's gauntlet page emits a 404 social image (owned by the
   OTHER bug numbered 131, `…og_image_points_at_a_card…`, UNDIAGNOSED). When
   fixed, shared verdict cards and links get a face. Do not build a generator
   here; contribute into that file if you touch it.
4. **Consolidation ping**: features_open/024 wants `platform/mailer` +
   `platform/httpguard` adopted into tools-api and its memory line says
   "message them first" — this workstream owns tools-api, so expect (or
   answer) that contact. Nothing owed yet.
   > **NOTE ADDED 2026-07-30 by the consolidation thread (features_open/024) —
   > appended, nothing above edited.** *"Nothing owed yet"* was written at 08:22
   > on 07-29 and is now out of date: the contact arrived at 13:34 that day as
   > `CONTRIB_2026-07-29_tools_api_client_identity_is_a_constant.md`, **in this
   > directory**, and it is not a request — it is evidence plus a ready patch.
   > **The finding is about YOUR service and does not depend on the consolidation
   > programme at all:** `client_ip_hash` is `sha256("172.18.0.1")` (the docker
   > bridge gateway) in **83 of 83 rows** since 07-25 — one distinct value, whole
   > table. So the per-IP limiter is one global bucket shared by every visitor,
   > and the stored identity column has never distinguished anybody. No attacker
   > involved; measured, not inferred.
   > The patch is three small edits following **your own `httperr` precedent** (a
   > `clientip` package + two one-line call-site changes at
   > `middleware/ratelimit.go:30` and `handlers/round.go:109`). It is yours to
   > take, reject or rewrite — **we are deliberately not touching tools-api.** One
   > thing in it is marked `[INFERRED]` and you are better placed than us to
   > settle it: that `CF-Connecting-IP` reaches the app process. We measured it
   > arriving at Caddy and measured Caddy forwarding it verbatim, but tools-api
   > has no header-echo endpoint and we would not add one to your service.
   > Acceptance check is in the CONTRIB: `count(DISTINCT client_ip_hash)` must
   > exceed 1 across visits from two different networks — a presence check passes
   > against the unfixed code, so do not use one.
5. **083 (gauntlet-engine-503, BY SLUG)** posture unchanged: armed log, check
   on a TRIGGER not a date. Its only catch remains the 07-28 cap event.
   Candidate 4 (max_tokens) still awaits a real burst of successful traffic.

## 4. Landmines earned 07-28 (the RUNBOOK holds the older list)

- **A docker IMAGE ID is not portable across `save|load`** — engine
  re-serialization gave two IDs for byte-identical content. Hash `/tools-api`
  (the BINARY), never compare image IDs across hosts.
- **A retag is not a rebuild** — 1188/1189 shared one image id built BEFORE
  the fix "in" them. Check `.ID`+`.CreatedAt`; re-grep after any roll you
  did not do.
- **`content-box` means `max-width:100%` is not a cap** — padding rides on
  top (measured 398px on a 358px row, twice). `box-sizing:border-box`,
  `min-width:0`, `max-width:100%` travel together.
- **`strings` does not exist in the browser-runner-adapter image** — grep the
  binary at `/app/browser-runner-adapter`; a strings-based check reads a
  successful deploy as failed. Always pair with a positive control.
- **The hardened kcat trigger can still double-fire** (two PUBLISH_OK, two
  orchestrations). Idempotent for assemble-only rerenders; NOT necessarily
  elsewhere — count orchestrations by payload after any dispatch.
- **The island has a dedicated KEY but shares the ORG spend limit** — key
  rotation cannot escape a cap. (07-28's outage capped the island too.)
- **Characters are not bytes, third bite of the week** — `len()` vs `wc -c`,
  `length()` vs `octet_length()`. When two sizes of "the same" content
  disagree, name the units before naming a culprit.
- **A "fresh visitor" is a NEW browser context** — clearing sessionStorage in
  a live page is undone by the `pagehide` save re-persisting a live round.

## 5. The rail, and its two deliberate exceptions

Nothing on vonc.com claims a number that is not true by construction; no
control changes state except as the consequence of a real API response —
including the REVEAL (only /round 200 or a live stored-round resume removes
`gi-sealed`) and the CARD (the button lives inside the verdict step, hidden
until /defend returns; every string on the card is a fact of that round).
`input_schema.llm_guidance` states the rail at every field. The two standing
exceptions, so nobody "fixes" them: the sessionStorage round resume, and the
deleted feed pre-render (its reversal is commented at the code — restoring it
would return the Enter button to revealing nothing).

## 6. Where things are

- Sources + backups, all committed: `p4_sources/*2026-07-28{c,d,e,f}_*`
  (c=flow order · d=sealed reveal · e=question card+emphasis · f=verdict card),
  builders `build_{reveal,ef,g}*.py`, harnesses `drive_{reveal,ef,g}*.py`,
  live verifiers `verify_live_*131*.py`.
- The delivery pattern that worked four times: guarded transaction
  (`WHERE updated_at = <read value>`) on template+js+rendered → assemble-only
  rerender (script `rerender_arena_vonc.sh` adapted per page) → live verify
  with the SHIPPED overflow JS extracted from the Go source.
- Commands + gotchas: `RUNBOOK_gauntlet_dead_cta.md`; island:
  `infra/island/RUNBOOK_island.md`.
- Plain-prose history: `README_where_we_are.md` (append-only).
- Milestone read-out: `SUMMARY_2026-07-28_gauntlet_dead_cta.md`.

> **CORRECTED 2026-07-29 (vonc6, same day):** the ruling's first build — the
> opinion ledger — is **SHIPPED and live-verified** (25-check harness + 13
> live checks incl. a real round on the served page). En route, 083 (by slug)
> was **CLOSED**: the armed log caught the judge thinking itself out of its
> 2048-token budget (claude-sonnet-5 thinks by default); fix live on island
> `v1.0.1198`, council APPROVED r1, 6 clean verdicts since. See
> `SUMMARY_2026-07-29` + NOTES tail. §3's item 1 is done; item 5's posture
> moves to CLOSED. Two new small facts: the page serves ONLY at
> `/tools/gauntlet/index.html` (no directory index — both bare variants 404),
> and `scripts/rerender_gauntlet_vonc.sh` is the hardened rerender for this
> page (prefer it over `republish_gauntlet_js.sh`, which still carries the
> racy kcat stdin form).

## 7. H — DECIDED by the owner, 2026-07-29 morning: **3 leading to 2**

> **Ruling:** the distribution experiment first (owner does the posting; the
> share card and daily provocation are the travelling artefacts), feeding the
> arena thesis (categories, one daily provocation each, communal views only
> once participants exist). **Plus a named feature direction in the owner's
> own words: "a (dated) personal history of your opinions might be a goldmine"
> — a dated personal ledger of what you argued and when.** Full ruling +
> design constraint (localStorage ledger, entries only from real /defend
> responses, no accounts in v1) recorded in
> `PLAN_2026-07-22_gauntlet_dead_cta.md` § OWNER DIRECTION 2026-07-29.
> **The next session's first build is the opinion ledger.** The option text
> below is kept for the record of what was decided between.

The original framing, for the record:

The user comment that started it: *"why argue with an AI when Perplexity or
Google is free?"* The site is honest, reliable and now designed; nothing on it
can manufacture visitors. The choice is which product thesis vonc.com pursues
— they lead to different work and different spend:

- **A. The examination thesis.** The Gauntlet is a TEST, not a conversation:
  a clock, an adversary, a judge, a keepable verdict. Chat has no stakes;
  this does. Choosing A means building depth — difficulty tiers, a personal
  history of your own real rounds, harder judges — all buildable inside the
  no-fabrication rail.
- **B. The arena thesis (the owner's own 07-28 hypothesis: liveliness +
  breadth).** The differentiator is that it is a SHARED daily thing — one
  provocation per day per category (current affairs, celebrity, films, news,
  finance), everyone argues the same one, and once there are participants you
  can see how positions divided. Chat is solitary; this is communal. RAIL
  (owner's own, recorded 07-28): group-opinion graphs come AFTER participants
  exist — a chart on an empty site is a fabricated crowd.
- **C. The distribution experiment (cheapest, reversible, answers A-vs-B with
  data).** Treat H as an audience question, not a build question: take the
  daily provocation and the share card to where people already argue (the
  owner posting them; the card exists precisely to travel), watch whether
  anyone follows and plays, and let real behaviour choose between A and B.
  Costs no engineering beyond the og:image fix already owned elsewhere.
- **D. The demonstration ruling.** vonc.com is an agentchassis test property;
  its job — prove the platform can ship an honest interactive product — is
  done. Park the Gauntlet as a showcase, spend nowhere, revisit if traffic
  ever arrives organically.

They are not exclusive: C before A/B is the evidence-first ordering, and D is
what C collapses into if nobody follows. What only the owner can supply: which
thesis he WANTS this site to test, and whether he will do the distribution
himself (C requires a human with an audience; no agent here should be posting
to forums on his behalf uninvited).
