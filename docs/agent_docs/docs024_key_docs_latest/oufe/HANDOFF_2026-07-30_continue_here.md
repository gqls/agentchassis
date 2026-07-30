# HANDOFF — oufe lane + the checker-layer queue (written 2026-07-30, work done 2026-07-29)

> **⚠ PARTIALLY SUPERSEDED 2026-07-30 (later the same day) — read this box before §3 or §7.**
> Current read-out: **`SUMMARY_2026-07-30_oufe.md`**; running log: the `NOTES_oufe.md` tail.
>
> - **§7's recommendation is DONE.** Option (1), `149` C1+C3, plus (3) B4 — commit
>   `f61dce806`, image `v1.0.1208` built and marker-verified, **not rolled**.
> - **§3's "blocked on reading `144` first" is STALE.** 144 closed and live on
>   v1.0.1203, and `e9de64d99` had already corrected the citation: the workflow
>   validator checks SHAPE, never page CONTENT, so nested-step validation never had
>   any bearing on the missing claims gate.
> - **§3's C1 target is WRONG, and §3's C3 figures were false when written.**
>   `page-content-writer` persists nothing (it is *called* by four of the six agents
>   that do), so the gate went to the persistence seam instead; and
>   `claims_unverified` was **not** zero — the detector had fired automatically
>   2m28s before `149` was committed. Both corrections are in `bugs_open/149`'s
>   banners, with the evidence.
> - **§5's oufe site figures are still NOT re-verified** — that part of this file
>   stands exactly as written, including its warning.
> - **§6 (build the second tool) is untouched and is now the lane's top item**, with
>   its blocker cleared for the second time.

**Supersedes `HANDOFF_2026-07-28_continue_here.md` as the cold-start pointer for this
lane.** That file is still worth reading for the pre-07-29 history; everything it says
about "next steps" is superseded here.

This session started as oufe work ("build the second tool") and became mostly a
fleet-wide investigation. Both threads are live and both are written up. Read §1, then
whichever of §3 / §4 you are picking up.

---

## 0. FIRST — two things that will bite you in the first five minutes

1. **The cluster token is EXPIRED as of this writing.** Every `kubectl` call returns
   `You must be logged in to the server (Unauthorized)`. That is the documented
   3-day kubeconfig expiry; the owner refreshes it. **Nothing in §3/§4 can be
   re-verified until it is back.** Every live figure in this file is stamped with the
   date it was measured — treat any of them older than a day as needing a re-run, and
   do not carry one forward into a new document unchecked.
2. **HEAD has moved ~82 commits since my last one** (many sessions, same tree). My
   three commits are `67c420a96`, `b15b1456f`, `c0903eb66`. Re-run `git status`
   before acting on anything; the tree is shared and mutable.

---

## 1. What I actually did, in one paragraph

I set out to build oufe's second tool. The first honest question — *where would a
reader find it?* — turned up that **nothing linked the first tool**, live since 07-28.
Fixing that led to a fleet census (11 unreachable tool pages across 5 sites), then to
the reason they were unreachable **at creation**, which turned out to be three
interlocking code defects and a repair route that structurally cannot repair. The
owner then asked for the whole checker layer to be listed as a workable queue, with
one named requirement (copy written off a discovery check must go through claims
checking). That is `bugs_open/149`. **The owner then overturned the headline of my own
Group A within hours, correctly.** Both corrections are recorded where the claims were
made. The second tool is still not built — deliberately.

---

## 2. The two corrections I was given, because they matter more than the findings

Read these before you write anything durable in this area. They are a matched pair and
I made both in one afternoon, in opposite directions.

**(a) "The cause is cadence, not code."** — wrong. I found the orphan detector sits in
a discovery agent and concluded the problem was that nobody ran it. Scheduling the
agent would have **detected all 11 and repaired none**. The cause was creation-time
code. What caught it: the owner asking *why were they unreachable when they were
created?* — a question I had not asked. I had explained why nobody **noticed** and
presented it as why it **happened**.

**(b) "These handlers have never repaired anything."** — also wrong, and the owner's
words are the rule worth keeping: *"the lack of evidence of these tools working is not
evidence that they don't work. they may not have run often."* `claimed_by` was NULL on
**all 37** rows: the handlers had never been offered one. I had read 27 rows in a
**terminal** status as a three-month backlog.

The pair is now in `016b §9` as two patterns explicitly cross-referenced —
*"a detected defect whose handler cannot act on it"* and its counterweight
*"a zero from an UNEXERCISED path looks identical to a zero from a FAILING one"*.
**Applying the first without the second produces the second error**, which is exactly
what happened to me, four hours apart.

