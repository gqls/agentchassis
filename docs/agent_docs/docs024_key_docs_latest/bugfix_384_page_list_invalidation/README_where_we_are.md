# Where we are — bugs_open/384 (plain prose, append-only, newest at the bottom)

## 2026-08-24 evening — picking the bug up

The bug: when an article gets its little listing picture (the "card"), the page that shows the list of articles is never told to go and look again. The list is a snapshot taken when the page was last built. A routine re-render of that page just re-ships the snapshot, so the card stays missing until something unrelated happens to rebuild the list properly. The dartsonline home page showed 4 of 12 cards without pictures for two days; the person who found it fixed that one page by hand this evening, so the site looks right now — but the mechanism is unchanged and the next card will do the same thing.

What I found on top of the filing: the same snapshot problem covers more than cards. The list picture falls back to the article's big hero image when there is no card, and a hero landing does not tell the list either. And the "tool" strips on several sites hold 14 stale entries right now — nobody sees it only because that strip doesn't display pictures.

The fix I'm building is one shared piece: "the data behind an article/tool list changed on this site — go and re-resolve every page that shows such a list". The card-maker and the hero-landing step both call it. A nightly-style check will also compare each list against a fresh answer and file a re-render if they differ, so any future producer we forget about is still caught. Pages marked as owned by a tool are skipped, because re-rendering those fails on purpose.

I've told the four other sessions whose files sit next to mine what I'm doing. The person who filed the bug is still online but working on something else; I've written into the bug file so they see it.

## 2026-08-24 late evening — the fix is written and tested; waiting on the review and the next build

The shared piece is built. When a card is made for an article, or a page's big picture lands and no card is coming, the code now finds every page on that site that shows a list of articles or tools and files a "re-resolve your list" request for each one. Those requests go through the same re-render machinery the news feed already uses, so nothing new has to be learned by the system — it just gets told, which it never was before. Pages that are owned by a tool are skipped because re-rendering those is refused on purpose and would just produce failures.

I proved the tests bite: I deliberately broke the code five different ways (declared a wrong source, un-declared a right one, removed the owned-page rule, deleted each of the two calls) and every time the right test went red. Then I put it back and everything is green.

Four other sessions checked my numbers while I worked, and three of the numbers I had written down were wrong or half-wrong — two of them in ways the other sessions themselves got wrong first. All corrected before anything was committed, and each mistake is written up with the check that would have caught it. The plain lesson: a number handed to you by another session is still just a number in a document; run it yourself before it goes in with a date on it.

Next: the reviewer council is looking at the change (budget half an hour), then it's committed so the next chassis build carries it. It does nothing until that build rolls. After the roll I'll prove it on a real card landing rather than trust the tests. The "safety net" sweep that would catch any future producer we forget is Phase 2, not yet built.

## 2026-08-24, later — reviewed, approved, and the safety net built too

The reviewer council approved the main fix (13 reviewers, none unreadable). Two of their advisory comments were worth acting on the same evening rather than filing:

- One reviewer asked how many pages a single card landing could send for re-render. I measured it: on one site, 26 — of which only ONE actually shows the picture. The other 25 are "tool strip" components that store the picture but never display it. So the shared lookup now only counts pages whose template actually renders the image (that brings every site down to 0–3), and there is a hard cap of 24 per landing that shouts in the log if it ever bites.
- Another reviewer said this "who consumes this data" lookup is the first general one in the platform and should be named as shared infrastructure rather than discovered later. I wrote a short RFC (052) stating the seam and the open question — whether the two older producers that hard-code their one consumer page should move onto it.

The safety-net sweep is also built now: a discovery check that compares each stored list against a fresh answer and files a re-render when a picture differs. It deliberately never "closes" anyone else's re-render request — I measured that no producer has ever retracted a re-render on this item type across 18,360 of them, so this would have been the first, judging 121 other producers' requests. Turning the sweep on is a held migration (603), to be applied by hand only after the build that carries the check has rolled.

Everything is committed (three commits) and the second review round is running. Nothing is live until the next chassis build rolls; after that, the proof is an induced card landing producing exactly the expected re-render requests — written up in the lane RUNBOOK.

## 2026-08-24, end of session — both reviews approved; nothing live until the next build

The safety-net sweep went through three review rounds tonight. Each round found something real: first, a place where an unreadable stored record could have been mistaken for "nothing to compare" (fixed and tested with a case that actually proves it); then two numbers I had written from memory instead of measuring — the migration now measures its own starting point when it is applied, and one status value I described as live does not in fact occur. Third round approved.

Where this leaves things: five commits on the branch, all reviewed, the build proven from a clean copy. None of it does anything until the next chassis build rolls. After that roll there are four things to do, written up in the lane's RUNBOOK: prove the roll, run the induced card-landing test, switch the sweep on by hand (migration 603 — it checks its own preconditions and tells you what to watch), and re-read the escalation number a week later.

