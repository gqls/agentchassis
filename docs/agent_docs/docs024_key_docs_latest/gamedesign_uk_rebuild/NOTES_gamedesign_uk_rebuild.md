# NOTES — gamedesign.uk rebuild

Append-only, newest at the bottom. Technical log: evidence, commands, what the system
actually said, and every misstep.

---

## 2026-09-02 — session opens: diagnose the live site

Owner: "look up previous threads for gamedesign.uk and fix the site, it is in a bad way."

Searched `docs024_key_docs_latest/`, `bugs_open/`, `bugs_closed/`, memory and git. **No
dedicated workstream directory existed** — every hit was an April–May 2026 pipeline doc
using gamedesign.uk as a test target. This directory is new.

**First DB check contradicted the premise as I'd assumed it.** `SELECT ... FROM sites
WHERE domain ILIKE '%gamedesign%'` → **0 rows**. Widened to `%game%` → one row,
`gamesdesign.co.uk`. Widened again to a whole-row text sweep,
`to_jsonb(sites)::text ILIKE '%gamedesign.uk%'` → still only gamesdesign.co.uk. So the
platform has no record of gamedesign.uk at all, yet the domain serves.

Live probe with the parked-domain control the memory index insists on:

```bash
curl -s -o /dev/null -w '%{http_code}' https://gamedesign.uk/this-path-does-not-exist-9z8x7.html
# 404 — not a catch-all, so its 200s are real pages
```

Crawled all nine linked pages. Six serve a literal `<main>\n\n</main>`.
`/privacy.html` + `/terms.html` 404 while linked from every footer. `/sitemap.xml` 404s.

### MISSTEP 1 — I called a dead CSS rule live damage

I read `--color-card-bg: #ffffff` against `--color-text: #e0e0e0` and computed 1.32:1,
and was about to report unreadable text on white cards.

**Wrong.** Extracting the classes actually present in the markup (regex over the body with
`<style>` stripped) returned 18 classes, **all header or footer**, and `card` is not among
them. The rule never instantiates.

**The check that caught it:** enumerate the classes the MARKUP uses before reasoning from
the stylesheet. **A CSS rule is not damage until the markup instantiates it.** Cost: none,
caught before it reached the owner — but only because I looked.

### MISSTEP 2 — I nearly reported six identical pages as duplicate content

`/`, `/about.html`, `/getting-started.html`, `/services.html`, `/tools.html` are all
~15.8 kB, which reads like one page served six times. md5 + `<title>` per page: all
distinct, all correctly titled. The similarity is that each is ~15.8 kB of inline `<style>`
wrapping an empty body. **Similar SIZE is not identical CONTENT** — hash before claiming.

### The one control that mattered for the empty-main claim

An empty `<main>` in served HTML could be client-side injection. Checked: the only
`<script>` is a 320-char mobile-menu toggle, and a fetch with a Chrome UA returns the same
`<main>\n\n</main>`. The claim survives its disconfirming test.

---

## 2026-09-02 — owner redirects: "the primary problem to fix first is why the adoption caused a broken site"

Dropped the rebuild and went after the mechanism.

**The sites repo is local** (`~/projects/sites`), and it is the deploy source, so the
damage has a git history. This is what made the diagnosis cheap — no cluster archaeology
needed.

Bisected `<main>` content length per commit on `gamedesign.uk/index.html`:

```
06b7b1251  2026-04-14  main_chars=5977
f9838491d  2026-04-16  main_chars=0     <- the break
```

`f9838491d` is titled **"Rerender: index.html"**, `6 insertions(+), 278 deletions(-)`. The
diff deletes the entire hero + features sections and rewrites the header nav in the same
commit. So the rerender **succeeded at the chrome and wrote an empty body**.

### The discriminating control — was this fleet-wide or gamedesign.uk alone?

Seven sites had commits on 2026-04-16. For each, compared `<main>` length before/after for
every `.html` touched that day:

```
ai-agent-orchestration.com   html_files_touched=25   emptied=0
finetuning.uk                html_files_touched=37   emptied=0
gaswholesalers.com           html_files_touched=26   emptied=0
leopardessconsulting.co.uk   html_files_touched=25   emptied=0
gamedesign.uk                html_files_touched=11   emptied=4
vonc.com                     html_files_touched=12   emptied=0
robot-hands.com              html_files_touched=14   emptied=0
```

**4 of 11 on gamedesign.uk; 0 of 139 across the other six.** The rerender path was working
fleet-wide that day. Whatever emptied these pages was specific to what was being done to
gamedesign.uk — which was the adoption.

Also worth separating: `tools.html`, `services.html`, `getting-started.html` read 0 chars
**both before and after** — those were never populated, a different failure (never-built)
from the four that were emptied (content-loss).

### The mechanism, in the adoption thread's own words

`docs024_key_docs_latest/old_design_and_styling/FOCUS_design_and_styling_adoption_HANDOFF_2026-04-16_v2(1).md`:

> **Pages:** 11 (clean — **wiped and recreated by latest adoption**)

and, in its own Problems list:

> ### P3 — Empty `<main>` Content
> **Root cause:** The `needs_content_page` work items haven't been processed yet…
> **Fix:** Let the `needs_content_page` work items complete.

So that thread SAW the empty pages, classified the state as temporary, and expected the
content cascade to fill them. It never ran to completion. The empty shells were **already
published** by then.

The adoption ran with gamedesign.uk as **both source and destination** (site
`15a6cb16-5a86-4541-a8e4-d7106239b6a4`) — it crawled the live site and wrote back into the
same site row, wiping the pages that were serving.

### Is the defect still live? NO — three guards, all landed AFTER the damage

Checked the code, then dated it, because a doc comment enforces nothing:

| guard | file | landed |
|---|---|---|
| refuse to deploy header + empty `<main>` + footer | `rerender_single_page_action.go:581-602` | **2026-05-12** (`d777cb4d2`) |
| sibling-layout rescue for "adoption convergence with no sections to carry" | `load_page_sections_from_spec_action.go` fallback 4 | **2026-06-08** (`856fc4a51`) |
| empty assembly with component rows now FAILS instead of reporting COMPLETED | `rerender_single_page_action.go:167-186`, `bugs_open/095` | **2026-07-27** (`6579e9ae1`) |

