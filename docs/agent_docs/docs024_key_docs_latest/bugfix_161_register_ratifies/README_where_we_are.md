# Where we are — bug 161's residual

Plain prose, append-only, newest at the bottom.

## 2026-09-03

You asked me to look at bug 161, resume it if nobody was on it, and prefer a fix that helps the
whole framework over one that helps a single case.

**Bug 161 was already closed, and I checked that it really was fixed rather than taking the
file's word for it.** It is fine. The false claim it was about is gone from the database and
gone from the six live pages, the daily check that guards it ran again yesterday, and the tool
it points at is unchanged. Nobody was working on it.

**So I asked what the closing session had left behind, and the answer was worse than what the
bug had been about.**

Every site has a small file of "things we are allowed to say" and "things we must never say".
The never-say list is the important half. For **two sites, that file had stopped being readable
altogether**, and everything that depends on it had quietly switched itself off — not
complaining, just behaving as though the site had never had one.

The two sites are **finetuning.uk**, off since 24 August, and **noted.co.uk**, off since 25
August. Between them that is ten never-say rules that were not being enforced. On noted.co.uk
they are the ones that matter most on a note-taking product: never claim the notes are
end-to-end encrypted, never claim we are GDPR compliant, never say military-grade or unhackable
or "you can't lose a note". On finetuning.uk one of the disabled rules exists because **you
personally ruled on 27 July** that a "~80% reduction in quote preparation time" claim had
nothing behind it and had to go. That rule was unenforceable for ten days.

**Nothing bad was actually published in the window** — I checked every deployed page on both
sites against their own rules and found no violations. This was a loaded gun, not a wound.

**Why it happened is the part I think is worth your attention, because nobody was careless.**
The file lets you record facts with a number attached — "11 tools live", "4 inputs". Somebody
writing finetuning.uk's file needed to record facts whose value is *words*, not a number: that a
model is licensed under "MIT", or "Apache 2.0", or that booking hours are "9am–5pm UK time".
That is an entirely reasonable thing to want to record. There was no way to write it, and the
consequence of trying was that **the whole file became unreadable, including the never-say
list**, with no error anywhere a person would see. Only a line in a log inside a server that
gets replaced every day. And it was spreading — that site went from three such entries to eight
in three days, as the author kept doing the sensible thing.

**What I changed:** one bad entry now costs you that entry and nothing else. The never-say list
survives. And the daily sweep now raises a proper job saying which entry is broken and what it
cost, so this can never be silent again. I also fixed a related blind spot: there are 27 facts
across five sites that nothing ever re-checks, and the daily sweep could not even count them —
it can now, so we can watch that number instead of guessing.

I did **not** add a way to record word-valued facts, though that is the real answer. It is a
bigger change to a shared format and deserves its own review, and it is no longer urgent now
that a bad entry is harmless. I have written down exactly when we should promote it: if these
broken-entry jobs keep appearing, that is the signal.

**Two things I got wrong, which I have written up properly.** I checked the six pages from bug
161 by guessing their web addresses instead of reading them from the database, got "page not
found" on all of them, and briefly thought something was badly broken — it wasn't; I had the
addresses wrong. And a script I wrote to check all 27 sites silently only checked **one** of
them and reported a clean result, which I nearly believed. Both are now logged with the check
that would have caught them.

**The independent review approved the change first time**, and was still worth reading closely —
it caught me writing "live" about a test I had run on my own machine against live data, which
is not the same thing and is exactly the sort of overstatement that misleads the next person.

**Where this stands right now:** everything is written, tested, reviewed and committed, but it
is **not yet running** — it goes live with the next server build, which is due shortly. I have
written down precisely how to confirm it is genuinely working afterwards, because "we deployed
it" is not the same as "it works", and I would rather check than assume. The two sites' broken
entries are small repairs for the teams that own them, and I have written to both.
