# NOTES — bug 252 og/lang assembly

Append-only, newest at the BOTTOM. Evidence, commands, and every misstep.

---

## 2026-08-20 (a) — pickup, and the premise check that changed the design

**Ownership swept before touching anything.** `scripts/who-owns.py 252` returns the ambiguous-number
warning and names the *disk* 252's lane (closed 08-15) — so its "OWNED" verdict is about the other
case. For this slug: last commit on the bug file is `f666408ed` (08-11, the owner's option-3
decision); no open `site_work_items` match og/lang; nothing in the `needs_diagnosis` queue overlaps.
The LMC lane's `HANDOFF_2026-08-15b_continue_here.md` §4.2 has this queued behind "verify 251 is
live", never started.

**251, the stated blocker, is already discharged.** `61abbdbd0` added `preferredPageURL`
(`rerender_single_page_action.go`), council corr `33fb41cb` = APPROVED round 1 with one advisory.
Confirmed live at the artefact: `about.html` serves `<link rel="canonical" href="…/about.html">`.
So A is unblocked and og:url has an existing helper to agree with.

**The premise check that mattered.** The bug file says an assembled page carries og:2. Rather than
inherit that, I counted the stored heads:

```sql
SELECT count(*) AS heads,
  count(*) FILTER (WHERE rendered_html LIKE '%og:url%')   AS og_url,
  count(*) FILTER (WHERE rendered_html LIKE '%property="og:title" content=""%') AS blank_ph
FROM site_components WHERE slot_name='head';
--  24 | 22 | 4
```

22 heads carry og:url. That is the opposite of "the tags are missing", so I went to the artefact:

```
$ curl -s https://ai-agent-orchestration.com/about.html | grep -iE '<html|og:|rel="canonical"'
<html lang="en">
    <meta property="og:title" content="">          <- template, blank
    <meta property="og:description" content="">    <- template, blank
    <meta property="og:type" content="website">
    <meta property="og:site_name" content="AI Agent Orchestration">
  <meta property="og:title" content="AI Agent Orchestration">     <- injectBrandHeadTags, DUPLICATE
  <meta property="og:url" content="https://ai-agent-orchestration.com/">   <- the HOMEPAGE
<link rel="canonical" href="https://ai-agent-orchestration.com/about.html">
```

Two og:title tags, and an og:url that contradicts the canonical on the same page. `git log -S'og:url'`
dates the cause to `d3f73a724` (imagery I1, after the bug was filed). **This is why the bug file's own
fix candidate would have fixed almost nothing** — it proposed adding blank placeholders to fill, and
the sites that have them already have them shadowed by a filled duplicate. Design changed to
remove-then-inject as a result (PLAN D1).

**Two of my own grep errors, same root, both caught within minutes.**

1. I grepped `'lang="en"'` across the Go tree and concluded `rerender_single_page_action.go:670` and
   `rerender_pages_actions.go:527` had already been fixed — both were cited by the bug file and
   neither appeared. **They were there all along**: the emitters build the string in Go, so the
   source reads `lang=\"en\"` (escaped) and a literal `lang="en"` grep cannot see it.
   *The check:* when grepping for HTML that Go EMITS, search the escaped form too
   (`grep -rn 'lang=\\"\|lang="'`), or search a fragment that cannot contain the quote (`<html lang`).
   Cheap and decisive. Logged in `WRONG_CALLS.md`.
2. My first two DB queries failed on column names — `site_components` has `rendered_html` and
   `slot_name`, not `html` / `component_type`. CLAUDE.md says schema first (`\d <table>`); I
   guessed twice, then read it. No damage (a failed query is loud), but it is the same habit as (1).

**A stale premise found in a comment, worth recording for whoever fixes the checkers.**
`discovery_checks/check_site_structural_validity.go` (~:55, :1029) documents at length that og:url
is excluded from `head_essentials_missing` "because the shared `<head>` cannot carry a per-page
value", and that `verify_site.py`'s `OG_PER_PAGE` exemption (`…/loanandmortgagecalculator_couk/verify_site.py:71`)
treats it as an accepted loss. Both premises are what this lane is removing. They will become false
when this ships, and neither is a code path that will fail loudly — added to the close-out list.

**Two facts about the fleet's own machinery that shaped the rollout, both confirmed in source:**
- `renderAndStoreSiteComponent`'s idempotence exit skips any head whose `rendered_html` is non-empty
  and whose `build_status <> 'pending'`. So **a Go change regenerates no stored head, ever.**
- `datahelpers/chrome_render_inputs.go` hashes the template and site_specs **by value**, and Go code
  is not an input. Hence: the og fix must work without a chrome rebuild (it does — it strips at
  assembly), and the lang fix gets its fleet propagation for free from the migrations.

## 2026-08-20 (b) — 090 dispatched, and a planning assumption that expired

`090_TRIGGER_needs_diagnosis_v1.sh` accepted the symptom (mechanism-only, no counts asserted, three
tables/symbols named). Intake corr `855be313-c96a-47d6-a92d-cdfa53de6b03`; the dispatch loop claimed
it and minted **run corr `af31ec22-5662-4798-91b9-b12132ebca70`** — that second one is the key the
artefacts are written under. Filed even though the mechanism is read directly from source, because
the durable claim here is cross-cutting (two producers, a shared artefact, a fingerprint that cannot
see the code) and CLAUDE.md's 2026-07-31 ruling puts the burden on the asserter.

> **CORRECTED, same session:** my plan recorded that the working tree does not compile and that both
> of my edit targets were held dirty by the bug-260 session, with a whole mitigation built around it
> (build from a clean `git archive HEAD` extraction; sequence the emitter edit behind a re-check).
> Re-checked at execution start: HEAD has moved to `2b3e61e8e`, that session has **committed**,
> `go build ./platform/orchestration/actions/` exits 0, and **both target files are clean**. The
> mitigation is no longer needed for correctness — the clean-extraction build is now belt-and-braces,
> not a workaround. Keeping the emitter edit in its own commit anyway: the file is one another lane
> touches often, and a one-line commit is the cheapest thing to cherry-pick around.
>
> *The check that caught it:* re-run `git log -1` + `git status --porcelain` + a build at the start of
> every phase, never trust the snapshot the plan was written against. On this tree a planning
> assumption about another session's tree state has a half-life measured in minutes.

**Migration numbering: 497 and 498 are BOTH already taken twice** (`497_escalation_owners_map…` and
`497_…_ROLLBACK`; `498_escalation_literal_markdown…` AND `498_schedule_meta_description_backfiller…`).
So the directory already contains a same-number collision from two lanes, and the highest number is
501. This lane takes **502** and **503**. Read the directory fresh before authoring — do not derive
the next number from a doc.

**Council scope widened under me, in my favour.** CLAUDE.md changed on disk during this session:
appliable migrations (`docs/agent_docs/sql_for_agents/NNN_name.sql`) are now IN council scope
(`bugs_open/314`), with the scope single-sourced in `scripts/council-scope.sh` and `DRY_RUN=1` on the
097 trigger available to test admission for free. So both migrations go through the gate with the Go,
in one submission, rather than the Go alone.
