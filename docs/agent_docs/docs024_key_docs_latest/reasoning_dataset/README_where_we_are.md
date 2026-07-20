I surveyed the whole platform rather than just the fix loop, and it reframes the project. Written up in reasoning_dataset/PLAN_capture_gaps_and_volume.md (628467b7d).

We were sizing this off the wrong thing

The fix loop is 445 rows. The platform has logged 40,785 LLM calls and holds 5,292 work items with terminal outcomes. The reasoning is already being produced at roughly 100× the rate we planned to harvest it.

What's missing isn't reasoning. It's the join between a decision and what happened next — and it's missing for a boring, fixable reason.

┌──────────────────────────────────────────┬──────────────────────────────────┐
│                                          │              today               │
├──────────────────────────────────────────┼──────────────────────────────────┤
│ LLM calls logged                         │ 40,785                           │
├──────────────────────────────────────────┼──────────────────────────────────┤
│ …carrying a work_item_id                 │ 2,894 (7.1%)                     │
├──────────────────────────────────────────┼──────────────────────────────────┤
│ …joinable to a terminal outcome          │ 1,165 — already 15× the fix loop │
├──────────────────────────────────────────┼──────────────────────────────────┤
│ items ever independently verified        │ 9                                │
├──────────────────────────────────────────┼──────────────────────────────────┤
│ human decisions with the reason recorded │ 0 of 316                         │
└──────────────────────────────────────────┴──────────────────────────────────┘


The gap that matters most

llm_call_log already has a work_item_id column and the logger already populates it when it's present. It usually isn't:

- content-quality-auditor — 7,119 calls, 0%
- visual-design-auditor — 4,032 calls, 0%
- site-review-agent — 3,987 calls, 0%
- feed-triage — 423 calls, 0%
- tool-recreation-handler — 181 calls, 100%

That last row is the tell: it's plumbing, not physics. ~15,500 calls where an agent formed a verdict and we can't tell whether it was right. One field threaded through four agents, no migration, no new LLM spend,
order-of-magnitude more labelled data. Everything else on the list multiplies whatever this produces, so it goes first.

Five more, briefly

- Human decisions keep the status and throw away the reason. 316 items resolved by a human; approved_by 0, resolution_path 0. Both columns already exist. "Rejected" teaches nothing; "rejected because it named a
  destination that doesn't exist" teaches a lot.
- complete is self-reported. 4,583 complete, 9 ever verified — all one item type, none since 14 July. Training on complete trains on the agent's own say-so, which is exactly the bugs_open/012 failure.
- Free signals unused. attempt_count: 60 items hit 3 attempts and 44 of those are stuck. Hard cases with ground-truth negatives, already recorded, currently invisible to any export.
- feed-triage is the sleeper. It already emits {score, reason, credibility, credibility_reason, source_tier, flagged} and batches many items per call — 423 calls carrying thousands of judgements, running daily since
  March. Best-shaped non-loop source on the platform and it wasn't on anyone's list. The three big auditors return prose instead; porting the loop's cite-or-abstain contract to them would make ~15k calls/month
  training-shaped.
- The log is lossy. LogLLMCall is fire-and-forget with a 5s timeout — rows vanish under load, and load correlates with the interesting runs. That's undetectable corruption after the fact.

The honest ceiling

Even with all of it, this platform makes hundreds of decisions a day, not millions. Six-month realistic scale is tens of thousands of outcome-labelled steps — fine-tuning and eval scale, not pretraining. Worth doing,
because a few thousand outcome-verified domain traces with real negatives is genuinely scarce; but the goal should be stated as a high-quality specialist corpus, not a big one.

Cheapest volume lever once the plumbing's fixed: replay /bugs_open/ — 20 cases with documented root causes and verification steps, so N bugs × M models generates graded trajectories with the answers already written
down.

One correction I should flag: I tried to measure "fix didn't hold" via recurrence and my query was wrong — a naive self-join reported 387,301 recurrences for page_rerender, which is the join exploding, not a finding.
It's in the doc as an open lead with the bad number explicitly not quoted.

Gaps 1 and 2 are platform/ changes, so not mine to make — they'd go to the owning threads through the council gate. Want me to draft those two submissions?

---

## 2026-07-18, later — drafting the submissions killed both of the gaps above

Two of the claims in the entry above are wrong. Leaving them there, as the rules
say, and correcting underneath.

