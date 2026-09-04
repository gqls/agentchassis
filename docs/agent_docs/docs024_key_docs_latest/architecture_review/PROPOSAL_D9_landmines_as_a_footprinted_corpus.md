# PROPOSAL D9 — landmines as a footprinted corpus, not prose in an auto-loaded file

**Status: DRAFT — proposal only. Nothing built, no config or code changed.**

**Provenance.** Written 2026-07-27 by session *"bugfix 61"* (the `bugs_closed/061`
med-scrape thread) at the owner's direction: *"with regard the landmines, we have the
council and the missteps list and we could also add landmines, see the architecture
council thread."* Handed to this thread rather than applied, because **D8 is this
thread's finding and this extends it**, and because `DECISIONS_open_for_owner_…` was
being actively edited (14:00 today) while this was written. Fold it in as D9, reject it,
or bin it — it is yours. A one-line pointer has been inserted at the end of §5 of that
document; nothing else there was touched.

---

## 1. The finding: D8's defect has a second symptom, and nobody had connected them

**D8 established** that the council structurally cannot read the misstep record —
`code_checks` indexes Go only (4,535 symbols, **0 markdown**; `WRONG_CALLS.md` has 0 rows
in `code_symbols`), and the SQL `checks` reach ten tables, none a document store. So
`WRONG_CALLS.md`, `/bugs_open/`, `/bugs_closed/` and every working doc are invisible to
every seat.

**The second symptom, measured 2026-07-26/27:** the auto-memory index grows without
bound for the *same underlying reason*. `MEMORY.md` auto-loads in full into every thread,
so a landmine that must be seen has to live there — because prose that cannot be queried
can only be delivered by broadcasting it. From the memory repo's own git history:

| time (07-26) | `MEMORY.md` bytes | |
|---|---|---|
| 22:11 | 20,613 | |
| 22:15 | 17,621 | ← compaction #1 (~3 KB cut) |
| 23:01 | 20,575 | **re-inflated past its pre-compaction size in 46 minutes** |
| 23:05 | 17,573 | ← compaction #2, by a different session |
| 07-27 11:51 | 17,978 | creeping again |

Two compactions inside one hour, by two sessions, the first fully undone before the
second. Composition today: **76 entries, mean 223 chars, only 2 over 400** — so this is
*uniform* bloat, and trimming the worst offenders provably cannot fix it. `MEMORY_closed.md`,
the archive created by the 2026-07-24 lifecycle split, is now **24 KB — larger than the
index it drains**, and near the same ~24.4 KB read cap.

**One defect, stated once:** *hard-won knowledge lives in prose; prose cannot be queried;
what cannot be queried must be broadcast; broadcast has no upper bound.* The memory index
is the broadcast channel, and it is now the binding constraint on both problems.

## 2. There is a ladder, and its middle rung has no mechanism

- **`WRONG_CALLS.md`** — raw incidents, append-only, human-read. One row is an anecdote;
  the file's own thesis is that the value is *"the tally of the cheap check that would
  have caught it — because a check that keeps appearing is one worth automating."*
- **A landmine** — the distilled check, attached to the thing it guards, delivered when
  that thing is touched. **No mechanism exists for this rung.** It is improvised as prose
  in `MEMORY.md`, and that improvisation *is* the bloat in §1.
- **A lint or a council seat** — full automation, once the tally justifies it. Exactly how
  `check_append_only_docs` earned its place in `scripts/pattern-check.py`.

`WRONG_CALLS.md` already states the promotion rule for rung 1 → rung 3. Naming rung 2
gives landmines a home that is not an auto-loaded file, and — the point for this thread —
gives the council something it can actually read.

## 3. Reuse before building: the store already exists and is already used this way

`public.doc_notes` — **370 rows today**:

| column | why it fits |
|---|---|
| `subject_type` + `subject_key` | indexed together (`idx_doc_notes_subject`). `subject_key` is the natural **footprint**: the path, table, or symbol the landmine guards |
| `categories` jsonb | GIN-indexed (`idx_doc_notes_categories`) — cross-cutting tags |
| `body` | the landmine text |
| `source_agent`, `created_by`, `created_at` | provenance and staleness, which `MEMORY.md` lines do not carry |

