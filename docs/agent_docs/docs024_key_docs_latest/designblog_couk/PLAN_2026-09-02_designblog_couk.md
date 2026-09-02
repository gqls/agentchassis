# PLAN — designblog.co.uk lane

**Opened:** 2026-09-02, owner-created ("Please make this thread about the
designblog.co.uk site"). Session name: "designblog.co.uk".

## What this lane IS

The site-quality owner for **designblog.co.uk** (remake №4 of the portfolio
positioning lane's 22 hosted-site remakes; built and gone live 2026-09-02 —
build record in `docs/agent_docs/docs024_key_docs_latest/portfolio_positioning/`,
brief = `portfolio_positioning/BRIEF_2026-09-02_designblog_co_uk_for_review.md`).
This lane corresponds with the portfolio positioning thread (which built the
site) and carries the owner's critique to the mechanism threads whose fixes the
site needs. Much of the critique is a **class**, not a designblog defect — the
owner said explicitly it applies to advertise.co.uk and websitepromotion.co.uk
too — so the default posture is: route class defects to the owning mechanism
thread, hold the designblog-specific instances here.

## What this lane is NOT

- Not the builder. The pipeline builds and rebuilds the site (OWNER RULING
  2026-08-04: every site goes through the framework; hand-patching a served
  artefact dies at the next rebuild).
- Not the owner of the design-sameness mechanism (site design planner /
  components / theme kits threads), the empty-listing-page class (experience
  loop detectors + portfolio positioning build pipeline), or copy voice (copy
  quality two stage). Contribute; don't compete.

## Decisions & phasing

1. **P1 (done 2026-09-02):** verify every point of the owner's critique against
   the served site before relaying it — all confirmed, see
   `CRITIQUE_2026-09-02_owner_site_review.md` §2 (this directory).
2. **P2 (done 2026-09-02):** route each point to its owning thread via
   cross-session messages (7 threads), and confirm receipt — the owner asked for
   receipt to be checked, not assumed.
3. **P3 (open):** track each routed point to a mechanism-level answer: why did
   four listing pages ship empty; why did brief-echo copy pass the copy gates;
   why no tools nav link; what changes make the next remakes visually distinct.
   Designblog-specific repairs (nav link, page content fills, copy rewrites)
   happen through the framework once the owning threads say which seam fixes
   them.
4. **P4 (open):** re-verify at the served artefact after fixes land; a green
   status is not a repaired page.

## Open questions for other threads (asked 2026-09-02)

- Portfolio positioning: are the empty listing pages a convergence tail (like
  seotools' tools arriving by rotation) or a defect — does the pipeline have any
  mechanism that ever fills the directory/glossary/inspiration/feed with items?
- Experience loop: did the listing-class and experience-promise detectors run on
  the four 09-02 remakes; if yes, what did they conclude; if no, why not?
- Copy quality two stage: how did brief-echo copy ("explains the brief, doesn't
  answer it") and the AI-sounding closers pass the two-stage gates on a build
  that postdates their 08-31 roll?
