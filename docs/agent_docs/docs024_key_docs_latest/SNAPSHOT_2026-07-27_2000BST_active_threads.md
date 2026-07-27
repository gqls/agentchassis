# Snapshot — active threads and where we are, 2026-07-27 ~20:00 BST (19:00 UTC)

**Why this file exists.** The owner expected the machine to crash (suspected memory
leak) and asked for the state to be written down before it did. A crash does not
destroy the working tree or the repo — it destroys *conversation context*, i.e. who
was doing what and which of the uncommitted edits belong to whom. That is what this
file preserves.

**This is a DELTA, not a replacement.** The body of "where we are" is
`OPEN_THREADS_RESTART_LIST.md`, rewritten today at **16:05 UTC** (`71ad14c67`). It is
622 lines, per-thread, with a cold-start doc named for each. **Read that file first.**
Everything below is what changed in the ~3 hours after it was written, plus the
volatile state it could not contain.

---

## 1. Standing facts, checked just now

| fact | value | how checked |
|---|---|---|
| Live chassis image | **`v1.0.1175`** | `kubectl -n ai-persona-system get deploy agent-chassis -o jsonpath='{...image}'` → `docker.io/aqls/agent-chassis:v1.0.1175` |
| `IMAGE_TAG` in makefile | `v1.0.1175` (line 17) | `grep -n IMAGE_TAG makefile` |
| Open bug files | **44** on disk (was 40 at 16:05) | `ls bugs_open/*.md \| wc -l` |
| Commits today | **240** | `git log --since="2026-07-27 00:00" --oneline \| wc -l` |
| Branch | `086_experience_loop` | `git branch --show-current` |
| Uncommitted paths | **33** | `git status --porcelain \| wc -l` |

> The restart list's three standing facts said **v1.0.1174**, rolled 15:11 UTC. That
> is now superseded — a second roll went out at **19:02 BST** (`22c6df962`) carrying
> the Gemini spawn-env fix. **Clock trap still applies:** this machine is BST (UTC+1);
> `git log` prints BST, `kubectl` prints UTC.

---

## 2. The single biggest change: the restart list's §0 emergency is CLOSED

The restart list opened with "**the page writer is armed to fail across the whole
estate tonight**" — `bugs_open/112`, spawned pods getting no `GEMINI_API_KEY`, with
the first predicted casualty the ~20:25 UTC `model-directory-publish`, and an owner
decision (§C item 1) about whether to revert the writer to Claude.

**That decision is no longer needed.** Between 18:55 and 19:41 BST the fix was
written, shipped and verified:

- `b3f19ac96` — give spawned agent pods `GEMINI_API_KEY`, **in both spawners** (the
  second spawner was found during the fix, not before it).
- `22c6df962` — chassis **v1.0.1175** built and rolled, live.
- `3c1a59151` — **verify step 1 PASSED**: the key reaches a spawned pod, *with a live
  control*. (Note the discipline — a positive control, not just a green path.)
- `6b5509ee3` — round 2 refactor: one provider-key allow-list read by both spawners,
  answering the council.
- `bb71f3dde` — records why fix candidate 2 was declined.

**Consequence for the owner:** §C item 1 of the restart list ("Gemini writer revert")
can be struck. The cost argument in it still stands on its own merits — Gemini spends
~1,815 thinking tokens per section against Claude's zero — and there is now
**four-model bake-off data** to decide on (`a51914ecf`): Grok amplifies the dryness
complaint, Fable fixes it, and one prompt clause nearly closes the gap for free.
That is a content-quality call, not an outage.

---

## 3. Threads that were actively committing in the last two hours

This is the real answer to "what threads are active". Fifteen lanes committed between
17:54 and 19:42 BST. Grouped by lane, newest first within each.