**And the categories in use are already landmines in all but name:**
`do-not-lock-derived`, `envelope-vs-payload`, `derived-fields`, `bug-020`, `bugfix-056`.
This is not a new idea being proposed — it is an existing practice that never got a name,
a convention, or a delivery path.

**Relevance-gating is likewise already proven here:** the gate fires two seats always and
the rest only when the submission's edited paths match their footprint (CLAUDE.md; a real
submission drew 10 of 16). That is precisely the delivery model a landmine needs — a
session touching `vet_med_price_scrape_action.go` should receive that file's three
landmines, not all seventy-six.

**Distinct from `bugs_open/108`** — do not merge them. 108 fixes the **code** index
(`code_symbols`: stale-but-FRESH, and no function bodies). D9 concerns the **prose**
corpus, which `code_checks` will never cover: D8 established it is Go-only *by design*,
not by defect. 108 makes the code index tell the truth; D9 gives the seats a second
corpus that has never existed. They are complements.

## 4. Proposed decisions

**D9(a) — Name the ladder** (incident → landmine → automation) in `WRONG_CALLS.md`'s
header and in `016b` §9, so the middle rung stops being improvised. Costs nothing;
prevents the next person doing what we have all been doing.

**D9(b) — Landmines become `doc_notes` rows**, `subject_type='landmine'`, `subject_key`
= the guarded path/table/symbol, `categories` for cross-cutting tags. No migration
needed; the table, both indexes and the practice already exist.

**D9(c) — Delivery to the council: add `doc_notes` to the schema hint.** This is D8a's
original proposal, which **D8a′ superseded only for the *minutes* case** (`council_report`
in `diagnosis_artifacts`, applied and live today). The landmine case is still open, and
it is a config change — live immediately, no image, no roll.

**D9(d) — Delivery to a *session*: OPEN, and the weak link. Do not let (a)–(c) be
approved as though this were solved.** Nothing currently queries `doc_notes` at session
start. Three candidates, none costed:
  1. a hook that queries landmines for paths in the session's opening diff;
  2. a standing discipline ("query landmines for what you are about to touch");
  3. piggyback the existing memory-recall mechanism, which surfaces topic files by
     `description:` frontmatter — but `doc_notes` rows are not memory files, so this may
     not reach them at all.

  **(2) is the tempting one and I recommend against it standing alone.** "Detail does not
  live in the index" is *already the written rule* in `memory-index-how-it-works.md`, and
  §1's measurements are what a discipline achieves against a standing incentive. Authors
  put landmines in the auto-loaded file because it is the only thing guaranteed to be read;
  that incentive survives any amount of exhortation.

## 5. Risks and what I could not verify

- **Staging risk, the serious one.** Moving landmines out of `MEMORY.md` before (d) is
  solved would *remove* protection sessions have today in exchange for protection they
  might get. Landmines in the index have genuinely prevented cross-workstream mistakes.
  **Sequence must be: build delivery, run both in parallel, measure, then drain the index
  — never drain first.**
- **[UNVERIFIED] the council's footprint map location.** I could not find it under
  `council_decide.config` (keys there are `error_step`, `max_rounds`, `review_fields`,
  `hard_veto_from`, `fix_correlation_id`), so it presumably lives per-`review_field` or in
  the mirrored `fix-proposer` row that `099_SYNC_gate_roster.py` copies. §3's claim that
  the gating mechanism is reusable rests on CLAUDE.md's description of the behaviour, not
  on my having read the config.
- **[UNVERIFIED] whether `doc_notes`' existing 370 rows need partitioning or curation**
  before they are exposed to seats, and whether `subject_key` granularity (file? function?
  table?) actually matches how landmines are phrased.
- **[UNMEASURED] migration cost.** The 76 index entries are prose; converting them to
  footprinted rows is manual curation, not a script. Nobody has sized it.
- **The same test D8a′ set for itself applies here:** the honest measure is not that the
  rows exist, but whether a seat's `checks` array ever cites them and whether a session's
  behaviour changes. Present text is not use.

---

## 6. EVIDENCE ADDED 2026-08-23 — two sessions hit one documented landmine within an hour, in opposite directions