The practical residue, and the thing to carry into any new work here:
**label every finding by the evidence class it rests on.**
- **MECHANISM** — you read the code path, it cannot do the thing, the artefact agrees.
  Survives "it hasn't run much". Act on these.
- **NEVER RAN** — a zero meaning unexercised. Good for prioritising cadence, worthless
  as a judgement on code. **Never let one motivate a rewrite.**

And the trap that hid it: one properly proven branch sat in the same paragraph as two
unexercised ones and **lent them its credibility**. Split them.

Both are logged in `WRONG_CALLS.md` (2026-07-29) with the cheap check that would have
caught each — for (b) it is one column, `claimed_by IS NOT NULL`.

---

## 3. Thread A — the checker layer: `bugs_open/149`

**Status: filed, ordered, unowned. Nothing has been fixed.** This is the one the owner
said "we'll work through". The file is the queue; it is ordered and each item carries
the query that measured it. Below is only what you need to decide where to start.

### The owner's named requirement is C1, and it is top of the queue

> *"The rewrite for `run_discovery_checks` must write copy that follows the claims
> checking like everything else."*

**Measured 2026-07-29:** of the **22 handler agents** discovery checks route work to,
**2** run `validate_page_content` (`page-build-handler`, `tool-recreation-handler`).
**`page-content-writer` has no validation step at all** — its whole workflow is
research-agent → `execute_llm_prompt` (gemini-pro-latest, 8000 tok) →
`render_component` → compile → complete. `internal-linker` is the same shape
(`plan_links`).

Three things to know before starting it:
- The LLM write is **inside a `loop` sub_workflow**, and per `bugs_open/144`
  sub_workflow steps are validated by *nothing*. **144 blocks trusting a fix placed
  there** — a correct step could be silently unwired and everything would still look
  green. Read 144 first.
- `validate_page_content`'s claims keys **default to `true`**
  (`validate_page_content.go:223,231,247`), so the work is mostly *adding the step*.
  **Set them explicitly anyway.** A silently-inherited default is precisely what
  caused the misrouting in Group A (see A4), and this queue should not repeat it.
- `report-builder` is the **only** agent that sets `check_claims: false`. Find out
  whether that is deliberate before switching it on — a report restating figures from
  a cited upstream may legitimately need different handling.

**Scope honesty:** `run_discovery_checks` itself writes no copy. The rewrite the owner
is asking for lives at the **handler seam**, and the honest scope is "every
content-emitting handler", not one action. Say so when scoping it rather than
delivering a narrower thing under the same name.

**Pair it with C3.** The detector for this (`unverified_claims`) is HITL-terminal *by
design* (correct — `check_unverified_claims.go:145`), but it lives in
`quality-discovery-agent`: **7 work items in its entire history, nothing since
07-17**, and `claims_unverified` has **zero rows fleet-wide**. That is **NEVER RAN,
not broken** — do not rewrite the check on that basis. The cheap first move is to seat
it in an agent that actually runs and see what it says; `claimscan` already finds real
instances on live sites (`bugs_open/147`), so a working detector on a working cadence
should too. **If it then stays silent, that is a finding.**

**So the exposure is the pairing:** no write-time gate *and* no working backstop.
That is why C1+C3 sit above the routing work, which is only a discoverability problem.

### The rest, in the file's order

- **B4** — an unregistered *or erroring* check name is a `WARN` + `continue`
  (`discovery_checks.go:141-146,152`). A config typo silently shrinks the run and it
  still reports success. Small, self-contained, and **every Group B number is
  untrustworthy until it lands** — that is why it is second.
- **B2** — no `nav_drift` item has **ever** been raised by a discovery agent (all 16
  from named sessions or `generic`), while the same agent kept raising other types.
  Cause `[UNMEASURED]`, three candidates named (dispatch coverage / swallowed check
  error / dedup suppression) with different fixes. **Establish which before changing
  anything.**
- **A6 → A2 → A3** — creation-time first (see §4), then stop routing `/tools/` to a
  handler that cannot act, then add the missing `orphan_tool_pages` → rebuild-listing
  route (the blog analogue already exists).
- **A4, A5** — `pages.in_header`/`in_footer` **default to `true`**, and two nav
  builders with opposite predicates. Both shared-mechanism changes wanting their own
  council round. Neither blocks the above.
