# The live arm test — STE and an improved house voice, run through the framework

**2026-08-09.** Companion to [`COMPARISON_2026-08-09_ste_vs_house_voice.md`](COMPARISON_2026-08-09_ste_vs_house_voice.md),
which was desk analysis. This is what happened when the blocks were actually put
through `page-content-writer`.

**Nothing shipped.** Every component on the target page was locked for the duration; all
five are byte-identical to the pre-run backup afterwards, on `content_data`,
`rendered_html` **and** `updated_at`. Site serving 26/26, toolgolden exact.

## 1. The target had to change, and why that is itself a finding

The owner chose **loancash.co.uk's homepage**. It cannot be done, and the reason matters
beyond this test.

**All 18 loancash pages are a single `Ported Page` component** carrying
`{"generator": "adoption-locked/1", "deploy_mode": "verbatim", "byte_source": "deploy-repo"}`.
The prose lives in `rendered_html` (8,445 bytes on the homepage); `content_data` holds
248 bytes of port metadata and no copy at all. There are **zero framework components on
the entire site**, so a content rewrite has nothing to read and nothing to write.

That also explains a standing mystery in that site's queue: **18 `needs_page` items sit
`failed` with "a section had no stored content_data"** — one per page, filed 2026-08-08.
They are not a transient build failure. They are the pipeline correctly reporting that a
verbatim port has no regenerable content. `loanandmortgagecalculator.co.uk` is in the same
state (42 ported components, 0 framework).

**[MEASURED 2026-08-09]** across all 21 deployed sites:

| | ported | framework |
|---|---:|---:|
| loancash.co.uk | 18 | **0** |
| loanandmortgagecalculator.co.uk | 42 | **0** |
| webdesign.co.uk | 97 | 9 |
| loancalculator.co.uk | 51 | 12 |
| every other deployed site | 0 | 20–124 |

**cookly.uk was the scientifically better target and was unavailable.** It is the only
deployed site with no voice spec, so it is governed purely by the house voice this change
touches. But a concurrent session was actively working it (435 references, 27 minutes
before the check) and it has open `needs_composition` / `needs_design` items. Firing at it
would have collided with live work.

So the run went to **loancalculator.co.uk's homepage**, capture-only.

## 2. How capture-only was made safe

Not by trusting a lock column name. The enforcement predicate was read from source —
`platform/orchestration/datahelpers/chrome_render_inputs.go:93`:

```
agent-writable  ⟺  locked_at IS NULL
                   OR (lock_type = 'timed' AND lock_expires_at IS NOT NULL AND lock_expires_at < NOW())
```

Three rows were writable (`prose-1`, `prose-2`, `prose-4`); `prose-0` and `tool-3` already
carried permanent locks. The three were locked with `lock_type='timed'` and a **24-hour
expiry** — non-writable immediately, and self-clearing if this session died mid-run rather
than leaving the page stuck. All five then returned `agent_writable = f` under the
production predicate itself, not under a paraphrase of it. Backup table:
`page_components_bak_20260809_stearms`.

**The `llm_call_log` truncation trap was checked, not assumed.** A cut completion logs
`output_tokens = NULL` and declares itself only in `error_message`. No arm truncated —
the largest call was 3,555 of 16,000 tokens.

**The writer runs as a CHILD orchestration.** `llm_call_log` rows carry the child's
`orchestration_id`, not the one the dispatch script prints. Querying by the printed id
returns zero rows while the run is plainly progressing. Join through
`orchestration_states.parent_orchestration_id`.

## 3. What the arms produced

Same page, same facts, same brief. Arms 1 and 2 are the existing record.

| | arm 2 — LIVE (voice H) | arm 3 — raw STE | arm 4 — house + 4 mechanisms |
|---|---:|---:|---:|
| sentences | 31 | 52 | 41 |
| mean words/sentence | 13.8 | **12.2** | 15.1 |
| max sentence | 27 | **23** | 27 |
| over 25 words | 3% | **0%** | 5% |
| over 30 words | 0% | 0% | 0% |
| contractions | 26% | **0%** | 22% |
| phrasal verbs | 3% | **0%** | 5% |
| banned modals | 0% | 0% | 0% |

**The writer can follow STE almost perfectly.** Zero contractions in 52 sentences, zero
phrasal verbs, zero sentences over the cap. Compliance was never the open question —
desirability is.

> **The sentence-ceiling mechanism proved nothing on this page, and I nearly reported that
> it did.** The live homepage was *already* compliant: max 27 words, 0% over 30. My
> site-wide figures — 20.3% over 25 words, longest sentence 49 — come from the guide
> pages. The homepage is the most-rewritten page on the site and had nothing left to fix.
> A ceiling that never binds cannot demonstrate its value. **To test that mechanism the
> target must be a page that currently violates it** — `guides/can-i-overpay` (34- and
> 38-word sentences) or `guides/uk-lending-landscape`.

## 4. The three findings that do hold

### 4a. STE's spelling rule breaks itself, within one page

STE mandates American spelling. Arm 3 wrote **"Amortisation" in section 2 and
"amortization" in section 4 of the same page**, plus "installments". Sections are generated
independently, so a global orthographic rule drifts *between* them — and the result
violates STE's own strictest rule, one word in one form used consistently.