Contributed jointly by the `bugs_open/326` and `loanzy.uk` lanes. It is offered as a
**measurement of the delivery path**, which §5 says is the honest test: *"Present text is not
use."*

### What happened

`LANDMINES.md:3248` has carried, since **2026-08-02**, an entry titled *"`orchestration_states`
keeps terminal rows ~24 HOURS — and `min(created_at)` says 20 days, because the statuses it
reaps are not the ones that set the floor"*. It is not a thin entry. It gives the measured
window, explains the false floor, prescribes bounding **per status**, and instructs the reader to
re-source durable claims from a table with no retention job — **naming `site_specs` and
`site_work_items` specifically**. Two further entries (`:2571`, `:7737`) cover the same table's
retention from other angles.

On the evening of 2026-08-23, within about one hour:

- the **`loanzy.uk`** lane wrote that `vertical-exemplar-researcher` had *"four runs in its
  entire history"* — that being the 24-hour retention window, not a history;
- the **`bugs_open/326`** lane (me) generalised a `0/4` from the same table into *"that branch
  has **never** fired"*, and hardened it into *"a defect"*.

Each caught the other's version. Neither caught their own, and **neither had read the entry**.
The recovery in both cases was exactly the move the entry prescribes — re-sourcing from
`site_work_items` + its archive, which gave **32 items across 27 sites since 2026-04-02** against
a visible four.

### Why we think this is mechanical, not two careless readers

- The `SessionStart` hook surfaces landmines by matching footprints against **files already dirty
  in the working tree**. This entry's footprint is a **TABLE**. A table cannot match a path, so
  the entry is **structurally unable to auto-surface** — not unlucky, ineligible.
- The only delivery route left is grepping `LANDMINES.md` for the table before querying it.
- `MEMORY.md` carries exactly that instruction — *"grep LANDMINES for the SYMBOL you are about to
  trust — the hook only matches files already DIRTY"* — and it auto-loads. Both sessions had it
  in context. Neither acted on it.

**A guard that is correct, well-written, indexed, and reachable only by a discipline nobody
performs is indistinguishable in its effects from an absent guard** — with the added cost that
its presence makes the gap invisible to anyone auditing coverage.

### The sharper version of the miss, from the 326 lane

I did not merely fail to grep. **The signal was in my own output and I read past it.** At ~17:00Z
I ran a whole-table grouping over `orchestration_states` and it returned
`build-dispatch-loop | 777 | min 2026-07-24` — a **month-wide floor** on a table that retains a
day. That is precisely the false-floor shape the entry describes, sitting in my terminal, two
hours before I made the claim it would have prevented. The entry would have told me what I was
looking at; nothing connected the two.

*(Bound on that observation: a per-status query at 19:40Z showed nothing older than 08-22, so
those rows had been reaped in the interval. That does not change the point — I had the anomalous
figure and did not treat it as a retention signal — but the month-wide floor is not reproducible
from the table now.)*

### What we did NOT do, and why it bears on D10

**We did not file a second landmine entry.** Writing the guard again does not fix the thing that
is broken. The failure is delivery, and a duplicate would make the corpus worse — one more entry
competing for the same unread channel, and one more thing for a coverage audit to count as
protection.

That reasoning is D10's own argument, arrived at independently by two lanes under pressure. **If
the proposal wants a case where the prose corpus demonstrably failed at its job, this is one**:
the guard existed, it was right, it was specific, it named the remedy, and it reached nobody.
Both incidents are in `WRONG_CALLS.md` under 2026-08-23, from both sides.

**The one design note we would add to §4:** footprinting by table would have fixed *this* case
only if the trigger fires on **querying** the table, not on editing a file that mentions it.
Neither of these sessions had `orchestration_states` in a dirty path — we were reading it live
through `psql`. Whatever D10 becomes, the delivery moment for a table-footprinted entry is the
first query against that table in a session, and we do not know what mechanism could catch that.
Stating it as an open problem rather than pretending the footprint alone solves it.

### 6a. A costed option for the §6 open problem — NOT BUILT, and deliberately so

§6 ends on an open problem: a table-footprinted entry's delivery moment is the **first query
against that table**, and nothing currently catches it. There is a concrete mechanism, and the
repo already runs one of exactly this shape.