- **B1** — six registered checks are in **no** agent, `validate_component_standards`
  among them (7 item types). **NEVER RAN, not broken.** Seat or delete. Expect a burst
  on first run — that is the check working, not a regression.
- **A1** — **not** a handler defect (corrected). The real item is recurrence branding:
  **20 of 24 repeat detections born `unresolved`** (terminal, non-dispatchable), 16
  distinct `item_key`s for 24 rows — the failure already pinned in
  `work_item_recurrence_test.go:20,103`. Pair with B2; may be the same mechanism from
  the other side.

### One near-miss worth inheriting

Six configured check names (`missing_*_tracker_*`, `missing_model_directory_*`) look
unregistered to a grep for literal `Name()` returns. They are registered **dynamically
per profile** at `check_directory.go:111-116`. I nearly filed them as dead config.
**A dynamic registration is invisible to a grep for string literals — enumerate via
the registry, not the source.** Cost nothing because I checked before filing; it would
have been a fabricated defect at the top of a queue the owner was about to work.

---

## 4. Thread B — `bugs_open/146`: why tool pages are unreachable at birth

**Status: oufe's own instance FIXED and verified; the fleet residual is unowned and
belongs to other lanes.** Four mechanisms, all MECHANISM-class:

1. **Both creators mark the page nav-worthy without anyone choosing it.**
   `deploy_tool_action.go:117` defaults `inHeader := true`;
   `create_tool_component_action.go:280` omits both columns and
   **`pages.in_header`/`in_footer` DEFAULT TO `true`**. So "the page carries a nav
   flag" is a schema default, not a declaration — and it is what routes these pages
   into the branch that cannot fix them.
2. **Neither creator finishes the job, in mirror-image ways.** `deploy_tool_to_site`
   sets the flags and writes **no `site_nav_items` row at all**;
   `create_tool_component` writes the nav item and never sets the flags. Neither
   re-renders chrome.
3. **`populate_nav_tables` — the only builder of nav from page flags — skips every
   `/tools/`, `/blog/`, `/guides/`, `/articles/`, `/case-studies/`, `/news/`,
   `/resources/`, `/insights/` URL by design** (`:294,339`): *"the parent listing
   represents them"*. Fleet-wide, **2 nav items point at a tool page, out of 95**.
4. **So the repair route is a closed loop.** Nav-flagged (1) ⇒ `nav_drift` ⇒
   `nav-updater` ⇒ `populate_nav_tables` ⇒ skipped (3). A `nav_drift` item raised
   **2026-07-24** is `complete`; the page still has **0 nav items, 0 chrome links**.

Plus: **two nav builders disagree** — `buildServicesHTML`
(`render_site_components_action.go:950`) includes any `in_header OR in_footer` page in
the chrome footer, so reachability depends on which builder last ran. And **a listing
page is not sufficient**: gamesdesign has `/tools/index.html` and still has 4 orphans
(two URL conventions, listing enumerates one).

**The trap that shaped the oufe fix, and that anyone touching the fleet residual must
respect:** the standard remedy is regenerate chrome, which **would have added the link
unaided** — and **would have deleted oufe's footer honesty note**, which is in no
template and no Go code and exists only in the stored artefact. Two greps established
that before I touched anything. So the page nobody could reach is fixed by the one
action that removes the disclosure saying the site can be wrong. **Check for
hand-patched chrome before regenerating anyone's.** Fixed instead with a guarded
targeted `replace()` (mig 268) whose VERIFY asserts the note survives.

**Also:** a site-wide `rerender-pages` **silently skipped the 3 `owned` pages** —
5 of 8 completed, 3 sat at `triaged` unclaimed, no error, orchestration `COMPLETED`.
Deployed individually with `049b` assemble-only. **After any chrome change, count
deployed pages, not orchestration status.**

**Note the title of 146 is wrong** and carries a correction banner — it blames 07-17 on
the wrong agent. `completeness-discovery-agent` ran to 07-25; 07-17 is
`quality-discovery-agent`. File not renamed (forward-only).

---

## 5. oufe lane state — as measured 2026-07-29, NOT re-verified (token expired)

- **8 pages live.** The tool is linked from the footer of **8/8 served pages** with the
  honesty note intact on all 8. oufe absent from the orphan census.
- **claimscan: 0 findings across 19 components.** Render audit run 5: 8 pages, **0
  firm findings**, 0 broken images, 0 overflow.
