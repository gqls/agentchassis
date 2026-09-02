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

---

## ADDENDUM 2026-09-02 16:2xZ — IT IS LIVE AND PROVEN. boxingonline is unblocked, and it is yours to fire.

**Live on `agent-chassis:v1.0.1354`**, probed at the binary (the startup provenance line
had scrolled) with a removed-string control: the deleted emitter text
`This chrome component's template could not be executed` is **ABSENT**, both new literals
**PRESENT**, nonsense absent.

**Proven end to end on the other casualty**, so you are not the guinea pig.
garden-tools.uk had the identical defect and its footer `rendered_html` had been **NULL
since 2026-08-23**. A `rerender-chrome` run stored it at 16:21:32Z — 2,427 bytes,
`digest_ok=true` — with its offending em-dash label **intact**:
`How We Assess Garden Tools — Our Methodology | Garden Tools UK`.

**What is left is one dispatch on boxingonline, and I have deliberately not made it.**
Your served footer is the 16:05 hand-patch and it is currently the only definition of that
footer, so a re-render **replaces** it. On a paid site mid-delivery that is your call, not
mine.

When you are ready, `rerender-chrome` is the surgical one — chrome only, no page
reassembly, no deploy, so it writes `site_components.rendered_html` and nothing else:

```bash
. scripts/kafka-publish-lib.sh    # assert the receipt; do not hand-roll kcat
# agent_type "rerender-chrome", input_data {site_id, domain} — garden-tools shape,
# full worked call in bugfix_423_chrome_utf8/HANDOFF_2026-09-02_continue_here.md §2
```
site_id for boxingonline.com is `d2aa5206-73bc-4707-a69c-2702c1eb9152`.

⚠ Expect **25–36 minutes** of queue latency and do **not** re-fire on a missing
orchestration row — that is the documented signature of ordinary latency, and a duplicate
costs a whole round.

Then your own pre-delivery probe still decides it: `rendered.footer=true`, the row's
`updated_at` moves, `rendered_html_digest = md5(rendered_html)`, **and the served footer
still carries NO contact block** (empty `sites.email`, gated at
`component_library.go:1988`). `bugs_open/423` closes on that check.
