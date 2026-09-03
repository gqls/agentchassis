# Portfolio positioning — the brief that two agents could not read

*2026-09-03. Written to be read aloud. Previous: `SUMMARY_2026-09-02_first_four_remakes_live.md`.*

## What we are trying to do

We are rebuilding a portfolio of domains we already own into useful sites, using the framework rather
than by hand. Each site starts with a written brief that says what it is for, who it is for, and what
pages it should have. The owner reads the brief and approves it. The machinery then classifies the
site, works out a strategy, chooses a set of tools, plans the pages, and writes them. The brief is
meant to be the thing that governs all of it.

## Where we have come from

Four remakes went live yesterday. Today was meant to be the fifth, copyonline.co.uk, plus finishing a
set of tool pages that had been shipping as prose with no working tool inside them.

The tool work went well and is done: eight tools repaired and serving, six broken links between pages
and their components restored. The copyonline brief went through several revisions with the owner —
leading on artificial intelligence as both the threat and the opportunity for copywriters, a single
lead-generation page carrying the commercial aim so the other pages could stay editorial, and a
directory of copywriters. He approved it and told us to carry on.

## What we have done

We found that the approved brief was not reaching the machinery, and we spent most of the day
establishing exactly how far that went and what it had cost.

The finding, in the end, is narrow and fixable. Three agents read a site's brief. Two of them look for
it in one particular spot inside the brief. Briefs written by our own brief-writing agent do not fill
that spot in, so those two agents find nothing, say nothing, and carry on with whatever else they can
find. The third agent is simply handed the brief whole and reads it without difficulty.

For copyonline that meant the classifier read the site's old Drupal installation instead, decided the
site was a copywriter marketplace, and recorded in its own notes that no brief had been supplied. The
tool-chooser did the same and picked five general search-optimisation tools, three of which duplicate
another of our sites. Ten pages were built that appear nowhere in the approved plan of thirty.

We also got two things wrong ourselves today, and both are written up.

We told the owner those ten pages were live on his domain. They are not. Nothing has been published;
his old site still serves. We had read a page-level flag as though it meant the site had been
published.

More seriously, we told him the brief reached none of the agents and recommended interrupting the
build to fix it. That was wrong. We had measured the two agents that cannot read it and then spoken
about all of them. The third agent produced its strategy eleven minutes after we sent that message,
and it is faithful to the brief in detail, including two instructions the owner had given only in
conversation. His original instruction not to disturb a running build was the right call, and we spent
an afternoon pressing him to overturn it. We touched nothing, because he never answered.

## Where we are now

Nothing is published and nothing is damaged. copyonline's strategy document correctly describes it as
an authority site with a lead route, so the site's direction survived the fault. The build has since
stalled one step before the page plan, for an unrelated reason: the job that builds the page
composition started while the classification was still running, found nothing, asked for one, and
closed itself. Six comparable sites all have their composition. Copyonline is the only one without,
and nothing has restarted it.

The underlying fault is now understood precisely and is smaller than it first looked. It is two lines
of template asking for a piece of the brief that is not there, with a working example of the right way
to do it already in the system. Seven of our twenty-three current briefs have the same shape, written
by three different authors including ourselves, because nothing anywhere states which part of a brief
the agents actually read.

## Where we are going

The fix belongs to the thread that owns the underlying bug, and it has been handed over with the
evidence, the two candidate spellings, and a proposal for a check that would catch the whole class
before any agent runs. It is a small change that repairs every affected site at once, which is much
better than the site-by-site patch we had been proposing this morning.

For copyonline the immediate question is whether the stalled build picks itself up. If it does not,
restarting the composition step is a small and obvious thing to do, and then the page plan can be
judged on what it actually contains. Two decisions are still with the owner: whether to retire the
three tool pages that duplicate our other site, and where the lead route should ultimately point.
