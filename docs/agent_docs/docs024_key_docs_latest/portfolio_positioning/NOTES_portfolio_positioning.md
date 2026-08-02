# NOTES — portfolio positioning

Append-only, newest at the bottom. Evidence, commands, what the system said, and every
misstep. The missteps are the point.

---

## 2026-07-31 — session 1: the framework, the register, and the check

### Ground measured before any design

- Owner supplied ~154 domains (82 finance listed singly + an insurance block compressed
  with `/ .uk / rates` shorthand, expanded to 72; `mortgagerepaymentsinsurance.*` appears
  in both lists → **152 distinct**).
- Clustered: **42 propositions**. Audit: every domain assigned exactly once, none
  invented, none double-counted (the assignment script asserts all three).
- Live DB: only `loancalculator.co.uk` and `loanandmortgagecalculator.co.uk` have `sites`
  rows; only those + `mortgagecalculator.co.uk` are in the deploy repo. **~150 greenfield.**
- **31 of 82** finance domains sit in propositions with **no differentiating axis in the
  name** (12 savings-rate variants, 8 loan generics, 6 rate-comparison, tool twins) — the
  number that shaped the whole framework: differentiation must be *assigned and recorded*,
  because for a third of the portfolio the name gives none.

### Owner decisions (P1–P4)

Thick sites (a correction — an earlier answer said thin; the correction supersedes),
roughly one per proposition, insurance included now, differentiation recorded from the
start. The deliverables: `PLAN_2026-07-31_differentiation_axes.md` (the axis catalogue,
three tiers, enforcement design), `REGISTER_positioning.md` (42 entries + machine-readable
claims table), `check_register.py` (the overlap guard the platform lacks),
`PORTFOLIO_domains.txt` (the canonical list).

### The checker caught ME twice, same day as the lesson was written

1. **First run: 71 violations, most false.** The prose register used shorthand
   (`rate(s).co.uk/.uk`, "+" lists) my parser could not read, so 68 domains looked
   unclaimed and one "domain" was the literal string `co.uk`. The tempting fix was to make
   the parser cleverer — the same shape as this morning's
   `fixing-a-checker-to-agree-with-a-broken-site`, from the other side. The right fix was
   to make the ARTEFACT checkable: a ```claims``` block, one entry per line, full
   spellings, which the checker parses *instead of* the prose. Prose for humans, claims
   for machines, and the two are cross-checked (every claims id must have a prose entry
   and vice versa).
2. **Then it caught a real one**: B9 (banking equipment) had no neighbours line — I had
   treated "different market" as an excuse not to say where its ground ends. Fixed in the
   register, not the checker.
3. **And my mutation test lied to me for one message**: `python3 … | tail -3; echo $?`
   reads **tail's** exit code, so the mutant printed FAIL and "exit=0" in the same breath.
   Re-ran with the pipe removed: exit=1. The fleet already has this class written down
   (`a-quiet-test-passes-when-the-rule-is-gone` is the cousin); the local lesson is that
   **a mutation test is itself a measurement and can be mis-instrumented** — read
   `PIPESTATUS[0]` or don't pipe.

### Verified state at end of session

| check | result |
|---|---|
| portfolio list vs claims table | **152 = 152**, 0 unclaimed, 0 extra |
| claims ids vs prose entries | **42 = 42** both directions |
| double-claims | 0 |
| entries with no neighbours/HOLD | 0 |
| coordinate collisions (family × audience × mode) | 0 across 26 claimed coordinates |
| checker mutation-tested (double-claim injected) | **exit=1, names both entries** |

### Open items this session deliberately did NOT decide

- ~40 ⚑OWNER twin-pair decisions (301 vs accepted duplicate) — rolled up at the end of the
  register; default recommendation is 301 to the built sibling.
- Build order across the 42 propositions — commercial data the owner holds.
- Whether `saving-rate.co.uk` takes the B6 "your saving rate" sense or joins B1's twins.
- `loancash.co.uk`: recommended **not building** (payday-adjacent name; the audience it
  attracts is the one FCA rules protect hardest) — owner has not yet ruled.

### [UNVERIFIED] / assumptions to keep honest

- Proposition clusters were assigned by reading the names, not by search-volume or SERP
  data. Where the owner has keyword data, it may re-rank *build order* but should rarely
  move a domain between propositions.
- The claim that "preparing" and "owning" journey stages are underserved is editorial
  judgement, not measured against competitors.

### 2026-08-02 — OWNER CORRECTION: the pipeline should be building everything