Arm 4 held British throughout (`amortisation` ×2, `instalments` ×1), matching live.

**The transferable point is not about spelling.** Any voice rule whose correctness is
*global to the page* will drift when the page is written section by section. A per-section
prompt cannot enforce a per-page invariant. That applies to terminology consistency and
synonym rotation too — the rules our checker already cannot see.

### 4b. Both arms tried to overwrite the owner's approved copy

The `lock_blocked_change` items name the slots the handler wanted to write:

```
Lock held: save_page_sections wanted to overwrite locked section "prose-0" ...
Lock held: save_page_sections wanted to overwrite locked section "prose-1" ...
Lock held: save_page_sections wanted to overwrite locked section "prose-2" ...
Lock held: save_page_sections wanted to overwrite locked section "prose-4" ...
```

`prose-0` is the owner's personally approved opening, locked
`loancalculator_owner_approved_20260809`. **A whole-page voice instruction was read as
applying to it, by both arms.** Nothing in either prompt said "leave the approved copy
alone", and neither arm inferred it.

This is direct confirmation of `HANDOFF_2026-08-09b` §2: **the lock is the control, not
the wording.** The rule "phrase it conditionally AND lock every sibling you are not
targeting" is not belt-and-braces — the belt does not hold on its own. Had these rows been
unlocked, this run would have destroyed owner-approved copy for the second time.

### 4c. Arm 4 caught a live house-voice violation — and invented links

The house voice says *start with the fact, never open by saying what something is NOT*.
The **live** copy breaks it:

> LIVE: "If you borrow money, you'll pay it back in monthly instalments. **But that monthly
> payment isn't just** the loan amount divided by the number of months you're borrowing it
> for. Instead it's split into two parts: Principal … and Interest …"

> ARM 4: "If you borrow money, you'll pay it back in monthly instalments. **Each payment
> splits into two parts:** Principal … and Interest … That's different from simply dividing
> the loan amount by the number of months you're borrowing it for."

Arm 4 leads with what the thing *is* and demotes the contrast to a trailing clause. That is
the rule working, on copy that is live right now and passed every previous review.

**But arm 4 also added two internal links to `prose-0`, which contains none** (verified:
zero anchors in the live component). The brief said preserve links exactly as they appear.
Adding links is outside a voice-only rewrite, and it is the kind of change that reads as an
improvement while quietly widening scope.

Arm 3 lost meaning in the same paragraph: *"shows how much of a deal is interest rather
than car"* became *"shows the interest amount within a deal"* — the comparison that was the
whole point is gone. It also wrapped its output in an extra
`<section data-component="ported-prose">` wrapper the original does not have.

## 5. What this changes about the recommendation

Nothing in the headline: **do not adopt STE as the house voice.** §4a's four mechanisms
stand, with one demoted and one sharpened.

- **Hard per-sentence ceiling** — still worth having, but **untested**. Retest on a page
  that currently breaches it.
- **Procedural/descriptive classification** — arm 3's `NOTE:` and `CAUTION:` blocks appeared
  where the page describes tool limits, unprompted by anything page-specific. The mechanism
  fires. Whether we want that register is a separate question.
- **Substitution table** — no evidence either way from this run.
- **Post-draft self-check** — no evidence either way. Both arms complied with their own
  ceilings, so nothing needed catching. Also needs a page that breaches.
- **NEW, and the most useful thing here: a page-level invariant cannot be enforced by a
  per-section prompt** (§4a). Any rule about consistency *across* the page — spelling,
  terminology, one-name-per-thing — needs a check that sees the whole page after assembly,
  not an instruction repeated to each section. That is an argument about *where* the rule
  lives, and it applies to the shipping question §4a still has open.

## 6. Reproducing this

```bash
# carriers (spec holds the voice block under test; status 'cancelled' so nothing polls them)
#   arm3 cd06f032-51d8-4615-9b8d-6b6a151feb29   arm4 91c8b30b-43bc-4217-9362-d082057594da
SRC_ITEM=<carrier> KEY_PREFIX=<arm> SUMMARY='<what it is>' ./voiceh_rewrite_v3.sh <page>

# the writer is a CHILD orchestration — the printed id logs nothing
SELECT orchestration_id FROM orchestration_states WHERE parent_orchestration_id='<printed>';

# truncation lives in error_message, never in output_tokens
SELECT step_name, output_tokens, max_tokens,
       (error_message ILIKE '%stop_reason=max_tokens%') AS truncated,
       length(response_text)
FROM llm_call_log WHERE orchestration_id='<child>' ORDER BY created_at;

# proof nothing was written — diff every column against the backup, not just content
SELECT b.slot_name, b.content_data::text=p.content_data::text,
       COALESCE(b.rendered_html,'')=COALESCE(p.rendered_html,''), b.updated_at=p.updated_at
FROM page_components_bak_20260809_stearms b JOIN page_components p ON p.id=b.id;
```

Scoring: `score_arms.py` (this directory) over the dumped `response_text`, reusing
`ste_audit.py`'s rule tables so arms and live baseline are scored identically.
