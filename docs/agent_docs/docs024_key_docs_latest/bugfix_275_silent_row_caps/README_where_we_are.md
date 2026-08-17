# Where we are — the tool suggester was only ever shown half the library

Plain prose, append-only, newest at the bottom.

## 2026-08-16 — what was wrong

When we decide which interactive tools a website should have — a mortgage calculator, a comparison
table, that sort of thing — we ask a model to pick from our library. The query that fetches the
library ended with an instruction to return only the first 30 entries, sorted alphabetically.

The library now holds 74. So the model choosing tools for every site has been picking from **30 of
74**, and which 44 it never saw was decided by the alphabet.

The reason nobody noticed is the interesting part. The model returns sensible-looking suggestions
whether it has seen 30 tools or 74. There is no error, no warning, nothing that looks wrong. You can
only find this by reading the query. And it has been getting worse on its own: when it was reported
two days ago the library held 68 and 38 were hidden; today it is 74 and 44.

## 2026-08-16 — what we did

The obvious fix is to remove the limit, and that would have been the wrong fix. The limit was not
purely arbitrary — 74 entries is a large chunk of prompt, and someone capping it was responding to a
real cost.

So I measured where the cost actually is, rather than arguing about it. Each entry has five fields,
and **one of them, the description, is 80% of the total** — some run to 2,500 characters, where the
median is under 400. The model does not need 2,500 characters to judge whether a mortgage calculator
suits a mortgage broker; it needs the first sentence or two.

So we bounded the descriptions at 200 characters and removed the row limit entirely. The result,
measured on live data: **the whole library instead of 41% of it, for 22% more prompt**. I checked that
200 characters still carries the meaning by reading the longest descriptions — they do.

**And then the more useful half.** This is not really one bad query; it is a shape. A row limit feeding
a model is invisible by construction, because the output looks the same either way. I counted every
such limit in our live configuration: 26, of which 19 are the harmless "fetch exactly one record" kind
and **seven are caps that could be silently hiding things**, this one being the worst.

All 26 run through a single piece of code. So that code now notices when a query comes back exactly as
full as its own limit — which is the signature of a truncated view — and says so in the logs, naming
the step. It changes nothing about what any query does or returns; it just stops the class being
invisible, including for limits nobody has written yet.

## 2026-08-16 — where this leaves it

The library fix is **live now** — it was a configuration change, and those take effect immediately
rather than waiting for a build. Verified on the running system: the limit is gone, the descriptions
are bounded, an earlier safety rule about backend-requiring tools is still intact, and **44 tools that
could never previously be suggested are now reachable**, the first alphabetically being an "Early
Settlement Estimator".

The detection half is Go, so it waits for the next build like any code change.

**Two things I have not claimed.** The bug report asks for proof from a real suggester run — that a
late-alphabet tool is genuinely absent from the model's prompt before and present after. That needs the
suggester to actually run, which has not happened yet, so the check I have done is necessary but not
sufficient, and the bug record says so. And **nobody has checked whether the other six caps are
actually hiding anything** — each is one query, and the new warning will answer it in production once
the code ships.

One small thing found along the way and deliberately left alone: one library entry has a blank name,
which means it sorts first and has therefore always occupied one of the 30 visible slots while telling
the model nothing useful — its description is internal developer notes. That is a content problem
rather than this one, and it belongs to whoever owns that tool.

## 2026-08-17 — the review sent it back once, and it was right to

The change went to the review council and came back **revise**. Four of its objections found real
things, and one of them changed what is now running.

**The one that mattered most.** To fit the whole library into the prompt I shortened each tool's
description to 200 characters — and I did it *silently*. A reviewer pointed out that this is the same
mistake as the bug I was fixing, just moved one level down: a tool whose distinguishing feature is
described in its third sentence now reads as generic, and neither the model nor anyone reading the
record can tell anything was cut. That is exactly right, and slightly embarrassing given what the fix
was for. Shortened descriptions now end with a visible "[…truncated]" marker, which costs about 3% and
means the model can at least see that there is more. I chose the marker over simply allowing longer
descriptions because doing that costs more than twice as much prompt and *still* cuts a third of them
silently.

**One objection was wrong about the file and right about me.** It said my migration never took a backup
before changing live configuration. It always did — the backup is there, timestamped. But the summary I
submitted for review showed the clever parts and left the safety line out, and a reviewer can only judge
what you show them. The lesson is worth keeping: when you summarise a change for review, show the
safety-critical lines, not the interesting ones.

**One was right in a way I would have missed.** Two reviewers flagged that some agent types have two
active configuration rows, where only one is actually used, so a targeted update can silently do
nothing. I checked: this agent has exactly one row, so the change was fine. But it was fine *by luck* —
the guard I had written could not have detected the problem it was supposed to guard against. That is
now fixed properly.

**And a minor one found a genuine process hole.** A reviewer asked me to confirm my migration number
was not already taken. It wasn't — but because I applied the change by hand rather than through the
usual tool, it had been recorded nowhere, which is precisely how two people end up claiming the same
number. Both are now on the register.

## 2026-08-17 — something I nearly told you that was not true

While checking whether widening the menu genuinely changes anything, I found that suggestions the model
cannot match to an existing tool get sent off to be **built from scratch**. So I reasoned the old cap
must have been making us rebuild tools we already owned — and the numbers looked damning: 18 of 19
"build this from scratch" requests named a tool already in the library.

It is a good story and it is false. I checked the dates before writing it down, and every one of those
18 library entries was created *at or after* the request that named it — meaning they exist *because*
of those builds, not in spite of them. That is the pipeline working normally, not waste.

I mention it because it is the third time today my reasoning has run ahead of my evidence, and the only
difference on this occasion is that I checked before telling you rather than after. The harm from the
original bug is what the report said it was — the suggester was choosing from half the library — and
there is no measured waste on top of it.

## Where it stands

The library fix and the truncation marker are both **live**. The detection code waits for the next
build. The review is on its second round.

Two things remain open and are written down where the next person will find them: the bug's own
end-to-end proof needs a real suggester run, which has not happened yet; and **two other agents have
the same defect** — one of them choosing from 10 of up to 107 items, which is worse than the bug I set
out to fix. Whether those become their own tickets is your call.
