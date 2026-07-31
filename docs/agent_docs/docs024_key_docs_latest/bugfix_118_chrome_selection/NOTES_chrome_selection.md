# NOTES — bugfix 118, chrome component selection

Append-only, newest at the bottom.

---

## 2026-07-31 — picking the bug, and why 118

Swept `bugs_open/` against the live `.jsonl` transcripts of every session active
in the last five hours (`grep -oh 'bugs_open/[0-9]\{3\}' | sort | uniq -c`),
because `scripts/who-owns.py` reads COMMITS and is blind to a session mid-fix.
118 came back with 4 mentions fleet-wide and no owning workstream; the one
session with a real cluster of them (`15765e72`) turned out to be the
`analytics_gtm` lane, whose 12:39 bulk edit to seven chrome components is GTM
injection into templates — not selection. Last commit naming 118: 2026-07-27.

Checked validity before starting rather than after: `render_site_components_action.go:556`
still carries the bare `WHERE function = $1 ORDER BY name LIMIT 1`, and the live
library still has three deactivated footers sorting ahead of two active ones.

## 2026-07-31 — the measurement that changed the fix, three times

**First shape I had in mind was the bug file's candidate 1: add `AND is_active`.**
Three measurements killed it, in order.

1. **`ORDER BY name` under `AND is_active` picks `header-leopardess`** — an
   ACTIVE FORK of leopardessconsulting.co.uk's header. A fork carries its
   parent's `function`, so it sits in the generic pool. The bug file only
   considered the footer's tie-break (`footer-theme-chrome` vs `site-footer`)
   and the header's answer is worse than the disease. `GetComponentByFunction`'s
   own doc comment has said "forks should only be accessed by component_id"
   since it was written; two of the three call sites never honoured it. A doc
   comment enforces nothing.