The owner asked whether loancash.co.uk was built by the submit-domain trigger or by this
thread. Answer: entirely by this thread; the trigger ran once, afterwards, in locked
adoption mode (zero LLM work items — generation structurally excluded). The owner then
ruled the error: **hand-building is not the method — the pipeline should be building
everything.** Recorded as `SUMMARY_2026-08-02_the_pipeline_should_build_everything.md`,
a CORRECTED marker on `SUMMARY_2026-08-01c` (which had credited "the machine" with the
full cycle), and a WRONG_CALLS entry (a success metric that flipped sign: zero generated
items passed adoption fidelity and refuted "the machine built it", in the same file).

Consequence for the lane: before any gap-closing, run the fresh-build experiment — one
cheap domain via `082_submit_domain_unified.sh` (no `--from`), register entry as
`--mission-file` AND pre-seeded via `set_divergence_specs.py`, marker sentence planted,
live-origin `verify_site.py`-class checks on the output. The six-gap list (README
2026-08-02 entry) is the candidate work plan; the experiment decides which gaps bite.

**Same day, sharpened by the owner:** the experiment is not "one cheap domain" — it is
loancash itself: run the same proposition through the pipeline and fix the pipeline
until it matches the hand-built site (nav, research, articles, copy). Safety rule: the
fresh build runs at a SHADOW domain (`loancash.uk` — unclaimed, no Cloudflare zone,
publicly unreachable), never at loancash.co.uk, whose row is live and locked. Before
the run: read the LIVE `domain-submitter` agent_definitions row (seed ≠ system) to
confirm what it does on an existing-domain collision, and expect its sites-repo commits
to need tidying afterwards. Benchmark rubric = extract from the hand-built site:
nav model, every-page footer disclaimers, fact-with-rule-name density per guide, copy
register, tool correctness fixtures (£15 / 0.8%/day / 100% ceiling).

**Shadow domain corrected by the owner, same day:** `lendzy.co.uk`, not `loancash.uk` —
owner-held, in neither the register nor PORTFOLIO_domains.txt (grep exit 1 on both, all
docs). Sharper test: no loan-cash name signal, so the L10 positioning must arrive
entirely via mission brief + pre-seeded specs (the gap-1 question itself). Still no
Cloudflare zone → publicly unreachable. Task #20 updated; summary carries the CORRECTED
marker.

### 2026-08-02 — lendzy.co.uk shadow build FIRED