**CORRECTED — "the gap that matters most … it's plumbing, not physics" is wrong.**
The 0% figures are real, but I read them as a field being *dropped* when it is a
field that was never *there*. Those auditors are producers: they raise work items,
they are never handed one to work on, so there is no id at that moment to
propagate. `tool-recreation-handler` shows 100% only because it is on the
dispatch path, which injects the id — a path none of the auditors is on. And
`feed-triage` never touches work items at all; it writes to a different table
entirely, so it would have read 0% for ever no matter what anyone changed.
Including it was my error. The join we want runs the *other* way — from the item
that got created back to the run whose reasoning created it.

**CORRECTED — "human decisions keep the status and throw away the reason" is
worse than stated.** I said the reason columns are empty. Then I checked whether
the reason was being kept somewhere else, and found the handlers do write a
resolution into a JSON blob — and that blob is empty too. Not 0 of 316. **0 of
4,570.** The human-resolution path has never once been called. Those admin
routes are live and nothing invokes them, and a third one was written, finished,
and never wired up at all. So my proposed fix — "record the reason properly" —
would have changed nothing, because nobody walks that path. The real question is
yours, not a data one: those 278 items waiting for human review are a queue
nobody works. Is that intended?

What I got out of being wrong twice: both errors came from reasoning about the
system from its *numbers* without reading its *code*. Both were caught before
anything shipped, by reading the code when I sat down to write a change plan.

**What went to the council instead.** Two submissions:

- **A — record where a work item came from.** Add one column holding the id of
  the run whose findings raised it. That is the corrected version of the first
  gap, pointing the right way round. It makes about 15,000 auditor judgements a
  month checkable against what actually happened to them, and separately it would
  give the first per-agent accuracy signal this platform has ever had — today an
  auditor that cries wolf twenty times looks identical to one that is right.
- **B — actually check that fixes worked.** This one is the better find. There is
  already a mechanism that re-checks whether a defect is really gone before
  marking an item done. It was built properly, it works, and it has been switched
  on for exactly **one** kind of item. Everything else takes the agent's word for
  it. That is why only 9 items in the platform's history have ever been
  independently verified.

## 2026-07-19 — both fired, both came back REVISE, and the council was right

I submitted both. Each takes about two minutes and pulls in whichever reviewers
match the files you touched — 8 reviewers on A, 10 on B.

**Both came back "revise", and I think that is the system working rather than
failing.** Neither was rejected; both got specific, checkable objections. I
checked every factual one against the live database and the code before
responding, and the reviewers were right on all of them.

On **A**, the sharpest objection was one I would not have caught. I had planned
to store the run id as a strict UUID and quietly write "unknown" whenever it
didn't parse. The reviewer pointed out that this can only ever turn a *real*
origin into something indistinguishable from a missing one — silently
undermining the exact link the change exists to create, while looking like
perfectly normal missing data. It is now stored as plain text, which cannot
fail. A second objection caught me widening the change onto a second code path I
had never actually justified; I have removed it.

On **B**, a reviewer found a genuine bug in the *existing* code that I was about
to copy. The current check treats "the component has vanished" as "the problem
was fixed" — but a vanished component is equally the signature of this
platform's most common failure, where a rebuild silently deletes content. So a
content-loss incident currently reads as a success, and I was about to spread
that blind spot to two more checks while citing it as the precedent to follow.
The revised plan fixes it at source: a missing target now records "could not
verify" instead of "verified". Nothing wedges — the item still completes — but
the false success becomes a visible unknown.

Two other things B got right: I called one item type "highest-volume" on figures
that were a 30-day window quoted as if a total (it is 180 items, the largest of
my three, but eighth overall). And one of my three picks — checking broken image
URLs — turned out to have **4 items and zero completions**, so the one choice
carrying real risk (an outbound network call on a hot path) had no volume to
justify it. Dropped. A reviewer also noted my anti-recurrence measure was a code
comment, which would not stop the next 46 item types doing the same thing; it is
now a test that breaks the build.

**Where that leaves us.** Both revisions are written and going back for a second
round. Neither change is mine to implement — they belong to the threads that own
that code; my job ends at a plan good enough to hand over.

The extraction work itself — the thing that actually produces the dataset — has
not started, and is not blocked by any of this.

**One thing I should own.** This thread has now made four wrong calls (two above,
two more in the technical log). Every one came from the same habit: reasoning
from data or documents without opening the code underneath, and being confident
while doing it. All four were caught before anything shipped, but three were
caught by checks this repo already mandates and I had skipped — and the fourth
was caught by the council, not by me.

