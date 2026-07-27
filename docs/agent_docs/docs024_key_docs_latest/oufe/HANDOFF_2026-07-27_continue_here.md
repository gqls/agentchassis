# HANDOFF — oufe.com / oxenunity.com, 2026-07-27 evening

**Cold-start entry point.** Read this, then `PLAN_2026-07-25_oufe.md` §1–2 for what
the site is and why, and `DECISIONS_2026-07-26_oufe.md` for what the owner still
has to choose. `NOTES_oufe.md` is the running log, newest at the bottom, and it
carries the missteps — those are the part you cannot rederive.

Supersedes `HANDOFF_RESUME_oufe.md`, which is now stale.

---

## 1. What is live right now

Verified by fetching each URL, not from the database.

| url | state |
|---|---|
| `https://oxenunity.com/` | 200 — hand-authored one-pager, no entity claims |
| `https://oufe.com/` | 200 |
| `https://oufe.com/about.html` | 200 |
| `https://oufe.com/cases/index.html` | 200 |
| `https://oufe.com/cases/thames-water.html` | 200 — **the case** |
| `https://oufe.com/tools/tool-recovery-waterfall.html` | 200 — **the tool** |
| `https://oufe.com/contact.html` | 200 (`build_status=needs_rebuild`, cosmetic) |

Site id `a0d7f1ae-f37e-4ea5-b30c-9012d1d14f39`. Evidence register: **22
quote-verified facts, 28 banned patterns**. Zero broken internal links (full
crawl). All copy in the house voice; one em dash remains site-wide and it is a
placeholder glyph in the tool's verdict field, deliberately.

**oufe.com needs no infrastructure work.** It is already on our Cloudflare
nameservers with the portfolio-sites Worker route bound, so content serves the
moment it lands in B2.

---

## 2. What the owner asked for last, and has NOT been done

He reviewed the site and asked for five things. **One is done, four are open.**

1. ~~Fix the unreadable Thames Water link~~ — **DONE**, see §4.
2. **More tools.** There is one. He wants several.
3. **Closer to the original spec: more explanation, helpful guides at each point.**
4. **A more readable layout** — different components, "less like a heavy text book
   and more like a readable important doc".
5. **Infographics for the harder concepts, and graphs where appropriate.**

> **Before starting 5, know this: there is no chart renderer.** The doctrine
> (charts are code-rendered from real figures, never drawn by a diffusion model)
> is real and enforced. The implementation is a gap. `go-echarts` is not in
> `go.mod`; the only chart code in the estate is
> `platform/orchestration/actions/report_charts.go`, a dependency-free inline-SVG
> helper written for one report page. So a graph means extending that, or drawing
> client-side inside a tool. Do not promise a chart before reading that file.

Items 3 and 4 interact with **`bugs_open/107` (the skeleton one — see §6 on the
number collision)**: every site gets hero › cards › call-to-action regardless of
subject, which is *why* it reads like a textbook. Owner has parked 107 itself as
"not a blocker for now", but a genuinely different layout for oufe either works
around that skeleton or fixes it.

---

## 3. Decisions — one answered, one still open

Both in `DECISIONS_2026-07-26_oufe.md`.

**O1 — the audience. ANSWERED 2026-07-27.** The owner chose "anyone learning how
this works": students, trainees, early-career analysts, and practitioners in
adjacent seats. Not narrowed to students (which would have made the site free
indefinitely), not narrowed to mid-market professionals. Applied to
`identity.target_audience`, to a new `audience` aspect (`audience.v1` — the site
had none), and to the mission brief. **Content is no longer blocked on this.**

**O6 — the radar ordering.** He ruled "direction 3 first, lowest risk"; this
workstream argued it is the highest risk available and built the dossier-plus-tool
path instead. **We proceeded on our own reading.** Needs ratifying or reversing
rather than being left implicit.

---

## 4. Framework changes shipped from this workstream

These outlive the site. All are live.

| what | where |
|---|---|
| House voice is the writers' **default** | mig 228, 230, then `SWEEP_house_voice_coverage.py` (3 of 26 → 7 of 15 real candidates) |
| Compliance council seat catches **overclaimed reliability** + illustration-not-authority | mig 223, corrected by 227, mirrored via `099_SYNC_gate_roster.py` |
| `grounded-explainer` — high-attention content lane that **cannot publish** | mig 224, 225, 229, 230 |
| oufe armed with overclaim patterns | mig 226 |
| Concept register: `covers-through:` stamps + coverage sensor | `docs026_concept_register/102_CHECK_register_coverage.py` |
| oufe WCAG link fix | `gqls/sites` commit `17c0d060e` |

**The house voice source of truth** is
`travelling_docs/pitch_pdf_source/REVERSE_ENGINEERED_STYLE_PROMPT_v3.md`, built by
the owner on 2026-07-17 over three rounds of critiquing the prompt's own output.
Rule 3 is the one he names most: say what a thing **is** before saying what it
isn't. It had never been wired to any writer before this session.

---

## 5. Bugs filed here (read before touching adjacent code)

| bug | one line |
|---|---|
| `094` | `049b_deploy_single_page.sh`'s `section_data_resolved` branch cannot work — omits `page_name` |
| `096` | a long council run head-of-line blocks every other generic dispatch. **Fix approved, committed, NOT applied** — see §7 |
| `097` | CTA integrity misses `content_data.cards[*].link_url`; a second instance on robot-hands was added by another session |
| `104` | `banned_claims` is per-site only, 10 of 15 sites unarmed incl. vetcomparison and idea.uk |
| `105` | `EvidenceFact.Kind` declared and read nowhere — it is the slot a promise needs |
| `106` | concept register froze 2026-07-13; **51 of 76 workstreams postdate it** |
| `107` (skeleton) | every site gets the same brochure layout. **Parked by the owner** |
| `122` | **generated CSS fails WCAG AA on four live sites** — see §4 above and §8 |