All three postdate 2026-04-16. The first is the one that would have prevented this: it is
real code (`if len(sections) == 0 { return "", assembly, nil }`, caller then returns
`skipped: true, html: ""`), not a comment.

⚠ **[UNVERIFIED AT THE ARTEFACT]** I have confirmed these guards are in HEAD and dated
them. I have **not** confirmed the running chassis binary carries them by reading a build
provenance stamp. Given they are 3–6 months old and the fleet rolls frequently this is
near-certain, but it is an inference, not a measurement — mark it as such if quoting.

### Why it was never repaired

The site row was later deleted outright. With no `sites` row and no `pages` rows, nothing
can dispatch a rerender at gamedesign.uk, so the empty artefacts have been frozen in the
bucket and serving since **2026-04-16 — 4.5 months**.

It is also invisible to the detector built for this class:
`scripts/audit-archived-still-serving.sh` (`bugs_closed/359`) enumerates
`pages.status='archived'` with a non-null `deployed_at`. gamedesign.uk has no `pages` rows
at all. **359 covers a retired PAGE; this is a whole SITE whose rows were deleted while its
artefacts kept serving** — the same gap one level up, outside that detector by
construction. Candidate for its own bug file; not yet filed (see README).

---

## 2026-09-02 — independent re-investigation on Fable (owner's request), and what it corrected

Owner: "please investigate these errors again using fable." Launched a fresh general-purpose
agent on `claude-fable-5` with the symptoms and evidence pointers, ordered to MEASURE before
reading my write-up, then grade it and hunt for what I missed. ~24 min, 87 tool calls, read-only.
Its scratch is under `scratchpad/fable/`.

**What held, reproduced independently:** the fabricated-URL control; the no-row state at three
widths plus a site-id text sweep; the `f9838491d` diff; **the 4/11 vs 0/139 discriminating
control, exactly**; both source and destination = gamedesign.uk (single-domain trigger,
archive `site_record`, 04-16 handoff); "wiped and recreated"; guards `d777cb4d2` (05-12) and
`6579e9ae1` (07-27) as real code; no serving-side enumerator anywhere; the sibling's 40 pages.

**What it refuted, and I re-measured before accepting — because a second report is still a
report:**

| my claim | what was true | caught by |
|---|---|---|
| "six of nine linked pages empty" | **five** empty + **two 404s**; whole dir **13 of 47** empty | walking the directory, tabulating by category |
| `tools.html`/`services.html`/`getting-started.html` "0 chars both before and after — never populated" | two of the three **did not exist** on 04-16 (born 04-18). My loop's `$c` was EMPTY for a file with no pre-cutoff commit, `git show :path` errored, python got empty stdin and printed **0**. **An absent file measured as an empty file.** | Fable's per-file walk; confirmed by re-reading my own script |
| homepage "5,977 chars" | metric-dependent: mine stripped tags but kept the inline `<style>` text INSIDE `<main>`; Fable's strips script/style → 2,473. Direction (→0) holds; the number needed its metric stated | comparing the two scripts |
| "three guards" incl. `load_page_sections_from_spec` fallback 4 | **two** publish guards + one **build-side rescue**; that file never refuses an empty list | reading its terminal branch |
| `rerender_single_page_action.go:581-602` | 652–673 for Fable, **680** at `11414e733` for my spot-check — the file moves daily | `grep -n` at HEAD; cite sha+string, never a bare line |
| "with it every handle any detector or repair path had" | **1,147 `site_work_items_archive` rows** still carry `site_id` + `domain` | Fable's archive query |
| `[UNVERIFIED]` running binary carries the guards | **VERIFIED**: stamp `a2732c72…` via `service_binary_capabilities`; all five commits ancestors; HEAD→stamp NO / stamp→HEAD YES as controls | Fable (the CLAUDE.md log grep is out of range at `--tail=20000`) |
| "20+ top-level domain dirs" | **36** match the Action's filter, **8** row-less — I guessed LOW. **Fable said 19, also wrong.** | `ls -d */` + the Action's own regex + `SELECT domain FROM sites` |

**What it found that I missed — the one that matters:**
`ai-agent-orchestration.com/roi-estimator.html` serves `<main>\n\n</main>` (control 404); row
`active`/`deployed`, **0 component rows**, **eight `page_rerender` items `complete` 08-26→09-02**.
Spot-checked myself: confirmed. So the guard I called "closed" fires, reports `skipped`,
`check_skipped` routes to `complete_skipped`, the item closes as complete, the empty artefact is
untouched. **"Closed" is true for NEW publishes only.** `bugs_closed/315`'s profile, live. Two
more from the same 04-18 fleet-wide born-empty wave still serve empty (`llm-cost-calculator.html`
archived-and-serving; `robot-hands.com/learning-center-article.html` url-shape mismatch). Other
lanes' sites — flagged to owner, not filed by me.

Also: `index.html-gamedesign.uk` is **untracked** (`??`, mtime 2026-05-20) and byte-identical to
the original hand-built homepage at `04dc4200b` — the best surviving copy of the original home
copy, if the rebuild wants to reference it. And **23 of 49 gamesdesign.co.uk page titles read
"… | GameDesign.uk"** — a brand leak on the SIBLING; that lane's, not mine.

All corrections applied in place in `bugs_open/432` (six `CORRECTED` blocks) and here.

---

## 2026-09-02, evening — fresh chassis build deployed; stamp re-read

Owner: "A fresh chassis build has been deployed." Re-read the running stamp rather than assume:

```
agent-chassis | ebf27c60377f | 231 pods | last_seen 2026-09-02 16:19:44   <- new
agent-chassis | a2732c7207da | 346 pods | last_seen 2026-09-02 16:19:30   <- old, still draining
```
Mid-roll: two stamps coexist. `ebf27c60377f` = "394: un-ignore render_truncation_acks.json …"
(2026-09-02). `git merge-base --is-ancestor`: `d777cb4d2` YES, `6579e9ae1` YES, `856fc4a51` YES;
controls HEAD→stamp NO, stamp→HEAD YES; old stamp `a2732c72` is an ancestor of the new (so it IS
a newer build, not a re-tag). **Nothing changes for this lane** — the guards were already live —
but the figure in the PLAN and HANDOFF now names the current stamp.

**Gotcha found on the way:** `service_binary_capabilities`' column is `service`, not
`service_name`; my first query errored. RUNBOOK §5 carries the working form.