**The shape already exists.** `scripts/block-git-stash.py` is a `PreToolUse` hook on Bash, wired
in `.claude/settings.json`, that inspects the command *about to run* and acts on it however the
command is compounded. The same shape fits here: match a Bash command containing `psql`, extract
table-shaped tokens from the SQL, look them up against `LANDMINES.md` **footprints**, and print
any match as advisory context **before** the query executes. That is precisely the moment both
sessions missed on 2026-08-23. It needs **no new corpus format** — it reads the footprints D9
already prescribes — so it is evidence *for* this proposal rather than a competing mechanism.

**Two constraints, and the second is load-bearing** (both from the `loanzy_uk_example_site` lane,
which raised the option):

1. **Once per table per session, not per query.** `psql` runs constantly here. An entry reprinted
   forty times becomes wallpaper — and wallpaper is the exact failure mode §6 documents. *Present
   text is not use*, and present text repeated is worse than absent text, because it trains the
   reader to skip the channel.
2. **Match FOOTPRINTS, not free text.** Grepping the whole file for a table name hits every entry
   that merely *mentions* it — three entries touch `orchestration_states` from different angles,
   and firing all three on every query is how a reader learns to ignore them. This constraint is
   what makes the option depend on D9's footprinting rather than substitute for it.

**Neither lane has built it, and neither will on the other's suggestion.** A `PreToolUse` hook is
a harness/config change: it means editing `.claude/settings.json` and intercepting **every session
on this machine**, not just the two that found the problem. A peer session asking is not authority
for that — it would be taking an instruction from a session rather than from the owner. So it is
recorded here as a costed option for the owner to rule on, and both lanes have flagged it to
theirs.

**What would make it worth building, stated as a test rather than an assertion:** the honest
measure is not that the hook fires, but whether a session's behaviour changes — the same bar §5
sets for the corpus itself. The 2026-08-23 pair is the baseline to beat: guard present, correct,
specific, naming the remedy, and reaching neither of two sessions actively looking at the table.

### 6b. A THIRD instance, and it breaks the framing of §6/§6a: this warning WAS delivered, on time, and still failed

Contributed 2026-09-03 by `inline_guide_imagery` and `dartsonline_traffic`, who hit it independently
in one afternoon. **It is not another delivery failure. It is the case where delivery worked and the
outcome was identical**, which is why it is worth adding rather than counting as a third tally mark.

**What happened.** The `SessionStart` hook printed the landmine *"`git diff | grep '^-[^-]'` cannot
see a deleted markdown BULLET"* to this session at start-up — **first of the six entries it
displayed**. It was read. Three hours later, checking whether a ledger edit had clobbered a peer's
entry, that session ran `git show <sha> -- LANDMINES.md | grep '^-[^-]'`, got **nothing** against a
`--numstat` of **1 deletion**, and briefly accepted the reassuring answer. The same entry had been
printed to the other lane at the top of *their* session too.

**Why §6a's mechanism would still have caught it, and why that matters.** The trap fires on a
**command shape** — inspecting a diff with a dash-excluding predicate — and a `PreToolUse` hook on
Bash sees exactly that, before it runs. So this case is **evidence FOR the costed option**, and it
extends it cheaply: match on **command shape**, not only table tokens inside `psql`. The two known
misses then share one mechanism.

**But it also falsifies the assumption underneath §6.** §6 reasons as though the problem is a
warning that *never arrives*. Here it arrived, correctly targeted in content, at the only moment the
current design has — and the failure was identical. **Delivery is not availability.** A warning
lands filed under the artefact it concerns and fires when you are thinking about something else
entirely: the session was not investigating a bullet, it was checking a diff.

**The structural gap, stated as sharply as we can put it — MANY LANDMINES ARE FOOTPRINTED BY
ARTEFACT AND FIRE ON A VERB.** This entry's footprint is a **path**
(`docs/agent_docs/docs026_concept_register/`) and it matched only because an unrelated `.tmp_check`
file in that directory happened to be dirty. **It was delivered by coincidence, three hours early,
filed under a dimension orthogonal to when it fires.** The same is true of others added the same
day: *alt text is not evidence of what an image shows* fires when you **verify imagery**, not when
you touch a component row; *do not gate on "is migration NNN applied?"* fires when you **check a
number**, not when you edit `sql_for_agents/`. A path footprint cannot predict any of those moments.

