# SUMMARY — 2026-08-02 — all four reconciliation deletes guarded, live and proven; the lane closes

Second summary in this lane's series, and the last. The first
(`SUMMARY_2026-08-01_site_a_live_and_proven.md`) was written when one of the four
call sites was done. This one is written because the work is finished, the bug is
closed, and the read-out genuinely differs — which is the test for writing one at
all.

---

## What we're trying to do

There is a pattern all over this platform called a *reconciliation delete*: a job
re-writes everything it can see, then deletes whatever it did not re-write, so that
rows for things which no longer exist do not pile up. The pattern is correct and we
depend on it.

It is catastrophic in exactly one situation — **the job did not see everything.**
Then "delete what I didn't re-write" means "delete most of the site". Nothing at any
of the call sites checked whether the run had actually seen the corpus before it
started deleting.

The goal was not to fix one of those. It was to establish a single shared rule —
*prove you saw enough before you are allowed to destroy* — and put every instance of
the pattern behind it.

## Where we've come from

The pattern has bitten before, more than once, and the case files say so: the same
delete-then-rebuild shape destroyed a working pathfinding game on one site, then
recurred independently on a second months later. The version that hurts most is the
quiet one — an emptied index does not visibly break, it just starts answering
"nothing found" with complete confidence.

`bugs_closed/135` built the decision rule and wired it to the first call site.
`bugs_open/165` was filed to say: there are three more, and they are unguarded.

Three things we believed at the start turned out to be wrong, and all three were
corrected by measurement rather than argument:

- **The bug file told us how to slice the data, and it was wrong twice.** For the
  page-sections site it proposed slicing per section name — but almost every page
  has exactly one section of each name, so any legitimate removal would score as a
  100% loss and the guard would refuse for ever. For navigation it proposed slicing
  per nav group — but the classifier deliberately *moves* pages between groups, so
  an ordinary re-homing reads as a total loss of one group. Both partitions would
  have produced a guard that cried wolf until someone deleted it.
- **We assumed the same slicing would transfer between tables. It does not.** A
  table whose contents are *authored* (by an LLM) ratchets: once a truncated run
  writes fewer rows, the next run is handed the same truncated input and agrees, so
  the damage becomes the new baseline. A table that is *derived* (recomputed by
  deterministic code) self-heals: the next healthy run projects the full set again
  and repairs it. Only the first kind needs a second opinion from a different
  source.
- **We nearly shipped a workaround and called it a fix.** More on that below.

## What we've done

**All four reconciliation deletes on the platform now refuse to run when the job
that would trigger them has not seen enough.** One shared rule, four call sites,
each supplying its own measurements — because the rule generalises and the slicing
never does.

- The guard for page sections went live on 31 July and was proven by deliberately
  breaking it.
- The guards for navigation and for the link registry went live on 2 August.
- Navigation was proven the same way: we planted extra rows on a live site so the
  rebuild would look like a big loss, watched it **refuse**, confirmed nothing had
  been deleted, and swept the planted rows away afterwards. The pass side was
  proven for free — a genuine site build had already cleared the guard the day
  before and recorded its numbers, which is better evidence than a synthetic run and
  cost one query.
- The link-registry guard cannot be proven, and we have said so plainly rather than
  letting it look undone. Its table has never held a row, and the agent that would
  have filled it was retired the same day. The guard is unreachable — but it is in
  place for whenever that path is revived, which is the whole argument for guarding
  a dormant delete.

Along the way the induction found a real bug in the shared code: the refusal message
told operators that the rows it declined to delete would be tidied up by a later
run. True for one of the four call sites, false for the other three, which refuse
the whole operation and tidy nothing. That message was being written into a durable
work item that a person reads. It is fixed and live.

**The most valuable thing we did was reject our own shortcut.** Guarding the link
registry created a disproportion: one page's refusal would fail an entire site
build, because the loop it sits in has no way to say "tolerate this one step". We
diagnosed that correctly — and then routed around it by making the action never
report failure. Four independent reviewers rejected that in the same round: we had
named the real cause and then stepped around it rather than repairing it. They were
right. We reverted to the honest contract and filed the missing capability as its
own case (`bugs_open/173`), which is still open and unowned.

## Where we are now

The bug is **closed**. Four of four deletes guarded, live and verified in the
running binary rather than in git — three of them proven by making them fire.
Nothing is owed by this lane.

Two things are left, and neither belongs here:

1. **`bugs_open/173`** — the missing "tolerate this one substep" control in loops.
   Latent: nothing is failing because of it today. It is what pushed us into the
   workaround, and fixing it would benefit four loops, not just the one that
   exposed it.
2. **A documentation-rot problem we measured but deliberately did not fix.** Roughly
   a hundred places in the code cite a bug by a path that has since changed, because
   bug files move directories when they close — and about a hundred of those sit in
   text that a human actually reads. It must not be swept by script: some bug
   numbers name two unrelated cases, so only the author knows which one they meant.
   The right fix is a check that fires when someone edits a file.

**And one honest note about how this lane failed.** Not a single measurement in this
work turned out to be wrong. What went wrong, three separate times, was citing a
*status* instead of checking one — a bug I said was open had closed two days
earlier; the bug I was working on closed while I was writing about it; a decision I
described as an open question had been made seven hours before I wrote the sentence.
I found and fixed nine instances in the morning, wrote up why it happens, and then
did it twice more that afternoon. Being told about this class does not prevent it.
The only thing that worked was running one query at the moment of writing the claim.

## Where we're going

Nowhere, for this lane — that is the point of the summary. The work is done and the
directory is a record rather than a workspace.

What it leaves behind for others:

- **A reusable rule** any future destructive delete can adopt, with the warning
  attached: copy the rule, never the slicing, and measure your own distribution
  before choosing one. Both sites that measured found their own bug file's proposal
  was wrong on the data.
- **A test standard.** Every guard here was proven by breaking it and watching the
  tests fail — including one mutation the tests were expected to *survive*, which is
  what separates "these tests work" from "these tests are brittle".
- **A verification recipe** for proving anything reached production, together with
  the three ways a check can quietly lie about it, all three of which caught us
  (`RUNBOOK` R-V1/R-V2). They share a direction: a mis-spelled check fails toward
  "your change is not there", which during verification is exactly the answer you
  are already half-expecting, so it reads as diligence rather than as an error.
- **An open question bigger than this bug.** Retiring the link registry's carrier
  revealed that four of the five remaining site builders have also never run, while
  sites are still being created every day. Something else is building them. That is
  worth understanding properly, and it is section 4 of
  `HANDOFF_2026-08-02_continue_here.md` — not a thing to resolve by retiring four
  more agents one at a time because a query says nobody picked them.