2. **`is_active AND forked_from IS NULL` still picks `site-header`, which is
   `component_level='section'`** — a 6.6KB page-section component, not chrome.
   016b §9 recorded the principle ("`site-head` (`component_level=section`) is
   unreachable as chrome") long before anything encoded it in a predicate. With
   the level filter added there is exactly ONE eligible row per chrome function,
   so the tie-break question the bug file raised disappears instead of being
   answered by the alphabet twice.
3. **The fallback only fires on a slot with NO `site_components` row.** All 14
   real sites have all three rows. So the filed claim that candidate 1 "changes
   the rendered footer on every site" — the reason it had been parked for an
   owner call for four days — is wrong about the code path. It changes what an
   UNASSIGNED slot gets. Live blast radius today: `loancalculator.co.uk`,
   created 2026-07-30, zero chrome rows.

That third one is the useful correction, because it turns a blocked bug into a
shippable one. What genuinely needs an owner call is repointing the 11 sites
already pinned to `footer-4-column` — which candidate 1 never proposed to do.
**The bug file collapsed "fix the selection" and "repair the fleet" into one
decision, and only the second is fleet-visible.**

## 2026-07-31 — what I found that the bug file did not

`site_work_items` has been carrying `deactivated_component` items — *"Site
component footer points to deactivated component 'footer-4-column'"* — since
**2026-07-17**, two of them stamped `[unresolved after 2 attempts]`. So the
platform detects this state perfectly well. Their `HandlerAgent` is
`rerender-pages`, which re-renders **the component the row already points at**:
the deactivated one. The routed repair is structurally incapable of repairing the
finding, which is why they age out rather than close.

I did not fix that here. It is a different defect (a handler whose contract it
cannot satisfy) and fixing it means repointing assignments, which is the
fleet-visible half. Recorded in the bug file and in `LANDMINES.md`, because a
`complete` deactivated_component item reads exactly like a repaired slot.

## 2026-07-31 — missteps

- **I nearly shipped `ORDER BY name` on `GetComponentByFunction` as "just
  determinism".** It is not: the ordered query and the unordered one could have
  disagreed, and if they had, every page BUILD's chrome would have changed as a
  side effect of a tidy-up. I only found out by running both (RUNBOOK R3). They
  agree — for exactly two functions, which is the entire population that has a
  choice. **An `ORDER BY` added to an existing `LIMIT 1` is a behaviour change
  until you have measured that it is not.**
- **My commit message names three "same-file passengers" that were not in my
  commit.** I checked `git diff` on `LANDMINES.md` and `000_concept_index.md`,
  saw other sessions' uncommitted entries, and wrote a paragraph naming them as
  riding along. Between that check and the commit, the 137 lane committed both
  files (`f0a52f42b`) — carrying MY lines as THEIR passenger — so my pathspec
  matched two clean files and silently took neither. The paragraph is wrong in
  the record. Forward-only, so it stays; corrected here and in `WRONG_CALLS.md`.
  The cheap check I skipped: re-run `git status --porcelain <paths>` in the same
  command as the commit, not two tool calls earlier. On this tree a two-minute-old
  `git status` is a guess.
- **`sites` has no `deleted_at`.** I wrote `WHERE deleted_at IS NULL` from habit
  and got a hard error. Cheap and self-correcting, but it is the third time this
  repo's schema has punished assuming a soft-delete column; `\d sites` first.
- **A 0-byte `go.mod` is a full disk, not a broken repo.** `git archive HEAD |
  tar -x` into `/tmp` produced `go: error reading go.mod: missing module
  declaration` — a message that reads like the module is broken. `/tmp` is a 16G
  tmpfs at 94%. Extract somewhere else and `wc -c go.mod` immediately.

## 2026-07-31 — what shipped

`b052249d8` (+ `a77034379` gofmt). One predicate (`chromeEligibleSQL`), one
slot→function map (`ChromeSlotFunction`), one resolver (`ResolveChromeComponent`)
returning `(component, eligible, error)`; the two ASSIGNMENT call sites routed
through it; `GetComponentByFunction` given `ORDER BY name` and nothing else.
Registered as **CLC-013** in the same commit, per the ordering-exemption
condition that still stands. Council submission
`5bc232d6-590a-4476-a6b1-4fb6f61751c6`, submitted before the commit, trailer
`Council-Submitted:` (never `Council-Reviewed:` on an unread verdict).

Verified against a clean `git archive HEAD` tree because the working tree carried
another lane's mid-edit compile error — `cmd/agent-chassis` builds, package tests
green, and the two ordering tests proven non-vacuous by deleting both `ORDER BY`
clauses and watching them go red.

**Inert until a chassis image rolls.** The bug stays OPEN until then; a fix
committed but not live leaves the defect reproducible, which is the standing bar.

## 2026-07-31 (later) — it went live mid-session, and the owner said repoint everything

The chassis rolled to **v1.0.1219** while I was mid-repoint discussion. Verified at
the pod on **both replicas**, never at the tag: the three strings this change added
are present (`no eligible component for function`, the level whitelist, the
`ineligible_chrome` key) with `RenderSiteComponentsAction` = 6 as a positive
control in the same exec. My commits (18:36/18:38/18:46 UTC) predate the pod start
(19:09 UTC) — but that ordering is not evidence, which is exactly why the grep
exists (`bugs_open/153`).

Owner ruled: **repoint all eleven now.** Done — 21 assignments in one guarded
UPDATE (11 footers + 10 headers; `head` cannot move, no eligible component exists),
the WHERE carrying `pageComponentAgentWritableSQL`'s predicate verbatim so a locked
slot would have been skipped. None were locked. Prior mapping saved to
`site_components_repoint_backup_20260731`. Then chrome re-rendered on all 11.

**Result: 28 of 28 header/footer slots across 14 sites now render from an ACTIVE
component.** relojistas' stored footer emits `<h4>Explore</h4>` where it emitted
`<h4>Our Services</h4>`, and its Contact column is correctly absent — `bugs_open/111`'s
gate working at last on the component that actually renders, which is the change
whose silent failure filed 118 in the first place.

### Two gates the repoint exposed, which sharpen 166 considerably

Neither was visible from reading the check. Both were found by dispatching the
handler at a real site and watching it report COMPLETED while changing nothing.

1. `rerender-pages` renders chrome only when `refresh_site_components` is true in
   `input_data` — there is a `check_refresh_components` conditional in front of the
   step. The detector sets it, so that half is right; any other caller gets a
   silent skip.
2. **Even with the gate open and the assignment already corrected, the slot is
   still skipped**, because the `!force` idempotence exit tests `rendered_html IS
   NOT NULL AND != ''` — not whether the component changed. A repointed slot holds
   its old HTML, so it reads as "already rendered". ⇒ **the repair needs
   `force_rerender: true` and the detector does not set it**, which is now 166's
   cheapest fix candidate.

### Misstep

I NULLed leopardess' stored footer to get past gate 2 **without saving its
contents first** — my backup table captured `component_id` only. The artefact is
regenerable from the template (and it did regenerate), and deployed pages serve
their own HTML so nothing outward-facing was at risk, but for a few minutes the
site had no stored footer and I had no copy. Logged in `WRONG_CALLS.md`.

### What is NOT done, and it is the honest half

**Stored chrome is correct on every site; the DEPLOYED pages are not.** They still
serve the old footer until the **206 `page_rerender` items** now sitting at
`triaged` drain — `bugs_open/117` is why (chrome is a stored artefact) and
`bugs_open/149` owns the queue. `curl relojistas.com` still shows `Our Services` as
of 19:26 UTC. Do not read "28/28 slots active" as "the fleet looks right".

## 2026-07-31 (late) — 166's mechanism, found by hand and then closed

Carrying on from the repoint. The two gates I hit by hand are the whole of 166, and
neither is visible from reading the check:

1. `rerender-pages` renders chrome only when `refresh_site_components: true` is in
   `input_data` (a `check_refresh_components` conditional). The detector sets it, so
   that half is correct by construction; a hand-run gets a silent skip with a
   COMPLETED status.
2. **Even with the gate open and the assignment already corrected, the slot is
   skipped**, because the `!force` idempotence exit tests `rendered_html IS NOT NULL
   AND != ''` — whether the slot holds bytes, never whether it holds the RIGHT
   component's bytes.

Fixed both directions (`39afbf697`, council `e242e9d3`): `repointIneligibleChromeSlot`
> **CORRECTED at council round 1 — renamed `repointRetiredChromeSlot` and NARROWED to
> `NOT is_active`; see the round-1 entry at the bottom of this file for why the version
> described in this paragraph was dangerous.**
moves a slot off an ineligible component onto the one `ResolveChromeComponent` names,
and the exit gained `COALESCE(build_status,'') <> 'pending'`. The repoint runs ABOVE
the exit — below it an unforced render returns early, and every scheduled chrome
rebuild is unforced. A test asserts the ordering rather than a comment claiming it.

Three refusals, each with its own test because each turns the repair into a different
bug if dropped: repoint only when an eligible alternative exists (so `head`'s 13 slots
are left alone rather than churned for a library gap no site can fix); write through
`pageComponentAgentWritableSQL` and file **no** lock item on the zero-row path (the 069
gate owns that decision); set `build_status='pending'` rather than clearing
`rendered_html` (I blanked one by hand and had no copy — the slot must keep serving
its old chrome until the new render succeeds).

**Measured before submitting: net live change today is ZERO.** All 42 rows are
`build_status='rendered'`, none pending-with-HTML; the only ineligible assignments
left are the 13 `head` slots, where the repointer declines. It changes the *next*
deactivation.

### Missteps, both logged in WRONG_CALLS

- **I registered the seam one commit late.** `render_site_components` now reassigns an
  assigned-but-ineligible slot where before it only assigned an unassigned one — a
  widening of what a shared action promises its callers — and CLAUDE.md requires the
  register entry in the *same* commit. I had done exactly that for the 118 seam three
  hours earlier and did not repeat it, because 166 felt like a bug fix rather than a
  seam. The test is what the caller is promised, not how the work felt.
- **I read a two-hour-old clock into a `now() - interval '40 minutes'` window** and
  briefly believed the rerender queue had drained. A wider window returning fewer rows
  than a narrower one run earlier is impossible unless the rows changed — that tell was
  free and I walked past it. The truth: **11 of 206 complete, 195 still `triaged`**,
  oldest stuck item fleet-wide from seven hours before. The chrome repoint will not
  reach a served page until that queue moves, which is `bugs_open/149`'s lane.

## 2026-07-31 (late) — council round 1 on 166: REVISE, and it earned its keep

Gated by `prior_art_librarian`. Nine seats, five abstained. The round is worth
recording in full because **one objection stopped a genuinely damaging change.**

`debug_historian` objected that "net live change today: ZERO" rested on a
point-in-time census, citing `WRONG_CALLS`'s history of figures drifting before a
fix shipped. Re-running it at revision time found **three live rows** my trigger
would have moved: `idea.uk`'s header and footer (active, unforked, section-level)
and **`leopardessconsulting.co.uk`'s header — that site's own ACTIVE FORK.** My
code would have replaced a client's bespoke header with the house one, silently,
on the next chrome render.

**The root confusion: I reused one predicate for two different questions.**
`chromeEligibleSQL` answers "which component may be CHOSEN as a library default?",
where a fork must be excluded. It does not answer "which existing assignment is
unacceptable?", where a fork is the supported mechanism. Narrowed the trigger to
`NOT is_active` — which is exactly what `deactivated_site_components` detects, and
when the right predicate turns out to be the one the existing detector already
uses, the generalisation was never load-bearing.

The other objections and what each earned:

- `bug_historian` (locked+retired declines silently, then the exit waves it
  through) — **right, fixed in code**: the refusal is now reported and surfaces in
  `ineligible_chrome`.
- `bug_historian` (the sibling audit) — **done, and the answer is do-not-touch**:
  `getSiteComponents`/`getAreaComponents` share the `rendered_html IS NOT NULL`
  shape but are READERS; a pending slot still serves its old chrome, so filtering
  them would blank the page. Pinned by a test.
- `editquality` (is `pageComponentAgentWritableSQL` even right for this table) —
  measured: `site_components` carries all four lock columns; 069 put them there so
  one predicate could guard both.
- `reuse_agent`/`guardian` (is `link_site_components` already the designated
  repointer) — measured: it appears in **no** agent that renders chrome, so putting
  the repair there would make it inert on every path that renders.
- `prior_art` HIGH (the on-topic landmine was not quoted) — procedurally fair: it
  is this lane's own entry, written this morning, and I addressed it without ever
  quoting it. Quoted in round 2.

Round 2 resubmitted on the same correlation so the trail accumulates.