**Brochure component library / fundamentallyai.com — the busiest lane today.**
Owner reported the site was "nothing like the brief" on mobile. The thread built a
*render audit* tool (`881b4b5cf` — "measure the page a visitor actually gets, not the
sources it was built from"), which found what fifty source-side checks could not.
Shipped a readable palette and pointed the site at 21 generated images it had never
referenced (`f56e78ea5`); three rounds on the CSS renderer (`3096a55a6`, `3ac63dc5c`,
`18d5bd714`). Filed **113**, **114**, `features_open/026`. Wrote a summary
(`541395988` — "the site was unreadable, not undesigned").
**Owed:** the corrected stylesheet is staged and **pending one publish command**
(`1cde867d8`) — memory records this as an owner step, `deploy_stylesheet_direct.sh`.
Also flagged: the homepage has **no carousel at all**, which is a brief requirement,
not a taste question (`9bc0f7883`).

**relojistas.com / traffic_probe.** Owner ruled: no contact route. Removed and
**verified live on 18 of 18 pages** (`738379817`). Favicon built from the logo's gear
glyph. Three wrong mechanisms were chased before the right one, and that chase is
written up (`0c026f8ff`) plus a runbook on changing site chrome and actually seeing it
(`26c77f7fa`). Filed **117** and **118** out of it. For-sale block configured but
parked on language.

**webdesign.co.uk.** Found **every content link on the home page was a 404**
(`b274c4ed5`), and filed **116** — *link integrity checks have never run on any site*.
Corrected two inherited warnings that had expired (`230a3653c`). Cold-start handoff
written (`1ee0e522f`).

**Gemini content provider.** The 112 fix above, plus the four-model bake-off.

**Experience register.** `799c0c97e` — **the contract shape was INVENTED; all nine
harvested entries would have been refused.** The thread caught its own error, logged
it to `WRONG_CALLS` (`3c8396578`), and shipped a mechanical guard (`45b2a3d90`). This
was the roll-blocked item in memory; the roll has now happened twice over.

**oufe.com.** Humanised voice is now the **default, not an option**, and the copy was
rewritten (`13805c29f`). Seed 210 fixed — `target_topic` was a topic nothing consumes,
so the lane fired into a void (`4834b0d50`). Found the documented re-render path cannot
render an owned page, and every tool page is one (`84407636e`).

**gauntlet dead-CTA / 083.** Round 2 closed the coverage gap the council found; the
thread corrected its own count (`7f281cea9`). Seven discard sites, not two, and a
third endpoint (`d18eec0a6`). Stays OPEN.

**idea.uk VM site.** Model usage now logged on **every** call — the cost record had
been gated on caching (`fb10b2659`). Cost measured to a floor, and two copy checks
that only *looked* like they passed (`a13147fed`).

**Bug 103 (tool meta-descriptions).** A second unguarded call site found; the census
undercounts by two (`0610ca06a`). Backfill SQL dry-run verified but **NOT applied**
(`1dd0329f0`). **This lane has uncommitted Go — see §4.**

**Bug 029 (hung spawns).** Corroboration from a brand-new lane, 2 for 2, roll-adjacent
(`a67d71d6e`) — then a **correction to its own evidence: 1 hang, not 2**
(`f3cdc3377`). Still the blocker under several other lanes.

**Post-roll triage sweep.** `67bed924c` and `48cb65da9` — what v1.0.1174 made live and
what is still owed; **closed 010 and 070**, re-checked 040, 096, 099.

**Council / architecture.** Filed **119** — one seat's malformed JSON voids a round
that eight seats approved. Fifteen council prompts corrected because the code tier is
dead on two of three lanes (`e410b0ab0`, ties to `bugs_open/108`).

**Smaller live lanes:** 109 (render-context serialiser derived from the struct, not a
hand-written list — `595c1f499`); 080 (gap-planner's `new_page` routed through
`CanonicalisePage` — `9759687d1`); 095 (an empty page assembly now explains itself
instead of reporting COMPLETED — `6579e9ae1`); 085 (evidence-chart restored at plan
level, containment lifted — `e18583c80`); 071 (proven by accident: **you cannot remove
a CTA, and trying makes it worse** — `8228fb748`).

---

## 4. Volatile state — the uncommitted tree (33 paths)

**This is the part a crash makes ambiguous**, because nothing on disk says who owns
these edits. Nothing here is lost by a crash; it is lost by the *next* session running
`git add -A`.

**Modified, tracked — looks like live work in progress:**
- `platform/orchestration/actions/create_tool_component_action.go` (+12)
- `platform/orchestration/actions/deploy_tool_action.go` (+43)
- untracked `platform/orchestration/datahelpers/meta_description.go` **and its test**
  → `[INFERRED]` these four are one change, the **bug 103 tool meta-description** lane,
  matching its two commits at 19:38–19:39 and its unapplied backfill SQL. It is a
  `platform/` change, so it needs the council gate before commit.
- `cmd/reasoningset/main.go` (+75), `cmd/reasoningset/extract.sql` (+58)
  → the **reasoning dataset** lane. Memory says next step was materialising two joins;
  the diff size is consistent with that.
- **17 × `deployments/kustomize/.../uk_001/kustomization.yaml`** — every one bumps
  `newTag: v1.0.1150 → v1.0.1174`. This is the **record of a fleet-wide deploy that
  happened and was never committed.** Note it says 1174, while chassis is now on 1175.
- `bugs_open/029_...hung_spawns...md` — modified, uncommitted.

**Staged deletions — both safe, do not panic:**
- `bugs_open/043_..._generated_page_copy_invents_quantitative_claims.md` → the file
  exists in `bugs_closed/`. A completed close, staged but uncommitted.
- `bugs_open/112_..._shipped_css_diverges_from_the_pinned_palette....md` → **this is a
  renumber, not a loss.** The chain is `bug 110` (`3126682c6`) → renumbered to 112
  (`35eebafe4`, "another session took 110 concurrently") → but 112 was *already* the
  Gemini key bug, so the brochure thread refiled it as **113**
  (`generated_palettes_inherit_the_layouts_light_literals`). Same workstream, same
  day, same owner report, sharper mechanism. The old text is still in `HEAD` and
  recoverable with `git show HEAD:bugs_open/112_...css...`.

**Untracked odds and ends:** `clearideas.bash`, `live.html`, `reasoningset`,
`scheduler` (the last two look like built binaries — `[INFERRED]`, not verified),
`docs/.../NOTES_placeholder_check.md`, `docs/.../bugfix_029_dispatch_gate/PLAN_2026-07-26_dispatch_gate.md`,
`docs/agent_docs/sql_for_agents/213_dispatch_gate_matches_dispatcher.sql`,
`214_build_dispatch_watchdog.sql`.

**If you are the session that owns any of the above: commit it narrowly, now.**
`git commit <your paths> -m "..."`. A long-lived dirty tree is shared mutable state,
not a private workspace.

---

## 5. Bug numbering — three collisions live right now

Worth knowing before anyone routes work by number:

- **112** was used twice on the same day (Gemini spawn key; CSS palette). Resolved by
  refiling the CSS one as 113 — but a doc written this afternoon may still point at
  "112" meaning either.
- **083** and **107** each appear **twice** in `bugs_open/` right now.
- **016** and **017** are each used by two different closed cases (long-standing;
  documented in `bugs_closed/README.md`).

**Resolve by slug, never by bare number.** Numbering is one sequence across
`bugs_open/` and `bugs_closed/` and is never reassigned.

New today since the restart list: **113, 114, 115, 116, 117, 118, 119** — seven bugs
filed in about ninety minutes, mostly by the brochure and relojistas/webdesign lanes,
mostly found by *rendering pages and measuring them* rather than reading code.

---

## 6. What to do first after a restart

1. **Re-run `git status` and `git log`.** This file is a snapshot; on this tree the
   half-life is a couple of hours, not a day.
2. **Read `OPEN_THREADS_RESTART_LIST.md`** — §A is "ready to resume", §C is the owner's
   queue. Apply the corrections in §2 and §5 above before acting on it.
3. **Strike §C item 1** (Gemini writer revert) — resolved by v1.0.1175.
4. **The one owner action still blocking a finished piece of work:** publish the
   corrected fundamentallyai stylesheet. The fix is measured (101 unreadable text pairs
   → 1) and committed; it is inert until published.
5. Before touching any bug, `scripts/who-owns.py <number|slug>` — and remember its
   blind spot: it reads *commits*, so a session mid-fix with an uncommitted tree (see
   §4 — there are at least three) is invisible to it.

---

## 7. What was verified here, and what was not

**Checked live, just now:** the running pod image (`v1.0.1175`); the makefile tag; the
open-bug file count; the full `git status` and its staged contents; `git log` for the
last three hours; the existence of the deleted 112 file in `HEAD`; the first six lines
of both 112-CSS and 113 to confirm the renumber; the kustomize diff on one service.

**Read from commit messages, not independently re-verified:** every per-lane claim in
§3. Commit subjects in this repo are unusually descriptive and several are explicitly
self-correcting, which makes them good evidence of *what a thread believes* — but they
are a thread's own claim about its own work, not a measurement made here. Nothing in
§3 was checked against the cluster or the live sites.

**`[INFERRED]` markers appear inline in §4** where I grouped uncommitted files into a
lane by their shape and timing rather than by anyone telling me.

**Not attempted:** any DB query, any site fetch, any council or work-item state. The
restart list's §B/§C/§D sections are carried forward unchanged and unchecked — they
were accurate at 16:05 UTC.
