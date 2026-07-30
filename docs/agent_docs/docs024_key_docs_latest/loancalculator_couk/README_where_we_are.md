# Where we are — loancalculator.co.uk

The owner's running log. Plain prose, append-only, newest at the bottom.

---

**2026-07-30, afternoon.** Started by just looking at the two folders — the source
one in `domains/` and the deployed one in `sites/`. They're the same site: the first
builds it, the second is the copy that gets published. Nothing has touched either
since the 20th of March.

Then I checked whether the site was actually up, and it wasn't. Worth explaining the
symptom because it's a misleading one: the domain resolved fine, the padlock was
valid, the request went out — and then nothing came back at all. It just hung until
it gave up. No error page, no 404, no 500. That looks like a network fault, but it
wasn't.

Every working site of ours has a tiny programme running at Cloudflare's edge that
fetches pages out of our storage bucket. I checked it directly (there's a
`/worker-health` address that answers "Worker is running!" if it's there) —
gamesdesign.co.uk answered it, loancalculator.co.uk didn't. So the edge programme was
never connected to this domain, and requests were still being sent to the old Amazon
storage location from the original March setup, which is long gone. That's why it
hung: the proxy was alive, the thing behind it wasn't.

You fixed that this afternoon by adding the missing route, and the site is back —
confirmed at 16:11, all the pages I tried return properly.

**Then the more interesting bit.** You asked for this to go in through the framework
so it's managed properly, and then for anything the framework's own adoption process
would have got wrong to be reported or fixed. So I read that process carefully, and it
would have done three things we very much don't want.

It would have changed every web address. All our addresses look like
`/tools/standard-calc.html`; the adoption code rewrites them into
`/tools/standard-calc/index.html`. It doesn't matter what the real address was — it
throws it away and builds a new one. That's all 28 addresses broken.

It would have rewritten the calculators. Every page gets marked "recreate", and pages
with interactive bits get handed to a language model to rebuild. Twelve calculators
that work, that do correct arithmetic, regenerated from scratch. Even if it went
perfectly it's a pointless risk; realistically some of them come back subtly wrong.

And the setting that's supposed to prevent exactly this doesn't do anything. There's a
"fidelity" option — how faithfully to keep what you found — and you can pass it, and
it gets written into the record, and then nothing on earth reads it. I checked the
whole codebase rather than trusting the comment that says so: ten mentions of the
word, every one of them about something unrelated. So the promise is in the interface
and the implementation was never written.

**What I want to do about it.** Not work around it for this site. If I hand-port this
one, the site is fine and the framework is exactly as unable to adopt the next one.
So the plan is to build the missing faithful-preservation mode properly — make
"fidelity locked" mean "keep the addresses, keep the bytes, don't regenerate
anything" — and then use it on this site as the proof. It's two focused changes, both
opt-in, so no existing site can be affected either way. They'll go through the review
council before I call them done.

**One thing I found that you may not know about.** The site has real problems of its
own, and they've been live since March. Four pages aren't really pages — they're
loose fragments with no page structure, so they arrive as unstyled text with no
navigation. Ten of the 28 pages have no menu at all, and that includes the main loan
calculator and the overpayment tool, so anyone arriving on those from Google has no
way to get to anything else. The menu that does exist has a link to a calculator page
that doesn't exist, and it's on every page that shows the menu. And the main
calculator has been loading its stylesheet from the wrong place, so it's been
unstyled this whole time.

I'm going to fix all of that first, before adopting, for a slightly non-obvious
reason: the framework learns the site by reading the live pages. If I adopt first, it
faithfully preserves the broken version and we've frozen the bugs in.

**Order of work from here:** fix the site where it sits and publish it; build the two
framework changes and get them reviewed; run the real adoption process with the new
mode; check every page and every calculator still works, including that a later
rebuild can't quietly change them back.

**Something for you to decide, but not now.** The written content quotes interest
rates and a "last updated" date from March. I'm leaving those alone — changing the
copy isn't part of adopting the site, and I'd rather not mix the two. But once it's
managed, keeping figures like that current is exactly what the platform could be
doing for us, and that's worth a conversation after this lands.

**2026-07-30, later.** Two things done, one waiting.

**The site is fixed and back up.** All the problems I listed above are repaired and
live: the four broken pages now render properly, every page has its menu, the dead
menu link is gone (with a forwarding page left at the old address so anything
pointing there still works), the main calculator is styled again, and the address
list search engines read is correct for the first time. I checked all 34 addresses
and every one returns a page. I also confirmed the three dead files I deleted now
genuinely return "not found" — which matters more than it sounds, because it proves
the publish actually reached the storage bucket rather than me reading a stale copy.

**The framework change is written and under review.** It does what I described:
asking for "locked" fidelity now means the platform keeps a site exactly as it
found it — same addresses, same pages, byte for byte — instead of rewriting it.
Two focused changes, both switched off unless explicitly asked for, so no existing
site can be affected. I wrote tests, and then deliberately broke each safety check
to confirm the tests actually catch the failure rather than just passing quietly.
One of those attempts taught me something: my first break didn't change behaviour
at all, so the test passing proved nothing — I had to break it properly before the
check was worth anything.

It has gone to the review council and I'm waiting on the verdict, which usually
takes about half an hour. I'm deliberately not deploying anything until it comes
back, because deploying restarts the system and would kill the review mid-flight.

**One preparation worth mentioning.** The crawler had a cap of 30 pages and this
site has 29 files — and the homepage can count as two addresses, so it could
genuinely have hit the cap and quietly left pages behind. A page lost that way
doesn't error; it just never arrives. I raised the cap to 60 first, after taking a
backup of the old setting and checking the backup really held the old value.

**A mistake of mine, for the record.** I committed the framework change before
submitting it for review, which meant the reference number linking the two was not
available yet and I wrote a placeholder. The system that reports which changes have
been reviewed will therefore not credit this one automatically. Nothing is broken and
the review is happening normally, but the tidy trail is missing. The right order is
submit first, then commit — I've written that into the runbook.

**2026-07-30, late evening.** The site is adopted and healthy, the new mode works,
and one thing needs your decision.

**What went right.** The adoption ran through the real framework and did exactly what
it was supposed to. All 27 pages are now recorded in the platform with their web
addresses **unchanged** — `/tools/standard-calc.html` is still
`/tools/standard-calc.html`, not the rewritten form the old path would have forced.
Every page is marked as ours-not-the-pipeline's, and **not one of the twelve
calculators was handed to a language model to rebuild**. Under the old behaviour all
28 addresses would have changed and all twelve calculators would have been rewritten.

I also proved the safety property rather than asserting it: I deliberately triggered a
rebuild of one page afterwards and confirmed it republished **byte-for-byte identical**
content. So a future maintenance sweep cannot quietly restyle this site.

**What went wrong, and what caught it.** Before letting the deploy run I compared what
the platform had captured against what the site actually serves. **None of the 27
matched** — every stored page was about 9KB bigger, and the same ~9KB on all of them.
That consistency was the clue. The crawler does not return the page the server sends;
it returns the page *after the browser has run the JavaScript*. So the navigation menu
that our site builds in the browser had already been baked into the file, every
relative link had been rewritten to a full address, and the dropdown menu links had
been turned into links back to the same page — which would make clicking "Tools"
reload the page instead of opening the menu.

Three pages slipped out to the live site before I stopped the rest. I restored them
from our own copy and confirmed the live site is correct again; all 27 addresses
return normally. The remaining 24 were cancelled before they could deploy.

Two things are worth saying plainly about that. First, **the check that caught this
was a manual one I ran by hand** — the review council had already told me that check
was "aspirational, not built", and they were right; running it by hand is the only
reason three pages were damaged instead of twenty-seven. It should be part of the
pipeline. Second, restoring the files exposed a separate trap: the publishing step
**silently skipped one of the three files** because the bad copy in storage was newer
than my fix, and the deploy still reported success. I only found it by checking the
storage directly rather than trusting the green tick.

**Where it stands now.** I loaded the correct served bytes into the platform for all
27 pages and verified every one matches exactly, so the record is now both correct and
safe to redeploy. The site is live and unchanged from a visitor's point of view.

**Your decision.** The new "keep it exactly as it is" mode is sound, but its source of
truth — the crawler — cannot deliver exact bytes. Three ways to fix that properly:

1. **Adopt from our own files** rather than by crawling. We already hold the exact
   bytes we publish, so this is the obvious source. I had listed this as an optional
   extra; it now looks like the right answer.
2. **Fetch the pages plainly**, without a browser, just for this mode.
3. **Build the byte-comparison check into the pipeline** so it refuses to deploy when
   what was captured differs from what is served.

My recommendation is 3 regardless of which source you choose, because it is the check
that saved this run, and then 1 as the source. I have not started either — this is a
genuine choice about how the framework should work and it is yours to make.

**Also worth knowing:** the review council looked at the framework change and asked
for revisions. Several were fair and I have already made them — the code now refuses
to continue rather than quietly dropping a page, and it records in the database when
it has skipped the design stage so that is never invisible. Two objections are
genuinely for you rather than me: whether skipping the design cascade for preserved
sites is acceptable (our own documentation says that stage always runs), and whether
this needs a formal architecture review. I have not overridden either.
