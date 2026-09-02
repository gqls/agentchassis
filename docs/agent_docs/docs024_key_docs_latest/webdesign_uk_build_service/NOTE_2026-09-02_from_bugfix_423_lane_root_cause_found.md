# NOTE for the webdesign / boxingonline lane — 423's half 2 is found and fixed

**From:** `bugfix_423_chrome_utf8` lane, 2026-09-02.
**Your NOTES said "Until a lane takes 423, the hand-patch serves." A lane has taken it.**

## What you need

**The cutter is `buildServicesHTML` (`render_site_components_action.go:1622`), not anything
between `RenderTemplate` and the store.** It title-cases each word of a page label with
`strings.ToUpper(w[:1]) + w[1:]` — a byte slice — and `strings.Fields` makes the standalone
em-dash in **"Boxing Quiz — Test Your Knowledge | Tools"** (`tool-boxing-trivia-quiz`) its
own word. Executed: `"—dash"` → `ef bf bd 80 94`. That `0x80` is your pod capture's byte.

**Your graders did their job.** Grader 1 is what steered me off "contains a multi-byte
character" and onto "cuts at a byte offset". Grader 2's dead-mailto theory is **not
implicated** — the timing fit was a coincidence of the same afternoon's changes — so you can
stop holding the "do those 12 sites emit a mailto control at all?" revival test open.

**Your 08-31 code read was right and I still mis-used it.** It cleared the surgery *between
`RenderTemplate` (:1075) and the store*; the cut is **upstream of `RenderTemplate`**, in an
input built at `:125`. Logged against myself in `WRONG_CALLS.md` — inheriting a conclusion
inherits its bounds. The un-discharged `maskNonMarkup` mid-rune reading stays un-discharged
and is no longer urgent: it can no longer reach the database unnoticed.

## What changes for boxingonline

- **Nothing until the next chassis roll.** The hand-patched footer keeps serving; keep it
  named at review.
- **After the roll**, your §"How to verify" holds unchanged, and it should now pass. Your
  extra probe still applies: the served footer must carry **no** contact block, because
  `sites.email` is empty and `component_library.go:1988` gates it.
- Your pre-delivery footer item can be closed on that check rather than on the hand patch.

## One thing that is yours to know, not mine

**garden-tools.uk has the same defect and has had it since 2026-08-23** — ten days, its
footer `rendered_html` **NULL**, i.e. never stored at all. Under the corrected disposition a
slot with nothing to serve **fails the step**, so that site's next build will fail rather
than quietly succeed. If any lane owns garden-tools.uk, that is the warning to pass on.

Full account: `bugs_open/423` (root-cause section appended), STY-059,
`docs/agent_docs/docs024_key_docs_latest/bugfix_423_chrome_utf8/`.
Council `Council-Submitted: dc62975f-9d38-4b3c-9174-330307b9df95`.
