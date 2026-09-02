# 432 — a site whose DB rows were DELETED keeps serving to the public, and every detector enumerates FROM those rows, so nothing can see it

**Filed 2026-09-02** by the `gamedesign.uk` session, on the owner's instruction to find why
an adoption broke a site. The broken site was the symptom; this is the reason it stayed
broken for 4.5 months. **Status: OPEN, UNOWNED.**

**This is an ABSENCE claim plus measured damage**, both first-hand. Read §4 before treating
the absence as settled.

> **CORRECTED 2026-09-02, same day, by an independent re-investigation on a second model
> (Fable) at the owner's request, then spot-checked by the filing session.** The core
> mechanism, the discriminating control (4/11 vs 0/139 — reproduced exactly), the no-row
> state and the detector absence all **held**. Seven details did not, and are corrected
> in place below, each marked. Two of the re-investigation's own refutations were wrong
> and are noted where they arise — **a second report is still a report; the filing
> session re-measured every load-bearing refutation before accepting it.** The most
> important new finding is §3a: the "defect is closed" claim is true for NEW publishes
> only — an already-served empty artefact with zero component rows is never repaired,
> and the rerender that "completes" on it is a skip.

⚠ **Number collision risk:** 431 was the highest at filing. Another lane took 425→429 today
after exactly this race. Resolve by slug, and `git log` the FILE PATH, not the number.

## 1. The one-paragraph version

Retiring a page sets `pages.status='archived'` and leaves the artefact serving —
`bugs_closed/359` built the detector for that. But a page (or a whole site) can leave the
database by a second door: **its rows are DELETED outright.** Nothing then marks it
archived, because there is no row left to mark. Every check that reasons about what we are
publishing enumerates `FROM pages` or `FROM sites`, so a domain with neither is **invisible
by construction** — not reported clean, absent from the report entirely, which reads as
nothing to say. `gamedesign.uk` has been in exactly this state, serving broken pages to the
public, since **2026-04-16**.

## 2. The damage, measured 2026-09-02 with a fabricated-URL control

```
CODE  BYTES   URL
404   2846    gamedesign.uk/this-path-does-not-exist-9z8x7.html   (control — not a catch-all)
200   15814   gamedesign.uk/                     <- <main>\n\n</main>, EMPTY
200   15831   gamedesign.uk/tools.html           <- EMPTY (born empty 2026-04-18; see correction)
200   15813   gamedesign.uk/about.html           <- EMPTY
200   15837   gamedesign.uk/getting-started.html <- EMPTY
200   15793   gamedesign.uk/services.html        <- EMPTY
404   2846    gamedesign.uk/privacy.html         <- linked from EVERY page footer
404   2846    gamedesign.uk/terms.html           <- linked from EVERY page footer
```

**The control is what makes this evidence:** the invented URL 404s, so this domain is not a
parked catch-all and its 200s are real objects. ~~**Six of nine linked pages serve a header,
a footer and a literal empty `<main>`.**~~

> **CORRECTED 2026-09-02:** **five** of the linked pages are empty; the other two "bad" ones
> are **404s**, which I had folded into one count. And the nav is not the surface: the
> directory holds **47 HTML files, all byte-identical to repo HEAD, of which 13 serve an
> empty `<main>`** — the five above plus `guide-post.html`, `jump-physics-tool.html`,
> `tools/index.html`, `tools/progression-architect/index.html`,
> `guides/rng-design/index.html`, `guides/tool-lanchester-combat-calculator-guide.html`,
> and two `blog/` posts [Fable, MEASURED 2026-09-02]. Caught by walking the directory
> instead of the nav.

Not client-side injection: the only `<script>` is a 320-char menu toggle, and a fetch with a
Chrome UA returns the identical empty `<main>`.

**DB state, three widening queries, all 2026-09-02:**