Submitted fresh (no --from) with MISSION_2026-08-02_lendzy_shadow.md:
`SUBMISSION_CORR=a62da5d3-3e08-450a-a981-bc002f7ea2cd`, orchestration
`fac748c0-3fa0-494f-96e0-9137560d7d48`. Queue was QUIET — domain-submitter COMPLETED
within ~2 min (the ~29-min budget is a ceiling, not a constant). Site
`8ff093d5-1f19-453b-9439-a10379bbcd76` (active/pending). Submitter wrote `submission`
(3,175 ch, carries the mission) + `mission_brief` (3,108 ch); one `needs_domain_research`
item triaged for domain-research-classifier. Hold poller armed 6h on
needs_content_page/needs_tool_recreation (scratchpad hold_lendzy.sh, defer-on-sight).
Seeding plan: after classifier/strategist write identity/content_direction, run
set_divergence_specs.py with a lendzy payload (L10 + SPEC MARKER exact phrase "checked
against the FCA handbook, rule by rule"; mission marker is "know the rules before you
borrow" — distinct phrases so output attributes to seam). Baseline-zero grep of
page_components for BOTH markers before any release. Benchmark scoring =
RUBRIC_2026-08-02_loancash_benchmark.md.

**11:20Z — kubeconfig token EXPIRED mid-experiment** (3-day expiry, owner-only refresh
from the Rackspace Spot console). Consequences: (1) watcher v1 false-fired on the error
text — replaced by v2 which matches row shape and waits through outages (misstep in
WRONG_CALLS); (2) the hold poller is BLIND until refresh — it self-heals (fresh kubectl
per loop) but while blind nothing defers content items. Mitigating fact: every pre-11:20
poll returned empty, so identity/content_direction were still unwritten ~64 min after
submission — the cascade is early and the unguarded window likely not tight. No action
at the cluster is possible or needed from this side; the platform continues regardless.

### 2026-08-02 evening — attempt 1 BUILT ITSELF during the outage; scored

Token restored ~18:26Z; the entire cascade had already run in the blind window (research
→ … → 20 pages planned, 15 built + committed to sites repo master, imagery, rerenders).
Hold plan doubly moot: outage AND wrong item type — live pipeline creates `needs_page`,
not the `needs_content_page` in 082's header. So attempt 1 = mission seam only; spec
marker baseline 0 as required. Full scoring in SCORECARD_2026-08-02_lendzy_attempt1.md.
Headlines: nav PASS (exact mission labels, deep-link help); facts-with-rules STRONG and
number test clean in prose; tools NOT BUILT (×3 "needs owner-aware build, not the
generic builder" — 24 dead links shipped); every-page invariants 3/15 (no chrome
carrier); canonicals 0/15; mission exact-phrase marker 0/15 (verbatim attenuates).
Two of my own counts corrected by the artefact (spelling-scoped grep; CSS in the number
census) — kept visible in the scorecard. NEXT: attempt 2 = seed L10 specs (they exist as
classifier rows — supersede) + regenerate ONE page + marker check = the #16 proof.

### 2026-08-02 late — owner ruling: build all 5 parked items fully; attempt 2 armed

Owner: "remove them from needs_human_review, they should be able to be built fully."
Mechanism read first: `reconcile_site_plan_action.go:37-40` parks tool-role pages BY
DESIGN (guard rail 1 of the experience loop — the generic builder clobbers owned
pages), and the parked items' own spec.fix names the sanctioned route:
tool-generator/create_tool_component. So the release = 3 new `needs_tool_recreation`
items (handler tool-recreation-handler, live conventions mirrored), blog needs_page ×2
back to triaged, owned_page_review ×3 closed as handled
(scratchpad/release_lendzy_items.sql).

**Chassis roll to v1.0.1229 landed mid-experiment** (owner warned). Attempt 1 ran on
v1.0.1228; everything from here runs on 1229 — a recorded variable, not a controlled
one. Release deliberately WAITS for roll-complete + 300s (dispatch near a restart is
silently dropped).

**Experiment design improved by the accident:** L10 specs SEEDED FIRST (gated dry-run,
then --apply; formatted 12,519→14,023 bytes, has_positioning=true read back), so the
tool + blog builds are themselves the attempt-2 probes — brand-new pages written with
seeded specs, against a measured 0/0 marker baseline. Spec marker appearing in any new
page = the #16 proof, on new writes, no regeneration needed.

**19:26-19:45Z — two missteps and a filed-bug finding during the release.**
(1) MY OWN hold poller deferred MY OWN three tool items within 20s of insertion — the
poller v2 window (to 04:17Z) outlived its purpose and its filter included
needs_tool_recreation. The class is `your-action-silences-your-own-detector`, inverted:
my guard silenced my release. Poller killed, items re-triaged.
(2) pkill self-kill AGAIN, new variant: the bracket trick guarded the pkill PATTERN but
the same command line carried the PLAIN path as a sed operand — pkill matched that,
exit 144, sed never ran. Rule: a pkill compound must not contain the plain target
string ANYWHERE in the command line.
(3) The blog pages are bugs 015-class, with evidence: page-build-handler no-op ("no
sections ready to build", attempt 2), and a sibling site's /blog/news-post.html shows
the same plan-template page unbuilt fleet-wide. The live 015 retype arm would flip
blog-index section-index→news-index and hand the blog to the NEWS pipeline — which
wires ongoing news generation onto a shadow site. OWNER CALL, not a silent wiring;
blog items left in needs_human_review meanwhile. Tools re-triaged 19:26Z, watcher on.

**19:55Z — #16 PROVEN.** about.html rebuilt post-seed with pages.content_direction
NULL: spec marker in 3 regenerated sections (hero quotes it verbatim). Site-spec seam →
writer prompt → saved copy, end to end. Also corrected earlier tonight: the mission
marker was in the homepage <title> all along (component-level census missed assembly
chrome — WRONG_CALLS'd, repeat offence). Site: 18/20 built + tools formula-verified;
remaining: blog (015-class owner call), favicon/meta-description/canonicals/chrome-
carrier seams, in-browser tool fixtures (needs serving).

### 2026-08-02 night — blog removed per owner ruling; chassis 1231; lane pausing for handoff

Owner: "I don't want a blog on this site." Executed in one transaction: 2
site_plan_pages rows DELETED (no soft column exists), 2 pages rows → status='archived'
(platform idiom, never built, 0 components), 2 items → 'rejected' with the ruling in
resolution_path. The 015-class finding stays on the seam list. lendzy open work items:
ZERO. about.html rerender (marker copy) COMPLETE and in the sites repo. Chassis now
v1.0.1231 (third binary today: 1228 attempt 1 · 1229 attempt 2 · 1231 current).
⚠ CLEANUP OWED: the acceptance_marker instruction is STILL LIVE in lendzy's seeded
content_direction — any future rebuild will keep weaving the marker phrase. Before
lendzy is ever real, strip that key and re-seed. Token load high → handoff written:
HANDOFF_2026-08-02_continue_here.md.
