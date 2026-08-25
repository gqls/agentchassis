# HANDOFF — ai-agent-orchestration.com. ⚠ SUPERSEDED. Written 2026-08-25 ~19:45Z.

> ## ⛔ SUPERSEDED the same evening by `HANDOFF_2026-08-25c_continue_here.md` — READ THAT FIRST.
> Its central correction — that "contrast is at zero" was FOUR pages of FORTY-TWO — stands, and is
> why this lane now audits the whole site. **Its remaining-work table is out of date: 17 firm
> failures are now 5.** The three `owned` tool pages and 3 of contact's 4 were fixed by migration
> `625` plus an assemble-mode deploy; the manoeuvre is written up in the new file's §1.

**Supersedes `HANDOFF_2026-08-25_continue_here.md`** (written 12 hours earlier, same day).
⚠ **That file's headline claim is WRONG and this file exists mainly to correct it.**

> ## ⚠ "Contrast is at ZERO" was true of FOUR pages. The site has FORTY-TWO. There are 17 firm failures left.
>
> Nothing regressed. `index` / `about` / `pricing` / `services` are still clean. **The failures are
> on pages this lane never measured**, because the four-page scope came from the originating handoff
> and nobody questioned the denominator. Every "44 → 32 → 8 → 0" figure in the earlier docs is
> *"…on those four pages"* and was not labelled as such.
>
> | measured 2026-08-25 evening, all 42 active pages | |
> |---|---|
> | pages audited | **40 of 42** |
> | firm failures | **17**, on 5 pages |
> | overImage (excluded by this lane's convention) | 23 |
> | **NOT MEASURABLE** | **2** — zeros there are silence, not a pass |

---

## 1. The 17, and what each needs

| page | firm | `rebuild_policy` | cause | fix |
|---|---|---|---|---|
| `/tools/agent-complexity-estimator.html` | 6 | **owned** | template HAS the 456 ink fix; placement stale since **2026-05-01** | owned-page chrome path |
| `/contact.html` | 4 | generic | 3 = 456 never propagated (stale since 08-11); 1 = white-on-amber button | ✅ rerender FILED; button needs a 457-shaped template fix |
| `/tools/automation-savings-estimator/index.html` | 3 | generic | **`html_ink` already TRUE — different cause** | needs diagnosis |
| `/tools/password-entropy.html` | 2 | **owned** | template never fixed (`#666` on `#0D1117`) | migration + owned path |
| `/tools/build-vs-buy-analyzer/index.html` | 1 | generic | `html_ink` TRUE — different cause | needs diagnosis |
| `/tools/tool-llm-cost-calculator.html` | 1 | **owned** | template HAS ink; placement stale since 08-11 | owned-page chrome path |

**11 of the 17 are one mechanism — the ink fix exists in the template and the placement was never
re-rendered.** That is a propagation gap, not a code gap.

### ⚠ DO NOT flip the three `owned` pages to `generic` to get a rerender through

They hold their whole tool — calculator and `<script>` — in a single verbatim component, which is
*why* they are owned. The LANDMINE is explicit: the composition loop commits freshly-written HTML to
the deploying repo **one step BEFORE** `save_page_sections` refuses, so the tool is replaced with
prose and shipped before any DB refusal saves you. Someone has already destroyed calculators this
way. Route: `refresh_owned_page_chrome.sh` (see the leopardess lane's `reconcile_footer_nav.sh`
sibling), **not** this lane's usual `template_changed` rerender.

## 2. ⚠ Two pages cannot be measured at all

`ai-readiness-quiz.html` and `tool-ai-agent-roi-estimator.html` — *"probe produced no result"*,
**reproducible**, both serving HTTP 200 at ~55KB. `render_audit.py` prints *"the zeros above are
silence, not a pass"*. Both are JS-heavy tool pages; a page that rewrites its own body can lose the
injected probe. **Until this is understood, no "the site is clean" claim is complete** — and one of
the 9 parked items sits on `ai-readiness-quiz`, so it can be neither confirmed nor retracted.

## 3. Corrections to the earlier handoff — verified here, found by another session

- **"17 parked `contrast_failure` items"** → **9**. Eight were `cancelled` 2026-08-24 19:11:22.
- **"the render audit has not run here since 2026-08-10"** → it visited **02:23Z on 2026-08-24**.

Both were counts measured once on 08-18 and repeated for a week. ⚠ **The 08-22 owner ruling on
dating counts was applied to new numbers and not to inherited ones** — that is the gap to watch.

**The 9 are sharper than the 17 ever were:** 6 sit on `about`/`services`/`index`, which now measure
0, so the retraction should close them; of the other 3, `news` and `tools` measure **0 firm** (so
also stale) and `ai-readiness-quiz` **cannot be measured**. So `bugs_open/296`'s test here is:
**8 of 9 should retract, 1 is undecidable by this instrument.**

## 4. In flight right now

- `page_rerender` for `contact`, `created_by='aiao-contact-propagate'`, **`triaged`**. Queued behind
  `bugfix_391_cta_relevance`, which holds 2 `content_rewrite` rows `claimed` on this site — one
  claimed row of ANY type holds the per-site mutex (`029`). **It drains; do not poke it.** When it
  completes, re-audit `/contact.html` and expect 4 → 1 (the button survives).

## 5. What is DONE and should not be re-derived

- Carousels live on `index` + `enterprise-reference-deployment`, opt-in, other 2 sites verified OFF.
- 10/10 case-study images serving 200.
- `NNN+` incident CLOSED — source fixed by `611`+`613`, live page verified clean on all 7 pages
  checked. Caused by my `557`; `WRONG_CALLS` logged.
- `bugs_open/364` (clock times read as business claims) LIVE — both current chassis stamps contain
  it (`a7459a44b`, `4c996e1b5`; a roll was mid-flight, ~4,896 / ~4,444 pods).
- Migrations in force: `469`, `557`, `559`, `560`, `611`, `613`.

## 6. ⚠ `writer_block_managed` — being handled by someone else, do NOT flip it by hand

Another session has prepared **`617` (HOLD, council APPROVED r1)**: it carries `611`'s prohibitions
into `writer_block_guidance` (CLM-029's first consumer), pre-composes `writer_block` to the real
composer's exact bytes, and adds a chassis guard refusing any pod predating the carry. **The blocker
the earlier handoff describes is being cleared by them.** `617` is `_HOLD` for ordering and applies
deliberately. Do not re-derive the analysis and do not flip the flag.

## 7. Next actions, in the order I would take them

1. **Verify the `contact` rerender landed** (§4) — cheapest, already filed.
2. **The three `owned` tool pages** (9 failures) via the owned-page chrome path. Read the
   flip-to-generic landmine FIRST. Biggest single win left.
3. **Diagnose the two `html_ink`-TRUE pages** (4 failures) — they already carry the token, so the
   cause is something else and guessing will waste a cycle.
4. **The `contact-form` button** (white on amber, 2.08:1). Same shape `457` fixed for `.stats-cta`;
   the fix is `--color-accent-text` with a two-level fallback. ⚠ **`contact-form` is on 20 SITES** —
   this is a fleet change, not a lane change. Measure per-site before repointing, and tell the
   consumers (owner ruling 2026-07-29 §3). `457` deliberately left siblings alone as unmeasured;
   one is now measured.
5. **The two unmeasurable pages** (§2) — an instrument problem, possibly worth a bug.

## 8. Practice notes that earned their place this week

- **Question the denominator.** A scope inherited from a handoff is a claim like any other. I
  verified everything I changed, at the artefact, and still reported a four-page number as a site
  number for a week.
- **Prose written into a prompt is INPUT, not documentation** — `557` shipped `NNN+` to the public.
- **Ask the CAPABILITY, not the commit**: `service_binary_capabilities` + `git merge-base`, never a
  binary grep for your own sha (`buildcapability.go`; two lanes burned).
- **Before concluding nothing is happening, ask what is `claimed`.** I nearly filed a duplicate
  rebuild while the real one was in flight.
- **A peer lane's CONTRIB is another doc** — twice this week the reported defect was narrower than
  the real class (2 writer_lines vs 5; "3 pages 404" vs 0).
- **Hard-wrapped docs make `grep -F` report false absences** — unwrap, or use `git diff --numstat`.
