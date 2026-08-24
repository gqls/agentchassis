# What needs your ruling — `bugs_open/326` and its neighbours

**Rewritten 2026-08-24 (second version) after you asked for a review before deciding.**
The first version, written the same day, contained three errors that would have pushed you
towards the wrong option. They are listed first, because you should know what changed and why.
Every count below carries the date it was counted.

---

## What the first version got wrong

**1. "The damage is growing ~26/day" — false.** I subtracted two snapshots taken a day apart
(635 → 661) and called the difference a rate. Re-measured by `created_at`: all 26 are ONE item
type (`contrast_failure`), on TWO sites, from TWO detector sweeps — 22 rows on `noted.co.uk` at
`12:29:23` and 4 on `idea.uk` at `22:23:49`, **same-second timestamps, one sweep each**. Over the
brake's own 7-day window the rate is **~8/day across all types**, and the part that belongs to
this bug (action requests) is **~3.4/day**. This is exactly the error the other lane named
yesterday — *if the evidence is a count, the claim must contain the count* — reproduced in a
document addressed to you.

**2. "~570 keys become live work" (option A's cost) — wrong by about 70×.** 570 is the number of
keys *armed* to two-strike if a repeat arrives. It is exposure, not volume. The actual extra load
under A is the rows that would otherwise be born dead — **~8/day** — each dispatching once after a
12-hour delay. And the landfill would **shrink**, not grow, for a reason below.

**3. "Option B is close to a no-op" — wrong for detectors.** A detector that *wants* the brake but
not the burial has no lever today; B would give it one. B is still useless for the population
326 is about (callers nobody has classified), but it is not a no-op.

**4. The RFC conflates two mechanisms with different problems.** This is the one that reshapes
the decision, so it gets its own section.

---

## The picture after re-measurement: two arms, two different problems

The brake in front of every keyed work-item insert has two arms. Yesterday's RFC treated them as
one thing. They are not.

### Arm A — "finished under 3 hours ago" → the request is DROPPED. No row, no error.

**This is 326's bug.** The customer re-submission that reported success and did nothing hit this
arm at 2h28m. It is unconditionally bad: **no caller, of any kind, wants its request destroyed
with nothing recording that it existed.** Its damage is structurally unmeasurable after the fact,
because it leaves no row.

- Customer path: **fixed** (migration 572, five build-chain steps declared).
- Still exposed: **14** config steps (2026-08-24, the census names them) and roughly **10** Go
  producers that file action requests with a key and no declaration (`emit_*`, `flag_*`,
  `seed_*`, `render_*`, `rerender_*` — 25 files call the helper, 4 set the flag; the rest split
  between action requests and detectors, and I have not audited each one).

### Arm B — "finished twice this week" → the row is born `unresolved`, which nothing picks up.

**This is the landfill, and it is two problems, not one:**

| | rows (2026-08-24) | what it actually is |
|---|---|---|
| action requests born dead | **431** (65%) | a *classification* failure — the brake should never have run on these |
| detector findings born dead | **230** (35%) | the brake **working as designed**: a detector re-found a fault the fixer said it fixed |

The 431 splits again by who filed them:
- `improve_tool` **205** — `tool-auditor`. **Historical.** Newest 08-15; the step now carries a
  per-item key suffix (the `bugs_open/321` fix), so the collision that fed this stopped.
  `[dates MEASURED; causal link INFERRED]`
- `page_rerender` **212** — the **discovery sweep, in Go** (`created_by`
  `completeness-discovery-agent` / `generic`). **Ongoing, ~3.4/day.** The config census cannot
  see it; only a code change reaches it.

The 230 detector rows are a different bug entirely. `contrast_failure`: **200 `complete` + 26
`unresolved`** by `css-patch-agent` in 7 days — the fixer reports done, the fault persists, the
detector re-finds it, the brake fires. That is the brake doing its job, and the defect underneath
is **`bugs_open/352`** (the finding names a selector that matches nothing, so the patch applies
to nothing and reports success). Neither option here fixes it, and option A would re-dispatch that
futile patch every 12 hours.

**And the duplication is the landfill's real disease:** 661 rows over only **247** distinct keys
— **2.68 dead rows per key**, worst key **91 rows in two days** (`audit_fix_gamesdesign.co.uk`).
Because `unresolved` sits *outside* the dedup index, every re-detection lands a fresh corpse.
That is why the pile grows even where the brake is right.

---

## The decision, re-costed

**What it is.** Whether the door that files work should be allowed to destroy a request, and
whether the row it parks should hold the dedup slot.

**What your rule says.** New authority on a shared mechanism ships opt-in per caller, unsafe side
OFF (2026-08-02 §2). The council's guardian vetoed the original change on exactly that shape,
and because 572 had already closed the customer bug without it. Both grounds stand.

**How the options measure now:**

| | fixes Arm A (326's class) | stops action-request landfill growth | fixes duplication | touches detector semantics | shape vs §2 |
|---|---|---|---|---|---|
| **A** defer both arms, fleet-wide | ✅ all callers | ✅ | ✅ (deferred row holds the slot) | ⚠ re-dispatches genuinely failed fixes every 12h | default change — the vetoed shape |
| **B** opt-in defer per caller | ❌ unclassified stay exposed | ❌ | only where opted in | ✅ gives detectors a choice | ✅ exactly §2 |
| **C** census → zero, then mandatory | config only | config only — **Go `page_rerender` bleed continues** | ❌ | ✅ untouched | ✅ no mechanism change |
| **D** defer **Arm A only** | ✅ all callers | ❌ | ❌ | ✅ untouched | narrower default change; **Arm A has no legitimate use, so the "safe side" is unambiguous** |
| **E** set the flag on the ~10 Go action-request producers | Go sites only | ✅ stops `page_rerender` | ❌ | ✅ untouched | ✅ per-caller, exactly §2 |

**None of A–E fixes the 230 detector rows** — that is `bugs_open/352`'s class. **Only A fixes the
duplication.**

**My view, as a view:** **D + E now, C alongside** — D is the smallest change that ends silent
destruction for everyone and has the strongest safe-side argument (nobody wants Arm A's
behaviour); E is the §2 shape and stops the only live action-request bleed. That leaves the
duplication for `RFC_010` / `033` D2, where it already lives, and the detector rows for `352`. If
you would rather take the duplication too, that is A, and its cost is ~8 extra dispatches a day —
not 570 of anything.

Where it lives: `architecture_review/RFC_048_…`, with the patch beside it. The patch is A-shaped;
D is a subset of it (delete the two-strike branch of the deferral, keep the within-cycle one).

---

## Decision 2 — migration 573 is SEPARABLE from decision 1, and I had that wrong too

**What it is.** `573_…_HOLD.sql` makes the front door fail loudly when a submission genuinely
queues nothing. It needs one config key (`on_dedup`) that only exists in the RFC_048 patch.

**What I said yesterday:** decline decision 1 ⇒ delete 573. **Wrong.** `on_dedup` was the RFC's
*second* edit, and it is independent of the deferral. The guardian objected to it at **low**
severity (that "zero consumers" was asserted, not shown); the architecture seat said it
**passes RFC_022 cleanly**. The objection is now answered with a query: **0** live workflow
conditions branch on `deduped`/`inserted` (2026-08-24).

**So 573's real options:** resubmit `on_dedup` alone as an ordinary opt-in field (a small council
round, likely approved), then apply 573 after the roll; **or** decide you do not want a loud front
door, and delete 573. **It does not wait on decision 1.** What it must never do is sit as a stale
`_HOLD` — that is the trap cleaned up on 524 yesterday.

---

## Decision 3 — the 14 undeclared config steps: only a name needed

`./scripts/audit-undeclared-recurrence.sh` lists them. I did not classify them — it is judgement
inside other lanes, and the draft that tried found its own counter-example (`claims-auditor`
genuinely needs the counter). One lane sweeps with owners consulted, or each lane runs the audit.
Nothing is blocked.

## Decision 4 — the 661 dead rows: already yours, now better understood

Still `RFC_010` / `033` D2's decision. What is new: 205 are historical and safe to close; 212 are
`page_rerender` a re-render would trivially resolve; 230 belong to `352`. And the pile is 247
keys wearing 661 rows, so "drain it" is a smaller job than the headline number suggests.

## Decision 5 — the `345` ownership collision: nothing broken

Code committed, green, inert. The other session is the council submitter. Only question: name one
owner so the next instruction does not fork.