```sql
SELECT id FROM sites WHERE domain ILIKE '%gamedesign%';              -- 0 rows
SELECT domain FROM sites WHERE domain ILIKE '%game%';                -- only gamesdesign.co.uk
SELECT domain FROM sites WHERE to_jsonb(sites)::text ILIKE '%gamedesign.uk%';  -- only gamesdesign.co.uk
```

No site row, no page rows, and the objects have been in the bucket since **2026-08-02**
(`last-modified`), ~~frozen since the content was emptied on **2026-04-16**~~.

> **CORRECTED 2026-09-02:** two populations, which I had conflated. **Five files LOST
> content** — four on 04-16 (`index.html`, `tools/index.html`,
> `guides/rng-design/index.html`, `tools/progression-architect/index.html`) and
> `about.html` on **04-18** (`c36418898`). **Eight were BORN empty**, never populated —
> and `tools.html` and `getting-started.html` did not exist until 04-18, so my NOTES line
> "0 chars both before and after" for them was an artefact of measuring a file that was
> not there (see WRONG_CALLS). The flagship page whose content was LOST is
> `tools/index.html` (1,106 chars → 0, `8eba6b30a`), not `tools.html`. Last render of the
> directory: 04-18 21:24.

## 3. How it got there (the origin, for context — NOT the defect this file is about)

The April 2026 adoption ran with gamedesign.uk as **both source and destination**. It wiped
and recreated the page rows — the adoption thread's own handoff says *"Pages: 11 (clean —
wiped and recreated by latest adoption)"* — and the rerender then published the empty
placeholders over the live HTML, at a time when nothing guarded that seam. The homepage went
5,977 chars (04-14) to 0 (04-16) in commit `f9838491d`, "Rerender: index.html", −278/+6.

~~**That publish defect is CLOSED**, by three guards that all postdate the damage:~~
~~`rerender_single_page_action.go:581-602` (2026-05-12, …), `load_page_sections_from_spec_action.go`~~
~~fallback 4 (2026-06-08, …), and the `bugs_open/095` split (2026-07-27, …).~~

> **CORRECTED 2026-09-02:** three errors in one sentence.
> 1. **Two guards, not three.** `load_page_sections_from_spec_action.go` fallback 4
>    (`856fc4a51`, 2026-06-08) is a **build-side layout rescue**; that file's terminal
>    branch still returns `sections: [], source: "none"` as a silent success and **never
>    refuses** anything. The publish-seam guards are `if len(sections) == 0 { return "",
>    assembly, nil }` (`d777cb4d2`, 2026-05-12) and `assembledToNothingDespiteComponents()`
>    failing the step (`6579e9ae1`, 2026-07-27). A third, `e6a8bb63b` (2026-07-30), guards
>    only the verbatim path.
> 2. **Never cite a bare line number on this file.** It read `:581-602` when I looked,
>    652–673 when Fable did hours later, and **680** at `11414e733` when I spot-checked —
>    `improvement_loop` is editing it daily. Cite the commit sha and the code string.
> 3. **"CLOSED" is true for NEW publishes only — see §3a.** The guards stop an empty page
>    being written; nothing repairs one already written whose row has zero component rows.
>
> The **running binary DOES carry all five commits** — stamp
> `a2732c7207da4f24ed3aceb6f62b238605db0530` from `service_binary_capabilities` (the
> CLAUDE.md log grep is out of range even at `--tail=20000`), each verified with
> `git merge-base --is-ancestor`, with HEAD→stamp NO and stamp→HEAD YES as controls
> [Fable, MEASURED 2026-09-02]. My NOTES' `[UNVERIFIED AT THE ARTEFACT]` is now verified.

**This file is about the second half: why nothing noticed for 4.5 months afterwards.** The
site row was later deleted, ~~and with it every handle any detector or repair path had~~.

