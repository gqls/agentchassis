# SUMMARY — 2026-08-16 — the required-fields repair router: built, live, drained; the triage promoter restored; the router engine scheduled

*The first summary for this workstream. Milestone: the mechanism the owner ordered on the
morning of 2026-08-15 is live and has done its work; two owner decisions taken since have
changed the platform's shape beyond this one type. Written to be read aloud.*

## What we are trying to do

Stop the human-review queue filling with findings the framework should resolve itself. The
owner ruled on 25 July that "the framework — not a person — should resolve every one of these
classes", and on the morning of 15 August ordered the first concrete instance: a fleet-wide
repair handler for `required_fields_missing`, the finding that says "this page component is
missing fields its template declares required". Forty-four of them sat unread, one on the
owner's own worked example (the gas unit converter), and nothing in the fleet claimed the type.

## Where we have come from

The type had been filed *flag-only* by design — no handler, parked straight into human review —
on the reasoning that the honest fixes ("give the site a data source, or remove the component")
are human decisions. That reasoning was sound for some of the findings and useless for the
rest, and because nothing ever looked at them, nobody knew which were which. The estate had a
pattern for exactly this problem from three days earlier: the two image "routers" the owner
asked for on 12 August — small config-only agents that ask the database one question per
finding and then close, park, or hand off to a repairer that already exists. This workstream
built the third router on that pattern, and it went through the council four times.

## What we have done

- **Measured before designing.** Of the 44 findings, 35 were not "missing content" at all but
  pages serving perfectly well from a single stored block of HTML — pages that an automatic
  "repair" would have *destroyed* by regenerating one template section over the whole page. Six
  pointed at components that no longer exist. Only three were genuinely repairable, and one of
  those (the gas converter) is a tool page the platform deliberately refuses to rebuild
  generically. So the handler is a router, not a repairer.
- **Built the router** (seed 410, register CQ-023): eight routes. It closes findings whose
  premise no longer holds, *with the evidence*; it parks the four classes a machine must not
  touch back into the review queue *carrying their classification, the danger, and the safe
  options*, pinned so the system cannot re-raise duplicates; and it converts the two repairable
  classes into work the page builder already owns. New findings route themselves from birth.
- **Let the council make it better.** Four REVISE rounds; two of them found genuine design
  errors, both then proven live and turned into routes: a repair that would have asked a prose
  writer to invent image addresses (the validation gate refused it), and a page rebuild that
  would have quietly produced nothing (it did). The council also forced the backlog assignment
  into a reviewed, guarded SQL file, and caught a stale figure of mine (logged in WRONG_CALLS).
- **Drained the backlog**: all 44 through the router — 36 parked with their facts, the rest
  closed on evidence; zero left unrouted, zero blocked.
- **Took the two decisions the council could not settle to the owner, and actioned both:**
  1. *The triage promoter.* The council objected that findings born "claimable" skip the
     observe-only stage — true, but the promoter that stage depends on had been switched off
     since May (bug 083), stranding findings forever. The owner ruled: rebuild the promoter as
     its own scheduled task. It is live (seed 430, register SCH-026): every fifteen minutes it
     promotes at most twenty findings whose handler is real and has completed that kind of
     work before; a brand-new type is held until a human runs one by hand. Overnight it drained
     the stranded pile from 70 to the 4 it correctly holds; 93 of 100 promoted findings
     completed. This workstream's producer went back to the proper "born unclaimable" convention
     the same hour, and that change is now live on the fresh chassis.
  2. *One engine, many classifiers.* This was the estate's third near-identical router. The
     owner said each handler should be modular and responsible for its own thing, and agreed
     that the modular unit is each type's *classifier*, run by one shared engine rather than a
     growing family of cloned agents. That is RFC_030 — ruled, and set up as its own lane
     (`router_engine/`) with a plan and a cold-start handoff; nothing built yet, deliberately.

## Where we are now

The mechanism is fixed and live; the backlog is clear; the promoter is live and has proven
itself once. Two council rounds are owed (a short closing round on the router's trail now that
both open questions are owner-ruled; a round 2 for the promoter with every objection already
measured and favourable). Bug 277 stays open only for its own week-long checks (no churn; the
two cancelled repair attempts re-raise and park correctly). Bug 083 stays open for two of its
three verify criteria (a first-ever `phantom_internal_link` completion; a spot-check of promoted
completions at the live page). Bug 033 — the queue itself — has its first "should not fill"
mechanism and a dated contribution block; its other pieces (the Retry button on handler-less
items, the owner's decisions B and D, identity) remain for other sessions.

## Where we are going

The router engine is the next real piece of work: a design round with the council on its shape,
then migrate this router first (its eight routes define the contract; its census and canaries are
the regression fixture), then the two image routers, which inherit its hardening. Meanwhile the
promoter should get one small door-closer (never promote diagnose/report-pipeline items) as a
fresh numbered migration, and the two other producers that adopted the born-claimable workaround
can return to the convention now that the stage exists. Full priority list:
`HANDOFF_2026-08-16_continue_here.md` beside this file.
