Opened 2026-07-19 (session "leopardess3") from an owner review of leopardessconsulting.co.uk:
four buttons he couldn't make sense of — *Start Ranking Free*, *See How It Works*, *Start the
Guide*, *Visit the Tool*. He was right on both counts: they are broken, and they are also
mislabelled, which is why they made no sense. **Diagnosis and plan are complete; nothing is
fixed.** That split is deliberate — the owner scoped the fix to a later thread.

Entry points: `PLAN_2026-07-19_cta_link_integrity.md` (defect classes A–H, phasing),
`NOTES_cta_link_integrity.md` (the evidence, including my own corrections),
`RUNBOOK_cta_link_integrity.md` (every query with its gotcha),
`SUMMARY_2026-07-19_cta_link_integrity.md` (the plain-prose owner read-out).
Bug: `bugs_open/023`. Pattern: 016b §9, indexed §10.

**What the four buttons turned out to be.** Four buttons, four *different* mechanisms, each
defeating a different check — which is the whole reason none was caught. One renders
`href=""` (a warning in `validate_page_content`, never a blocker). One renders
`href="#guide-start"` where no such id exists (fragment scope is skipped by every check, and
nothing anywhere resolves a fragment against a page's ids). One points at a hostname that
does not resolve (external scope is skipped by every check, and there are zero HTTP
reachability checks in `platform/`). The last carries a frozen `source:static` label
belonging to a *different tool* — a Bayesian ranker — served by
`bayesian-ranking-hero-tool_pre_037`, a `_pre_037` row that is the sole live component for
its function.

Underneath all four: a button's label and its URL are unrelated schema fields, and nothing
expresses "a label implies a destination". Static labels re-apply their fallback every render
(`plan_sections_action.go:1210-1218`), bypassing `required`/`on_missing` entirely, so the
words always appear whether or not there is anywhere to go.

**Two findings that outrank the buttons.** First, it is fleet-wide and measured, not
estimated: 51 dead or suspect controls across 7 of 11 sites, and **75 of 89 URL-bound CTA
anchors in the component library are ungated (~84%)** — an agreed, documented invariant
(LNK-005, "an unresolvable destination renders nothing") that almost nothing enforces.
Second, and sharper: **one of the four was correctly detected on 2026-07-17**, right
component and right page, two days before the owner clicked it — and filed at
`needs_human_review`, which triage never promotes and no handler consumes. A grep of
`platform/` for `unresolved_cta` / `cta_names_unknown_destination` / `dead_control` returns
emission sites only, zero consumers; 34 sit open on this site. That is a *delivery* gap, and
it is a different fix from the detection gap. Adding checks without building the handler
would convert a visible problem into a larger invisible one — which is why I recommended
against running the experience loop on this yet, despite the owner asking about it.

**A correction worth carrying forward.** I first wrote that the fabricated hostname
`leopardess.contactforsales.com` was "assembled from two real domains in the owner's estate",
implying the model knew what he owns. The owner challenged it — he thought his owning
`leopardessconsulting.com` was coincidence. He was right, and checking produced a far better
mechanism: the `.com` is simply the obvious variant of the site's own name, while
`leopardess.contactforsales.com` is a **transform of the real contact email in the site's own
identity spec** (`leopardess@contactforsales.com`, `@`→`.`). The parts were true and
in-context; only the recombination was invented. That turns an unanswerable question ("is
this hostname plausible?") into a one-line deterministic check needing no network call — now
plan step P1.5 — and it generalises: **six sites carry `<label>@contactforsales.com` in their
current identity spec**, so any of them can produce the same fabrication. The general lesson
went into 016b §9: when you catch a fabrication, work out which real in-context tokens it
recombined, because that rule is usually cheap to test for.

**Owner-approved alongside it (P4.1):** 301 `leopardessconsulting.com` → `.co.uk`, path
preserved. It fixes one of the four buttons immediately and is independent of all code work.
The trap is flagged in every doc here: it makes a *fabricated* URL start resolving, which is
not the defect fixed — the field still invents on the next build, and the other page's button
stays dead.

**Git note (commit-per-task).** Three narrow pathspec commits: `db9a4259b` (bug + plan),
`47a86c61b` (016b), `9b5b117bb` (the correction + redirect). A lesson for the next thread:
`9b5b117bb` used a *directory* pathspec for this folder and swept in a stray file — a raw
terminal capture of my own chat reply, carrying the very claim I had just corrected. Nothing
was lost and nothing of anyone else's was taken, but it is the hazard CLAUDE.md warns about,
one level down: **a directory pathspec is not a narrow pathspec.** This file is that stray,
rewritten to match the convention the other workstream dirs use.

**Where to start if you are picking this up.** Read the plan's §3 first. The highest-leverage
change is not the obvious one: replace the hardcoded six-entry `ctaFieldNames` map with CTA
pairs derived from `input_schema`. Four migrations (091, 096, 097b, 098) have already
hand-patched that map with the same lesson, and until the label/url pairing exists as data,
the check everyone wants cannot be written at all.
