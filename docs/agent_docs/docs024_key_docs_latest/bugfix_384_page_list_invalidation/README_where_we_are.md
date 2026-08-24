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
