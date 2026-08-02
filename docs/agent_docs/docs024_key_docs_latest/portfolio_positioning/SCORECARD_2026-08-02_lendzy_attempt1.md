# SCORECARD — lendzy.co.uk pipeline attempt 1, against the loancash benchmark

**2026-08-02 evening.** First-ever pipeline build of this portfolio, scored against
`RUBRIC_2026-08-02_loancash_benchmark.md`. Timeline: submitted 10:16Z, full cascade ran
unattended (and unwatched — the kubeconfig token was dead 11:20Z–~18:26Z, so the entire
build happened in the blind window), 20 pages planned, 15 built and committed to the
sites repo (`origin/master`, "Rerender: …" commits), 5 items parked in
`needs_human_review`. Scored from the BUILT ARTEFACTS in git, not from component rows —
two of my own component-level counts were falsified by the artefact (below).

**Attempt 1 measures the MISSION SEAM ONLY.** The specs were never seeded (the cascade
outran the plan during the outage — and the hold poller was watching for
`needs_content_page`, while the live pipeline creates **`needs_page`**: the trigger
script's own header documents the former. The seed is not the system, again). Spec
marker baseline correctly 0.

| # | dimension | verdict | evidence |
|---|---|---|---|
| 1 | nav model | **PASS** | Exact mission labels: "Check your loan" / "Your rights" / "Free help now", the third a DEEP LINK to `/cant-pay.html`, not a hub. Zero directory-shaped refs site-wide. One addition (About). |
| 2 | every-page invariants | **FAIL** | Footer is a generic Quick-Links dump. Independence line on **3/15** pages, does-not-lend on **2/15** — present only where an individual page writer chose to put it in body copy. No chrome mechanism carries mission constraints. Free-help orgs in footer: none (in body on cant-pay + breathing-space). |
| 3 | facts with rules | **STRONG** | CONC 5A on 5 pages, CONC 5.2A on 4. price-cap.html opens answer-first with the rule named and a check-the-source pointer (`fca.org.uk/handbook/CONC`), worked £200 example. Loan-shark numbers on the two right pages. **Number test PASS: zero unsanctioned figures in prose across all 15 pages** (early 82/70/90 "violations" were CSS widths — strip scripts/styles before a number census). |
| 4 | copy register | **PROMISING** (sampled) | Sampled page is answer-first and protective. Full judged pass across 15 pages still owed. |
| 5 | tools | **FAIL — the headline** | None of the 3 tools built. All parked: "Owned page tool-price-cap-checker is not_built — **needs owner-aware build, not the generic builder**" ×3. Every page links to them: **24 dead links** on the deployed artefact. |
| 6 | structural validity | **MIXED** | ld+json parses 15/15. Internal refs clean EXCEPT the 24 tool/blog links to unbuilt pages. **Canonicals 0/15 — entirely absent** (checked any-attribute-order). |
| 7 | compliance boundaries | **PASS** | No `<form>` anywhere, no lead-gen, no lender recommendations found. |
| 8 | markers | mission **0/15** · spec **0 (correct baseline)** | The exact-phrase homepage instruction ("know the rules before you borrow") did NOT survive mission→classifier→writer — not even the fragment "before you borrow". Verbatim elements attenuate through the research chain. |

Also: the pipeline imposed **blog-index + blog-post** pages the mission never asked for
(both unbuilt, parked in review) — the standard site shape leaking past the brief.

## Measurement corrections (kept visible)

1. I reported "independence line 0/20" from component-level greps for the SPELLED-OUT
   name. The artefact says 3/15, phrased "not affiliated with the FCA" — a grep proves
   absence only for its spelling, and chrome added at assembly is invisible in
   component rows. **Score the artefact.**
2. I nearly reported market-rate violations from a number census that included inline
   CSS/JS. Stripped, the prose is clean.

## What the diff points at (seam list for attempt 2+, in priority order)

1. **The owned-page/tool builder** — `owned_page_review`'s own summary names the gap:
   "needs owner-aware build, not the generic builder." For this portfolio the tools ARE
   the product. Options: build that path, or the earned hybrid (hand-built tools in via
   locked components, pipeline builds everything else).
2. **Every-page invariants need a chrome-level carrier.** Per-page writers cannot
   produce "on every page"; footer/chrome generation must accept mission/spec
   constraints (compliance lines, footer link groups).
3. **Canonicals are simply not emitted** — head-template gap, cheap fix, direct SEO
   value fleet-wide.
4. **Links to planned-but-unbuilt pages ship** — assembly links the plan, not the
   built set. Either sequence links after builds, or block deploy while flagship pages
   are unbuilt.
5. **Verbatim elements attenuate through the mission seam** — exact phrases (brand
   tags) need a spec- or chrome-level carrier. Attempt 2 (seed L10 specs, regenerate
   one page, marker check) measures whether the spec seam carries them better — that is
   also the #16 proof.
6. **The standard shape leaks** (blog pages off-mission) — find where the planner gets
   its default page set.

## Standing state

Site `8ff093d5-1f19-453b-9439-a10379bbcd76`, publicly unreachable (no Cloudflare zone —
by design). 5 items in `needs_human_review` are an OWNER QUEUE: 3 tool builds
(recommend: hybrid route or fix seam 1 first), 2 blog pages (recommend: reject as
off-mission). The platform committed the build to the sites repo master — expected
deployer behaviour, harmless while unreachable; tidy only when the experiment ends.