> **CORRECTED 2026-09-02:** one handle survives. **1,147 rows in `site_work_items_archive`**
> carry `site_id = 15a6cb16-5a86-4541-a8e4-d7106239b6a4` and `domain = 'gamedesign.uk'`,
> created 2026-04-02 → 2026-04-15 [Fable, MEASURED 2026-09-02]. They date the adoption
> runs (04-02, 04-03, 04-15 ×2, mass `wont_fix` at 04-15 11:34:24 — the fingerprint of the
> 04-16 handoff's "full cleanup procedure"), and they show the platform **did** file
> `empty_section` on `index`/`tools`/`about` in April and closed 32 of them as "superseded
> by duplicate". The deletion of the `sites` row itself is **undatable** from the DB
> (cascade-deleted dependants, no audit rows, ~2-day `orchestration_states` retention);
> bounded alive at 2026-04-20 (composition verified on it in that day's handoff) and gone
> by 2026-09-02.

### 3a. The adjacent LIVE finding — an already-served empty page is never repaired

`https://ai-agent-orchestration.com/roi-estimator.html` serves `<main>\n\n</main>` today
(invented-URL control on the same domain: 404). Its `pages` row is **`active`, `deployed`,
0 `page_components` rows**, and it carries **eight `page_rerender` items marked `complete`
between 2026-08-26 and 2026-09-02** [MEASURED 2026-09-02, both sessions]. The guard fires,
the step reports `skipped: true`, page-rerender's `check_skipped` routes that to
`complete_skipped`, the item closes as `complete` — and the empty artefact is untouched.
That is `bugs_closed/315`'s profile ("rerender skips and completes"), live on a site
with a row. Two more on other sites from the same 04-18 fleet-wide born-empty wave still
serve empty: `ai-agent-orchestration.com/llm-cost-calculator.html` (row `archived`,
deployed 04-18 — `bugs_closed/359`'s class, still serving) and
`robot-hands.com/learning-center-article.html` (row url is `/blog/…`, served at root — a
`pages.url` probe would miss it) [Fable, MEASURED 2026-09-02]. ~~**Not filed separately by
this session** — they are other lanes' sites; flagged to the owner.~~

> **UPDATE 2026-09-02 evening:** owner ruled "reopen 315 and hand it to that site's lane" —
> done. `bugs_open/315` is REOPENED and owned by `site_ai_agent_orchestration`, which built the
> fix the same day (`8eca969cb` + `8a0b927f5`, Council-Submitted `2be8ec34`; inert until a
> chassis roll): both the consumer skip-branch and the producer loop now file a deduped
> `needs_content_page` instead of a guaranteed-skip rerender. Their census, corrected by the
> council from my "one instance": **14 rows** `active` + `deployed_at IS NOT NULL` + 0
> component rows. `llm-cost-calculator.html` (archived-and-serving) is recorded there as an
> ~~**owner decision** — retract vs un-archive-and-build~~ — **CORRECTED 19:40Z (AI page 3, at the
> artefact):** the canonical tool ALREADY EXISTS at `/tools/tool-llm-cost-calculator.html` (built
> 2026-09-02, 67 KB); the flat-URL page is a stale empty DUPLICATE still linked from the homepage
> and `/tools.html`, plus a third empty shell at `/guides/tool-llm-cost-calculator-guide.html`. So
> not a build and not a naive retract (that 404s live nav): repoint two nav links, then retract two
> shells — a content decision for that site's owner, recorded in 315. `robot-hands.com/learning-center-article`
> is taken into that session's robot-hands list. Also surfaced there: `componentless_pages` is
> a built discovery check enabled in ZERO discovery agents, and would not catch a `sections=[]`
> page anyway — the real gap sits between `check_sectionless_pages` and it.

## 4. The absence, scoped — and the disconfirming check I ran

The claim is "no check enumerates the SERVING surface and reconciles it against the DB".
Census of `cmd/` (**31 services as of 2026-09-02**), counting enumeration axis:

| service | `FROM pages` | `FROM sites` | bucket-side |
|---|---|---|---|
| component-render-check | 0 | 0 | 1 |
| config-key-audit | 3 | 4 | 3 |
| content-loss-check | 1 | 1 | 0 |
| instanceaudit | 0 | 0 | 5 |
| webdesignport | 1 | 1 | 0 |

**The two bucket-side counts are the disconfirming candidates, and I opened both rather than
trusting the grep.** `instanceaudit`'s five hits are a Go `map[string][]string` named
`buckets` used to group classification results. `component-render-check`'s single hit is the
word "bucket" in a prose comment about where a parse failure is counted. **Neither touches
storage.** So the claim survives the check that could have refuted it.

Everything that reasons about what we publish starts from a DB row. `pages.status` has no
value meaning "this row was deleted", because a deleted row has no values at all.

### What this is NOT a duplicate of — the family, and where each stops

Grepped `bugs_open/` and `bugs_closed/` before filing. Every member of this family
**presupposes a row exists**; that is exactly the assumption that fails here.

- `bugs_closed/359` — no detector for a retired page still serving. Its fix
  (`scripts/audit-archived-still-serving.sh`) enumerates `pages.status='archived'` with a
  non-null `deployed_at`. **A deleted row is not an archived row.** Same gap one level up.
- `bugs_open/098` — archiving a page does not retract it. Presupposes the archive.
- `bugs_open/304` — retracting the LAST page of a site cannot unpublish it. Adjacent (whole-
  site end of the seam) but still driven from rows.
- `bugs_open/429` (filed today, OWNED by the delivery lane — do not compete) — the publish
  mirror never propagates deletions, so a retracted page persists at the hosted copy.
  **That is origin-clean / mirror-stale. This is origin-serving with nothing to enumerate.**
- `bugs_open/356` — discovery checks select pages on the BUILD axis only. Closest in spirit
  (a selection-axis defect) but its subject is a row on the wrong axis; here there is no row.
- **Added 2026-09-02** (missed at filing, found by the re-investigation):
  `bugs_closed/185` (detectors select on `deployed`) and `bugs_closed/315` (a rerender that
  skips reports COMPLETED) — 315 is the profile §3a shows still live.

## 5. Why it is a real defect and not tidy-up

- **It publishes content nobody has decided to publish.** Whatever the site was for, no
  current record says it should be serving anything, and it is serving broken pages under a
  domain we own.
- **The blind spot is structural, not a missed case.** Adding a status value cannot fix it;
  the row is gone. Any detector for this class has to start from the serving surface (bucket
  keys, or DNS + a crawl) and reconcile INTO the DB — the opposite direction from every
  check we have.
- **The population is unmeasured, and that is the first thing a fixing thread owes.** I have
  one instance, found by hand because the owner pointed me at the domain. ~~`~/projects/sites`
  holds **20+ top-level domain directories as of 2026-09-02**; how many have no `sites` row
  is a straightforward query nobody has run.~~ **Do not read "1 instance" as a rate.**

> **CORRECTED 2026-09-02, and the correction was itself corrected.** The deploy Action's
> filter (`deploy-to-b2.yml:34`, `^[^/]+\.[^/]+/$`) matches **36** top-level directories,
> not "20+" — I guessed low. Fable's report put it at **19**, which was also wrong. Measured
> against `SELECT domain FROM sites`: **8 of 36 have no `sites` row** [MEASURED 2026-09-02]:
> `gamedesign.uk`, `nanangmrk.com`, `oxenunity.com`, `puzzlegenerators.com`,
> `testllmlog.example.com`, `website-design.com`, `websitedesign.com`, `wykefarm.uk`.
> Fable probed four of its six and found only gamedesign.uk carrying empty-`<main>` damage
> (the others are hand-built, mostly with no `<main>` element at all). `oxenunity.com` is a
> known hand-authored single page (`oufe/RUNBOOK_oufe.md` §1). The other row-less serving
> domains are **unprobed by this session** — a fixing thread's first census.

## 6. Fix candidates, ordered by what closes the door

1. **Reconcile the serving surface against the DB (closes it).** Enumerate top-level
   directories in `b2://portfolio-sites/` (or the `~/projects/sites` tree, same set) and
   report any whose domain has no `sites` row, or has one with no `pages` rows. This is the
   only candidate that makes the bad state *visible* rather than *unrepresentable*, and it
   needs no schema change. Cheap: the listing is one `b2 ls` and one `SELECT domain FROM
   sites`. Both controls from `audit-archived-still-serving.sh` apply unchanged and must be
   carried over — an invented URL must 404, a known-good sibling must 200.
2. **Make deletion impossible where archival is meant (closes the door upstream).** If no
   code path is supposed to hard-delete a site or its pages, add the guard and let archival
   be the only retirement. Requires a census of what deletes today
   (`080_trigger_adoption.sh` does exactly this to `pages` by design) before it can be
   scoped — a reset script for a dev loop is a legitimate deleter.
3. **Retract the orphan (fixes the instance, not the class).** Whatever is decided for
   gamedesign.uk specifically, that decision is separate from this file.

## 7. How to verify a fix

- **STEP 0, before anything:** re-run §2's probe with its control. If the invented URL stops
  404ing, this domain has been repointed and the whole measurement is void.
- The detector must find gamedesign.uk **while gamedesign.uk is still in this state**, and
  must find nothing on a domain that has a healthy site row (`gamesdesign.co.uk` is the
  natural negative control — 40+ deployed pages as of 2026-09-01).
- **A zero is not a pass** until the demand control fires: seed a throwaway directory with no
  site row and confirm the check reports it. A reconciler that enumerates nothing reports
  clean, which is this class's false all-clear — the same profile as 359's.

## 8. On the 090 diagnosis loop, per the owner ruling of 2026-07-31

This file asserts a cross-cutting structural absence, so the ruling applies. **I did not run
the loop, and state the substitute plainly rather than omitting it:** the absence is narrow
enough to enumerate exhaustively by hand, and I did — every service in `cmd/`, by
enumeration axis (§4), opening both bucket-side candidates that could have refuted it rather
than trusting a grep count. The damage half is first-hand with a per-domain fabricated-URL
control, and the origin half (§3) is settled by the sites repo's own git history plus a
seven-site discriminating control (gamedesign.uk lost content in 4 of 11 files touched on
2026-04-16; the other six sites lost 0 of 139).

**The obvious hole in that substitute, closed in the same session rather than left as
homework.** `cmd/` is not the whole detector surface, so I extended the census to the other
two places a serving-side check could live:

- `scripts/audit-*` / `check-*`: three touch site data
  (`audit-archived-still-serving.sh`, `audit-experience-promises.py`,
  `audit-listing-class-promise.py`) and **all three enumerate `FROM pages`, none
  serving-side**.
- In-chassis Go: the files matching `portfolio-sites|ListObjects|b2 ls` are
  `platform/publish/{publisher,b2worker}.go`, `platform/storage/s3.go`,
  `platform/orchestration/actions/{publish_site,zip_deliverable}_action.go` — **every one is
  the WRITE path.** The only code that lists bucket objects is the publisher putting them
  there. Nothing reads that listing back to ask what is serving.

So the claim stands at its full width: **across `cmd/`, `scripts/` and the in-chassis Go,
nothing enumerates the serving surface and reconciles it against the DB.** Independently
re-derived 2026-09-02 (Fable): four `ListObjects` callers, all write-path or keyed from a
row; 88 `FROM pages` / 43 `FROM sites` across the discovery checks, entered per `site_id`;
no Cloudflare zone enumeration in any check. One useful pointer from that pass: the deploy
Action already calls Cloudflare's `zones?name=` — **a serving-side enumerator has an API to
use**, and the true surface is *zones × bucket prefixes*. The one untested assumption left
is that the deployed `cmd/` CronJob images match the tree (neither session checked).
