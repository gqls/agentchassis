# HANDOFF — 2026-08-03b — the retraction seam is adopted, live, and has retracted real findings

**Supersedes `HANDOFF_2026-08-03_continue_here.md`** (which is still worth reading for the
`bugs_closed/168` history and the traps in §6 — they all still hold). This one covers what the
successor session did: **§4.1 of that handoff, "Decision 1 ADOPTION", is DONE.**

Read this, then `NOTES_deployed_asset_path.md` (newest at the bottom) and
`architecture_review/RFC_010_discovery_checks_can_raise_a_finding_but_not_retract_one.md`.

**Everything below is committed, live and verified. Nothing is half-applied.**

---

## 1. What changed, in one paragraph

`check_empty_sections` is now the first real adopter of the RFC_010 retraction seam (WII-009).
It closes `empty_section` findings whose slot it has **positively re-observed** as rendering
content, reusing `emptySectionVerdict` — the pure predicate already written for the completion
gate — so there is one answer to "is this section empty", not two. It went live on `v1.0.1243`
and, on its first real sweep, **retracted four findings raised in April 2026 that nothing in the
platform could previously close** — while leaving six others on the same site open.

## 2. State — all verified 2026-08-03, chassis `v1.0.1243`, both replicas

| thing | state |
|---|---|
| `check_empty_sections` retraction | **LIVE AND PROVEN.** Council `97923026` APPROVED round 1. |
| Commits | `2287606d1` (adoption) · `27891fab8` (council hardening) · docs `d983de570`, `3d05fc828`, `0acf30fd1` |
| First retractions | **4**, on `leopardessconsulting.co.uk`, sweep correlation `4401d952` |
| Fleet `result ? 'resolved_at'` | **0 → 4** (was 0 in all history) |
| Remaining retractable | **14 across 5 unswept sites** — they close on those sites' next sweeps |
| RFC_010 Q1 (two-strike) | **OPEN** — owner ruled accept-as-is + track. See §5. |
| Decision 2 dedup half | **NOT STARTED, still blocked.** Unchanged — see the previous handoff §4.3. |
| `bugs_open/179` | **OPEN**, unowned. Unchanged. |

**The verification command, and the reason it is shaped oddly:**

```bash
for POD in $(kubectl -n ai-persona-system get pods -l app=agent-chassis -o name | sed 's|pod/||'); do
  A=$(kubectl -n ai-persona-system exec $POD -- sh -c 'strings /app/agent-chassis | grep -c "re-observed healthy: all"' 2>/dev/null|tail -1); A=${A:-0}
  H=$(kubectl -n ai-persona-system exec $POD -- sh -c "strings /app/agent-chassis | grep -c 'COALESCE(page_id::text'" 2>/dev/null|tail -1); H=${H:-0}
  echo "$POD adoption=$A hardening=$H"
done
```

⚠ **This change removes no string literal, so `bugs_open/153`'s positive+negative recipe does not
apply.** The substitute — and it is *stronger* for a purely additive change — was to take the
pod-grep **before** the roll and date it: `re-observed healthy: all` was **0 on both replicas of
`v1.0.1238`**, and is **1 on both of `v1.0.1243`**. A stale same-tag binary would still read 0, so
the transition is the proof. **If you ship an additive change, take that baseline first; it costs
one command.** `$N=${N:-0}` remains mandatory (`grep -c` exits 1 and prints nothing on zero).

## 3. THE ONE THING TO UNDERSTAND BEFORE ADOPTING THE SEAM ANYWHERE ELSE

**Count the producers of your `item_type` before you write a line of code.**

`Resolved` closes rows by `(site_id, item_type, item_key)` — which is **coarser than the producer
set that shares that key**. The owner's ruling of 2026-08-02 actively *encourages* several
producers to converge on one key, and **13 item types already have ≥2 Go producers.**

The near-miss, which is why this is first: `check_undeployed_assets` was the obvious adopter —
95 open items, the biggest stale population in the queue, and a switch arm that already observes
*"Deployed."* and discards it. It is **the wrong one**. `undeployed_asset` is also filed by
`write_render_audit_findings_action` by deliberate co-dedup, and its finding — *"this image
serves broken on a real page"* — **requires** the HTML to reference the asset, which is exactly
what `check_undeployed_assets` reads as healthy. The two producers' evidence is *positively
correlated*, so adopting there would have retracted **every render-audit 404 finding fleet-wide**,
silently, on the next sweep.