Wrote the RUNBOOK (the standing five were four until now) and `HANDOFF_2026-09-02_continue_here.md`.
The lane is BLOCKED on owner decisions A1–A4 (HANDOFF §4); nothing dispatches until they land.

---

## 2026-09-02, ~16:45–17:10Z — owner rulings executed, build dispatched

Rulings (verbatim shape): email `gamedesign@contactforsales.com` · look different, ask theme
kits · clear old files · present the brief · 315: reopen + hand to that site's lane · 432: do
here · oxenunity.com should have a row · the rest adopted later with oversight · tell
gamesdesign.co.uk to stop using our name · carry on.

**Theme kits' answer, the load-bearing half:** kits are NOT live (`to_regclass('theme_kits')`
NULL, no binary carries `apply_theme_kit`) — seed values directly. Palette cascade rung 1 is
`mission.preferred_palette` (the classifier does not write `mission`); typography cascade is
REVERSED (`design_intent.typography.reference_values` first), so seed both. `colour_mood` must
AGREE with the values (boxingonline: prose said dark, values said light, resolved silently to
the values). **Owner ruling today: `reference_values` is NOT a pin** — RFC_059 proposed one and
was withdrawn. Seeded = what the site starts from. The colour-churn landmine I cited earlier
in this lane is superseded on that point.

**432 fix, ordering deliberately before the pre-seed** so the demand control fired on the real
case: `audit-rowless-serving-domains.sh` first run → gamedesign.uk NO-ROW-AND-SERVING, plus
nine others. Bucket has 55 prefixes; the sites repo 36 — `secured-loan-calculator.com`,
`tilingcalculators.com`, `wykefarm.co.uk` exist only in the bucket. **Enumerate the artefact,
never the proxy** — my own repo census (8) was an undercount for the same reason Fable's (19)
was: neither looked at the bucket.

**Seed applied 17:04:35Z**, JSON validated before apply. Verify block returned both rows and
four current pinned specs. Re-ran the reconciler on three prefixes: gamedesign.uk and
oxenunity.com → ROW_NO_PAGES, nanangmrk.com (untouched control) → still NO_ROW. Register
IMP-059's verify-later half 1 satisfied.

**Retraction.** `~/projects/sites` local master is 14,738 commits behind origin with ONE
unpushed commit from another session (apis.uk restore, 14:02) — not mine to push or discard.
Did the deletion in a detached `git worktree` at `origin/master`, pushed `HEAD:master`. First
push rejected (a git-adapter rerender landed in between — expected); re-created the commit on
the new tip and pushed: `40bd35f19`. 58 tracked files removed, `404.html` kept, 34,848 lines.
Verified at the artefact with a cache-bust: `/` `/tools.html` `/about.html` → 404 (2846 B fleet
page), `/404.html` 200, invented URL 404, `b2 ls` → 1 object. ⚠ `git worktree remove` left the
shell cwd in the sites repo and my next `git add` in "agentchassis" silently ran there —
"pathspec did not match". Absolute paths, always.

**Dispatch 17:07:55Z.** Last chassis pod start 15:53Z (the ~300 s rule holds). 082 needed
`bash ./…` — the script is not executable. Correlation `f07313f6-976c-4593-9e5e-44892008fb74`,
orchestration `2069ee9e-…`, LANDED COMPLETED. **The brief did NOT land in `mission`** — 082
sends it as `input_data.mission_brief`, `persist_mission` had nothing to write, so my
`mission.preferred_palette` row was never merged over at all. `mission_brief.text` (2,892 chars)
is what `domain-research-classifier` reads (`site_specs.specs.mission_brief.text`, verified in
its definition). Site id `8f17eb73-fc74-4718-8371-b3125bc4e414`. `needs_domain_research`
triaged. Budget hours.

**Brand leak:** no lane owned gamesdesign.co.uk when I looked; routed to positioning (GD1
holder); the owner then created a dedicated `gamesdesign.co.uk` session [783baf] and it now has
the instruction and the package. Positioning's recount: four current specs carry the string
(not six — something superseded two in between; re-count before acting). Class bug still owed
by me.

---

## 2026-09-02, ~17:20Z — `bugs_open/438` (theme kits) and what the cascade did in its first 6 minutes

**438, filed by theme kits from my `mission_brief` datum:** on the FRESH path `persist_mission`
reads `input_data.mission` (which 082 never sends) and its `error_step` `persist_mission_brief`
writes aspect **`mission_brief`** — so aspect `mission`, the palette cascade's rung 1, is never
populated. Measured by them: `mission_brief` on 22 sites, `mission` on 2 (one of them mine, an
hour old); `palette_source` across 31 resolved compositions = `design_intent_values` ×30,
`archetype_default` ×1, **`mission_hint` ×0** — the rung the code calls "most authoritative"
has never fired in production. My seed is safe BECAUSE of the bug.

**Their tripwire, narrowed by reading the merge:** `write_site_spec` always deep-merges
(`site_spec_actions.go:246`; only arrays overwrite wholesale, `:14`). A repointed
`persist_mission` would ADD `text` beside `preferred_palette`, not replace it. And my
submitter run COMPLETED 17:07:57Z, so the step cannot re-run without a re-submission. Told
them, so the 438 fixer does not "fix" the merge away. **Watch item, not a blocker.**

**The cascade, measured:** classifier COMPLETED 17:09:24 (~2 min after triage — no queue this
evening); `needs_vertical_research` claimed 17:13; `vertical-exemplar-researcher` crawling at
17:13:20. Site id `8f17eb73-fc74-4718-8371-b3125bc4e414`.

**The classifier superseded my `design_intent` — `pinned=t` did NOT protect it** (manual row
17:04:35 → `is_current=f`; classifier row 17:11:32 current). But the overwrite went the SAME
WAY: bg `#F5F0E8` (seed `#F4F1EA`), accent `#9B4E2A` (seed `#A6521F`), `dark_light: light`,
colour_mood "Warm off-white ground — the colour of uncoated paper…". The brief's referent
sentence steered the classifier's own write to within two hex steps of the seed. Divergence:
`heading_font` = Playfair Display (seed Merriweather) — and typography's rung 1 IS
`design_intent`, so Playfair will likely win. Serif either way; recorded, not fought.

