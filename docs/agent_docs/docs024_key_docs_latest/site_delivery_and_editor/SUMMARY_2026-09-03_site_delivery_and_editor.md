# SUMMARY — 2026-09-03, site delivery and editor

*Written to be read aloud. Supersedes nothing: `SUMMARY_2026-08-26` still stands as what we
believed then, and the gap between the two is part of the record. Previous in the series:
`SUMMARY_2026-08-26_site_delivery_and_editor.md`.*

## What we are trying to do

Take a paid order for a website, build the whole thing without a person touching the HTML, and
hand it to the customer with a link and an email — and be able to say, item by item, why we
believe it is good enough to send. The delivery half of that chain is this lane: the review the
owner does, the approval he gives, the zip we cut, the email that goes out.

## Where we have come from

A week ago the chain existed but had never carried a real order. Then boxingonline.com came in as
the first paid build, and the owner read the finished site the way a customer would. He came back
with fourteen things wrong with it. That review changed what this lane is for: not "does the
delivery machinery work" but "can we tell the difference between a site that looks finished and a
site that is finished". He also set the rule we have been working under since: everything gets
fixed before the delivery email goes out.

Most of the fourteen were not delivery faults at all. They were places where the machinery had
done what it was told and produced something nobody would want — a logo with its background baked
in, cards with empty summary lines, a call-to-action that explained the site to the reader instead
of talking to them, a fight calendar with no fights in it. Each one belonged to a different thread,
and the last ten days have been mostly about routing them and then verifying the results rather
than fixing them here.

## What we have done

Today three of the owner's points closed properly, meaning we checked the thing he would see and
not the system's opinion of it.

The logo is now genuinely transparent. That took two attempts by another thread, a bug in the
safety check that was passing bad results, a fix, a release, and the owner topping up the image
account. We confirmed it by fetching the file the site actually serves and measuring the
transparency in it, four fifths of the canvas, clear to every edge.

The home page's article cards now carry their one-line summaries and have lost the site name that
was being appended to every headline. That one needed a rebuild of the page, and the rebuild kept
being refused by a guard that stops a page losing half its text in one save — because the
correct repair was to cut eleven hundred characters of padding down to under two hundred. The
guard was right and the refusal was wrong at the same time. The owner authorised a narrow,
recorded, reversible exception for one run; it worked, it closed itself, and no other site's build
was affected by it.

The contact page the owner asked to be removed now returns "not found" at the public address, with
nothing linking to it.

Two things arrived that he did not ask for but will notice: the analytics tag and the cookie
consent banner are now live on the pages.

Along the way we filed two platform bugs from things this site exposed. One is a rule that quietly
parks the header-and-footer refresh for a site after it has succeeded twice in a week, which had
stalled that refresh on twelve sites. The other is a step that appends a junk row on every run and
now fails outright, which is why the refresh on this site has to be done by hand until it is fixed.

Last, we turned the fleet acceptance checklist into a script, ran it over all twenty pages, and
found nothing the owner's review had not already caught.

## Where we are now

The site serves what we say it serves. The checks that could have gone badly went well: the
owner's personal email is nowhere on it, no page is orphaned, no picture points at a missing file,
the logo has a real alpha channel, there are no leftover placeholders.

What is still wrong is what the owner already knows about and what other threads own: the articles
promise dated news and deliver general essays, the comparison tool asks the reader for all its
data, the fight calendar has no calendar in it, and every page except two carries a single picture
— the logo — with one hero image shared across seven pages. All four have the same root: nothing
in the system yet gathers dated, checkable facts about a subject, and that work is now staffed.

One small thing is with the owner right now. The repaired call-to-action says "the calendar below"
and there is nothing below it. The framework's copy editor is rewriting that sentence, and by
design the proposal stops in his review queue for approval before it reaches the page.

Delivery is still held, exactly as he ruled. No customer access token has been issued.

## Where we are going

Three steps, in order. The owner approves or rejects the one sentence waiting in his queue. Then
he decides whether the remaining four items are close enough to send, or whether delivery waits
for the fact-gathering work — that is his call, not ours, and the read-out that supports it is
kept honest in three columns rather than two. Then the delivery rehearsal itself runs end to end
for the first time on a real order: review, approve, cut the zip, send the email to the address on
the order rather than the one on the site.

The thing worth keeping from this week is not any of the fixes. It is that every one of them was
believed done at least once before it was done, and what caught the difference each time was
looking at the artefact — the served page, the image file, the row — rather than at a status that
said complete.