```bash
grep -rn --include=*.go -E '(ItemType|itemType):[[:space:]]*"<your_type>"' platform/ internal/ | grep -v _test
```

More than one file? Then ask the harder question: **is my positive observation a REFUTATION of the
other producer's predicate, or merely unrelated to it?** Unrelated is not good enough — the UPDATE
closes the row either way. Both this and the early-return trap (§4) are in `LANDMINES.md`.

## 4. What is next, in the order I would take it

### 4.1 More adopters — the pattern is now proven and cheap to copy

`check_empty_sections.go`'s `findResolvedEmptySections` is the worked example. The shape:
enumerate **from the item side** (walk the slots/targets your items name for this site, ask the
database what is there now), never from the absence of findings. Single-producer candidates with
real open populations, all enabled, all single-producer (verified):

| check | item_type | open rows |
|---|---|---|
| ~~`check_required_fields_missing`~~ | `required_fields_missing` | ~~59~~ **DONE 2026-08-03 — see below** |
| `check_misdirected_cta` | `cta_names_unknown_destination` | 100 (**107** open when re-measured 08-03; still being filed) |
| `check_sprite_css_missing` | `needs_sprite_css` | 10 (all `unresolved`) |
| `check_voice_tells` | `voice_tells` | 25 |

> **UPDATE 2026-08-03 (later session) — row 1 is ADOPTED, committed `ba3aae47f`, docs
> `b312c409a`, council `64430363-a42a-4028-b84a-9a25ab707441`. NOT LIVE until the next roll.**
> 6 of the 59 retract on evidence; 50 refused as still-missing, 3 as holding no deployed
> component. Full working in `NOTES_deployed_asset_path.md` (bottom). Three things the next
> adopter should take from it:
>
> 1. **The trap list below is INCOMPLETE, and trap 1 is the one that fooled me.** It says to look
>    for a leading `if len(findings) == 0 { return }`. `check_required_fields_missing` has no such
>    guard, looked safe, and was inert anyway — its retraction-skipping `return` sat **mid-loop**,
>    fired by a 25-finding noise cap. **Read every exit between the scan and the retraction.**
>    Filed as its own `LANDMINES.md` entry.
> 2. **A NEW REFUSAL CLASS, for any check whose predicate is "the config declares X, the data
>    lacks X".** An unreadable schema/config yields NO required set, which computes to *nothing
>    missing* = healthy. It is the inverse — the observation could not be made — and unguarded it
>    retracts hardest exactly when a component's schema was dropped. Also in `LANDMINES.md`.
> 3. **A mutation that PASSES is not proof the guard is redundant.** Deleting the "no deployed
>    component" refusal left every test green because a guard in *series* (the NULL-schema
>    refusal) shadowed it under today's join. It took a synthetic row to pin it alone.
>
> `page_id` also inverts here: all 70 rows carry a **NULL** first-class column (this check's
> filing half never sets `WorkItemSpec.PageID`), so the `COALESCE` fallback is the arm that
> actually fires — the opposite of `empty_section`, where it only pins a preference. Fixing the
> filing half is a **live follow-up**, deliberately not folded in.

⚠ **`check_image_url_404` is tempting (a real HTTP probe is the cleanest possible positive
observation) but another lane is actively editing it** — co-ordinate first. (Still true 08-03:
`gofmt -l` shows it dirty in the shared tree.)

**Three traps, all paid for already** (and see point 1 above — trap 1 is **narrower than it
reads**):
1. **The early return.** Most checks open with `if len(findings) == 0 { return }`. Correct while a
   check can only file; exactly backwards once it can retract, because a site with zero findings
   is the *only* site it fires on and precisely the one whose stale items need closing. It is
   green in every test that has a finding. **Write the zero-findings retraction test first.**
2. **Refuse the ambiguous case.** A target that has *vanished* is equally "fixed" and "silently
   deleted by a rebuild" — this platform's most repeated failure. 10 of 47 items were in that
   state. Do not retract them.
3. **`(page_id, slot_name)` is not unique** (`bugs_open/156`) — nor are many other "obvious" keys.
   Check for multi-row targets and be conservative when they disagree.

