# Where we are — the council gate and config changes

## 2026-08-20 — the reviewers can finally see the changes that matter most

There is a panel of AI reviewers that checks changes before they go out. It had a rule about
what it would look at, and the rule was written as a list of folders. That sounds harmless, but
on this system the folder list quietly excluded the single riskiest kind of change we make.

Most of what alters how the site-building agents behave is not program code at all — it is
configuration, written as small database files. Those live in a documentation folder for
historical reasons, and the reviewers' rule treated anything in a documentation folder as
writing, not as machinery. So two thirds of the changes that alter live behaviour could not be
reviewed at all. You could force it through with an override, but nothing then distinguished
someone who knew what they were doing from someone dodging the check.

Worse, it was excluding exactly the changes where review pays. When someone forced a
configuration change through recently, the reviewers found a genuinely serious problem with it.
The program-code change they were happy to look at, they had nothing to say about.

**That is fixed.** Configuration changes are now reviewed as a matter of course. Writing, and the
scratch files that accompany a change without being the change, are still ignored — those really
would just cost money. And rather than patch the one place the bug was reported, I found the rule
had been copied by hand into three separate tools that had drifted apart, and replaced all three
with a single shared definition.

**The part worth telling you about.** I put the fix through the reviewers themselves, and they
approved it — but one of them found a real mistake in it. I had excluded a category of file on
the grounds that the automatic installer skips them. It turns out those files are the ones
applied *by hand*, deliberately, when the order matters. They are live configuration, and I had
just exempted them from review — which is the very mistake this whole fix exists to correct,
made inside the fix itself. I checked it myself before believing the reviewer, and they were
right. It is corrected, and the reasoning is written down where the next person will hit it.

I think that is the best possible evidence for the change: the first thing the widened reviewers
caught was a flaw in the change that widened them.

There is one loose end I have left deliberately: the reviewers still cannot look at changes to
their *own* tooling — only at configuration and program code. Fixing the gate still requires the
override. Nobody has asked for that yet, so I have not decided it unilaterally; it is written
down as an open question rather than quietly done.

Nothing needs a decision from you. The cost is worth mentioning: this will mean somewhat more
review activity than before — I measured it at roughly a fifth more, against a system whose
review costs were cut by three quarters earlier this month. I have written down a specific
threshold at which we should switch to a cheaper reviewer panel for configuration rather than
turn the reviewing back off.

---

## 2026-09-02 — picking this lane back up, for the loose end it left

This lane closed on 20 August. I have reopened it to finish the one loose end named in its own
close-out note, and to check the other one is still a loose end at all.

**First, what the two loose ends were.** When we fixed the review gate so it could look at
configuration changes, we left two things undone and wrote them down. One: the reviewers still
could not look at changes to *their own tooling*. Two: there was a fourth copy of the rule that
says "which files are migrations", living in a different script, and it had already fallen out of
step with the others.

**The first one has largely fixed itself, by other people.** Two more widenings happened after we
went quiet — on 23 and 24 August the owner ruled that the detector code for the nightly check
fleet, and then the commit-time checker script, should both be reviewable. So the gap we named has
been closed from two directions by other threads. What remains of it is narrow: the review gate's
own two scripts still cannot be reviewed. I have not touched that, because nobody has asked for it
and it is a judgement call, not a defect.

I want to flag one good sign in that. When those two widenings went in, each of them correctly
edited *both* halves of the scope — the rule itself and a second, hand-kept list in the coverage
report. That second list is a trap this lane wrote up in August precisely because it is easy to
miss. It did not catch anyone out. The warning worked.

**The second one is real, and I have now measured it rather than asserted it.** Here is the thing
in plain terms.

There is a script that runs on every single commit, in every session, and quietly reads what you
are about to commit looking for a handful of known mistakes. One of those mistakes is a database
change that cannot safely be run twice — if it ever gets run a second time it stops the whole
migration system dead, and that has happened for real: one such file blocked everything for three
days.

For that check to work, the script has to decide which of your files *are* database changes. It
decides that by the filename. And its rule for filenames was written to only recognise names in
lower case. The actual migration system accepts capital letters too. So any database change with a
capital letter anywhere in its name has been sailing straight past the check, silently, since the
check was written.

**How much?** 743 files are real, live database changes. The check could see 738 of them. So five
were invisible. That is the honest number, and I nearly reported it as 660 — I used the wrong tool
to compare the two lists and it gave me a nonsense answer that I did not immediately disbelieve. It
is written up in the shared mistakes log.

**And now the good news, which I want to state as plainly as the bad.** All five of those invisible
files are, as it happens, already written safely. So nothing is broken today. What we have is a
hole in a smoke alarm, not a fire. The next file of that shape is the one that would have got hurt.

**One thing turned out to be the opposite of what everyone assumed, including me.** There is a
convention here for a database change that has to wait its turn: you park it with a special name
until the ordering is right. The obvious assumption — mine, and the assumption of the lane that
writes the most of these — is that a parked file is not yet live, so the check need not look at it.

That is backwards. I checked what actually happens to parked files over time: they get renamed to
drop the parking suffix, 26 of them in the last month. And then I found the reason, in the code:
the system **refuses** to record a parked file as "already done" — it will only record it under its
grown-up name. So the sequence is forced. You apply the parked file by hand, you rename it, and
only then can you record it. In the gap between the rename and someone remembering to record it,
the system sees a file it thinks has never been run, and runs it again.

So parked files are not the safest category for this check. They are the **most** dangerous one,
because they are the only kind that is *guaranteed* to be applied by hand before the system knows
anything about them. The check should look at them hardest, and it currently does not look at them
at all.

**Working with another thread on this.** The lane that writes most of these files is active right
now, so I told them before changing anything under them. That went well and in both directions.
They first told me parked files are never renamed — which was true of their own practice and not
of anybody else's, and they corrected it themselves within the hour once they checked the house
rule. In return, my heads-up made them go and look at their own files, and they found **seven** live
database changes that had been applied by hand and never recorded — which is exactly the dangerous
state described above, sitting there unnoticed since August. They have since recorded all seven.

That is worth saying out loud: the message that fixed a real problem was not the fix. It was
telling a neighbouring thread what I was about to touch.

**One new thing to raise, which is not mine to fix in this change.** They suggested I also add a
check for "files that have been applied by hand but never recorded". I looked, and that check
already exists — it is what the migration tool prints by default, for free, every time anyone runs
it without arguments. Their seven files would have appeared on that list on any day for a month.
So this is not a missing safeguard; it is a safeguard nobody runs. That is a different kind of
problem and deserves its own item rather than being bolted onto this one. I am filing it, at their
request, and crediting them.

**Nothing needs a decision from you.** The change itself is small and cannot break anything — the
script it touches is advisory and is forbidden from blocking a commit. I am putting it through the
reviewer council, which is possible only because of the widening this very lane shipped in August;
before that, fixing this would have required the override.
