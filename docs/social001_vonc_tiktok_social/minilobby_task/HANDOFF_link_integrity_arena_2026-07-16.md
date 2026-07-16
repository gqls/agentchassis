# HANDOFF & SUMMARY — vonc link integrity, the Arena, and where we go next
*2026-07-16. Read top-to-bottom. This is the current bootstrap for a fresh chat.*

---

## The one-paragraph version (read this out)

We set out to fix vonc.com's broken links and "phantom" tools — buttons that said
"Enter the Gauntlet" but went to the contact page, and copy that sold an "Arena"
that didn't exist. Rather than patch vonc by hand, we fixed the **root cause** (a
platform-wide bug where a button's web address was locked to a fixed page
regardless of what the button said) and taught the platform to **find and repair
this class of problem on any site automatically**. We proved it works: before we
touched vonc, the automatic checker independently found every broken link we knew
about. We then built the **Arena** as a real, interactive page and wired every
"Arena" button to it. Everything is live and verified on the real site. The
automatic checker is now doing its job so well that it has surfaced a further
batch of the same kind of issue on other page components — which is exactly what
we want — and closing those is the next step.

---

## Where we are (all live and verified on vonc.com)

**Root-cause fix, fleet-wide.** The bug: a component's call-to-action URL was
sourced from a fixed page slug (`pages.contact` / `pages.services`) and re-applied
on every render, so no content edit could ever win — 143 of ~144 hero/CTA buttons
across 8 of 10 sites pointed at `/contact.html` no matter what the copy promised.
Migration **091** removed that lock; the link resolver now chooses a real,
relevant destination (tools/games first, then content hubs).

**Automatic detection, fleet-wide.** Migration **092** switched on three checks
for every site: `phantom_internal_links` (links to pages that don't exist),
`misdirected_cta` (button copy names one page, the link goes to another), and
`incomplete_page_group` (a promised set of pages where some were never built).
We **proved the loop**: a discovery run on vonc *before* any repair independently
found the 19 misdirected buttons, the 2 dead links, and the Arena buttons.

**vonc repaired.** All 9 archetype pages + the hub now send their primary button
to the Gauntlet and secondary to the Quiz; the two dead `/how-it-works` links are
fixed (**093**); every "Arena" button reaches the Arena.

**The Arena is real.** `https://vonc.com/tools/arena/index.html` — a genuine
client-side tool (daily provocation, file-your-take with local save, the five
Arena Reactions — Genius/Delusional/Suspicious/Based/Cursed — and a remix-chain
visual), built by the platform's tool generator with zero bespoke code, live and
in the footer nav next to Gauntlet and Quiz.

**Residual cleanups shipped.** Gauntlet-CTA's off-brand static labels unlocked
for authoring (**096**); the provocations "Enter today's Arena" button pointed at
the Arena (**097/097b**); a false-positive guard added to the phantom-link checker
so it stops flagging deliberately-empty template shells (Go, committed — rides the
next chassis image).

**Also caught and fixed en route:** a real field-name bug in migration 095 that
silently left Catalyst's button on the wrong page (fixed in **095b**); the tool
generator ignoring the planned URL slug (page reconciled to the right address);
and a stale navigation link left over from that rename (nav rebuilt, clean).

Everything above was verified **by artifact** — actual database rows and live
`curl` of the deployed site — never by trusting a job's "complete" status.

---

## Where we're going (next steps, in priority order)

### 1. Broaden `applyCTARecompute`'s scope — *the top job*

`applyCTARecompute` is the Go function (in `rerender_page_sections_action.go`)
that, during a "cta_links_stale" page re-render, actually **repairs** a
misdirected button: it recomputes the button's URL from the site's real pages
(interactive tools first, then content hubs) and writes it back. Today it only
runs for two generic component types — `hero` and `call-to-action` (the set named
in `ctaFieldNames`). But the `misdirected_cta` *checker* scans **every** button on
every component, so it correctly detects mismatches on other button-bearing
components too — `archetype-grid`, `archetype-combinations`, `gauntlet-cta`,
`content-block-about` — which the repair function then **can't touch**. In other
words, our detection is broader than our repair, so the loop finds these but can't
close them. "Broadening the scope" means teaching `applyCTARecompute` about those
additional component types and their differently-named URL fields (each component
names its fields its own way — `cta_url`, `primary_cta_url`, `cta_primary_url`,
etc.), plus giving their URL fields the same source-unlock treatment 091 applied
to hero/CTA, so that a single re-render repairs them the same automatic way —
turning today's "human review" items back into self-healing ones, fleet-wide.

### 2. Investigate the gauntlet/quiz `needs_rebuild` flag
Both tool pages got marked `build_status='needs_rebuild'` during the close-the-loop
run (they still serve fine). Determine whether that flag is spurious or real, then
clear it or rebuild the two tools via the tool generator (never the generic page
builder — TP-004).

### 3. Content-writer pass on the unlocked components
096 *unlocked* gauntlet-cta's labels but deliberately did not hand-author stat
numbers (anti-fabrication rule). A proper content pass should author real Spark
copy for gauntlet-cta and the other non-hero components once #1 makes their links
self-heal.

### 4. Loose ends
The quiz's "Get Your Full Report" button has an empty link — decide if a
"full report" feature is intended or the button should go. The auto-created Arena
guide page isn't in a blog listing. Both minor.

---

## Bootstrap for a new chat

- **Live site**: vonc.com, site_id `9ec3b9ee-5b08-461b-b4f8-9e1e03579c74`.
- **DB**: `kubectl exec -n ai-persona-system postgres-clients-0 -- psql -U clients_user -d clients_db`.
- **Deployed chassis**: `v1.0.1123` (WS1 Go verified live at v1.0.1117; the
  phantom-link runtime-fill guard is committed at `9752bc68d` and rides the next
  image — verify in-pod before relying on it).
- **Migrations applied (live)**: `091`–`097b` in `minilobby_task/`. `090` is the
  precedent (the static-label landmine on `content-block-about`).
- **Read for detail**: this file → `RUNNING_NOTES_minilobby_task.md` (the
  2026-07-15/16 execution log at the bottom) → `RUNBOOK_link_integrity_task.md`.
- **The core lesson, reusable everywhere**: a `static` / `site_specs` /
  `renderer`-with-fallback URL source **re-applies on every render** — only
  `source:"llm"` or a component inside `applyCTARecompute`'s set lets a content
  edit win. Migrations 091, 096 and 097b are all this one lesson.
- **Standing landmines**: never full-rebuild `/tools/arena/index.html` (TL-001,
  clobbers the widget); leave `provocation-card` / `lobby-grid` blank fields alone
  (deliberate runtime-fill shells); park `needs_page:provocation` after any
  reconcile; the tool generator ignores the plan slug and doesn't enqueue deploy
  (reconcile + dispatch by hand, verify by artifact); the build-dispatch loop
  stalls on zombie claims (reset-and-retry).

---

## Commits this session (branch `085_debug_and_feature_loops`)
- `9752bc68d` — WS1 Go incl. the phantom-link runtime-fill guard
- `6264e3ebb` — Arena migrations (095b/096/097/097b) + guard unit test
- `283d1c07f` — execution log + honest close-the-loop state
- (migrations 091–095 landed via an earlier checkpoint commit; all applied live)