**Prove every guard by mutation.** Seven mutations were required to fail on this change; six of
the seven guards would have looked fine under a green test run.

### 4.2 Watch the remaining 14 retract, then stop claiming and start counting

```sql
SELECT s.domain, swi.item_key, swi.result->>'reason'
FROM site_work_items swi JOIN sites s ON s.id=swi.site_id
WHERE swi.result ? 'resolved_at' ORDER BY swi.completed_at DESC;
```
Expect it to climb from 4 toward 18 as the five unswept sites get their next sweeps. **If it
stalls at 4, the question is whether those sites are being swept at all** — that is a scheduling
question, not a defect in this code, and the standing lesson applies: *a silent mechanism is
usually undriven, not missing.*

> **MEASURED 2026-08-03 (later session): it is STILL EXACTLY 4.** All four are the original
> `leopardessconsulting.co.uk` retractions from sweep `4401d952`; none of the 14 have closed.
> So the branch above resolves to the *scheduling* arm: those five sites have not been swept
> since. **Do not re-run this query hoping for a different number — go and look at whether the
> sweeps are firing**, or dispatch one yourself as the 08-03 session did to produce the first
> four. Recorded so the next reader does not spend the query twice.

### 4.3 RFC_010 Q1 — the two-strike interaction (owner-ruled, tracked, not fixed)

See §5. It is written up as **Q1** in the RFC with the measurement attached.

### 4.4 Decision 2's dedup half, and `bugs_open/179`

Both unchanged from the previous handoff (§4.3 and §4.2 there). Still real, still blocked on the
87 duplicate rows and the asymmetric ordering.

## 5. OWNER RULING 2026-08-03 — the two-strike interaction

Three council seats (`guardian` medium, `improvement_guardian` low, `bug_historian` low)
**independently** asked for a human decision: a retraction writes `complete` onto an **existing**
row, which feeds `insertWorkItem`'s two-strike counter **identically to a real handler fix**. At
2 strikes the next genuine detection of that key is born `unresolved` and undispatchable — the
landfill, created by the mechanism meant to drain it.

> **RULED: accept as-is, track as a follow-up.** Measured **0 of 17** affected (all were ≥15 days
> old; the counter only sees rows created within 7 days), and `insertWorkItem` sits on the insert
> path of every work item in the estate — changing it from inside a check adoption is the
> `bugs_closed/124` shape the guardian seat exists to veto.

Recorded as **RFC_010 Q1**. Whoever picks it up: `workItem.recurrenceExpected` is the existing
lever that already exempts a class of items from that counter — read it before inventing a second.

## 6. What the council round taught, beyond this change

**APPROVED at round 1**, 15 seats, 3 advisory objections, none high. Four advisories were
checkable and were checked rather than filed (one produced code — the `page_id` hardening; three
were discharged with evidence).

⚠ **The finding worth carrying: the gate cannot see the docs half of your work.** Four seats
independently objected that "you say you will file this hazard to LANDMINES/register but no edit
does it". The gate **refuses docs client-side**, so that edit could never have appeared in the
plan they reviewed — the docs had in fact already shipped. The previous round's lesson was
*"evidence you hold and do not cite is evidence you do not have"*; this is its sharper form —
**evidence you *cannot* cite, because the submission schema excludes it.** The fix is cheap: name
the docs commit in the `rationale`, which is the only channel the gate leaves open.

## 7. Correlations

| what | id |
|---|---|
| council (APPROVED r1) | `97923026-2b2d-4925-b9a3-de6f70c49d2b` |
| the sweep that produced the first 4 retractions | `4401d952-4b1b-472c-b364-4d9fedb369f1` |
| previous handoff's council (`168`, APPROVED r3) | `abd9b119-d274-43bf-a03f-cf45bfb6b881` |
| RFC_010 Decision 1 council (APPROVED r1) | `846f4f3d-8958-4e4c-be81-d5f02e20852d` |

## 8. Environmental, will bite you

**`/tmp` is a 16G tmpfs and was at 97%** (14G of session scratchpads). The Go linker writes there,
so `go build ./...` fails with `mapping output file failed: no space left on device`, which reads
exactly like a code error. Fix without touching other sessions' data:
`TMPDIR=/home/ant/.cache/buildtmp go build ./...` (235G free on `/`).