**Palette rung 1 is intact** (`mission` row byte-identical to the seed: no `text`, palette
present, `created_by` mine). When composition resolves, `palette_source` on this site is the
**first live test of `mission_hint`**: `mission_hint` ⇒ rung 1 fired for the first time in prod;
`design_intent_values` ⇒ rung 2 won WITH a populated `mission` row present, which would sharpen
438 from "nothing populates it" to "rung 1 does not read what is there". Owed to theme kits
either way.

---

## 2026-09-02, ~17:35Z — 438 measured fleet-wide from this lane's three error rows

Theme kits took the "three error rows per fresh submission" datum and counted it over the
30-day `agent_error_log` retention: `persist_mission` 16 rows / 12 sites, `persist_roadmap`
16 / 12, `persist_roadmap_brief` 14 / 11 — and **`persist_mission_brief` itself 6 / 3**. So the
step that happened to carry MY brief (verified: `mission_brief.text`, 2,892 chars, and the
classifier read it) **fails on a quarter of the sites it runs on**; on those, the submitter
wrote neither `mission` nor `mission_brief` and the classifier ran with no brief at all.
`082` sends no roadmap key whatsoever (`grep -c roadmap` = 0), so both roadmap steps fail by
construction on every fresh submit — 30 of 52 error rows are from steps that cannot succeed.

Their own correction: the four `persist_*` steps' `error_step` fields form a **linear
continuation chain** (`persist_mission → persist_mission_brief → persist_roadmap →
persist_roadmap_brief → create_research_item`), not a designed recovery pair. Mission being
"rescued" by mission_brief is incidental ordering. For this lane: **verify the brief landed on
every fresh submission — `SELECT length(data->>'text') FROM site_specs WHERE aspect =
'mission_brief' AND is_current AND site_id = …` — never assume it from a COMPLETED submitter.**
(RUNBOOK §7 gets this line.)

---

## 2026-09-02, ~17:45Z — 315 accepted by its new owner; fix built same day

`AI page 3` (site_ai_agent_orchestration) accepted the reopen with every figure re-verified at
the artefact. Fix: `8eca969cb` + `8a0b927f5` (writeWorkItem refactor after council r1 objected
to a hand-rolled ON CONFLICT), Council-Submitted `2be8ec34`, **inert until a chassis roll**.
Both doors closed: consumer skip-branch and producer loop now file a deduped
`needs_content_page`. Criterion 1 driven to the framework's edge, not hardcoded: a hand-filed
`needs_content_page` parked honestly at `needs_human_review` (`sections=[]`, no spec), so they
filed `needs_content_planning` first — the page's SHAPE is the planner's/owner's call.
**Census corrected by the council: 14 rows, not my "one instance"** (`deployed_at IS NOT NULL`
is the honest census predicate; the liveness predicate gave 2). Prior art surfaced:
`componentless_pages` is built, enabled in ZERO discovery agents, and needs sections PRESENT —
roi-estimator sits in the gap between it and `check_sectionless_pages`. Recorded in 432 §3a.

**Owner decision surfaced by them, carried here:** `ai-agent-orchestration.com/llm-cost-calculator.html`
— archived, still serving empty (359's class). Retract, or un-archive and build. Not acted on.

---

## 2026-09-02, 17:38Z — composition resolved: `palette_source = mission_hint`, first time in production

Cascade timings (all `[MEASURED]`): submitter 17:07:55 → classifier 17:09:24 → vertical research
17:16:47 → strategy → briefing → site plan complete ~17:36 → **composition 17:38:00**. Thirty
minutes, no queue — the "budget hours" line in the HANDOFF was written for a busy fleet.

**Site plan:** 5 pages — `/index.html` (landing), `/articles/index.html` (section-index),
`/about.html`, `/contact.html`, `/blog/article.html` (blog-post). Fan-out: `needs_composition`,
`needs_design`, 5× `needs_page`, 1× `needs_rerender`, 4× `needs_imagery`. Classifier's
classification: `category: editorial` — *"not a tool platform … does not host tools (those
live on the sister domain)"*. The positioning constraints held at the plan level.

