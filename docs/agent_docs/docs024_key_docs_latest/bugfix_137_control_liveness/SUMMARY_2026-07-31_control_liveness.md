# SUMMARY — 2026-07-31 — one definition of "is this control alive"

## What we're trying to do

Bug 137 reported something unusual: two pieces of code that both answer "is this
button real, or is it a dead thing dressed up as one" sit in the same function,
run in the same pass, and give **opposite answers** about one element on a live
site. Neither was misbehaving — each did exactly what it was written to do. The
person who filed it said honestly that they did not know which was right, and
flagged that the reading which made their own red result disappear was
suspiciously convenient. It was filed because a reviewer asked for it to be,
rather than quietly deferred.

The job was to reconcile them, and to do it in a way that holds for the framework
rather than for that one element.

## Where we've come from

The two mechanisms share an escape hatch: a section can be marked "the browser
fills this in later", and anything inside it is excused, because its links are
meant to be placeholders. Sensible.

The problem was **how the excuse was tested** — a plain "does this text appear
anywhere in what I was handed". So its reach depended entirely on how much HTML
the caller passed in. Give it one section and it means "is this section a
placeholder", which is right. Give it a whole page and it means "does this page
contain a placeholder anywhere", which excuses **everything else on the page**.

That line existed in nine places. Nobody had written the second behaviour down
and no test caught it, because read at any single site it looks obviously
correct. One earlier thread had noticed and worked around it in its own corner,
with a comment explaining the reasoning — the right answer, applied where it
could not be enforced.

So the reported disagreement was a **symptom**. The cause was that the excuse had
no fixed size.

## What we've done

Replaced the nine hand-written copies with one shared judgement that is scoped to
the *element*, not to whatever the caller passed. Two forms — one for code working
on raw text, one for code working on a parsed page — both derived from a single
definition, so they cannot drift apart.

Measured on the real pages rather than argued: on vonc.com's home page the old
test excused **100%** of the page and the new one excuses **12.6%**, which brings
two genuinely dead buttons into view. On the page bug 137 is actually about, both
mechanisms now agree the element is excused, and the page comes out byte-for-byte
unchanged. That is the reconciliation the bug asked for.

**The council reviewed it three times and changed it twice**, and both changes
made it better:

- **Not everything should be narrowed.** Some consumers *check* and file findings
  for a human; one *edits the page*, and the way it "fixes" a dead link is to
  delete the link and leave the words. There is already a written warning about
  that. Narrowing a checker only surfaces more findings; narrowing an editor makes
  it rewrite more pages. So the editors keep the wide excuse — for them, doing
  less is the safe direction. The platform had already ruled on this next door,
  where dead controls are deliberately routed to a human "because picking a fixer
  automatically would guess".
- **A gate, not a note.** Adding a tenth copy of this predicate used to cost
  nothing and tell nobody. It now fails the build unless the author names the
  scope they intend, or records the site as a deliberate exception with a reason.

## Where we are now

The code is committed, builds clean, and its tests are proven load-bearing by
deliberately breaking the fix in both directions and checking that the right tests
fail. **The bug is still open, correctly**: the fix is not in the running image —
verified at the pod, not at git, with a positive control in the same command.

Two things are outstanding and neither is a defect in the work:

1. **The third council round could not finish.** It died in a way that looks
   exactly like a rejected submission but is not — a reviewer's own model call hit
   a fleet-wide API usage limit that lifts at **00:00 UTC on 1 August**. Three
   unrelated runs died the same way in the same minutes. The submission is valid
   and must be **resubmitted unchanged**; editing it would be chasing a fault that
   is not in it.
2. **It ships when the chassis next rolls** with these commits in it.

The most useful thing this session produced is probably not the fix. It is the
rule underneath it — **a shared guard tested against "whatever the caller passed"
has no fixed blast radius** — and its counterweight, that **narrowing such a guard
is safe for code that reports and unsafe for code that acts**. Both are now written
where the next thread will meet them.

## Where we're going

- After 00:00 UTC on 1 August: resubmit the unchanged round-3 plan under the same
  correlation, read the verdict, and act on it.
- After the next chassis roll: confirm at the pod, re-run the measurement on the
  affected page, and move the case to `bugs_closed/`.
- Left deliberately undecided, and recorded rather than buried: whether deleting a
  dead link and keeping its text is the right repair at all. It is a real question
  about a live mechanism, it is not this bug's to settle, and the code now says so
  in the place someone would look.
