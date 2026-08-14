# HANDOFF — bug 122 lane. START HERE. Written 2026-08-14 (afternoon).

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
stylesheet had re-rendered. It now rests on **one** — and that one is nobody's to hold. **Any lane
re-rendering any of those 14 sites, for any unrelated reason, changes their link colours, and nobody
has to intend it.**

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
