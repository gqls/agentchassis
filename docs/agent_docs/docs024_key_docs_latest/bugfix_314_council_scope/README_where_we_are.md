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
