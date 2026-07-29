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
2. **Witness 131-B on the acceptance lane.** The clause is live and did real
   work in the C build harness, but no `improve_tool` tier-4 run has exercised
   it in production. Either wait for a natural run or fire one acceptance run
   at any tool page with criteria including `no_horizontal_overflow` on
   mobile. The first genuine catch (or clean sweep) closes 131-B fully.
3. **og:image** — vonc's gauntlet page emits a 404 social image (owned by the
   OTHER bug numbered 131, `…og_image_points_at_a_card…`, UNDIAGNOSED). When
   fixed, shared verdict cards and links get a face. Do not build a generator
   here; contribute into that file if you touch it.
4. **Consolidation ping**: features_open/024 wants `platform/mailer` +
   `platform/httpguard` adopted into tools-api and its memory line says
   "message them first" — this workstream owns tools-api, so expect (or
   answer) that contact. Nothing owed yet.
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

## 7. H — the decision the owner is being asked to make

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
