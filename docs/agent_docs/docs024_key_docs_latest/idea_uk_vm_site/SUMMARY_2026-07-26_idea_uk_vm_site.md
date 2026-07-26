# SUMMARY — idea.uk (2026-07-26)

*Previous summaries: `SUMMARY_idea_uk_vm_site.md` (07-18, the migration) and
`SUMMARY_2026-07-25` (mid-build, four guides and one tool). This one marks the pipeline
COMPLETE. Written to be read aloud.*

## What we're trying to do

idea.uk is the place someone with an idea works it out properly: plain-English guides for every
stage of an idea's life, free tools that give an honest steer, and one paid product — the £29
Verified Idea Report — that everything funnels towards. It is also the site the rest of the
fleet will copy, one maturity rung at a time.

## Where we've come from

Ten days ago this was a migration project. Three days ago the site was nine sound pages with an
empty Guides section and one tool. Yesterday morning it had four guides and a patent checker;
by last night the whole journey existed.

## What we've done

The pipeline the owner asked for on the 24th is **built and live, end to end**:

- **Nine hand-written guides**, reading in journey order on a self-populating hub: creating
  ideas → building it → testing it → user acceptance → feedback loops → patents → copyright →
  funding ways → funding sources. The legal and funding guides state no figures by policy (a
  stale number is indistinguishable from an invented one); the AI-law section says plainly that
  the law is unsettled rather than guessing.
- **Two free tools built on the same safety rule** — gate before you score, because some answers
  are decisive on their own. The patent checker refuses to tell someone who has already
  disclosed publicly that they look patent-ready; the funding-fit finder tells someone seeking
  living-cost money the honest thing nobody advertises (almost nothing funds runway), and only
  past its gates composes a route map across grants, equity, debt and customer money.
- **The paid tool extended to match its own sales page** (owner's ruling): the report now leads
  with an assessment of the idea the customer actually submitted, carries "Check it yourself"
  source links under every finding, discloses its AI use in the report itself, and renders an
  honest "too early to assess" outcome instead of a padded verdict. Deployed to the box;
  first end-to-end run still pending (see below). The sales page got a light pass so it now
  describes both halves of what customers receive.
- **A day of found-and-fixed defects along the way**: the tools page's dead buttons, missing
  diagram and invented "8 free tools" statistic; the paid report absent from its own tool
  listing; the guides hub silently capping at six cards with nine guides live; and an
  unoverridable-fallback schema defect fixed fleet-wide twice, contributed into the owning bug.
- Everything authored is locked against regeneration; everything that lists is derived and
  picks up new content automatically — after nearly freezing one derived listing with a
  misplaced lock, which is now a written rule with a guard in every lock script since.

## Where we are now

The site is whole: 9 guides + 4 tool cards (the £29 report, two free finders, the audience
check), every button pointing somewhere real, every path ending at the report. What remains is
**operational, on the owner's box, not in the site**: five of the owner's own stale test orders
hold all five order slots (so /confirm answers "at capacity"), and the settings file sets the
contact address twice with the stale dead address winning. The fix is four pasted commands, then
one real end-to-end report to see the new format arrive. The deploy itself is verified good —
the new binary has been running since yesterday 15:11.

## Where we're going

Once the box is cleared: the first real report in the new format is the proof of the whole
funnel. After that the pipeline grows sideways rather than forward — more tools where stages
earn them, news/content feeding the empty News section, and the owner-flagged margin question
(each report now costs roughly double the AI spend). And idea.uk takes up its intended role as
the top-rung exemplar for the fleet-wide site maturity ladder being planned in its own thread.