**Composition (`resolved_composition`, resolved_by site-design-planner):**
| axis | source | result |
|---|---|---|
| palette | **`mission_hint`** — rung 1, **fleet baseline was 30/1/0, now 30/1/1** | `palette-gamedesign-uk-a6a70287`: **byte-for-byte the seed** (#F4F1EA / #FFFFFF / #33302B / #6E6558 / #A6521F / #23211E / #6B655C / #DDD6C9) — NOT the classifier's #F5F0E8/#9B4E2A at rung 2 |
| typography | `fingerprint_font_family_match` | heading Playfair Display (classifier's design_intent = rung 1 of that cascade, as predicted); body **Libre Baskerville** — serif throughout |
| layout | `library_match` from 5 candidates | **`magazine-grid`** (editorial, `scheme` empty) — NOT theme kits' `soft-editorial`; my `layout_preference` was in the superseded design_intent row, so the tag-match had only prose |

**438's gate: diagnosis HOLDS.** Nothing populates `mission` on the fresh path; when something
does, `extractPaletteSignal` reads it. Sent to theme kits with the values.

**Watch for the served stylesheet:** the owner's "reference_values is NOT a pin" ruling means
the render overlay may move off these; read `--color-*` in the served CSS after deploy and
report the values, not a complaint.

---

## 2026-09-02, ~17:50Z — first page live, unstyled; `article` parked honestly

**`/about.html` deployed 17:42:47** — 200, 10,455 B, `content_hash` set, 3 sections / 3 component
rows, **5,335 chars of main text** (script/style stripped). Opening: *"Writing about game design
for people who already do it. This site is aimed at the parts of the job that a design canon does
not cover well: what happens when a spec meets an engineering team, when a balance pass runs into
sign-off politics…"* — the brief, in the site's voice. `index` (4 sections), `contact` (3),
`articles-index` (3) building.

**`article` → `needs_human_review`** at 17:43:38: `page-build-handler no-op: no sections ready to
build` via `mark_no_ready_sections`. The planner created `/blog/article.html` (blog-post, nav
label "Article | gamedesign.uk") with **0 sections** — a slot with nothing to fill. The handler
parking rather than rendering an empty page is the post-April guard working; the same shape AI
page 3 hit on roi-estimator. **Owner/planner call:** leave it parked until there is an article,
or cancel the slot. Not an error.

**⚠ Page-before-stylesheet.** `/assets/css/styles.css` **404** while `about.html` serves: the
`needs_design` item (webdesign-agent, priority 8) is still `triaged`, and `css_themes` row
`274632b8…` has no `css_content` yet. So the first public page is bare HTML until design lands and
the priority-99 `needs_rerender` restyles. Normal cascade ordering, but a real (transient)
exposure — a fresh site's first minutes are unstyled by construction. Watch that `needs_design`
is CLAIMED; if it sits triaged for long, that is the throughput lane's starvation shape
(`RUNBOOK s15` in their docs: a triaged build item can starve behind older backlogs).

Benign warn row at 17:42:39 (`page-rerender`/`render_page`): "Outbound link suppression SKIPPED —
refused-target set unavailable or site has not shipped (bugs_open/328)". Expected pre-ship.

---

## 2026-09-02, ~18:05Z — SITE LIVE. Verification at the artefact, and the palette reading

**Served, cache-busted, control 404** `[MEASURED 2026-09-02 ~18:00Z]`:

| path | code | bytes | main text |
|---|---|---|---|
| `/` | 200 | 59,023 | 2,118 — title "gamedesign.uk — The Practice of Game Design" |
| `/about.html` | 200 | 10,455 | 5,335 |
| `/contact.html` | 200 | 7,733 | 1,984 (first probe 000, second 200 — 359's retry rule; transient) |
| `/articles/index.html` | 200 | 8,396 | 2,148 |
| `/blog/article.html` | 404 | — | parked slot, **linked from nowhere** — not a dead link |
| `/privacy.html`, `/terms.html` | 404 | — | **not in the plan and not linked** — the old footer linked them, the new one does not |
| `/sitemap.xml` | 200 | 507 | the four live pages, article correctly absent |
| `/assets/css/styles.css` | 200 | 19,977 | LAYOUT: magazine-grid |
| `/assets/images/favicon.png` | **404** | — | the one dead internal reference; logo.png 200 |

Constraint sweep on the four live pages: `href="mailto:"` empty **0** (footer now
`mailto:gamedesign@contactforsales.com`); "game room" 0; "GameDesign.uk Pro" 0; "calculator" 0;
links to `gamesdesign.co.uk` on `/` and `/about.html` — the sibling cross-link positioning asked
for. Homepage opening: *"Game design, examined as a practice, not a pitch. Writing for people who
already run the reviews, own the sign-off and have opinions about the pipeline, on the parts of
the job that are more judgement than arithmetic."*

**THE PALETTE READING — the "not a pin" ruling, measured on a hand-seeded composition:**

| slot | seed (= composed palette row, `mission_hint`) | classifier `design_intent` (rung 2) | **served `styles.css`** |
|---|---|---|---|
| background | #F4F1EA | #F5F0E8 | **#F5F0E8** ← classifier's |
| accent | #A6521F | #9B4E2A | **#9B4E2A** ← classifier's |
| surface | #FFFFFF | — | #EDE7DB |
| primary | #33302B | — | #2C1F14 |
| secondary | #6E6558 | — | #5C4033 |
| text | #23211E | — | #1E1410 |
| text-muted | #6B655C | — | #6B5A4E |
| border | #DDD6C9 | — | #D4C9BA |

**The composed palette row and the served stylesheet disagree on all eight slots.** Where the
classifier wrote a value, the CSS carries the classifier's; the rest are re-derived warmer
browns. So `palette_source=mission_hint` was true at composition and **did not reach the
stylesheet** — the render overlay (webdesign-agent) took its values from `design_intent`, not
from `resolved_composition.palette_id`. Under the owner's 2026-09-02 ruling this is permitted
and expected ("full authority to ignore our set of themes"). The DIRECTION held completely —
warm paper, earth accent, serif throughout (Playfair / Libre Baskerville), light — which is
what the owner asked for. **Recorded as values, not a complaint**, per theme kits' request.
Whether `resolved_composition.palette_id` should describe the served site is theme kits' /
site_design_planner's seam question; noted to them, not filed by me.

**Open items on the site:** `article` parked (0 sections — owner/planner: leave or cancel);
3× `unresolved_cta` (gated, no dead links, "no eligible content hub" — self-resolves when
articles exist); `site_unreachable:detected` (stale — `/` is 200; clears on next rotation);
4× post-design `page_rerender` in flight (the restyle); 2× `needs_page:triaged` (post-design
re-files). Missing favicon: `/assets/images/favicon.png` 404.

**Cascade wall-clock: dispatch 17:07:55 → homepage live ~17:56 → styled by ~18:00. Under an
hour.** The HANDOFF's "budget hours" was a busy-fleet figure.

**432 reconciler, re-run:** gamedesign.uk must now be OK (row + 5 pages) — verify below.

---

## 2026-09-02, ~18:30Z — the palette mechanism, and a correction to my own landmine

Theme kits verified the served table byte-for-byte and read the mechanism: `render_css_from_spec`
→ `buildPaletteMap(comp.Palette, specPalette)`, **8 core slots are spec-wins by design
(DES-003/DES-042)**; `analyze_design` reads `design_intent`, never the composed palette row. So
`mission_hint` won the composition and lost the render, by construction. **No submission-side
seed can put a chosen core colour on a site**; the lever on served colour is the BRIEF, which the
classifier turns into `design_intent`, which is what the overlay reads. 438 §6a-ter records it as
the ruling working, not a defect — and warns future readers not to file it as one.

**Corrected my own LANDMINES entry (written ~17:45, wrong at ~18:30) in place**, and logged the
wrong call: I wrote a prospective remedy from the composition record without first reading the
stylesheet. The SEED file's header comments (§2 "your reliable lever") are now historically
accurate about composition and wrong about the artefact — left as written with this NOTE as the
correction, since the file is applied history, not guidance. Theme kits' own advice to me was
wrong the same way and they said so first.

---

## 2026-09-02, ~18:45Z — sibling renamed; one inbound deep link I broke without telling them

**gamesdesign.co.uk session:** rename to **"GamesDesign.co.uk"** executed (owner confirmed
positioning's recommendation) — 4 specs superseded retire-then-insert with `adopted_from` preserved
by an in-transaction guard, 22 plan titles, 23 page titles + 1 meta, 30 page_components, backups
`bak_gdcouk_rename_20260902_*`, 32 rerenders dispatching. 439 stands as filed; they will not claim a
fleet-wide zero from their manually-renamed rows.

**A miss of mine, surfaced by them:** their `guide-p2p-architecture` page carried a live absolute
link `https://gamedesign.uk/games/p2p-networking/index.html` ("Launch P2P Simulator"), inherited at
the June adoption. My 17:05 retraction (`40bd35f19`) removed that path; it has been **404 since**,
and I had not looked for inbound links before clearing. Censused now: `page_components.rendered_html`
across their site → **exactly one** absolute link into `gamedesign.uk/` fleet-wide, this one. Their
own copy of the page is 200 at the same path on their domain; told them to repoint. No redirect
from my side — the editorial seat does not host games (positioning), and a redirect would put the
tool kind back on the wrong domain.

**The check I skipped, for the RUNBOOK:** before retracting a domain's tree, census inbound links
from every OTHER site we control —
`SELECT s.domain, p.name FROM page_components pc JOIN pages p ON p.id=pc.page_id JOIN sites s ON s.id=p.site_id WHERE pc.rendered_html LIKE '%<domain>/%' AND s.domain <> '<domain>';`
— and tell the owners before, not after. One query; I ran it 100 minutes late.

---

## 2026-09-02, ~19:10Z — sibling loop closed, verified at their artefact

gamesdesign.co.uk lane reports: 'GameDesign.uk' on zero of 31 serving pages (case-sensitive,
cache-busted); the P2P button repointed to their local `/games/p2p-networking/index.html`; one
rerender raced their link-fix commit by seconds and briefly re-served the dead URL, cleared by a
second dispatch after ~19:00Z. **Spot-checked myself** `[MEASURED 2026-09-02 ~19:10Z]`: `/`,
`/about/`, `/guides/p2p-architecture/`, `/tools/` — `GameDesign.uk` **0** on all four,
`GamesDesign.co.uk` 30/5/3/15; the button's href is `/games/p2p-networking/index.html`. No
further dependency from their tree on the old one. **Owner instruction "tell gamesdesign.co.uk to
stop using our name": DONE and verified at both artefacts.**

---

## 2026-09-02, ~19:25Z — owner rulings on the four open items, executed

Owner: no privacy/terms pages · yes favicon · cancel the article slot · llm-cost-calculator "not on
this site unless the story is about using LLMs for games design and playing games".

- **Privacy/terms:** nothing to do. Not planned, not linked, ruled unnecessary.
- **Article slot:** `SEED_2026-09-02b_…` applied — `needs_page` → `cancelled` (reason in result),
  page `article` → `status='archived'` (never deployed, in nav nowhere; 266's guard now refuses
  any deploy). `site_plan_pages` row deliberately left — the plan is written whole and this lane
  will not hand-edit a versioned plan. **WATCH:** the next rotation must NOT re-file work at the
  archived page; if it does, that is `bugs_open/356`'s class (fixed in tree 08-22; whether the
  roll carried it is a stamp check).
- **Favicon:** `needs_brand_head_assets:favicon` filed by hand in the `undeployed_assets` check's
  exact shape (handler `asset-deployer`, pipeline build, priority 60, dedup key so the check will
  not twin it). `derive_brand_head_assets` resizes `logo.png` (200) — deterministic. Fleet record
  for the type: 22 complete / 44 unresolved / 10 failed — not a sure thing; monitor armed.
- **llm-cost-calculator.html:** it is on **ai-agent-orchestration.com**, not this site — I had
  listed it without saying so loudly enough. Relayed the owner's PRINCIPLE to AI page 3 with that
  caveat; on its face the principle points at un-archive-and-build there. Their call / owner's.

**~19:40Z correction (AI page 3, measured at the artefact):** my "the principle points at
un-archive-and-build" reading of the llm-cost-calculator question rested on the tool NOT existing.
It does — `/tools/tool-llm-cost-calculator.html`, built 2026-09-02, 67 KB of real content. The
archived flat-URL page is a stale empty duplicate still linked from nav; a third empty shell sits
at the guide URL. Real action: repoint two nav links, retract two shells — that site's owner's
call, recorded in 315. **The cheap check I skipped: before recommending "build X", ask whether X
already exists at another URL.** One `SELECT url FROM pages WHERE site_id=… AND url ILIKE '%llm-cost%'`.
Corrected in 432 §3a; nothing routes here.

---

## 2026-09-02, ~19:55Z — favicon live; the owner's list is empty

`needs_brand_head_assets:favicon` claimed → complete within ~25 min. **At the artefact**
`[MEASURED ~19:55Z, cache-busted]`: `/assets/images/favicon.png` 200, 8,640 B, `image/png`, PNG
signature valid, **64×64**; `/assets/images/og-card.png` 200, 159,864 B; `logo.png` 200. Homepage
`<head>`: `<link rel="icon" href="/assets/images/favicon.png">` + `apple-touch-icon`. The
monitor's own final probe printed `000` — a curl transport failure inside the script (its
`|| echo 000` concatenated with `-w`), the same 000-then-200 shape `contact.html` showed at 18:00.
**A status is not an artefact, and a monitor's probe is not either: I read the file.**

Every open item from the 18:10 HANDOFF block is now resolved or ruled: (1) article slot
CANCELLED + page archived; (2) legal pages ruled unnecessary; (3) favicon LIVE; (4)
llm-cost-calculator is ai-agent-orchestration's, and moot (tool exists) — theirs; (5) 432 stays
OPEN pending scheduling of IMP-059 (keyboard-run today) — a platform item, not this site's.

**Still watching, not blocking:** the next discovery/planner rotation must not re-file work at
the archived `article` page (356's class). One query, tomorrow:
`SELECT item_type,status,created_at FROM site_work_items WHERE page_id='2ea5d983-b798-4bb2-b30a-5e3047369561' AND created_at > '2026-09-02 19:20';` — expect 0 rows.

---

## 2026-09-02, ~20:30–21:30Z — the owner reviewed the site and it was WRONG; the lane reopened

Owner, verbatim: *"this site needs to be seen again by the checkers. please run the improvement
loop over it. it suffers from the same problems that designblog.co.uk etc suffered with. please
correspond with that blog to determine the best fixes. We need to change the design and copy. hero
images are missing e.g. articles/index.html that same page shows an explanation of the brief and so
on. It is a game design site why isn't it full of games and images and excitement — please add that
to the errors list it is a major error."*

**The 20:00 "closeable" SUMMARY was wrong at the ARTEFACT** — I verified presence (pages, links,
sitemap, favicon) and never read the site as its reader would. Every presence check passed on a
brief-shaped void. The root causes are mine and are stated in `bugs_open/446` §3.1: the
`imagery_style_guide` I seeded BANNED game imagery (the hero prompts say "no game imagery"
verbatim; the pictures are pencils and paper), and the brief asked for "restrained practice
journal". Briefs govern; the pipeline obeyed.

**Measured (cache-busted, control 404):** 1 `<img>` per page (logo); `/articles/index.html` = 0
articles + brief-echo prose incl. a "What they avoid" list (owner-banned negative-identity copy);
its hero = `url('/assets/images/hero.jpg')` → **404** (planner requested no site-scope hero; every
other site probed HAS one — controlled, not a fleet claim); home/about/contact heroes exist as CSS
backgrounds under a 50–70% black wash — **and about + contact both render `hero-home.jpg`**;
`hero-about.jpg` / `hero-contact.jpg` referenced on 0 pages. My first census said otherwise
because `grep -l hero-about` matched `data-component="hero-about"`, not the filename — anchor on
filename+extension with a control (WRONG_CALLS-worthy; SITE_DEFECT_CATEGORIES 10.3b carries it).

**Improvement loop run (corr `8b2473ab`, 20:00Z):** the auditors filed **27 record-mode verdicts**
that say everything the owner said — zero articles, index writes about itself, empty featured
section, hardcoded hero overlay, CTA `#8b0000` clashing with the accent, "second heading defines
what the site is not", no author, and `gamedesign@contactforsales.com` reads as a placeholder to
this audience. **`[verdict, not dispatched]`, all 27.** The checkers see it; record mode acts on
nothing. 446 §4a.

**Cross-lane (owner: "correspond with that blog"):** designblog's CRITIQUE is near word-for-word
ours; their routing joined rather than duplicated: 444 (brief-echo; my hub is a fourth mechanism —
no article pages exist; their resolver DOES hold `section-index`, corrected after I said otherwise),
114 (imagery; inline guide imager's predicate-derived population: **7 components, 157 instances,
61 of 65 page heroes orphaned, hero-tool 76**, one counter-instance that may be the cheaper fix),
components (mechanism: site-wide `hero_url` injection, aliasing gated on an image-TYPED field —
**they own the migration**; I verify at my artefact), theme kits / site design planner / copy
quality two stage (routed by designblog; not re-opened by me). **Migration 718 landed 19:59Z**: the
planner prompt now EXPECTS content imagery — my re-plan inherits it.

**Filed:** `bugs_open/446` (the owner's "major error", with my root cause stated), CONTRIB to 444,
`SITE_DEFECT_CATEGORIES.md` §10 (the owner-directed errors list — "add that to the errors list"
means THAT file): 10.1 the spec bans what the vertical is made of · 10.2 index with zero members ·
10.3 hero over a CSS-url 404 · 10.3b the WRONG hero passes presence checks · 10.4 zero interactive
elements on an interactive vertical · 10.5 no detector; a reviewer with a referent.

**Re-seeded (`SEED_2026-09-02c`):** imagery_style_guide v2 (game art, big, bold, stylised where the
game is; stationery banned instead), evidence_base v2 (name released games, describe observable
play; ban sales/scores/internal-decision claims and brief-echo headings as SHAPES), design_intent
v2 (games magazine, not print journal). Brief v2 (`MISSION_2026-09-02b`, 3,590 chars): "full of
games, images and energy… name the games… playable illustrations embedded… never explain its own
brief". **Re-dispatched 20:10:59Z via 082, corr `aab87c0c-f731-4a64-a20f-0e81fc5c8375`**; brief
v2 verified in `mission_brief.text`; classifier queued. Monitor armed.

**Owner decisions surfaced by the auditors, not mine to make:** (1) the contact email domain —
`contactforsales.com` reads as a placeholder to senior studio professionals (×3 verdicts); (2) an
author / editorial identity — the evidence rules forbid inventing one, so the site is anonymous
until an identity is supplied; (3) newsletter/RSS — no repeat-visit mechanism (feed lane's
mechanism exists; undriven per site).

---

## 2026-09-02, ~22:30Z — cluster token expired mid-rebuild; watching at the artefact instead

`kubectl` → `Unauthorized` (the 3-day kubeconfig expiry; owner refreshes). The DB monitor's last
reading ("plan imagery rows:" empty) was this, not a plan change. Stopped it. Re-armed a curl-only
monitor on the served site — sitemap URLs, the four hero `url()`s (filename-anchored), the articles
page's text length and whether "What they avoid" is present — which is the verification that
matters anyway. Last DB-side state before the token died (~22:20Z): cascade #2 through classifier
(20:16) and vertical research (claimed 20:17); classifier's v2 `design_intent` = `bold-creative`,
game imagery, gold accent `#D4A017`, "sensibility of a magazine". Plan not yet written.

**Also done ~22:10–22:25Z:** `growth_posture='hold'` set (SEED e; WDS-020 live in stamp ebf27c60
by ancestry check; first of 39 sites; unexercised here until the next loop run); 447 candidate 3
STRUCK on the loop owner's refutation; 444 CONTRIB completed with page_type + sections + the
second mechanism (the planned article had `parent_section` EMPTY at `/blog/`).

**When the token is back, first three reads:** the new `site_plans` row (pages, roles,
`parent_section`, imagery rows — against 444's predicate and 446 §3.3's "site-scope hero"), any
`add_tool` filed in RECORD shape (WDS-020's demand test), and the served hero `url()`s after the
rerender for `components`' migration 721 (their first live test).

---

## 2026-09-03, ~09:30Z — fresh build read; handoff #2 written

Stamp `7bf1ff674021` (09-03; roll in progress 91/322 at 09:16Z). Ancestors, HEAD→stamp NO:
`6525b45ae` (444 gate), `c2349955d` (WDS-020), `d777cb4d2`, `8eca969cb` (315). Migration 720
(444's gate needs it) NOT verified — `schema_migrations` has no `version`/`name` column by those
names; check `\d schema_migrations`. The `needs_briefing` I filed at 08:31:19Z was still `triaged`
at 09:16Z: no lock, nothing claimed, the dispatch loop cycling gamesdesign.co.uk / fundamentallyai /
gaswholesalers every ~4 min — `dispatch_throughput/RUNBOOK` (bug 413: a pinned old row freezes a
site's age; younger sites starve; "hours" is normal). Not hand-spawned; recorded. New cold-start
doc: `HANDOFF_2026-09-03_continue_here.md`; third SUMMARY. Artefact monitor timed out 09:16Z;
re-arm in the next session.

## 2026-09-03, ~10:40–10:55Z — rebuild #2 PLANNED, and it has ZERO articles again. Cause found: a planner refusal

**The chain moved while the last session was writing the handoff.** `needs_briefing`
`95d834f8` claimed 09:33:16Z, complete 09:34:35Z; `needs_site_plan` `173744d7` claimed
10:38:06Z, complete 10:40:40Z. So the "still `triaged` at 09:16Z, starving" state in the
handoff's §0 table was ~1 hour from resolving itself. **Nothing was hand-spawned.**

**The new plan `c920da7a` (10:40:21Z) has FOUR pages and ZERO article pages** — `index`
(landing), `articles-index` (section-index), `about`, `contact`. Identical page set to rebuild
#1. §6's first check therefore FAILS: "expect N article-role pages parented under the articles
section" → N = 0.

**This was NOT the gate dropping them.** I read the planner's raw output
(`llm_call_log` `7b3bffdd-64dc-4a97-bb00-7633aa7271f8`, 25,400 in / 4,072 out against
`max_tokens` 16,000 — **not truncated**, so the CLAUDE.md `output_tokens == max_tokens` trap does
not apply): the planner PROPOSED exactly those four pages. The gate's `drops` list was empty.

**Cause, in the planner's own `strategy_notes`:**

> "The blog-post type is satisfied by the **blog infrastructure**; individual posts are not
> planned as static pages here."

There is no blog infrastructure. `[MEASURED 10:48Z]` active `blog-post` pages with non-empty
`sections`: webdesign.co.uk 52, dartsonline.com 23, finetuning.uk 22, seotools.co.uk 14,
**gamesdesign.co.uk 13** (our own sibling) — all ordinary planned pages built by the normal
pipeline. `[MEASURED 10:47Z]` the same reasoning appears in **3 of 32** `plan_site` runs in 30
days, and the other two are **designblog.co.uk** (09-02 16:10Z) and seotools.co.uk (09-02
16:13Z). Full evidence + the falsification framing: CONTRIB appended to `bugs_open/444` today.

**MISTAKE THIS LANE MADE, corrected here.** My 09-02 CONTRIB into 444 called this "the plan
created ONE article page with ZERO sections… so the type has no producer because no content pages
exist at all" and read it as a parenting accident. **It is a refusal, and it is not site-specific.**
The 09-02 reading would have had us fix parenting and re-run — which is exactly what we did, and
it reproduced the defect. What caught it: reading the planner's own reasoning instead of only its
output. **Cheap check that would have caught it on 09-02: `SELECT response_text` for the plan_site
call, not just the resulting `site_plan_pages` rows.** The rows tell you WHAT was planned; only
the response tells you WHY the rest was not.

**And the mission could not have been clearer.** Mission v3 (seeded 09:45:50Z by the previous
session, 55 min before the planner ran) says "The site launches with real articles, not a
description of what the articles will be like. A page that lists articles must list articles". I
confirmed it reached the model by reading the RENDERED prompt (line 110), not the seed file.
**The planner received that instruction and overrode it.** So there is no per-site lever here:
`site_plan_directives` is not one either — all 1,922 rows are written BY the planner
(`write_site_plan`) and "directive" appears 0 times in the rendered prompt. Input never, output
always.

**720 resolved (the handoff's open ⚠).** APPLIED and LIVE — `enforce_listing_sources: true` on
`validate_plan`, and rule 3's narrowed text at position 25019 of `plan_site.prompt_template`. But
**absent from `schema_migrations`** (which holds 721/723/724/726/727/728 and no 720). Told 444.

**The gate's first live rebuild run was CORRECT**: 2 `capability_gap` rows,
`gap_kind=producer_missing` — `index`→`blog_posts`, `articles-index`→`section_children:articles-index`.
Neither page dropped, rightly: both are realised, so the 001 preserve guard kept them.

**Decision: let the chain run.** The remaining items (`needs_page` index, 5× `needs_imagery`,
`needs_rerender`) were still `triaged`/unclaimed at 10:45Z. The realised `index` ALREADY carries
`["hero","featured-content","content-listing","generic-text-block"]` — the article-grid shape is
live from rebuild #1, so finishing the build adds no new defect and does deliver the game imagery,
the 721 hero test and the palette. It cannot add articles. **Do not report rebuild #2 as fixing
the owner's articles complaint — it does not.**

## 2026-09-03, ~11:00Z — MISROUTE corrected, and the finding got sharper for it

**I filed the planner-refusal finding at the wrong lane first.** Messaged `site design planner`
as the owner of the planner; that session replied that it owns **composition resolution only**
(`resolve_composition_layout_action.go` + siblings — layout/typography/palette), while what I
found is `build-site-planner`/`plan_site` page planning. **Different agent, no code overlap; the
names collide.** They also said they had confirmed the same split with the `bugs_open/427` lane on
day one for this exact reason. Correct route: **`bugs_open/428`**, actively owned.

**Cheap check that would have caught it:** before messaging a lane named after a mechanism, grep
the agent `type` string it actually owns (`build-site-planner` vs `site-design-planner`) rather
than matching on the English words. The two names differ by one hyphenated token.

**The reroute made the finding better, which is the useful part.** 428 already shipped migration
**687**: the planner must now name each omitted `recommended_page_types` entry in `strategy_notes`
with a per-type reason. I confirmed 687's rule reached our 10:40:15Z call by grepping the RENDERED
prompt ("omitted named type with no per-type reason in strategy_notes is a gap, not a decision").
**So our call is the first POST-687 instance, and 687 WORKED** — the planner named `blog-post` and
gave a reason. **The reason is false** ("satisfied by the blog infrastructure"), and it asserts
"All four types are present" in the same paragraph as explaining the absence of one. 428's §3
sample is entirely pre-687 (2026-05-14 → 2026-08-31), so nobody had audited 687's output yet.
Residual for that lane: the "note why" obligation is satisfiable with a hallucinated justification
and nothing checks it against the estate — while 444's gate computed the contradicting fact three
seconds later in the same validation pass. CONTRIB appended to 428; pointer left in 444.
