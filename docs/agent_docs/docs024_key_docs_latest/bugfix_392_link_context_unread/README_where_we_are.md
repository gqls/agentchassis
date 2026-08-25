# Where we are — bug 392 (plain prose, append-only)

## 2026-08-25 — the lane opens

Yesterday another thread noticed that when our writing system can't load the list of pages it's
allowed to link to, it carries on and writes the page anyway — with an instruction saying "do not
put any internal links in this". That behaviour is deliberate and we're not changing it: failing
the whole build because one query was slow would be worse. What nobody had built is the bit
afterwards. The system writes a note to itself saying "this page went out without links", and
then nothing ever reads that note. The page stays link-less indefinitely.

I picked this up today because it was unowned — nobody had a working directory for it and nobody
had touched it.

**The first thing I did was check whether it is still a real problem, and the honest answer is
"yes, but not in the way the bug file says".** The note has only ever been written twice, both on
the same minute yesterday, and both of those pages are already fine — one got rewritten an hour
later by coincidence, the other's build failed so it never published at all. So there is no fire
burning right now. What there is, is a hole: next time a database query is slow, the same thing
happens and again nobody notices.

**The more interesting finding is how big the surrounding problem is.** I counted every published
page on every site and asked how many carry no internal links written into their body text at
all. It's 411 out of 736. Not all of those are broken — a calculator page reasonably links
nowhere — but on the page types where linking is normal (articles, guides, ordinary content
pages) it's 187. So the specific bug we were asked to fix accounts for about one page in two
hundred of the visible symptom. The owner agreed we should detect the damage itself rather than
just this one cause of it.

**I got one measurement wrong along the way and want it on the record.** My first count said 140
pages, not 411. I was counting links in the finished HTML, which includes the navigation menu and
the buttons the template adds — so a page whose writer wrote no links at all still looked like it
had two or three. Counting what the writer actually produced is a different number and a much
worse one. This is logged in the shared wrong-calls file, because the same mistake would flatter
any future check built on it.

**The good news is that almost nothing needs building.** The machinery to fix a page's prose
already exists and runs about thirty times a day. There is already a check that finds the exact
opposite problem (a page that nothing links *to*), sitting in a folder with about a hundred
sibling checks, all of which get scheduling, de-duplication and safety brakes for free. So this
becomes one new check in an existing family, not a new service.

**One thing I found by measuring that I would not have guessed:** of those 187 pages, 48 are on
sites where a customer owns the page and we're not allowed to rewrite it — and it's 48 out of 48,
meaning *every* owned page is link-less. Filing repair jobs against those would have created 48
stuck records of exactly the kind another open bug was filed about last week. So the check will
report those and not act on them, and I'll tell that thread what I found, because "every owned
page has no links" is its own separate problem and not mine to fix.