**So D10 has two axes, not one.** The `subject_key` footprint §4 prescribes is right for *"what does
this entry guard"*. It is not, on its own, an answer to *"when should this entry appear"*. We are
not proposing a schema change — `doc_notes` rows can carry both — only that the two questions be
kept distinct in whatever D10 becomes, because collapsing them is what produced a correctly-written,
correctly-targeted, correctly-delivered warning that reached a reader and changed nothing.

**Neither lane is building anything**, for exactly the reason the section above already gives: a
`PreToolUse` hook is a harness/config change affecting every session on this machine, and a peer
session asking is not authority for it. Recorded for the owner, flagged to ours.

**One empirical note that bears on §5's success test.** Across four measurement failures between
these two lanes in one day — a wrong artefact, a format-assuming predicate, an expectation-sized
time bound, and this dash-excluding predicate — **all four were caught by the OTHER lane re-running
the query, and none by the lane that wrote it.** That is not a coincidence of attention: the
predicate *is* the author's understanding, so re-reading your own query re-applies the assumption
that produced it. If D10 wants a measure of what actually catches this class today, the honest
answer is a second party with the same question and a different encoding — which argues for making
the corpus **cheap to re-derive from** (commands and controls in the entry, as the newer entries
carry) at least as strongly as it argues for better delivery.

#### 6b (i). SHARPENED, and it corrects §6b's own framing: the two axes are NOT collapsed by accident — the verb axis has a mechanism, and its trigger condition is SELF-SUSPICION

`dartsonline_traffic` read the hook's source, which states the design more plainly than either
lane's write-up did. **Verified at the file here rather than relayed**, `scripts/landmines-session-start.py`,
`matches()` — the guard at **:69–70** and the docstring at **:58–65**:

```python
    Substring both ways: a footprint may be a directory ('cmd/') that a path sits
    under, or a full path that equals it. Bare table/command footprints
    ('agent_definitions', 'git commit') will not match a path and are deliberately
    NOT reported here — a session start is about files. Those are found by grep,
    which is what the standing discipline in CLAUDE.md is for.
    """
    hits = []
    for e in entries:
        for fp in e["footprints"]:
            if "/" not in fp and not fp.endswith(".go") and not fp.endswith(".py"):
                continue                      # not a path-shaped footprint
```

And the discipline it hands off to, `CLAUDE.md:566-568` verbatim: *"**Still grep it yourself for
table, command and symbol footprints**, which cannot match a path."*

**So §6b's "collapsing them is what produced the failure" was wrong, and this is the correction.**
The two axes are not conflated. The hook serves the artefact axis **by design, with its reason
written down**, and hands the verb axis to a documented manual step. Nothing here is an oversight.

**Which makes the gap much more precise, and worse.** The verb axis is not unmechanised — it is
mechanised as *a human remembering to grep*, and **its trigger condition is self-suspicion**:
"before touching anything **unfamiliar**". `[MEASURED 2026-09-03]` neither session was doing
anything unfamiliar when it fired. One was checking a diff it had just produced; the other was
editing a style guide it had written itself an hour earlier. **A discipline conditioned on doubt
cannot catch confident competence, and confident competence is precisely the state in which these
traps fire** — that is the same shape as this corpus's recurring finding that a right conclusion
propped up by a count nobody took survives every review.

**The coincidence claim now has two independent instances, one per lane.** In this lane the
`grep '^-[^-]'` entry was delivered because a `.tmp_check` file in `docs026_concept_register/` was
dirty. In theirs, the *same entry* was delivered because a **deleted bug file left dirty by a lane
rename** matched its `bugs_open/` footprint. **Neither had anything to do with diffs or markdown
bullets.** Two sessions, two unrelated paths, the same entry, both delivered hours early and both
ignored at the moment it mattered. That is the artefact axis working exactly as specified and
being orthogonal to the firing moment, observed twice.

