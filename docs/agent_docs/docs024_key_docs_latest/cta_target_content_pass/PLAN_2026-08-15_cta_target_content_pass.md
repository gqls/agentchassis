# PLAN — CTA target content pass (commissioned by the owner, 2026-08-15)

**The commission (owner, in chat, 2026-08-15, to the 268 lane):** "accept it
as a floor for now and now commission a content pass to vary the targets
page by page to the most appropriate tool." Spun out of
`bugfix_268_cta_buttons_fleet` (closed) — this lane inherits its context but
is a NEW deliverable, not a bug.

## The problem, precisely

The 2026-08-15 resolution re-run gave every resolvable label-less CTA a real
destination. Where the button's wording named a real page, label-match chose
it (good). Where it did not, the positional fallback chose **the site's
top-ranked interactive page** — the same one for every such page on the
site. Measured 2026-08-15 (query in RUNBOOK): **16 sites have ≥6 rows on
their modal target; worst are finetuning.uk (39 rows on
/tools/password-entropy.html), ai-agent-orchestration.com (36, same tool)
and gaswholesalers.com (28)** — and `/tools/password-entropy.html` is the
modal target on THREE sites, sometimes topically absurd (an AI-services
page's main button pointing at a password checker). Everything works; it is
repetitive and occasionally off-topic. The owner accepts it as a floor; this
pass raises it.

## Mechanism — compose two existing pieces, build nothing (candidate 1)

The framework must write the content (owner rule 2026-08-06) and topical
choice is content judgement, so:

1. **Writer step:** per page, a `content_rewrite` (`mode=edit_live`) whose
   `content_guidance` lists the site's REAL tool pages (name + url + one-line
   purpose, generated from `pages`) and asks the writer to reword the CTA
   LABELS to name the most appropriate tool for THIS page's topic — labels
   only, prose untouched.
2. **Resolver step:** the existing `cta_links_stale` re-render label-matches
   the new wording to the tool it names (`applyCTARecompute`, the same
   machinery the re-run used) and writes the url. The 268 carry keeps it
   through future rewrites.

No new Go, no new seam; the LLM never touches a url key (renderer-sourced,
writer never emits them). **Known caveats to design around:** label-match
overlap ties on incidental words (`bugs_closed/253`); the self-link gap and
double-target quirk recorded in `bugs_open/248` (a fix there would land
before or during this pass, ideally); `content_guidance` had "no readers" in
one filing (`bugs_open/271`) — VERIFY the writer actually receives it on
this item type before trusting step 1 fleet-wide (the 268 D1 rewrite DID
follow its guidance on 2026-08-15, so it reaches page-build-handler items;
re-verify per handler).

## Phasing

- **Phase 0 (done here):** population measured; mechanism candidate named.
- **Phase 1:** one-site canary (suggest finetuning.uk — worst offender,
  fresh in context): tool-list guidance generator + dispatch + matched-pair
  verify (labels changed, urls follow the labels, prose untouched, no valid
  link lost).
- **Phase 2:** fleet, site by site, same recipe.
- **Open question for phase 1:** whether to widen
  `candidatesFromHubs` to guide pages (the 248 note's alternative) instead
  of relying purely on wording — that IS a Go change and would need the
  council gate; decide after seeing canary label-match quality.

## Decisions and their reasons

- 2026-08-15: floor ACCEPTED (owner) — this pass is quality, not repair; no
  urgency, no census to drive to zero.
- 2026-08-15: composed-mechanism candidate chosen over new resolver code
  because both halves are live, proven this week, and the pass stays
  reversible page by page.
