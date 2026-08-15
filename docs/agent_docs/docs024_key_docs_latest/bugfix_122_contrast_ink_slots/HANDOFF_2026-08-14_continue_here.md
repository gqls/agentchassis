# HANDOFF — bug 122 lane. START HERE. Written 2026-08-14 (afternoon), status block updated 2026-08-15.

## ⚡ 2026-08-15 STATUS — read this first, then the body

**v1.0.1300** (pods 20:36Z 08-14, stamp `a2a691213` — chassis provenance line scrolled, stamp read
from the adapter and probed into the chassis binary with both controls). Ancestry:

| commit | in v1.0.1300? | consequence |
|---|---|---|
| retraction `5639a1103` | **LIVE** | Monday unchanged |
| ink round 2 composited `8ad05d01a` | **LIVE** | — |
| 5.0 + kill-switch round 1 `d4bbbf645` | **LIVE** | **the fleet now emits 5.0 values on any re-render** (all call sites wired, so round 1 behaves correctly today; the zero-value hazard is for FUTURE unwired callers) |
| 5.0 round 2 `e0f239118` | **NOT LIVE** | **THE GATE SAYS HOLD, and it is right** — the running version is the one the council said REVISE to. Behaviourally identical for the canary, but the owner's gate should not run on a REVISE-flagged binary when the APPROVED one is a roll away. `829a8f3e` stays `deferred` |

**THE "RARE" THIRD-PARTY EXPOSURE FIRED WITHIN HOURS OF BEING NAMED.** `visual-design-audit` — a
routine pipeline — filed `needs_design_review` items that re-rendered **cookly.uk** (14:22/14:24Z)
and **robot-hands.com** (15:25Z) on 08-14, on the then-live **4.5** binary (v1.0.1298). Served now:
robot-hands `--color-primary-ink: #8a97bd`, cookly `--color-accent-ink: #af4625` — **the 4.5
round-2 brand-tinted values are user-visible on two sites**, pre-empting the owner's staged gate in
the SAFE direction (legible, brand-hued, just at the threshold he revised away from).
dartsonline (`#F0F2F7`, held item intact), webdesign.co.uk and vonc are untouched.

**Monday therefore grades on the `#8a97bd` branch of the §3a table** (round 2 at 4.5 underneath;
both must-retract rows still clear at 4.56 / 5.16, verdicts unchanged) — **unless another
re-render lands first: read the served ink again on the day.** This is the third time this
handoff has had to say that, and the second time events proved it.

**The 226 is now 225 + 1 CANCELLED**, and the cancellation is NOT ours and NOT the retraction:
`contrast_failure:/gripper-payload-calculator.html#A.cta-btn`, cancelled 16:36Z 08-14 with **no
`resolved_by`, no `reason`** — the retraction always stamps both. Probably the bug-268 buttons
lane (the cta-btn/gradient family is in their remit now) — unconfirmed, left as an open question.
Consequence for §3b: robot-hands' first-pass ceiling is **33**, not 34. All three
`/selection-guide.html` canary rows are still `deferred`.

**Next trigger unchanged: a roll whose stamp has `e0f239118` as ancestor** → un-defer + file
webdesign (§3a-bis) → owner looks → "Go". Saturday's pricing task (§4.1) is now TODAY.

---


> **STATUS IN ONE PARAGRAPH.** The retraction is built, council-APPROVED, and **live**, now on
> `v1.0.1298` / `bc39e7bf5` (re-verified at the binary with controls today). It has **still never
> run**, and cannot until **2026-08-17 14:54:23Z**, because the render-audit rotation is a 7-day
> per-site window — that is the mechanism working, not a fault. Nothing is owed before then except
> one dated pricing task on **08-16**. Everything else in this lane is a **dated verification**, and
> §3 is the procedure for it. Read §1, then §3. Nothing is on fire; the queue is empty, no site is
> locked out, and the 226 rows are parked and stable.

Supersedes `HANDOFF_2026-08-12c_continue_here.md` **for state**. That file remains the reference for
*why* the retraction is shaped the way it is (§1b's five council objections, §2's correction to the
still-failing set) and it is still correct — do not re-derive that reasoning, read it.