**What this does to §6a's costed option.** It strengthens it and narrows what it must do: the hook
is not a replacement for a missing mechanism, it is **the automation of the manual step whose
trigger condition is known to be unsatisfiable** — a `PreToolUse` matcher does not need the session
to suspect anything, which is the whole of the difference. Match on command shape as well as on
table tokens, and the two known misses plus both of today's share one mechanism.

**Still not built, and still not ours to build.** Unchanged from §6a and §6b: a harness change
touching every session on this machine, and a peer session asking is not authority. Both lanes have
routed it to their owners. **What we would flag to whoever rules on it:** the success test in §5
should be run against *unsuspecting* sessions specifically, because a warning evaluated by someone
who already knows to look for it measures the wrong population.

#### 6b (ii). The §5 success test, settled between the two lanes — my proposed criterion was UNRUNNABLE and is withdrawn

**`inline_guide_imagery` proposed "run the success test against *unsuspecting* sessions
specifically". `dartsonline_traffic` refused it and they are right. It is withdrawn.** The concern
behind it stands (a warning evaluated by a reader who already knows to look for it is the
fixed-tree error), but **suspicion is not observable in the artefacts**: nothing a session leaves
behind separates *"did not think to grep"* from *"grepped for an unrelated reason and happened to
be covered"*. Any proxy — *did they grep `LANDMINES` before touching X?* — **is a predicate encoding
an assumption about mental state**, which is the family that produced four errors between these two
lanes in one day. Worse, it reads rigorous and **could never return a failing result**, which is
this corpus's own definition of a measurement that is not one.

**The runnable form, theirs, adopted:**

> After the mechanism ships, do `WRONG_CALLS` entries **of the covered class** stop appearing, while
> entries for traps with **no** hook coverage continue at their prior rate?

It measures outcomes rather than mental states, carries its own control, needs no cooperation from
the sessions being measured, and is blind to suspicion by construction — an unsuspecting session
that avoids the trap simply never files an entry, which *is* the signal.

**Their two stated cautions, kept:** it is a **lagging** indicator (you learn it worked from
absences accumulating, and an absence is weak evidence early), and `WRONG_CALLS` is
**self-reported**, so it counts *caught-and-logged* traps — a mechanism that prevents a trap and one
that stops people noticing they hit one look identical. That is why the no-coverage control matters
more than the headline count.

**A third caution, measured here rather than reasoned, and it changes how the test must be read**
`[MEASURED 2026-09-03; predicate: bullets opening `- **YYYY-MM-DD`, which is the entry form — the
file's other 1,250 top-level bullets are continuation lines within entries]`:

| period | entries |
|---|---|
| 2026-07 (from the 30th) | 6 |
| 2026-08 (whole month) | 27 |
| **2026-09 (three days)** | **41** — 19 on the 2nd, 22 on the 3rd |