---

## 2026-07-19, later — six council rounds, one wasted, and where I've stopped

Both went back for second and third rounds. Here is the honest state.

**A is nearly there.** It started with five objections from two reviewers, came
back with two (both procedural — name how you'll verify the deploy against the
running pod, and ship a rollback file with the migration; both added), and then
on the third round pulled in a reviewer that hadn't fired before, because the
plan now touched a database migration. That reviewer asked one genuinely good
question: there is already a "batch" column on the same insert — have you shown
it isn't already doing the job you want your new column to do? I checked. It
isn't: it's a fresh random id per run, tied to nothing, so it can group an audit
run's items but can't tell you *which* run made them. The objection doesn't
stand — but the reviewer was right that I'd assumed it rather than shown it.

**B is not converging, and the reason is fair.** The objection count stayed at
eight across both rounds. The one that stung, correctly: one of my three edits
was described as "a stub dressed as an edit" — I'd written a paragraph of
comments explaining what the function should do, where the other two edits had
real code. That's exactly right. A reviewer also caught my plan contradicting
itself (I quote a figure showing nine items in a particular state, and elsewhere
assert nothing in the code produces that state — both true, but I never
reconciled them, so it reads as an error).

**I've stopped at six runs, and I think that's the right call.** B's remaining
objections would be answered by writing the actual code and listing out
forty-seven item types — which is implementation work in code this thread isn't
allowed to touch. Iterating further would mean doing the owning thread's job
inside a JSON file. A is two small additions from likely approval, if you want
one more round.

**What the exercise taught about the gate itself**, which I think is worth more
than either submission: it earns its cost on the *first* round. It caught two
real design defects — a silent-failure mode I'd built into A, and a genuine bug
in existing live code that B was about to copy. By the third round it's mostly
enforcing plan completeness, and a plan detailed enough to satisfy it is nearly
the change itself. Submit once, take the design hits, then hand over.

**And a fifth wrong call, this one with a price.** Twice the runs didn't show up
in the database after I fired them. I waited about seven minutes, concluded the
platform had silently dropped them, checked the one documented cause (didn't
apply), and re-fired one — spending a council run on a duplicate. Every single
one of those runs landed. They just take several minutes to appear when the
platform is busy, and it was. Same mistake as the other four in a different
costume: I concluded a *mechanism* from an *absence*, without waiting long enough
to know. The other four cost credibility; this one cost money.

---
## 2026-07-19, end of day — handed over

Both are now with the threads that own the code.

**B went to the work-item-completion thread**, and that placement is itself worth
a note: I first assigned it, from the code's git history, to the thread that
originally built the verifier. Reading the workstream documents rather than the
commit log showed a different thread owns this class — and its own plan, written
the day before, **already names this exact gap, down to the same item type I'd
picked**. So the handoff went there, with a pointer left for the thread that built
the verifier, since the defect sits in their code and they may hold context we
don't. (Sixth wrong call of the workstream, same shape as the rest, caught in
about two minutes and for free.)

**Along the way I filed the bug the council found.** The verifier that checks
whether a defect is really gone reports *success* when the thing it's checking has
vanished — and a vanished component is exactly what happens when a rebuild
silently deletes content. So a content-loss incident currently gets recorded as a
verified fix, by the mechanism built to stop us trusting self-reported successes.
That's `bugs_open/032`, with a conservative fix that leaves item flow unchanged
and turns a false success into a visible "couldn't check".

I did *not* file a second bug for the coverage problem — that this verification
mechanism has been switched on for one item type out of about fifty. Another
thread had already filed that exact pattern yesterday, from a council objection of
its own, so I added mine as an instance underneath rather than starting a rival
account. The general lesson also went into the debugging guide so nobody re-walks
it.

**A has no owner, and that's the honest blocker.** The file it changes has no
active thread and no workstream tracking it — its last meaningful commits are
generic and old. Rather than dump it on the nearest plausible thread, I've filed
it where the motivation lives and flagged it for you to assign. It's two small
edits from likely approval, and both are written out so whoever picks it up
doesn't re-spend the council rounds.

**Twice today my documents were swept into other sessions' broad commits.**
Nothing was lost and I've verified the content landed intact, but it's worth you
knowing it happens — the repo's own guidance says committing narrowly protects
others from you, not you from others, and today was a live demonstration.

**What's actually left for this workstream:** the extractor. None of the above
blocks it. That's the piece that produces the dataset, and it hasn't been
started.

---