---

## 1. What is live, measured today with controls

| | |
|---|---|
| release | **`v1.0.1298`**, pods started **2026-08-14 08:58Z** |
| stamp | **`bc39e7bf5`** — `browser-runner-adapter` and `render-audit-adapter` print it in their own startup provenance line; `agent-chassis` verified by binary probe (`bc39e7bf5…` **PRESENT**, yesterday's `69612d692…` **absent**, so the probe discriminates) |
| the retraction (`5639a1103`) | **LIVE** — `git merge-base --is-ancestor` → true |
| ink derivation round 1 (`12cf55015`) | **LIVE** |
| ink derivation round 2 (`8ad05d01a`) | **LIVE** — this is the one that matters, see §2 |
| 226 `contrast_failure` rows | **unchanged**: all `deferred`, `retracted` **0**, `max(attempt_count)` **0** |
| rotation | `site-render-audit-rotation` enabled, **0 sites due**; `robot-hands.com` first at **2026-08-17 14:54:23Z**, then `loancalculator.co.uk` 15:54Z (**0** open rows — anything retracting there is wrong), `cookly.uk` 16:55Z |

**"Did my fix ship?" is a git query, not a grep.** The startup provenance line **scrolls** on busy
services — an empty grep means "not in range", never "unstamped". Fall back to the binary probe, and
**always run a negative control** in the same breath:
```bash
kubectl -n ai-persona-system logs -l app=<svc> --tail=400 | grep -m1 -o '"git_commit":"[a-f0-9]*"'
POD=$(kubectl -n ai-persona-system get pods -l app=agent-chassis -o jsonpath='{.items[0].metadata.name}')
kubectl -n ai-persona-system exec $POD -- grep -aq "<expected-sha>" /proc/1/exe   # and one that must be ABSENT
git merge-base --is-ancestor <your-commit> <the stamp>
```
Never `strings` (absent from the images). Never grep for "some 40-hex string" (matches Go's digit
table). Read the stamp of the **service you mean** — a release can straddle commits.

---

## 2. The one thing that changed today: the ink derivation is LIVE but DORMANT

Another lane (`bugs_open/122` §2 and §7, session `581eb30a`) repaired `legibleInkFor` so
`--color-<x>-ink` keeps the brand hue instead of collapsing to `--color-text`. **Both rounds are now
in the running binary.** But **nothing schedules a re-render**, and a site's `--color-primary-ink`
only changes when its `styles.css` is regenerated. Their rollout is deliberately staged —
dartsonline first, under an owner ruling, widening one site at a time.

**Measured today: robot-hands still serves `--color-primary-ink: #E2E8F0`, with no page-level
override.** So the site has not re-rendered and the canary is in its clean state.

⚠ **This makes the discriminator MORE necessary, not less — and the reason is worth reading slowly,
because it is the kind of change that happens silently.** Until 08:58Z today, "nothing has changed
for any visitor" rested on **two independent facts**: the code was not in the binary, *and* no
stylesheet had re-rendered. It now rests on **one** — and that one is nobody's to hold. ~~Any lane
re-rendering any of those 14 sites, for any unrelated reason, changes their link colours.~~

> **CORRECTED 2026-08-14 (late), raised by `581eb30a`, verified here with a WIDER sweep than
> theirs:** the struck sentence is **false, in the safe direction**. `/assets/css/styles.css` is
> regenerated by exactly **one** step in the live fleet — `webdesign-agent`'s `generate_css`
> (`render_css_from_spec`, the sole caller of `buildLegibleInkDefaults`). Checked across every
> live agent's workflow steps for `%render_css%`, then re-checked for `%css%` and `%style%` to
> catch other spellings: the five additional hits *select* a collection, write **sprites.css**, or
> *read* CSS for fingerprinting — none regenerates styles.css. `page-rerender` / `rerender-pages` /
> `rerender-site` / `rerender-chrome` only write the `<link>` tag (`rerender_pages_actions.go:558`).
> **So page and site re-renders CANNOT move the ink values.** The exposure is not "any unrelated
> re-render"; it is **"someone dispatches `webdesign-agent` at one of those sites, or files an item
> whose `handler_agent` is `webdesign-agent`"** — and the held item covers the only one filed. The
> two-protections-to-one framing stands directionally; the likelihood was overstated by a wide
> margin, and a warning people must ignore to do their jobs is one they stop reading. (Live case in
> point the same evening: a session re-rendering dartsonline page sections was, on the struck
> model, one command from spending the owner's gate; on the measured model it cannot touch it.)

So: **read the ink immediately before grading. Never carry a reading forward**, including the
`#E2E8F0` recorded above. This lane has already made the "it was live this morning" mistake once
(`12c` §1c), about this same site, and the window is now shorter than it was then.

(The other lane has flagged the same shift to the owner in its own terms: his pending ruling on the
visual change went from a hard gate to a soft one overnight, with no commit of theirs involved.)

---

## 3. MONDAY, 2026-08-17 ~14:54Z — the whole procedure

The first exercise is a single-site canary **by construction** (the rotation's `pre_query` is
`LIMIT 1`), so it costs nothing. Baseline is captured: `retracted` **0**, 226 `deferred`, 0 carrying
`batch_id`.

### 3a. FIRST, establish which mechanism you are testing — git query first, hex second

```bash
git merge-base --is-ancestor 8ad05d01a <the service stamp>   # AUTHORITATIVE
curl -s https://robot-hands.com/selection-guide.html | grep -- '--color-primary-ink'   # page block WINS
curl -s https://robot-hands.com/assets/css/styles.css  | grep -- '--color-primary-ink'   # else this
```

| served ink | meaning | canary verdict |
|---|---|---|
| **`#E2E8F0`** | no re-render — **the expected state, and the clean one: my retraction is under test** | table in 3b stands |
| `#94a0c2` | **round 2 at the owner's revised 5.0 threshold** (see the revision note below) — 5.88 / 7.20 on the canary grounds | table in 3b stands, both rows still retract |
| `#8a97bd` | round 2 at the original 4.5 — worst-of-four 4.56:1 | table in 3b stands, both rows still retract |
| `#7d8bb6` | round 1 **without** round 2 — measures 4.55 declared but **3.93 composited** | `card-link` may file fresh — **their regression, not a retraction bug**. Stop and tell them |

> **REVISED 2026-08-14 (evening) — three further owner rulings, relayed via the `581eb30a` session
> and each verified at the artefact before being written here:** the ink threshold moves **4.5 →
> 5.0** for this change (the "unless someone says otherwise" branch of the AA default — his call),
> a **kill-switch is wanted** after all, and widening waits for **"Go after I have seen
> dartsonline.com"**. Consequences, all checked:
> - The dartsonline rebuild item `829a8f3e` is **HELD** (`deferred`, hold note in
>   `spec.held_2026_08_14`, one-line restore inside) so the owner's gate shows the values that will
>   actually ship. **Un-defer and re-file at 5.0 for dartsonline AND webdesign.co.uk** once the
>   code change lands. The `581eb30a` session owns that code change;
>   `platform/colour/contrast.go` and `palette_specialised_slots.go` etc. are **frozen to this lane
>   until they message done** — a pathspec commit cannot protect either side from a same-file
>   passenger.
> - **dartsonline's served CSS has ZERO live ink consumers** — its one grep hit is line 937, which
>   is prose inside the renderer's own comment. **A grep counts strings, not consumers.** Its single
>   real consumer is `page_components`-side, one eyebrow on `/index.html` — so the darts canary
>   shows the owner ONE small label. **webdesign.co.uk is the second canary** because five layouts
>   carry `a { color: var(--color-accent-ink, …) }` in the stylesheet itself — every in-prose link.
> - At 5.0, verified two-ways: dartsonline `#94A0C2` / `#F18072` (5.122 / 5.125); robot-hands
>   primary is also `#94A0C2` (worst 5.077). webdesign accent `#915E2C` (5.151) is
>   replication-only — pin it when the code lands.
> - My own consumer census ("4 components / 37 placements") was **page_components-only and missed
>   the layout surface entirely** — the five-surface list in `bugs_open/122` §6 was already on
>   record and I had read it. A census answers the question you encoded.

### 3a-bis. THE RE-FILE, ready to run — but gate it on the RUNNING BINARY, not on the commit

**Verified 2026-08-14 evening, after the hold:** the watcher's full timeline shows `829a8f3e` was
NEVER claimed (`triaged` 16:48→17:11, `deferred` from 17:12, `claimed_by` empty, `attempt_count` 0),
and dartsonline's served stylesheet is **byte-identical** to the banked before-state (sha
`16eb767f…`, matched post-hold). No 4.5 rebuild ever ran. Second canary pre-flighted: webdesign.co.uk
pin present (07-25), collision guard **0 rows**, before-state banked
(`before_css_webdesign/…`, sha `50d55d8d…`, `--color-accent-ink: #2b2b2b` == text, the pre-fix
collapse).

**⚠ SEQUENCING: their sha existing is NOT the trigger. Go changes are inert until an image is built
and ROLLED** — un-deferring after the commit but before the roll re-renders at 4.5, which is
exactly what the hold exists to prevent. The gate is ancestry **of the running stamp**:
```bash
kubectl -n ai-persona-system logs -l app=agent-chassis --tail=2000 | grep -m1 -o '"git_commit":"[a-f0-9]*"'
git merge-base --is-ancestor e0f239118 <the stamp>     # TRUE, and only then:
# (e0f239118 = ROUND 2 of their change, after a council REVISE found the inkPolicy zero-value
#  defect; d4bbbf645 is its ancestor, so gating on e0f239118 covers both. Gate on e0f239118 —
#  a stamp carrying only d4bbbf645 would run the version whose zero value silently emits nothing.)
```

> **THE CURRENT STATE, measured 2026-08-14 evening (and corrected the same evening — read both
> halves).** The 5.0 change is committed (`d4bbbf645` + comment followup `ec9a0ee2f`; council corr
> `d60aab29…`, **APPROVED at round 2** — the trail reads REVISE → APPROVED, and the REVISE found
> the zero-value defect, so the round paid for itself) and **the fleet does not run it**: v1.0.1299's stamp is
> `6f8efa158`, probed with three controls, and `merge-base --is-ancestor d4bbbf645 6f8efa158` →
> **false**. Retraction and ink round 2 both still live. **The item stays held until that query
> returns true for `d4bbbf645` specifically.** *(Superseded the same evening: gate on
> **`e0f239118`** — round 2 after a council REVISE; see the gate block above.)*
>
> > **CORRECTED 2026-08-14, caught by the `581eb30a` session, verified here at the clocks:** the
> > first version of this note read "v1.0.1299 rolled WITHOUT it" and filed the incident under
> > `bugs_open/249`'s roll-straddling hazard. **Wrong mechanism.** 1299 was built from `6f8efa158`
> > (committed **14:28Z**) and rolled 15:32Z; `d4bbbf645` was committed **18:27Z** — three hours
> > AFTER the roll, and `6f8efa158` is its *ancestor*. A build cannot carry a commit that did not
> > exist. **Nothing was omitted and there is nothing to investigate: the next roll carries
> > `d4bbbf645` normally.** The distinction matters because the two readings predict different
> > futures — "a roll can silently skip a committed change" sends the next thread hunting a pinning
> > bug this case is not an instance of, and would pollute `249`'s evidence base with a
> > non-example. The RULE survives its corrected example and needs no incident to justify it:
> > **gate on ancestry of the running stamp, never on "a roll happened" — and equally, never on "a
> > commit exists".**
>
> ⚠ When the next roll lands, the probe's positive control must be the NEW stamp, not `6f8efa158` —
> only a build's own sha is stamped, so yesterday's stamp reading "absent" on a newer binary is the
> probe working, not failing. This bit once already tonight.
Then, in order:
```sql
-- 1. restore the darts item (the hold note carries this too)
UPDATE site_work_items SET status='triaged'
 WHERE item_key='css_rerender_ink_round2_dartsonline_com_20260814' AND status='deferred';
-- 2. file the second canary
INSERT INTO site_work_items
  (site_id, source, pipeline, item_type, severity, priority, status, item_key, summary, spec,
   created_by, approval_mode, max_attempts, handler_agent)
SELECT s.id, 'operator:bugfix_122_contrast_ink_slots', 'build', 'needs_design_review', 'medium',
   100, 'triaged', 'css_rerender_ink_round2_webdesign_co_uk_20260814',
   'Second canary (owner-added): re-render styles.css at the revised 5.0 threshold. Five layouts put accent-ink on every in-prose link, so THIS site exercises the decision.',
   jsonb_build_object('bug','bugs_open/122','domain','webdesign.co.uk',
     'reason','ink_derivation_round2_owner_gated_canary',
     'owner_ruling','2026-08-14: 5.0; Go after I have seen dartsonline.com'),
   'operator:bugfix_122_contrast_ink_slots', 'auto', 3, 'webdesign-agent'
FROM sites s WHERE s.domain='webdesign.co.uk';
```
> **⚠ `handler_agent` must be the FIRST-CLASS COLUMN, not a spec key (corrected 2026-08-14 late).**
> My original dartsonline INSERT put it in `spec` only. The dispatcher and claim path read the
> COLUMN (`load_work_item_actions.go:676`, `claim_work_item_action.go:132`), and an empty column
> means the item is claimed then immediately **`blocked`** — "No handler_agent set" — the moment it
> is un-deferred. Found by the council's `debug_historian` seat nudging a re-check of the hold; the
> held row is repaired (`handler_agent='webdesign-agent'` set 2026-08-14 while safely `deferred`),
> and this staged INSERT now matches the proven 08-09 template shape. While `deferred`, the item is
> unclaimable regardless (claims filter `status IN ('triaged','approved')`), so the repair was safe.
>
> **The hold itself is a SINGLE status lock, and all SEVEN un-park paths are verified shut** (six
> checked by both lanes independently; the seventh found by `581eb30a` following a code comment
> AFTER both lanes had published their enumerations as complete): the two claim predicates
> (`triaged`/`approved` only), the promoter (`detected` only), release-on-unhealthy (needs a prior
> claim), the stale reaper (`triaged`+build), **the admin dashboard** — retry
> (`site_admin_handlers.go:886`) resets `{needs_human_review, failed, blocked, unresolved}` →
> `triaged`, resolve (`:930`) closes that set plus `triaged`, `deferred` in neither — and
> **`feasibility-recheck`**, a live enabled `scheduled_tasks` row (600s) whose `pre_query` promotes
> `blocked` → `triaged` where the handler exists: `WHERE wi.status='blocked'`, so closed for
> `deferred` (verified at the live row). **A table is a very convincing way to publish an
> unenumerated absence** — both lanes did it, the same evening the census lesson was written.
>
> **⚠ RELEASE-TIME behaviour change from the column repair, know both halves before you un-defer:**
> with `handler_agent` now set, `blocked` is **self-healing within one 600s tick** (the recheck's
> `EXISTS(agent_definitions…)` test now passes). That is what you want for the rebuild actually
> running — a transient block retries itself. It also means **a block will no longer hold this item
> still**: if you need to re-hold it mid-flight, set `deferred` explicitly. Before the repair,
> `blocked` was a permanent trap; after it, `deferred` is the ONLY parking state.
>
> **And note what the seat's hazard actually was, because both directions got stated before the
> truth did.** `debug_historian` feared the empty column would let the canary FIRE early. The
> `581eb30a` lane then reasoned it could fire via blocked → operator-retry. **Neither: it could
> never render at ANY threshold** — the retry resets *status only*, the column stays empty, and the
> next claim re-blocks it. The broken filing's real failure mode was a **block/retry jam** at the
> owner's gate moment, burning operator attention while looking like a stuck queue. The repair
> removed the jam; `deferred`'s absence from the retry list is what keeps the *held* item beyond an
> operator's reach. Same objection, two wrong theories, one real defect — a seat's QUESTION can be
> worth more than its theory, and the measured answer beat both lanes' first readings.
```sql
-- (gate sha updated: round 2 of their change is e0f239118 — gate on THAT)
```
**Grade against these, written before any run** — full-file diff vs the banked shas first, then:
| site | slot | expect at 5.0 | status |
|---|---|---|---|
| dartsonline | primary-ink | `#94A0C2` (5.122) | verified two-ways |
| dartsonline | accent-ink | `#F18072` (5.125) | verified two-ways |
| webdesign.co.uk | accent-ink | `#915E2C` (5.151) | now verified two-ways — their pinned test computed it independently |
| webdesign.co.uk | primary-ink | **unchanged `#5c6b5d`** | the no-op branch: it already clears 5.0 (5.32 / 5.65, verified) — an unchanged primary there is CORRECT, not a failed rebuild |

**Dartsonline's served ink, three-way after the rebuild runs:** `#F0F2F7` = nothing shipped ·
`#8a97bd` = **the 4.5 binary ran — the roll didn't carry `d4bbbf645`; stop, do not show the owner** ·
`#94a0c2` = 5.0 live, correct (accent `#f18072`).

**The `pickInkOn` split — reviewed here as asked, and CONCURRED, with standing.** Their change
splits `inkFloorContrast = 4.5` out so the 5.0 ruling reaches only `legibleInkFor` (the `-ink`
slots: links and eyebrows on the page ground) and NOT `pickInkOn` (the `-text` slots: labels on
filled controls). They flagged this as their judgement about the ruling's reach, top risk in their
submission. **It is the only reading consistent with the owner's own default rule, which was given
in THIS session:** "as a default we only need to get to AA **unless someone specifically says
otherwise in the brief**". The 5.0 was said about *this change* — the ink margins he was shown.
Nobody said anything about filled-control labels; therefore they sit on the default, and the
default is AA. Raising `pickInkOn` too would apply "specifically says otherwise" to a mechanism
that was never in the brief. The split is not caution — it is the ruling's own structure,
implemented.
Then the owner looks at dartsonline (and webdesign, where the links are); his "Go" gates widening.

**Why the hex alone cannot carry this:** round 1 emits a *plausible navy* that passes any eyeball
check while failing on the real ground. "It's a navy, so the fix is in" reaches the right conclusion
from the wrong evidence. Both commits now travel together, so branch 3 is unlikely — kept because
another lane's roll plus an unrelated re-render is outside anyone's control.

### 3b. THEN grade the rows — per selector, never by count

`robot-hands.com` ceiling: **34** open rows across **21** pages. A first pass may close **at most 34**,
and only on pages it actually measured. The sharpest test is one page with its own control:

| `/selection-guide.html` row | required | why |
|---|---|---|
| `…#A.info-card-grid__card-link` | **RETRACT** | migration `368` fixed it |
| `…#SPAN.info-card-grid__eyebrow` | **RETRACT** | same |
| `…#A.cta-btn` | **STAY OPEN** | a *type* error, not a contrast one — see §4 |

Same page, same run, opposite required outcomes. **If all three close, the scope is wrong** — read
`12c` §1b(1) and stop. No count of closures can draw that distinction.

```sql
-- did anything retract, and did the mechanism say so?
SELECT item_key, status, result->>'resolved_by', result->>'reason', result->>'resolved_at'
FROM site_work_items WHERE item_type='contrast_failure' AND result ? 'resolved_at'
ORDER BY result->>'resolved_at' DESC LIMIT 20;
-- the park, which now drains on its own
SELECT status, count(*), count(*) FILTER (WHERE result ? 'resolved_at') AS retracted
FROM site_work_items WHERE item_type='contrast_failure' GROUP BY status;
```
The action's result carries `retracted`, `retracted_parked`, `retraction_scope_pages`, and
`retraction_unavailable:true` when the adapter is too old to say what it measured. **Seeing
`retraction_unavailable` means the adapter has not rolled — that is the version-skew branch working.**

### 3c. THEN, and only then, unpark

One `UPDATE` at the foot of migration `389`, predicated on `spec->>'parked_by' = 'migration_389'`.
Row-level backup: `scratchpad/backups/backup_park_contrast_failure_20260811.tsv`. **Expect fewer than
226 remaining by then — that is the point, not a discrepancy.**

---

## 4. Also open in this lane

| | item | state |
|---|---|---|
| 1 | **📅 2026-08-16 — price the discovery-rotation ramp.** Calls AND tokens (baseline ~248k input tok/h idle; driven sweep ~806k/h). Queries at the foot of `sql_for_agents/395_enable_quality_discovery_rotation_slow_ramp.sql` | **owed, dated** |
| 2 | `A.cta-btn` — **root cause found 08-13** and it is a defect class of its own: `.cta-btn-primary` reads `--color-cta-bg` into a `color:` slot, and that token is a **`linear-gradient`** on 5 of 10 sites. Not a valid `<color>`, so the declaration is discarded, `color` **inherits** `#ffffff` over a `#ffffff` button. **Confirmed at the instrument** (filed row: `fg` and `bg` both `rgb(255,255,255)`, ratio 1, sample "Run MatchMatrix"). 16 of 17 filed `%cta-btn%` rows fleet-wide; the 17th is a control at 2.27:1 (valid token, merely pale). **No ink fix or repoint can close it.** Written up as `bugs_open/122` §11 | filed, unowned |
| 3 | `bugs_open/212` §8 — component-painted grounds (~24 failures) | **owner's**, architecture |
| 4 | `dark_section_audit` straddles the same generic hole as `contrast_failure` did | `bugs_open/213`'s call or the owner's |
| 5 | ~~A note claiming a bare `curl` 403s on every site~~ — **CLOSED 2026-08-14, and it was OURS.** The claim originated in **this lane's own** `LANDMINES.md` entry (source line "2026-08-06, bugfix_122 lane"), was lifted verbatim into another lane's test comment, and is false: 7/7 pinned domains return **200** bare, identical with a browser UA, on two separate egress paths. Corrected at source (`26a2a6541`), both copies now fixed. **Keep the shape:** the *true* fact is one entry over (`Python-urllib` gets 403), and a `urllib` fact generalised into a `curl` fact survived eight days because it sat as a **parenthetical inside a correct command** — the command works, so nothing invites doubt about the sentence attached to it | **closed** |

---

## 5. Standing traps this lane has paid for

- **Grade per selector, never by fleet total.** It rose 109 → 112 while every targeted failure closed.
- **A filed count is not a found count**, and **226 is a FLOOR** — the audit was capped at 25 pages until `v1.0.1288`.
- **Resolve a CSS custom property at the layer that WINS.** A page `<style>` block beats the site stylesheet, and page-level palette blocks are routine here. Reading `styles.css` alone gave "3 of 8" where the truth was **5 of 10** — the convenient check *under-reported* the damage.
- **A `var(--x, fallback)` whose `--x` is DEFINED BUT OF THE WRONG TYPE is worse than an undefined one** — the fallback is dead code and the property inherits. In `LANDMINES.md`.
- **A figure from a probe you fed by hand: the INPUTS are the claim, not the output.** A `[MEASURED]` on the result certifies arithmetic that was never in doubt and launders the fabricated half. Two of the other lane's published hexes came from invented grounds; both were caught only by a second implementation transcribing inputs from the artefact. `WRONG_CALLS.md` 08-14.
- **A `file:line` in a handoff is a pointer, not a quotation.** Open the file.
- **A pathspec commit protects others from your STAGED files and does nothing about your STALE ones**, in either direction. `git show --stat` **after** committing is the only account that cannot be overtaken. `WRONG_CALLS.md` 08-14.
- **`deferred` is in NEITHER status list** — not terminal, not closed. "Parked" does not mean "nothing will touch this".
- **A passing mock cannot assert a negative — MUTATE.** And a test that pins a *unit* says nothing about whether its *caller* passes the argument: the other lane deleted their compositing loop and the whole package stayed green.
- **Never run `run-migrations.sh --apply` on this tree.**
