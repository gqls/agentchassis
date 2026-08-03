# SUMMARY — 2026-08-03b — fixed and live; one fix reviewed, approved and still inert

> Second summary of the day. The first (`SUMMARY_2026-08-03_the_site_was_the_symptom.md`)
> was written mid-flight: it said the homepage was "queued and will land", the
> council verdict was "in flight", and the second site had not been touched. All
> three have since resolved, and the owner has taken the one decision that was
> blocking. That is the inflection this entry records — not a restatement.

## What we're trying to do

The owner reported finetuning.uk as "looking terrible" and asked for it to be
fixed **by the framework rather than by hand**, with three harder asks attached:
run the audit checks and see what is wrong including visually; make sure the
handlers are picking work up and fixing it; and check that the framework catches
everything.

The site was the symptom. The framework was the subject.

## Where we've come from

Nineteen broken images — eight on the homepage, eleven on `/about.html` — each a
120px broken-image icon in a grey circle. `departments-grid` had been forked from
a staff-photo component and repurposed for departments; its schema correctly
declared `icon` as a name, but its template still rendered that name into
`<img src="…">`. The page was asking the server for a photograph called "cpu".

Behind that sat two larger facts. The check that owns broken images could not see
the shape at all — in one run, on those two pages, it filed five findings for
missing case-study photos and stood beside nineteen worse ones. And nothing on the
site had ever been *attempted*: every item read `attempt_count = 0`, because the
dispatcher claims only `triaged` work and the sole promoter of `detected →
triaged` sits inside a loop whose schedule has been off since 2 May.

## What we've done

**Fixed the component at source** (`293`), targeting `<i data-lucide>` — not a
new approach, but what the `features` component on the same page already does.
Verified beforehand that every affected page loads the icon library.

**Taught the check to see the shape** — a third predicate, its own finding kind,
six negative controls guarding against the classic failure of widening a regex
until it reports nothing. Committed and tested.

**Drove the framework end to end.** Wrote a manual improvement-loop trigger
(`294_TRIGGER_…`, registered as IMP-050) because the one the documentation names
does not exist in the tree. Fired it at the site: **128 items promoted to
triaged**, the first dispatchable work here since April, and the handlers began
claiming.

**Caught the trap that would have made all of it cosmetic.** 68 rerenders were
already queued and 42 carried no `spec.reason` — which routes them to the branch
that re-staples a page from the stored HTML being replaced. They would have
completed, reported success, and changed nothing. Queued explicit
section-regenerating items instead (`294`, `295`, `296`).

**Finished the class fix on the second affected site** without taking on its
backlog — two rerenders only, because firing the full loop there would have set
off 37 items on a site nobody asked about.

## Where we are now

**finetuning.uk is clean, verified at the served page rather than at a status:**

```
https://finetuning.uk/            broken images 0  (was 8),  18 icons
https://finetuning.uk/about.html  broken images 0  (was 11), 11 icons
```

**The fleet census has gone 31 → 8**, and the 8 are one remaining page:

```
ai-agent-orchestration.com/            broken images 0  (was 8),  17 icons  ✔
ai-agent-orchestration.com/about.html  broken images 8            claimed, in flight
```

**The council approved the check fix on round 2.** Round 1 was REVISE and was
right: a standing landmine pairs that file with a sibling check and I had not
read it. Checked properly, they cannot collide — the sibling matches two literal
paths, both containing characters the new pattern structurally excludes. That
analysis now lives in the landmine record and syncs to `doc_notes`, which was a
reviewer's suggestion, rather than sitting in a Go comment where no agent finds it.

**The check fix is reviewed, approved, and NOT live.** Pod-grepped both replicas
with a positive control: the new symbol is absent, the control present. It needs a
chassis build. **The owner has taken this decision — they will roll the chassis
later, when things are quiet.** That is the right call and it is now the only thing
standing between the framework and this defect class; nothing further is needed
from this lane to make it happen.

**The queue is still draining** — 62 complete, 115 triaged on finetuning.uk. The
substantial remainder is content, not layout: 17 phantom internal links, 28
high-severity misdirected CTAs (copy naming a different page than the link goes
to), 25 components missing schema-required fields.

## Where we're going

**One decision remains the owner's: the 204 findings parked across 10 sites.**
Unchanged and still the largest finding here. Re-enabling `improvement-sweep`
fixes the fleet at once but was switched off deliberately and would start changing
ten sites immediately; per-site firing is safe and does not scale; bulk-promoting
skips the step that decides what is worth doing, on a queue where 235 items have
already failed at least once. The recommendation stands: keep firing deliberately,
and give the three-month disconnection its own answer rather than settling it as a
side effect of whoever is fixing a site that week.

**After the roll, one thing is worth re-running** to close the loop honestly: the
same fleet census plus a discovery run, to confirm the new predicate fires on real
data rather than only in tests. Until then the fix is proven in unit tests and
approved by review, which is not the same as proven in production.

**Two things are recorded and deliberately not fixed.** Eight of
ai-agent-orchestration.com's icon values are not icon names at all — `strategy`,
`research`, `operations`, `quality`, `development`, `design`, `data`, `content` —
now verified live sitting in `data-lucide` attributes that lucide will not
resolve. Those render as an empty badge rather than a broken image: an improvement,
not a repair, and a `content_data` defect on that site rather than a template one.
And a council seat made the broader point that this closes one shape of a general
problem — schema-declared strings rendered into structural HTML sinks with no
bind-time validation that the string resolves. The architecture seat approved the
point fix on the same reasoning; the generic guard remains unbuilt and is worth
its own decision.