`095` (wrong `slot_name` renders nothing, reports COMPLETED) was **closed by
another session** — it is in `bugs_closed/` now. Its lesson still applies.

---

## 6. Traps that will cost you time

**`107` is an ambiguous number.** Two unrelated bugs share it:
`107_..._every_site_gets_the_same_homepage_skeleton.md` (mine) and
`107_..._gemini_client_starves_thinking_models_of_output_budget.md` (another
session's). **Refer to either by slug, never by number.** This is the documented
trap in `bugs_closed/README.md` and it just happened again.

**Rendering an owned page.** Every tool page is `rebuild_policy='owned'`, and
`save_page_sections` hard-refuses those, so `TRIGGER_rerender_page.sh`'s default
`section_data_resolved` can never render a tool. Use assemble-only after
populating `rendered_html` from `html_template`. Full recipe in `RUNBOOK_oufe.md`
§8b.

**`slot_name` must equal the component's function name** *and* appear in
`pages.sections`. `'main'` matches nothing, renders nothing, and reports
`COMPLETED | complete_skipped`.

**`curl 000` is not a 404.** It is "no response obtained" and says nothing about
the resource. Use `--retry 3 --retry-all-errors`. `$?` after a pipe is the last
command's status, not the script's. Both cost me a false report today.

**A contrast check must resolve the element's actual background.** My first audit
compared everything against the page background and flagged the header as failing
at 1.03; it sits on a white bar and scores 17.40. "Fixing" it would have dropped
it to 2.6.

**The two content writers keep their prompts at different paths.** One UPDATE
covers one of them. Verify by type, never by rowcount.

---

## 7. The one approved thing not yet applied

`bugs_open/096` — give council traffic its own dispatch lane. Owner approved it.
Manifest committed as `e88852825`. **Not applied**, for two reasons both recorded
in the bug:

- **Never `kubectl apply -k` the overlay.** Sixteen files under `deployments/` are
  modified and uncommitted by other sessions. Use:
  ```bash
  kubectl -n ai-persona-system set env deploy/agent-chassis \
    EXTRA_REQUEST_TOPICS=system.agent.scheduled.requests,system.agent.council-gate.requests
  ```
  Current value for rollback: `system.agent.scheduled.requests`.
- **A roll costs an in-flight run.** `Agent.Shutdown()` waits 30s then closes
  regardless, and a council step is a single LLM call running for minutes. I
  watched for ~90 minutes and councils were in flight on *every* check — the
  window never came, which is the bug describing itself.

**Rollout order is load-bearing** (`agentbase/agent.go:429-432`): roll, confirm
the new topic has a consumer, and only then point
`097_TRIGGER_council_review_v1.sh` at it. A producer aimed at an unconsumed topic
piles messages where nothing runs them.

Owner's options are written up at the foot of 096. Recommendation: an overnight
window.

---

## 8. Suggested next moves, in order

1. ~~Get an answer to O1~~ — **done 2026-07-27**, see §3. Content is unblocked.
2. **Apply the council lane** at a quiet hour, then run the **Tier-4 browser
   acceptance** for the waterfall tool — still owed. Markup being present is not
   evidence a tool works; that claim belongs to the headless run.
3. **Fix the three other WCAG-failing sites** (`122`): dartsonline, robot-hands,
   vonc. The oufe fix is the worked example; the measurement script is in the bug.
4. **Then the owner's layout/tools/guides/infographics work.** Read
   `report_charts.go` first, and decide whether to work around the skeleton or fix
   it.

---

## 9. Rails that must not be relaxed

- **No figure enters a brief, spec, identity or content_direction.** Only the
  evidence register, with a source. A number in a spec is a *given* and beats
  every writer-side rule — that is how invented figures were written back over
  correct ones on another site with all guards live.
- **A clean claims report on this site means little.** The deterministic number
  scan has no finance vocabulary and excludes currency outright.
- **Never publish a figure about a real company without a source URL and a capture
  date.** The vetcomparison rails, inherited whole.
- **The grounded lane must keep its inability to publish.** It ends at
  `needs_human_review` and there is no flag to change that. An automated content
  lane that *can* publish will eventually publish something wrong unattended.

---

## 10. The thing worth carrying forward

Four times this week the platform already had what I was about to build, or had
already answered the question I was asking: the banned-claim scanner could always
catch reliability overclaims; `evidence_base.Kind` was the slot for a promise;
`format_research_content` already capped source size; and the fleet-share decision
had been filed months earlier with a trigger that quietly lapsed.

Each was invisible for the same reason — **it was not running**. Code search finds
what executes and leaves no trace of what is merely available.

The companion lesson, and the one the owner named: every misstep had a designed
interceptor — the diagnosis loop for a structural claim, the reuse seat for a
proposal that duplicates existing machinery — and all of them are opt-in, chosen
by the person about to make the mistake. Confidence is the gate, and confidence is
what is broken at exactly that moment.

Both are written up in `docs024_key_docs_latest/WRONG_CALLS.md`, 2026-07-26 and
2026-07-27 entries, with the cheap checks that would have fired.