- Open work items were all `needs_human_review` / `wont_fix` — nothing auto-dispatchable.
- Earlier this session (pre-compaction): the render-audit dispatch was fixed at source
  (mig 256 wrote `initial_step` where the chassis reads `start_step`); a firm WCAG
  failure on the contact form's submit button was fixed in the `gqls/sites` repo
  (`ae0c28d51`); the platform's **first evidence-timeseries** shipped (mig 265/266,
  Thames leakage, 5 observations each with its own citation); the tool-guide intro and
  footer link shipped (267/268).

**Re-verify all of the above once the token is back**, before quoting any of it.

### Standing rails for this lane — do not relax them

- **No figure enters any brief, spec, identity or content_direction — only the
  evidence register, with a source.**
- **Never publish a figure about a real company without a source URL and capture date.**
- The grounded lane must keep its **inability** to publish.
- **Section G (liability cap) stays parked.**

---

## 6. THE NEXT TASK, and why it is still not done

**Task #6 — build the relevant-alternative tool (oufe's second tool).** Design settled
and recorded; nothing built.

- **What it does:** move the counterfactual recovery and watch which classes lose their
  veto (s.901G Conditions A and B). **Hypothetical figures only.**
- **Conventions:** model on the waterfall tool — `rw-`-prefixed CSS, consent gate,
  `color: var(--color-surface)` on `background: var(--color-primary)`.
- **Heed `bugs_open/126`:** the consent gate must be **its own first acceptance
  check**, or an automated repair loop will "fix" the tool by deleting the disclaimer.
- **It now has a footer "Explore" slot to be linked from** — that was the whole point
  of this session's detour.
- **Grounding already exists** and carries no `context_terms`, so "75%" in prose is
  registered fleet-wide on this site: `CIT-709026d97df11645` (s.901G/75%),
  `CIT-917cae2baf2a9069` (Condition A), `CIT-81d532ab22dcb359` (Condition B),
  `CIT-f729a1b60a30481` (relevant alternative), `CIT-3cd41ecf235e9df9` (cram down).

**Why I did not build it:** the lane's next-step was "add a tool", and the honest first
question turned into an afternoon. **A second unreachable tool would have doubled the
invisible surface.** That trade still looks right, and the blocker it exposed is now
cleared.

**Also open from the 07-27 owner review:** #2 more tools (in progress), #4 a more
readable layout. And the premise-branching / deepthink design decision
(`DESIGN_2026-07-28`) is at step 4 of its own first slice — *"only then design the
lane… put it through the council gate"*.

---

## 7. Choosing what to do first

Three live options; they are genuinely different kinds of work.

1. **`149` C1+C3 — the claims gate.** The only item in the queue with a
   *content-correctness* consequence rather than a discoverability one, and the owner
   named it. Blocked on reading `144` first. **This is what "we'll work through them"
   most plausibly meant.**
2. **Oufe's second tool** (Task #6). Self-contained, design settled, unblocked now.
3. **`149` B4** — make silent check-skipping loud. Small, and it makes every other
   Group B measurement trustworthy.

**My recommendation: (1), reading `bugs_open/144` before touching anything**, then (2)
as the lane's own next step. (3) is a good filler if the cluster token is still down —
it is a code change verifiable by test rather than by the fleet.

**Nothing here is dispatched, in flight, or half-applied.** No uncommitted work of
mine is outstanding; the only modified file in my area at handoff was
`bugs_open/128`, which belongs to another session.

---

## 8. Where everything is

| what | where |
|---|---|
| checker-layer queue | `bugs_open/149_HANDOFF_2026-07-29_discovery_checker_layer_defect_queue.md` |
| tool pages unreachable at birth | `bugs_open/146_HANDOFF_2026-07-29_eleven_tool_pages…md` (**read its banner**) |
| blocks C1 | `bugs_open/144` (`sub_workflow` validated by nothing) |
| the two patterns, as a pair | `016b §9` — *"a detected defect whose handler cannot act on it"* + its counterweight |
| both wrong calls, with the cheap checks | `WRONG_CALLS.md`, 2026-07-29 entries |
| the running technical log | `oufe/NOTES_oufe.md` (newest at the bottom; the last three entries are this session) |
| plain-prose history for the owner | `oufe/README_where_we_are.md` |
| last milestone read-out | `oufe/SUMMARY_2026-07-29_oufe.md` — **written before the orphan discovery, so it predates §3/§4.** A new summary is arguably due; I did not write one because the queue is filed and unstarted, and "where we are now" would mostly restate this handoff. Write it when C1 lands. |
| commits | `67c420a96` (146), `b15b1456f` (149 filed), `c0903eb66` (149 A1 corrected) |
