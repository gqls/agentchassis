# PLAN — 2026-08-25 — CTA destination relevance (`bugs_open/389`)

## What we are trying to do
Stop the framework choosing a site's most prominent button by a sort that cannot know what the
site is about. Filed on an owner report: a password-strength toy was the call-to-action on an
AI-orchestration consultancy.

## Phasing
- **Phase 0 (DONE, 2026-08-25):** find the cause; prove it at code, data and served bytes; size
  the blast radius; separate the live half from the historical half. Filed as `bugs_open/389`.
- **Phase 1 (BLOCKED — owner decisions):** four decisions, in the bug file and the handoff §3.
  They are genuinely different kinds — content, data, platform, repair — and picking one does not
  imply the others.
- **Phase 2:** implement whichever platform option is chosen. Architecture-scope; council gate
  before or alongside the commit; per 2026-08-02 §2 new authority ships opt-in, default OFF.
- **Phase 3:** repair the 80 stored values via `bugs_closed/268`'s fleet re-run, and only then.
  Constrained by `bugs_open/248` (recompute clobbering authored links).

## Decisions and their reasons
- **File, do not patch.** Three `UPDATE`s would clear today's embarrassment in a minute. Rejected
  as the whole answer: the resolver minted a fresh instance **today**, so data repair without a
  mechanism decision is undone by the next run. Recorded because the temptation is real and the
  next session will feel it too.
- **Do not widen to the borderline case yet.** `webdesign.co.uk` → `tool-ab-test-calculator` from
  66 tools might be fine. Its `nav_order` is the default 100, not a fossil 1, so it is a different
  and weaker shape. One human glance before it joins the bug.
- **Corrections to the originating read, kept visible:** my first framing — "13 sites show a
  deliberate hide-vs-rank contradiction" — was **wrong** and is retracted in the bug file and in
  NOTES. `in_header=false` is the majority state for tool pages (62.7%), so it does not carry the
  meaning I gave it.

## Risks
- **The coupling is in the fix as well as the bug.** Using `nav_order` to fix ranking moves the
  visible menu on any site where the page is `in_header=true` (that is `ai-agent-orchestration.com`
  today).
- **A repair run re-mints.** Ordering matters: mechanism first, data second.
- **Relevance scoring invites false positives** and the existing `semantic_tags` are already
  misleading — the migration that claimed to *narrow* password-entropy's affinity **added**
  `tech`, `cybersecurity`, `developer` to it.

---

## RESIZED 2026-08-25 (same day, after adversarial review) — the phasing above is superseded

**What changed:** the review confirmed the mechanism and found the loop the diagnosis sat inside —
the label match runs ahead of the positional pick, and `stampCTADestinationGuidance` has the writer
produce copy naming whatever was picked. A wrong pick therefore becomes **label-locked**, and a
`nav_order` fix cannot reach it. Full working: `bugs_open/391` §THE FEEDBACK LOOP.

**The bug number changed too:** this lane's bug is now **`bugs_open/391`**, not 389 (collision with
the `bugfix_308` lane's re-file, which was 2m25s earlier; 390 was taken in the interval).

### Revised phasing

- **Phase 1 (BLOCKED — owner decisions):** now **five**, not four. The fifth is the standing
  commission (honour / re-scope / withdraw), which is a decision about the owner's own 08-15
  instruction and cannot be folded into the repair.
- **Phase 2 — ranking first.** Whatever platform option is chosen must change the **ranking**, not
  the loaders: the loaders have a third consumer (the site header CTA fallback) whose output is
  never persisted, so a loader change moves every site's header button invisibly. And an opt-out
  flag must also be read by `LoadCTALabelUniverse`, or it has a hole exactly the shape of this bug.
- **Phase 3 — the content pass, RE-SCOPED not skipped.** ~20 label-locked fields, selected by
  query (in the bug's §4), not 16 sites. This reverses what I wrote this morning.
- **Phase 4 — repair**, verified at the served bytes, never by work-item status (`bugs_open/389`).

### Decisions and their reasons — additions

- **"File, do not patch" still holds, but for a sharper reason than I gave.** It is not only that a
  data fix is undone by the next run; it is that a data fix **cannot reach the damage the owner
  reported at all**, because those three buttons are label-locked.
- **The recommendation is no longer candidate 1 alone.** An opt-out is reactive and does not make
  the bad state unrepresentable — it makes the good state sayable. Candidate 1 **paired with**
  candidate 4 (a detector for the anomalous-`nav_order` shape) is what earns "closes the class".
- **RFC_022 must be engaged before booking a council round**, including the consumer enumeration it
  requires — asserting the shape without the query is itself the objection.

### Risks — added

- **The locked set grows.** Every positional mint gets copy written for it, so the population
  needing the content pass increases while the ranking stays unfixed. That is the argument for
  doing phase 2 before phase 3, and it has a clock on it.

---

## OWNER DECISIONS LANDED 2026-08-25 (all five) — phasing now fixed

| # | answer | consequence for this plan |
|---|---|---|
| 1 | tool "can disappear everywhere" | retirement authorised, **sequenced last** — 91 refs + footer + 3 listings |
| 2 | yes, change the numbers | **DONE + verified**; `SQL_2026-08-25_demote_password_entropy_nav_order.sql` |
| 3 | yes, build the lever | candidate 1 + candidate 4, RFC_022 owed, council before/with the commit |
| 4 | "whatever you suggest" | re-resolve then verify at served bytes; never by work-item status |
| 5 | re-scope the commission | ~20 label-locked fields by query, not 16 sites |

**Decision recorded, with its reason:** the demotion value is **900, not 200**. At 200 it ties with
the sites' other tools and the tiebreak is alphabetical on `name` — `password-entropy` precedes
every `tool-*`, so it would have kept winning on two of three sites. **A demotion that joins the
pack is not a demotion.** Guarded in the SQL by an abort-unless-exactly-three check.

**Decision recorded, with its reason:** retirement is **not** first, despite being decision 1 and
fully authorised. Deleting the page ahead of the copy rewrite strands ~91 references and leaves ~20
buttons naming a tool they no longer point at — `bugs_closed/299`'s exact defect, manufactured by
our own repair. Authorisation is not a sequence.

**New open question for the owner, raised not assumed:** the library component
`tool-password-entropy` is `is_active = true`, so it can still be handed to a *new* site. "Disappear
everywhere" may or may not extend to that switch; it is separate from the three pages and has been
left alone.
