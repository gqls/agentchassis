# HANDOFF 2026-08-25 — the greenfield route is now gated on ONE unowned bug, and the only thing left to learn needs ONE build

**Lane:** `loanzy_uk_example_site` — the unaided one-shot greenfield route (submit a domain, touch
nothing, measure what the platform produces on its own). Worked example on disk:
`garden-tools.uk`, built 2026-08-23/24, **deliberately unrepaired**.

**Supersedes** `HANDOFF_2026-08-24_garden_tools_finished_and_what_must_be_fixed_before_the_next_domain.md`
(same directory), which is bannered with the four things in it that went stale. That file is still
the reference for **§3b (the owner's review, verbatim)** and **§3's measured recipes** — nothing in
this file replaces those. The pre-run
`HANDOFF_2026-08-23_garden_tools_continue_here.md` remains the reference for the pre-flight recipe
and DNS/zone setup.

> ## ⚠ SUPERSEDED the same day by `HANDOFF_2026-08-25b_the_canary_served_the_owner_reviewed_it_and_the_council_already_exists.md`
> **The canary this file said was "the only thing left" was authorised, built, cut over and reviewed
> by the owner within the day.** §3 (`376`'s mechanism — now with a second failure mode and a
> verified fix design in the bug's §11) and §5 (the harness — now parameterised, control-gated and
> chrome-stripped) still stand. §4's decision is taken. Read the b-file.

---

## 1. State of the worked site `[MEASURED 2026-08-25 09:57Z, at the served pages, cache-busted]`

Site `16784842-f7d8-4467-bb5b-eb1fb5c1caba`, domain `garden-tools.uk`. Re-measured this morning with
`./after_test.sh` (now promoted into this directory — see §5).

| serving (7) | bytes today | | 404 (5) | why |
|---|---|---|---|---|
| `/index.html` | 66,815 | | `/buying-guides/index.html` | `section-index` — deliberately out of `206`'s fix |
| `/how-we-assess.html` | 67,782 | | `/brand-directory/index.html` | `entity-directory` — `206`, fix live, cannot re-mint |
| `/seasonal-planner.html` | 67,419 | | `/entities/brand-profile.html` | `entity-page` — same, and **unlinked** (orphan) |
| `/about.html` | 65,608 | | `/blog/buying-guide-post.html` | `blog-post` — the layout-less class, `206` §(b) |
| `/care.html` | 65,462 | | `/tools/finder/index.html` | `tool` — owner-gated to human review, working as designed |
| `/contact.html` | 57,753 | working 2-input form | | |
| `/affiliate-disclosure.html` | 54,440 | | | |

**Unchanged from 08-24:** the 7/5 split; **9 dead links across 4 distinct targets** (`/tools/finder/`
×4, `/buying-guides/` ×3, `/brand-directory/` ×1, `/blog/buying-guide-post.html` ×1 — the home page
carries 3); `PROMISE UNMET` fires on `seasonal-planner` and nothing else.

**Changed, and worth knowing why:** every serving page is **exactly +420 bytes** larger than the
08-24 table. That table was measured 09:05Z; `pages.deployed_at` on all seven is **14:00–14:04Z the
same day**. So the figures in that handoff were one deploy stale when they were written. A uniform
delta across seven pages of different lengths is chrome, not content — and the structure counts
confirm it.

**`bugs_open/381`'s pre-fix baseline SURVIVES and is still the estate's only clean one:**
**0 tables, 0 `<strong>`, 0 content lists** across all seven served pages, longest `<p>` 104 words,
3 distinct month names on a page headed *"What your shed needs, month by month"*.
⚠ The `li=8` the harness prints on every served page is **navigation chrome** — identical on all
seven. It is not a content list. Do not report it as one.

**Open review items (one MORE than the 08-24 handoff lists):** 4 × `needs_page` no-op ·
1 × `owned_page_review` (tool-finder) · 1 × `needs_section_data` (contact email) ·
8 × `unresolved_cta` (at least one provably stale) · **1 × `claims_unverified`**
(`claims_llm_garden-tools.uk`, filed 2026-08-24 16:52Z — the first that step has ever filed anywhere,
and it names the owner's own two example sentences).

**⚠ STANDING INSTRUCTION, UNCHANGED: do not repair this site.** Its whole value is that it is
unassisted. If the owner wants a working garden-tools site, that is a different job with its own lane
doc — say so explicitly, do not quietly fix this one.

---

## 2. What changed since the 08-24 handoff — read this before acting on that file

### 2a. The owner RETRACTED the §3a authorisation (2026-08-25)

He had authorised releasing the parked `needs_page:brand-directory-index` row so `bugs_open/206`'s
fix could be proven on this site. **This lane took the correction back to him** (`4741cf682`) and he
has retracted it. **Do not revive the plan.** The two reasons, both measured:

1. **Clearing the row alone is inert.** `reconcile_site_plan` is carried by exactly **one** agent
   (`build-site-planner`) and nothing schedules it. A quiet site never re-reaches the fixed code.
2. **Making it non-inert means a full re-plan** — `plan_site` → `write_site_plan` → `sync_pages`,
   which overwrites `pages.sections` and re-emits design and imagery. That is `bugs_closed/001`'s
   hazard and it would destroy the baseline in §1 that `381` is holding.

**`206`'s proof therefore rides the next greenfield build.** The test is the **MINT**, not the page:
a fresh `needs_page` for an `entity-directory` role carrying `handler_agent='directory-build-handler'`,
filed by reconcile with no hand routing. ⚠ Use `spec->>'page_role'` — **not** `page_type`, `page_id`
or `filename`, none of which exist on a reconcile-minted row (0 of 134 fleet-wide).

### 2b. `bugs_open/380` — TAKEN, fixed, and the Go half is now LIVE at the binary

The 08-24 handoff calls it "UNOWNED, highest priority of all". It was taken ~15:15 BST that day.
Lane: `docs/agent_docs/docs024_key_docs_latest/bugfix_380_claims_fail_open/`.

- **Config, live 2026-08-24:** `check_opted_in` **deleted** — the auditor no longer skips a site with
  no evidence base, and now fails **closed**; the planner mandates `"facts": []` and bans briefing a
  methodology page as practice; the writer's no-register / no-operating-history arm (599) went live
  after the owner read and approved its plaintext.
- **Go, live on chassis `v1.0.1337`** (pods 09:27Z; that lane verified provenance `4c996e1b5` with
  `git merge-base --is-ancestor`). **Independently confirmed here at the binary**
  `[MEASURED 2026-08-25 09:56Z, both replicas]`: `practice_claim` PRESENT, the family's own reason
  string *"no recorded operating history"* PRESENT, `validate_page_content` PRESENT (positive
  control), `zzz_not_a_real_symbol_control` ABSENT (negative control).

> **⚠ LIVE IS NOT FIRING.** The practice family runs inside `validate_page_content`, which runs during
> a build or rerender. This site is quiet, so the check is live and **asking nothing here**. A clean
> result on `garden-tools.uk` today is not evidence — nobody has put a question to it.

> **⚠ The `build provenance` log line had already scrolled by 09:54Z** on a 26-minute-old chassis pod
> (`--tail=6000`, nothing). That is the documented trap behaving exactly as documented: an empty grep
> means "out of range", not "unstamped". The binary capability probe with two controls has no shelf
> life — use it, and never a discovery grep for "some 40-hex string".

### 2c. `bugs_open/381` — TAKEN, both halves live, AND three new components built

Also listed "UNOWNED" on 08-24. Lane:
`docs/agent_docs/docs024_key_docs_latest/bugfix_381_inexpressive_composition/`.

- **Planner arm live 16:53Z** (migrations 591–593): `component_expresses()` derives from each
  component's own template whether it can produce a list, table, repeating set or only paragraphs,
  and all three planner menus now print it.
- **Writer arm live ~16:58Z** (594/595): the four prose slots that carry most estate prose now carry
  real `llm_guidance`, and RULE 10 is re-addressed to `html`.
- **Three new components built overnight** — `checklist`, a period `calendar` (generic over months,
  quarters or seasons) and a `comparison-table`. This closes the gap the 08-24 handoff recorded as
  open ("there is no generic checklist/steps/table/calendar component").
- **Their measured result: 72% of treated writes now produce a list, against a 10% baseline; 100%
  produce a subheading.**

### 2d. `bugs_open/376` — STILL UNOWNED. It is now the only one of the three that is.

No lane directory, no commits since 2026-08-23. See §3.

---

## 3. THE GATE — `bugs_open/376`, and why it now blocks four lanes rather than one

**What it is, plainly.** Early in a greenfield build, a step called `vertical-exemplar-researcher`
asks an LLM to name the three best existing websites in the vertical, then crawls them. Firecrawl
refuses some hosts outright. **The crawl steps have no `on_error`**, so one refusal discards the
whole step — including the crawls that succeeded. And `create_next_item`, the **only** producer of
`needs_strategy` estate-wide, is the last step in that chain. So a refusal is **terminal**: the build
stops and never reaches the planner.

**The rule this offends.** Research is N-of-3, not a transaction. A stage that has two good crawls in
hand should not throw them away because a third host said no.

**How this case measures.** Submission 1 of `garden-tools.uk` died exactly this way. On that vertical
the refused host (`thespruce.com`) appeared in **4 of 5 observed draws** `[MEASURED 2026-08-23]`, and
`max_attempts=3`, so the item usually dies before a lucky draw. Each cycle is 30–60 minutes.

**⚠ Two corrections already on the file — read `bugs_open/376` §4b before quoting it.** The pool is
**biased, not fixed**: the fifth draw re-drew without the refused host and did use a
`competitors_found` candidate, refuting this lane's own earlier "retry is structurally futile" claim.
The severity case never depended on that and is unchanged.

**Fix candidates, in the order the bug file ranks them:**
1. **`on_error` tolerance on each crawl step, with a stated floor** (e.g. proceed on ≥2 of 3).
   Config-only. Cheapest real fix.
2. **Persist refused hosts and exclude them at selection.** The only candidate that gets cheaper
   over time — it removes a 4-in-5 tax on this vertical permanently.
3. *(separable, different owner)* the `competitors_found` branch rarely firing — a site classified
   `hub`/`content` is being compared against retailers. Not required for the fix above.

**⚠ Verifying any fix here: the refused crawl's step record reads `"success": true`** — that is a
dispatch receipt, not a result. Join on `request_id` and verify at the artefact. In `LANDMINES.md`.

---

## 4. THE ONE OPEN OWNER DECISION — and it is narrower than it was yesterday

Three lanes wanted a post-fix greenfield build. **The writer half of that question has answered
itself for free**: 271 `page-content-writer` `generate_content` calls have run since the fix
`[MEASURED 2026-08-25 10:10Z]`, and `381` has measured the effect on the treated population.

**What is still unanswerable without a new-site build:**

| lane | what only a greenfield build can show |
|---|---|
| `bugs_open/381` | does the **planner** choose the new `checklist` / `calendar` / `comparison-table`? They have **never been placed on a page**. The component-choice step runs on new-site builds only. |
| `bugs_open/380` | does a build with no evidence base still hallucinate practice claims, now that all three mechanisms fail closed? |
| `bugs_open/206` | does reconcile **mint** `entity-directory` at `directory-build-handler`? (Cannot be tested on `garden-tools.uk` — see §2a.) |
| this lane | what the unaided route produces on its second run, against a dated first run |

**The options, unchanged in shape:** (i) authorise a new greenfield domain; (ii) name one page on an
existing site for a single writer run; (iii) no dedicated canary — measure whatever builds next.

**What is different today:** option (ii) now buys **almost nothing** — it tests the writer arm, which
is already proven. Option (iii) has delivered everything it can. **Only (i) answers any of the four
rows above**, and (i) walks straight into §3.

**The `381` lane's stated preference, which this lane endorses:** the *subject* matters more than the
domain. Something genuinely structured — a buying guide or a how-to — exercises all three new
components; a two-page brochure exercises none.

**No session may resolve this alone.** Both lanes have said so in writing; so does this one.

---

## 5. What is in this directory now that was not before

- **`after_test.sh` — PROMOTED** from the 08-24 session's scratchpad, per that session's own
  instruction. 148 lines, unchanged except a banner. Sections: `311` collateral · `260` template
  leak · pages table · the artefact (http/bytes/inputs/buttons/CTA/leak/identity) · `328` dead links ·
  **PROMISE vs DELIVERY**.
  - ⚠ **Section (a): re-pin, and CHECK — do NOT dismiss a `*** HTML CHANGED ***` line.**
    > **CORRECTED 2026-08-25 10:29Z, hours after I wrote it.** This bullet originally read *"all
    > eight moved 2026-08-20 … `*** HTML CHANGED ***` means the pin is old, not that your build
    > collided. Re-pin before a build; **ignore section (a) otherwise**."* **That inverted the
    > instrument.** Caught by the `bugs_open/381` lane re-pinning at 10:27:16Z for the
    > `homegarden.uk` build; verified here independently at 10:29Z.
    > `[MEASURED 2026-08-25 10:29Z]` **all eight match the 08-23 pins exactly — 8 of 8, html AND
    > schema — and `content_components.updated_at` on every one reads 2026-08-20, untouched for
    > five days.** The 08-20 move is true as *history* and false as *current state*; I inferred
    > forward from the event and wrote the inference as an instruction. With the pins current, a
    > CHANGED line is a **real collision** — the one alarm that section exists to raise — and my
    > wording would have had the next reader dismiss a true positive mid-build.
    > Full entry: `WRONG_CALLS.md`, 2026-08-25.

    **Operative:** re-pin before a build (one query, and `LANDMINES.md`'s entry has it). If you
    have not, treat a CHANGED line as **news**: check that component's `updated_at` before
    deciding whether it was you or `bugs_open/283`.
  - ⚠ Its `li=` column counts **all** `<li>`, chrome included (8 per served page here). Read it as a
    delta against the baseline, never as an absolute content-list count.
  - It validated on live data before this run: `PROMISE UNMET` fires on `seasonal-planner` and stays
    silent on the other 11 — a check that can come out both ways.

---

## 6. WHAT IS NEXT ON THIS LANE, in order

1. **Take `bugs_open/376`.** It is this lane's own filed bug, it is the only unowned one of the
   three, and it is the gate on the single artefact four lanes now need. Fix candidate 1 is
   config-only. Do candidate 2 alongside if the round allows — it is the one that stops paying the
   tax. Council-gate it (config on `agent_definitions` is in scope via the migration widening).
2. **Then, and only with the owner's authorisation, the greenfield canary** — §4. Pre-flight recipe
   in `HANDOFF_2026-08-23_garden_tools_continue_here.md`; re-pin the `311` md5s first; expect
   time-to-first-agent = queue depth ÷ ~90s.
3. **Read the canary's rendered prompts THE SAME DAY.** `[MEASURED 2026-08-25 10:33Z]` — measured
   here after the `bugs_open/381` lane pointed out that both of us had been repeating "~24h" from
   memory without ever running it: **7,668 rows in `orchestration_states`, oldest 1d 00:57, 0 older
   than 48h**, distribution 1,137 under 6h · 6,365 at 6–24h · 166 at 24–48h. So the tail runs to
   **about 25 hours, not a clean 24** — but the operative advice is *same day*, because a prompt you
   plan to read "tomorrow morning" is a coin-flip. `llm_call_log` keeps prompts AND responses with no
   such window (it is the training corpus, not a log); `site_work_items` ∪ `site_work_items_archive`
   for item history.
   > **Why this is dated rather than stated:** had the answer come out at 6h, this line would have
   > sent the next reader looking for a row that no longer existed — and an absence reads as a failed
   > build. **An unmeasured caution is a coin-flip written in the imperative.** Same defect class as
   > §5's corrected banner; see `WRONG_CALLS.md`, 2026-08-25.
4. **Leave `garden-tools.uk` alone** until the owner says otherwise, and keep §1's baseline dated in
   any doc that quotes it.

**Not this lane's work, tracked so nobody re-chases it:** `380` and `381` are owned and mid-flight —
contribute, do not compete. `206` is owned and six council rounds in. The card-wall composition
question (`381` §7) needs an owner decision about page composition before it is worth filing at all.

---

## 7. Falsifiers for this handoff

- `376` gaining an owner, a lane directory, or an `on_error` on the crawl steps.
- Any of the five 404 pages beginning to serve — someone acted, or a fix rolled and reconcile ran.
- `sites.last_reconciled_at` for this site moving off **2026-08-23 20:15:50** — the baseline is then
  at risk and §1 must be re-measured before anyone quotes it.
- The apex serving anything other than the current **66,815**-byte index.
- A new-site build appearing anywhere on the estate — that is the canary, authorised or accidental,
  and §4 becomes answerable.
- Firecrawl beginning to accept `thespruce.com`. The whole `376` hazard rests on one host's blocklist
  status, which is not ours and can change without notice.