**The logging rate rose roughly twentyfold in two days.** So **"continue at their prior rate" cannot
be read as an absolute baseline** — a before/after count would be dominated by whatever is driving
that surge, not by the hook. **The concurrent control is doing all of the work, and the statistic
has to be the RATIO of covered-class to uncovered-class entries in the same window**, which is
robust to the overall rate moving. Read as a ratio the test survives; read as a before/after count
it does not. (Totals confirmed: 74 dated entries against **870** `###` entries in `LANDMINES.md` —
the peer's 869 is same-day drift, entries are being appended hourly, including four today.)

⚠ **And the surge is fleet-wide, not these two lanes inflating their own instrument** — **3, 4 or 5**
of the 41 September entries name either lane, **depending on what "naming a lane" means, and all three
figures are correct for their own predicate** (recorded so nobody later reads them as a contradiction):
**3** matching the exact slugs on the entry's *dated opening line only*; **4** matching the exact slugs
*anywhere in the entry*; **5** on a broader match adding bare `dartsonline` and `grip-styles`.
⚠ **The 3 was mine and it is the narrowest for a familiar reason — my predicate was scoped to the
opening line while my claim was about entries, so it missed one entry that names a lane in a
continuation line.** That is the fifth instance in two days of a predicate answering a different
question from the claim it supports, and it surfaced only because the other lane's number differed
from mine and neither of us assumed the other had slipped. **Whichever figure you take, ~88–90% is
other lanes** and the conclusion is unchanged. So the throughput is real and broad, which is good for
detecting a change. But it introduces the **mirror of caution two**: a hook that raises landmine
awareness generally could increase logging (better noticing) at the same time as it prevents traps,
and those two effects move the covered-class count in **opposite** directions. The uncovered-class
control absorbs the awareness effect only if awareness rises equally across both classes — which is
exactly what a hook targeted at one class would not do. **Stated as a known weakness of the best
available test, not as a reason to prefer the unrunnable one.**

**Both lanes now agree the criterion is the ratio form with those three cautions attached, and
neither is building the mechanism.** Owner's ruling.

### 6c. A FOURTH instance — and it puts a NUMBER on §6a's constraint 2, which was the only part still stated as judgement

Contributed 2026-09-04 by `bugfix_384_page_list_invalidation` and `ai-agent-orchestration`, who hit
one entry from opposite sides within an hour. It is a §6-shaped case (table footprint, no delivery),
not a §6b one — but it is worth adding for two things the first three did not supply.

**What happened.** The entry is *"`content_components.name` AND `.function` DISAGREE ON 336 OF 442
ACTIVE COMPONENTS"*, footprinted on those columns. It records the `resolveComponent` /
`loadComponentSchemas` name-or-function fallback, **both call sites with line numbers**, and the
figures **14 of 16 NULL-`component_id` rows resolve, 2 are stranded**.

- The **384 lane wrote that entry on 2026-09-03.**
- On **2026-09-04** the same lane wrote a census over `page_components` joined to
  `content_components`, keyed on `pc.component_id` alone — the exact error the entry warns about —
  and shipped it to a peer lane and into a runbook twice. It then **re-derived 16/14/2 from
  scratch**, arriving at the numbers already in its own entry.
- The peer made the **same** error independently, on the same row, in the same hour.
- The 384 lane then used `grep 'func resolveComponent'` to tell the peer their (correct) citation
  named the wrong function — a grep that **structurally cannot match a closure**, which is what the
  relevant `resolveComponent` is.

**Datum 1 — the entry now carries FOUR "it existed and I did not read it" notes, two of them from
the lane that wrote it.** §5's bar is behaviour change; here the corpus failed the lane that
authored the guard, twice, on consecutive days. **None of the four was a reading failure.** Nobody
reached the entry to misread it. That is the distinction §6 draws and this is its cleanest case.

**Datum 2 — and this is the contribution: `grep -c` gives constraint 2 a number.**
§6a says *"Match FOOTPRINTS, not free text … grepping the whole file for a table name hits every
entry that merely mentions it — three entries touch `orchestration_states`"*. Three sounded like a
nuisance. `[MEASURED 2026-09-04]` on this file:

| token | lines mentioning it |
|---|---|
| `content_components` | **173** |
| `page_components` | **333** |

**A free-text trigger on either is unusable on its first firing, not on its fortieth** — it would
print a third of the file. So constraint 2 is not a refinement of the mechanism, it is a
precondition for it, and D9's footprinting is what makes §6a possible at all rather than merely
tidier. ⚠ **Note also that the peer had the meta-instruction loaded**: their memory carried
*"grep LANDMINES yourself for the SYMBOL you are about to trust"*, dated two days earlier, at
session start. **Delivering an instruction to grep is not delivering the entry** — the instruction
fired and the behaviour did not follow, which is §6b's finding one level up.

**Two design notes for whoever builds §6a, both from this case:**

1. **Extract EVERY table token in the statement, not the leading `FROM`.** The query that went
   wrong read `page_components` and joined `content_components`; the entry is footprinted on the
   **joined** table. A hook keying on the primary table would have stayed silent.
2. **Command-shape matching (§6b) should include the SYMBOL grep, not only diff predicates.**
   `grep 'func <name>'` over a Go package is a command shape with a known blind spot — it cannot
   match `name := func(…)`, and in this package the closure is the definition that matters. That
   is now its own `LANDMINES` entry; it is listed here because it is a second command shape the
   same hook would catch, and because it was the tool used to *dispute a correct citation*.

**Neither lane is building it**, for the reason §6a already gives: a `PreToolUse` hook is a
harness change affecting every session on this machine, and a peer session asking is not authority.
Recorded as a fourth datapoint for the owner's ruling.
