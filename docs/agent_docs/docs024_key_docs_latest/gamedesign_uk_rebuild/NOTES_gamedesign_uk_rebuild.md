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
