# SUMMARY — 2026-08-14: the last two tool tests shipped, and a four-month-old broken logo is proven fixed

*The read-out on finishing the batch-8 tool-testing backlog and on `bugs_open/248` (the
"images get published under the wrong filename" bug), written to be said aloud.*

## What we're trying to do

Two separate but related things. First: make sure every interactive tool on our sites
actually gets a real, automated "does this still work" test, not just a page that loads.
Second, picked up along the way: fix a bug where freshly generated images — logos, hero
photos — sometimes get saved to the site under the wrong filename, so the page that's
supposed to show them gets a broken image instead, silently, with nothing reporting an
error.

## Where we've come from

The tool-testing work had two calculators left on its list: a ranking calculator on the
games-design site, and a cost calculator that appears on five different sites. Separately,
someone had noticed that one client's site (Gas Wholesalers) had been showing a broken logo
for four months, traced it to two small mistakes in how the system names files it deploys,
and written a fix — but the fix hadn't been reviewed, hadn't shipped, and nobody had checked
whether it actually worked.

## What we've done

Finished both calculators — wrote proper tests for each, including deliberately breaking
them in several different ways first to prove the tests would actually catch a real
problem, not just wave a page through.

Then took on the broken-logo fix. It went to our automatic review process, which is meant to
catch problems before they ship, and it earned its keep: four rounds of back-and-forth, and
every round found something real. Round one caught that the fix's own safety net (falling
back to a saved copy of an image when the usual lookup fails) could, in one specific case,
grab the wrong caller's data — traced and fixed with one small, targeted change. Round two
found that the exact same small mistake existed in a second place nobody had checked yet —
fixed the same way, and round two also flagged two things that turned out NOT to be
problems: one old bug the fix seemed to bump into had actually already been fixed months
ago, and another flagged issue was real but belonged to a completely different, already-known
bug on a different part of the system. Round three asked two very specific, checkable
questions — "are you sure the part of the config you're editing is actually the part that
runs?" and "could there be a second, hidden copy of this configuration that your fix
missed?" — and both came back clean once checked directly, rather than just reasoned about.

Round four raised something different: not "this specific fix is wrong," but "should this
whole shortcut ever be allowed, anywhere in the system, even where it isn't causing a
problem today?" That's a real question, but it's not one four rounds of automatic review
can settle by itself — it needs a person to decide. So rather than go for a fifth round, we
stopped, wrote the question down properly, and moved on.

While all of that was happening, the software rolled out a new version on its own — nothing
to do with us — and it happened to include the fix. Rather than just assume it worked
because the version number said so, we found the exact broken logo, told the system to try
deploying it again, and watched. It worked: the logo now saves under the right name, and we
checked the actual live web page ourselves rather than trusting a success message. The image
loads. Four months of a broken logo, fixed and confirmed.

## Where we are now

Both calculators are live, tested, and proven to fail their tests if deliberately broken —
so the tests are worth something, not just decoration. The tool-testing backlog we'd been
working through is now clear.

The broken-logo fix is live in the running software and proven correct on the one image we
tested it against. It has not yet been formally approved by the automatic review — that's
sitting on the bigger, "should this shortcut exist at all" question from round four, which is
waiting on a person, not on us doing more checking. A second broken image on a different
site (a mortgage calculator's homepage photo) almost certainly has the same fix available to
it, but we haven't gone and checked that one yet — worth saying plainly rather than assuming.
And there are roughly 150 other images across 15 sites that got saved under the wrong name
before this fix existed; fixing the code stops new ones from happening, but it doesn't
un-break the old ones. Someone still needs to plan out how to safely re-run the fix across
all of them.

## Where we're going

Three concrete things are left, all written down in detail for whoever picks this up next:
send the "should this shortcut exist at all" question to a person for a real decision rather
than more automatic review; check the second broken image the same careful way we checked
the first; and design a proper, careful plan for cleaning up the ~150 already-broken images,
rather than improvising something quick. Given how much ground this session covered, we've
written all of this up as a proper handoff and stopped here rather than pushing further.
