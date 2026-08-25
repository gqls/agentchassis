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
