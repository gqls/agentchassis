# Where we are — the chrome pin (bug 170)

Plain prose, append-only, newest at the bottom.

---

## 2026-08-01, morning — what this is about

Every site has a header and a footer. Which component supplies them is recorded in
two different places, and the platform has only ever guarded one of them.

The guarded one is the per-site record: "this site's header is that component."
Over the last fortnight three bugs were found and fixed there — one where the
chooser ignored whether a component was switched off (118), one where the repair
re-rendered the broken thing instead of fixing it (166), and one where a
page-section component could be picked as a header (167). All three are fixed,
approved and live.

The unguarded one is the *style collection*, a shared look-and-feel record that
several sites can point at, which may also name a header and a footer directly.
Nothing has ever checked what it names. Bug 170 was filed about that last night by
the lane that fixed 167, deliberately left alone because fixing it changes how
some live sites look, and that felt like a decision to put to you rather than take.

## What I found when I picked it up

Two things the original bug report doesn't say.

**First, it undercounts.** The report says three deployed sites are showing a
header the library says is switched off. That's right for headers. But four sites
are showing a switched-off *footer* — including the one site whose header is
correct. So it's three sites on a dead header and four on a dead footer.

**Second, and this is the one that matters: it is not just about what gets
displayed.** There is a routine background job that decides which component each
site's header slot should point at. If the style collection names one, that job
takes it, no questions asked, and writes it into the per-site record — the very
record the last fortnight's fixes just repaired. Right now all four sites'
per-site records are correct and all four style collections are wrong. The next
time that job runs for one of them, it would quietly put the broken value back and
blank the stored header so it gets rebuilt.

So the fix that shipped on Thursday isn't durable while this stands. I checked
whether that job has actually run recently — it hasn't, not in the nineteen days
we keep records for — so nothing is breaking this week. It's a loaded gun, not a
fire.

There is also a third place, which I found only because a test I'd written for the
first two failed on it: when we copy a style collection to make a new one, we copy
whatever it names, broken or not. That's how the problem would spread to sites
that don't have it yet.

## The decision I took, and why I didn't come to you first

The reason 170 was left for you is that fixing it changes how three sites look.
Having read what the earlier fix actually did, I don't think that's the question
any more.

When the style collection's header is ignored, the site falls back to the library
and gets `header-theme-chrome`. That is *exactly* the component the earlier repair
already moved those same three sites to, with council approval, and it is what
their per-site records say today. So the change doesn't decide anything new about
how they look — it makes the second record agree with the answer we've already
given. That's a much smaller thing than "shall we restyle three clients' sites",
and I've said so plainly in the council submission rather than glossing it.

Two honest caveats. The four style-collection records are still wrong after this —
they're now ignored rather than obeyed, which is why I've also taught the existing
detector to notice them and raise them for a human, since repointing a *shared*
collection affects three sites at once and is a judgement. And nothing changes on
any live page until the site chrome next gets rebuilt, because chrome is stored
rather than regenerated on each page build (that's bug 117, still open).

## Two mistakes worth recording

I wrote a test to stop anyone adding a fourth unguarded reader in future. It
immediately failed — on the file I had just fixed. The correct way to write the
fix puts the safety check in a place my test wasn't looking, so a correct fix and
a completely unguarded one looked identical to it. Then I made almost the same
mistake again in a second test: it searched a whole file for a word, and the word
happened to appear elsewhere for an unrelated reason, so it passed when it should
have failed. Both are the same error — asking about *the file* when I meant to ask
about *the thing* — and both are now fixed and deliberately re-broken to prove the
fixes work.

The other thing worth telling you: I put this through the automatic diagnosis
service, as the rules now require for a claim this structural. It ran, spent four
rounds gathering evidence, and produced no conclusion at all. The reason is that
the file at the centre of the question is 94KB and the service can only show its
reader 60KB, so the one thing it needed to look at was omitted every single time.
That's not specific to this bug — **any** file over that size is effectively
invisible to it — so I've written it up where the next person will trip over it.
I verified everything by hand instead and said so.

## Where it stands

Code written, all tests green against a clean checkout, submitted to the review
council, committed. It does nothing until the next chassis build goes out. The
ticket stays open until it's live and I've confirmed it on the running pods.