Two open questions for you, neither urgent: RFC 052 asks whether the two older "tell my one listing page" mechanisms should move onto the new shared lookup; and the `rebuild_blog_listing` step still writes empty images into blog listings — harmless today (no blog index lists a page with a card) and now caught by the sweep when it fires.

---

## 2026-08-25 (afternoon) — you decided all four, and all four are done

You answered the four open questions: turn the sweep on, generalise the lookup now, fix the
action, and fix the tool-cta entries by changing the template. Then, when I measured what the
template change would actually look like, you made a second call: derive the missing images
first. Here is where each one landed.

**1. The sweep is on.** I applied the held migration by hand after re-checking that the running
software actually knows about the check — it does, on 301 pods. The sweep has run twice since
and, exactly as predicted, found nothing to fix. That is the expected result: it is insurance
against a future producer nobody wires up, not a repair of anything broken today.

One caveat I want to be straight about. To know the sweep is *working* rather than *blind*, I
need to see it report "I looked at N listings and all N were current". So far it has only
visited a site with an empty listing, where the honest answer is "nothing to compare". So the
proof is still outstanding — not failed, just not yet obtained. It should arrive as the rotation
reaches a busier site.

**2. The lookup is generalised.** This was the piece of shared machinery that answers "who is
using this data?". It could only answer that question about images. It can now answer it about
any kind of data — news items, directory entries, products — and the two older pieces of code
that had their own private answer now ask this one instead. Reviewed and approved first time.

Two things worth telling you. First, I found that a claim we had written down about those two
older pieces was **wrong**: we had recorded that each one had a single page name hard-coded into
it, and neither did. They were less broken than we had said. I have corrected that in the
document where it was written, because it had already been quoted onward into an architecture
proposal. Second, the test I wrote to keep this honest immediately caught an error in my own
work — and then turned out to be looking through a keyhole, reporting something correct as
wrong. I fixed the keyhole rather than excusing the case, because a test with a blind spot is
worse than no test.

**3. The action is fixed — and the review caught something real.** The blog listing rebuild was
blanking out the images. The reviewers sent it back once, and they were right to. My own code
comment described a way the thing could fail silently — if the data shape ever shifted, the page
would quietly keep its old contents and report success — and I had written that down as a known
risk without actually closing it. It is closed now: that case fails loudly instead.

They also asked why a query had no limit on it. Checking, I found something better than the
question: the same listing has two pieces of code that build it, one of them capped at 24 items
and the other uncapped. On one site with 40 articles they would have disagreed by 16. Both now
share one setting.

Two of the reviewers' own checks disagreed with numbers I had put in the submission. I re-ran
both rather than argue. One of them was right and I was too narrow — I said "47 images are
blank", meaning the pages this code writes; across the whole estate it is 55. The extra 8 are on
pages this code never touches, and I checked each one: they are blank because those articles
genuinely have no picture, which is correct. The other check was mine to stand behind, and I
have the row that proves it.

**4. The tool strips now show pictures.** This is the visible one. When you said "change the
template", I measured what would actually appear and found a problem worth putting back to you:
144 of the 228 pictures would have been full-width page banners squeezed into small cards — all
on loancalculator, whose tool pages had no proper thumbnail at all. You said derive the missing
ones first, so I did.

That worked for loancalculator — all ten pages now have a proper cropped thumbnail, and the
banner problem is gone entirely. It did **not** work for loanandmortgagecalculator, and the
reason is worth knowing: the thing that makes these thumbnails **crops an existing picture**. It
does not draw a new one. Those 19 pages have no picture of any kind to crop, so the job
completed honestly having produced nothing. Their strips will simply show text, which is the
designed fallback and looks fine.

Everything else is live and I have checked it on the actual pages, not just in the database: six
pages re-rendered so far, every one showing real thumbnails, and none showing an empty picture
box. The rest are working through the queue.

**One page refused to re-render, and I want to flag it rather than bury it.** A page on
ai-agent-orchestration was stopped by a safety guard: re-rendering it would have thrown away
half its layout, so the system refused and wrote nothing. That is the guard doing its job. It is
**not** caused by anything I changed — that same page already failed three times yesterday, and
those were the only other refusals anywhere in a fortnight. Something about that one page's
stored version cannot currently be rebuilt from its own source. It is a real problem, it is
someone else's to pick up, and nothing is damaged in the meantime.

**Something I bumped into that is not ours.** A test the house rules tell every author to run
before committing cannot be run at all — it refers to a name that was renamed two days ago and
never updated, so it fails to compile. It is another team's file and I have not touched it, but
it means that particular check has been silently unavailable to everyone since 23 August.
