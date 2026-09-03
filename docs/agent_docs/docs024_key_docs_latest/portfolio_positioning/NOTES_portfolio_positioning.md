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

### 2026-08-02 late night (fresh session from the handoff) — seam 1 built: the every-page carrier

Task #21 seam 1 (chrome-level carrier for every-page invariants), designed, measured,
tested, submitted. Mechanism read end-to-end first: the ONE live assembly path is
assemblePage (page-build-handler deploys via call_agent→page_renderer→
rerender_single_page; zero live agents use assemble_page/InjectFooter — measured), and
chrome is the stored site_components artefact, so the footer is the carrier. THE
MECHANISM ALREADY EXISTED, REGISTERED: STY-050 (per-site chrome config via gated
input_schema field, config.* → site_specs site_config; GTM on idea.uk was consumer 1).
**Near-misstep, caught before shipping:** I had drafted the council submission framing
this as a NEW seam ("shared_mechanisms" register category that does not exist) — reading
the register category file before writing the entry surfaced STY-050, whose own
CORRECTED block records exactly this failure (shipped, then found head-seo-standard had
the same pattern since 05-13). The check that caught mine: open the category file and
read the tail before inventing an entry. Submission reframed as consumer 2 + new
vocabulary key; source switched to the config.* idiom; STY-051 registered (entry +
index row).

Built and proven tonight:
- footer_compliance_lines_test.go — 4/4 PASS: unset byte-identical, empty-array
  byte-identical (navigateMap: empty array = unpopulated), set renders both lines,
  wrong-type contained (whole template drops to regex fallback, footer survives).
  Test old-constant md5 == live row (95801d67…), new-constant md5 == SQL literal
  (eea3fb69…) — the thing tested IS the thing applied.
- SQL_2026-08-02d_seam1_footer_compliance_carrier.sql — md5-drift-guarded template +
  input_schema update, lendzy site_config seed (2 mission lines), DO/RAISE asserts
  (a bare verify SELECT cannot stop a COMMIT), propagation item in section C
  (needs_rerender / rerender-pages / refresh_site_components — the shape that ran on
  lendzy today at 16:04Z).
- Prior-art sweep BEFORE shipping: compliance_lines 0 hits fleet-wide;
  footer-with-disclaimer_pre_037 active but 0 slots; vetcomparison content_direction
  asks writers for per-page footer disclaimers in PROSE (demand-side evidence; named
  in STY-051 as candidate consumer 3). NB my ILIKE '%footer_disclaimer%' matched that
  prose because _ is a single-char wildcard in LIKE — a lucky accident this time, but
  the match was NOT the key I searched for; escape underscores when the distinction
  matters.
- Concept-index headline did not reproduce: file said 1,711; the documented grep said
  1,705 pre-STY-051. Recorded in the index header per its own rule, not diagnosed.

NEXT in this session: 097 submit → commit (test + SQL + register + NOTES, pathspec,
Council-Submitted trailer) → apply SQL → propagation item → artefact census (both
lines on every built page, gqls/sites origin/master, strip script/style first —
NOTE the compliance <style> tag is INSIDE the block, so census the <p> text, then
separately confirm the div).

### 2026-08-02 late night (cont.) — seam 1 APPLIED + propagating; seam 2 committed; seam 3 rescoped

Seam 1 execution: SQL applied clean (drift guard passed, both DO asserts held);
propagation item 6c95718a (needs_rerender / rerender-pages / refresh_site_components)
fired at 22:33:47Z with chassis pods 24 min old. Stored footer chrome re-rendered at
22:36:11Z (2,593 → 3,008 bytes, FCA line present). **The item read `complete` at
22:36:17Z with 4/18 pages carrying the block — a `complete` work item is not a
repaired artefact, again:** the page rerenders trickle through the git-adapter queue
at ~30s/page ("Rerender:" commits landing one by one on gqls/sites master). Census
watcher armed; verify 18/18 before claiming the seam closed. Council run for corr
56ab6e23 live, `review_architecture` executing (the architecture seat is FIRING on
exactly the class it was seated for).

Seam 2 (canonicals + meta description) DONE as code, INERT until a roll:
- assemblePage is the single live assembly path (measured: zero live agents carry
  assemble_page; page-build-handler deploys THROUGH page-rerender). Canonical was
  emitted by NOTHING on any path (spelling-aware sweep of platform/).
- injectCanonicalLink mirrors injectPageJSONLD (named skips incl. fragment/query
  URLs; idempotent; respects an existing canonical either attribute order; identity
  byte-identical to the JSON-LD @id — asserted IN the test, not assumed).
- spliceMetaDescription: targeted fill of the ONE fleet-wide blank shape (measured
  by DISTINCT substring over every stored head), legacy first-empty-content fallback
  kept (its wart documented + deliberately unasserted), attribute escaping NEW,
  and correct-or-absent REMOVAL of the blank tag when the page has no description.
- Misstep, small: my first legacy-fallback test asserted a case (attribute-order-
  reversed tag) the ORIGINAL code never covered — the failing test corrected my
  belief about what "legacy behaviour" was. Kept as the documented non-assertion.
- Tree wouldn't compile mid-run: another session's in-flight edit on
  page_role_upsert.go (bug-175 lane, mtime seconds before failure). Verified via
  clean `git archive HEAD` overlay: seam tests + FULL actions package green.
- Committed 9c7a8e9e4 with Council-Submitted: 4cffcebb-9774-45e9-971c-0f057058f795.
  OWED after the next chassis roll: pod-grep positive control (e.g. the literal
  "injectCanonicalLink") AND negative control, then a lendzy rerender + census.

Seam 3 (favicon) RESCOPED by measurement — the platform half is NOT missing:
discovery's check_undeployed_assets files needs_brand_head_assets → asset-deployer →
derive_brand_head_assets (derives favicon + OG card FROM the approved logo; 11
complete fleet-wide). Lendzy's gap is UPSTREAM: no logo asset, sites.logo_url empty
— nothing to derive from. So seam 3 on lendzy = imagery-lane work (a logo), not a
platform seam. Open residual: whether discovery sweeps a shadow site at all —
check before relying on auto-filing.

### 2026-08-02 late night (cont. 2) — both verdicts APPROVED round 1; objections settled

Seam 1 corr 56ab6e23: APPROVED, 5 advisory objections. Settlements:
- editquality "is_active/forked_from not shown": checked live — e6347680 is_active=t,
  forked_from NULL, component_level=site, function=site-footer; 14/14 footer slots
  point at it by component_id (the render path is the slot JOIN, not library
  selection, so the selection landmine doesn't bite here).
- editquality "site_id unverified": sites row for 8ff093d5 reads name=lendzy.co.uk
  (queried before seeding).
- bug_historian "full-literal UPDATE may revert bugs_closed/111's {{if or .email
  .phone}}": the new literal was built by python byte-exact old.replace(anchor,
  block) on the md5-verified live bytes — the wrapper is provably untouched (it is
  in both test constants).
- bug_historian "wrong-type → silent whole-template fallback, first operator-config
  exposure; is the fallback loud?": answered — logger.Warn ONLY, no error row.
  FIXED beyond the objection: declared-type guard at the gap-fill
  (resolvedValueSatisfiesDeclaredType, commit 2046b6975) refuses a non-array on an
  array/list-declared field → block renders ABSENT + named Warn instead of the
  whole slot degrading. Guard scoped by measurement: 53 array/list fields ranged,
  16 unreferenced, 0 bare-output → cannot break a working render.

Seam 2 corr 4cffcebb: APPROVED, 4 advisory objections + 1 low note. Settlements:
- editquality "skip logic only in a comment": implementation is attribute-based
  Contains(rel="canonical") both quote styles + a reversed-attribute-order TEST.
- bug_historian "sibling file may duplicate the head-splice": settled decisively —
  rerender_pages_actions.go's rerenderSinglePage is wrapped by
  RerenderSitePagesAction which has NO registry entry and NO references outside its
  own file: unreachable dead code. compile_page_sections (page-content-writer) IS
  live and renders heads via RenderHead, but its output is split into sections by
  save_page_sections and the shipped artefact is assembled by assemblePage — the
  canonical lands at the last writer.
- bug_historian "other writers of the stored blank-tag shape": irrelevant by
  design — the splice runs at ASSEMBLY, downstream of every stored-head writer.
- guardian "changes head bytes fleet-wide": intended; inert until roll; pod-grep +
  census owed (already recorded).
- guardian "legacy fallback wart deserves a named TODO": recorded here + documented
  in the test as a deliberate non-assertion. If it bites, the fix is to delete the
  fallback and measure who screams — do not extend it.
- prior_art_librarian "grep absence claim / ugrep silent-zero landmine": the sweep
  returned 10+ (unrelated) matches, so the tool was demonstrably not in its
  silent-zero failure mode; the absence claim is about rel= emission among hits.

Register discipline note: both commits landed with Council-Submitted; both corrs
now APPROVED and read — 098 credits automatically, no amends (forward-only).

### 2026-08-02 ~23:00Z — seam 1 VERIFIED COMPLETE on the artefact

Strict census at gqls/sites origin/master b1d7c98ad, scripts stripped first:
**18/18 pages carry line 1 (does-not-lend), 18/18 line 2 (FCA independence),
18/18 the footer-compliance block.** Zero missing. The three tail pages
(breathing-space, credit-unions, report-loan-shark) drained through their own
page_rerender items (filed by rerender-pages itself at 22:36:16Z) — the parent
item reading `complete` at 22:36 while children queued is the known
a-complete-work-item-is-not-a-repaired-artefact shape; only the artefact census
closes the claim. Scorecard metric moves 3/15 → 18/18. Seam 1 CLOSED as built:
carrier live, seeded, propagated, measured.

### 2026-08-03 morning — v1.0.1238 rolled by owner: seam 2 VERIFIED LIVE; seam 4 reframed by evidence; 090 filed

Roll verification done at the POD (rule: a roll is not evidence): both replicas of
v1.0.1238 carry injectCanonicalLink (4 strings), resolvedValueSatisfiesDeclaredType
(1), the "no canonical emitted" log literal (1); the zero-control returned 0 (a Go
comment never reaches a binary, so it doubles as the nonsense-string control — the
POSITIVES are what prove the grep pipeline). Then one assemble-mode rerender of
about.html (item d1319b89, fired 7 min after pod start): artefact now carries
`<link rel="canonical" href="https://lendzy.co.uk/about.html">` and the blank
description tag is GONE (count 0; page has empty meta_description — correct-or-
absent exercised). **Seam 2 CLOSED end to end.** Remaining lendzy pages gain
canonicals at their next natural rerender — no sweep fired, deliberate; one
needs_rerender item (proven shape) sweeps them if wanted sooner.

Seam 4 REFRAMED by the retained runs (all 3 tool orchestrations still in
orchestration_states, 19:31/19:34/19:38Z yesterday, COMPLETED, no __step_error):
the handler's deploy DID run — deploy_result shows a commit at 19:33:51Z — but it
published the tool's SELF-CONTAINED document (own <head>, meta description blank,
NO site chrome). The hand-fired 19:42:43Z items were compensating for MISSING
ASSEMBLY, not a missing dispatch. Current artefact is fine (assembled by the
19:44 hand-fired rerenders, compliance block added by the 22:36 seam-1 sweep; no
canonical yet = last rerender predates the roll, consistent). So the seam is
"tool deploy path bypasses assembly", not "no rerender queued" — the handoff's
seam-4 line was a correct observation with a wrong mechanism guess. Durable
cross-cutting claim ⇒ 090 FILED per the CLAUDE.md norm:
**RUN_CORRELATION_ID f2252404-257b-49a1-bf3d-6de8b5b294b0** — read the verdict
BEFORE designing the fix (find artifacts by that corr, not the intake id).

### 2026-08-03 late morning — seam 4 REFUTED: the tool handler was never broken

The 090 run came back **UNVERIFIABLE (scope-not-narrowing)** with a precise
"still needed" list; completing that list first-hand (declared substitute per the
07-31 ruling) REFUTED the seam entirely:
1. The handler's own deploy_result (orch dcce231c, FULL 28,878-char HTML — not the
   2KB preview) contains header + nav + site-footer, one <html>: **assembled**.
2. The artefact history has exactly TWO commits for the tool page: 626c8e15d at
   19:33:50Z (the handler's own deploy, chrome verified in the blob) and 12dd4a26f
   at 22:38:59Z (seam-1 sweep). **The 19:42Z hand-fired items produced NO commit** —
   they had nothing to do; their completion was misread as having done the work.
3. No code read needed once 1+2 landed: the handler calls the page renderer, the
   renderer assembles — as designed.

So: TWO stacked wrong calls (both in WRONG_CALLS.md) — the 08-02 session's
"queues NO rerender" and this session's "bypasses assembly" (read off a truncated
preview; the SAME truncation shape stopped the 090 loop). A REFUTED backlog item
is a success: the seam list shrinks by one and the tool pipeline needs nothing.
Backlog after this: seam 5 (dead links to planned-but-unbuilt pages) is next;
then seam 6 (planner shape, owner-call component); seam 7 blocked on serving.

### 2026-08-03 midday — seam 5 CLOSED by census + prior art: zero dead links remain

Census on the current built artefacts (git show from origin/master, all 18 pages,
root-relative double-quoted hrefs checked against the actual file set): **0 distinct
dead internal targets.** The 24 dead links of attempt 1 were measured on chassis
1228; the 097 lane's content_data link resolver went LIVE in 1229, the 3 tools were
built, the blog rows were removed, and the 22:36Z seam-1 sweep re-rendered every
page on 1231 — the class fix belongs to the 097 lane (CLOSED+LIVE, both arms
induced, per its own records), and lendzy's instance is confirmed clean. Scope
notes: the census does not cover JS-built anchors (the 180 landmine — that residue
is the 097/180 lane's) or non-rooted hrefs; for "does the artefact ship 404 links"
the file-set denominator is the right one.

**Backlog position after today:** seams 3, 4, 5 all dissolved under measurement
(imagery input missing / never broken / fixed by another lane). Remaining REAL
items: seam 6 (planner imposes the standard shape past the mission — needs an
owner call on the 015 retype arm before or alongside the design) and seam 7
(in-browser fixtures, blocked on serving). The two genuine platform gaps the
scorecard found (seams 1+2) are shipped, approved, and verified live.

### 2026-08-05 — copy provenance CONFIRMED hand-written; style trials drafted; blog ruling ratified

Owner flagged loanandmortgagecalculator.co.uk's hero as LLM-styled and asked, before
any fix, whether it went through the framework. **It did not — confirmed three
ways:** sites-repo history (2 commits, both authored cqls directly, "new combined
loan + mortgage site" 07-31, no Rerender: pattern); platform shape (41 pages ×
exactly 1 `ported-page` whole-document component = adoption import); and specs (a
manual content_direction exists but NO writer ever ran to consume it). The copy is
an unframed CLI LLM's freehand — the owner's hypothesis holds. Four read-aloud
rewrites drafted against the house v3 style prompt
(travelling_docs/pitch_pdf_source/REVERSE_ENGINEERED_STYLE_PROMPT_v3.md — its rule
3 IS the flagged tic): COPY_STYLE_TRIALS_2026-08-05_hero_rewrites.md. Next steps in
that doc; live writer prompt content [UNVERIFIED — read before designing].

**Blog ruling ratified by owner 2026-08-04 ("please go ahead"), with his
clarification recorded:** blog ≠ news, but editorial feature-building FROM news is
also blogging. The ruling shape: mission vocabulary gains an explicit blog
declaration (none / editorial-from-news / curated-features); the planner must never
plan the unbuildable section-index blog shape and never plan any blog undeclared;
the 015 retype arm is the per-site migration tool for editorial-from-news ONLY,
never an automatic repair. Condition before any regulated site takes
editorial-from-news: verify the news writers read content_direction [UNVERIFIED].

**lendzy DNS (owner changed NS):** Nominet HAS it (whois + parent delegation both
show ivan/betty.ns.cloudflare.com); Cloudflare answers authoritatively; public
resolvers serve the proxied A records. HTTP TIMES OUT — zone resolves but nothing
serves (no origin wired). ⚠ Before it ever serves: strip the acceptance_marker
from content_direction + rebuild about.html (marker copy is IN the live artefact),
logo, register row.

### 2026-08-05 — voice chosen (H), prompt built, SEEDED; model stays gemini; rebuild queued

Owner chose trial H after a round-2 refinement (walk-in conditional openings —
candidate rule 14). VOICE_gentle_explanatory_v1.md written: 10 rules + anchor +
the 4 approved before/after pairs seeded as worked examples (models imitate
exemplars more reliably than rules). Model premise CORRECTED from llm_call_log:
the writer runs gemini-pro-latest since 07-27 (Sonnet 4.6 before, NEVER Sonnet 5);
owner ruled: leave it on gemini. **Finding during the seed: the site's manual
content_direction (07-31) ENCODED the rejected style** — banned contractions,
em-dash definitions, conclusion-first openings — so the seed was a REVISION, not
an append: 4 rules replaced (string-guarded in both keys AND formatted — every old
string matched exactly once), 4 added, voice_exemplars key + formatted exemplar
section. Applied as a supersede (07-31 rows preserved); read back verified
(19 rules, 22,980-char formatted, current row 2026-08-05). The spec is INERT until
writers run — all 41 pages are rebuild_policy='owned' ported wholes; the
whole-site rebuild (owner: "do the whole site — I'll check it then") is the
handoff's task with hazards named there (owned-page route, decompose first,
23 calculator pages = TL-001 clobber class, no --fidelity high).

### 2026-08-12 — fleet build-out scoping: register-vs-DB gap, pipeline capability census, and the oufe.com graph double-check

**[MEASURED, direct DB query]** Of the 152 domains in `PORTFOLIO_domains.txt`,
exactly 4 have a `sites` row: `loanandmortgagecalculator.co.uk` (41 pages),
`mortgagecalculator.co.uk` (31), `loancalculator.co.uk` (27), `loancash.co.uk` (22).
All `status='deployed'`. The other 148 have no row at all. Cross-checked against the
register's own status legend: only 4 of 43 propositions (M1, M2, L1, L10) carry any
status tag — the register never claimed more than this, so "148 greenfield" is
consistent with the register's own text, not a discovery that contradicts it.

**[MEASURED, adoption-lane NOTES]** All 4 built sites are hand-written or adopted
(byte-preserving crawl) from pre-existing live sites — none was produced by the
generative pipeline end to end. The one real pipeline test is `lendzy.co.uk`,
deliberately kept outside the portfolio/register precisely so the experiment
couldn't be mistaken for a launch; it has no Cloudflare zone (not publicly
reachable) and is currently 22 pages, up from "18 of 20 planned" on 2026-08-02.

**[MEASURED, code read]** `platform/orchestration/actions/v3_site_actions.go:3226`,
`validate_site_plan`: `maxPages := 20` — a ceiling the validator truncates DOWN to
(preserving already-realised pages first), not a floor. Real runs land at 15-18.
Config-only fix: `agent_definitions.default_config` for `build-site-planner`,
`workflow.steps.validate_plan.config.max_pages`. Not exposed as a submit-time flag.

**[MEASURED, code + doc read]** Capability census against the "tools, guides,
infographics, graphs, newsfeeds, directory listings" ask:
- Tools: pipeline genuinely generates working calculators (`tool-generator` ->
  `create_tool_component_action.go`); arithmetic checked correct once against known
  constants (0.8%/day, GBP 15 cap, DISP 1.6.2R) but no automated correctness gate
  exists; defaults to parking in `needs_human_review` on first contact with a new
  site.
- Newsfeeds: fully wired, self-healing (`seed_content_sources_action.go` ->
  `content-feed-orchestrator` -> `feed-ingester` -> `content_feed_items`, plus a 6h
  heartbeat task and discovery-check backstop).
- Directory listings from live web search: the verified-claim mechanism
  (`directory_claims.go`, quote-must-be-present-in-fetched-source) is real and
  proven, but wired to exactly two unrelated verticals (an AI-model directory, a
  company adoption tracker) — zero wiring for finance/insurance. The other directory
  mechanism (`entity-directory` page type / `directory-build-handler`) is a dead
  end for this: `bugs_open/206` establishes it has never rendered on any live page
  fleet-wide.
- Graphs/infographics: `evidence-chart` (CSS bar chart, gated on a site's own cited
  `evidence_base`) is the only real capability; a general infographic/multi-chart
  generator is explicitly unbuilt, on the platform's own roadmap as future work
  (`PLAN_imagery_loop_closure.md`).

**[MEASURED, direct DB query, 2026-08-12]** Double-checked the graphs claim against
`oufe.com` specifically (asked by the owner, not assumed): `/cases/thames-water.html`
does render `evidence-chart` (confirmed via its CSS class in stored
`rendered_html`). Its three tool pages have no chart/canvas/graph markup at all —
an `<svg>` hit on `/tools/tool-recovery-waterfall.html` turned out to be a decorative
icon, not a chart, on inspection of the actual markup (initial ILIKE '%svg%' screen
was a false positive for "graph"; had to read the actual bytes to tell). Also found
a second, genuinely separate chart mechanism, `report_charts.go`
(`renderBarChartSVG`/`renderHeadroomChart`), but confirmed via its sole caller
(`create_report_page_action.go`) that it's hardcoded to `page_type='report'` pages
driven by `score_grippers` — a physics-scoring mechanism built for the unrelated
robot-hands/gripper-dossier vertical. `oufe.com` doesn't use it (its case-study page
is a plain content page). Net: the "evidence-chart only" claim holds; no counter-
evidence found. Side finding, not part of this workstream: oufe's own "recovery
waterfall" tool doesn't visually render a waterfall despite the name — flagged to
the owner as a possible future fix on that site, out of scope here.

**Enforcement-gap re-confirmation [MEASURED, 2026-08-12]:** structural-validity gate
still has zero generalized implementation anywhere in `platform/`; bug 161's
fleet-wide fix (72% of registered facts are prose-sourced and unverifiable by
construction, figure unchanged since 2026-07-31) is still parked at "architecture
review, RFC before code," unstarted; fidelity dial still only has `locked` wired.
No activity on any of the three in the 24h preceding this session's start.

Full phased plan from this session: `PLAN_2026-08-12_fleet_buildout.md`.

> **CORRECTED 2026-08-12, same session, before acting on it.** Phase A4 of that plan ("raise
> `max_pages` from 20 to ~24-25") is moot. Direct query of the LIVE `agent_definitions` row
> for `build-site-planner` shows `validate_plan.config.max_pages = 80`, not 20 — some other
> session already raised it (`docs/agent_docs/sql_for_agents/053_build_site_planner.sql`
> carries the comment "raise validate_plan max_pages 20 -> 80 (don't truncate; 80-page
> ceiling)"), unrelated to this build-out. The earlier research agent's `maxPages := 20` code
> citation (`v3_site_actions.go:3226`) was accurate as a fallback default, but the live
> config already overrides it. Caught by checking the live row myself before writing to it,
> not by trusting the secondhand research figure — exactly the kind of drift CLAUDE.md warns
> a session-start snapshot goes stale within minutes. No action needed for A4.

**Phase A5 check [MEASURED, 2026-08-12]:** bug 251 (canonical) is FIXED, committed
(`61abbdbd0`, 2026-08-11 16:50) **and confirmed LIVE** — `git merge-base --is-ancestor
61abbdbd0 ef1374426` is true, `ef1374426` is the only `makefile` (IMAGE_TAG) commit in the
window and bumped to `v1.0.1291`, and both running `agent-chassis` pods are on
`v1.0.1291` (started 2026-08-12T14:55Z). Bug 252 (og:/lang) is NOT yet implemented — only
the locale-mechanism decision has landed (`f666408ed`, "option 3"); no commit touches
`rerender_single_page_action.go` for og:/lang since 251's fix. Owned and actively worked by
`loanandmortgagecalculator_couk` (their own handoff: "`bugs_open/251` → then `252`. Order
matters"), not something to compete with. **Consequence for this plan**: `canonical_mismatch`
auto-repair (A1) can be enabled once bug 251 is confirmed live — it now is. Bug 252 is not a
hard gate on Phase C per the plan's own text (informational/pacing only) but re-check its
status nearer the pilot, since a fresh-built site with dropped `og:` tags on 503+ pages is
exactly the kind of defect this build-out shouldn't multiply unnecessarily.

**Phase A1 shipped [2026-08-12]:** `check_site_structural_validity.go`, four checks
(`dead_internal_link_live`, `canonical_mismatch`, `structured_data_invalid`,
`head_essentials_missing`), committed `c66a83e9e`. Flag-only, self-clearing, not yet
enabled on any live discovery agent (deliberate follow-up). Council-submitted, corr
`51cb66fb-e4bc-46ec-8bbd-a4a561da14a0` — **committed before the verdict landed and the
commit carries no `Council-Submitted:` trailer**, because the submission went out after
the commit; check the verdict by correlation, not by the commit message, before writing
`Council-Reviewed:` anywhere (forward-only forbids amending the original commit to add
the trailer now). Built by a background agent working in an isolated worktree — its
finding along the way: **bug 251 is already fixed on this branch** (found via
`preferred_page_url_test.go`/`preferredPageURL` existing), so `canonical_mismatch` is a
regression guard, not groundwork for an open bug.

**LANDMINE, self-inflicted, worth recording so nobody repeats it:** copying the agent's
files into the main tree, then running `git stash -u` to isolate a test run, swept up
EVERY other concurrent session's uncommitted WIP in this already-dirty shared tree
(confirmed: `revalidate_unverified_claims.go`, `store_generated_component_action.go`,
and more). A concurrent session then continued editing `revalidate_unverified_claims.go`
WHILE it sat in my stash, so `git stash pop` correctly refused (local changes would be
overwritten) rather than silently reverting their newer edit — git's own safety net held,
but only because it refused rather than merged. Recovered by `git checkout stash@{0} --
<untracked path>` (fails on untracked paths — use the saved patch file / re-copy from
source instead) for my own two files and re-applying my saved `.patch` for the one edited
file, **never touching the stash's copy of the other session's file** since a newer live
version already superseded it. The stash (`stash@{0}`) is left in place, not dropped —
its content is now redundant with the live tree, and dropping someone else's swept-up WIP
is not this thread's call to make even when redundant. **Do not `git stash -u` on this
tree without a narrow pathspec** (`git stash push -u -- <your paths>`) — the whole-tree
form assumes a clean tree that this repo, by its own design (CLAUDE.md's opening
paragraph), never has.

**Phase A2 (RFC_025) shipped [2026-08-12]:** both stages committed `3129cceea`
(`refresh_evidence_base_action.go` + new `refresh_evidence_base_rfc025_test.go`, 9 new
tests incl. the `gd-trials`-shaped induced-fault canary and three RFC_017 fail-closed
cases). `datahelpers/claims.go` confirmed byte-for-byte unchanged (empty `git diff --stat`)
— the RFC's ratified §9 Q2 constraint holds. Council-submitted, corr
`9fd94852-ff79-496b-96b5-78a8d3619162` (again committed before submission — same pattern as
A1's `51cb66fb...`, no trailer on either commit; both need checking by correlation, not
by commit message, before either gets a `Council-Reviewed:` trailer anywhere). RFC_025
itself stays short of status `IMPLEMENTED`: the mechanism is proven but no real fact has
been retyped to use `artifact_check` yet, and the submission flagged an unresolved scope
gap (`artifact_check.component_id` isn't verified to belong to the fact's own site) for the
reviewers to rule on.

**Phase A is now code-complete, both pieces council-submitted, verdicts pending:**

| item | commit | council corr | status |
|---|---|---|---|
| A1 structural-validity gate | `c66a83e9e` | `51cb66fb-e4bc-46ec-8bbd-a4a561da14a0` | submitted, verdict pending |
| A2 / RFC_025 (bug 161 fix) | `3129cceea` | `9fd94852-ff79-496b-96b5-78a8d3619162` | submitted, verdict pending |
| A3 fidelity dial | -- | -- | no action, already ruled acceptable |
| A4 max_pages | -- | -- | moot, already 80 live |
| A5 bugs 251/252 | -- | -- | 251 live-confirmed, 252 owned elsewhere |

**Next session should check both verdicts before treating Phase A as done**:
```sql
SELECT correlation_id, created_at, metadata->>'decision' FROM diagnosis_artifacts
WHERE correlation_id IN ('51cb66fb-e4bc-46ec-8bbd-a4a561da14a0','9fd94852-ff79-496b-96b5-78a8d3619162')
  AND kind='council_report' ORDER BY created_at;
```
then move to Phase B (finance/insurance directory producer) per
`PLAN_2026-08-12_fleet_buildout.md`.

**Both verdicts came back REVISE, 2026-08-12** — real, substantive findings, not
process nitpicks. "Revise, don't defend" applied:

- **A2/RFC_025 (corr `9fd94852...`)**: fixed and resubmitted (`RESUBMIT_CORR`, same corr,
  trail preserved), commit `9652f4d52`. Three real defects closed: (1) [HIGH] the
  `artifact_check` regex had no anchoring — a bare `10000` pattern would substring-match
  `100000`, reproducing bugs_open/161's own documented landmine INSIDE the fix meant to
  close it; refused at parse time now. (2) [MEDIUM, 4 reviewers independently] `component_id`
  wasn't scoped to the fact's own site — a fact could "verify" against another site's
  component; now joined through `pages` and site-scoped, fails closed on cross-site refs.
  (3) [MEDIUM, architecture seat] the `changed`/stale_evidence-raise decoupling from the
  first round touched the PRE-EXISTING citation branch, outside RFC_025's ratified scope —
  reverted for citation/sql (exact prior behaviour restored), the new capability scoped to
  a new `ArtifactCheckDrifted` counter via an extracted, unit-tested `shouldRaiseStaleEvidence`
  predicate. One HIGH objection (guardian: does write-back silently delete untyped keys via
  the typed struct?) was a **false alarm** — verified directly by reading
  `writeRefreshedEvidenceBase`, which marshals the untyped map; the relevant LANDMINES.md
  entry itself names this function as one of the two safe writers. Cited as evidence, not
  argued from memory.
- **A1 (structural-validity gate, corr `51cb66fb...`)**: revision dispatched to a background
  agent (real findings: sitemap coverage promised in the rationale but not built — now being
  added as a fifth narrow check; a genuine but not-fully-argued overlap with
  `check_missing_structure.go` — that one checks the DB, this one checks the LIVE served
  page, needs the distinction written up; one objection (claimed `preferredPageURL` doesn't
  exist) already personally verified FALSE — it's real, at
  `rerender_single_page_action.go:1050`, test-pinned).

### 2026-08-13/14 — Phase B execution (see HANDOFF_2026-08-13_continue_here.md for the plan)

**B1+B2 committed** `6f26570e4` (3 finance kinds in all profile tables, kind-scoped HITL
keys, `evaluate_directory_features`, DIR-001 register entry + index row). Council corr
`69a619e6-5152-45d8-ae01-5d30a0c7776f`. **Round 1: REVISE** (guardian gating + compliance).
Split three ways: (1) real defects fixed in `035f72365` — "verified facts" headline
overclaim → "cited facts", and the non-price ruling got its MECHANICAL half
(`financeKindFieldAllowlist`, closed per-kind field vocabulary enforced at registration,
price-shaped fields structurally unregistrable, refusals to the HITL queue, tested);
(2) submission under-naming (registry.go, coverage tests WERE edited in round 1 — the
edit list just didn't name them; round 2's does); (3) the guardian's "changes LIVE kinds'
key shape" — TRUE and owned rather than narrowed, with the reader census as evidence
(grep: nothing programmatic matches the bare keys; half-migrating the key space would
recreate the bugs_open/213 collision). Round 2 resubmitted 2026-08-14, verdict pending.

**B1 implementation note**: a background agent began B1 and was killed mid-B1b by an API
session limit; its partial work was extracted from its worktree, completed and reviewed
line-by-line in the main session. Two pre-existing bug-231 image-test failures in the full
actions suite belong to another session's in-flight WIP (verified passing at clean HEAD in
a throwaway worktree).

**B3b APPLIED [2026-08-14]**: `finance-directory-researcher` agent seeded (active; all its
actions ship in the OLD binary, so the seed is safe pre-roll) + three weekly discovery
tasks in `finance_directory_pipeline/SEED_*.sql`, **enabled=false on purpose** — the
non-price allowlist only exists in the Phase B binary, and research on the old binary
would enforce the compliance ruling by prompt alone (the exact gap the council objected
to). Enable-gate: pod-grep `"per kind (Phase B kind-scoped keys)"` on BOTH replicas, then
`UPDATE scheduled_tasks SET enabled=true WHERE name IN (...)`. Committed with the seeds.

**Still owed for Phase B**: round-2 verdict; B3a components seed (background agent
authoring now); the Phase B image roll; B3c publish-trigger fix; B3d wiring migrations;
B3e planner rule; B3f enablement; legacy constant-key row cleanup (disciplined shape
recorded in the round-2 submission's grounded_in); B4 first runs + HITL review.

**A1 revision complete and resubmitted [2026-08-12]**: fifth check `sitemap_entry_dead_live`
built (reuses `probeInternalLinkTargets` verbatim), header sections added distinguishing
`head_essentials_missing` from `check_missing_structure.go` (DB vs live) and
`dead_internal_link_live` from five existing DB-only link mechanisms. Commit `9e410ce85`.
Resubmitted with `RESUBMIT_CORR=51cb66fb...`; verdict pending as of this note.

**A2/RFC_025: APPROVED at round 2** (corr `9fd94852...`, 2026-08-12 20:42). Full detail in
the RFC file itself (§10). **Phase A's blocking gate is therefore satisfied on the RFC_025
side** — only A1's resubmitted verdict remains outstanding before Phase A can be called
fully done.

**A1 round 2: REVISE again, but much lighter — 10 of 12 reviewers approved outright**
(2026-08-12 20:53). The one real, gating finding (editquality + prior_art_librarian,
independently, same underlying issue): a header comment added in round 2 claimed
`check_missing_structure.go`'s DB predicate was a meaningful discriminator — this collided
with a standing landmine (`pages.rendered_header/rendered_footer/rendered_head` are
VESTIGIAL fleet-wide). **Verified live, not just cited**: `SELECT count(*), count(*)
FILTER (WHERE length(rendered_header)>0), count(*) FILTER (WHERE length(rendered_footer)>0)
FROM pages` → 683/0/0 — the landmine holds, ten days on and 121 more pages later. Corrected
the comment in place (commit `a86dd0349`) — the honest differentiation is actually
*stronger* than the retracted one: a DB predicate that fires identically on every page
provides no signal to overlap with at all. Also independently verified (not just re-cited)
reuse_agent's two LOW notes: `check_phantom_internal_links.go`'s `phantom_internal_link`/
`unbuilt_internal_link` item types are real, and `loadStructuralDomain` has exactly one
caller inside `loadStructuralPopulation`, no duplicate loader. Resubmitted (round 3), queue
was clear, verdict pending.

**A1: APPROVED at round 3** (2026-08-13 10:26, corr `51cb66fb-e4bc-46ec-8bbd-a4a561da14a0`).
Neither implementation commit carries a trailer (all predate their submissions) — this
note and the parallel one in the RFC/commit trail are the durable record; resolve by
correlation, never by commit message.

**PHASE A IS NOW FULLY DONE — both blocking pieces implemented AND council-approved.**
Per the owner's sequencing ruling and `PLAN_2026-08-12_fleet_buildout.md`, Phase B
(finance/insurance directory producer) is unblocked. Note the checks are approved but
NOT yet enabled on any live discovery agent, and the RFC_025 code has not yet rolled —
enablement migration + a chassis roll are Phase-B-adjacent follow-ups, not blockers on
starting B's own build.

**PHASE A LIVE [MEASURED 2026-08-13 ~16:45Z]:** chassis `v1.0.1295` (rolled ~13:53Z by the
D2/observability lane) carries all six Phase A commits — verified at the artefact per the
standing recipe: stamp `69612d692` probed in `/proc/1/exe` on BOTH replicas (control sha
absent), each commit confirmed via `git merge-base --is-ancestor`. The log-line half of
the recipe was unavailable (startup line rotated out — the TIME-LIMITED landmine another
session filed today as `6ceeaba1b`). Owner's Phase A gate is fully satisfied. Phase B
approved and planned; cold-start = `HANDOFF_2026-08-13_continue_here.md`.

**Side finding while double-checking the round-2 correction (owner asked for one more
check): bug 270 filed.** `check_missing_structure.go` — the very check the council
objections forced a comparison against — turned out to be LIVE and firing on a predicate
that cannot be false (vestigial columns, empty on all 683 pages): 43 `needs_rerender`
items since April, ~25 completed full-site rerenders dispatched for nothing,
`dartsonline.com` 3-for-3, fired again 2026-08-12. Full census + fix candidates in
`bugs_open/270_HANDOFF_2026-08-13_missing_structure_check_fires_on_vestigial_columns...md`.
Unowned; candidate 1 (retype predicate onto `site_components`) is the door-closer.
Also a standing confounder for any rerender-churn investigation (bugfix 117's class).

**B1 round 3: REVISE again [2026-08-14 18:17]** — the loop is now form, not substance:
the gating HIGH is the submission schema's own ≤8-edit cap vs 10 edited files (named
inside an edit's rationale instead; reviewer flags the array anyway — structurally
unsatisfiable as asked); bug_historian's medium MISREADS group-by-kind as "caps emission
to one item per kind" (the item aggregates the kind's failures in spec.rejected, refreshed
per pass — same as the pre-existing pattern, now per kind); reuse_agent flipped
approve→object on a point answered in round 2. Two cheap checks run anyway and clean:
`agent_definitions` references to `evaluate_directory_features` = 0, to
`mortgage_lender_directory` = 0 (prior_art's lows). Gate is advisory; verdict trail
recorded; 4th-round-or-proceed is the owner's call. Nothing else blocks on it — the roll
is the real gate for everything remaining. Milestone summary:
`SUMMARY_2026-08-15_guardrails_live_directories_built.md`.

### 2026-08-15 — Phase B ACTIVATED on v1.0.1301; round 4 at owner direction; bug 270 owned elsewhere

Owner rulings: bug 270 is another thread's (hands off); run council round 4. Fresh roll
v1.0.1301 verified on BOTH replicas via the gate literal + absent-sha control — carries
Phase A AND Phase B. Executed in the recorded order: components seed regenerated with the
honesty clause (compliance r3 low) and APPLIED (6 section-level rows; ROLLBACK-dry-run
validated first); tasks enabled with `last_triggered_at=now()` (first fire deferred —
supervised force-trigger is the next session's B4); legacy constant-key rows pre-counted
(exactly 2, ids 39b5153f/35350447) and cancelled with a successor-key note. Round 4
submitted: states the 8-edit cap as fact instead of claiming completeness, answers
bug_historian's aggregation misread with the mechanism, shows the full allowlist, records
the activation, and carries the two zero-count agent_definitions checks. Cold-start:
`HANDOFF_2026-08-15_continue_here.md` (supersedes 08-13). Owner decision list delivered in
chat and recorded in the handoff §3.

### 2026-08-15 (later session) — B4 supervised first runs begin

**Cross-thread pointer (owner, in the session brief): copy-voice work is live in another
thread** — session "copy quality two stage", id `79d969f9-0009-4540-84cc-2557222db288`.
Relevant to handoff §3a (mortgagecalculator.co.uk voice review): that review should not be
duplicated from this lane; treat that session as the active worker on voice/copy.

**Observed on pickup, benign but unexplained writer**: all three finance discovery rows
carry `last_completed_at=2026-08-15 13:54:40.636839+00` — identical to the microsecond
(one statement), while `last_triggered_at` still holds the 10:51:40 activation stamp and
`updated_at` still reads 08-14 08:12. Not the scheduler (`stampCompleted` writes both
stamps), not the admin handlers (both bump `updated_at`); so a manual psql UPDATE, most
plausibly clearing the in-flight/timed-out cosmetic state the activation left
(`last_triggered_at` set, `last_completed_at` NULL). No orchestrations of either
researcher type existed before B4 fired, so no run was hidden behind it. [INFERRED as to
motive; the column evidence is measured.]

**B4 run 1 — mortgage-lender, fired 14:3x UTC** via
`UPDATE scheduled_tasks SET last_triggered_at=NULL WHERE name='mortgage-lender-directory-discovery'`.
Pre-flight: chassis pods 179m old (no 300s window issue); open-work-item check clean for
the three finance kinds (the open `directory_citation_unverified` rows are the MODEL lane's
kinds — protocol/company/model — another workstream). Orchestration
`c516508b-a1ae-43ad-8f39-63aa07f48b8f` spawned within a tick, step `search_web`,
AWAITING_RESPONSES. Watching to terminal state; results below.

**Run 1 result (COMPLETED 14:32:51, ~2.5 min): mechanics ALL work, the SET fails the bar.**
Chain exercised end-to-end: search → prepare (4 urls of 10 results) → scrape → extract →
verify_and_register → kind-scoped HITL reject item. 3 entities registered, 8 claims all
status `found`; 1 reject item `directory_citation_unverified` ("Directory
(mortgage-lender)", 2 candidates, both `fetch_error` HTTP 403 on one Eversheds page). No
price-field refusal this run (nothing attempted a price field). **The defect run 1
exposed: 2 of 3 entities were CATEGORY-shaped** ("FCA-regulated mortgage lenders
(general)", "UK Specialist Lending Sector (nonbank)") — true, citable facts hung on
sector pseudo-entities. Cause read from the artefacts, both layers: (a) retrieval — the
research_query described the MARKET, so the scrape set was 3/4 market-level pages (FCA
how-to, KBRA RMBS research, Eversheds market-study commentary, BoE statistics) and only
1/4 an actual lender (Family Building Society, the run's one real entity); (b) extraction
— the prompt said "name the specific provider" but never forbade aggregates, and the
citation gate CANNOT catch this class because the defect is entity shape, not claim
truth. The sibling directory-researcher (model lane) has the same latent gap — unfired
there because model sources are model-specific; NOT edited (their row), noted here.

**Fixes applied (all config, live immediately, one transaction where possible):**
- **Migration 423** (`sql_for_agents/423_…named_firm_rule.sql` + ROLLBACK): named-firm-only
  rule appended to extract_claims after the price rule, 206's replace()-idiom, snapshot
  first (verified in `agent_definitions_backup` holding the PRE-change text,
  14:40:14Z), anchor unique, idempotent WHERE, verify t|t|t. NB found while writing it:
  a `\n` inside a LIKE pattern matches literal `n` (backslash is LIKE's escape char) —
  use position() for anchor checks over jsonb ::text.
- **research_query re-aimed at NAMED firms, all 3 kinds** (guarded UPDATEs, 1 row each).
  Old values for reversibility: mortgage-lender "UK mortgage lenders FCA authorised:
  banks, building societies and specialist lenders; residential, buy-to-let and
  later-life product ranges; regulator status and firm reference numbers" · savings
  "UK savings account providers: banks and building societies, FSCS protection, product
  types (easy access, fixed term, ISA), regulator status and firm reference numbers" ·
  health "UK private medical insurance providers: insurers and underwriters, cover types
  (inpatient, outpatient, mental health, dental), regulator status and firm reference
  numbers".
- **The 2 category entities archived** (status was the only live filter — consumers read
  `status='active'` only, `queryresolve/directory_items.go:108`), reason written into
  `attributes.archived_reason`. Family Building Society stays active.
- **HITL item `dc891e85` ruled complete/discard** with the reasoning in `result` (its two
  candidates cited the 403 page AND hung on a category-shaped entity).

**Run 2 fired ~14:47Z** (same force-trigger) to measure both fixes. Applying the two
fixes together rather than one-per-run was deliberate: they act at different stages
(retrieval vs extraction), and supervision time bounds the run count.

> **CORRECTED 2026-08-15 (same session): run 2 actually fired 14:41:11Z and FAILED at
> search_web — MY re-aimed query broke it.** Orchestration `42f72cd9`, error "search
> query not found - check 'query', 'topic', or 'query_field' config". The config was
> byte-identical to run 1's; what changed was the query VALUE — my re-aimed wording was
> 275 bytes, and `web_search_action.go` extractSearchQuery's query_from path drops any
> resolved query with `len >= 200` as a "likely LLM error message", then falls through
> to an error that misdirects at CONFIG KEYS, never mentioning length. (Run 1's original
> query was 183 bytes — under the cap by accident.) Caught in ~2 minutes because the run
> was supervised. Also: my first watcher missed run 2 entirely — I'd guessed a
> `created_at > 14:45` filter while the DB clock said 14:41; filter times belong to the
> DB's clock, not your sense of elapsed time. Queries shortened to 184-185 bytes (same
> intent, plain ASCII), live rows + seed updated, cap documented in the seed header.
> The len<200 conflation of "long query" with "LLM error message" is a platform wart —
> recorded here and in the seed; not filed as a bug this session (config-side fix
> suffices for DIR-001's fixed queries; a fleet-shaped 016b/bugs filing needs a grep of
> other query_from consumers first, which the next session can do off this note).

**Run 3 fired ~14:45Z** with the shortened queries (mortgage-lender kind again).

**Run 3 (`ffc22155`, COMPLETED 14:45:54): the named-firm rule HELD — and retrieval got
worse.** Scrape set 4/4 regulatory/market pages (FCA how-to, FCA handbook SUP16, FCA
authorisation page, the same Eversheds commentary): my shortened query's "FCA
authorisation, firm reference number" vocabulary pulled the REGULATOR's own pages. The
extractor registered NOTHING from aggregate-only sources — zero new entities, zero
claims, zero rejects — which is exactly what the 423 rule prescribes and the first
positive evidence it works (run 1's identical source-shape produced two category
entities). Lesson: regulator words belong to EXTRACTION, not retrieval — the FCA-footer
facts live on lender pages, so the query must hunt LENDERS. Query iteration 2 (mortgage
kind only until proven, then mirror): "list of UK mortgage lenders and building
societies: named member firms (Building Societies Association, UK Finance) and each
firm's mortgage range: residential, buy-to-let, later life" (183 B). Run 4 fired
~14:52Z.

> **CORRECTED (same session): run 4 fired 14:47:11Z, not ~14:52** — the same
> local-sense-vs-DB-clock error the run-2 correction above describes, made again within
> minutes of writing it down. Timestamps in these notes must be READ from
> orchestration_states/scheduled_tasks, never estimated.

**Run 4 (`26c4f5ac`, COMPLETED 14:48:23): retrieval FIXED, and two new mechanical
failure modes exposed — 8/8 candidates rejected.** Scrape set was the right shape at
last: UK Finance largest-lenders list, BSA homepage, Family BS, IBISWorld. The extractor
produced REAL named firms (Nationwide, Yorkshire BS, Coventry BS from IBISWorld; Family
BS from its own page) — 423's rule demonstrably steering. But verification rejected all 8:
- `citation_lost` ×3 (Family BS): quotes were multi-paragraph BULLET BLOCKS
  (`"...\n\n- Owner Occupier..."`); the refetch's text extraction renders layout
  differently, so a bullet-block quote fails verbatim match even though every word is on
  the page. Run 1's single-sentence quotes had passed — failure of SHAPE, not truth.
- `fetch_error` HTTP 405 ×4 (IBISWorld): the aggregator refuses the verifier's refetch,
  so any claim cited there is dead on arrival — a wasted scrape slot every time it ranks.
One claim DID register (Family BS `lender_type` "building society", superseding run 1's
hyphenated value). **Fixes: migration 424** (quote must be ONE CONTINUOUS passage of
running text; `ibisworld.com` → exclude_domains; snapshot-first, pre-checks 1|f|f,
verify t|t), seed synced. Only the evidence-named domain excluded — the CLASS
(refetch-blocked aggregators) will grow; add members as runs name them.

**FINDING — run 4's 7 rejects were SILENTLY SUPPRESSED from the HITL queue by MY OWN
14:40 ruling.** No `directory_citation_unverified` item was created or refreshed:
`writeWorkItem` (load_work_item_actions.go, two-strike rule) drops any keyed write whose
`item_key` had a terminal (`complete`/`failed`) row created <3h before — my completing
`dc891e85` at 14:40:42 made every reject write under
`directory_citation_unverified:mortgage-lender` a silent no-op until ~17:40. The rejects
survive ONLY in orchestration `collected_data` (~24h retention). [INFERRED from code +
row absence + timestamps; the branch's own log line could not be confirmed at the pod —
see next paragraph.] Class: "your own action can silence your own detector"
(memory index). Structural fix candidate: the reject emitter should set
`workItem.recurrenceExpected` (exists for exactly this; the emitter doesn't set it) —
Go change, inert until a roll, NOT applied this session. ALSO of note for the two-strike
half: after 2 terminal rows in 7 days a third write flips to status `unresolved` — a
weekly HITL queue that a human dutifully completes will hit this on week 3.

**PUZZLE, unresolved: this action's log lines never reach `kubectl logs` on either
chassis pod** — not the unconditional `verify_and_register_directory_claims: complete`,
not the HITL-write literal (the same string the deploy gate probes in `/proc/1/exe`, so
the binary carries it), not the suppression line — across runs 1/3/4, 90-minute window,
both pods checked individually, while OTHER actions-package callers
(workflow_actions.go, database_actions.go, ai_actions.go) DO print there, and the verify
step's processing_history pins execution to `sspqq`. The DB writes prove the code ran.
Consequence for practice: for THIS action, absence of its log line is not evidence it
didn't run — the code comment at directory_claims.go:449 claims that line is "the only
pod-greppable evidence", and in production it is not greppable at all. [MEASURED absence;
mechanism UNDIAGNOSED — worth a 090 run if it matters beyond this lane.]

**Structural observation for Phase C/D (not a defect):** list pages (UK Finance largest
lenders, BSA members) name many firms but state few closed-vocabulary facts about each,
so even perfect retrieval of lists yields ~0 claims from them; per-firm provider pages
carry the facts but arrive ~1-2 per search. The 6-step workflow has no second hop
(list → per-firm pages), so realistic yield is a few firms per weekly run, accumulating.
Either accept slow accumulation or add a two-hop discovery step later — owner-visible
trade-off, park for Phase C/D.

**Run 5 fired 15:00:30Z** (DB clock, read not estimated) with 423+424 both live.

**Run 5 (`d0f4bcaf`, COMPLETED 15:01:46): 5/5 candidates REGISTERED, zero rejects — the
full loop now works honestly.** A NEW real firm entered the register: Mansfield Building
Society, cited to its own retirement-mortgage page; Family BS gained a clean RIO
product_types claim. Current mortgage-lender register: 2 active entities (Family BS,
Mansfield BS), 3 current claims, every one a named firm citing its own page. Minor
quality note for the eventual publish review: Mansfield's product_types value
("Residential mortgage lending available in England, Wales and Scotland") is a coverage
statement rather than a product enumeration — true, cited, not embarrassing, but the
kind of thing the per-site manual review should eye.

**B4 status at session end (credits): mortgage-lender kind PROVEN through the whole
chain** (discover → scrape → extract named-firm-only → verify verbatim → register →
kind-scoped HITL on rejects), with 5 supervised runs' evidence. Savings-provider and
health-insurer NOT yet run — their queries carry the same named-firm pattern (applied
this session, 184 B each) but each needs its own supervised first run per the B4 recipe:
force-trigger, watch, read candidates+registration from collected_data (do NOT rely on
the HITL queue while a completed reject item is <3h old — see the suppression finding
above), review the registered set, iterate the query/prompt if the kind's sources
misbehave differently. Register accumulation is slow by design (~1-2 firms per run;
weekly cadence + manual force-triggers).

### 2026-08-15 (third session) — B4 continues: savings-provider and health-insurer supervised first runs

**Pre-flight (all measured 17:09Z, DB clock)**: chassis pods started 11:28/11:29Z (~5.7h,
no 300s window issue); zero open work items keyed to either kind; migrations 423/424
already live from the earlier session.

**Query mirror APPLIED BEFORE FIRING (iteration 2 → savings + health).** The handoff's
recipe said "run then iterate", but run 3 (`ffc22155`) had already proven the exact
failure the two pending queries carried: regulator vocabulary ("FCA authorisation, firm
reference number") pulls the regulator's pages and yields zero. The mortgage iteration-2
membership-list shape was proven by runs 4-5 and its own note said "then mirror" — so the
mirror IS the recorded plan, done before spending a run re-demonstrating a documented
failure. Guarded UPDATEs (WHERE pinned old value), 1 row each, verified 184 B / 171 B:
- savings-provider NEW: "list of UK savings account providers: named member banks and
  building societies (Building Societies Association, UK Finance) and each firm's
  accounts: easy access, fixed term, cash ISA"
- savings-provider OLD (reversibility): "named UK savings providers: individual banks and
  building societies; each firm's accounts (easy access, fixed term, cash ISA), FSCS
  protection, FCA authorisation, firm reference number"
- health-insurer NEW: "list of UK private medical insurance providers: named member
  insurers (Association of British Insurers) and each firm's cover: inpatient,
  outpatient, mental health, dental"
- health-insurer OLD: "named UK private medical insurers: individual firms and
  underwriters; each firm's cover (inpatient, outpatient, mental health, dental), FCA and
  PRA authorisation, firm reference number"

**Seed drift found and fixed**: `SEED_finance_directory_scheduled_tasks.sql` still
carried all three ITERATION-1 queries while claiming "this file now matches live" — the
earlier session's iteration-2 update reached the live mortgage row but not the seed. All
three seed queries now match live, with a dated revision note naming run 3 as the
evidence.

**Run 6 (savings-provider, first ever) fired 17:11:28Z** (DB clock, read not estimated)
via `last_triggered_at=NULL`. Watching via monitor to terminal state; results below.

**My watcher failed silently — the run did not.** The monitor I armed on run 6 polled
`SELECT id ... FROM orchestration_states` — but the column is `orchestration_id`, and the
`2>/dev/null || true` I wrapped the psql in (meant for transient kubectl failures)
swallowed the SQL error every poll, so the monitor sat in silence for its full 15-minute
timeout while the run completed in 95 seconds. Same class as the `||true` watcher lesson
already in the memory index: FOREGROUND-TEST the exact watcher query before arming it —
the arm-time test is the only thing that distinguishes "quiet" from "broken". Run 7's
watcher uses the query proven in foreground below.

**Run 6 (savings-provider, `c2cd7f55`, fired 17:11:28Z, COMPLETED 17:13:17Z, ~95s):
15/15 candidates REGISTERED, zero rejects — first-run success, no iteration needed.**
The mirrored iteration-2 query retrieved the right shape immediately: GOV.UK's
HMRC-approved ISA managers list, the BSA members list, Family BS's own savings pages,
Moneyfacts. 13 active entities, all real named firms (Nationwide, Coventry, Skipton,
Leeds, Yorkshire, Principality, Newcastle, Nottingham, Cambridge BSs; Al Rayan Bank,
Monzo Bank, NS&I, Family BS), 15 current claims, all `found`, all cited. Independent
spot-check: Nationwide FRN 106078 and Monzo FRN 730427 both confirmed present on the
cited GOV.UK page by direct fetch.

**Kind-specific structural note (contradicts the mortgage-kind observation, usefully):**
the mortgage lesson "list pages name many firms but state few facts" does NOT hold for
savings — GOV.UK's ISA-managers list states per-firm FCA references inline, so one list
page yielded 12 fca_firm_reference claims in one pass. The two-hop discovery question
(Phase C/D) is therefore kind-dependent: savings accumulates fast from official lists,
mortgage stays slow until a second hop exists.

**Quality note for the eventual per-site review**: Coventry and Skipton `product_types`
values are ISA-component enumerations from the GOV.UK list ("Cash ISA, Cash Junior ISA")
— true, cited, but narrower than the firms' actual savings ranges. Same class as
Mansfield's coverage-statement note from run 5.

**Run 7 (health-insurer, first ever, `8b6f8e12`, fired 17:28:58Z, COMPLETED 17:30:36Z,
~84s): 13/15 registered across 7 real named insurers, 2 rejects — mechanics clean, the
SET fails the review bar four ways.** Entities: Bupa, Aviva, AXA Health, VitalityHealth,
The Exeter, Freedom Health Insurance, WPA — the named-firm rule held (zero category
entities on this kind's first run; contrast mortgage run 1). The four defects, read from
the registered set:
1. **A £-amount inside an allowed field's value**: WPA cover_types "inpatient, outpatient
   (£350 included as standard); ...". The mechanical non-price gate blocks price FIELDS,
   not price CONTENT — a benefit limit is exactly the volatile figure the owner's ruling
   keeps out. Go-side residual noted below.
2. **Same-run duplicate destroyed the better claim**: two bupa.cover_types extractions
   from one page; last-write-wins at registration, and the survivor was the weaker
   benefits blurb ("24/7 remote GP access and dental...") while the superseded row held
   the real enumeration ("inpatient, outpatient, mental health..."). Read from
   directory_claims is_current=false.
3. **Marketing prose in underwriter**: bupa.underwriter = "Bupa (no shareholders;
   reinvests profits into services)" — ethos, not a firm name.
4. **forbes.com is refetch-blocked** (HTTP 403 ×2, both rejects) — the 424 class
   (ibisworld) gained the member run 4's note predicted; per that recorded policy it goes
   into exclude_domains.
Also: **source concentration** — all 12 current claims cite ONE broker comparison guide
(drewberryinsurance.co.uk). Verifiable and honest, but no GOV.UK-equivalent surfaced for
this kind (the ABI homepage scraped but yielded nothing extractable); flagged for the
per-site review rather than fixed by query over-fitting.

**Migration 428 applied 17:34Z** (`sql_for_agents/428_…value_hygiene_and_forbes.sql` +
ROLLBACK): (A) no monetary amount anywhere inside a value; (B) at most ONE claim per
(provider, field), most complete enumeration wins — because a later duplicate OVERWRITES;
(C) underwriter = the underwriting firm's name alone; (D) forbes.com → exclude_domains.
Same shape as 423/424: snapshot-first (backup verified PRE-change), anchors pre-checked
unique (1|1|1), all edits idempotent-guarded, DO/RAISE verify so a miss aborts the COMMIT
(the RFC_006 lesson — a SELECT-only verify cannot stop one). **Go-side residual, fix
candidate for a roll**: the registration-time allowlist could also scan VALUE CONTENT for
monetary amounts; today the only control on rule A is prompt text, and a doc/prompt line
is not a mechanical control.

**Run 8 (health-insurer, post-428) fired 17:35:02Z.** Success reads: forbes.com absent
from the scrape set; no £/GBP inside any newly registered value; no same-run duplicate
(entity, field) registrations; any underwriter value a bare firm name. If the Drewberry
page re-extracts, cleaner values should SUPERSEDE run 7's two offending claims; if run 8
does not touch them, curate manually (archive precedent: run 1's category entities) and
say so here.

**Run 8 (health-insurer, post-428, `297ca621`, fired 17:35:02Z, COMPLETED ~17:36:2xZ):
14/15 registered, 1 reject — ALL FOUR 428 RULES PROVEN AT THE ARTEFACT.**
- forbes.com absent from the scrape set (slot went to mytribeinsurance.co.uk);
- currency-in-value count across ALL current health claims: 0 — WPA's £350 value
  superseded by an amount-free modular description;
- no same-run (entity, field) duplicate; Bupa's real cover enumeration RESTORED as
  current (the run-7 inversion healed by supersede, is_current tells the story);
- the new underwriter claim (Saga) is a bare firm name: "Bupa".
Three new entities: National Friendly, Saga Health Insurance, General & Medical. The one
reject (aviva.cover_types, citation_lost on a longer quote) is the verbatim gate working;
aviva keeps its valid run-7 claim.

**Curation + HITL (runs done first, per the two-strike order):** bupa.underwriter
ethos-prose value retired (is_current=false, superseded_at set, no successor — better
absent than marketing prose; run 8 never re-extracted that field so supersede could not
heal it). HITL item `9acefafa` (health-insurer) ruled complete/discard with the reasoning
in result — run 8's refresh had already replaced the two moot forbes rejects with the one
aviva citation_lost. NB the 3h reject-suppression window for
`directory_citation_unverified:health-insurer` now runs until ~20:37Z; next scheduled
runs are a week out, so it expires harmlessly.

**B4 CLOSED — all three kinds proven through the whole chain under supervision.**
Register at close (active entities / current claims on them): mortgage-lender 2/3 ·
savings-provider 13/15 · health-insurer 10/15 — 25 entities, 33 claims, every one a named
firm, every claim cited, zero price content. (A count query that LEFT JOINs claims
without the entity status filter reads mortgage as 7 — the extra 4 hang on run 1's two
ARCHIVED category entities, invisible to consumers via the status='active' filter;
measure on active entities.) Run count per kind: mortgage 5 (runs 1-5, prior session),
savings 1 (run 6 — first-run clean), health 2 (runs 7-8 — 428 between them). Next: B3c
(publish-trigger fix), B3d, B3e, B3f per the handoff §2.

### 2026-08-15 (fourth session) — B3c DONE: the publish leg goes kind-aware (migration 429, applied + proven live)

**What shipped**: `sql_for_agents/429_directory_publish_trigger_kind_aware_fan_out.sql`
(+ ROLLBACK) — the trigger's find-sites query now returns one row per DUE (site, kind)
across all six kinds (per-kind opt-in via the spec key, per-kind deployed component,
per-kind publishable claims mirroring QueryDirectoryEntries: active entities,
is_current+found), `ORDER BY random() LIMIT 12` replaces `ORDER BY s.domain LIMIT 5`
(deterministic alphabetical starvation → positive probability every cycle for every due
pair), and the publisher collapsed from the hard-coded model→company→protocol 7-step
chain to ONE render→commit pair parameterised by `input_data.kind`. Per-kind commit
messages ride the trigger rows via `commit_message_field` (historical messages kept
verbatim). Seed `SEED_directory_publish_trigger.sql` synced to live in the same task
(the researcher-seed drift lesson applied proactively).

**Mechanism findings that shaped the design (all read first-hand in the Go)**:
- **A publisher-internal loop over kinds is IMPOSSIBLE config-only**: loop iterations
  suffix output fields (`_N`) and `coordinator.go prefixConfigStepReferences` rewrites
  only a dataRefKeys whitelist that does NOT include git_commit's `files_field` — a
  render→commit pair inside a loop reads a field that no longer exists. Hence fan-out
  at the TRIGGER (its spawn+call loop is the proven pattern), publisher stays linear.
- **The 07-26 silent-model-default trap is closed UPSTREAM of the action**: `kind` joined
  the publisher's `input_contract.required`, and call_agent's ValidateInputContract
  fails a call missing a required field (read at call_agent.go:1005-1013). The
  `kind!` strict marker was considered and REJECTED: pods run v1.0.1302 started
  11:28Z; RFC_029's strict-marker commit (1806371ef) landed 14:07Z the same day —
  config must not outrun the binary. [READ, not induced — the loud-fail path was
  verified in source, not by firing a broken call.]
- **`"kind": "input_data.kind"` resolves via Strategy 0** (kind is Optional in
  RenderModelDirectoryInputSpec; the closed-set literal override ignores non-profile
  strings) — pinned by the existing test case "reference that resolved" in
  TestDirectoryKindResolvesFromLiteralStepConfig.
- **Multi-iteration same-role spawn+call is iteration-aware** (findAgentByRole prefers
  the `_N`-suffixed spawn of the current iteration) — mattered because the new shape
  runs 3 iterations TODAY where the old ran 1.

**Applied 18:10Z** after snapshot_agent backups of BOTH rows (two-arg form →
agent_definitions_backup, reason '429...: pre-update'); zero in-flight orchestrations
checked first; both UPDATEs one transaction; verify DO/RAISE includes
`EXECUTE 'EXPLAIN ' || query` so a JSON-mangled SQL string aborts the COMMIT. NULL-trap
caught in my own first draft: `position(... IN trg_query) = 0` is NULL (not true) when
the path is missing, so plain `<>`/`=` verifies could never fire — rewrote with
IS DISTINCT FROM + explicit NULL guards before applying.

**First kind-aware run (force-triggered 18:10:51Z, trigger fired 18:11:12Z, COMPLETED
by ~18:13Z): PROVEN AT THE ARTEFACT**:
- 3 publisher orchestrations, one per (site, kind): company `a082f0fd`, model
  `45c0c0bf`, protocol `5442a492` — all COMPLETED.
- Per-kind entity counts DIFFER (the 07-26 kind-collapse check): company 44, model 40,
  protocol 8 — same magnitudes the 411 session measured, so the register kept growing
  while the leg was being fixed.
- Correct per-kind files (adoption-tracker/model-directory/protocol-tracker + -full)
  and per-kind commit messages carried through input_data.
- Served JSON fresh at each run's own completion second (updated_at 18:11:39 /
  18:12:17 / 18:12:42Z, all 200) — DB→render→commit→served round-trip confirmed.
- Finance kinds produced ZERO rows — correct self-gating (no site opts in yet; the
  dry-run of the query before baking it into config returned exactly the 3 AI pairs).
- `rerender_queued=0` on all three kinds — IDENTICAL pre-429 (the 17:33Z run returned
  0/0/0 too), so observed-unchanged, not introduced; noted to the model_directory lane
  in the CONTRIB, not chased from here.

**Council**: submitted FORCE=1 (411 precedent — config migration under docs/, scope
filter would refuse it uncredited). `SUBMISSION_CORR=a7c99b84-f70f-4f34-b8e9-b12813e8639e`.
Committed with `Council-Submitted:` per the standing rule; verdict to be read later.
Schema lessons for the next submitter: `.plan.summary` required; operation vocabulary
is modify|add|remove|config_change ('create' refused); `.plan.risks` is a STRING.

**Residuals, stated**: (1) the VALUES kind→spec_key→components mapping in the trigger
SQL is a LOCKSTEP contract with Go's directoryPublishProfiles — a kind added in Go
without a row here silently never publishes; LANDMINES entry added, DIR-001 updated
(adding a kind is now SEVEN data-only places, the trigger row is the seventh). (2) No
staleness ordering — random() is fair-in-expectation only; a real due-stamp needs a
bookkeeping table, deferred deliberately. Next: B3d (wire evaluate_directory_features),
B3e, B3f per the handoff §2.

### 2026-08-15 (fourth session, continued) — 429 verdict APPROVED round 1; B3d DONE (migration 432)

**429 council verdict (corr `a7c99b84`): APPROVED, round 1, 18:24Z** — "approved with 5
advisory objection(s) — none high-severity", 5 abstained. The two editquality advisories,
dispositioned with evidence rather than argued:
- "whole-workflow replacement may drop continue_on_error (the 427-class silent
  overwrite)" — the reviewer saw the abbreviated sketch; the applied file's full literal
  carries it, and VERIFIED LIVE post-verdict: `process_sites.config.continue_on_error =
  true`, `max_iterations = 12` on the live row.
- "no row-count guard against the two-active-rows fleet pattern" — the pre-flight DO
  block's first two checks are exactly-one-active-row guards for BOTH types (RAISE on
  count<>1); present in the applied file, not shown in the sketch.
Commit `0af2c21f9` carries `Council-Submitted:` and 098 credits it automatically now the
verdict is approved — no amend, forward-only. Lesson for future sketches: show the
guards and the loop config lines; both advisories were sketch-visibility artefacts.

**B3d APPLIED (migration `432_wire_evaluate_directory_features_b3d.sql` + surgical-inverse
ROLLBACK), ~18:55Z, on the fresh v1.0.1303 pods (rolled 18:45Z by another session —
config-only change, no dispatch inside the 300s window):**
- improvement-loop: `enrich_directory_features` inserted between `enrich_news_feed` and
  `load_audit_state`; news edges re-pointed on BOTH success and error paths so a news
  failure still reaches directory enrichment and either failure still reaches
  `load_audit_state` (291's property preserved). site_id = `site_record.site_id`.
- domain-research-classifier: same step between `write_classification_spec` and
  `write_content_direction_spec` — greenfield builds get the flag at plan time.
  site_id = `input_data.site_id`, error_step continues the build.
- Guards pinned the LIVE 291-shaped edges (drift → refuse); snapshots for both rows in
  agent_definitions_backup (reason '432…: pre-update'); in-transaction DO/RAISE verify
  passed. Surgical jsonb_set edits, NOT whole-workflow replacement — deliberately
  opposite to 429's approach because these two agents are other lanes' machinery and the
  ROLLBACK must not clobber unrelated later edits (it removes the step by `#-` and
  re-points edges back, guarded).
- **Finding: the improvement-loop consumer is wired but UNDRIVEN — `improvement-sweep`
  is DISABLED** (`enabled=f`, interval 900s, last fired 2026-08-14 16:34Z). The loop
  wiring self-proves only when the sweep re-enables (whoever owns that call) or a manual
  cycle runs; the classifier wiring proves on the Phase C pilot's greenfield build. The
  memory pattern "a silent mechanism is usually undriven" applies — do not read the
  absence of enrichment rows as a 432 failure.
- Council: `SUBMISSION_CORR=47785bb5-ca66-4aed-819f-2bd29277b80d` (FORCE=1, 411/429
  precedent). Committed with `Council-Submitted:`; verdict to read next session.

**v1.0.1303 note**: the roll postdates RFC_029's strict-marker commit (1806371ef,
14:07Z), so `kind!`-style strict input mappings are PROBABLY now available on the
running binary — [UNVERIFIED at the artefact; check the pod's build provenance stamp
before relying on it]. 429 deliberately avoided the marker and needs no change; a
future tightening pass could adopt it once verified.

### 2026-08-15 (fourth session, closing) — 432 round 1 REVISE, revised and resubmitted round 2

**Round 1 (corr `47785bb5`, 20:49Z): REVISE** — gating objection from editquality, echoed
by guardian and debug_historian: no version pin on the WHERE clauses (the two-active-rows
landmine). Disposition, honest on both halves:
- **Forward migration: the objection was a sketch-visibility artefact** — the applied
  file's pre-flight refuses on count<>1 active rows for BOTH types (refuse-on-ambiguity,
  deliberately safer than a max(version) pin, which silently picks a row in exactly the
  state a human should inspect). Fresh measurement for the resubmission: both types have
  exactly ONE active row (version 1); the fleet's two-row types are chief-strategist,
  content-creator, content-creator-contact, site-component-architect — neither is ours.
  Same lesson as 429's advisories, now COSTING A ROUND: **quote the guard block verbatim
  in the sketch; reviewers can only see what the sketch shows.** Third occurrence = this
  goes in the submission checklist.
- **ROLLBACK: the objection was REAL and is FIXED** — the un-run ROLLBACK lacked the
  count guards (a SELECT INTO over two active rows picks one arbitrarily while the
  UPDATE hits both). It now carries the same exactly-one-active-row RAISEs, with a
  comment crediting the round. The applied forward file is NOT edited — it is the record
  of what ran.
- **Architecture seat (medium, not gating): this is the SECOND hand-authored
  evaluate_news_feed-style splice** (news was first, directory is second); a third
  starter-kind recommender needing the same treatment is the RFC trigger for a shared
  abstraction (a generic ordered enrichment-step list on these agents). RECORDED HERE AS
  THAT TRIGGER — whoever writes the third splice should open the RFC instead of copying.

**Round 2 resubmitted** same correlation (`RESUBMIT_CORR=47785bb5…`, run orch
`5cb56ab9…`): guards quoted verbatim, measurements in grounded_in, ROLLBACK fix named.
Verdict to read next session.

### 2026-08-15 (fifth session) — 432 round 2 APPROVED; advisories dispositioned; B3e applied + live

**432 round 2: APPROVED 22:19:07Z** (corr `47785bb5`, all 11 non-abstaining seats approve,
6 abstained, `decided_by: all reviewers approve`). Commit `bbb0cfa89` carries
`Council-Submitted: 47785bb5…` and is credited automatically by 098. Five low advisories,
none gating; dispositions, each with the check run rather than asserted:

- **guardian edit-1 (output_field collision)**: MEASURED CLEAN 22:25Z — the only steps in
  either agent writing `directory_features_enrichment` are the two `enrich_directory_features`
  steps 432 itself added (jsonb_each sweep over both live workflows, 2 rows, both ours).
- **guardian edit-2 (rollback assumes its inverse targets)**: REAL, FIXED in the un-run
  `432_…_ROLLBACK.sql` — pre-flight now pins ALL six edges it overwrites or could orphan
  (news `next_step` AND `config.error_step`; the directory step's own `next_step` +
  `error_step` in BOTH agents), with refusal messages naming the third-lane-splice
  scenario. Pinned values read from the live rows 22:27Z, not from the forward file.
- **architecture (trigger must outlive lane NOTES)**: DONE —
  `architecture_review/RFC_031_hand_spliced_enrichment_steps_want_an_ordered_list.md`
  filed as the durable trigger record (population measured live: 3 splices, 2 actions;
  includes the drift evidence that news is NOT in the classifier while directory is).
  Trigger unchanged: the THIRD deterministic recommender builds the ordered list.
- **prior_art (scheduled_tasks unverifiable from the seat)**: RE-CONFIRMED 22:31Z by
  direct read — `improvement-sweep` enabled=f, interval 900s, last_triggered_at
  2026-08-14 16:34:16Z. The handoff's claim stands; the loop consumer remains
  wired-but-undriven, classifier consumer proves on the Phase C pilot.
- **debug_historian (no SELECT…FOR UPDATE)**: acknowledged, no change — matches the
  platform's existing verify-block idiom; the in-transaction verify would abort on the
  drift it describes. Noted here so the next migration author sees the trade-off named.

**B3e DONE, LIVE, VERIFIED AT THE ARTEFACT** — migration
`sql_for_agents/433_planner_directory_rule_b3e.sql` (+ surgical-inverse ROLLBACK with the
same guards), applied 22:24Z after a clean dry run (COMMIT→ROLLBACK copy: pre-flight
passed, UPDATE 1, in-txn verify passed):
- ONE `Directory rule:` paragraph after the News listing rule + RULES entry 18 after rule
  17, spliced surgically on `plan_site.config.prompt_template` via jsonb_set +
  replace() on the DECODED string (no whole-config ::text replace, no \n escape
  subtleties — a deliberate tightening of 206's idiom).
- Mapping is verbatim from `directoryCheckProfiles` (check_directory.go:72-149); the
  composition (hero → `<x>-listing` → call-to-action, header+footer nav) matches the three
  LIVE pages on ai-agent-orchestration.com AND MissingDirectoryPageCheck's own suggestion.
- Directory page types verified to survive the Go validator: normaliseRole passes unknown
  roles through; CanonicalisePage default arm preserves type, name=slug, url=/<slug>.html —
  the exact shape of the live pages. No section-index flattening trigger applies.
- Pre-flight guards: exactly-one-active-row (measured 1, version 1), anchors exactly-once
  (measured 1|1|0 pre-apply), already-applied refusal; verify block is DO/RAISE.
  Backup-row guard is PRESENCE not exactly-one — snapshot_agent runs outside the txn, so
  a dry run legitimately leaves a surviving row; exactly-one there would refuse the real
  apply after any dry run (found while designing the dry run, worth keeping as practice).
- Read back live post-apply: both splices in position, canonical table untouched, rule 17
  and news rule intact at count 1.
- The 12 directory components need NO vocabulary injection: measured no `requires-backend`
  tags, component_level='section', is_active — they already flow through load_components
  into available_components. 433's rule supplies the WHEN and the exact name/page_type
  routing, which is the part the DB list cannot carry.

Next: B3f (434, the 194/215 jsonb_set pattern on completeness-discovery-agent), then
council submission for 433 + commit, then the Phase C pilot.

### 2026-08-15 (fifth session, cont.) — B3f applied + live; both submissions dispatched

**B3f DONE, LIVE** — migration
`sql_for_agents/434_enable_finance_directory_and_structural_checks_b3f.sql` (+ surgical-
inverse ROLLBACK), applied 22:38Z after a clean dry run. Eleven checks enabled on
completeness-discovery-agent: the three finance directory pairs and the five Phase-A
structural checks (`dead_internal_link_live`, `canonical_mismatch`,
`structured_data_invalid`, `head_essentials_missing`, `sitemap_entry_dead_live`).
Checks array 32 → **43, read back live**.
- **Correction to the handoff's framing**: B3f said "the 6 directory checks"; the three
  AI-kind pairs (model/adoption/protocol) were ALREADY enabled by 194/215 — measured in
  the live array before writing the migration. The six that were missing are the FINANCE
  pairs. Same count, different six; the migration enables what was actually absent.
- **Binary precondition proven at the artefact, not inferred from the tag** (an
  unregistered check NAME is silently skipped, so a config-only enable can look applied
  and do nothing): pod `agent-chassis-584b6fcf-9mtqd`, `grep -aq … /proc/1/exe` —
  `missing_mortgage_lender_directory_section` PRESENT, `sitemap_entry_dead_live` PRESENT,
  positive control `missing_model_directory_section` PRESENT, negative control
  `missing_zzz_nonexistent_check_qqq` ABSENT. Both controls in the same breath, per
  LANDMINES; no `strings`, no discovery grep.
- **Zero work items expected until the Phase C pilot opts in** — the directory checks
  self-gate on the per-site flag (no site carries a finance key today) AND on current
  found claims of that kind (B4 populated all three). That silence is the design, not a
  failure; the pilot is what arms them. The five structural checks fire on any dispatched
  completeness sweep — their real-site finding volume is UNMEASURED (A1 shipped them
  flag-only for that reason), so the first post-434 sweeps want watching.

**Both migrations submitted to the council gate** (config under docs/, so `FORCE=1`):
- 433 (B3e planner rule) → corr `53ae1501-abe0-4bce-8376-bf20e220faf7`
- 434 (B3f checks enable) → corr `1b087280-b43b-4ea6-82c1-818c4f3cbef8`
Commits carry `Council-Submitted:` (verdicts pending; 098 credits them on approval).
Budget ~30 min each — the council itself is 2–5 min, the dispatch queues behind the fleet.

**NEW submission-schema gotcha, fourth in the series** (429 cost two, 432 one, this one
cost a rejected submit): **a sketch whose every non-blank line is a comment is REFUSED
client-side** — *"a fix plan proposes changes, not observations"*. 434's first draft
sketched the migration as an annotated comment block (accurate, readable, and rejected).
Fix: sketch the REAL statements (guards, UPDATE, verify block) and move the commentary to
`.rationale`/`.grounded_in`. Note this compounds the 432 lesson rather than replacing it:
reviewers can only see what the sketch shows, AND the sketch must be code.

### 2026-08-16 — 433 round 1 REVISE: the gating objection answered at the artefact, and one objection that earned a real change (441)

**Council latency, measured**: 433/434 were submitted ~22:45Z on 08-15 and did not START
until **16:08Z / 16:10Z on 08-16** — ~17.4 hours queued. CLAUDE.md's "~30 minutes" is a
floor under normal load. A `fix_plan` artifact with no `council_report` row means RUNNING,
not dropped; diagnose by artifact KIND, never by elapsed time.

**433 round 1: REVISE** (`decided_by: gating objection from editquality`, 5 abstained, not
truncated). Dispositions:

- **editquality [HIGH, gating] — "the rule names `content_features`, but is that data
  actually piped into the prompt via input_fields? If not, the rule is inert."** A sharp
  objection and the right failure mode to name. **ANSWERED AT THE ARTEFACT, and the first
  read looked like it CONFIRMED the objection** — the newest rendered prompt contains
  `content_features` exactly TWICE, both occurrences inside the RULE text, none in the
  data. The trap there is that the rule text and the data both contain the same substring,
  so a naive `LIKE '%content_features%'` cannot tell them apart. Re-asked in the DATA form
  (`content_features:map[`, which the Go map rendering produces and the rule text cannot):
  **5 of 66 real `plan_site` renders carry it, and the same 5 carry `news_feed:map[`.**
  The other 61 are sites whose classification spec has no such key — so the count tracks
  the data and could have come out otherwise. Verbatim from the 08-11 render:
  `content_features:map[news_feed:map[reason:… recommended:true separate_page:true …]]`.
  End-to-end precedent for the same mechanism: the News listing rule has produced **7 live
  `news-index` pages across 6 domains**, including relojistas.com's localised
  `noticias-index` — exactly what that rule's text asks for.
- **bug_historian [MEDIUM] — "does the prompt teach byte-identical component names? a
  paraphrase is dropped silently."** **RIGHT, and it produced migration 441.** Two parts:
  (1) the underlying safety claim is now MEASURED — `content_components.name <> function`
  for all 12 directory rows (`name` is prose, "UK Mortgage Lender Directory"), and 433
  teaches the **function** values, which is the canonical resolution target:
  `componentNameResolver.resolve()`'s FIRST arm is *already a valid function → return
  unchanged* (`v3_site_actions.go:3896-3898`). Teaching `name` would have been exactly the
  defect the seat feared. **I had measured `function` and never checked whether `name`
  differed — the plan was on the right side of this by construction, not by checking.**
  (2) A real weakness the objection lands on: 433 enumerated every component and page name
  EXCEPT the listing component, which it asked the model to DERIVE by suffixing. The one
  derivable name was the one exception. **441 enumerates all six listing names** and adds a
  sentence stating that a paraphrased component name is dropped from the plan silently.
  Applied 2026-08-16 after a clean dry run (UPDATE 1 + UPDATE 1); read back live.
- **reuse_agent [MEDIUM] — "does the OLDER path become a no-op, or a second creator?"**
  A no-op, by construction: `MissingDirectoryPageCheck` returns empty when `pageCount>0`
  (`check_directory.go:338-348`); `MissingDirectorySectionCheck` the same on a
  `page_components` join (251-266). The plan-time path satisfies the check's precondition.
- **debug_historian [MEDIUM] — "no `snapshot_agent()` before the jsonb_set."** The applied
  file's FIRST line is exactly that, and the guard RAISEs if the backup row is absent.
  **Sketch-visibility artefact — the FOURTH round this lane has lost to showing reviewers
  less than the file contains.** Now a checklist item, not a lesson: *quote the snapshot
  call and the guard block verbatim in every sketch.*
- **guardian [MEDIUM] no `SELECT … FOR UPDATE`**: acknowledged unchanged (same as 432 —
  matches the platform's verify-block idiom; the in-txn DO/RAISE aborts on that drift).
  **[LOW] blast radius**: now named — `plan_site`'s output feeds the greenfield build
  pipeline and `recompose_pages` redesign runs. **[LOW] prompt cost**: now MEASURED rather
  than estimated — real renders are 40,897–85,183 bytes (n=66), so 433's ~2.4KB is ~3–6%.
- **prior_art [LOW]**: the claims it could not verify from its tier are the ones measured
  above; nothing was left asserted.

**Side finding, unlooked-for**: migration **439** (another lane, landed mid-session) added
`menu_field: "available_components"` to `validate_plan`, so the planner's own menu rows now
join the valid set alongside the section/element base. Strengthens 433/441 — both
resolution paths cover the 12 — and is a reminder that this prompt has several editors.

**Ordering trap created by 441, stated in both headers**: 441 edits text INSIDE 433's
inserted block, so while 441 is applied, 433's surgical-inverse ROLLBACK cannot match its
literal and **refuses** (by design). To undo the pair, run **441's ROLLBACK first**.

**Round 2 resubmitted** same correlation (`RESUBMIT_CORR=53ae1501…`): 4 edits (433 forward
unedited as the record of what ran, 433 ROLLBACK, 441 + its ROLLBACK), 12 grounded_in
items, snapshot+guards quoted verbatim. **434's verdict (`1b087280…`) was still executing
at the time of writing — read it next.**

### 2026-08-16 — 434 APPROVED round 1; three seats raised one shared concern, and the measurement cleared it

**434 (B3f, corr `1b087280…`): APPROVED at round 1**, 16:21:17Z —
`decided_by: approved with 3 advisory objection(s) — none high-severity`, 6 abstained,
not truncated. Commit `a1b92f609` carries `Council-Submitted:`; 098 credits it on report.

**The shared concern, raised INDEPENDENTLY by editquality, guardian and debug_historian:**
`snapshot_agent()` runs before `BEGIN`, so it autocommits outside the guarded transaction —
and my own stated workflow (dry run, then real apply) calls it TWICE with the same
`…: pre-update` reason. All three cited the 2026-08-16 LANDMINE: a replay's snapshot is
labelled *pre-update* and holds *post-update* config, so "the latest snapshot" is the wrong
pre-image.

**MEASURED, using that landmine's own content-not-recency check — and it clears:**
```
SELECT snapshot_reason, snapshot_taken_at,
       (default_config#>>'{workflow,steps,plan_site,config,prompt_template}' LIKE '%Directory rule:%')
FROM agent_definitions_backup WHERE type='build-site-planner' …
```
433's two rows: **both `f`** (pre-change). 434's two rows: **both 32 checks** (post-change
is 43). 441's two rows: **both `f`** on its new clause. **Why the seats' reasoning does not
reach this case:** a dry run is the same file with `COMMIT`→`ROLLBACK`, so the UPDATE is
undone and the second snapshot still sees the pre-change row. The decoy needs a replay of an
**already-succeeded** apply — 417's case, not this one. Two consequences worth keeping:
a pre-flight that **refuses when already applied** cannot produce the decoy at all (the
replay aborts before the UPDATE), and duplicate `pre-update` rows should be CHECKED rather
than assumed poisoned — distrusting a good snapshot loses the restore path just as surely.
Contributed back to the LANDMINES entry as a dated refinement, with the measurements.

**guardian [MEDIUM] — "no `handler_agent` verified for the five structural checks; will
items pile up unconsumed?"** Answered by the file's own header
(`check_site_structural_validity.go:86-96`): all five register with **`HandlerAgent ""`
deliberately** — the flag-only idiom shared with `check_asset_reference_404` and
`check_site_unreachable`. Findings surface as visible `detected` items and are
intentionally NOT dispatchable. The header also explains why an auto-repair would be
actively harmful today: a `canonical_mismatch` fixer is gated on bugs_open/251's fix being
reachable by every render path that owns a `<head>`, and today it covers only the
page-rerender path, not `AssemblePageAction`'s three other callers — so an auto-fixer would
rewrite a canonical that the next render would immediately un-fix. **Not a gap; a designed
posture the seat could not see from the sketch.** Recorded here so it is not re-raised.

**editquality [LOW]** — the UPDATE's WHERE re-checks only one of the eleven names while the
DO block checks all eleven: accurate, and the transaction boundary makes the window nil.
Noted, unchanged. **guardian [LOW]** — filed as `add` rather than `config_change` naming the
owning pipeline: convention point, taken for the next one.

**Status: Phase B's council trail is now 429 APPROVED · 432 APPROVED (r2) · 434 APPROVED ·
433 round 2 pending** (`53ae1501…`, resubmitted 16:28Z with 441 attached).

**CORRECTION to commit `0b125c532`'s message (same session, 2026-08-16).** That message
says LANDMINES.md rode along in it carrying the `mortgagecalculator_couk_adoption` lane's
restructure as *my* passenger. **It did not — the file is not in that commit at all.** What
actually happened is the mirror image: while I was writing the message, that lane committed
LANDMINES.md themselves (`d0dd4bec9`), and **my contributed bullet went in as THEIR
passenger.** By the time my pathspec commit ran, the file was clean, so git took only the
two remaining files and reported exactly that.

Nothing is lost — the bullet is at HEAD (verified: `git show HEAD:…/LANDMINES.md | grep -c
"CONTRIBUTED 2026-08-16, portfolio_positioning lane — the DRY-RUN"` → 1) and the doc_notes
sync + verifier dispatch had already run (corr `2c4189d9`). Forward-only forbids an amend,
so the correction lives here.

**The transferable bit, and it is the reason this is written down:** I checked the passenger
situation *before* composing the message and then committed *after* — and on this tree that
gap is enough for the answer to change. A same-file passenger check is only true at the
instant `git commit` runs. **Read the commit's OWN output**: it prints the scope block
naming every file it actually took, and mine said two files while my message described
three. The check that would have caught it costs nothing — compare the scope block against
what the message claims, before moving on.

### 2026-08-17 — 433 round 2 APPROVED: Phase B's council trail is COMPLETE (429/432/433/434 all approved)

**433 round 2 APPROVED** 2026-08-16 16:38:18Z — *"approved with 3 advisory objection(s) —
none high-severity"*, 4 abstained, not truncated. **All four Phase B rounds are now
approved: 429 · 432 (r2) · 433 (r2) · 434.** 098 credits the `Council-Submitted:` commits
automatically at report time.

**Fresh chassis v1.0.1305 rolled** (both replicas, 2026-08-16 22:07/21:08Z). Phase B is
config-only so a roll cannot undo it, but B3f's checks DO depend on the binary carrying
their names, and a fresh build is not automatically a newer commit — so re-probed with
controls on the new pod (`agent-chassis-5657f446c7-q7b82`):
`missing_mortgage_lender_directory_section` PRESENT · `sitemap_entry_dead_live` PRESENT ·
`evaluate_directory_features` PRESENT · POS control PRESENT · NEG control ABSENT.
**Preconditions hold on 1305.** (The `build provenance` startup line had already scrolled
after ~13h — absence there means "not in range", not "unstamped".)

**Advisory dispositions:**

- **debug_historian [MEDIUM] — the post-UPDATE verify blocks never re-check `p IS NULL`
  before doing arithmetic. REAL, and it is the "a check that cannot fail" class**: if the
  SELECT returns no row, `p` is NULL, every `length()`/`position()` is NULL, `NULL <> 1` is
  NULL, no IF fires and **the verify PASSES having inspected nothing** — while the
  pre-flight in the same file refuses on exactly that condition. Asymmetric in the direction
  that matters. **FIXED in both un-run ROLLBACKs (433 + 441)**, each with the reason inline.
  The applied FORWARD files are deliberately NOT edited — they are the record of what ran
  (432's precedent) — so the gap is recorded here instead: in-transaction the row cannot
  vanish, which is why this is a latent-not-live defect, and the next planner migration
  should carry the guard from the start.
- **editquality [MEDIUM ×2] — "the 433 rollback sketch is not a runnable script" and "441's
  rollback UPDATE contains a literal `...` elision that will never byte-match".** Both are
  **sketch-visibility artefacts: measured `grep -c '\.\.\.'` on the 441 ROLLBACK file = 0**,
  and the 433 ROLLBACK file is a complete `BEGIN`/`DO`/`UPDATE`/verify/`COMMIT` script. But
  this is the **FIFTH round this lane has lost to showing reviewers less than the file
  contains**, and the second where my own abbreviation *looked exactly like a bug a
  reviewer should catch*. **New hard rule, not a lesson: NEVER elide inside a sketch's
  string literals, and never sketch a rollback as fragments — paste the whole file if it is
  under the size cap.** A reviewer cannot distinguish brevity from breakage, and objecting
  is the correct thing for them to do with an elided literal.
- **bug_historian [MEDIUM] — "441 patches the SYMPTOM; validate_site_plan drops an
  unresolvable section with no error, no work item, no log."** **The class is right, the
  current-state premise is OUT OF DATE and the correction matters for the pilot.** Read at
  source: `validate_site_plan` calls `recordDroppedSectionNames`
  (`component_name_resolver_menu.go:208-221`), which files a DURABLE finding via
  `LogActionFindings` under provenance `validate_plan`; there is a per-drop `Logger.Warn`;
  and `warnUnrecordedDrops` (lines 244-252) makes even a FAILED record loud, explicitly so
  the report cannot reproduce the silence it removes. That is bugs_open/282's fix, another
  lane, already shipped. So a dropped directory section **is observable today** — which
  turns the seat's objection into a concrete PILOT CHECK (below) rather than an open risk.
  The seat's LOW rider is fair and stands: the resolver's drop behaviour is generic across
  the whole rule, not just the listing field 441 narrowed.
- **reuse_agent [LOW] dual-path** (reactive check vs proactive planner) — answered in round
  2's grounded_in: both checks return empty once the page/section exists, so the older path
  becomes a no-op, not a second creator. **guardian [LOW]**: 433 and 441 are not
  independently reversible (441 first) — correct, by design, stated in both headers and the
  handoff. **prior_art [LOW]**: `llm_call_log` is outside its queryable schema, so the
  5-of-66 evidence is credible-but-unverified from that seat; it is reproducible by anyone
  with DB access and the query is in round 2's grounded_in.

**PILOT CHECK earned by the bug_historian round** — after the Phase C plan is written, ask
whether validate_plan dropped any section name for that site, instead of inferring from a
missing page:
```sql
-- table is agent_error_log (NOT agent_errors); key on the error_code, which is
-- unambiguous — `action` carries values like 'validate_plan' / 'apply_gap_plan:new_page'
SELECT occurred_at, agent_type, action, error_message, context
FROM agent_error_log
WHERE error_code = 'PLAN_SECTION_NAME_DROPPED'
  AND site_id = '<pilot-site-id>'
ORDER BY occurred_at DESC;
```
A row naming a directory component = the planner emitted a name that did not resolve, which
is precisely the failure 441 exists to prevent and the one a silent-looking success hides.

> **CORRECTED 2026-08-17, before anyone ran it:** the first version of this block queried
> `agent_errors` with a `details` column. **No such table** — it is `agent_error_log`, and
> the payload column is `context`. Caught by doing the thing this file keeps telling people
> to do: `\d` the table before writing SQL into a doc. An unrun query in a handoff is a
> claim like any other, and this one would have failed in the next session's hands at
> exactly the moment they needed it.
>
> **And the zero it returns today is DEMAND-CONTROLLED, so it can be trusted** (checked
> 2026-08-17, because "0 drops" and "the recorder never runs" look identical):
> `PLAN_SECTION_NAME_DROPPED` = **0 rows, all history** — but the findings door is
> demonstrably live (**12,158 rows in 7 days across 63 agent types**, newest 12 minutes
> before the check), and the recorder is in the RUNNING binary (`grep -aq
> "PLAN_SECTION_NAME_DROPPED" /proc/1/exe` on v1.0.1305 → PRESENT, POS control PRESENT,
> NEG control ABSENT). So the zero means *no section name has been dropped*, not *nothing
> is watching*. **Residual, stated:** the door is proven for other call sites, not
> positively demonstrated for `validate_plan` specifically — no drop has ever occurred to
> prove it end to end. The pilot is its first real exercise, which is another reason to run
> the query there rather than assume the silence.

### 2026-08-17 (cont.) — Phase C pilot SEEDED and DISPATCHED; the seed's own claims guard was inert

**Pilot dispatched** ~11:4xZ, corr `fb048d5f-b4b3-49c8-bc02-2810bbe209aa`.
`domain-submitter` COMPLETED; `needs_domain_research` triaged to
`domain-research-classifier`. Mission stored at 3,010 chars with all three marker sentences
intact (`urgency without alarm`, `10-second answer`, `non-price facts`) — the 082 script
single-lines and escapes the file, so verifying the markers survived is not ceremony.
Pre-seeded `evidence_base` + `imagery_style_guide` both survived the submitter (`is_current`).

**PRE-FLIGHT THAT PAID FOR ITSELF — and it produced `bugs_open/292`.** Before dispatching I
asked what the classifier actually writes for finance sites, because the directory flag
depends on it. Measured: `industry` is **NULL** and `site_type` `"interactive-platform"` on
all four comparable sites — *neither is in `verticalDirectoryMap`*. So for this whole family
the **domain-derived signal is the only one that can fire**. Chasing that led to the defect:
the domain-signal loop ranged the map (randomised) appending EVERY match into a
first-match-wins dispatch, over a map that deliberately mixes recommending and
NOT-recommending entries. `mortgage-refinance.co.uk` — M4, the pilot's own family, because
`"refinance"` contains `"finance"` — **flipped per run**. Reproduced on iteration 1; fixed;
600×3 green. Pilot domain unaffected (single keyword). Full write-up in the bug file; the
transferable half is in 016b §9.

**THE SEED'S OWN GUARD WAS DEAD, AND MY VERIFY BLOCK SAID IT WASN'T.** All six
`banned_claims` patterns inert on first apply — dollar-quoted SQL passes bytes literally and
JSON *then* unescapes, so `\\\\b` stored `\\b`. `claims.go` falls back to `QuoteMeta` only on
a compile ERROR, and a double-escaped pattern is *valid regex that matches no English*: it
compiled, the fallback never fired, the guard was loaded, listed, **counted**, and dead.
My verify asserted `jsonb_array_length = 6` and passed. **Six inert patterns count exactly
like six working ones** — the check could not have come out otherwise, which is the entire
objection to it. What caught it: probing four must-match strings and one must-not.

Three failed corrections, each worth keeping because they are the same family as the bug:
1. the `£` guard searched for the character it wanted to CONFIRM — this channel rewrites
   `\uXXXX` into the character it denotes, so what I typed is not what landed (`chr(92)` now);
2. the `LIKE` guards were unreadable — **in `LIKE` the backslash IS the escape character**,
   so `'%\b%'` means "the letter b" and would have passed on any pattern containing `b`
   (`position()` now, which has no escape semantics);
3. the first probe used Postgres `~`, but production compiles with Go RE2 and **PG spells
   word boundary `\y` while `\b` is backspace** — a check in the wrong engine, which is its
   own false signal. Semantics now pinned in Go beside the production compile call
   (`datahelpers/claims_banned_pattern_escaping_test.go`, 6 must-catch + 4 must-allow + the
   mechanism itself).
Then the corrected guard was **run against the still-broken rows and required to flag them**
(6 of 6) before being trusted on the fixed ones. LANDMINE filed.

**Landmine footprints, corrected mid-session**: another lane filed an entry warning that
prose and globs cannot substring-match a path. Ran their parser against my two entries —
both had prose/glob parts, and comma-splitting had cut junk fragments (`config-only`,
`no roll needed"`) out of my parentheticals. Both rewritten as short comma-separated keys and
re-verified through `landmines_lib.parse()`. Reading their entry cost a minute; my entries
were partially inert without it.

**292 APPROVED unanimously at round 1** (corr `d9ca49ae…`, 2026-08-17 11:15Z, `decided_by:
all reviewers approve`, 6 abstained) — and the queue was ~3 minutes this time, against 17.4
hours two days ago. Worth noting for anyone sizing a submission: the latency is not a
property of the gate, it is fleet load, so neither figure is "the" number.
One LOW advisory, editquality: the regression test assumes `matchVerticalDirectory`'s
signature and param order, and a wrong guess would fail to compile — *"which is the only
thing standing between this defect and a silent return"*. Already satisfied: the test
compiles and passes (600×3 green, and it FAILED on iteration 1 before the fix, which is what
establishes it can detect the defect at all). No change. Phase B + 292 now have four
approvals and one unanimous.

### 2026-08-17 (cont.) — the pilot is BLOCKED, and not by anything in this lane: the build pipeline is head-of-line stuck fleet-wide

Watched the pilot for 20 minutes; `needs_domain_research` never left `triaged`. Rather than
wait, checked the dispatch path — and the fleet's whole build pipeline is stalled.

**The chain is alive.** `build-pipeline-trigger` is `enabled`, 60s interval, and completed
**27 runs in 40 minutes**; `build-dispatch-loop` likewise; 48 work items across 11 types and
6 sites were touched in the window. So this is not a dead fleet and not a dead scheduler —
which is exactly what makes it hard to see.

**The selector takes ONE row, oldest-first, fleet-wide.** From the live config,
`build-pipeline-trigger.find_dispatchable_site`:
`… WHERE s.locked_at IS NULL AND wi.status IN ('triaged','approved') AND wi.attempt_count <
wi.max_attempts … ORDER BY wi.created_at ASC, wi.priority ASC, wi.id ASC LIMIT 1`.
Note it does NOT filter `pipeline='build'` (the scheduled task's own gate query does) — it
takes the oldest eligible item in the estate, whatever it is.

**The head row never moves, so nothing behind it ever runs.** Measured: the last 8
consecutive trigger runs all picked the SAME site (`webdesign.co.uk`,
`6b49db8e-…`); its item `8c921926-…` (`add_tool`, created 11:19:24Z) still reads
`status=triaged`, `attempt_count=0` — while `updated_at` bumps every run (12:08:35Z), so the
row IS being touched and is going nowhere. `call_dispatch`'s response carries that item back
inside a `pending` block with `has_items: true`, and the orchestration finishes
`COMPLETED` / `current_step=complete` with **no `__step_error`**. Every layer reports success.

**My pilot is 5th of 36 eligible items** and cannot be reached until the head clears. Ordering
is `created_at ASC`, so new arrivals queue BEHIND me — my position can only improve; it is
head-of-line blocking, not starvation-by-reordering (which is the defect 429 fixed in the
publish trigger, `ORDER BY domain LIMIT 5`).

**What the blocking item is, and why this is NOT ours to clear.** It is
`add_tool_novel_webdesign.co.uk` — the `webdesign_tool_rebuilds` lane's owner-directed
REPLACEMENT #2, rebuilding a ported tool at an EXISTING page (`page_id` set,
`/tools/ab-test-calculator/index.html`, handler `tool-generator`). That is exactly
**`bugs_open/286`**'s scenario, whose fix is BUILT + council-APPROVED but **NOT LIVE** — the
Go rides the next chassis roll. **Hands off**: another lane's item, another lane's bug, and
CLAUDE.md says never cancel a failing row pre-diagnosis.

**It is also NOT `bugs_open/029`.** 029 is hung spawns saturating the `dispatch` concurrency
group — orchestrations stuck in `AWAITING_RESPONSES`. Here every orchestration COMPLETES.
Different mechanism, same fleet-wide consequence.

**Filed `090` rather than asserting the root cause** (CLAUDE.md's default: cross-cutting, and
the cause is non-obvious after a proper look — I can show the selector and the stuck row, but
NOT why the dispatcher returns an item as `pending` without claiming it).
**`RUN_CORRELATION_ID=5fbb7f4c-9968-4b95-9048-caad202cea4a`** — claimed by the dispatch loop.
Artifacts are keyed under THAT id, not the intake one.

The open structural question the run should settle, and it outlives this incident: the
selector has **no skip and no backoff** for a head item that is returned but never claimed,
so one un-claimable row halts every build on the estate while all instrumentation reads green.

### 2026-08-17 (evening) — THE PILOT BUILT, AND ALL THREE PROOF POINTS PASSED

The head-of-line blocker cleared at 12:12Z (webdesign.co.uk's `add_tool` completed — that
lane's own business), and the pilot then ran the whole pipeline unattended:
`needs_domain_research` 12:20 → `needs_vertical_research` 12:29 → `needs_strategy` 12:34 →
`needs_briefing` 12:39 → `needs_site_plan` 12:44 → `needs_composition` 12:46 → `needs_design`
13:13 → pages. **So the 090 I filed was answered by events, not by the loop** — the stall was
transient head-of-line blocking, and the diagnosis run (`5fbb7f4c…`) is still worth reading
for the structural half (no skip, no backoff), but the incident is over. Do not read its
verdict as describing a live outage.

**PROOF POINT 1 — 432 LIVE AND CORRECT.** The classification spec carries exactly:
`content_features.mortgage_lender_directory = {recommended: true, kind: "mortgage-lender",
separate_page: true, reason: "Mortgage sites gain authority from a cited, verified directory
of UK lenders"}`. The pre-flight predicted this and named the reason (`industry` NULL,
`site_type` not in the map, so the DOMAIN signal is the only one that can fire) — prediction
and outcome match.

**PROOF POINT 2 — 433 + 441 LIVE AND CORRECT, on a real build.** The planner produced:
| page | page_type | sections |
|---|---|---|
| `index` | landing | hero, mortgages-repayment, brief-explanation, info-card-grid, **mortgage-lender-directory**, call-to-action |
| `mortgage-lenders` | **mortgage-lenders** | **hero, mortgage-lender-directory-listing, call-to-action** |
Name and page_type both exactly the `directoryCheckProfiles` values; composition exactly the
prescribed order; `in_header=true`. **The homepage section AND the dedicated page, both from
the rule, on the first real build.** 441's enumeration did its job: the listing component
name is byte-correct, which is the one the old text asked the model to derive.

**PROOF POINT 3 — zero `PLAN_SECTION_NAME_DROPPED` for this site.** So no section name was
silently dropped, and proof point 2's composition is what the planner actually emitted rather
than what survived a cull.

**COST BASELINE (the Phase C deliverable).** Joined to the pilot's own 83 orchestrations, NOT
a time window — a window would have swept in council, diagnosis and landmine-verifier spend
from other lanes and inflated it several-fold:
**43 LLM calls · 389,406 input tokens · 120,822 output tokens · 11 assets.**
By agent: page-content-writer 29 calls (276,562 in / 51,592 out) — the overwhelming majority;
component-creator 6; classifier 2; vertical-exemplar-researcher 2; planner, strategist,
briefing, webdesign-agent 1 each. **[FLOOR, not a total]** — the join catches orchestrations
whose `collected_data` names the site, so any run that does not is missed. Treat as a lower
bound. (The payload search costs >120s and timed out once; run it in background.)

**WHAT FAILED — the pilot's real yield.** Nothing here is caused by this lane's work:
- **11 × `failed to get latest commit/base tree`** (10 `needs_imagery` + 1 `needs_page`
  deploy) — a deploy-path failure, logged under `LLM_API_ERROR`, which is a misleading code.
  **`sites.github_repo` is EMPTY — and that is NORMAL, not the bug**: 6 of the 8 most recent
  sites are empty, including established ones, because they serve by the B2 route. So the
  obvious reading ("the repo was never created") is wrong, and the real question is why the
  asset/page deployer took a git path on a B2 site.
- **20 × `unrendered_template` `{{end}}` blockers** on 2 pages (`CONTENT_VALIDATION_FAILED`).
  **Checked specifically: NOT my banned_claims** — the guard did not over-fire and blocked
  nothing. A component is leaking raw Go template syntax into rendered output.
- **6 × `component_validation_rejected`** for `mortgages-repayment` (a homepage section the
  planner chose), **2 × `needs_new_component` failed** at `store_generated_component`
  ("generated template …"), **1 × `needs_rerender` failed** after 3 retries (timeout).
- **10 × `unresolved_cta` + `needs_section_data` + 2 × `needs_page` at
  `needs_human_review`** — the HITL queue, expected on a new site, not failures.

**THE ROLL SHIPPED NO NEW CODE, AND I CHECKED RATHER THAN ASSUMED.** Pods restarted 14:42/
14:43Z but the tag is **unchanged at v1.0.1305**, and the binary carries **none** of this
session's commits (`e0d662243`, `1268ae2ef`, `07b2ea6d5`, `a1b92f609`, `7aff17b21` all
absent; POS control present, NEG absent). `286`'s `adopt_existing_page` IS present, so it was
already in 1305. **`bugs_open/292`'s fix is therefore STILL NOT LIVE.** Corroborated
independently: another lane filed a MEMORY entry the same day — *"a FRESH BUILD CAN SHIP NO
NEW CODE — same-tag rebuild serves the CACHED image: 203 commits unshipped 08-17 while pods
looked new"*. **`IMAGE_TAG` must be bumped; a restart on the same tag is not a release.**
Note one probe arm was worthless and I should not have written it: `grep -aq "bestDomainKey"`
tests for a LOCAL VARIABLE name, which Go strips — it would read ABSENT on a binary that has
the fix. The commit-stamp probe is the one that carries the result.

### 2026-08-17 (evening, cont.) — the pilot's deploy failures are a LIVE FLEET-WIDE OUTAGE, not a pilot fault

Chased the 11 deploy failures and they are the local edge of something much larger.

**The split is clean, and it is not credentials.** Every site taking the 404 has
`github_repo = (EMPTY)`; every site with a real repo has none:
| site | github_repo | base-tree 404s (4h) |
|---|---|---|
| remortgagecalculator.uk | (EMPTY) | 11 |
| robot-hands.com | (EMPTY) | 9 |
| loancalculator.co.uk | (EMPTY) | 4 |
| noted.co.uk | vm-sites | 0 |
| webdesign.uk | vm-sites | 0 |
| oufe.com | (EMPTY) | 0 |
**The adapter is HEALTHY**: its own log shows `deploy_page` succeeding for noted.co.uk at
16:25:37Z with `repo_name: vm-sites`, `success: true`. So the token works, the pod works, and
"GitHub is down" / "the token expired" are both refuted. **Caveat on oufe.com**: 0 there is
absence of ACTIVITY, not evidence of success — it did not deploy in the window, and I have
not shown it would succeed.

**Scale and timing (measured):** ~832 `base tree` 404s since **13:31Z**, still firing at
16:14Z. `page-rerender` 488 · `build-dispatch-loop` 254 · `asset-deployer` 40 ·
`image-build-handler` 12 · `rerender-pages` 10 · `page-build-handler` 9. **808 of them carry
a NULL `site_id`**, so anyone counting by site sees a small fraction of the damage.

**Two things make it hard to see, and both are worth fixing regardless of the cause:**
1. It is logged under **`error_code = 'LLM_API_ERROR'`** — a deploy fault filed under an LLM
   code. Grepping for a deploy problem does not find it.
2. Most rows have no `site_id`, so per-site triage under-reports it ~70x.

**Filed `090`: `RUN_CORRELATION_ID=75220928-935a-4e5d-8982-802992b0af34`.** What I could NOT
establish first-hand, and therefore did not assert: WHICH component chooses the git route
versus the bucket route, and what changed at 13:31 so that no-repo sites reach the git path
at all. `deploy_config` is `{}` on the affected sites AND on long-established ones that have
deployed fine for weeks, so the route is decided somewhere else.

**Not this lane's to fix** — shared deploy infrastructure — but it is the reason the pilot
site is built and not published, so it belongs in this lane's handoff as the blocking item.

### 2026-08-17 (late) — v1.0.1307 IS a real build, 292 is LIVE, the outage is over, and B3f's volume is finally measured

**The tag moved: v1.0.1305 → v1.0.1307.** Another lane found and fixed the same thing I
reported, independently and in the commit message: *"BUILD: bump IMAGE_TAG to v1.0.1306 —
v1.0.1305 was reused, so the cluster has been re-serving a cached 08-16 binary and 24 code
commits across ~10 lanes are inert"* (`aa9c7b74f`), then `e24dc0e6c` → v1.0.1307.

**`bugs_open/292`'s fix is LIVE.** `git merge-base --is-ancestor e0d662243 e24dc0e6c` → true,
and the pods (started 17:05:46Z) run the image tagged by that commit (16:36:35Z). So the fix
shipped. **292 can be closed once someone confirms at the artefact** — I am asserting
ancestry, which is the documented test, not a binary read.

> **CORRECTED — MY OWN PROBE METHOD WAS WRONG TWICE TODAY, and it nearly produced a confident
> false negative.**
> 1. `grep -aq "<my commit sha>" /proc/1/exe` **only matches if the build was made from
>    EXACTLY that commit.** The binary carries ONE stamp — the commit it was built from — not
>    every ancestor. So ABSENT means "not built from this commit", *not* "does not contain the
>    fix". I read it the wrong way at first and would have reported 292 as unshipped.
> 2. `grep -aoE "[0-9a-f]{40}"` to *discover* the stamp returned **20 strings, none of them a
>    real commit** (`git cat-file -e` refused all 20). That is the trap LANDMINES already
>    describes — it matches Go's internal digit table. I ran it having read that entry.
> 3. `grep -aq "bestDomainKey"` (yesterday) tested a **local variable name**, which Go strips.
> **The method that works, in order:** read the service's own `build provenance` line
> (`kubectl logs -l app=<svc> --tail=300 | grep -m1 'build provenance'`) — but it is a STARTUP
> line and had already scrolled here after ~1h; failing that, take the tag-bump commit from
> `git log -p -- makefile` and run **`git merge-base --is-ancestor <your-commit> <that
> commit>`**. Never a discovery grep, never a symbol that is not a string literal.

**THE DEPLOY OUTAGE IS OVER, and the zero is demand-controlled.** Since the 17:05Z roll:
**0** `base tree` 404s, against **117** other errors in the same window (so the error door is
live) and real deploy demand — 89 `page_rerender` triaged, 7 claimed, 2 failed, 16
`contrast_failure` completed. Compare ~832 404s in the preceding three hours. **The 090
(`75220928…`) was overtaken by the roll** — it still has no verdict; read it for the routing
question if it lands, but the incident is closed. **What I never established stays
unestablished**: which component chose the git route for a no-repo site, and what changed at
13:31. If it recurs, that is the question.

**B3f's structural checks — the "[UNMEASURED] volume" from the last handoff is now MEASURED:**
| check | items | sites | first seen |
|---|---|---|---|
| `head_essentials_missing` | **247** | 8 | 2026-08-16 10:01Z |
| `dead_internal_link_live` | 6 | 3 | 2026-08-17 12:32Z |
| `canonical_mismatch` | 4 | 3 | 2026-08-17 12:32Z |
| `structured_data_invalid`, `sitemap_entry_dead_live` | 0 | — | — |
247 is a real backlog, not noise — but all are `detected` and **flag-only by design**
(`HandlerAgent ""`), so they consume no handler capacity and dispatch nothing. Treat as
intelligence to triage, not an incident. **The two silent checks are the ones to be careful
about**: zero could be "clean estate" or "never exercised", and I have NOT distinguished them.

**The six directory checks are at ZERO — and this time that is a PASS, provable rather than
assumed.** The pilot is the only opted-in site and it HAS its `mortgage-lenders` page with the
right composition (verified directly at proof point 2), so
`missing_mortgage_lender_directory_page` SHOULD be silent. The net stayed quiet because there
was nothing to catch, which is only a meaningful statement because the page was checked first.

**Pilot unchanged since 16:21Z** — its failed and `needs_human_review` items are untouched.
Now that deploys work again, the failed `needs_imagery`/`needs_page`/`needs_rerender` items are
candidates for a retry, but **their `attempt_count` is exhausted (1–3 of max 3)**, so they will
not retry themselves. That is the next concrete action for this lane.

**Pilot's outage casualties RE-QUEUED (2026-08-17 late).** The `stale-work-item-reaper` only
touches `triaged` rows 48h+ (`pre_query` read, not assumed), so nothing was going to retry
these. Re-queued **11** items — 10 `needs_imagery` + 1 `needs_page` — with a predicate that
names the site, `status='failed'`, the `%base tree%` error signature, AND
`attempt_count < max_attempts`, so it cannot reach anything else. **`attempt_count` left
ALONE at 1 of 3**: they have budget, and resetting it would erase the fact that they already
failed once. The in-transaction guard required exactly 11 triaged and exactly **3 still
failed** afterwards — the three genuine failures (`needs_new_component` ×2 at
`store_generated_component`, `needs_rerender` timeout, all 3/3) are deliberately NOT retried,
because they failed for their own reasons and would simply fail again.

**`bugs_open/292` → `bugs_closed/292`** — fixed AND live in v1.0.1307, which is the bar.

**CORRECTION ×2 on the retry watch (2026-08-17 18:35) — both were my errors, and both would
have misled.**

1. **I reported "the retries are succeeding" on the strength of `complete:1`. That row was
   PRE-EXISTING** — the single `needs_imagery` that completed at 13:27Z, before the outage.
   **No retry has completed.** The watch's own first line (`t+20s: claimed:1 complete:1
   triaged:9`) shows the 1 was there from the start; I read a baseline as a result. The check
   that would have caught it costs nothing: a count that is already non-zero at t=0 cannot
   evidence anything at t+n — **take the baseline before you start watching, or watch a delta.**
2. **My watch then printed "A RETRY FAILED AGAIN — outage not fixed". That verdict is WRONG.**
   The loop keyed on `failed:[1-9]` — on the FACT of a failure, not its CAUSE. The new failure
   is `call_asset_deployer … timed out after 3 retries`, and **base-tree 404s remain at 0 since
   the roll**. A different failure is not evidence about the outage. The rule I wrote into the
   handoff ("fail again → the outage is NOT fixed") was too crude for the same reason and is
   corrected there.

**Actual state:** the base-tree outage stays FIXED (0 since 17:05Z). The live constraint is
now **asset-deployer TIMEOUTS: 18 since the roll, across exactly 1 site** — the pilot, which is
the only site with imagery work in flight, so this is concentrated rather than fleet-wide.
Pilot `needs_imagery`: 8 triaged, 1 claimed, 1 complete (the old one), 1 failed at
**attempt 2 of 3 — it still has a retry**. Assets unchanged at 11, so nothing new has landed.
**Too early to call either way**, and that is the honest read rather than a lean in either
direction: image generation is genuinely slow, and one timeout in a draining queue is not yet
a pattern. What to measure next is whether `assets` for this site rises above 11.

### 2026-08-18 — THE FULL CHAIN IS PROVEN AT THE RENDERED ARTEFACT (and my "decisive" measurement was wrong)

**v1.0.1308 running** (tag moved again — 1307 → 1308; the reuse problem is genuinely fixed).
Base-tree 404s since the 08-17 roll: **0**, now over ~17 hours. The outage stays closed.

**The retry worked.** `needs_imagery` went 1 complete / 10 failed → **9 complete / 2 failed**;
`needs_page` 4 → **9 complete**. So 8 of the 11 items I re-queued succeeded.

> **CORRECTION — the measurement I named as decisive was the WRONG ONE, and I wrote it into
> the handoff.** I said: *"does `assets` for this site rise above 11?"* It did not, and that
> is CORRECT BEHAVIOUR, not a failure. The 8 completed items reference **8 asset ids that all
> predate the outage** (created 13:26–14:10Z); **zero new rows**. The images were generated in
> the original run; the step that failed was `call_asset_deployer` — **deployment**, not
> generation — so a successful retry deploys an existing asset and must NOT mint a duplicate.
> `"stored": true` in the result refers to the S3 object, not to a new DB row.
> **The lesson: I picked a counter that measures the wrong half of the pipeline.** Before
> naming a measurement "decisive", say which STEP it observes and check that the step is the
> one that failed. The right checks here were item completion plus the rendered artefact.

**THE END-TO-END PROOF — the whole point of Phases B and C.** `mortgage-lenders` is
**`build_status='deployed'`, `deployed_at` set**, and its three components are rendered:
`hero` (3,010 chars), **`mortgage-lender-directory-listing` (4,882)**, `call-to-action` (2,373).
Read the rendered HTML rather than trusting the status — it contains:
- the heading *"UK mortgage lenders, listed"*;
- the compliance posture **in the copy itself**: *"notes facts such as their regulator status
  and the types of mortgage they offer. **It does not list rates, fees or APRs**, because those
  change daily and depend on your circumstances, so check current pricing with the lender
  directly or with a broker"* — the owner's non-price ruling, surviving all the way to the page;
- real cited entries: **Mansfield Building Society** and **Family Building Society**, each with
  `lender_type` / `product_types` claims and a **`source`** link per claim.

So the chain is closed at the artefact: researcher → verified quote-checked claim → register →
kind-aware publish → `evaluate_directory_features` flag → planner rule → page → deploy →
rendered page naming real regulated firms with citations and no prices.

**Where the pilot actually stands:** 3 of 6 pages deployed (`mortgage-lenders`, `next-steps`,
`about`); 3 at `needs_rebuild` (`index`, `what-your-number-means`, `six-month-checklist`).
`sites.build_status` is still `pending`. Remaining work, unchanged in character: 3 ×
`needs_new_component` failed, 1 × `needs_rerender` failed, 2 × `needs_imagery` failed, and a
HITL queue of 10 × `unresolved_cta` + 4 × `needs_page` + 1 × `needs_section_data`.

### 2026-08-18 (cont.) — the pilot is DEPLOYED BUT UNREACHABLE: its domain still serves a registrar parking page

Asked for the pilot's URL and nearly handed over a false one. **`curl` returned 200 for
`https://remortgagecalculator.uk/mortgage-lenders.html` — and the body is
`<script>window.onload=function(){window.location.href="/lander"}</script>`.** That is the
registrar's parking page, which answers **every** path with 200. The status code was not
merely uninformative, it was actively misleading, and the estate's own rule caught it:
**trust the rendered artefact, never the status.** A `%{http_code}` probe against a parked
domain is a check that cannot fail.

**So the pilot's true state is: BUILT and DEPLOYED TO THE BUCKET, NOT SERVED.**
`pages.build_status='deployed'` + `deployed_at` set is truthful about what the pipeline did —
it wrote the tree to `b2://portfolio-sites/remortgagecalculator.uk/` — but **nothing points
the domain at the serving worker**, so no visitor can reach it. `deployed` and `reachable` are
different facts and this lane had been treating them as one.

**The worker/DNS chain works where DNS exists** — the demand control that makes this a
configuration gap rather than a broken pipeline: `ai-agent-orchestration.com` serves real
built pages through the same worker (`scripts/cloudflare/worker.js`, objectKey =
`<hostname><path>` in bucket `portfolio-sites`), including all three live directories, with
real content (*"Gemini 3.7 Flash · Google · Google's most capable Flash model…"*).

Bucket objects cannot be verified directly from here — `s3.us-east-005.backblazeb2.com/portfolio-sites/...`
returns **401** for the pilot AND for the known-good established site, so the 401 says
"private bucket", not "missing object", and is not evidence either way.

**LIVE AND VIEWABLE TODAY (all 200, real titles, real entries):**
- `https://ai-agent-orchestration.com/model-directory.html` — *AI Model Directory*
- `https://ai-agent-orchestration.com/adoption-tracker.html` — *Enterprise AI Agent Adoption Tracker*
- `https://ai-agent-orchestration.com/protocol-tracker.html` — *Agent Communication Protocol Tracker*

**NOT viewable:** the pilot's `mortgage-lenders.html`. Its content is verified in
`page_components.rendered_html` (two cited building societies, non-price copy) but it is
behind an unpointed domain. **A DNS/domain step is missing from the fleet-build path**, and
Phase E would hit it on all ~140 domains — worth deciding whether the pipeline should own it
or whether it is deliberately manual.

### 2026-08-18 — COST BASELINE CORRECTED: the figure I published in four documents was a snapshot of a still-growing set

**Published: 43 calls · 389,406 in · 120,822 out. Actual: 73 calls · 663,759 in · 184,596 out** —
about **70% higher**. Confirmed identical across two independent runs of the same query.

**Why the first number was low, and it is not a query bug.** Same CTE, same join, run ~2 hours
apart. The CTE selects orchestrations whose `collected_data` names the pilot's `site_id` — and
**`collected_data` is written as an orchestration progresses**, so a run that was mid-flight
when I measured had no `site_record` yet and was invisible to the join. I measured a set that
was still being written and reported it as a total. The `[FLOOR, not a total]` marker I put on
it was correct in direction and far too weak in magnitude — a floor 70% below the answer is
not a useful floor.

**The check that would have caught it:** re-run the measurement once the work is quiet and
require the two runs to agree before publishing. A single measurement of a live system is a
sample, not a total — and this lane already knows that (`a record goes stale faster than its
reader can tell`); I applied it to other people's data and not to my own.

**Corrected baseline — ONE domain, TEXT generation only:**
| model | calls | input | output | at today's rates | at standard rates |
|---|---|---|---|---|---|
| claude-sonnet-5 | 57 | 545,887 | 95,247 | $2.04 | $3.07 |
| claude-sonnet-4-6 | 15 | 98,976 | 85,709 | $1.58 | $1.58 |
| claude-opus-4-6 | 1 | 18,896 | 3,640 | $0.19 | $0.19 |
| **TOTAL** | **73** | **663,759** | **184,596** | **$3.81** | **$4.83** |

Rates: Sonnet 5 $3/$15 per Mtok with a $2/$10 introductory rate through **2026-08-31**;
Sonnet 4.6 $3/$15; Opus 4.6 $5/$25. **So the fleet figure moves on 2026-09-01** — ~$534 for
140 domains today, ~$677 after. Anyone quoting the cheap number after August is quoting a
rate that has expired.

**What this figure still does NOT include, stated so nobody treats it as all-in:**
- **The 11 images.** Different provider, not in `llm_call_log` — genuinely unmeasured, not
  estimated-as-small.
- **Rework.** This build ran through the deploy outage and repeated work; a clean run should
  be cheaper, and I have no clean run to compare against.
- **Cache reads.** `llm_call_log` has no cache columns, so cached input is billed at a lower
  rate than the arithmetic above assumes — the true figure is at or below these numbers on
  that axis, and above them on the images axis.

**The fix is the plan the owner already has**: run the next domains one at a time and measure
each from a quiet start to a quiet finish. Three clean runs beat one contaminated one, and
this pilot cannot be re-run clean.

### 2026-08-18 — OWNER RULINGS: Phase C SIGNED OFF; nmsvr.uk parked; DNS via Cloudflare; copy handed to its lane

**1. PHASE C IS SIGNED OFF (owner, 2026-08-18).** The gate the plan put before Phase E is
passed. What was signed off: the machinery proven end to end at the rendered artefact, and the
cost baseline below. Phase D's decisions are now the only thing between here and Phase E waves.

**2. nmsvr.uk is PARKED (owner).** *"we can leave nmsvr.uk for now."* Own-nameservers is not
being built. The reason it was raised and then parked is worth keeping: sites are served by a
Cloudflare Worker (`portfolio-sites-router`) reading `<hostname><path>` from the
`portfolio-sites` B2 bucket, so authoritative DNS elsewhere would take the worker OUT of the
request path. Revisit only alongside a decision to change how sites serve.

**3. DNS: point domains at Cloudflare, in bulk groups (owner).** Confirmed mechanism, measured
2026-08-18: `ai-agent-orchestration.com` → `alexis/leah.ns.cloudflare.com` and serves;
`remortgagecalculator.uk` → `ns1/ns2.dan.com` (Dan.com parking) and serves a lander on EVERY
path. See the runbook file for the exact steps and what I could and could not do.

**4. The directory-page copy voice is the `copy_quality_two_stage` lane's (owner).** Handed
over with evidence in
`copy_quality_two_stage/CONTRIB_2026-08-18_the_negative_default_survives_a_POSITIVE_identity_spec_on_directory_pages.md`.
SendMessage failed — that session is not currently running — so it is a document, not a
message. **The finding that makes it worth their time: that lane's 2026-08-12 root cause
(negativity inherited from `identity.key_differentiators`) does NOT fit this case** — I read
the site's differentiators and they are entirely positive ("Fast deployments", "Solid
enterprise-level security"…) while the CTA copy is still negative, five days after their fix.
Positive input, negative output: a second path, or a site-scoped fix that did not generalise.
**I did not rerender and did not touch the writer** — the affected pages are another lane's
site, and the obvious fix may be the wrong one.

### 2026-08-18 — COST BASELINE INCLUDING IMAGERY (owner asked for images, over-estimated)

**Text is measured. Images are NOT, and cannot be from here** — checked: no cost/price column
exists anywhere in the schema for assets or LLM calls, and the images come from a **Google**
model (`banana/gemini-3-pro-image-preview` generated all 11 pilot assets; fleet-wide the split
is 350 banana / 79 SDXL / 88 derived). So the per-image rate below is an **[ASSUMED]** figure
with the arithmetic exposed, not a measurement — substitute the real rate from the provider's
billing console and the table still works.

**Text (measured, 3 agreeing runs): $3.81/domain today · $4.83 from 2026-09-01** when Sonnet
5's introductory rate ends.

**Per-domain TOTAL at today's text rate:**
| images/site | @$0.04 | @$0.08 | @$0.15 | @$0.25 |
|---|---|---|---|---|
| 11 (pilot actual) | $4.25 | $4.69 | $5.46 | $6.56 |
| 30 | $5.01 | $6.21 | $8.31 | $11.31 |
| 50 | $5.81 | $7.81 | $11.31 | $16.31 |

**Fleet of 140, today:** 11 img → $595–$918 · 30 img → $701–$1,583 · 50 img → $813–$2,283.
**From 2026-09-01:** 11 img → $738–$1,061 · 30 img → $844–$1,726 · 50 img → $956–$2,426.

**The owner has said he wants "loads of imagery on each site", so read the 30–50 rows.** The
headline: **imagery overtakes text as the dominant cost somewhere between 30 and 50 images per
site at any rate above ~$0.08.** That is the number worth pinning down before Phase E, and one
provider invoice after the next few builds settles it — the *count* per site is ours to
choose, the *rate* is theirs to publish.

### 2026-08-18 — OWNER RULING: L9 reassigned — loanzy.uk is an EXAMPLE SITE, no register entry

*"we can use loanzy.uk as an example site which means no prior registry entry and it should be
built just from the webdesign.uk prompt."* Started by the owner in a separate thread.

**This resolves the Phase D conflict item.** L9 (loan brandables) had been "direction
unassigned, deliberately" and was contested between this lane and the `webdesign` lane; it is
now neither lane's proposition — it is a demonstration of what the framework produces from a
prompt alone. **Removed from this lane's build order** (`PLAN_2026-08-18_first_50…`).

Worth noting for whoever runs it: "no prior registry entry" means the directory machinery will
correctly do **nothing** there — `evaluate_directory_features` writes no flag without a
matching vertical, so no `content_features` key, no directory page, no directory checks firing.
That is the designed behaviour, not a failure, and it makes loanzy.uk a clean control for what
the framework does *without* the Phase B additions.

### 2026-08-18 — first-50 build order proposed, awaiting approval

`PLAN_2026-08-18_first_50_build_order_FOR_APPROVAL.md`. Five waves, M→B→I per the standing
ruling, 42 fresh builds + the pilot + 8 named reserves.

**The finding that shaped it, and it cuts against the ruling's order:** the mortgage-lender
register holds **2 entities / 3 current found claims**, against savings-provider 13/15 and
health-insurer 10/15 (measured 2026-08-18). M is the family ruled FIRST and has the LEAST
directory data — eleven mortgage sites built today would each carry the pilot's same two-row
lender directory. **Recommendation: force-trigger the finance researcher on `mortgage-lender`
before Wave 2** (the B4 procedure, already proven). It does not block Wave 1, and it is the
difference between a 2-row and a 15-row directory across the whole M family.

### 2026-08-18 — WAVE 1 STARTED: build #1 dispatched, researcher armed, one conflict raised not resolved

**Owner approved the first-50 order** ("do as you direct") and ruled on B8/B9/I10 (recorded in
the register with shapes, not just yes/no).

**Mortgage-lender researcher force-triggered** (`last_triggered_at=NULL` on
`mortgage-lender-directory-discovery`, the B4 procedure, guarded to exactly one armed task). It
is weekly and last fired 08-15, so it would not otherwise run until 08-22. This is the
2-lenders-vs-13 gap; it does not block Wave 1.

**BUILD #1 DISPATCHED: `adversecreditmortgage.co.uk` (M5)**, corr
`4a128539-8a51-4e4c-a410-15dffa3f6946`. Seeded first (site row + email, evidence_base,
imagery_style_guide); `needs_domain_research` triaged; mission stored at 3,185 chars with its
marker intact.

**The seed carries the pilot's escaping lesson as a mechanical check.** Single backslash in the
file this time, and the verify block **probes rather than counts**: it refuses on any pattern
containing a double backslash (the signature of the inert-guard bug) and requires ≥6 patterns
to carry a real word boundary. Then proved in **Go** semantics, matching production's
`regexp.Compile`: **6 of 6 must-catch strings caught** ("Guaranteed acceptance…", "No credit
checks required", "Fixed at 6.49% APR"…) and **4 of 4 must-allow strings allowed**. The pilot's
six patterns were inert and its verify passed; these are not, and it was proved the same way it
would have been disproved.

**Why M5 first, not M3 as originally listed — and this is a reorder, not a substitution.**
`mortgage-rates.co.uk` is a rate-table proposition, and (1) the owner's own non-price ruling
forbids per-lender rates, (2) we have no verified-facts pipeline for rates, so it would have
shipped as **a rate site with no rates** — exactly what the pilot's empty roster produced, but
on a site whose entire premise is the numbers. **Held at the end of Wave 1 with three options
put to the owner** (market-level dated sourcing / reshape away from live rates / wait for a
rates pipeline). M5 is a better build #1 anyway: it stress-tests the claims guard on the
audience where a false promise does actual harm.

**Watch on this build:** it should get `content_features.mortgage_lender_directory` at
classification (domain contains "mortgage"), then a `mortgage-lenders` page — and its directory
will show **2 lenders** until the researcher run lands. Also expect the negative-CTA copy
defect handed to `copy_quality_two_stage`; it is unfixed, so build #1 will exhibit it.

### 2026-08-18 — DNS: `remortgagecalculator.uk` zone CREATED and configured; awaiting Nominet

Owner minted a proper token (`~/.config/cloudflare/portfoliotoken`, All zones —
Zone:Edit/DNS:Edit/Workers Routes:Edit, no expiry) after correcting my wrong instruction (the
zone-creation right is under the **Zone** group scoped to all zones on the account, not under
Account). **Done from here:**
- zone `c7ef25edb1221fb4ffc4d4dade271781`, status `pending`
- **nameservers: `alexis.ns.cloudflare.com`, `leah.ns.cloudflare.com`** — same pair as the
  working sites; these go into Nominet (owner)
- one proxied apex A record → `192.0.2.1` (TEST-NET-1, RFC 5737 — cannot route anywhere; the
  worker answers before origin)
- worker route `remortgagecalculator.uk/*` → `portfolio-sites-router`

**Config matched to the reference, not guessed.** With DNS:Read finally available I read
`ai-agent-orchestration.com`: **one apex A record, one route, nothing else.** I had added a
proxied `www` CNAME — removed it, because with no `www.*` route it would proxy to a dead origin
and fail, which is worse than not existing.

**FLEET-WIDE FINDING, not fixed:** `www.ai-agent-orchestration.com` **does not resolve at all**
(apex returns 200). No `www` record or route exists on the reference zone, so by implication on
none of the 36. **Anyone typing `www.` gets a DNS failure.** Left alone deliberately — fixing
one zone would diverge it from 35 others; it is an owner decision, recorded in the runbook.

**Still blocked on Nominet** — until the nameservers change, the domain keeps answering from
Dan.com's parking lander. **Remember the trap: a parked domain returns 200 on every path**, so
verify by reading the body, never the status code.

### 2026-08-18 — OWNER: BUILDS HALTED; classifier gets the register (option 2, filed as RFC_037)

**HALT IS IN FORCE.** Owner: *"Stop the builds until we sort out the classifier and which
builder flow we are using."* Implemented with the platform's **own** pause switch rather than
by mangling work-item statuses: `sites.locked_at` is exactly what
`build-pipeline-trigger.find_dispatchable_site` excludes on (`WHERE s.locked_at IS NULL`).

- Locked: `adversecreditmortgage.co.uk` (build #1, mid-flight) and `remortgagecalculator.uk`
  (pilot — it still has failed/HITL items that could be retried).
- `locked_by` states who and why, so another session finds a reason and not a mystery.
- **Verified against the dispatcher's own predicate**: both now read
  `dispatcher_would_select = false`. Build #1's `needs_strategy` stays `triaged` — **queued work
  is preserved, not cancelled**; clearing `locked_at` resumes it exactly where it stopped.
- Checked before using it: nothing auto-clears `sites.locked_at` (the `locked_at` hits in Go are
  all `site_components`/`pages`), so the pause is durable rather than a lease that expires.
- **Build #1 got as far as `needs_strategy` and its classifier had already run — proof point 1
  passed a second time**: `mortgage_lender_directory`, recommended, `separate_page: true`.

**Option 2 chosen for the classifier → `architecture_review/RFC_037_…md`.** Filed rather than
built: it adds an input to `domain-research-classifier`, a shared seam every site in the fleet
passes through. The RFC carries the measurement (7 finance sites → 2 distinct classifications,
`industry` null on all 7), the prior art it must NOT duplicate (`vertical_landscape` already
does this outward, at competitors, and never at our own portfolio), and four design questions
I deliberately did not answer alone — chiefly where the register data lives, and whether the
input is advisory or a binding collision check.

**The two paused decisions are COUPLED, and that is the thing to decide first.** If each site
keeps a hand-written mission carrying its register positioning, the classifier change is
belt-and-braces. If sites are to be built from a short prompt — the `loanzy.uk` model, *"built
just from the webdesign.uk prompt"*, no register entry — then the classifier reading the
register is **the only remaining place differentiation could come from**, and RFC_037 becomes a
precondition for the fleet rather than an improvement to it.

---

## 2026-08-18 (evening) — three measurements: DNS after the Nominet change, the citation queue, and whether the two flows are actually two flows

### 1. DNS — the nameserver change LANDED; `www` was never a nameserver problem

Measured 2026-08-18, from this session, resolver + HTTPS in one pass:

| host | DNS | HTTP |
|---|---|---|
| `ai-agent-orchestration.com` (reference) | `2606:4700:3033::6815:eef` | 200 |
| `www.ai-agent-orchestration.com` | **NXDOMAIN** | — |
| `remortgagecalculator.uk` (pilot) | `2606:4700:3037::6815:11f8` | 200 |
| `www.remortgagecalculator.uk` | **NXDOMAIN** | — |

**The pilot is serving its own pages at its own domain for the first time** — body read, not
status code (`<title>Remortgage Calculator UK — Your Number, in Seconds</title>`, 40,726 bytes,
the framework's own CSS-variable stylesheet). That closes the §3 blocker in
`HANDOFF_2026-08-18b`: the Nominet delegation to `alexis`/`leah.ns.cloudflare.com` is live.

**`www` is unchanged and the nameserver change could not have changed it.** Delegation decides
*which nameserver answers*; `www` needs *a record to answer with*, and there is none — the
reference zone carries one apex A record and nothing else, and the worker route is
`<domain>/*`, which does not match `www.<domain>`. So the RUNBOOK's fleet-wide finding stands
verbatim, now with a second zone confirming it. **[MEASURED]** — the disconfirming result was
available: had a `www` record existed, `getent hosts` would have returned an address.

### 2. The citation queue — what is actually in it (`directory_citation_unverified`)

Four rows open, one per directory kind, all at the `system.internal` pseudo-site
(`eac60db8-…`), all `needs_human_review`, all `handler_agent='human-review'`:

| item_key | rejected candidates | classes |
|---|---|---|
| `…:company` | 10 | citation_lost |
| `…:model` | 12 | citation_lost |
| `…:mortgage-lender` | **1** | citation_lost |
| `…:protocol` | 4 | citation_lost, fetch_error |

**The handoff's "4 items" is 4 QUEUE ROWS, not 4 candidates** — 27 rejected candidates in
total, and only ONE of them is a mortgage lender. Re-read `HANDOFF_2026-08-18b` §3 with that
correction: working the mortgage-lender row cannot produce more lenders, because it holds one
reject and that reject is not a new lender.

The single mortgage-lender reject: `family-building-society` / `product_types`, class
`citation_lost`, url `…/mortgages/later-life-mortgages`. **It is not a loss.** Two current
`found` claims for that entity already stand (`lender_type`, `product_types`, both verified
2026-08-15) — this is a re-extraction whose longer quote no longer matches verbatim, the same
shape the 08-15 health-insurer row was ruled on. The verbatim gate working, not a regression.

**Why the register still shows "2 lenders" against 4 entity rows** — the two extra rows are
`status='archived'`: `fca-regulated-mortgage-lenders` ("FCA-regulated mortgage lenders
(general)") and `uk-specialist-lenders-sector`. Both are CATEGORY-shaped, not named firms, and
the 08-15 B4 session's "discard" ruling archived them and added the named-firm rule to
`extract_claims` (migration 423). `QueryDirectoryEntries` filters `status='active'`, so they
cannot reach a page. The two active ones are Family Building Society and Mansfield Building
Society. **So the directory's problem is ACQUISITION (the researcher finds ~2 firms per run),
not the review queue.**

### 3. The two builder flows ARE one flow — measured at the handler, not asserted from the script

`DECISION_2026-08-18_two_builder_flows_side_by_side.md` asserts "same entry script and same
agent graph" from the script. Here is the same claim measured at the work items the builds
actually produced — the row that names the agent that did the work:

| domain | needs_site_plan | needs_composition | needs_page | needs_imagery |
|---|---|---|---|---|
| remortgagecalculator.uk (A) | `build-site-planner` | `site-design-planner` | `page-build-handler` ×14 | `image-build-handler` ×11 |
| adversecreditmortgage.co.uk (A) | `build-site-planner` | `site-design-planner` | `page-build-handler` ×18 | `image-build-handler` ×20 |
| loanzy.uk (**B**) | `build-site-planner` | `site-design-planner` | `page-build-handler` ×16 | `image-build-handler` ×16 |

Identical handlers, identical item types, identical order
(`needs_domain_research → needs_strategy → needs_briefing → needs_site_plan → needs_design →
needs_composition → needs_page/needs_imagery`). **loanzy.uk drove off the same work-item queue
as both flow-A sites.** Its rows are `cancelled` because the site was cleared, not because a
different mechanism built it.

**Where "pageflow builder" comes from, and why it is a red herring here.** `pageflow-builder`
is a live agent, and the classifier's prompt still tells it to emit
`"recommended_builder": "pageflow-builder"` (`003_site_classifier.sql`). But: only TWO live
agent definitions mention the string at all (`domain-research-classifier`, which writes it, and
`pageflow-builder` itself, which is named by it) — and **no Go code in `platform/` or
`internal/` reads `recommended_builder`** except doc-comment examples of the generic
`agent_type_field` dynamic-dispatch mechanism (`spawn_actions.go`, `call_agent.go`). It is a
field from the older intake route that the build pipeline does not dispatch on. **[MEASURED]**
by `default_config::text LIKE '%pageflow-builder%'` over live non-snapshot agent definitions,
and by grep over the two Go trees.

### 4. Misstep in this session, caught by my own aggregate

Reading `family-building-society`'s claims I saw two rows with `is_current = t` for the same
`field` (`product_types`) and started to write it up as a supersede defect. The fleet-wide
check — `GROUP BY entity_id, field HAVING count(*) > 1 WHERE is_current` — returned **0 rows**,
which refuted it in one query. The cause: `UNIQUE(kind, slug)`, so Family Building Society
exists TWICE, once as `mortgage-lender` and once as `savings-provider`, and my slug-only filter
joined both entities' claims into one listing. **A per-entity read cannot see a per-(kind,slug)
key** — filter by kind, or join and print it.

---

## 2026-08-18 (evening, second session) — why the pilot has no calculator, and the directory going 2 → 25

### 1. The missing tools: BOTH keys are right and they are DIFFERENT keys

`remortgagecalculator.uk` served with no calculator. Full case + fix candidates:
**`bugs_open/311`**. Diagnosis run through `090` — **CONFIRMED, first iteration**
(run corr `8aa2e283-129f-41d1-93a0-6dcacbbabeae`, intake `5f0798b3`).

The chain, each step measured:
1. Planner planned section `mortgages-repayment` on `index`.
2. `SelectComponentByType` matches `WHERE section_type = $1 …` — the component that
   IS that calculator (`function='mortgages-repayment'`) has **`section_type` NULL**.
   No match → `not_found` → `needs_new_component`.
3. `store_generated_component` matches `WHERE function = $1 AND forked_from IS NULL`
   — **this finds it**, so the store believes it is a regeneration.
4. The field-contract guard refuses: the new schema drops the 8 fields
   (`button_1, heading_1, label_1..label_6`) that
   **loanandmortgagecalculator.co.uk's** stored `content_data` is keyed on. The
   guard is RIGHT — overwriting would empty that site's live sections.
5. Three attempts (12:51, 14:12, 19:02 on 08-17), all identical. Section left with
   `component_id=''`; page built, deployed, served. `status='active'`.

**The alternative explanation is excluded, not assumed.** The diagnosis flagged
that one site retrying its own regeneration would look identical in the error log.
Settled by joining through `page_components` (there is no `site_id` on
`content_components` — which is exactly why the writer cannot see whose row it is):

| function | rows | depended on by | requester |
|---|---|---|---|
| `mortgages-repayment` | 1 | loanandmortgagecalculator.co.uk | **remortgagecalculator.uk** |
| `loans-credit-health-check` | 1 | loanandmortgagecalculator.co.uk | **loancalculator.co.uk**, **loanzy.uk** |

**Live right now, not historical:** `loans-credit-health-check` retried at 18:02,
18:07, 18:10, 18:14, 18:17, 18:21, 18:25 today, same 18-field rejection each time.

**Class size** [MEASURED]: of 140 active base `component_level='section'` rows,
**26 have no `section_type`** — each invisible to the selector and each a trap for
the next site that names its `function`. 79 `component_level='tool'` rows are
invisible to this selector by construction.

**The part that stings:** an active `tool-mortgage-repayment` section component
(section_type set, 10,760-char template, live since 2026-05-06) was sitting there.
The site needed no new component at all.

### 2. The directory: 2 → 25 mortgage lenders in ~15 minutes

**The measurement that redirected the work** — all-history yield per source domain,
a query that could have come out the other way:

| kind | source | claims | firms |
|---|---|---|---|
| savings-provider | www.gov.uk | 14 | **12** |
| health-insurer | mytribeinsurance.co.uk | 8 | 7 |
| health-insurer | drewberryinsurance.co.uk | 7 | 7 |
| mortgage-lender | fca.org.uk / kbra.com / familybuildingsociety / mansfieldbs | 2/2/2/1 | 1 each |

Every high-yield source in the register's history is a **multi-firm enumeration**.
Every single-firm page yields one firm. The mortgage kind had never had an
enumeration page in its scrape set: its four slots on 08-18 went to two
market-overview pages that name firms without stating quotable facts about them
(ukfinance.org.uk's largest-lenders table, **bsa.org.uk's HOMEPAGE rather than its
member list**) plus two single-society pages. Candidates 2, registered 1.

So the cause was **source shape and slot count**, not the pipeline. Migration
**`471_widen_finance_directory_discovery.sql`** (applied 18:32:55Z, DO/RAISE verified). **⚠ cite the FILENAME:** another session applied a different `471_floor_held_remedy_partitions_failures_first.sql` at 18:30:39Z — `schema_migrations` keys on filename, so both are live and the ledger is right; a bare "471" is what is wrong. Changes:
- `max_scrapes` 4 → 10, `num_results` 10 → 20, `max_snippets` 5 → 8. Sized against
  `bugs_closed/062` (a batch_scrape reply over ~1 MB is dropped and the caller
  starves): the 08-18 run carried **85 kB of scrape_results for 4 URLs**
  (orchestration `a5ba225c`), so 10 URLs projects to ~210 kB — ~5× margin. Measured.
- four enumeration-shaped mortgage queries (BSA member list, adverse-credit
  specialists, buy-to-let, FSCS-protected firms), all < 200 bytes (web_search drops
  a ≥200-char query and blames config keys).
- the prompt line *"Third-party listicles are weak"* narrowed — it was the exact
  opposite of what the register's own history shows.

**Result, four runs, ~15 minutes:**

| query | urls | candidates | registered |
|---|---|---|---|
| BSA member list | 10 | 15 | 13 |
| adverse credit | 10 | 15 | **15** |
| buy-to-let | 10 | 15 | 7 |
| FSCS protected | 10 | 12 | 7 |

**Active mortgage lenders 2 → 25.** Named firms, no category-shaped entities (the
423 named-firm rule held). The adverse-credit cohort — Bluestone, Kensington,
Pepper Money, Vida, Foundation Home Loans, The Mortgage Lender, Aldermore, United
Trust Bank — is exactly what `adversecreditmortgage.co.uk` had nothing to list.
Review queue took 8 mortgage + 5 savings rejects, which is the verbatim gate
working at a normal rate (42 registered / 57 candidates).

### 3. www — what the Nominet change did and did not do

Nameserver delegation landed (`remortgagecalculator.uk` serves its own pages).
`www` was never a delegation problem. Route inventory across all 39 zones, measured:
**24 carry a wildcard `*.<domain>/*` route** (need the DNS record only), **12 have
the apex route only** (need both), and **4 must be left alone**: `idea.uk` and
`relojistas.com` have no route to the worker at all (a proxied A → 192.0.2.1 there
is a 522 black hole), `webdesign.uk` deliberately 302s to webdesign.co.uk, and
`relojistas.com` already serves www. `cookly.uk` and `dartsonline.com` already
301 correctly. `robot-hands.com` and `leopardessconsulting.co.uk` have a www
record with nothing serving it — they time out today, and the change fixes them.

So the earlier "www resolves NOWHERE across all 36 zones" was **too uniform**:
8 zones had a www record, in four different states.

### 4. Missteps this session

- **I wrote "no zone has a www route" from a script whose route check had
  crashed.** The python got `D` as argv instead of env, printed nothing, and
  `[ "" != 0 ]` is TRUE — so eight unconfigured zones reported as "already
  configured". A reading that is not a number must be refused, not compared; the
  script now does that explicitly.
- **I hypothesised that directory facts come from a firm's own page** and was
  going to exclude the trade-body pages on that basis. The per-source yield query
  refuted it in one shot: the single highest-yield source in the register's whole
  history is a gov.uk list page. Widening on the wrong theory would have made the
  mortgage kind worse.

---

## 2026-08-18 (night) — the www→apex rollout, measured at every step

**Deploy** (owner-run, 20:02:37Z): `deploy_worker.sh` reported `success: True` and
`bindings now: []`. **That empty array is the API not echoing them, and it is
indistinguishable from the credential-stripping outage the README warns about** — so it was
verified at the artefact instead: 5 of 6 sampled apexes served 200 with real titles and
plausible byte counts immediately (a stripped B2 binding fails EVERY page, since the worker
SigV4-signs each origin GET with them), `/worker-health` answered, and a missing path still
returned the worker's own 404.

**Full post-deploy sweep, all 39 apexes:** 36 serving. The 3 exceptions are all pre-existing
and none is a regression: `apis.uk` and `ugg2.com` have **no `sites` row at all** (parked
domains that carry a worker route), and `loanzy.uk` was cleared by the owner — though its DB
row still reads `status='deployed'` with 19 active pages, so anything asking the database
will believe it is live.
> One apex (`loancash.co.uk`) read 404 on the first pass and 200 on retry, with
> `cf-cache-status: DYNAMIC` (so live, not a cached answer) — transient, during script
> propagation. **A single reading immediately after a fleet deploy is not evidence.**

**Fan-out:** `add_www_redirect.sh --apply` — **28 DNS records + 7 routes, 0 failed.** Only 7
routes because 24 zones already carried a wildcard `*.<domain>/*` route; adding an explicit
`www.<domain>/*` there would have been redundant. Skipped 3 (see runbook). One zone was
applied and verified ALONE first (`remortgagecalculator.uk`) before the other 35 — the
worker's redirect branch had never executed anywhere at that point.

**Verification: 36/36 applicable zones return `301 → https://<domain>/`**, path and query
preserved (`www.…/about.html?x=1` → `https://…/about.html?x=1`), follow-through 200.

### The two readings that said FAILURE while the change was correct

1. **A newly created worker route 522s.** `https://www.remortgagecalculator.uk/` returned
   **522** on the first request while `…/mortgage-lenders.html` — same host, same route,
   same second — already returned 301. On this estate the origin is `192.0.2.1` (TEST-NET-1,
   which can never answer), so 522 is the *expected* signature of "no worker attached", i.e.
   the reading confirms the fear. 5/5 clean within a minute.
2. **My own resolver's negative cache outlived the record.** `www.vonc.com` and
   `www.webdesign.co.uk` gave `curl: (6) Could not resolve host` for ~4 minutes while the
   Cloudflare API returned the records (`created_on` 20:07:43Z / 20:07:46Z) and
   authoritative DNS answered `rcode 0` with proxy IPs. The dry run had queried both names
   while they were NXDOMAIN, and that negative answer was cached locally. Proven working by
   pinning the IP: `curl --resolve "www.vonc.com:443:104.21.87.185"` → `301 →
   https://vonc.com/`. **Same command, same minute, opposite answers, depending only on who
   was asked.**

Both are in `LANDMINES.md` and the runbook's ✅ DONE section. The dangerous remedy in each
case is the obvious one — delete the record and re-run — which destroys a correct change.

### Correction carried back into the runbook

The 08-18 finding *"`www.` resolves NOWHERE across all 36 zones"* was an **inference from
the reference zone**, and wrong in both directions: 8 of 39 zones already had a `www` record,
in four distinct states, and two of those (`cookly.uk`, `dartsonline.com`) were already doing
exactly what the owner asked for. The exceptions were precisely the zones a blind fan-out
would have broken — which is the argument for classifying per zone rather than looping.

---

## 2026-08-19 — drift check against ~120 overnight commits, and two things that changed the picture

### 1. What was re-verified rather than assumed

| claim | state 2026-08-19 | how |
|---|---|---|
| chassis build | **v1.0.1314, `d3590ca46`** | binary probe, **three-way control**: build sha PRESENT · yesterday's `0b185bad2` **ABSENT** · nonsense sha ABSENT. The `build provenance` startup line had scrolled past `--tail=20000` on both pods, so the log route was unavailable |
| both sites locked | ✅ still, `locked_by` names this lane | `sites.locked_at` |
| directory | **25 mortgage lenders, 17 savings** | unchanged overnight |
| `www` → apex | 301 on all 5 sampled zones incl. the two that were resolver-stale | followed the redirect |
| `bugs_open/311` fixed? | **no** — `store_generated_component_action.go`, `create_tool_component_action.go`, `component_selector.go` all untouched since filing | `git log <sha>..HEAD -- <paths>` returned empty |

### 2. 311 got worse, and then got a cross-lane answer

**Worse:** the `loanzy_uk_example_site` lane reproduced it on a greenfield build and lost **7 of
7** tool sections, then measured the served artefact — `loanzy.uk/tools/loan-comparison-calculator/index.html`
returns 200, 22,600 bytes, **zero `<input>` elements**. A calculator page with no calculator,
live, with nothing in the page to show a reader that anything failed. Our filing had asserted
"built, deployed and served without it"; they attached a URL to it.

**Answer:** `architecture_review/RFC_036` is the same design fact (function fleet-wide, identity
per-site) at the **tool** level, filed independently, with an owner direction and a written path
(§9.3). Until this morning `grep 311` there and `grep 036` here both returned **0** — two lanes
at the same wall, invisible to each other.

**They are distinct, and I checked rather than assumed** — the incumbents
(`mortgages-repayment`, `loans-credit-health-check`, `loans-car-finance-calculator`) are all
`component_level='section'`, and `idx_cc_tool_function_unique` is **partial on
`component_level='tool'`**, so the index cannot be what refuses them; the field-contract guard
is. RFC_036's rows are `tool`-level and are not touched by that guard.

**The consequence that made the cross-link worth writing:** RFC_036 §9.3's change fires only
when a **library (`tool`-level)** row claims the function — so building it would fix their half
and **leave 311 entirely live**, while reading as "tools are fixed". Both files now cite each
other. One council submission covering both writers is cheaper than two rounds.

### 3. The pilot's blocked pages are `bugs_open/260`, and the dead links are the SAME defect

An inbound note from the loanzy lane pointed out that our `HANDOFF_2026-08-18` TODO 6 (file the
`{{end}}` leak) was already `bugs_open/260`, and predicted that blocked pages leave dead internal
links. **Tested on our live site and confirmed:** `/`, `/about.html` and `/next-steps.html` each
link to `/six-month-checklist.html` and `/what-your-number-means.html`, **both 404**. They are
nav links, so it is site-wide, not one stray reference.

So the lane's `{{end}}` TODO and its dead-internal-link TODO were **one defect counted twice** —
the links are dead *because* the pages were blocked. Contributed as §12 of 260. Their headline
still says "no live damage"; it is now wrong on two live sites, and I left the headline to its
owner rather than editing it.

### 4. Summary written

`SUMMARY_2026-08-19_first_sites_live_and_the_wall_the_fleet_would_have_hit.md`. The test in
CLAUDE.md is whether the five headings would differ substantially from the last one: 08-17's
"where we are now" reads *"the pilot site is built but not published… nothing is live yet"*,
which is no longer true in either half. A genuine inflection, so a new file rather than silence.

---

## 2026-08-19 (later) — owner rulings recorded, two threads contacted, and an inference that survived six domains and then met its counter-example

### Owner rulings taken this session

| decision | ruling | where recorded |
|---|---|---|
| RFC_037: where the data lives | **a database** | RFC_037 §5, register header |
| RFC_037: what is authoritative | **the database**; the markdown register stops being the source of truth | same |
| RFC_037: advisory or binding | **advisory** (the binding collision check is deferred, not rejected) | same |
| RFC_037: the ~40 non-register sites | **superseded — a registry for them AND for the rest of the ~2,000 .uk domains** | same |
| 311 / RFC_036 | **one submission covering both writers; treat as a PRECONDITION** for wave 1 | handoff §3, relayed to the `bugs_open/311` session |

**The hard blocker created by the registry ruling, recorded so it is not discovered late:**
there is **no inventory of the ~2,000 domains** — not in the repo (`z_bundles/old/domainsubmit1.txt`
is a log dump, not a list), and not in any `information_schema` column named `domain`/`domain_name`.
Today: **44 register entries · 153 domains in `PORTFOLIO_domains.txt` · 43 `sites` rows.** The
list must come from the owner; it blocks the build, not the design.

### Correction to the owner's premise, stated because acting on it would have been wrong

The owner asked me to contact "the thread that did `remortgagecalculator.co.uk`, which has the
full directory". **That domain does not exist** and no other site carries the lender directory
[MEASURED 2026-08-19]: `sites` holds `remortgagecalculator.uk` (ours), `mortgagecalculator.co.uk`
and `loanandmortgagecalculator.co.uk`; the only `mortgage-lender-directory*` components anywhere
are on **our** pilot; `mortgagecalculator.co.uk` has one `guide-lender-restrictions` guide page
and no directory; and a sweep of five live finance sites for the new lender names (Skipton,
Bluestone, Kensington, Pepper Money, Vida, Aldermore, Leeds BS) returned **zero hits on any of
them**. Our pilot still serves 2 lenders. **I did not guess a thread and message it** — asked the
owner instead.

### The 260 exchange — an inference offered, tested by someone else, and half-refuted

Their ask was CSS classes from our blocker rows, to decide narrow-fix vs render-seam. **The rows
have no classes and an empty `location`**, so that route is dead — reported as a negative rather
than worked around.

What the rows do carry is the token set, and it discriminates better: `{{end}}`×9,
`{{.label}}`×1, `{{if`×9, `{{range`×1. **The 20 is two detectors each capped at 10**, which
explains why every instance fleet-wide reports exactly 20. `{{.label}}` inside a range means an
**array loop that never executed** — not a parse error on arbitrary markup.

Intersecting with what the plan intended (`site_plan_sections`, plan `e743e9b4`):

| page | planned components | array fields | emits `{{.label}}` |
|---|---|---|---|
| six-month-checklist | hero, **mechanism-flow**, info-card-grid, call-to-action | mechanism-flow(steps), info-card-grid(cards) | **mechanism-flow only** |
| what-your-number-means | hero, generic-text-block, **mechanism-flow**, faq, call-to-action | mechanism-flow(steps) | **mechanism-flow only** |

So: `mechanism-flow` is the only array-carrying component on **both** pages and the only one
whose template contains `{{.label}}`. **Offered as a falsifiable inference, explicitly not an
attribution** — nothing in `agent_error_log` names a component.

**Their test: it held on 25 of 26 events across seven domains** — six of which it was never built
on. **The 26th refuted the narrow branch**: `webdesign.co.uk` leaked `{{ variable }}`, with
spaces, a different producer entirely. So the population is **not homogeneous**, and the fix goes
to the **render seam across all 304 components**, not to one component's schema. Both sides
flagged the same limitation unprompted: the value lists inherit the 10-cap, so the finding is
*"consistent with mechanism-flow on every occurrence"*, never *"proven"*.

**Why this is worth the space:** the useful outcome was not the 25 that agreed — it was the 1
that did not, and it arrived because the claim was stated in a form someone else could break.
Also a correction to our own §12: the "four items" are **two pages observed twice** (a build item
plus its re-render sibling), which reads as more occurrences than it is; their census counts
distinct pages and that is the better number.

---

## 2026-08-20 — the guard went live, the estate got inventoried, and four of my own claims were wrong

### 1. CGV-033 is LIVE, and how that was established

Chassis `v1.0.1317`, build point **`2d13d530d`** (2026-08-19 22:21Z). Establishing it took two
corrections of method:

- The `build provenance` startup line had **scrolled past `--tail=200000`** on both pods, so the
  documented log route was unavailable. Found the build point by probing candidate commits from
  the build window against `/proc/1/exe`.
- **I first probed for MY OWN commit's sha and got "absent" on all three candidates** — including
  `d3590ca46`, which had been PRESENT the day before. That is the landmine my own handoff quotes:
  the binary carries the ONE commit it was built from, so your commit is always absent and the
  reading means nothing. The correct test is `git merge-base --is-ancestor <commit> <build point>`.

By ancestry: the guard (`b8e2e9cbe`), its round-2 tests, `cmd/regcheck` and the attestation
recorder are all **in**. The section-editor arm (`ae7a8d739`, committed today) is **not** —
it needs the next roll.

**Not proven at the artefact.** No refusal has been observed; both sites are locked so there has
been no demand. A zero here is no-demand, not a working guard.

### 2. The estate: 1,567 domains classified from the registry export

| class | n |
|---|---|
| PARKED (marketplace/for-sale DNS) | 1,247 |
| NO_DELEGATION (registered, NS never set) | 207 |
| **CLOOK / real hosting** (`dns.uk-noc.com`) | **62** |
| REGISTRAR_DEFAULT | 19 |
| OTHER (unknown — 2 AWS, 1 Hetzner, 8 domainmanage) | 18 |
| CLOUDFLARE (ours) | 14 |

The 62 serve real content (`cartoon.co.uk` 133 kB, `businessinsurancequotation.co.uk` 78 kB).
**People's-name domains: 5.** **50 test candidates picked.**

**Using the export's own `dns0..dns9` beats a live lookup twice over:** it is the registrar's
record, and it distinguishes "registered, never delegated" from "lookup failed" — a live query
returns an ambiguous nothing for both, and 207 domains are in that state.

### 3. Four of my claims, corrected — three by the owner, one by the council

- **"The planner does not know what we can build." FALSE.** `build-site-planner` has a
  `load_components` step querying the live library, and `plan_site`'s `input_fields` include
  `available_components`. It is handed the catalogue every run. *Useful by-product:* tool-level
  components are included only if the site already has `plan_includes_tools` AND the tool is
  already on one of its own pages — so a greenfield build sees no library tool, which is
  `bugs_open/311`'s upstream half.
- **"Games have no mechanism." FALSE.** `tool-drop-rate-tuner` is 22,230 chars of live
  interactive JS on `gamesdesign.co.uk`, with `tool-xp-curve-designer` and
  `tool-gacha-pity-designer` alongside. A game is an interactive tool.
- **"NXDOMAIN may mean the registration lapsed." Owner: "no nameserver usually means I never set
  a nameserver."** My reading would have sent someone to re-buy domains he already owns.
- **"ScanAllBannedClaimsWithSuppressed is the single function every enforcement surface calls."
  FALSE, and the council caught it.** `section_editor_actions.go` has zero claims-guard
  references and writes `page_components.rendered_html` directly. Fixed in code.

### 4. Two tool defects the real data exposed, and one claim I made about my own work

- **The name extractor was wrong in BOTH directions.** Forename-first: 35 NAME verdicts on the
  real estate, **3 of them people** (`christmasbasket` → "chris"+"tmasbasket", `annualreports` →
  "ann"+"ualreports"). Compound-first: `jamesbrown`, `davidsmith`, `peterhiggins` all became NO,
  because those surnames ARE dictionary words. The order that satisfies both: **known
  forename+surname beats compound; compound beats a speculative split.** 35 → 5, all correct.
  *A precedence bug reads exactly like a data bug.*
- **"The first N that pass" is not a sample.** The picker returned 50 domains beginning with a or
  b, because the eligible list is sorted. Subject variety survived by luck. Replaced with a
  deterministic stride. It had also picked `anne-marie.co.uk` — a person's domain — so it now
  excludes NAME verdicts.
- **I wrote in the 08-20 handoff that council round 3 had been submitted. It had not** — two runs
  existed for the correlation. Caught by counting the runs, which is a one-line query. *A claimed
  submission is the cheapest false claim to make and the easiest to check.*

---

## 2026-08-20 (later) — the brief writer, built, run, and read

### The run

Migration `510` applied and DO/RAISE-verified. Fired at `indoorplanters.co.uk` (one of the 50,
site row created **locked** so nothing could build even by accident), orchestration
`1ea45228-2135-4754-b3d9-d3d792b87df0`. Result:

- `mission_brief` spec written, **12,095 chars**, `source='brief-writer'`
- `needs_brief_review` item at **`status='needs_human_review'`** — held, as designed
- **nothing built** — no pages, no plan, no other work items

### The hold was proven BEFORE it was built on

`approval_mode` is the semantically right mechanism and the dispatcher honours it. Tested against
the dispatcher's own predicate verbatim (`load_work_item_actions.go:709`) on two disposable rows
inside a transaction that was rolled back:

| row | approval_mode | status | dispatched? |
|---|---|---|---|
| A | auto | triaged | **yes** |
| B | manual | triaged | **no** |
| B (after) | manual | **approved** | **yes** |

Both directions. ⚠ **This lever had never been used — all 10,311 rows were `auto`.** But
`create_work_item` has **no `approval_mode` config key**, so an agent cannot set it without a Go
change. `status` is supported, and `needs_human_review` is equally invisible to the dispatcher, so
that is what shipped: zero code, and the estate's existing HITL idiom. `approval_mode` support in
`create_work_item` is recorded as the better long-term shape (BLD-024 verify-later).

### Is the brief any good? — the only question that mattered

**A completed run is not a good brief**, and generic-with-the-nouns-swapped was the failure mode
the prompt argues against. Read it. It is specific:

- **Names real competitors from its own research** and classifies them: general houseplant sites
  (Homestead and Chill, The Spruce) that treat pots as secondary; retailer blogs (Plantatorem,
  Greenhouse Studio) that are "thinly disguised product catalogues"; academic sources (PMC,
  ScienceDirect) inaccessible to a general reader — then states the gap between them.
- **Takes a stance**: *"against the trend of treating pots as pure decor objects divorced from
  plant health, and against vague care advice that gives no concrete guidance."*
- **15 content items across 8 kinds**, honestly prioritised: 5 `core` (planter size, drainage,
  material guide, plant-to-pot matching, a Planter Finder tool), 7 `valuable` (retailer directory,
  comparison table, pot-size calculator, care guides for 30–50 plants), 3 `aspirational`.
- **Three tools with real inputs and outputs**, not "a calculator" — e.g. Pot Size Calculator:
  current pot diameter + root condition → recommended diameter range.
- **It marks its own uncertainty**, unprompted and in the estate's own idiom:
  `research_quality: "adequate — … no direct competitor analysis of UK-specific sites was
  available, so the directory opportunity and UK market gap are INFERRED rather than confirmed"`,
  confidence 0.78. That is the discipline this repo spends `WRONG_CALLS.md` teaching, arriving in
  a generated artefact without being asked for.
- `regulated_subject: false`, correctly.

### Cost, and a measurement problem worth naming

**One brief: 133,948 input tokens, 3,172 output, `claude-sonnet-4-6`, NOT truncated**
(`output_tokens < max_tokens` checked — a completion at the cap is a CUT, not a finish).

Against the estate's measured baseline for a whole site build ($3.81 for 663,759 in / 184,596
out), a brief is ~16% of that token volume — so **an upper bound of roughly $0.60/brief, ~$900
across 1,500**. It is an UPPER bound and the real figure is lower: output tokens cost several
times input, and this run is unusually input-heavy (five scraped pages), so a proportional
estimate over-states it. **[ESTIMATED, not measured]** — no cost column exists.

The input is the cost driver, and it is `max_scrapes: 5`. Halving the scrapes roughly halves the
brief.

⚠ **The spend is not attributable.** The call logged under `agent_type='generic'`, not
`brief-writer` — `created_by` bottoms out at that literal when an agent does not set
`config["source"]` (LANDMINES). So brief-writer spend cannot be separated from anything else's
until the step sets a source. Worth fixing before 1,500 runs make the question interesting.


---

## 2026-08-24 (b) — wiring `render_sitemap`, and the second defect that fell out of measuring the first

### The before-figure, and why the handoff's own number needed re-deriving

The handoff set the close condition as *"a site that did not have a sitemap has one, fetched and
read"*, with **8 of 25** as the before-figure. Re-measured today across every live site:

- **8 of 28** live sites serve a sitemap of ours. The NUMERATOR held; the DENOMINATOR had grown
  by three since 08-20. This is the staleness-by-addition shape the counts ruling exists for.
- Judged by BODY, not status code, and the difference is three sites:
  - `adversecreditmortgage.co.uk` — 200, 171 bytes, one `<loc>` for `/lander`. Parking provider's
    file. The handoff warned about this one and it is real.
  - `noted.co.uk` — 200, **27,414 bytes, `content-type: text/html`**. Its own homepage, served for
    any path. **A second instance of the same trap, which no document recorded.**
  - `webdesign.uk` — **302** to `webdesign.co.uk/`.
- The discriminating test, which is mechanical and could have come out otherwise: extract every
  `<loc>`, strip the origin, match the path against that site's `pages` rows. Ours score **17/18
  to 98/98**; the parking file scores **0/1**. That separation is what makes it evidence rather
  than a judgement call.

### The wiring decision, with the cost measured rather than left as a risk

Round 1's submission listed probe cost at fleet scale as a risk for the reviewers to check. That
is not evidence, so it was measured first: **735 listable pages across 28 live sites, avg 26.3,
max 135** (`webdesign.co.uk`).

| | per firing | when it fires |
|---|---|---|
| rotation, one site/tick | ~26 GETs | 1800s tick, site due after 3 days → ~245 GETs/day |
| page-deploy path | whole site (135 for webdesign.co.uk) | every page change |

Rotation first, on two grounds: it is the only one that reaches the **20 sites with no sitemap at
all** (the deploy path only helps a site that gets rebuilt), and it is bounded. The deploy-path
half stays open — a new page waiting up to 3 days is a real cost, just a smaller one than
re-probing a 135-page site on every edit.

### The misstep: I confirmed the canonical defect from ONE page and nearly shipped a fix 8× too wide

Comparing the served sitemaps against `pages` rows showed `/` in every existing sitemap and
`/index.html` in what the action would emit. I fetched `dartsonline.com`'s homepage, read
`<link rel="canonical" href="https://dartsonline.com/">`, and called it confirmed. **That is one
data point, and the obvious fix from it — `TrimSuffix(p, "index.html")` — would have been wrong.**

Fetching a section index in the next breath gave the opposite convention:
`https://dartsonline.com/guides/` declares canonical `.../guides/index.html`. The listable
population is **27** rows of exactly `/index.html` against **228** of `<section>/index.html`, so
the suffix fix repairs 27 and breaks 228, each against its own canonical tag — and **every broken
URL would still return 200 and pass the generator's own probe**, so nothing downstream would have
objected.

⚠ The counter-example was already in data I had fetched: the existing sitemaps I was comparing
against contained `/guides/index.html` next to a bare `/`, one column over. Logged in
`WRONG_CALLS.md`. The check that generalises: **when a rule is about a pattern in paths, sample it
at two different depths before generalising.** 10 of 10 sites checked agree on both halves, across
all three builders (B2, vm-sites, git-hosted).

Fixed in `5c9acf1bd` by whole-path match. `TestRootIndexIsCanonicalisedButSectionIndexIsNot`
pins the asymmetry and was mutation-proven **in both directions with changes that still compile**:
suffix-match → 5 section cases fail; delete the canonicalisation → root and both-forms cases fail.

### What is committed, and the one thing that is not done

- `5c9acf1bd` — the canonicalisation fix + tests (Go: **inert until the next fleet roll**)
- `0bce1db39` — migration `590`: the `sitemap-refresh` agent + `sitemap-refresh-rotation` task
- `ff55133ac` — register **SEO-007** (`render_sitemap` had **no register entry at all** —
  `bugs_open/106`'s exact failure mode)
- `5f67b977a` — two LANDMINES entries; `WRONG_CALLS.md` row

**Migration 590 is NOT applied.** The live-DB write was refused by this session's tool-permission
classifier, not by any check of ours. Everything up to the write is done: scoped dry run clean,
probe transaction ran to its own COMMIT and rolled back, and all four verify guards were INDUCED
to fail (typo'd action name; conditional default that never reaches the commit; `locked_at` guard
removed; `files_field` pointing at a field `render_sitemap` never writes). **So the close
condition is NOT met — no site has gained a sitemap yet.** Applying 590 is the next action.

### Three things checked so they are not re-derived

- **The action is live in the running chassis** (v1.0.1333), binary-probed with controls: registry
  description PRESENT, invented string ABSENT, a known older action PRESENT. So the wiring needs no
  image roll — only the canonicalisation fix does.
- **`locked_at IS NULL` is load-bearing.** `adversecreditmortgage.co.uk` is locked by the owner
  HALT of 08-18. A sitemap commit **is** a deploy; the sweep must not drive one against a halt.
- **The other consumer of the stamp table was checked, not assumed** (owner ruling 07-29 §3):
  `site-discovery-staleness-check` CROSS JOINs its own three `DISCOVERY_AGENTS` and its findings
  loop iterates that same list, so a fourth `agent_type` in `site_discovery_rotation` cannot create
  a false finding there. `render-audit-agent` is already precedent for a non-discovery type.

Council round 2 submitted on the same correlation `8a004aab-be85-4d6d-bdb1-4fb114f1d64b`
(`RESUBMIT_CORR`), running at `review_reuse_agent` when last checked. Round 1's REVISE — *"a
registered-but-uncalled action reproduces the diagnosed defect in a new form"* — is what this round
answers.


### 2026-08-24 (c) — council round 2: **APPROVED**, 7 advisory objections, every one run down

Verdict at 14:26:40, correlation `8a004aab-be85-4d6d-bdb1-4fb114f1d64b`, *"approved with 7
advisory objection(s) — none high-severity"*. Approval is not a reason to skip the objections;
two of them were real and one is still open.

| seat | objection | resolution |
|---|---|---|
| editquality (med) | agent seeded without `description` cannot be spawned | **Sketch elision only.** The migration's column list (line 58) includes `description`. Verified against the landmine's own query. No change |
| editquality (med) | is `scripts/site-discovery-files.py` still driven? A second writer would race | **VERIFIED DEAD as a driver.** 0 `scheduled_tasks` reference it; no CronJob, no CI. Hand-run only, so no two-writer race. It still has the canonicalisation defect — recorded, out of scope |
| editquality (low) | pre_query alias `sid` undefined | Sketch shorthand. Real SQL is `SELECT s.id AS sid`. No change |
| **bug_historian (med)** | **`url_count=0` collapses "opted out" and "unexpectedly zero" into one silent no-op** | **ACCEPTED, NOT FIXED — see below** |
| reuse_agent (med) | did you look for an existing path normaliser before writing one? | **Right to ask, and the answer is MUST NOT REUSE** — see below |
| tooling_provenance (med) | the canonicalisation decision has no travelling-docs row | Partly addressed: `LANDMINES.md` (which syncs into `doc_notes`), register **SEO-007**, and this file |
| **guardian (med)** | **pre_query has no vm-sites/B2 distinction; grounding only showed B2 routing** | **VERIFIED SAFE, and the reason is now pinned in the migration** — see below |
| guardian (low) | a THIRD periodic git-committer against the shared repo | Named. RSS commit is the precedent at the same cadence class |
| debug_historian (med) | `sites.status` is informational; enumerate the distribution before scoping on it | **Enumerated 2026-08-24:** `deployed` 26, `pool` 17, `active` 2, `test` 2, `system` 1. `IN ('active','deployed')` selects exactly the **28** the before-figure was measured over, and `pool` (unbuilt domains) is correctly excluded. Matches all three existing rotations |
| prior_art (med) | the `DISCOVERY_AGENTS` safety claim needs a fresh grep, not a citation | **Fresh grep run:** `check.py:49-53`, exactly 3 hardcoded. Claim holds |

#### reuse_agent's objection is the most useful thing in the round, because the answer inverts it

`datahelpers.NormalizePagePath` (`links.go:215`) **does exist** and it **does** strip `index.html`.
Reusing it would have been wrong, and its own doc comment says why:

```
//	/tools/index.html    -> /tools
//	/index.html  /  ""   -> /
```

It strips the suffix from **section indexes too**, because it exists to make two hrefs
**comparable for equality** — the link checker needs `/guides/` and `/guides/index.html` to match.
That is the opposite of emitting a canonical URL. Calling it from `absoluteURL` produces **exactly
the mutation `TestRootIndexIsCanonicalisedButSectionIndexIsNot` catches**, and would have rewritten
all 228 section indexes against their own canonical tags.

**So the transferable point is not "reuse the helper" but "a normaliser for COMPARISON and a
normaliser for EMISSION are different functions that look identical".** The seat could not have
known that without reading it, and asking was correct.

#### guardian's vm-sites objection: safe, but only by an ABSENCE, so the absence is now load-bearing

4 of the 28 live sites are `vm-sites` (`idea.uk`, `noted.co.uk`, `relojistas.com`, `webdesign.uk`);
the other 24 have an empty `github_repo`. The rotation will hit all of them.

Routing is correct **by construction**: `git_commit` calls `resolveGitRepoNameDB`
(`git_deployer_actions.go:170` → `helpers.go:232`), which reads `sites.github_repo` per domain.
⚠ **But its FIRST branch is an explicit `config['repo_name']`** — so adding that key to the step
config would send those 4 sites to the wrong repo, and `LANDMINES.md` records exactly that failure
for the hand-run script which hardcodes it: **kcat exits 0, the adapter logs no error, GitHub shows
the commit, and the served file never changes.** The migration now carries a `DO NOT ADD
'repo_name'` comment at the step, because the correctness here lives in a key that ISN'T there,
which is invisible to anyone reading the config.

#### bug_historian's objection is REAL and is NOT fixed — the one thing this round leaves open

`check_has_urls` routes `url_count = 0` to `complete` for two opposite cases: the site opted out,
and the pages query returned nothing. Given sites carry **26–135** pages, the second is almost
certainly a defect, and nothing files a work item or error row. This is 016b §9's *"page build
completes having built nothing"* shape, inherited from `check_has_rss` which this deliberately
copies.

**One correction to the objection, stated because it changes the priority:** it says the site would
*"silently and permanently"* stop being maintained. **Not permanently** — the rotation re-selects
every 3 days regardless, so a site recovers as soon as its pages come back. The real cost is the
**absence of a signal**, not a stuck state.

**The fix, designed here so the next session does not re-derive it:** the action already
distinguishes the cases in its `reason` string, but routing config on a prose literal is its own
trap. Give `render_sitemap` a machine-readable `skip_reason` (`opted_out` | `no_listable_urls`),
then branch: `opted_out` → `complete`, `no_listable_urls` → record an error row before completing.
That is a Go change plus a follow-up migration, so it is a separate round — and it cannot be
applied before `590` anyway.


---

## 2026-08-25 — the fleet swept itself overnight: **8 → 26 of 28**, and two instrument faults

### The sweep, measured not assumed

`590` applied 2026-08-24 15:31. The rotation then ran **unattended for 13.2 hours**, 15:32:13 →
04:44:38, one site per 30-minute tick.

- **27 sites swept** (28 live minus `adversecreditmortgage.co.uk`, correctly held out by
  `locked_at`).
- **27 orchestrations, ALL `COMPLETED`.** Zero `FAILED`, zero partial.
- **Every stamp reconciles to `runs = 1`** — the check the previous handoff called for. **Zero
  dropped dispatches.** The ~300s post-roll dispatch dead-zone never bit, because the roll landed
  at 09:27 today, well after the last tick.
- Then it went quiet, correctly: with all 27 stamped inside the 3-day threshold, nothing is due
  until 2026-08-27. That is the designed steady state, not a stall.

### Fleet coverage, judged by body

| | 2026-08-24 14:00 | 2026-08-25 10:00 |
|---|---|---|
| serve a sitemap **of ours** | **8 of 28** | **26 of 28** |

Every one of the 26 scores a **perfect n/n** loc-to-`pages`-row match. The two that do not:

- `adversecreditmortgage.co.uk` — still the parking provider's 1-`<loc>` `/lander` file. **Correct:
  it is under the owner HALT and the rotation excludes it by design.**
- `webdesign.uk` — **302s every path** to `webdesign.co.uk`. It WAS swept (`runs=1`) and did commit
  a file to `vm-sites`, but the domain serves nothing of its own. This turned out to be a defect —
  see below.

### The canonicalisation fix proved itself, by accident, as a natural experiment

I did not have to construct a before/after. The chassis rolled **between the first and second
ticks** on 08-24 (~15:45, picking up `5c9acf1bd`), which split the sweep cleanly:

- `robot-hands.com`, swept **15:32:13, pre-roll** → homepage loc `https://robot-hands.com/index.html`
- **all 26 sites swept 16:02 onward, post-roll** → homepage loc `https://<domain>/`

One binary, one behaviour, either side of a known boundary. ⚠ **`robot-hands.com` is therefore the
one stale artefact.** Its stamp was cleared 2026-08-25 ~09:55 (the one-line remedy the previous
handoff documented) so it re-runs on the current binary; the task fires every 30 minutes and was
last at 09:49, so the correction lands ~10:19.

### INSTRUMENT FAULT 1 — my own verification query scored my own fix as a regression

The first census after the sweep read **25 of 28** and reported `apis.uk` as **NOT OURS** (0 of 1
locs matching). `apis.uk`'s sitemap was perfect.

Cause: canonicalisation makes the emitted homepage (`/`) differ from the stored `pages.url`
(`/index.html`) **on purpose**, and my match test predated the fix. Every site scored exactly
**n−1**; `apis.uk`, whose only deployed page IS the homepage, scored 0 of 1, i.e. total failure.

**The tell was uniformity.** One unmatched entry on 26 different domains is not 26 independent
faults — a real coverage problem is ragged. Canonicalising the `pages` side of the join fixed it:
**25 → 26**, every per-site score n/n. Written up as a landmine.

### INSTRUMENT FAULT 2 — and it found a real bug

`webdesign.uk` being the one swept site still serving nothing was the loose thread. Its sweep had
reported `url_count: 7, probe_dropped: 0` — a perfect result for a domain that redirects
everything away.

**`probeOK` has always carried the rule and never implemented it:** *"Only a 2xx qualifies. A
redirect is deliberately NOT listed."* The client was `&http.Client{Timeout: ...}`, and Go follows
up to 10 redirects transparently, so the probe saw the **200 at the end of the chain**.

Proven both ways, 2026-08-25: all 7 of `webdesign.uk`'s pages return **302 un-followed, 200
followed**. `probe_dropped: 0` — the field whose entire job is to report this — was the most
convincing part of the wrong answer.

⚠ **The doc comment made it worse, not better.** Anyone checking whether redirects were handled
found the rule stated clearly and correctly, beside code that did not implement it. That converts
"unhandled" into "handled" for every reviewer who reads it.

**Blast radius measured BEFORE writing the fix**, not left as a reviewer's risk: sampling up to 3
listed URLs per domain across all 27 swept domains, exactly **1 of 27** has any 3xx — and it is
`adversecreditmortgage.co.uk`, already excluded. **No live site loses a URL.** `webdesign.uk`
correctly drops 7 → 0, and the existing empty-sitemap refusal then stops the commit.

Fixed in `54ba65b25` (`CheckRedirect` → `http.ErrUseLastResponse`), pinned by
`TestProbeDoesNotFollowRedirects`, mutation-proven — the bare client yields *"reported status 200,
want 302"*, verbatim the live reading. Council `25157bab-4b6d-40c5-a218-98148b60daf6`.

**Same class as yesterday's canonicalisation defect:** the probe proves FETCHABILITY and the action
keeps mistaking that for CANONICALITY. Both were stated in the header; both were unimplemented.

### The wrong call: I trusted a measuring instrument I had not checked

I justified yesterday's binary probe with *"both controls discriminated correctly"*. A landmine
added the **same day** by another lane says exactly that does not establish it — BusyBox `grep`
over `/proc/1/exe` gives false absences **while both controls pass**.

The conclusion survived on luck: the failure mode is a false ABSENCE and my target read PRESENT
(and 27 live runs settled it anyway). I found out by accident — grepping chassis logs for `build
provenance` matched the landmine being synced into `doc_notes`.

**I applied "grep LANDMINES for the symbol you are about to trust" to the SUBJECT and not to the
TOOL.** Both faults this session were in instruments, neither in the thing being measured. Logged
in `WRONG_CALLS.md`.

### 2026-08-25 (b) — both fixes closed out

**`robot-hands.com` corrected.** Stamp cleared ~09:55, re-selected **10:19:48** on the first tick
after (the task's own interval is 30 minutes — a cleared stamp does NOT re-run within 30 seconds,
which cost me three impatient checks). Orchestration COMPLETED, `url_count` 35, `probe_dropped` 0.

At the artefact: `<loc>https://robot-hands.com/</loc>`, was `<loc>https://robot-hands.com/index.html</loc>`.
**35 locs preserved, 35/35 matching `pages` rows, and the file is exactly 10 bytes smaller — the
length of `index.html`.** Same URL, both directions, on live data.

**Fleet re-checked live at 10:22 — ZERO of 28 domains still emit a non-canonical homepage.**

**Redirect fix APPROVED**, correlation `25157bab-4b6d-40c5-a218-98148b60daf6`, 2026-08-25 10:16:29,
*"all reviewers approve"*, **zero objections** — unlike the wiring round, which drew 7 advisory ones.
Worth noting why: this submission carried the blast-radius census (1 of 27 domains affected, and
that one already excluded) **in the submission itself**, rather than listing it as a risk for the
reviewers to check. That was the exact criticism of round 1 of the wiring, and doing it up front
removed the whole category.

⚠ **`54ba65b25` is Go and therefore INERT until the next roll.** `webdesign.uk` still serves its
wrong sitemap until then. Verification recipe is §1 of the 08-25 handoff.


---

## 2026-08-25 (c) — chassis v1.0.1339, and the deferred objection turned out to be LIVE

Roll landed 19:07 (pods `669b45fdb4-*`). `54ba65b25` (the redirect fix) is in HEAD, so it is
carried.

### The fleet grew on its own, which is the mechanism working

**31 live sites**, up from 28. Three arrived since yesterday — `homegarden.uk`, `lampenkap.com`,
`cv1.co.uk` — and the rotation picked all three up without anyone doing anything. That is the whole
point of `590`: a new site gets a sitemap because it exists, not because someone remembered.

**Census 2026-08-25 19:35: 27 of 31 serve a sitemap of ours**, every one a perfect n/n match.

### ⚠ MY OWN RECONCILIATION QUERY NOW PRODUCES FALSE POSITIVES — fix it before using it

The `runs = 0` query I wrote into the 08-25 handoff reported **six** sites as dropped dispatches
today. **All six had run.** `orchestration_states` retains COMPLETED/FAILED for only ~24h
(database-cleanup step 3), and their rows had been reaped: oldest surviving is 2026-08-24 19:05
against stamps going back to 15:32.

**And the remedy I documented is destructive when fed this.** "Clear the stamp so it re-runs"
against six healthy sites re-probes ~150 URLs for nothing and reads as "the sweep is broken".

**The query is only meaningful for stamps NEWER than the retention window.** Corrected form in the
handoff. Same shape as MEMORY [[a-closer-census-cannot-see-what-it-succeeded-at]] — a rolling
window makes success indistinguishable from never-happened.

### ⚠ AND A SECOND BAD QUERY, caught BEFORE it reached the handoff

Hunting for other affected sites I wrote `count(*) FILTER (WHERE p.deployed_at > r.last_selected_at)`
and got **24 of 24 domains**, most with large counts — which would have read as fleet-wide
breakage. **`pages.deployed_at` is UPDATED on every redeploy, not set once at first deploy**
(dartsonline.com pages `created_at` 2026-07-06, `deployed_at` 2026-08-24/25). So it measured
ordinary rerender churn.

The reliable detector for this defect needs the **artefact**, not the DB: *stamped, has pages now,
and serves no sitemap of ours*. That census found exactly two.

### THE REAL FINDING — the bug_historian's objection is live, and worse than it was framed

`homegarden.uk` and `cv1.co.uk` both swept successfully and both serve **404**. Not a wiring fault
— they were selected **before their pages deployed**:

| site | stamped | first page deployed | pages deployed after |
|---|---|---|---|
| `homegarden.uk` | 10:50:18 | 12:47:09 (**~2 h later**) | 20 of 20 |
| `cv1.co.uk` | 12:21:48 | 13:47:04 (**~1.5 h later**) | 3 of 3 |

Both sweeps did exactly the right thing — `candidate_count: 0`, `rendered: false`, refused to
publish an empty sitemap, no commit. **The selector was the only thing wrong.** The stamp had
already been written, so both fully-built live sites would have sat sitemap-less until **2026-08-28**
and nothing would have said so.

**The objection framed the cost as a missing signal. The live cost is the CONSUMED SLOT** — a log
line would not have helped either site; not being selected would have. That reframing is the useful
part, and it changed the fix: it belongs in the **selector**, not in observability.

Both stamps cleared by hand. Fixed for good in **migration `622`** (applied 2026-08-25 19:45):
an `EXISTS` guard so a site with no deployed active page is never selected and never stamped.

⚠ **The guard is DELIBERATELY WEAKER than `render_sitemap`'s own filter, and `622`'s verify block
RAISES if anyone tightens it** (it fails on `pg.noindex` or `pg.expires_at` appearing). Reasoning:
the action also excludes noindex and expired; testing only `status` + `deployed_at` is a strict
superset, so drift can only make the selector pick a site the action then finds nothing for — which
is today's behaviour, no worse. A mirrored guard would need permanent lockstep and its drift mode
is a site **silently never selected**, strictly worse than the bug. **A guard that cannot fail in
the dangerous direction beats one that is currently exact.**

⚠ Renumbered **591 → 622** mid-session: another lane took 591 while this was being written and the
tree had moved to 621. `ls` the directory immediately before naming a migration, not when you start.

Residue, named not hidden: a site whose pages are **all** noindex or **all** expired still burns its
slot silently. Needs the machine-readable `skip_reason` (Go). Population today: **zero**.

Council `c88f5c0f-cca2-4753-bd6c-9fabc93b100e` submitted for `622`.

### 2026-08-25 (d) — the redirect fix PROVEN at the artefact

`webdesign.uk` re-ran 19:59:02 on chassis v1.0.1339. Same 7 candidates as the pre-fix run; now
`url_count` **0**, `probe_dropped` **7**, `rendered` **false**, `reason` *"no listable URLs —
refusing to publish an empty sitemap"*, and **no `sitemap_commit_result` key** — the conditional
routed to `complete` without committing.

Before (2026-08-24 20:06): `url_count 7, probe_dropped 0`, committed to `vm-sites`.

The probe now sees the 302s it was always documented to refuse. **A domain that redirects every
path away correctly ends with no sitemap of its own**, so `webdesign.uk` stays uncovered by design
rather than as a gap.

Still unproven when this was written: `622`'s selector guard, whose two test sites (`homegarden.uk`,
`cv1.co.uk`) were queued behind `webdesign.uk` at ~20:28 and ~20:58. Recipe in handoff §1b.

### 2026-08-25 (e) — `homegarden.uk` recovered, and a precision note on what that does NOT prove

Re-selected 20:29:30. `candidate_count` 20, `url_count` 20, `probe_dropped` 0, committed. At the
artefact: **HTTP 404 → HTTP 200**, `application/xml`, 2,237 B, **20 locs, 20/20 matching** its
`pages` rows, homepage listed as `/` (so `5c9acf1bd` applies to newly-swept sites too).

⚠ **This proves the manual REMEDY, not the `622` GUARD.** `homegarden.uk` now has deployed pages, so
the guard is not even consulted for it. What recovering demonstrates is that clearing a stamp
returns a site to the front of the queue and it renders normally — which is the documented repair,
worth knowing works, and a different claim.

`622`'s guard is presently evidenced by its three induced verify failures and the live `pre_query`
text. Its BEHAVIOURAL proof needs a site with zero deployed active pages at selection time — the
next newly-seeded site. The falsifiable form is in handoff §1b: **a site with `deployed_pages = 0`
and a non-null rotation stamp means the guard did not hold.** Stating it that way because "the
sweep looks fine" would not distinguish a working guard from an unconsulted one.

`cv1.co.uk` still 404 at 20:30, queued ~20:59.

### 2026-08-25 (f) — `622` APPROVED, and its one objection was worth checking

`c88f5c0f-cca2-4753-bd6c-9fabc93b100e`, approved 20:31, **1 advisory objection, low**.

`editquality`: the anchor is a bare string and **`replace()` is GLOBAL**, so if
`AND s.locked_at IS NULL` occurred twice in one `pre_query` both would silently be guarded — and my
pre-flight counted matching **ROWS**, which says nothing about occurrences **within** a row. That is
a fair criticism of the migration as written.

**Checked 20:32: it did not fire.** The live row has exactly one `locked_at` clause, one guard, one
`-- 622:` comment (counted by `length(x)-length(replace(x,needle,''))` over the needle length). So
the applied state is correct and nothing needs undoing.

**The lesson for the next anchored migration**, and it is cheap: assert the POST-state occurrence
count, not the pre-state row count. One extra line in the verify block would have made the objection
unnecessary — and unlike the row count, an occurrence count cannot be satisfied by a partially
matching row.

### 2026-08-25 (g) — all three verifications closed. **29 of 31.**

| site | render | artefact | verdict |
|---|---|---|---|
| `webdesign.uk` | `rendered:false, url_count:0, candidates:7, dropped:7`, no commit | HTTP 302, 0 locs | **correct — no sitemap is the right end state** |
| `homegarden.uk` | `true, 20, 20, 0`, committed | HTTP 200, 20 locs, **20/20** | recovered from 404 |
| `cv1.co.uk` | `true, 3, 3, 0`, committed | HTTP 200, 3 locs, **3/3** | recovered from 404 |

**Final census 21:02: 29 of 31 serve a sitemap of ours.** The two that do not are
`adversecreditmortgage.co.uk` (owner HALT, excluded by `locked_at`) and `webdesign.uk`
(redirect-only). **So every site that SHOULD have a sitemap has one** — the close condition from the
08-24 handoff is met in full, against a starting figure of 8 of 28.

⚠ **`622`'s guard still has not fired** and the handoff says so in its own section (§1d). Both
recoveries came from hand-cleared stamps on sites that now HAVE pages, so the guard was never
consulted. Recording it separately because "29 of 31 and everything green" is exactly the kind of
result that would let an unproven guard ride along unnoticed.


## 2026-08-26 — the sitemap follows the deploy (`642`), the generator's last non-canonical writer fixed, and a line of mine swept into another lane's commit

### (a) Morning state check — §1d of the 08-25b handoff

`622`'s guard is **still unconsulted**: all 31 live unlocked sites have ≥1 deployed active page
(smallest: `apis.uk` 1, `lampenkap.com` 1) [MEASURED 2026-08-26 ~09:30 BST, the §1d query]. Nothing
to conclude in either direction; the falsifiable form stands. Rotation last ticked 08:10Z with
nothing due — correct, every stamp is inside its 3-day window.

⚠ The migrations dir had moved **634 → 641** between the handoff and my naming a file; `635` was
already taken. `ls` immediately before naming, again — this lane's third collision-avoided.

### (b) The deploy-path half of SEO-002, reframed before it was built

`590` deferred it with: *"the deploy path would re-probe a whole site on every page change (135
requests for webdesign.co.uk, every time)"*. **Right about the wrong design.** Firing a render per
edit is expensive; making an edited site *due* costs nothing new, because the rotation already has a
hard cost cap — one site per 1800s tick, so ≤48 renders/day whatever the due set says. The event
belongs in the SELECTOR, where the cap already lives. No Go, no roll, in council scope as a migration.

What I measured before writing it [all 2026-08-26]:

- **Churn is waves, not a rate.** Distinct sites with ≥1 `pages.updated_at` per day: 08-19 **2**,
  08-20 **2**, 08-21 **1**, 08-22 **0**, 08-23 **3** — then 08-25 **27**, 08-26 (to 09:30 BST) **21**.
  All 31 sites had ≥1 page changed since their last render (wave backlog). So a bare
  "changed ⇒ due" clause saturates during a wave (fine — the tick drains 31 in ~15.5h) and needs a
  **quiet period** so a site renders once after the wave, not mid-wave and again after.
- **`updated_at` alone is the signal.** Every `UPDATE pages` statement in the Go tree bumps
  `updated_at = NOW()` alongside whatever it changes — deploy stamp (`deploy_evidence.go:294`),
  retraction, `needs_rebuild`, section stores; **12** call sites as of 2026-08-26 — and **0** rows
  fleet-wide have `updated_at < deployed_at`. So one column covers deploy, redeploy, retraction,
  noindex flip and expiry-set, and the pre_query names no visibility column (`622`'s verify RAISES
  on `pg.noindex`/`pg.expires_at` — the mirror that drifts). The only trigger on `pages` is
  nav-cache invalidation; the bump is convention in every statement, not a trigger. [INFERRED,
  no instance today] a NEW writer that forgets the bump would be invisible to `642`'s early
  branch — the 3-day floor catches it.
- **The floor must stay unconditional.** If the quiet test gated both arms, a permanently-busy
  site would be silently never selected — the exact drift mode `622` refused. So:
  `(floor) OR (changed-since-render AND quiet 30 min)`.
- **Rejected, not deferred:** "probe only the changed URL and merge" (the handoff's other option).
  Go plus reading the served artefact as input, to save ~26 GETs per render. Wrong trade here.
- **Accepted costs:** over-trigger on pre-deploy edits (mid-build is excluded by the claimed-build
  guard); empty-diff commits when two same-day renders produce identical bytes — the git adapter's
  no-op skip covers deletions only (`github_client.go:250`), so identical content pushes an
  empty-diff commit that SUCCEEDS (no FAILED runs, just repo noise).
- Scheduler: a pre_query returning zero rows is a clean no-op that stamps and rotates
  (`cmd/scheduler/main.go`, the `bugs_open/048` comment) — a quiet tick costs nothing.

### (c) `642` applied and PROVEN the same day

Applied ~10:30Z, scoped via `MIGRATIONS_DIR` to a scratch dir holding only my file — the canonical
dir's pending set was **614–638**, all other lanes' (several `_HOLD`). Apply NOTICE:
**`0 by age, 20 by change-and-quiet`** — down from 28 at design time an hour earlier: 8 sites had
gone busy again as the wave continued. That is the quiet gate holding, seen through the count.

**The proof was built into the apply moment.** Every stamp was 08-24/25, so under the OLD rule
**nothing** was age-due before **2026-08-27 16:02:34** (`loancalculator.co.uk`, oldest stamp + 3d).
Any selection before that instant can only be the new branch. Then:

| tick (UTC) | site | url_count | dropped | committed | would have been age-due |
|---|---|---|---|---|---|
| 10:42:35 | `loancalculator.co.uk` | 28 | 0 | yes | 08-27 16:02 |
| 11:17:09 | `agritec.uk` | 13 | 0 | yes | 08-27 16:33 |
| 11:47:35 | `finetuning.uk` | 51 | 0 | yes | 08-27 17:34 |
| 12:17:56 | `apis.uk` | 1 | 0 | yes | 08-27 19:35 |
| 12:48:26 | `leopardessconsulting.co.uk` | 37 | 0 | yes | 08-27 21:07 |
| 13:18:56 | `cookly.uk` | 15 | 0 | yes | 08-27 17:03 |
| 13:49:00 | `garden-tools.uk` | 14 | 0 | yes | 08-27 18:04 |
| 14:19:30 | `loanzy.uk` | 25 | 0 | yes | 08-27 21:37 |

**8 ticks, 8 distinct sites, 8 COMPLETED orchestrations, zero dropped, zero duplicates**, each a
day-plus ahead of its floor. At the artefact for the first: `https://loancalculator.co.uk/sitemap.xml`
HTTP 200 `application/xml` 3,560 B, **28 `<loc>` = 28 listable rows**, homepage as `/`, and the served
body carries the two **`2026-08-26`** lastmods (`legal.html`, `tools/credit-health-check.html`) —
exactly the pages whose edits made the site due. Change → early selection → render → commit → served.
At 14:20Z: 0 age-due, **14** due-and-quiet, **11** changed-but-busy — draining at the tick rate.

Precision on what this does and does not prove: it proves the EARLY branch selects, and that what it
selects renders and publishes correctly. It does **not** directly show the quiet gate DEFERRING a
site — that is only visible in the due counts moving (28 → 20 → 14/11 at three instants), not in a
selection that was withheld. Good enough; stating it so nobody reads the table as more than it is.

### (d) Council `6e448adb-1e03-4e2e-a3dd-42bc6857ff24` — APPROVED 10:35:39Z, round 1, **10 minutes** after submission

Not 29. Two advisories, both run down:

- **`editquality` (medium):** *"the pre-flight only counts occurrences of the anchor — it never
  asserts that exactly one scheduled_tasks row carries that name"* (the byte-copied
  `build-pipeline-trigger` sibling precedent). **It does — the FIRST `DO` block raises unless
  `count(*) = 1` by name.** The reviewer judged from the plan SUMMARY, which named the
  occurrence-count assertion and not the row-count one. Checked live anyway: **1** row named; no
  other row targets `sitemap-refresh`; none carries the new clause. Did not fire.
  **Lesson: the submission summary must list EVERY pre-flight assertion.** A reviewer cannot open
  the file, so an unlisted guard is an absent guard to them — and a REVISE round on it would have
  been pure waste. Cheap to do; I did not, and drew a medium for it.
- **`debug_historian` (low):** concurrent ticks could double-select a site under the early branch
  before the stamp lands (the 583/584 single-flight family). Pre-existing and untouched by `642`
  (the stamp is a CTE in the same statement; `642` only widens the due set). `kafka-scheduler` runs
  **1 replica** [MEASURED 2026-08-26], so the loop is sequential; today's 8 ticks produced 8 distinct
  sites and 8 orchestrations. Noted, not acted on.
- 6 seats abstained; `bug_historian`, `reuse_agent`, `guardian`, `improvement_guardian`,
  `render_guardian`, `constitution`, `mission`, `prior_art_librarian`, `architecture` approve.
  `architecture`: *"point_fix … exactly the shape 590 deferred toward"*.

`107327c6b` carries `Council-Submitted:`; 098 credits it at report time, no amend.

### (e) `scripts/site-discovery-files.py` — the last writer of the non-canonical homepage, fixed (`bcb4645ff`)

Whole-path rule (`/index.html` → `/`) at the one point all three consumers read from
(`live_pages`), mirroring the action's `absoluteURL`. Population among listable rows
[2026-08-26]: **30** root `/index.html`, **261** section `index.html` (must pass untouched), 507
other, **0** missing a leading slash, **0** absolute — so the Go's defensive slash handling was not
needed here. Exercised against `cv1.co.uk` into scratch: root emitted as `/`, both section indexes
untouched, 3/3 probed 200, `llms.txt` picks up `/` too.

**Side observation, not this lane's item:** the script's rule-3 detection fired on `cv1.co.uk` —
Cloudflare's managed `robots.txt` block is merged in, currently disallowing ClaudeBot, GPTBot,
Google-Extended, CCBot, Amazonbot, Applebot-Extended, Bytespider, meta-externalagent and
CloudflareBrowserRenderingCrawler. Whether the estate wants AI crawlers blocked is the owner's call;
flagged in the handoff.

### (f) A line of mine was swept — and that is the documented system, working

My in-place LANDMINES correction (the canonicality entry's source line: "the script instance is
fixed too") reached HEAD in **`b3bddba60`** (imagery lane, 12:19 BST) as a same-file passenger.
Nothing lost, forward-only holds, message says nothing about it — exactly the shape CLAUDE.md
describes. Consequence I chose deliberately: **I did NOT run `landmines-verify-dispatch.sh`.**
Another lane (`dispatch_throughput`) has a NEW uncommitted entry in the file right now, and my sync
would consume its new-entry status so the verifier never checks theirs — the trap CLAUDE.md names.
My change is a status correction to an existing entry, not a new trap; the `doc_notes` copy lags
until the next lane syncs. Recorded so the lag is not read as an omission.

The session process restarted mid-task (~12:30 BST); one background verdict lookup was lost and
re-run. Nothing else was affected — every write had landed.

### (g) Register and handoff

`seo.md`: SEO-007's caller bullet carries the dated CLOSED note with the proof; verify-later
re-cut (deploy-path done, 3-days reframed as a re-probe floor, `622` guard + `skip_reason` residue
still open); SEO-002's line-132 sentence struck with the fix. New handoff
`HANDOFF_2026-08-26_continue_here.md` supersedes 08-25b.

### (h) 18:42Z re-check — `642` draining cleanly; `622` still unconsulted; idea.uk's dropped=1 explained

§1a of the 08-26 handoff, run at 18:42Z: **16** early selections since apply (10:42Z–18:26Z),
all COMPLETED, ~30-min cadence, 15 of 16 with `probe_dropped=0`. Due set **0 / 15 / 10**
(age / change-quiet / change-busy) — change-quiet holding ~14–15 rather than falling, because
the pool refills as it drains (10 sites had page writes inside the last 30 min). Rotation
provably ticking (18:26Z selection), so the stalled-rotation branch is excluded by direct
evidence, not by the counter.

**idea.uk `probe_dropped=1` is by design and PERMANENT — do not re-chase it.**
`dropped_sample = ["/privacy.html (301)"]`. That is the idea_uk_vm_site lane's decided legal-page
collision (their RUNBOOK, decision RESOLVED 2026-07-18, proven live 08-02): the VM's tool app
serves `/privacy` and nginx 301s the static `/privacy.html` to it. idea.uk has exactly ONE
`pages` row among the legal trio (`/privacy.html`, active, noindex=f; no `/terms*` or `/refund*`
rows), so every idea.uk render will be candidate 23 → kept 22, dropped 1. Consequence: the
privacy page is absent from idea.uk's sitemap — correct, since the recorded URL redirects and
the `/privacy` that 200s is the tool's page, not this row. Their lane's serving decision; not a
sitemap defect; nothing filed.

§1b at the same instant: **no** `deployed_pages=0` row carries a stamp — `622`'s guard still
never violated and still never consulted (min `apis.uk` at 1). The "31 sites" in this file's
morning entry vs 30 rows now is arithmetic, not a departure: `adversecreditmortgage.co.uk` is
locked (since 08-18) and the §1b query excludes it via `locked_at IS NULL`; 30 + 1 = 31.

### (i) The remake programme starts — first brief fired at `advertise.co.uk` (18:48Z)

The sitemap arc is closed and §1's checks were green, so the session moved to the largest §3
item: the 22 hosted-site remakes (`DECISION_2026-08-20_remake_the_hosted_sites.md`). The 08-21
precondition — the component-name collision — is GONE (`bugs_open/311` closed 08-24, verified at
the artefact by that session), so nothing technical gates the small end any more.

**First pick: `advertise.co.uk`**, per the decision's §5 ordering (single-pagers with strong
generic names, insurance last). Checked before firing: no register row (like `indoorplanters`,
the FIRST proven brief — the register hand-off is conditional, not required), no sites row, no
open work items, chassis pods ~5h old (300s spawn rule clear), no other session on the remakes
(commit sweep since 08-25: idea.uk / tool-rebuild / 404 lanes only).

- sites row created per the exact `buytoletcalculator.uk` precedent: `status='test'`,
  **LOCKED**, `locked_by` naming this decision — `d991a5b8-428f-44c1-b3eb-e50f44326fd9`.
- `scripts/fire-brief-writer.sh advertise.co.uk` with a short direction carrying the two facts
  research cannot find: it is a remake of a hosted single-pager (owner ruling 2026-08-20), and
  the estate-neighbour boundary the missing register row would have supplied
  (websitepromotion.co.uk / seotools.co.uk / webdesign.uk are not its ground).
- **PUBLISHED, receipt asserted** — `CORR c89cd031-728c-409c-9f85-5a880f42a727`,
  `ORCH 17af3a16-83c3-4f12-92cb-67b81e59a8bf`.

**[IN FLIGHT as of 18:49Z — orchestration live at `scrape_pages` within a minute of publish]** — a receipt proves the broker took it, not that the work happened.
Verify at the durable record (orchestration status → `site_specs` `mission_brief` →
`needs_brief_review` item held at `needs_human_review`). Outcome to be appended below; if this
entry ends here, the session died mid-poll and the three queries above are the pickup.

Also noted while checking the queue: the 08-20/21 test briefs (`indoorplanters.co.uk`,
`buytoletcalculator.uk`) still sit at `needs_human_review`, unreviewed — the gate provably
holds for days. `advertise.co.uk` is the first PRODUCTION brief behind them; the owner now has
three to read, and only this one asks for a release decision.

### (j) `advertise.co.uk` before-snapshot saved; the "single-pager" call corrected in mechanism, upheld in consequence

Saved before anything can overwrite it (the vinrose precedent):
`salvage/advertise.co.uk/index.html` — 20,453 B, verified real HTML with `file`, title
"Advertise". **The DECISION doc's "single-pager" classification is wrong in mechanism**: it is a
Drupal 7 instance with internal `?q=node/N` pages (the census keyed on "no internal links from
the homepage", and the links are there). **But right in consequence**: probed `?q=node/79` —
"Adweek news", 543 visible chars, an RSS-feed aggregation stub (Adweek / RealWire / WebProNews
feeds); the node pages are syndicated headline lists with no original content. Nothing further
worth salvaging; the homepage snapshot is the record. No change to the remake decision.

### (k) The advertise.co.uk brief: COMPLETED 18:51Z, verified at all three durable records

Run time ~3 min from publish (18:48Z → `scrape_pages` 18:49 → `write_brief` 18:50 →
`complete|COMPLETED` 18:51). Verified at the artefact, not the status:

- `site_specs` `mission_brief` is_current, **15,915 bytes, 13 top-level keys** (proposition,
  audience, reader_intent, content_plan, tool_opportunities, directory_opportunity,
  differentiation, stance, must_nots, regulated_subject, research_quality, confidence 0.78,
  open_questions) — structurally complete, not a truncation fragment.
- `needs_brief_review` item held at `needs_human_review`. Site row stays LOCKED.

**The direction landed.** The differentiation section names and stays off all three estate
neighbours: *"deliberately separated from websitepromotion.co.uk (promoting websites
specifically), seotools.co.uk (SEO as a discipline) and webdesign.uk (design)"*. Proposition:
plain-English reference on advertising ITSELF — what it is, how it works, who pays — explicitly
"not a service that does advertising"; must_nots refuse agency/media-buyer presentation, ASA/
CAP/Ofcom/ICO/CMA/FCA implication, results promises, and un-disclaimed tool advice.

**Its open_questions do the register's missing work**: Q1 says outright that the portfolio
position is unrecorded and asks the owner to confirm broad-reference vs lead-gen; Q2 asks
whether advertise.co.uk is the HUB of the marketing-adjacent cluster — **the answer shapes the
websitepromotion/seotools/webdesign briefs, which is why no further briefs were fired tonight**;
Q3 monetisation; Q4 UK-only confirmation; Q5 news-stream scope.

Third production-shaped run of the machine, third hold that held. The owner's queue is now three:
indoorplanters (test, 08-20), buytoletcalculator (test, 08-21), **advertise.co.uk (real, 08-26)**.

### (l) The lendzy acceptance marker: the 08-05 debt had already come due — found, source-fixed, filed as `bugs_open/414`

> **CORRECTED 2026-08-27 (by the 414 fixing session; refutation marked in the bug file):** "Source
> FIXED live" and "regeneration can no longer re-plant" were FALSE — `domain-strategist` had
> PARAPHRASED the instruction into the current `strategy` aspect on 08-12, a surface build-site-planner
> and webdesign-agent read, so the mandate stayed live for ten more days. My own census RETURNED that
> row and I cleared it on its FIRST matching window (the benign slogan) — and I censused the
> instruction's vocabulary, not the payload phrase. Full anatomy: WRONG_CALLS 2026-08-27
> (two tallies). Fleet payload-census 0 across ALL aspects as of 08-27 (their strip, row `0326a892`,
> re-verified by this lane at the live DB). Their §7 also corrects my population: 14 archived
> component rows, the guide re-emitting on FOUR regenerations — the census I published was
> current-state only.

Found while correcting this lane's stale memory-index entry: its landmine line ("lendzy's seeded
content_direction still carries the acceptance_marker instruction — strip before any real use")
was still true, and "real use" happened weeks ago — lendzy.co.uk serves 19 pages.

- The planted phrase **"checked against the FCA handbook, rule by rule"** is in 3 components'
  `content_data` and **SERVED**: `/about.html` ×2, the affordability-complaint-checker guide ×1
  [MEASURED 2026-08-26, curl by body]. An unverifiable compliance claim on a finance site.
- **Worse: an open `content_rewrite` item (needs_human_review) canonised it** — *"The site's
  core differentiator — FCA-rule-level accuracy checked guide by guide"* — the audit fleet
  adopted the tripwire as identity and queued work to reinforce it.
- Fleet census [2026-08-26]: **1 site** carries a marker; apis.uk/webdesign.co.uk "exact phrase"
  hits are innocent; lendzy's second mandated phrase ("know the rules before you borrow") is a
  benign slogan, left in place.
- **Source FIXED live**: `content_direction` revised — current row `81ddcc40-…`, the 08-02 row
  minus `positioning.acceptance_marker` + the `formatted` tail line, applied server-side under
  a guard asserting the exact tail; history preserved (`61ef7033-…` superseded, residue intact).
- **Copy repair OPEN** — and the trap is that the queued `page_rerender: about` canNOT fix it
  (rerender regenerates from `content_data`, where the phrase lives). Needs a content rewrite of
  the 3 components; the held "differentiator" item must be rejected/rewritten against the bug.
- Filed: `bugs_open/414_…` (evidence, census, repair path, why no 090 — verbatim string identity
  at every hop). 016b: §9 pattern ("an experiment's tripwire ships with the spec — and the audit
  fleet canonises it") + §10 index entry. No 090: owner-ruling-07-31 escape hatch stated in file.

## 2026-08-31 — re-ground after four quiet days + the fresh chassis roll; one live blemish found and remedied

Session resumed on the owner's ask (fresh chassis deployed ~08-30 evening — pods 12h old at
08:4xZ; provenance line already scrolled, as expected on a busy service). Nothing of THIS lane's
was inert-until-roll (all our changes were DB config + docs); the roll matters here because it is
what `bugs_open/414`'s class fix was waiting for (theirs to verify, §7j recipe).

**(a) Readings at 08:3x–08:4xZ:**
- **642 rotation ticking on the new build**: latest selection 08:12Z (fundamentallyai, 26 urls,
  0 dropped); 5 of the last 6 runs COMPLETED. Due set **1 / 2 / 5** — small numbers, mechanism
  keeping up; gaps between ticks are normal when nothing is due.
- **622 guard**: still unviolated, still unconsulted (min `apis.uk` at 1 deployed page; no
  zero-with-stamp rows; several sites born 08-22→08-25 all entered with pages).
- **Brief queue UNCHANGED since 08-26**: all three (`indoorplanters`, `buytoletcalculator`,
  `advertise.co.uk`) still `needs_human_review`. The remake programme is paused at the human
  gate by design — the owner reading + releasing is the only path forward.
- **lendzy served bodies CLEAN** — 0 occurrences of the planted phrase on /about.html and the
  guide (my own curl, both 200). Consistent with 414 §7m/n; the bug stays theirs to close
  post-roll.

**(b) `oufe.com` serves 404 on `/sitemap.xml` — found, remedied, verification owed.**
This morning's 03:38Z selection FAILED at `commit_sitemap` (git-adapter `TIMEOUT`, 03:40Z);
render was fine (19 urls = 19 deployed pages, 0 dropped). The served artefact decides:
`/sitemap.xml` = **404**, so the commit decisively did NOT land (— NOT the §7k
"failed-but-landed" class; I checked precisely because `sitemap_commit_result:ok` sat in
collected_data. ⚠ that key's presence is not a landed commit). Later runs 04:09→08:12 all
COMPLETED, so the timeout was transient. **The trap that turns one transient failure into a
3-day outage: selection STAMPS the rotation before the outcome**, so the failed run consumed
oufe's slot (next natural retry: page-change or the 09-03 floor). Remedy, mechanism-native:
backdated the stamp guarded on the exact failed-run value (`now()-'4 days'`; ⚠ the column is
NOT NULL — you cannot clear it, only backdate). oufe is now age-due; expect re-selection within
~1–2 ticks of 08:45Z. **Verify at the served body (19 locs), not at the run status.**
Residual worth watching, not yet filing: **a FAILED run consumes the slot** — 1 observed
instance in ~100+ runs since 08-26; if it recurs, file it with the retry-design question.

**(c) oufe VERIFIED HEALED at 09:46Z** — re-selected 09:44:25Z (the backdate worked; the selector
took oufe ahead of gaswholesalers on stamp age), and the served body is the proof: `/sitemap.xml`
**200, 2,262 B, 19 `<loc>` = 19 deployed pages**. 404 → n/n inside an hour of finding it.

**(d) Owner comments on the briefs → `DECISION_2026-08-31_best_in_vertical_fullness_and_the_advertise_marketplace.md`.**
Four rulings recorded: best-in-vertical standing fleet-wide (the MECHANISM is the copy lane's
unshipped `PLAN_2026-08-25_best_in_class_propagation.md` — do NOT duplicate it here; manual
control in the fire direction until it ships) · never differentiate by omission, copy-but-improve
(assessed: covered in practice in the held briefs, not guaranteed by mechanism) · advertise.co.uk
future = selling ad space on OUR network, preferably directly — this ANSWERS the brief's Q3;
compatibility analysed element-by-element: nothing forecloses it, three lines amend at pivot
(proposition's "not a service", the broker must_not, "it does not sell it"), one guard NOW
(no "we sell nothing" load-bearing served copy — the 414 lesson) · sites should be QUITE FULL —
measured: the brief is 16 items / 8 kinds ≈ 30–40+ pages via fan-outs, but glossary + news sit
in `aspirational`; release edit prepared (two promotions + explicit fan-out counts), NOT applied.
**Q2 (cluster hub) is still the open owner question — Q3 is answered, Q2 is not.**

## 2026-09-02 — three owner rulings executed: the edit applied, the cluster gate lifted, three briefs fired

Owner's follow-up (recorded in the DECISION file's EXECUTED addendum): negative-identity claims
default OUT of copy fleet-wide (legal/finance compliance the exception) · apply the fullness
edit · **Q2 ANSWERED — advertise.co.uk is the flagship of the marketing cluster** ("advertise.uk"
in the message is nowhere in the portfolio or `sites` — read as a slip, flagged back).

- **Brief revision `5dac12fd` applied** (original `c9210c3e` preserved; first attempt aborted on
  a syntax error and ROLLED BACK — verified single-current before retrying with a CTE form).
  Guards asserted every index before edit. Contents per the DECISION addendum. Review file
  re-rendered. **Build still HELD** — edit ≠ release; the owner has not said "go build".
- **Cluster briefs fired** (~10:0xZ): websitepromotion `a6fae8ee` · seotools `9ca54346` ·
  designblog `d8eb90be`. Sites rows `test`+LOCKED. Directions carry flagship deference +
  per-domain vertical + best-in-vertical/no-omission + fullness + no negative-identity copy.
  Before-snapshots: `salvage/<domain>/index.html` (8,058 B / 7,741 B / 39,286 B, all real HTML).
  [IN FLIGHT — outcome below; if absent, the three verification queries in
  `scripts/fire-brief-writer.sh`'s output are the pickup.]
- **CONTRIB to copy_quality_two_stage**: the negative-identity ruling handed to the lane that
  owns copy doctrine + the propagation carrier. We do NOT build fleet copy machinery here.

**Outcome (closes the in-flight marker above): all three cluster briefs COMPLETED and verified by
11:52Z** — 13 keys each, 15/18/20 plan items (the fullness direction visibly landed: all at or
above the first brief's 16), confidence 0.78/0.78/0.87, all three held at `needs_human_review`,
all three carry the negative-identity rule and name the flagship (websitepromotion in
differentiation; the other two elsewhere — their verticals sit further from the flagship's
ground, review-gate judgement). Review renders: `BRIEF_2026-09-02_<domain>_for_review.md` ×3.
The owner's queue: SIX briefs (3 real, 2 test, advertise edited-awaiting-"go build").

## 2026-09-02 (later) — "go build advertise": the first remake RELEASED, build running

Owner's message: *"I own advertise.co.uk not advertise.uk, sorry. go build advertise."* — the
domain confirmation closes the 09-02 slip question exactly as we read it (advertise.co.uk; the
lane had flagged it back), and "go build" is the release word §1d was waiting on.

**Release executed 12:13–12:14Z, per the review item's own `how_to_release`** (spec of work item
`518ed780`: "Create the needs_domain_research item for this site (handler
domain-research-classifier)…" — no held research item existed, verified before acting):

1. `needs_brief_review` `518ed780` → `complete`, `approved_by='owner'`, result records the
   release, the brief revision (`5dac12fd`, still `is_current`, owner original `c9210c3e`
   underneath — re-verified before release) and the owner instruction verbatim.
2. Inserted `needs_domain_research`, `item_key='research_advertise.co.uk'`, status `triaged`,
   handler `domain-research-classifier`, priority 5 / severity high / pipeline build — the
   domain-submitter Flow A shape (matched against `research_farmerinsurance.uk`), source
   `brief-release`, spec carries domain + brief revision. Dedup index checked first: key free.
3. **The deliberate release act, separate transaction, LAST**: site row `d991a5b8` off
   `test`+LOCKED → `status='active'` (upsertSite's value, pilot-seed precedent), `locked_at`
   NULL — guarded on the exact prior values. Work items first, unlock last, so the moment of
   release was a single act on an otherwise-prepared board.

**Dispatch verified at every hop, not assumed**: selector-eligibility dry-run (the live
`find_dispatchable_site` CTE, expanded) showed advertise as the ONLY eligible site fleet-wide;
`build-pipeline-trigger` (30s cadence) fired 12:14:58Z; item CLAIMED by `build-dispatch-loop`
12:15:21Z — first tick after release; handler orchestration `e44a44d7-682d-4a14-840d-74c34c6aa78b`
EXECUTING `read_site_specs` for advertise.co.uk at 12:15:2xZ. Flow A read-and-extends: the
classifier is reading the specs that include the owner-edited mission_brief.

Watch-points for the end-to-end (this is remake №1 — the thing we watch, per §1d):
- `SELECT current_step, status FROM orchestration_states WHERE orchestration_id='e44a44d7-…';`
- the cascade after the classifier: strategy → briefing → plan → build items on site
  `d991a5b8`; spawn→call handshake races are a known fleet flake — never cancel a FAILED
  handler row pre-diagnosis (memory: 2 COMPLETED / 2 FAILED all-history).
- at build review: the DECISION's guard — no negative-identity copy baked into chrome/footer
  ("we don't sell advertising" must not become load-bearing served copy).
- the site row should reach `deployed` by pipeline action, not by hand.

### (b) First stall of remake №1 — a NULL `name` planted 08-26 fired at release; all six brief rows carried it

The cascade ran clean through classifier (12:15→12:17Z, mission_brief untouched — `5dac12fd`
still current, four new aspects written) → vertical research → strategy → briefing, then
`needs_site_plan` attempt 1 FAILED: `ensure_site_record` → `upsertSite` → *"Scan error on
column \"name\": converting NULL to string"*. Cause: the minimal site rows this lane creates
for the brief-writer (domain+status+locked, buytoletcalculator precedent) leave `name` and
`network_id` NULL; `upsertSite`'s RETURNING scans BOTH bare (no COALESCE — unlike
email/company/tagline etc.), and its ON CONFLICT arm only bumps `updated_at`, never backfills.
So the trap arms at brief time and fires at RELEASE, weeks later, on the first item whose
workflow runs `ensure_site_record`.

- **Data fixed 12:5xZ, all SIX rows in one guarded UPDATE** (advertise + the five held briefs
  — every one carried the same NULLs, so each future release would have stalled the same way):
  `name = domain`, `network_id = 00000000-0000-0000-0000-000000000002` (the only network in
  use, 33 sites). Item retries at 13:01:04Z (attempt 1/3 consumed; dispatcher retry, no hand
  re-fire needed).
- **Door closed at the source**: `scripts/fire-brief-writer.sh` header now spells out the full
  INSERT (name+network_id included) and the exit-3 message points at it. The next lane creating
  a site row for a brief gets told at the point of use.
- Platform-side residual, deliberately NOT fixed here: `upsertSite` could COALESCE name in the
  RETURNING or backfill on conflict — that is a platform/ change (council scope) and the data
  fix removes the live need; noting it rather than shipping a seam mid-watch.

### (c) gamedesign.uk steer for a peer lane — new register family GAMES/GAME DESIGN, rows GD1+GD2

A peer session (owner-directed) is rebuilding gamedesign.uk FRESH via 082 and asked this lane
for the differentiated position vs gamesdesign.co.uk (the healthy sibling, another lane's build,
rebuilt 09-01). Steer given, grounded not invented: P5 applied (.co.uk=authority/.uk=instrument)
→ gamedesign.uk = the professional/studio PRACTICE side (editorial/process/workflow for teams;
leads/producers/professional designers), sibling keeps free tools+guides+learning; commercial
slot PREPARED never claimed (the advertise §3 pattern; sibling's live strategy records a
paid-tier path — but the literal "GameDesign.uk Pro" name is in NO current sibling spec,
verified, so the name was flagged unrecorded). Collisions named: designblog.co.uk (design
editorial, brief fired today) · cartoon.co.uk PROTECTED · gamerooms.co.uk stub · writesy.uk.
Neither game domain had a register row (verified '%game%' sweep — only the gamerooms stub).
**Rows written to `positioning_register` (DB ONLY — MD copy untouched, two-copies rule):
GD1 gamesdesign.co.uk (documents the occupied position; their build untouched) · GD2
gamedesign.uk (proposed 2026-09-02; neighbours + must_nots machine-readable; revise on ask if
the owner-reviewed brief lands elsewhere).** Register-before-build satisfied for their dispatch.
Their orphan finding (site serving with pages rows DELETED — invisible to
audit-archived-still-serving) is theirs to file; flagged the 359 family for their grep.

### (d) gamedesign.uk peer lane: steer ADOPTED as binding; and their root cause raises our register debt's stakes

Peer confirmed GD2 unchanged (their commit eba9c3bb6); all four steer points + collision list
binding on their brief; they take the oufe pre-seed route WITH name+network_id set explicitly.
Their April root cause (adoption ran gamedesign.uk as both source and destination; rerender
published empty placeholders pre-guards; CLOSED by three dated guards, their code-read) needs
nothing from GD1/GD2. **The fact that is OURS to keep: a platform-deployed domain served
publicly for 4.5 months with NO sites row and NO pages rows — invisible to
audit-archived-still-serving (keys on pages.status='archived'; deleted rows are not archived
rows). A register row is currently the only artefact that would ever notice such an orphan.
That upgrades the "21 portfolio domains have no register row" debt (handoff §3, counted
2026-08-21) from tidiness to a serving-side blind spot** — worth saying when the owner next
prioritises. The bug filing for the orphan/detector gap stays THEIRS (owner steering their
lane; they are grepping the 359 family first).

### (e) Remake №1 mid-build: plan complete on retry, full fan-out, and three `owned_page_review` items closed on artefact evidence

Site plan attempt 2 (orch `21ef4ac2`) passed `ensure_site` — the (b) fix held — and COMPLETED;
plan `046c9eee`. Fan-out: 18 `needs_page` + 18 `needs_imagery` + 24 `page_rerender` + content
items; 7 tools built via `add_tool` (tool-suggester) — fullness visibly landed.

**Three `owned_page_review` items filed 13:09:48Z at `needs_human_review`** ("Owned page
tool-X is not_built — do NOT route to the generic page builder"). Checked at the artefact, not
the status: all three components ALREADY deployed with real payloads (16,705 / 16,953 /
14,553 B, created 12:57–13:06Z — BEFORE the reviews were filed, so not a completion race;
the validator reads some earlier/other surface [INFERRED — validator code unread]). Routing
verified correct on every neighbouring item: the tool pages' content items come from
`tool-deployer` itself (component-aware by construction); the generic `needs_page` items
target only the three GUIDE pages. **Closed all three with the evidence in `result`** —
nothing depended on them, nothing automated ever clears `needs_human_review`, and a genuine
future not_built re-files freely (dedup key released on completion).

Observation worth handing to the checker-layer lane (149 family), not fixing here: a
plan-validation review item whose condition self-resolves minutes later has NO auto-clear and
NO dependent it gates — it just sits in the human queue for ever. On a fleet of 22 remakes
that is 3×22 stale rows unless the validator re-checks at read time.

### (f) Two mid-build review items: one re-queued on evidence, one left for the owner

- **Content blocker on the ab-test tool page**: writer emitted "More tools coming soon." →
  placeholder check blocked (rightly). Detail was NOT in the item/orchestration — it lives in
  `agent_error_log` under the SECOND row (severity `warning`, "see context.issues"; the generic
  CONTENT_VALIDATION_FAILED row carries no issues — worth remembering when chasing any
  "N blockers" message). Nothing persisted (validate precedes save). Re-queued triaged for a
  fresh writer run; escalate only on recurrence.
- **`brief_supplies_negation` (needs_human_review) LEFT OPEN — it is the owner's by its own
  spec.** The flagged phrase is an instructional style negation in `content_direction.formatted`
  ("tell the reader what to do with the information, not just what the information is"),
  `mandated: false`, 0 mandated onto pages. My read for the owner: keep as is — it is writing
  guidance, not a negative-identity claim; the ruling's target class ("we don't sell X") is
  absent. The check's own metadata already exempts 11 instructional negations; this one tripped
  the x_not_y shape matcher. Owner's word applies the fix text (whole-object content_direction
  rewrite) or closes it as accepted.

### (g) FLEET LLM OUTAGE at 14:43Z — Anthropic credits exhausted, mid-remake; build paused, self-resuming

At 14 of 18 plan pages built, two `needs_page` runs (news-index, uk-ad-spend-reference) failed
with the provider's own words: *"Your credit balance is too low to access the Anthropic API."*
Verified at the instrument the LANDMINES entry prescribes — `llm_call_log` success column,
never orchestration statuses: **ok goes 2/1/... to 0 at 14:43Z, all calls failing since** —
a fleet-wide LLM outage, not a build defect. Owner push-notified (terminal; no Remote Control).

- The failed items self-requeued with `retry_after` ~15:00Z, attempt_count still 0 — when
  credits return the build resumes UNAIDED; no re-queue owed. Watch attempt burn if the outage
  spans multiple retry windows (max_attempts 3).
- Recovery is proven ONLY by the first minute with ok>0 in `llm_call_log` (landmine: the fleet
  looks green throughout — dispatch plumbing completes without calling any model). A session
  monitor watches for exactly that.
- Owner hint recorded in the push (memory: the-fleet-key-is-not-on-the-default-console-org):
  the fleet key is NOT on the default console org — pick the account by the key's "Last used".
- Also this window: earlier `needs_section_data` on the contact page = missing business email;
  supplied `advertise@contactforsales.com` (estate catch-all convention, 11/12 recent sites),
  superseded the identity spec with the email at BOTH the flat and nested paths (the
  identity.email vs identity.contact.email landmine), completed the handlerless item —
  `swi_no_handlerless_promotable` (a check constraint) refuses to re-queue such items, learned
  by a rolled-back attempt. Contact page's own build item was still queued, so it picks the
  datum up naturally.

### (h) Owner: "the 3 new briefs are good" → cluster builds RELEASED (remakes №2–4), ~15:1xZ

Approval unedited, so the briefs stand as fired (websitepromotion 15 / seotools 18 / designblog
20 plan items). Same recipe as advertise (§ this file, morning): three review items → complete
(`approved_by='owner'`, instruction quoted in result) · three `research_<domain>` items created
triaged (Flow A shape) · **emails set BEFORE release** (`<name>@contactforsales.com`) — the
door the advertise build hit at its contact page, closed proactively this time · three site
rows `test`+LOCKED → `active`, unlock LAST, one guarded UPDATE asserting exactly 3 rows.
Dispatcher will interleave them with advertise's remaining drain (selector = oldest eligible
site first). Compact monitor armed for the trio (state deltas + errors only, exits at 3×
deployed). Advertise at this point: ~15/18 plan pages done post-outage, imagery nearly done,
rerenders queued. The owner's review queue now holds only the two TEST briefs
(indoorplanters, buytoletcalculator) + the advertise negation flag (recommendation standing:
keep as is).

### (i) Remake №1 BUILT AND DEPLOYED — 16:23:25Z; the domain still serves Drupal because DNS was never cut over

Site row flipped to `deployed` by the pipeline at 16:23:25Z. Verified at ARTEFACTS, not statuses:
- `gqls/sites` repo holds the complete `advertise.co.uk/` tree; built homepage 72,283 B,
  title "Advertise.co.uk — The UK Guide to Advertising", **0 negative-identity phrases** (the
  DECISION guard held); 24 pages, 20 deployed + 4 in the convergence wave (`needs_rebuild`:
  index, contact, regulation-map, ad-auctions blog — imagery weave + rerender items draining).
- All 6 `unresolved_cta` items verified OVERTAKEN at the repo artefact (3 resolved CTA hrefs
  per guide, 0 empty) and closed with evidence. Owner's negation flag stays open (theirs).
- ⚠ **A "deployed" site row does NOT mean the domain serves it.** advertise.co.uk still serves
  the OLD Drupal 7 site: NS = dns.us-noc.com/dns.uk-noc.com (legacy host, A 62.182.23.30).
  My first curl mis-read the Drupal "Page not found | Advertise" 404s as the new site's — the
  worked lesson: a branded 404 title is NOT proof of which stack answered; read the Generator
  meta. The RUNBOOK (dns_pointing_a_domain_at_the_serving_worker) governs the cutover:
  CF zone (needs the DNS-scoped token — `~/.cloudflare/dns-token.env` was NEVER created; only
  the Workers-scoped 404-token exists) → owner sets NS at Nominet (manual, no access) →
  apex proxied A + worker route (scriptable). **Same cutover will be owed for
  websitepromotion/seotools/designblog** — consider doing all four as one Nominet batch,
  which is exactly the bulk shape the 08-18 owner ruling prescribed.

### (j) Three owner asks in one evening burst — all routed, none executed out of lane

1. **gamesdesign.co.uk naming (owner via the gamedesign.uk peer, then directly)**: the site must
   stop using the brand "GameDesign.uk" (23/49 titles as of 09-02, inherited verbatim at the
   June adoption). Owner then ruled THIS lane must not touch the site and started a dedicated
   session (`gamesdesign.co.uk` [783baf]). What this lane did: GD1 records the ruling + the
   recommendation ("GamesDesign.co.uk", domain-as-brand, recommended-not-decided); full evidence
   package handed to the new session (measurement — four current specs carried the string at
   ~17:45Z, down from the peer's six, re-count before acting — plus supersede-not-update,
   case-sensitive replace, ~23 rerenders); the class bug ("adoption carries source company_name
   verbatim") stays the gamedesign.uk peer's to file.
2. **WebProNews RSS feed (owner)**: details measured and CONTRIB'd to the news lane
   (`news_feed_ingestion/CONTRIB_2026-09-02_..._webpronews_feed_candidate.md`, commit
   ebc050732) + messaged to the `feed lane` session. Key fact: the "current consumer" is the
   OLD Drupal advertise.co.uk (dies at DNS cutover); wholesale-import caution attached.
3. **indoorplanters.uk acquired (owner)**: register stub `GDN1b` written (twin of the briefed
   .co.uk; P5 pair decision owed before either builds).

## 2026-09-02 (evening) — INBOUND from the new nominet_domain_management lane: the DNS cutover HAS RUN and the four domains are dark until the CF zones exist

Telling you, not just measuring (owner ruling 2026-07-29 §3), because your live task is
watching remake №1 end-to-end and its domain no longer resolves:

- The registry (`dns1.nic.uk`, measured ~17:00Z) already delegates **advertise.co.uk,
  designblog.co.uk, seotools.co.uk, websitepromotion.co.uk** to alexis/leah — the owner's
  "Nominet batch" happened. **No Cloudflare zone exists for any of the four**, so the edge
  REFUSES and each goes dark as resolver caches (2-day TTL) drain. Your 16:23Z "domain still
  serves Drupal" was cache, and the WebProNews note's "dies at DNS cutover" consumer is
  dying NOW, not at some future date.
- Recovery is staged (`scripts/domains/cf-zone-bootstrap.sh`, owner-run; asked for in the
  nominet lane's README). Once the zones are active the four serve the portfolio router —
  №1's remake appears there when its build deploys; nothing for your lane to do except know
  that an HTTP probe of advertise.co.uk today measures DNS state, not build state.

**RESOLVED same evening (~19:00Z): all four zones ACTIVE, all four domains SERVING** —
the twist was a Cloudflare NS-pair mismatch (new zones get betty/ivan, the owner's batch
had pointed at alexis/leah); repointed at Nominet via the new nominet.py client, verified
at the registry and at the edge by body fetch. **advertise.co.uk now publicly serves the
remake** (200, 75,562 B, title "Advertise.co.uk — The UK Guide to Advertising") — №1's
launch is complete. ⚠ **For your lane's eye: designblog/seotools/websitepromotion.co.uk
also serve full titled sites** (69–72 KB bodies) while their briefs sit at
`needs_human_review` — whatever those bodies are, they predate their remakes; your lane
owns whether that interim content should stand.

### (k) ALL FOUR remakes DEPLOYED in one day — serving now gated ONLY on Cloudflare zones

advertise 16:23Z · seotools ~18:2xZ · websitepromotion ~18:3xZ · designblog ~18:5xZ. Convergence
waves still drain on all four (rerenders, designblog's 6 content_rewrite items, one 424-class
imagery retry on websitepromotion — that bug lane's). Standing review piles for LATER batch
verification at artefacts: ~54 unresolved_cta across the trio (the advertise precedent says most
resolve when hubs render) · seotools' 7 + websitepromotion's 1 owned_page_review (TRUE until the
design-discovery sweep files their evaluate_tools — seotools DEPLOYED without its 7 planned
tools; self-heals via the sweep, invisible while the domain is dark) · websitepromotion's
claims_unverified (owner) · advertise's negation flag (owner).

**DNS state 18:0x–19:0xZ: all four domains LAME-DELEGATED** — owner set alexis/leah at Nominet
(correct pair, matches the account's 36 zones) BEFORE zones existed; registry answers SERVFAIL
"no reachable authority". Old sites dark; new sites unreachable. Remedy = create the four zones
in the CF account (dashboard, or the never-created DNS token per RUNBOOK's 08-18 recipe); then
apex proxied A (dashboard/token) + worker route (this lane's 404-token CAN, step 5). Zone watch
armed; routes get added the moment zones appear. ⚠ the runbook's order (zone FIRST, then NS)
exists precisely because of this failure shape — worth a LANDMINES entry if it recurs.

### (l) gamesdesign.co.uk rename EXECUTED by its session; GD1 name DECIDED; new pair dependency

Owner confirmed "GamesDesign.co.uk" in that session; GD1 moved recommended→DECIDED with their
execution record (4 specs superseded properly, 23 titles + 30 components swapped, 32 rerenders,
backups bak_gdcouk_rename_20260902_*; class bug filed by the gamedesign.uk session as
bugs_open/439). NEW pair fact on both GD rows: sibling's guide-p2p-architecture deep-links to
gamedesign.uk/games/p2p-networking/ — GD2's rebuild brief must keep/redirect that path or have
the sibling drop the link; both site sessions informed.

### (m) Rename serve-verified by the gamesdesign session; Pro-tier name/home recorded UNDECIDED

Their verification: 0/31 serving pages carry the old brand; the p2p link races were caught and
re-dispatched. Register residue recorded on GD1+GD2: "GamesDesign.co.uk Pro" now exists in one
served meta (contact) + the never-built premium.html title — my earlier "Pro name in no current
spec" was TRUE of specs but the name lived in pages.title/meta (a lesson: a spec census is not a
pages census). The paid tier's NAME and HOME stay undecided; owner call at GD2 brief time.
Their lane docs: docs024_key_docs_latest/gamesdesign_couk/ (dc1764660, bf0dc007d).

### (n) DOMAINS LIVE — all four serve the new sites; verified at the bodies ~19:4xZ

Owner created the zones ("the domains are now live"); all four resolve to Cloudflare proxy IPs
and serve 200 with the new homepages (advertise 75,562 B "The UK Guide to Advertising" — 0
negative-identity, 0 Drupal; websitepromotion; seotools; designblog). Advertise deep pages:
tool 200 91,532 B, glossary 200, news 200. Expected residuals: advertise /sitemap.xml 404
(642 generates on next rotation selection — verify tomorrow); seotools /guides/index.html 404
(its convergence + tool sweep still pending). Zone watcher stopped.
⚠ **The 404-token listed NO zone for any of the four while all four were serving through the
worker** — its zone visibility is scoped or account-partial; never read that token's empty
zone list as zone absence. Recorded in the DNS runbook.

### (o) Owner critique of the remakes (via designblog session) → bugs_open/444 filed; §9/§10 entries

The owner reviewed designblog and said much applies to advertise/websitepromotion. Verified on
advertise same evening — TRUE and worse than it looks because every naive check passes:
channels-directory 0 entries, glossary 0 terms, news 0 items, all serving 200 at ~60KB of
brief-echo prose. Root causes split three ways (feed: mechanism live but 0 content_sources —
undriven; directory: DIR-001 wired but the vertical's KIND doesn't exist; glossary/showcase:
no producer found [absence claim]). Also class-wide: no tools nav link/hub on any remake (plan
precedes tools). Filed **bugs_open/444** with fix candidates ordered by door-closing (top:
plan validation refuses/degrades zero-item listing pages); 016b §9 pattern ("count the ITEMS,
never the bytes") + §10 row added. Answered the designblog session's (a)–(c) with the
measurements; their instances stay theirs, advertise's are ours. **Consequence for the next 18
briefs: until 444's candidate (1) or (2) lands, every brief that plans a directory/feed/
glossary page is planning a page that will ship empty — fire directions must carry that.**
The owner's "best in class" directive routed by the designblog session to the design lanes.

### (p) Theme kits' differentiation offer taken up for the next 18 briefs

Via the designblog session: theme kits measured 36/37 sites rendering identical chrome
(ChromeSlotFunction hardcodes it; 10 chrome-eligible header functions unused; 6/40
style_collections pin header_component_id) and offer per-remake chrome+layout+structure
recommendations AS DATA, applicable per-site today. Components + vigilant designer measured
the same shape independently (top-10 components carry 78–87% of slots; advertise AND
websitepromotion in the list). Messaged theme kits directly: asked for the data shape + the
application point for the next briefs (window exists — briefs HELD on 444 enablement), and
put retro-differentiation of the four live sites explicitly behind the new-brief work pending
their + the owner's call. **Next fire directions will carry: theme-kit differentiation data +
444 enablement (sources/kinds) or listing-page holdback + best-in-vertical/no-omission/
fullness/no-negative-identity (standing).** Cross-thread digest with attribution:
designblog_couk/NOTES_designblog_couk.md.

### (q) Theme kits' measured answer reshapes the differentiation plan — levers ranked by proof

Their CONTRIB (this dir, 2026-09-02) deflates two of their own claims and ranks the levers:
- **LAYOUT = the lever this week, proven + brief-driven**: 9 of 18 layouts used, 73% of sites
  on three, NINE never used; layout survives the design overlay (colour does not). Goes in the
  BRIEF (decided at composition from classification + style_direction prose).
- **COLOUR = brief-with-REFERENT technique** (proven on gamedesign.uk: brief named the
  sibling's actual hex values to differ from; overlay landed within two hex steps). Never plan
  colour via themes/palettes/pins (spec-wins; owner's 09-02 ruling).
- **CHROME = strongest lever, UNPROVEN**: all 6 existing pins point at the very component the
  default picks — "pin honoured" and "pin ignored" are indistinguishable in all current data.
  Decisive experiment (theirs, ~10 min): pin ONE site to header-with-search, rerender, diff
  SERVED html vs unpinned sibling. **Decision: fold the experiment into remake №5's build**
  (pin pre-release-rerender — no live-site touch, answers the question before the other 17
  fire); the owner can pull it earlier onto advertise/websitepromotion if he wants a live one
  differentiated sooner.
- **STRUCTURE: do NOT size on page_archetypes** — 94.4% of live pages match no
  defaultSectionsForPage output; the lever is the planner prompt (components lane's ground).
- **`contact-hero` poison** (their §6): 3 plan rows + 1 page name a component with ZERO rows
  (renders an empty slot); working pages use `hero-contact`. Vigilant-designer thread owns the
  repair; OUR vigilance: at brief/plan review of the 18, grep the plan for contact-hero.

**Fire-direction template for the next 18 (supersedes (p)'s checklist):** per-domain vertical +
best-in-vertical + no-omission + fullness + no-negative-identity (standing) · LAYOUT intent
naming a distinct layout, preferring the nine unused · COLOUR referent = nearest estate
neighbour's actual values · 444 enablement (feed source rows / directory kinds) or listing-page
holdback · chrome pin data per §5 once the №5 experiment proves the mechanism.

### (r) The platform violated a P5 seat split (bugs_open/447) — recorded on every pair surface

tool-suggester (reads identity+classification only; no positioning vocabulary) proposed SIX of
gamesdesign.co.uk's tools by name onto gamedesign.uk; the brief-fidelity guard RECORDED the
violation but dispatches nothing. Held reversibly by their lane. The class ("a dispatch-mode
agent that never reads the seat cannot honour a seat split") is now on GD1 (mechanism), GD2
(machine-readable intent: hosts_tools=FALSE, consumer = 447's opt-in build), GDN1b (the
indoorplanters pair inherits the watch), and RUNBOOK_remake_release §2 (eye every
evaluate_tools wave on a paired site until 447 lands). Their 446: owner re-ruled gamedesign's
vertical execution (louder, gamier) — GD2 substance unchanged.

### (s) Session close — fresh chassis deployed (owner message, late 2026-09-02); handoff finalised

Nothing of this lane's rides the roll (DB config + docs + work items only today). Cautions
recorded in the handoff: ~300s no-dispatch window post-pod-start; riding lanes (424, 444's
gated code) verify per-service at the stamp; 444's migration 720 stays HELD until their
round-3 verdict regardless of the roll. Handoff finalised at `2837ab63a` — cold start for the
next session is HANDOFF_2026-09-02_continue_here.md.

### (t) 2026-09-02 ~21:20Z — short session: 444 APPROVED, two sitemaps explained, seotools' tool holds proven TRUE at the body, then the cluster token expired

Cold start from the 09-02 handoff; lane dir clean (every dirty file in the tree is another
session's — kustomize overlays, 444's bug file, platform tests).

- **444 round 3 APPROVED 20:53:22Z** (`orchestration_states` corr `c0990eb3`: 57e04dbe REVISE
  19:56Z → 5cb63d3e REVISE 20:20Z → 8e041944 `complete_approved` 20:41→20:53Z). Their
  uncommitted bug-file note says 720 APPLIED + verified live, Go gate `6525b45ae` inert until a
  roll. Read `720_planner_listing_source_gate.sql`: anchored replace of the planner's rule 3
  (drops the "entity-directory may have empty sections" licence, adds the listing-source rule)
  + `validate_plan.config.enforce_listing_sources=true` (RFC_022 shape, one live consumer).
  Runbook §2 + handoff addendum carry it. NOT re-verified at the DB — token expired first.
- **Sitemaps**: advertise + seotools `/sitemap.xml` 404; the other two 200. The handoff's
  "expected, 642's next selection" was the right prediction for the wrong reason — the
  selection HAD happened (advertise 16:37:46Z, seotools 17:38:46Z), inside the lame-delegation
  window, and the probe dropped 22/22 and 14/14 → `url_count` 0, run COMPLETED, stamp advanced,
  nothing published. Trap recorded in runbook §3. Self-heals via the change-and-quiet arm
  (due set 0/14/5 at 21:0xZ, rotation ticking, last 20:43Z).
- **seotools' 7 tool pages** — measured at the body with a CONTROL in the same run:
  `/tools/<slug>/` ×7: 200, 55,684–56,860 B, the tool's H1, **0 form / 0 input / 0 select /
  1 button (mobile-menu toggle) / 3 scripts**. advertise's 3 real tools, identical probe:
  1 form, 11/11/0 inputs, 0/1/2 selects, 4 scripts. So the seven `owned_page_review` holds are
  TRUE — prose shells at the tool URLs — the OPPOSITE reading from advertise's (e) on the same
  item text. Recorded as a runbook §4 check. Who wrote the shells is [DB]-dependent (open:
  `save_refused_incomplete` on serp-snippet-previewer 20:40Z; design rotation last tick
  18:47:45Z vs 20:15–21:00Z for the other three — cadence or stall UNVERIFIED).
- **advertise listing pages** re-probed: `/news/index.html` 200 / 61,077 B, h3=1;
  `/channels-directory/index.html` 200 / 60,420 B, h3=1 — consistent with 444's 0-item
  measurement (a directory of channels would carry many h3s); `/glossary.html` is the real
  path (my `/glossary/` guess 404'd — path shape, not absence).
- **Token expired 21:08:03Z** mid-query (iat 08-30 21:08:03Z, exactly 3 days; memory file's
  python check confirms). The batch that died would have read: seotools tool `page_components`,
  advertise's `site_unreachable` + 2 FAILED `needs_page`, the odd websitepromotion HITL
  `needs_page`, one `unresolved_cta` spec for the batch-verify shape, the two dark-window
  sitemap orchestrations in full. All listed in the handoff addendum as first reads after the
  owner refreshes the kubeconfig. Owner told in README.
- Misstep worth a line: none caught this session; the one near-miss was reading "sitemap 404 =
  not yet selected" off the handoff — the DB said "selected and empty". The check that
  caught it was the routine §1a rotation query, which is why it is routine.

> **CORRECTED 2026-09-02 21:1xZ (same session, ~10 min after writing):** every "22:xxZ" stamp in
> (t), the handoff addendum and the runbook edits above was **BST written as UTC** — the token
> expired **21:08:03Z** (22:08:03 BST), and this session ran ~21:05–21:20Z. The memory file's
> python check prints `datetime.fromtimestamp` (LOCAL time) with no zone label, and I stamped its
> output "Z". Caught by the dispatch_throughput lane's commit message (`dc8bc044e`: "kubeconfig
> expired 21:08Z") disagreeing with mine by exactly one hour, then proven with
> `fromtimestamp(exp, timezone.utc)` and `date -u`. The cheap check: any timestamp you are about
> to write with a Z came from a tool that printed the zone, or it did not. Fixed in place in all
> three files; the memory snippet now prints UTC; WRONG_CALLS row added.

> **CORRECTED 2026-09-02 21:3xZ by the 444 session (their message):** my (t)/addendum inference
> "designblog's 21:04:44Z deferral ⇒ the roll carried `6525b45ae`" was HALF right. The row IS
> their repair firing (novel Reason string; `featured_post` one of five predicted unregistered
> bases), but it proves only `dbb218a41` (defer half, 20:08:04Z) rolled — the gate `6525b45ae`
> is a separate commit 16 min later and only the per-service stamp discriminates. Handoff
> addendum struck-and-corrected; runbook §6 now says an empty receipt query is ambiguous until
> the stamp is read. The check I skipped: two commits can ride one roll or not — a symptom that
> names one of them is evidence for that one only.

### (u) 2026-09-02 ~21:2x–22:2xZ — token back; §1a worked at the artefacts; seotools' shells diagnosed → `bugs_open/450`

- **Access**: owner refreshed the kubeconfig ~21:2xZ (444 session's message); every query below
  is post-refresh. 444's gate PROVEN LIVE by them (`560a24c07`) — runbook §2 rewritten; the
  receipt anti-churn rides the NEXT roll; CLAUDE.md's two deploy-proof recipes are landmined
  (no `build provenance` line on backend services; BusyBox `grep -aq` false-absences) — use
  their NUL-split probe.
- **CTA batch**: joined all 36 open `unresolved_cta` on the trio to `page_components.content_data`
  by (page_name, slot_name=section_name) → 11 had every reported-missing field filled; curled
  each served page cache-busted and counted `href="<value>"` (1–7 each, ≥1 required). Closed the
  11 (`closed_by`/`resolution`/`verified_at_artefact` — same shape as (i)); 25 remain with an
  empty field ("no eligible content hub" for `secondary_cta_url`). Query to re-run: the CTE in
  this session's transcript — reproduced in RUNBOOK? No: it is one-off; the shape is
  `spec->'missing'` × `pc.content_data->>m`.
- **seotools' 7 tool pages**: `page_type='tool'`, `rebuild_policy='generic'`,
  `sections=[hero-tool, generic-text-block]`; `site_plan_sections` for plan `5895b7ae`
  (current) says the SAME — the plan asked for a shell. `page_component_history` for all seven:
  every write's `source_item_id` = an `unbuilt_internal_link` item (20 filed 19:39:49Z, one per
  link; 26 writes over 6 pages 19:57–20:41Z; robots-txt-tester 4 writes = 2 items × 2 slots).
  No `evaluate_tools`/`add_tool` ever filed (live + archive); design-discovery has no rotation
  row for seotools or websitepromotion; rotation selections 09:43/12:43/15:44/18:47Z = ~3 h
  cadence. Guard: `save_page_sections_action.go:186` → `pageIsOwnedForGuard`
  (`owned_page_guard.go:176-190`) = `rebuild_policy='owned'` only. **Control**: advertise plan
  `046c9eee` (13:09:36Z) names `hero-tool,tool-guide-intro,tool-ab-test-calculator,tool-cta` —
  its tools existed (12:57–13:06Z, design rotation hit advertise 12:43:56Z). Fleet census with
  the control: 61 deployed tool-type pages without a tool-level component across 10 sites,
  advertise 0/3. Filed **`bugs_open/450`** (first-hand verification stated + 090 fired: intake
  `40879ff3`, run `96e97dc4`, `diagnosing` 21:42Z — verdict unread at close); 016b §9 + §10;
  LANDMINES entry (verifier armed); runbook §2 trap + §2b one-shot-discovery mitigation
  (UNEXERCISED; №5 canary #3; not for seotools until tool-deployer's behaviour on a
  pre-existing page row is known).
- **advertise**: `uk-advertising-regulation-map` never built — `mechanism-flow` `steps[].branches`
  array-vs-string, 4 failures = `bugs_open/437`'s instance (6 sites, filed today by loanzy).
  `site_unreachable` left for the availability probe (auto "probe recovered" precedent 21:22Z).
- **websitepromotion**: `needs_page` d0a5c53f HITL was a placeholder false positive on natural
  prose ("asking to be added to the relevant page") — re-queued once per (f).
- **Misstep, recorded**: my first handoff addendum wrote "who built the shells needs the DB" —
  right — but the (t) reading "design rotation last tick 18:47Z vs 20:15–21:00Z for the others,
  stall UNVERIFIED" resolved as CADENCE (3 h), i.e. the [UNVERIFIED] marker did its job: it was
  the first thing checked and it would have been wrong to act on.

> **CORRECTED 2026-09-02 ~22:0xZ (444 session, in 450 itself, `ad1b3b1fa`):** my (u) line "candidate 1
> is 444's gate generalised" overstated it — the plan-side hold cannot close the phantom-link
> door and risks starving the tool rotation (a tool's producer is outside the plan by design).
> 450's candidate order now reads (2)/(3) first, (1) conditional on §7 + a sibling key. What
> caught it: asking the adjacent lane the one question their gate could answer, in their file.
- **(u) follow-up, verified at the body**: websitepromotion `needs_page` d0a5c53f (re-queued
  21:39Z) → claimed, writer ran clean, deployed 21:41:02Z, page `deployed_at` 21:45:39Z;
  `/blog/promote-website-free-uk.html` 200 / 70,325 B with `hero-free-promotion.jpg` as the hero
  background-image, 0 × "to be added". My first probe 404'd at the ROOT path — the page lives
  under `/blog/`; read `pages.url` before curling (probe-page-url lesson, again).
- **090 on 450**: still `diagnosing` at 21:48Z (route step); a background watch prints the
  verdict when the item leaves `diagnosing`; if this session ends first, the query in the
  handoff addendum 2 is the pickup.

### (v) 2026-09-03 08:2x–08:4xZ — overnight read: 450 CONFIRMED + §7 answered at the rows; tools arrived under other names; cluster duplication surfaced

- 090 run `96e97dc4` → **CONFIRMED** 22:11:33Z (grounded on `owned_page_guard.go`'s constant and
  the seotools rows; it pulled `check_phantom_internal_links.go` by content first). The verdict
  is in `site_work_items.result` of the `needs_diagnosis` item — my handoff's `doc_notes`
  pointer was wrong (0 rows); corrected in 450 and addendum 3. My background watch was killed
  at 21:52Z with the item mid-run, so this was read cold this morning.
- Design rotation reached seotools 21:48:01Z — 12 min after my (u) reading "never selected" —
  and websitepromotion 03:49Z. `tool-deployer` created its OWN rows (8 + 7 tools, real
  `component_level='tool'` rows, 15–22 KB); planned names vs suggester names: 0/7 and 0/1
  matched. Shells re-deployed 00:07–00:09Z by a rerender wave, 0 forms at 08:2xZ. wp's
  sectionless planned tool page took the other fork (`mark_no_ready_sections` ×7 HITL +
  `needs_content_page`). 450 §7 answered; 444's deadlock branch refuted; runbook §2/§2b refined.
- Cluster duplication [MEASURED at component names, which carry the source site]: CPM/CPC on
  all three marketing-cluster sites; five more tools pairwise. Appended to `bugs_open/447` as
  a cluster-scale instance; owner question in README.
- Sitemaps: advertise 22 urls 05:52Z, seotools 35 urls 06:53Z — the change-and-quiet arm did
  re-select both within ~10 h, as the runbook §3 trap predicted. `site_unreachable` auto-closed
  21:50Z. Neither needed a hand.
- Not ours, noted: `capability_gap handler_missing: affiliate-link-manager` on both cluster
  siblings (revenue_shape findings with no registered agent) — a fleet registration gap.

### (w) 2026-09-03 ~08:5xZ — OWNER RULING: build the 8 planned tools; cluster duplicates KEPT; chassis roll within the hour

Ruling verbatim: "build the tools. duplicated tools across the cluster can be kept. Please be
aware that a new chassis is being prepared and will be deployed in the next hour."
Sequencing: a roll kills in-flight orchestrations and the ~300 s post-restart no-dispatch window
drops spawns — so the 8 `add_tool` items are PREPARED now (shape read from seotools' completed
rows) and FIRED after the roll is seen and settled. Open question being read before firing:
how `tool-deployer` names/creates its page when a row with the target name already exists
(the 7 shells) — the ruling wants the tools under the ALREADY-LINKED urls.
- **(w) follow-through**: `adopt_existing_page: true` LIVE on tool-generator `save_tool` (286
  closed, TL-044, arm committed 88897190e 08-16 — in v1.0.1355 and in whatever rolls next);
  `pages_site_id_name_key` UNIQUE is what the plain arm would hit. No library tool claims any of
  the 8 functions (nearest: robots-txt-GENERATOR, gtm-channel-fit). No backend provisioning for
  generated tools exists (`tool_backend_provision.go` header) → redirect-chain checker HELD;
  CWV via keyless PSI v5; meta-tag from pasted source. 7 items composed with build briefs from
  the shells' own meta/hero text (the brief's intent), guarded, dry-run ROLLBACK passed; watcher
  armed on "all chassis pods Running + newest start > 20:57:10Z + 420 s". wp `target_page`
  copied from the suggester's wp rows: `new` (wp has no tools page). File copied into the lane
  dir as `SQL_2026-09-03_fire_planned_tools_450_instance.sql`.
- **450 addendum-worthy**: the suggester's evaluate_tools reasoning on seotools treated the
  shells as deployed tools and suggested complements ("directly complements the existing
  robots.txt tester") — the shells fooled the suggester as well as every checker.

### (x) 2026-09-03 09:05:25Z — chassis v1.0.1356 rolled (first new pod 08:56:01Z, pair settled 08:58:07Z + 420 s); the 7 `add_tool` items FIRED

Monitor saw the roll at 08:56:03Z (2/3 Running, rolling update) and SETTLED at 09:05:10Z;
fired at 09:05:25Z from the file of record — `INSERT 0 7`, all guards passed, COMMIT. Items
(`triaged`, `tool-generator`): seotools 1961c217 robots-txt-tester · d6189102
core-web-vitals-checker · 876a0c54 title-tag-scorer · 095e55ca serp-snippet-previewer ·
6a757c3c keyword-difficulty-estimator · dce904b8 meta-tag-checker · websitepromotion 00b4981a
channel-prioritiser. Redirect-chain checker HELD (owner asked). A monitor reports each item's
terminal state and whether its page gained a `component_level='tool'` row; verification at the
BODY (runbook §4) follows. The 10-min-capped shell watcher was still polling at 09:04:50Z; any
later attempt of its own hits the dedup guard and aborts — the file is idempotent-safe by design.
- **(x) follow-through — OWNER RULING (second, ~09:1xZ): "accept a lesser redirect checker."**
  8th item fired from `SQL_2026-09-03b_fire_redirect_checker_lesser.sql` (same guards): a
  paste-in chain analyser (hops, loops, 301/302/307/308 semantics, collapse recommendation,
  timeline view) + a best-effort browser fetch that reports only `response.redirected`/final URL
  where CORS permits, prints a `curl -I -L` for the user otherwise, and says what it cannot see.
  All 8 planned tools now in the queue; the running monitor's exit threshold is 7, so read the
  8th by hand after it reports.

### (y) 2026-09-03 09:2x–09:5xZ — remake №5 chosen and its brief FIRED: copyonline.co.uk (corr `8aac8250`, orch `7b627de7`)

- **Queue reality for the 8 tool builds**: dispatch picks ONE site per ~4-min tick by oldest
  pending item across 21 pending sites; seotools sat 20th, websitepromotion 19th at 09:23Z
  (sites with 30–59 items ahead re-win until drained — 413's fair-share shape). No governor
  shed (0 traces; 674 still held). ETA 1–2 h; not jumping a shared queue. Monitor armed.
- **№5 selection**: the five non-twin single-pagers have NO register rows and NO sites rows
  (nor do the four live remakes — the brief-writer flow never required one; the brief is the
  reviewed positioning record). Old pages snapshotted to `salvage/<domain>/index.html`
  (fridge-magnets = merchandise RSS aggregator; conferences = conference-feed aggregator;
  copyonline = 2015 two-sided copywriter marketplace; dsgn = empty shell whose only menu item
  was "Experienced Copywriter"; catalogues = 2007 AdSense home-shopping page). For three canary
  duties the pick needs tools-friendly + no listing pages + uncrowded → **copyonline.co.uk**,
  positioned as the UK copywriting AUTHORITY (guides + browser tools), marketplace = prepared
  later slot (advertise §3 pattern). fridge-magnets next; conferences after (feed
  pre-enablement via `content_features.news_feed` is its natural shape).
- **Register**: CW1 copyonline (proposed; seat = setter/teacher of commercial copy; neighbours
  dsgn/writesy/advertise/websitepromotion/seotools with rules; must_nots incl. the 444/447
  classes) + CW2 dsgn.co.uk STUB (design, NOT copy — the collision found in its old menu).
  New family "COPYWRITING / CONTENT". DB only; MD untouched (two-copies rule).
- **Site row**: `3d965325`, complete (name, network_id, email copyonline@contactforsales.com),
  `test` + LOCKED — the script's header recipe.
- **Direction (template v2)**: vertical + best-in-vertical + no-omission + fullness + explicit
  NO listing pages (marketplace prepared, never claimed) + cluster deference (advertise
  flagship / websitepromotion / seotools, never duplicate tools — 447) + no negative identity +
  LAYOUT soft-editorial (one of the nine never-used) + COLOUR referents = designblog's and
  advertise's SERVED `:root` values (both share `#8b0000` CTA red and `#6b0000` hover — the
  generic-theme default; told to avoid it) + UK/ASA. Verbatim in the fire output / transcript;
  `BRIEF_2026-09-03_copyonline_co_uk_for_review.md` to be rendered when the brief lands.
- **Canary duties at release** (not at brief): §5 chrome pin on the style_collection before the
  release rerender · §5b imagery.sections on the plan · §2b one-shot design discovery the moment
  the plan completes. Recorded here so the release step does all three.
- **(y) follow-through — first tool landed 09:34:09Z (`tool-robots-txt-tester`), and it ADOPTED
  the shell**: `page_adopted: true`, page `6feb9797` (born 09-02 16:13Z), URL unchanged, no
  duplicate row, tool component 20,839 B. Dispatch reached seotools at 09:31Z — far sooner than
  the 20th-of-21 queue position implied, so the "oldest pending item" ordering is not the whole
  selector; do not predict ETAs from it. ⚠ **Two components now sit at position 2** (the new tool
  AND the shell's `generic-text-block`) — `create_tool_component` hardcodes position 2. Recorded
  in 450; check the served order once the `page_rerender` (4970265d) drains, and decide per page
  whether the prose block is retired or the tool bumped.
- **(y) outcome — copyonline brief COMPLETED 09:31:17Z, held at `needs_human_review`** (orch
  `7b627de7` COMPLETED 09:31:33Z). 13 keys (the cluster-brief shape), confidence **0.82**,
  regulated_subject false, **24 content_plan items** (6 core guides + 4 core tools, then
  valuable/aspirational) — **no listing page planned**, and `directory_opportunity` states the
  copywriter directory is "explicitly gated behind the bugs_open/444 condition … no directory
  content should be planned, stubbed, or referenced in live copy at launch" — the direction's
  hardest instruction came back verbatim as a rule. Render:
  `BRIEF_2026-09-03_copyonline_co_uk_for_review.md`.
  ⚠ **Watch at plan review: "The Copywriter's Glossary" is typed `guide`** — correct, and it is
  exactly 444's stated blind spot (BLD-028: a listing page typed `content` passes the gate). As a
  guide the definitions are written inline; if it becomes a listing page it ships empty.
- **Two monitor lessons this session**: (1) the copyonline watcher emitted NOTHING for 25 min —
  its `kubectl exec` was timing out under load, and `[ -z "$st" ] && continue` swallowed it, so
  silence looked like "no change". The brief had in fact completed at 09:31Z. **A poll loop whose
  empty result is indistinguishable from its no-change result reports nothing either way** —
  emit on the empty case too. (2) The tool-build monitor's exit threshold is 7 while 8 items now
  exist; the 8th (redirect checker) needs a hand check.
- **437 answered for advertise (their message, 2026-09-03)**: our two `failed` `needs_page` items
  need NOTHING — `reconcile_site_plan`'s sweep re-mints on its own — and firing before their Go
  half `a0044e73b` rolls would burn a third attempt and earn the `[unresolved after 2 attempts]`
  brand, which (unlike `failed`) IS kept in the open set and blocks re-minting for ever. Gate and
  pre-fire checks now in the handoff. **Their root cause is a second instance of a class this
  lane already carries in memory: the prompt's own JSON exemplar declared `branches` as a STRING,
  so the model copied the demonstration it was shown** — and the flattening happened UPSTREAM
  (the section planner projects a component's element schema to a flat list of NAMES, collapsing
  a nested array-of-objects to a scalar). 119 refusals across 6 sites; the writer obeyed
  throughout and the type gate was right every time. My own filing called it "the writer emits a
  STRING", which read as writer misbehaviour — corrected here.

### (z) 2026-09-03 ~10:0xZ — the layout lever is NOT what this lane has been told, and it reframes `bugs_open/445`

Triggered by the 445 lane asking a question I could not answer from belief: *does a brief that names
a layout bypass the tag matcher?* Measured instead.

- **A brief cannot name a layout.** `resolve_composition_layout_action.go` scores ONLY
  `classification.category` + `classification.industry_tags` against the library, with a
  light/dark **scheme** derived from `design_intent.style_direction`. Its own comment is explicit:
  *"layout resolution never consults design_intent at all"* and *"Human-set signals
  (mission/design_intent driving the tag matcher below) are not consulted for layout today."*
  The only short-circuit above the matcher is a **theme kit** naming a layout id. So my template-v2
  line "LAYOUT intent naming a distinct layout" reaches the resolver **only** as (a) the light/dark
  scheme and (b) whatever the CLASSIFIER chooses to write into `industry_tags` after reading the
  mission brief. **CORRECTION to the theme-kits CONTRIB's §3 claim ("layout intent belongs in the
  BRIEF") as this lane applied it:** the brief is an influence on a classifier, not a lever on the
  resolver. `soft-editorial` in the copyonline direction is a hope, not an instruction — **verify
  at №5's composition, do not assume it landed.**
- **How the three sites actually got magazine-grid — and it is not a "compromise match".** The
  classifier wrote the literal string **`magazine-grid` into `industry_tags`** for advertise,
  designblog AND websitepromotion (their tag arrays start with it). A tag that IS a layout's name
  scores against that layout directly. **[MEASURED 2026-09-03] 12 sites fleet-wide carry a layout
  NAME inside `classification.industry_tags`** — 8 × `magazine-grid` (advertise, apis.uk,
  boxingonline, designblog, gamedesign.uk, homegarden, relojistas, websitepromotion) and
  4 × `affiliate-hub` (garden-tools, loancalculator, mortgagecalculator, seotools); those are the
  only two layout names appearing. **[UNVERIFIED]** that each of the 12 then resolved to the layout
  it named — I could not find where the resolved layout is stored per site (`css_themes` has no
  `site_id`; `style_collections` has no `layout_id`; `site_plan_sections.layout_id` is per-section).
  The tag census is the solid half.
- **Why that matters to 445**: their proposed signal fires on a WEAK/compromise best match. If the
  classifier is naming the layout, the match is STRONG and their detector would stay silent on the
  exact three sites that motivated the bug. Told them before they build it (message + this entry).
  This is the "a detector tuned to the wrong mechanism is silent on its own motivating case" class.
- **(z) follow-on — the owner's seotools critique reached us via the designblog lane, and it is the
  450 class already in repair.** Owner: "many seotools tools are description pages … a major error".
  Verified by them at the served bytes on `/tools/serp-snippet-previewer/` (0 inputs). Answered with
  the sequence: shells built by the phantom-link repair BEFORE any tool existed → rotation landed
  21:48Z but built 8 tools under DIFFERENT names (0/7 matched) → owner ruled build them → all 7
  completed 09:30–09:54Z. Their example's new component carries **2 inputs + 1 textarea + 1 script**
  [MEASURED 09:5xZ at `page_components.rendered_html`]; invisible only because 7 rerenders sit
  `triaged`. **Their sharper point, taken: the shell's promising COPY is a defect in itself and the
  repair does not remove it** — `generic-text-block` stays, and the tool lands at the same
  position 2, so each repaired page will serve both. Decide per page at the body.
- **Three owner rulings taken into the direction template** (runbook §6): page_archetypes APPLY
  (supersedes the theme-kits CONTRIB line); glossary/inspiration held in briefs until 444's producer
  lives, with a prose `guide` as the compliant form; feed-shaped pages stay section-index filling
  from child pages.
- **(z) chrome pin RE-SCOPED after a theme-kits correction — and the measurement is worse than
  their framing.** They flagged that pinning selects but does not populate, and that the four
  alternatives need 11–19 template variables. Measured here [2026-09-03]: **37 of 39 sites supply
  ZERO header `content_data` keys** (2 supply any; max 5), against candidates requiring
  cart-or-nav 11 / with-search 12 / minimal-tool 16 / with-categories 16 / docs 19. **So the fleet's
  identical chrome is a DATA problem, not a selection problem** — the default `site-header` wins
  because it is the only one that needs no data — and their own earlier "selection is the
  bottleneck" framing is superseded (as is my template-v2 reading of it). RUNBOOK §5 rewritten: the
  one-UPDATE form is withdrawn; the safe experiment supplies the vocabulary FIRST, then pins, and
  the third branch of the three-way read now says empty/broken means *the pin was honoured and the
  data is missing*. For copyonline, `header-minimal-tool`'s tool vocabulary is the fillable one
  (it ships four tools); search needs a real endpoint, cart and docs are wrong for it.

### (aa) 2026-09-03 10:2xZ — the repair wave is BUILT but not PUBLISHED, and that is a queue, not a fault

- **State**: 7 of 8 tools built (09:30–09:54Z, no retries, every one adopting its existing page at
  the existing URL). **All 7 `page_rerender` items still `triaged`**, oldest since 09:34Z; the 8th
  build (`tool-channel-prioritiser`, websitepromotion) still `triaged` since 09:05Z. Nothing on
  either site is claimed; neither site is locked.
- **Checked rather than assumed** (memory: *detection works; schedule and dispatch do not*): the
  `build-pipeline-trigger` ticks (10:21:20Z), the dispatch loop is EXECUTING, and **22 page_rerenders
  completed fleet-wide in the last 15 min** — the machinery is healthy. It has simply not returned
  to our two sites: since 09:52Z it has picked gaswholesalers, gamesdesign and robot-hands only.
- **Why**: the selector takes the site whose OLDEST pending item is oldest. Our items were created
  09:05–09:54Z, so every site with an older backlog outranks us; seotools/websitepromotion are
  ~13th–15th. Sites ARE draining (gamesdesign 36 → 0, gaswholesalers 30 → 6 within the hour), so
  this resolves itself. ⚠ Not purely FIFO though: `vetcomparison.uk` has held position 1 since
  06:57Z with 28 items and is not being picked — [UNVERIFIED] why, not ours, but it means "oldest
  first" is not the whole selector and ETAs from it are unreliable (I already got this wrong once
  today, predicting 20th-of-21 and being served in 26 minutes).
- **Not intervening.** Jumping a shared queue for this lane would be the wrong act on a fleet where
  every other site's work is equally real. The URL watcher reports the moment each tool serves.
- **`bugs_open/450` class fix taken by a fixing lane** (`bugfix_450_tool_page_shells/`, their
  `587666be8`, inert until the next roll). They verified BEFORE writing it that our wave cannot
  deadlock on their guard: `deploy_tool_action.go` inserts the tool `page_components` row (:517)
  *before* raising its companion `needs_content_page` (:564), so the page already has a tool when
  the follow-up work is filed and their derived predicate is already false. Our 09:34Z completion
  is recorded in their file as the handover case. **Two consequences that are OURS:** (1) a repaired
  page is generically re-buildable again, so the position-2 leftover `generic-text-block` is not
  held still for us — decide it at the body; (2) from the next roll, the remaining unrepaired shells'
  `unbuilt_internal_link` items terminate **`wont_fix`** with an `owned_page_review` receipt
  (`spec->>'refusal_class'='tool_pending'`) instead of rebuilding — **that is the guard working, not
  a wave of new failures**, so do not count terminal statuses and panic. They want to hear if our
  repairs turn up a shell their predicate misjudges (their [UNMEASURED] case: a page wrongly typed
  `tool` whose generic rebuild is genuinely wanted).
- **(aa) follow-on — the 450 shell census goes GREEN at component-attach, not at publication, and
  I nearly had it as my success metric.** Their correction, verified here 10:27Z: the census
  predicate asks whether a live `component_level='tool'` row exists on the page. **seotools has
  vanished from the census entirely** (61 → 54 shells fleet-wide) while **7 components attached /
  0 pages published** — every one of the seven still serves prose to the public. So the census is
  a repair-INITIATED count, not a repair-COMPLETED one, and the gap is an unbounded queue delay.
  Same family as *a `complete` work item is not a repaired artefact* — and a nastier instance,
  because here the green reading is a fleet-wide census a future session would quote as evidence.
  **Our acceptance test stays the served body (form/inputs), never the census.**
- **And the census cannot see the variant that matters most to the class.** It requires
  `p.deployed_at IS NOT NULL`. websitepromotion's `tool-channel-prioritiser` — the SECTIONLESS
  fork, which the fixing lane says is their best evidence for holding empty-sectioned tool pages —
  was never deployed, so it is **excluded from the census that measures its own class**. A count
  of shells cannot see a page that never shipped; the parked-HITL variant needs its own predicate.
  Sent to them.
- **(aa) closed out — the 450 census was a floor TWICE.** The fixing lane re-measured all four
  variants: filed predicate **54 / 9** today · +4 pages / 4 sites never-shipped (my finding,
  incl. our `tool-channel-prioritiser`) · **+9 pages / 4 sites whose only tool component is
  INACTIVE** (their own finding — the filed census never tested `cc.is_active` while their guard
  does, so finetuning, robot-hands, ai-agent-orchestration and gaswholesalers were invisible
  entirely) · **what the guard actually refuses: 67 pages / 16 sites**. Filed figure was 61 / 10.
  They corrected the bug file's Verify block to carry attached-vs-published and the never-shipped
  arm, superseded rather than deleted the old number, and are NOT retriggering a mid-flight
  council round for it (directional evidence, not load-bearing on the design) — they carry the
  correction into round 2 or record it against the verdict. **The root they name is the keeper:
  the census that SIZES a bug was never the predicate the FIX uses, and the two drifted.** Folded
  into the fleet memory `a-census-must-be-derived-from-the-code-it-models` as its fourth form,
  with the go-green-early twin, both credited to the 450 lane.
- **(aa) mechanism named by the experience loop (via designblog), and it is the general form of
  what I measured:** `build_status='deployed'` records that a deploy ONCE happened, not that the
  CURRENT components have been deployed — nothing invalidates it when a component is added later.
  Our seven: `deployed_at` 00:08–00:19Z against components 09:34–09:54Z, a nine-hour gap, stored
  HTML holding the controls and served bytes holding none. **Their detector consequence is the
  keeper: a rule that reads STORED `rendered_html` goes FALSE-CLEAN exactly when the divergence
  opens** — theirs caught all seven at 07:41:01Z and passed them at 09:54Z. They are fixing it
  (refuse clean where the newest component postdates `deployed_at`) and re-running past passes.
  Appended to the existing LANDMINES entry ("A data repair RACES the sweep that publishes it")
  rather than filed as a competitor, with the query. **Also theirs, and worth carrying: rule B had
  detected all seven hours BEFORE the owner's critique — the gap was routing and visibility, not
  detection.**
- **copyonline's prose glossary confirmed COMPLIANT** with the owner's glossary ruling by the
  designblog lane's reading too; my plan-review check stands as the control.
- **OWED: advertise.co.uk has the same missing tools-hub shape** designblog just fixed via
  migration 726 (their 2.5/2.6 nav fix). Ours to do when we next touch advertise — the 444
  critique's "no tools nav link/hub on any remake" class.

### (bb) 2026-09-03 ~10:4xZ — my layout-name-in-tags CAUSE is REFUTED by the 445 lane; the census survives, the causation does not

I claimed in (z), in the runbook and in a WRONG_CALLS entry that the three remakes landed on
`magazine-grid` **because** the classifier wrote `magazine-grid` into `industry_tags` and the
matcher scored it. **That is wrong, and they disproved it three ways** (all re-runnable, and they
rebuilt the scorer to check their reading — the replica reproduces the system's own recorded score
EXACTLY on 29 of 30 sites):
- not via tags — magazine-grid's own `industry_tags` are `publication,news,blog,opinion,long-form,editorial`, **no self-name**;
- **not via the description bonus** — the path I never checked and my best shot: the bonus tests for
  `" magazine-grid "` / `" magazine grid "` inside the description, and the description opens
  *"Publication layout with featured article, main 2/3 + 1/3 sidebar grid…"*. Neither string present.
  Same for `affiliate-hub`, whose description says "affiliate CTA", never "affiliate hub";
- not via category — that bonus fires on `editorial`, which those sites carry independently.
**So the tag is INERT in the scorer.** My census (12 sites carrying a layout name in `industry_tags`)
stands as a fact and as a prompt-hygiene problem; it is not the cause of anything.

**What is actually happening, and it is worse.** Seven sites resolved to `magazine-grid` on
**exactly one matched tag** — `editorial-publication` — at the identical score 3.05 (tags 2.30),
addressing **7–10% of each site's declared identity** while the other nine or ten tags matched
nothing: advertise, apis.uk, designblog, homegarden, oufe, relojistas, websitepromotion. Eight more
sit on `tool-portal-light` on three tags, score 8.31, ~23% coverage — **seotools is one**. And four
sites are recorded BY THE SYSTEM as `tags 0.00` with `layout_source: library_match` (webdesign.uk on
brochure-bold; farmerinsurance / garden-tools / vetcomparison on industry-hub) — a layout matching
NONE of the site's tags, recorded as a successful match, because `IsFallback` requires the TOTAL to
be zero and the category/description bonuses lift it above zero unaided.
**The right instrument is COVERAGE — the fraction of a site's own identity its layout addresses —
not the score, which mostly measures how many tags a site happens to have.** Our three sites are at
7–8%: not a compromise, a near-miss dressed as a match.

**Two things this lane gains:**
1. **Where the resolved layout lives** (I looked and failed): `sites.style_collection_id` →
   `style_collections.css_theme_id` → `css_themes.layout_id` → `layouts`. Better still,
   `site_specs` `aspect='resolved_composition'` records `layout_name`, `lineage.layout_source`,
   `lineage.layout_candidates` and the matcher's own score in a `reasoning` string — 33 sites have
   one. **That is how we verify №5's layout, and it retires the "unverifiable" note in §5.**
2. **⚠ WATCH FOR №5: copyonline could become the EIGHTH member of the `editorial-publication`
   cluster.** It is a content/editorial site; if its classification lands that tag it goes to
   magazine-grid at ~7% coverage, i.e. exactly the sameness the brief's `soft-editorial` direction
   was meant to avoid — and the direction cannot prevent it, because a brief cannot name a layout
   (z). **Read `resolved_composition` the moment composition runs, before the release rerender.**
   Their design point, adopted: the unit of a library gap is a CLUSTER, not a site — seven per-site
   items are a queue nobody drains.

### (cc) 2026-09-03 ~11:0xZ — the layout problem is a TAG-VOCABULARY gap, measured: 9 of 18 layouts are reachable by ZERO sites

Prompted by the 445 lane's never-used split (`soft-editorial` scores above zero on 27/33 sites but
only on the same-scheme bonus, never on tag fit). I asked the reachability question directly —
**for each layout, how many sites emit ANY of its `industry_tags`?**

| reachable by | layouts |
|---|---|
| 23 / 21 sites | `tool-portal-dark` (14 tags), `tool-portal-light` (11 tags) |
| 5 / 4 / 3 / 2 / 1 | magazine-grid 5 · technical-precise 4 · social-lobby 3 · brochure-formal 2 · brochure-bold, ecommerce-storefront, high-energy 1 each |
| **ZERO** | **`soft-editorial`, `affiliate-hub`, `comparison-aggregator`, `docs-sidebar`, `industry-hub`, `media-grid`, `portfolio-kinetic`, `tool-first-landing`, `utility-tool` — 9 of 18** |

[MEASURED 2026-09-03: `layouts.industry_tags` ∩ every current `classification.industry_tags`.]

**The mechanism of the whole sameness problem, and it is not a library gap.** The two winners are
the only layouts whose tag vocabulary has been grown into the language the CLASSIFIER actually
speaks — `interactive-platform`, `tool-portal`, `practitioner-platform`, `editorial-publication`.
Every other layout still carries 4–6 tags in a *designer's industry* dialect the classifier never
emits: `wellness, lifestyle, bakery, artisan` · `law, consultancy` · `boxing, martial-arts` ·
`podcast, gallery`. So 15 of 33 sites funnel onto two layouts not because the library is short of
designs but because **only two of them are addressable**. Adding an 18th layout without tags in the
emitted dialect produces a 10th unreachable one — 445 suspected this; it is now measured.
**The deeper design fault: a layout's tags name an INDUSTRY, while what makes a layout right for a
site is its FORM (reading-first, long-form, tool-dominated). The two vocabularies answer different
questions, and the classifier mostly emits shape words.**

**Consequences for this lane, both immediate:**
1. **`soft-editorial` is UNREACHABLE for copyonline — zero sites emit any of `wellness, lifestyle,
   bakery, artisan, personal-brand, long-form`, and a B2B copywriting site will emit none of them.**
   The layout is aesthetically right (warm, reading-first, serif) and tagged for a market we are not
   in. My template-v2 line "prefer the nine unused layouts" was aiming at the nine that CANNOT BE
   REACHED. Corrected in the runbook.
2. **FALSIFIABLE PREDICTION for №5, recorded before the build so the outcome can disconfirm it:**
   copyonline will resolve to **`tool-portal-light`** (it ships four browser tools, and that layout
   carries `tool-portal`/`tools`/`interactive-platform` plus `editorial-publication`), with
   `magazine-grid` the alternative if the classifier leads with editorial words. **Not
   `soft-editorial`, whatever the brief says.** Read `site_specs` `aspect='resolved_composition'`
   the moment composition runs and record which way it went — either outcome is evidence for 445.
- **(cc) closed by the 445 lane: the fix was SPECIFIED IN APRIL 2026 AND NEVER BUILT.** Migration
  `103_site_design_planner.sql` (lines 175–215) defines the `resolved_composition` lineage contract
  and specifies three things that do not exist: **`layout_match_score` "(float 0-1) — tag-overlap
  score for chosen layout"** — a NORMALISED score, i.e. exactly the coverage measure they rebuilt,
  and **0 of 33 rows carry the key**; a **threshold** (its own worked example reads *"scored
  utility-tool=0.82 above threshold 0.5"*) — absent from the code; and
  `layout_source: "needs_new_layout_candidate"` as a recordable value — **never written once**
  (31 `library_match`, 2 `library_fallback` — ⚠ **their current-rows figure; re-derived here over
  ALL rows including superseded, 2026-09-03: 36 rows, 34 `library_match` + 2 `library_fallback`,
  and `needs_new_layout_candidate` STILL never written.** `site_specs` retains superseded versions
  in-table under `is_current=false` rather than moving them to an archive, so `is_current` is this
  table's rolling-window trap; the claim strengthens rather than breaks). So the gap signal was left firing on the only
  condition available, total score zero, which the category/description/scheme bonuses lift above
  zero unaided — hence the four sites recorded `tags 0.00` as a successful `library_match`.
- **And the classifier's live prompt promises a mechanism the code cannot keep**, verbatim: *"If no
  existing tag fits this site well, coin a new one using the same style — an unmatched tag will
  trigger a library-growth review work item rather than silently fail."* It does not.
  **[MEASURED 2026-09-03, theirs] 188 of 216 distinct terms the classifier emits across 33 sites
  (87%) match no layout.** The model obeys its instruction perfectly and its coined tags vanish;
  four terms decide everything. **Fleet lesson: a prompt that tells a model its output will be
  HANDLED is a promise the CODE must honour, or the model's compliant work disappears silently and
  at scale.** Written to memory.
- They adopt the **"industry dialect vs FORM"** framing with attribution, and it gives the archetype
  a hard acceptance test rather than a taste judgement: **a new layout is only real if the classifier
  ALREADY emits tags it can match** — `content-hub` (4 sites, matched by nothing) is a candidate
  anchor; `long-form` sits in magazine-grid's tags and is matched by nobody.
- **Their correction, which this lane must heed too:** a subagent report told them migration 689 was
  never applied and the theme-kit branch was dead code. They checked instead of repeating: **689 IS
  applied**, `theme_kits` and `page_archetypes` exist, 4 kits seeded; the branch is inert for a DATA
  reason (0 adoptions), not a schema one. The report had read a **concept-register entry written
  before the migration was applied and never updated** — the exact landmine this tree already
  carries. **We cite the register too: verify a register status against the mechanism before
  repeating it.**
- **№5 composition check upgraded** to capture `layout_candidates` + the `reasoning` score, per
  their request — a "not soft-editorial" outcome is not enough; the score decides what it means.
- **(cc) a correction OFFERED and a check that paid off sideways.** The 445 lane corrected its own
  "one `needs_new_layout_candidate` row fleet-wide, ever" to **2** — it had queried only
  `site_work_items` and missed `site_work_items_archive` (33,350 rows since 2026-02-22; corrected
  denominator **63,007 work items ever**). Both rows are the same degenerate arm
  (`"fallback — no classification tags"`), so their conclusion is unchanged and sharper: the number
  of times this mechanism has ASSESSED THE LIBRARY and reported it short is still **zero**, now
  zero of two. **I had not carried that figure** — what I carried was the different `layout_source`
  claim above — so nothing of mine needed retracting. **But their check applies to my claim one
  table over, and running it improved mine:** `site_specs` keeps superseded rows in-table
  (`is_current=false`), so my "33 rows" was the current-rows count; over all 36 rows the answer is
  the same and the population is bigger.
  **The generalised check, worth more than either instance: before quoting an ALL-TIME count, ask
  where the closed/superseded copies went.** Two different answers on this estate —
  `SELECT table_name FROM information_schema.tables WHERE table_name LIKE '%work_item%'` (4 tables,
  one a real archive) versus `site_specs`, which archives nothing and versions in place. A census
  that names neither is a census of the present tense.

### (dd) 2026-09-03 ~11:0xZ — a discovery sweep re-minted advertise's stuck page ahead of 437's fix, and the queue is advancing not starving

- **Queue check, because 90 min with no publication deserved evidence rather than patience:**
  pending sites **21 → 16**; websitepromotion **19th → 11th**, seotools **20th → 14th**. We are
  advancing. Still 7 rerenders `triaged`, 0 of 7 pages published, 8th build unclaimed since 09:05Z.
  Sites ahead carry big backlogs (finetuning 53, idea.uk 45, relojistas 33). **Not intervening.**
- **advertise gained ~57 items 10:22–10:36Z** — a full `discovery` sweep (rerenders 23, content_rewrite
  13, needs_content_planning 10, needs_content_page 7, imagery/assets/acceptance the rest).
- **⚠ The stuck regulation-map page was RE-MINTED AUTOMATICALLY at 10:34:50Z** — `e75f5880`
  `needs_content_page`, `triaged`, source **`page-rerender-empty-skip`** ("Page … has 0 component
  rows — a rerender cannot help; build it"). **This is the re-mint 437 predicted, by a DIFFERENT
  route than the one they named (not `reconcile_site_plan`), and it has arrived BEFORE their Go
  half `a0044e73b` is live** (cluster still v1.0.1356). It will dispatch on its own and, on their
  account, fail the same way. **Told them, did not touch it.** The general point handed over: the
  window between an automatic re-mint and their roll is a hazard the other five sites share and it
  is gated by nobody's restraint — which makes their roll more urgent than a normal fix.
- 445's simulation of a `content-hub-tools` archetype against the live fleet **partly disconfirms
  their own remedy**: it breaks the seven-site cluster up (websitepromotion 8%→28%, advertise
  7%→16%, relojistas 10%→22%, homegarden 8%→17%) but **designblog stays at 6% and apis.uk at 8% —
  still winning on a single tag**. So an archetype fixes the SAMENESS and not the FIT for all of it,
  and which of our sites benefits is decided by the classifier's tags, upstream of any brief we
  write. Their tag set was invented, so they are moving to choosing it by simulation against the
  fleet rather than by taste — one artefact serving both halves of 445.
- **Our composition read gains a fifth field** at their request: the site's own
  `classification.industry_tags`, which decides whether a future archetype would rescue copyonline
  or leave it at ~7%.
- **(dd) answered by 437, and advertise is SAFE — do not panic when the re-mint fails.** `e75f5880`
  carries its own key `needs_content_page:288baf25-…` (page UUID), while the two terminal rows are
  `needs_page:uk-advertising-regulation-map` and `page_rerender:…`. The two-strike arm counts
  terminal siblings on the SAME key in a rolling 7-day window, so it sees zero and starts fresh
  (attempt 0/3). It will fail on the unfixed writer; **one failure on a fresh key does not brand
  it.** advertise is at 0 keys with ≥2 terminal siblings. **Expected and harmless — it is the
  second and third failure on the same key that matter.**
- **Three other sites are NOT safe** [MEASURED by them 2026-09-03 ~11:00Z]: farmerinsurance.uk
  **21 of 24** keys, remortgagecalculator.uk 6 of 24, loanzy.uk 3 of 15 — 30 keys where the next
  automatic re-mint is born branded and stuck. Not ours; recorded because our sites share the
  producer. ⚠ **Their own caveat, and it is the important part for anyone quoting it: the
  threshold counts a ROLLING 7-day window, so it DECAYS.** A key at 3 siblings today falls below 2
  as they age out and becomes re-mintable with nobody doing anything, and climbs again on every
  sweep. **It is a snapshot of a moving quantity, not a backlog — re-run 437's §Verify query rather
  than citing the number.** What does NOT decay is an `unresolved` row once written.
- **Two corrections of theirs, one caught by our report:** (1) `reconcile_site_plan` did NOT re-mint
  our page — `fileBuildAskForEmptyPage` did (`rerender_single_page_action.go:1276`), a different
  producer, item type and key shape. **"Which producer re-mints this page" is not something to
  reason about; several can, and they do not share a key.** (2) Their "born via a raw INSERT subject
  to neither ladder nor two-strike" described `reconcile_site_plan` only; our item goes through the
  shared `writeWorkItem` door and IS subject to both. They also declined to hold the item type
  themselves — a hold on a shared seam is a config change wanting its own review — and surfaced the
  release urgency to their owner instead, which is the right division.
- **Deliverable state 11:06Z, unchanged for 90 min:** 7 rerenders `triaged`, **0 of 7 published**,
  8th build still `triaged` since 09:05Z. Queue advancing (we moved 8 and 6 places), not starving.

### (ee) 2026-09-03 11:39:14Z — OWNER INSTRUCTION EXECUTED: the classifier now reads the register (migration 734), and it carried a bigger fix with it

Owner: *"please fix the classifier to read the register."* RFC_037's DB half shipped long ago
(511, 194 rows); its classifier half had never been started. Both halves are now live.

- **(A) The larger change, found while wiring the smaller one.**
  `classify_and_extract.config.input_fields` is a **strict allow-list**
  (`ai_actions.go` `extractDataForAiAgent` → `datahelpers.ExtractFields`) and **`layout_taxonomy`
  was never in it** — although `read_layout_taxonomy` runs, populates `collected_data` correctly
  (verified: the key is present in real runs), and the prompt references it. So the taxonomy was
  **fetched and dropped at the template boundary**. MEASURED at the rendered artefact
  (`llm_call_log` `a01d9e76`, 2026-09-02 20:15:25Z) the model was shown:
  `Current library tags (match these…): null` and `The library currently has <no value> active
  layouts`, then told to coin a tag if none fits. **That is the mechanical cause of the 445 lane's
  87%-unmatchable finding: the arithmetic of an empty list, not a vocabulary mismatch.** They
  verified it independently at the rendered prompt before agreeing, and asked me to carry the fix.
  **A step built in April has never once delivered its payload — the next classification is NEW
  behaviour, not a restoration.**
- **(B) The register input.** New `read_positioning_register` step (`query_database` — reuse, no Go),
  `output_field: positioning_register`, between `read_layout_taxonomy` and `classify_and_extract`.
  Emits ADVISORY prose: recorded position, audience, mode, stance, must-nots, and each recorded
  sibling boundary; defers to a Pre-Defined Mission in its own words (ruling answer 3).
- **Inert by construction** (ruling answer 4): '' unless the row is not `exclude_from_build`, has a
  non-empty `proposition`, and does not self-declare `direction unassigned`. **TESTED per site
  before writing the migration** — copyonline 2,602 B · gamedesign.uk 1,677 B · webdesign.uk,
  farmerinsurance.uk, advertise.co.uk, seotools.co.uk all EMPTY. The query always returns exactly
  one row, so the variable can never render `<no value>`.
  ⚠ **The third guard exists because of a real live case:** entry_code **L9** ("loan brandables
  (direction unassigned, deliberately)") sits on **6 domains including `webdesign.uk`**, a design
  site. Without the guard the classifier would have been told a design site is a loan brandable.
  **That is a register DATA defect this lane owes a fix for, not something to leave to a string
  match.**
- **Guards, dry run, rollback**: 17 `RAISE EXCEPTION`s; anchor uniqueness; refuses if already
  applied; asserts exactly one new template variable and unchanged if/else/end balance; verify
  block asserts both new inputs, all four originals, the step, the `output_field`, the rewire and
  three prompt survivals. Dry-run with `COMMIT`→`ROLLBACK` passed before applying. ROLLBACK file
  includes a **variant that keeps the taxonomy fix**, so reverting does not silently restore a
  known defect.
- **Council: `Council-Submitted: f0ad8366-d489-440d-8a3b-59b000de0ff2`** (advisory; migrations are
  in scope). Applied without waiting per the 2026-07-29 ruling that review here is after the fact.
- **COVERAGE, stated honestly: 194 register rows, only 17 with a `sites` row, and NONE of the four
  2026-09-02 remakes has an entry.** So the mechanism is live and mostly inert. **Writing register
  rows for the built and briefed sites is now this lane's highest-value follow-up** — the estate-wide
  inventory the widened ruling wants is still an open ask on the owner, but coverage is now a data
  task rather than an engineering one.
- **NOT YET PROVEN at a real classification** — nothing has classified since 11:39Z. **copyonline is
  the canary and it has an entry**: at its release, read the rendered prompt and confirm (1) the
  library tag list is populated, not `null`, and (2) the register block is present. Told the 445
  lane the apply time so they can date their before/after boundary.

### (ff) 2026-09-03 ~12:0xZ — copyonline brief REVISION 2 (owner), and 734's council REVISE answered

- **OWNER DIRECTION**: AI is the biggest force on this audience, threat and opportunity at once, and
  *"AI is also pretty bad at it"*; a lead-generation page in the nav carries the conversion aim so the
  rest of the site can stay as planned; the CTA points at a directory (a randomised listing at first),
  later at the owner or a lead buyer; webdesign.uk may route leads here.
- **He asked whether he had missed AI in the brief. He had not — and the absence was MINE.** Measured:
  exactly **3** `\bAI\b` mentions in the original, ALL defensive (a stance against pretending AI needs
  no editing; a must-not against hyping AI tools; one *aspirational* page listing AI among four
  contested questions). No page about it, no positioning. **My own fire direction said "no hype about
  AI writing" and the writer obeyed.** Said so plainly rather than letting him think he misread it.
- **Revision applied 11:56:00Z** (`0d9311fb` current; brief-writer original `c1b5c41d` RETAINED,
  superseded — retire-then-insert as SEPARATE statements per the runbook trap; guarded on exactly-1
  current row AND on the review item still being `needs_human_review`, i.e. refuses if the build was
  released underneath me; verify asserts 2 rows, ≥3 AI-titled plan items, exactly 1 `lead-route` page).
  24 → **28** plan items.
- **The design decision worth keeping**: the owner's two aims conflict, and his own answer (one page,
  in the nav) is the right one — implemented as `kind: lead-route`, plus a must-not that stops it
  bleeding into the guides. **The randomised directory is STAGE 2 and deliberately NOT planned**: we
  have no copywriters to list, so it is the 444 empty-listing class, and claiming a panel we do not
  have is a claim the site cannot stand behind. Stage 1 routes to the owner and says so. Both staging
  questions went into `open_questions` as HIS calls, not mine.
- **AI framing guarded both ways**: new must-nots forbid hype AND sneering, because neither is
  checkable, and require every claim about machine output to be shown in a worked before-and-after.
- **734's council verdict: REVISE (editquality, 6 abstained, 11:48:39Z).** The objection: the rationale
  claimed the migration inserts a template variable but **the sketch showed only three `jsonb_set` calls,
  none touching `prompt_template`** — so on my evidence the field would be extracted and never rendered.
  **The seat was RIGHT about my evidence and WRONG about the code**, and that distinction is the whole
  point: my sketch under-described the change. Verified at the live row that
  `{{.positioning_register.block}}` IS in the prompt (and survived 445's later migration 735 intact,
  along with my `input_fields` and the chain). **This is the same defect class the migration itself
  fixes, one field along — a good catch on a plan I wrote carelessly.**
- **The seat also asked the better question — blast radius across the 25 agents sharing this
  allow-list. I ran the census and REFUSED TO REPORT A NUMBER, because the obvious query is invalid**:
  a naive `{{.root}}`-vs-`input_fields` diff counts loop-scoped fields inside `{{range}}` blocks as
  missing roots (worked example: `build-site-planner.plan_site` has
  `{{range .existing_pages}}- name: {{.name}} | page_type: {{.page_type}}`, and those are element
  properties, correctly absent). The only evidence that convicted the classifier was the RENDERED
  prompt. Handed the fleet question to the `--template-input-fields` lint (WFA-024) which can model
  scoping. **I nearly published a 25-agent figure that would have been wrong in the same way three of
  today's other corrections were.**
- Round 2 resubmitted on the same correlation `f0ad8366` with the corrected sketch, the live
  verification, the disclaimed census, and a new risk the objection implies: the ROLLBACK must remove
  the prompt variable too, or reverting would leave a reference to an unsupplied field rendering
  `<no value>` — the exact failure being fixed.

### (gg) 2026-09-03 12:4xZ — ALL EIGHT TOOLS BUILT; the sectionless variant repaired CLEANER than the shelled ones

- **8 of 8 complete.** Seven on seotools 09:34–09:54Z, and websitepromotion's
  `tool-channel-prioritiser` at **12:36:06Z** — it waited ~3.5 h for the dispatcher, not for anything
  wrong. Every one adopted its EXISTING page at its EXISTING url (`page_adopted: true`), no duplicate
  page rows minted.
- **The 8th is the interesting one, and it repairs BETTER than the seven.** Because it was the
  SECTIONLESS variant it never got a prose shell, so its page now carries **exactly one component —
  the tool, 29,859 B — with no leftover `generic-text-block` and no position-2 collision.** The seven
  seotools pages each carry the tool AND the shell's prose at the same position 2.
  **So the shell is not merely cosmetic damage: it leaves debris the repair cannot remove.** The
  variant that "parked" (7 HITL items and a needs_content_page, which read as worse at the time)
  turns out to be the one that heals cleanly. Told the 450 lane — it strengthens their case for
  holding empty-sectioned tool pages, from the opposite direction to the one they argued.
- **Still unpublished: 0 of 7 seotools pages, and the 8th is `planned`/`deployed_at` NULL.** The
  7 `page_rerender` items remain `triaged`. **Checked rather than assumed** — seotools is unlocked,
  `deployed`, has 43 dispatchable items and ZERO claimed siblings, i.e. fully eligible; it has simply
  moved 20th → 14th → **5th** in the pending-site queue behind relojistas, remortgagecalculator,
  dartsonline and idea.uk. Advancing, not starving. Its 7 rerenders are not even its oldest item
  (`nav_drift` 09:34:09.482Z is), and the loop takes ≤8 items per pass, so 43 items is several passes.
- **v1.0.1358 rolled 12:06–12:07Z.** Nothing of ours rode it (all DB config). 450's guard has not
  taken effect yet either — **0 `owned_page_review` rows carry `refusal_class='tool_pending'`** and
  the 19 `unbuilt_internal_link` rows touched since 12:00Z are all still `triaged`, so their
  `wont_fix` wave has not begun.
- **437: HOLD REAFFIRMED and the goalposts moved honestly.** v1.0.1358 DOES carry their `a0044e73b`,
  but their own go/no-go probe now returns FALSE post-roll — the writer prompt still shows the flat
  `"branches": "..."` exemplar and builds still fail identically. They withdrew "fire after the roll"
  and published a full elimination list (ancestry with a negative control; both literals in
  `/proc/1/exe` with present/absent controls; no service skew; migration intact; helper correct in
  isolation) plus their leading hypothesis (a schema serialised as a JSON string, where their helper's
  first guard is stricter than `extractArrayItemFields`' entry condition) — and called that asymmetry
  their own design error. **Our `e75f5880` stays untouched: one failure on a fresh key is harmless;
  it is the second and third on the SAME key that earn the sticky brand.** Their lane doc:
  `docs024_key_docs_latest/bugfix_437_writer_prompt_nested_shapes/HANDOFF_2026-09-03_continue_here.md`.

### (hh) 2026-09-03 ~12:5xZ — PRE-REGISTERED: copyonline's layout prediction, three arms, recorded BEFORE it resolves

The 445 lane shipped the `content-hub-tools` archetype (migration **736**, register **DES-087**,
applied 2026-09-03), choosing its tags by simulation against the live fleet **and against this lane's
`CONTRIB_2026-09-03_seventeen_remaining_remakes_for_tag_simulation.md`** — 7 of 14 proxies land on it,
including both Christmas twins, which they note is CORRECT because layout is FORM and twin
differentiation is RFC_037's job, not the layout's. Tags:
`content-hub, interactive-tools, editorial-publication, long-form, long-form-content, editorial-guides,
guides, research-publication, content-platform`.

**My prediction for copyonline, now with three arms. Written down before the answer exists:**
1. **`tool-portal-light`** — my original call (it ships four browser tools; that layout carries
   `tool-portal`/`tools`/`interactive-platform`/`editorial-publication`).
2. **`magazine-grid`** — if editorial words lead. Now a SURPRISE if it happens: `content-hub-tools`
   ties it on category and beats it on any second editorial tag.
3. **`content-hub-tools`** — if the classifier emits `content-hub` / `editorial-guides` /
   `long-form-content` / `research-publication` alongside `editorial`.
**Not `soft-editorial`, whatever the brief says** — unchanged, and the reason is unchanged
(zero sites emit any of its six tags).
**This will be the first classification ever to run with the taxonomy actually rendered (our 734) AND
with the form-over-industry guidance (their 735), so it is a canary for three changes at once** — which
is also its weakness as evidence: a single run cannot attribute an outcome among three changes. Send
all five fields (`layout_name`, `lineage.layout_source`, `lineage.layout_candidates`, `reasoning`,
`classification.industry_tags`) so they can separate them.

**MEASURED FOR THEM, because they had not proven it: `76db94fc7` did NOT ride v1.0.1358.**
NUL-split probe on the running chassis pod with both controls in the same pipeline —
`enforceListingItemSources` = **2** (present control, 444's live gate), `zzz_invented_absent_control`
= **0** (absent control), and their three symbols `layout_match_score` / `coverage_pct` /
`fit_threshold` all **0**. The commit IS an ancestor of HEAD, so this is the "in HEAD but not in the
binary" gap, not a missing commit. **Consequence: `lineage.layout_match_score` will NOT appear on
copyonline's row; coverage stays regex-out-of-prose; the old shape is expected and is not a
regression.** Told them.

**DECISION (ours): do NOT re-compose the live remakes yet.** Their simulation says advertise would go
7→16% and websitepromotion 8→28% coverage if re-composed. Both stay as they are. Reasons: fix-forward
is the estate's posture; re-composition on a live site is a design change with its own blast radius;
and copyonline is about to tell us how a FRESH classification behaves with a real tag list, which is
strictly better evidence than guessing from a simulation. **The condition that would change this:
copyonline lands on `content-hub-tools` (or another good fit) with a clean served result — then
re-composing advertise and websitepromotion becomes an evidenced improvement rather than a hope.**
- **(hh) confirmed by a SECOND instrument.** The 450 lane supplied the running chassis's own stamp —
  `v1.0.1358` reports `build provenance git_commit d0252fd4dab2a3a583d1cc8eb8e1b26e9c422d85` — which
  makes ancestry available, the instrument CLAUDE.md prefers because it has no timezone and no
  binary-probe caveats. Ancestry against that stamp, with a negative control:

  | commit | lane | in the running build? |
  |---|---|---|
  | `76db94fc7` | 445 fit scoring | **NOT an ancestor** |
  | `587666be8` | 450 class guard | IS an ancestor |
  | `6525b45ae` | 444 listing gate | IS an ancestor |
  | `a0044e73b` | 437 writer fix | IS an ancestor |
  | `HEAD` | control — newer than the build | NOT an ancestor ✔ |

  **Both instruments agree on all four.** The NUL-split symbol probe said 445's three symbols were
  absent and 444's `enforceListingItemSources` present; ancestry says the same. **Record the stamp:
  reading `d0252fd4d` once turns every "did my fix ship?" question into `git merge-base
  --is-ancestor` — cheaper and less error-prone than probing per symbol.**
- **450's guard is therefore genuinely LIVE and their zero is not idleness** — they re-powered the
  test on ACTIVITY rather than wall-clock after our exchange (199 orchestrations, 234 item
  transitions, 63 component writes since the roll, zero to a shell page; historical base rate
  275/17,205 = 1.60% predicts 1.01, observed 0, **p ≈ 0.36 — still uninformative**, needs ~125 more
  fleet writes). They said so rather than banking a good-looking number. **The improvement is in the
  instrument, not the result.**
- **They corrected their own argument on our sectionless finding and attributed it** — "no shell is not
  harmless" is right about the cost BEFORE repair and wrong about the cost AFTER it; the gate saves
  cleanup as well as face. **OWED BY US: the served-body reading of the position-2 collision once the
  seotools rerenders land** — whether a visitor actually sees the tool and the leftover prose fighting
  for the slot, or whether the renderer's ORDER BY happens to put the tool first. The database cannot
  answer that; only the body can.
- Their `[UNMEASURED]` case is closed from our side (none of our 8 was a mis-typed tool page); the sole
  remaining fleet candidate is `idea.uk` `/report.html` — six components, no tool. Not ours.
- **(hh) cause of the not-rolled, from 445 before they handed off — a shape worth keeping.** Their
  `76db94fc7` was committed AFTER the pods started at 12:06:47Z, because they were running mutation
  proofs when the build was cut. **The right trade** (an unproven predicate change fleet-wide is worse
  than one roll later) and it produces a specific, recurring confusion: a commit that is in HEAD, was
  intended for "the build you were told about", and is genuinely absent from it. **"It should ride the
  build" is a claim about TIMING, and timing is exactly what nobody can see on a shared tree.** They
  rewrote their handoff to read MEASURED: NOT ROLLED, so a fresh session does not file the old row
  shape as a regression. They confirmed our probe with their own (`enforceListingItemSources` 2,
  `layout_match_score` 0, `weak_tag_fit` 0, absent control 0).
- **Attribution, refined by them:** one run cannot separate 734/735/736, but the five fields make it
  TWO-THIRDS separable — `classification.industry_tags` alone says whether 734+735 changed what the
  classifier SAYS; `layout_name` plus the reasoning score say whether 736 changed what the matcher
  DOES with it. Good enough to act on, and worth stating that way when the result lands.
- **445 has handed off.** A fresh session inherits their handoff and will read our copyonline result.
  Our re-composition gate stands unchanged and they endorsed it.
- **Queue at 12:50Z**: seotools still 5th (relojistas 9, remortgagecalculator 39, dartsonline 6,
  idea.uk 44 ahead); dispatcher visited websitepromotion 12:31Z, idea.uk 12:49Z. 7 rerenders still
  `triaged`; the 8th page is `planned`/undeployed. **The URL watcher timed out at 3 h with 0 of 8
  serving — honest but useless, so it is re-armed with a 45-minute heartbeat**: a watcher that only
  speaks on success cannot distinguish "not yet" from "dead", which is the same silence-is-not-success
  trap that cost me 25 minutes on the brief this morning.
- **(gg) tail — a one-line caution turned up a FALSE RECEIPT, and the lesson is the keeper.** I
  suggested the 450 lane separate "mis-typed page" from "page whose tool was removed" before spending
  their last `[UNMEASURED]` case on `idea.uk /report.html`. They did both checks and found something
  better: `page_component_history` shows **no tool-level row has ever been attached**, so it was never
  stripped — and its six components include `report-request-form`, a form at SECTION level. Probed at
  the body with a real tool as the control in the same run: `idea.uk/report.html` 1 form / 8 inputs /
  100,896 B against advertise's ab-test-calculator 1 form / 11 inputs. **Indistinguishable from a real
  tool by their own test.** So their council `[UNMEASURED]` risk closes at **zero pages harmed** — and
  refusing that page's generic rebuild is CORRECT, because a rebuild would clobber a working form.
- **The finding: their receipt was asserting something false.** It told an operator the generic builder
  "would publish prose about a tool that is not there" — false for this page, where the tool IS there
  at the wrong level, so an operator would go hunting for a missing tool and find a working one. They
  rewrote the summary to state the measured fact (no tool-LEVEL component), leave the consequence
  conditional, and tell the reader to check the served body first; the test that pinned the old wording
  was asserting the falsehood, so it now asserts the opposite.
- **The transferable shape, in their words: "the predicate was right, and every sentence I wrapped
  around it was an untested inference."** A check's PREDICATE gets measured; the summary, the `fix`
  advice and the consequence do not — and the operator acts on the sentence, not the predicate.
  `no tool-level component` → `mislabelled` → `prose about a tool that is not there` is three
  unmeasured steps. **The cheap check is the one I keep being asked to run on my own claims: what ELSE
  could produce this same reading?** Belongs in 016b §9; theirs to write up, not ours to duplicate.
- **Their caution back, accepted with one clarification:** a build stamp is trustworthy PER SERVICE, so
  a lane reading the CHASSIS stamp must confirm their code lives in the chassis. Checked for 445 —
  `fork_theme_composition.go`, `install_site_composition_action.go` and
  `resolve_composition_layout_action.go` are all `platform/orchestration/actions/`, i.e. chassis — so
  the chassis stamp was the right instrument for their commit and the not-rolled result stands.

### (ii) 2026-09-03 13:0x–14:0xZ — the repair SURVIVED but its component LINK was destroyed; and my own watcher's test was unsound

- **Six of the seven seotools tool pages took generic writes 13:05–13:24Z** (`needs_content_page` →
  `page-build-handler`, minted by our own repair wave's downstream). The 450 lane spotted it and
  reported their guard had not refused and had left no receipt. **Their alarm was right about the
  writes and wrong about the outcome, and so was mine.**
- **MEASURED at rows AND bodies.** `save_page_sections` deleted all three rows at 13:05:14.988Z and
  re-inserted ~80 ms later. **The tool's `rendered_html` SURVIVED** (17,938–23,953 B per page) and all
  six serve working tool controls now, with instance-scoped ids
  (`id="c-tool-robots-txt-tester-fetch-domain-input"`), 78–85 KB bodies, 6 scripts.
  **What died is `component_id` — NULL on exactly those six tool rows.** Control group with the id
  intact: the seven suggester-built tools plus `tool-serp-snippet-previewer`, the one page that never
  took a generic write. The library component is untouched and still `is_active`.
- **Three consequences. The middle one is the dangerous one.**
  1. Serving is correct today.
  2. **Any predicate joining `page_components → content_components` on `component_id` cannot see these
     six** — so 450's shell census counts six serving tools as shells, and their 67-page population
     now over-reports by six. That is a SECOND instance of the census-versus-predicate divergence they
     wrote up this morning, inside their own instrument.
  3. **The real exposure is the next rebuild**: anything regenerating from `component_id` rather than
     from stored HTML has nothing to regenerate. These pages are one rerender from losing their tools
     for real, and nothing in their appearance would warn anyone.
- **MY OWN WRONG CALL, and it is the kind I have been correcting in others all day.** My URL watcher
  declared a page LIVE on `forms > 0 OR fields > 2`. **Two of the six serve a working tool with ZERO
  `<form>` elements**, so the rule passed them on a field count that a prose page with a search box
  would also satisfy — and it reported "TOOL LIVE" for `robots-txt-tester` at 13:06:41Z with
  `forms=0 fields=4`. It was right by luck. **A test that would pass a prose page cannot certify a
  tool page**, and I built it after spending the morning telling other lanes to control their checks.
  The sound test is the tool's own instance-scoped id (`c-<function>-`), which nothing else emits.
- **First alarm was mine too:** I read `tool_rows = 0` off a join and briefly believed six tools had
  been destroyed. The join was failing, not the data. **A LEFT JOIN that silently drops a row reports
  absence indistinguishably from deletion** — the check that separated them was `component_id IS NULL`
  versus `points_at_missing_row`, and it took one query.
- **(ii) FIXED 2026-09-03 14:01:10Z — the six orphaned tool rows are relinked.** Guarded UPDATE:
  refuses unless exactly 6 orphaned rows exist and each maps to exactly ONE active site-specific
  tool component (`cc.function = pc.slot_name`, `name LIKE '%-seotools-co-uk'`); byte safety asserted
  by snapshotting each row's `md5(rendered_html)`, length and digest into a temp table BEFORE the
  update and comparing after. `UPDATE 6`, all verifications passed, `rendered_html` untouched, no
  rerender triggered (none needed — the served bytes were already correct).
  **Verified at the served bodies with a SOUND test: the tool's own instance-scoped id
  (`id="c-tool-<function>-"`), which nothing else emits** — 11–19 occurrences per page on all six,
  plus 40 on websitepromotion's channel-prioritiser. **7 of 8 tools genuinely serving.**
  The eighth, `tool-serp-snippet-previewer`, is built and linked (15,335 B, `component_id` intact)
  but shows 0 tool ids because its page has never been re-rendered since last night
  (`deployed_at` 00:08:16Z) — its `page_rerender` is still queued, and its
  `unbuilt_internal_link` item failed 3× on the component floor.
- **MY VERIFY BLOCK'S OWN FALSE ALARM, and it is the third instance of one class today.** The first
  version asserted `rendered_html_digest IS DISTINCT FROM md5(rendered_html)` across ALL the site's
  tool rows and REFUSED with "the repair altered bytes". The predicate was right; the conclusion was
  false and the scope was wrong. `IS DISTINCT FROM` convicts a **NULL** digest; 4 unrelated rows
  carry one; and **a NULL digest is a normal state — 206 of 3220 rows fleet-wide, with only 34
  genuinely mismatched.** My update never touched `rendered_html` at all. **The fix was to compare
  the rows I ACTUALLY UPDATED, before against after, rather than asserting an absolute property of
  a population I had not measured.** Same shape as 450's false receipt and as my own layout-cause
  claim: a correct test wrapped in an untested inference.
- **Class NOT repaired here, and reported instead:** `idea.uk` and `loanzy.uk` each carry one
  orphaned tool row (other lanes'), and **20 orphaned `page_components` rows of all slot kinds exist
  across 7 sites** [MEASURED 2026-09-03 14:0xZ]. The mechanism: `save_page_sections`' insert takes
  `component_id` from `section.ComponentID` in the payload it is handed
  (`save_page_sections_action.go:1036-1041, 1125-1127`), so a preservation path that carries a slot's
  HTML without its ComponentID nulls the reference on re-insert. The 450 lane has made this their
  lane's first job ahead of their blocked migration; measurements sent.

### (jj) 2026-09-03 14:32:42Z — ALL 8 TOOLS SERVING, and the position-2 "collision" turned out not to be one

- **All 8 verified serving** by the sound test (the tool's own `id="c-tool-<function>-"`, which
  nothing else emits): 11–19 occurrences per seotools page, 40 on websitepromotion's
  channel-prioritiser. The last to land was `tool-serp-snippet-previewer`, whose rerender finally
  drained. Deliverable complete: the owner's ruling of this morning is executed end to end.
- **The owed reading for the 450 lane, as OBSERVED not reasoned.** Served byte offsets on all seven
  seotools pages: `hero-tool` ≈ 51,0xx → `generic-text-block` ≈ 54,6xx → the tool ≈ 61–63,xxx.
  **Order is hero → prose → tool, consistently, on all seven.**
- **And the prose is not debris — it now correctly INTRODUCES the tool below it.** Read on
  robots-txt-tester: *"The Robots.txt Tester reads that file the way a crawler does, then shows which
  of your URLs are blocked, which are allowed, and where two rules contradict each other… Paste in a
  URL or the robots.txt content itself, and the tester works through the disallow and allow
  directives in the order a crawler applies them."* That is accurate copy about a tool that is
  present.
- **So the 13:05–13:24Z generic writes IMPROVED these pages, and the inversion is the finding.**
  The original shell prose (written 2026-09-02, before any tool existed) promised a tool that was not
  there — 444's exact symptom. The rewrite happened AFTER our repair put a real tool on the page, so
  the builder had the tool in context and wrote prose that describes it truthfully. **The only
  casualty of those writes was the `component_id` reference, which is now repaired.** My earlier
  framing — "debris the repair cannot remove", "the tool and the prose fighting for the slot" — was
  wrong on both counts: they do not share a rendered position, and the prose is not leftover.
- **Corrected for the record, because I said the opposite twice:** to the 450 lane at ~12:4xZ ("each
  repaired page will serve both the tool and the prose describing it, in an order nothing declares")
  and to the owner. The order IS stable across seven pages, and "the prose describing it" is a
  feature here rather than a fault. What nothing declares is the DB `position` (two rows shared
  position 2 before the rewrite); the rendered order came out right, and I should not have inferred
  the served order from the stored position in the first place.

### (kk) 2026-09-03 15:49Z — REMAKE №5 RELEASED: copyonline.co.uk

Owner: *"otherwise the brief is good carry on"* — the release word. Per RUNBOOK §1:
- **TX1 15:49:05Z**: review item `5f9fe299` → `complete`, `approved_by='owner'`, result records the
  instruction verbatim, the brief revision in force, and that the lead-route destination is
  deliberately undecided and NOT blocking. `needs_domain_research` created `triaged`, handler
  `domain-research-classifier`, key `research_copyonline.co.uk`. Guards asserted: 1 held review item,
  a complete-and-locked site row (the `ensure_site_record` trap), a free research key, exactly 1
  current mission_brief.
- **TX2, separate and last**: site `3d965325` `test`+LOCKED → `active`, `locked_at` NULL. Verified.
- **PRE-ENABLEMENT DONE BEFORE THE PLAN, which is the whole lesson of 444/450**: scheduled task
  `copywriter-directory-discovery` created (weekly, `directory-researcher`), so the kind compiles
  entries before any page names it. ⚠ **Scope choice made deliberately and flagged to the owner: the
  query targets ORGANISATIONS (agencies, studios, marketplaces, professional bodies), NOT named
  individual freelancers.** A business listing carries no personal data, no consent question and no
  removal obligation — the exposure he raised himself — and it matches his own instinct that a
  competitor site will not complain. Widening to individuals is one line and his call. The query
  demands named organisations explicitly (migration 423's precedent: two rows archived as
  "category-shaped entity, not a named firm").
- **№5's canary duties, still owed at the right moments:** §2b fire the one-shot design discovery
  once classification exists and BEFORE the plan step (watcher armed to flag the window) · §5 chrome
  experiment in its RE-SCOPED form (supply the header vocabulary first, then pin — 37 of 39 sites
  supply zero header keys) · §5b check `imagery.sections` on the plan · and the pre-registered
  three-arm layout prediction, with all five fields owed to the 445 lane.
- **(kk) OWNER DECISION 2026-09-03 ~15:5xZ, DEFERRED BY HIS OWN CONDITION — carry it forward.**
  *"we can point to organisations in preference to myself, but don't change things if it's already
  running we can change it another time."*
  **The decision: the lead route's PRIMARY action is the listed ORGANISATIONS (the directory), and
  routing the enquiry to the owner becomes the secondary path.** That is a change to the
  `lead-route` page only; the rest of the brief is unaffected.
  **NOT APPLIED, and the reason is the interesting part.** At 15:55:30Z nothing on the site had been
  claimed and no classifier spec existed, so his condition ("if it's already running") was not met
  and the edit was free. I prepared revision 4 with the condition ENCODED AS A GUARD — refuse if any
  item is claimed, refuse if a `classification` spec exists — and at 15:56:30Z the guard **fired**:
  `REFUSED (owner condition): 1 item(s) already claimed`. One item was claimed inside that sixty
  seconds. **So the brief stands at revision 3 (`d070dca4`) and the build is undisturbed.**
  The prepared statement is `apply_r4.sql` in this session's scratch; it is one targeted edit to the
  lead-route page's `what` plus an `open_questions` entry recording the answer, and it re-applies
  cleanly whenever the change is cheap.
  **When it is cheap again:** after the build completes and before launch — the lead route was
  deliberately written so its destination can change without a redesign, so this is a copy edit on
  one page plus a rerender, not a replan. **Do it then, or hand it to the next session.**
  **Worth keeping as a practice note:** encoding the owner's condition as a SQL guard rather than as
  a judgement I make at the keyboard is what made this safe — I checked, believed I was clear, and
  was wrong by sixty seconds. A guard that refuses is better than a check that passes.
- **(kk) ordering oddity at №5's release, recorded not acted on.** Within 6 minutes of the site going
  `active` the discovery sweep created `needs_composition` + `needs_design` + `evaluate_tools`
  (15:55:06Z, `created_by='generic'`). **`site-design-planner` then claimed composition BEFORE any
  classification existed and raised its own `needs_domain_research` as a `side_effect`**, key
  `backfill_classification_for_composition` (15:56:41Z, claimed 15:58:26Z). So the classifier is now
  running from the BACKFILL item, while my release item `319d8d15`
  (`research_copyonline.co.uk`) is still `triaged` and unclaimed.
  **Different `item_key`s, so `idx_swi_dedup` blocks neither** — the site may classify twice, the
  second superseding the first. Wasteful rather than destructive, and **I am not cancelling mine**:
  the backfill exists to unblock composition, whereas the release item is the Flow A driver for the
  whole chain (research → strategy → briefing → plan), so cancelling it could stall the build.
  **The shape is today's recurring one — a producer running before its prerequisite exists.** Here
  the platform recovered by raising the missing input itself, which is the good version of it.
  Two things for whoever picks this up: watch whether classification is written twice, and note that
  **releasing a site makes the discovery sweep act on it within minutes**, which is earlier than the
  runbook's §2b window assumed.

### (ll) 2026-09-03 16:01:39Z — **734 ROLLED BACK: my register step was breaking the classifier fleet-wide**

- **The defect:** `read_positioning_register`'s query uses `$1` but the step config carried no
  `params` array. `query_database` binds parameters from `config.params` (offer-analyser's
  `load_premise`: `"params": ["site_record.site_id"]`). Runtime: `query failed: expected 1
  arguments, got 0`. **Every classifier run after 11:39Z failed at that step.**
- **Realised damage, measured:** two copyonline classifications FAILED (15:52:02Z, 15:59:09Z),
  burning one attempt on each research item. No other site attempted a classification in the window,
  so the blast radius was our own newly-released build. Both items retain **2 of 3 attempts** and
  retry at 16:22:05Z and 16:29:12Z, so the build resumes unaided.
- **Rolled back 16:01:39Z, keeping the half that works.** The register step and its prompt variable
  are gone; **`layout_taxonomy` stays in the allow-list** (a separate, working fix, and the one that
  matters most to 445's finding). Verified: step absent, chain restored, prompt no longer references
  the register, taxonomy reference and allow-list entry both intact. Snapshot taken first.
- **My test could not have caught it, and that is the lesson.** I verified the query on six sites
  before writing the migration — by `sed`-substituting a literal for `$1`. **The substitution removed
  the exact thing that was broken.** A test that edits the statement to make it runnable is no longer
  testing the statement. And my 17 guards plus the `COMMIT`→`ROLLBACK` dry run all assert the CONFIG
  IS WELL-FORMED; none asserts the step can RUN.
- **The check I owe every future workflow change: make it run once.** Fire one job and read
  `orchestration_states.current_step/status`, or execute the query through the action's own path.
  Twenty seconds, and it is the difference between "applied" and "working".
- **Owed:** tell the 445 lane (they dated a measurement boundary to 11:39Z for a change that never
  worked); the council submission `f0ad8366` round 2 describes a now-rolled-back change and needs the
  correction; and the fix-forward is `params: ["input_data.site_id"]` **only once that path is proven
  to resolve in this agent** — the classifier has no `ensure_site_record` step, so it does not have
  offer-analyser's `site_record.site_id`.
- **(ll) №5's tool sweep ran EARLY and well — and produced a live `bugs_open/447` instance the brief
  explicitly forbids.** `evaluate_tools` completed 16:01:52Z, **before classification and before the
  plan**, which is exactly the ordering runbook §2b wanted and it happened by itself: the discovery
  sweep acts within minutes of unlock. **So §2b needs rewriting — the one-shot task it prescribes is
  unnecessary; the natural ordering already does it.**
- **Six tools suggested. Three fit the site; three are a sibling's ground.**
  - FIT: `Website Brief Starter`, `Copy Brief Builder` (brief-building for clients),
    and **`Insight Injector` — "inject real facts, customer stories and proof points into an AI
    prompt so the resulting copy doesn't [sound generic]", which lands squarely on the owner's AI
    angle** and is the best of the six for this site.
  - ⚠ DUPLICATES OF `seotools.co.uk`, by name and by function: `SERP Snippet Previewer`,
    `Title Tag Scorer`, `Keyword Intent Classifier` — **all three exist as live pages on seotools
    today**, two of them built by our own repair wave this morning.
- **This is 447 exactly**: the suggester reads identity+classification and cannot see positioning, so
  it cannot honour the brief's must_not *"no duplicate of advertise/seotools/websitepromotion tools
  by name (447)"* — and here it ran with **no classification at all** (only `mission_brief` existed),
  so it had even less context than usual. It contradicts the brief's own differentiation, which says
  this site owns copy QUALITY and leaves SERP fit to seotools.
- **NOT ACTED ON, deliberately.** The owner ruled this morning that cluster duplication may be kept,
  and separately that a running build must not be changed. All six items are `triaged` on a live
  build. **Flagged to him with a recommendation rather than cancelled** — the cost is positioning
  dilution, not breakage, and the decision is his. If he wants the three dropped, cancelling three
  `triaged` `add_tool` rows before they build is cheap; after they build it is a page retirement.
- **(ll) №5's tool wave landed: 5 of 6 built, 1 LOST SILENTLY, and the cancel window is narrower than
  I told the owner.**
  - **Built (pages exist, `build_status='planned'`, NOT deployed):** `insight-injector`,
    `website-brief-starter`, **`keyword-intent-classifier`, `serp-snippet-previewer`,
    `title-tag-scorer`** — each with 1 tool component, plus a companion guide page each (10 pages).
    All five came from the LIBRARY (`library_source` set) and went through `tool-deployer`
    `deploy_tool` 16:06:23–16:08:15Z.
  - **⚠ LOST: `Copy Brief Builder`** (`tool-copy-brief-builder`, item `807779e4`). It is the ONLY one
    with `library_source: null` — a custom generation. Its item reads **`complete` at 16:05:54Z and
    its `result` contains the full `generated_html`** (styled fieldsets, labels, inputs — a real
    tool). **No page exists for it. No component. No error row.** The tool's HTML lives in a work
    item's result column and nowhere else.
  - **That is `bugs_open/218`'s shape** — *"tool-recreation's failure path discards the finished tool
    while the item reads `complete`"*, whose defect B is UNFIXED — but via a different producer
    (`add_tool` → `tool-generator`, not `needs_tool_recreation`). **Not filed as new; routed as a
    second instance.** The estate-level lesson is the one already in memory: a `complete` work item
    is not a produced artefact, and here the artefact is sitting in the row that says complete.
  - **CORRECTION to what the owner was told 10 minutes earlier.** I said cancelling the three
    seotools duplicates was "cheap now, a page retirement later". The `add_tool` items are already
    `complete`, so cancellation is off the table — but **all ten pages are `planned` and NOT
    deployed**, so the honest position is: retiring them now is cheap *because nothing is public*,
    and it becomes expensive at deploy, not at build. Told him.
  - Also logged by `tool-deployer`, five times: *"suggestion carried no related_pages, so no
    cross-links were emitted"* — the suggester ran with no classification, so it produced no
    related-page hints and the tools land uncross-linked. Cosmetic today; worth knowing that a
    pre-classification sweep costs cross-links.
- **(ll) ⚠ THE ROLLBACK IS NOT YET PROVEN, AND NEITHER IS THE HALF I KEPT.** At 16:23Z: **0 classifier
  errors since the rollback — and 0 classifier RUNS.** A post-fix zero with no demand behind it is
  not evidence (memory: *a post-fix ZERO needs a DEMAND control*). I nearly reported "no errors" as
  reassurance; it means nothing until one runs.
  **Sharper, and I glossed it when reporting to the owner: `layout_taxonomy` in `input_fields` has
  ALSO never executed.** Every run since 11:39:14Z failed at my register step, which sat BEFORE
  `classify_and_extract` in the chain — so the taxonomy fix has never reached the prompt either. The
  next successful classification is the FIRST test of both the rollback and the surviving half, and
  if the taxonomy list breaks the render (it is a large array injected into a template that has never
  received it) that will surface then, not before.
  **Item `319d8d15` is now due (`retry_after` 16:22:05Z passed); `a428a8f5` becomes due 16:29:12Z.**
  Watch for a `classification` spec, and read the rendered prompt afterwards to confirm the library
  tag list is populated rather than `null` — that read is the actual proof, and it is owed to the
  445 lane as their real before/after boundary.

### (mm) 2026-09-03 16:3xZ — why №5's classification sat 14 minutes past its retry: SELECTION happens only when nothing is claimed

**Symptom:** both `needs_domain_research` items priority **5** (the lowest number on the site, i.e.
first), retry-due since 16:22:05Z, attempts 1/3, handler active — and still `triaged` at 16:33Z with
**0 classifier runs since the rollback**, while priority-**50** `needs_content_page` items were being
claimed and completed throughout.

**Two hypotheses of mine, both DISPROVEN before I found the answer** (recording them because the
disproving is the useful part):
1. *Priority starvation* — refuted: `load_items` is `ORDER BY wi.priority ASC`, so 5 outranks 50.
2. *Spend-governor shed of research* — refuted at the rows: `governor_state.shed_level = 0`,
   `governor_admits('needs_domain_research') = true`.

**The actual mechanism, read at the selector's own SQL** (`build-pipeline-trigger.find_dispatchable_site`):
its eligibility CTE ends
`AND NOT EXISTS (SELECT 1 FROM site_work_items active WHERE active.site_id = wi.site_id AND active.status='claimed')`.
**A site with ANY claimed item is invisible to SELECTION.** Sampled six times over two minutes:
copyonline held exactly **1** claimed item continuously.

So the sequence is: at selection time (~16:0xZ) the research items were **not yet retry-due**
(`retry_after` 16:22/16:29), so the loader's 8-item batch took content pages only; the loop then works
that batch serially, and **each claim keeps the site out of re-selection**, so the loop never re-loads
and never sees the items that became due mid-batch. **Not a defect and not starvation — a newly-due
high-priority item waits for the CURRENT BATCH to drain**, because selection and loading happen once
per batch, not once per item.

**Self-resolving**, and the prediction is falsifiable: when the batch empties and the claim clears,
the selector picks copyonline again and the priority-5 research items load FIRST. 11 content items
remain at ~7 min each, but the batch is 8, so the gap should arrive well before those drain.
**If a classification has still not run once the site next shows zero claimed items, this explanation
is wrong and something else is holding it.**

**Practice note for this lane:** after a failed item's `retry_after` passes, the retry is not due
"in 30 minutes" but "at the next SELECTION of that site", which on a busy site is a different and
longer thing. Do not promise the owner a retry time from `retry_after` alone — I did, twice.
- **(mm) WATCH ITEM for №5's plan — `bugs_open/463` × `bugs_open/444` interlock, relayed by the 463
  lane (misrouted to us; they thought this lane owned the 444 gate — it does not, that is the
  `bugs_open/444` session, dir `bugfix_444_empty_listing_pages/`; told them).**
  Mechanism, in their words: inside `ValidateSitePlanAction`, `reconcilePlanWithRealised` runs first
  and its **Pass C drops any planned page whose FIRST PATH SEGMENT equals a realised section index's
  stem** — which cannot tell a child (`/articles/x.html`) from a flat page colliding with the hub
  (`/articles.html`), so it deletes every NEW child of a section index. 444's gate then runs ~350
  lines later, `countSectionChildren` finds zero, and holds the hub with
  `builder_needed=section_children:<page>`. **Each guard is defensible alone; together they make an
  empty section index permanently unfillable, and each one's evidence reads as a reason for the
  other.** Worked case: gamedesign.uk 2026-09-03 — capability_gap 10:40:18Z, five children dropped
  14:15Z.
  **Why it is ours to watch:** copyonline's brief plans a guides section and a directory, so a
  section-index hub with children is likely, and **its plan does not exist yet**. If a
  `section_children` capability_gap appears on copyonline, it is probably this pair and NOT the 444
  gate misfiring — send the instance to the 463 lane rather than filing it.
  They are fixing the Pass C side, deliberately NOT taking 463's candidate 3 (flat article URLs
  outside the section prefix) because it would trade a dropped page for an orphaned one and break the
  `section_children` resolver.

### (nn) 2026-09-03 16:54:12Z — the taxonomy fix is PROVEN LIVE, and the first result is 10/10 matchable tags that describe the WRONG SITE

**Proof, at the rendered artefact** (`llm_call_log`, the run that classified copyonline 16:54:12Z):
the prompt now carries `Current library tags (…):` followed by a **populated JSON array**, and
**"The library currently has 19 active layouts"** — against `null` and `<no value>` in every prompt
before today. **This is the first classification in the platform's history to see the tag library**;
the step was built in April and has never delivered its payload. The 445 lane's reworded sentence is
live in the same prompt ("Prefer a tag from the list above whenever one describes this site… Coin a
new tag only when…"). **The rollback is proven too: 1 run, 0 errors.**

**The headline number, and why I am not celebrating it. Copyonline emitted 10 tags and ALL 10 are
matchable against an active layout — 100%, against the historical baseline of 28 of 216 distinct
terms (13%) that the 445 lane measured.** n=1, one site, so it is a datum and not a rate.

**⚠ THE TAGS ARE MATCHABLE AND THEY DESCRIBE THE WRONG SITE.** Emitted:
`marketplace, directory, community-platform, b2b, professional-services, content-platform,
creative-agency, interactive-platform, practitioner-platform, industry-hub`, with
`category=hub`, `site_type=interactive-platform`.
Copyonline's brief positions it as **an editorial authority on writing commercial copy, with tools**,
whose directory is ONE page and whose **marketplace is explicitly "a PREPARED later slot, never
claimed" and a must_not**. The classifier has typed it as a **marketplace / directory / community
hub** — the old 2015 site, not the one the brief describes. Not one tag says copywriting, editorial,
guides or long-form.

**So the honest reading is that the fix worked and may have made accuracy WORSE, in exactly the way
the instruction now invites:** "prefer a tag from the list above whenever one describes this site"
rewards reaching for a library tag, and a loose fit from the list scores better than an accurate coined
one. **Matchability is not accuracy, and coverage cannot tell them apart** — a site typed entirely in
borrowed vocabulary scores 100% coverage.
**This matters most to the 445 lane**, whose fit metric will look spectacular from this run onward and
whose pre-registered disconfirmation band was about coverage rising. Told them: the rise is real, its
cause is my fix, and it is not evidence of better fit.
**Alternative explanation I cannot exclude at n=1:** the brief itself leads with the old marketplace
heritage and the new directory, so the classifier may be reading the brief correctly and the BRIEF may
be over-weighting the directory. That is testable on the next remake and on this one's re-classify.
**No `resolved_composition` yet** — composition has not run, so the layout prediction is still open.

### (oo) 2026-09-03 17:0xZ — **EVERY BRIEF THIS LANE HAS WRITTEN HAS BEEN INVISIBLE TO THE CLASSIFIER.** Five built sites were classified from web research alone.

Found by the 445 lane while reading the same 16:54Z prompt I was reading, and verified here at the
artefact for three sites.

**The mechanism.** `classify_and_extract`'s template guards on the PARENT and prints a CHILD:
```
{{if .site_specs.specs.mission_brief}}## Pre-Defined Mission
This site has a strategic mission provided by the owner. Use this as STRONG guidance … the mission is the primary source.

{{.site_specs.specs.mission_brief.text}}
```
A brief-writer `mission_brief` is a **structured object** — `proposition, audience, stance,
content_plan, must_nots, reader_intent, differentiation, open_questions, tool_opportunities,
directory_opportunity, confidence, research_quality, regulated_subject` — and carries **no `text`
key**. So the guard opens, the preamble asserting an owner mission renders in full, **and the mission
renders `<no value>`.**

**MEASURED at the rendered prompt** (`llm_call_log`, most recent classification of each):
advertise.co.uk, seotools.co.uk and copyonline.co.uk all show the heading, the "primary source"
sentence, and then `<no value>`, followed straight by Search Results.

**MEASURED across the fleet:** of 23 current `mission_brief` specs, **7 lack `.text` — and they are
exactly this lane's seven brief-writer briefs**: advertise, buytoletcalculator, copyonline,
designblog, indoorplanters, seotools, websitepromotion. All 16 with `.text` are hand-authored or
older missions. **The split is not random; it is by producer.**

**So all four live remakes and copyonline had their `identity`, `classification`,
`content_direction` and `design_intent` written from web research alone**, while the model was told an
owner mission existed and was primary. copyonline's marketplace/directory tags are now explained:
research on `copyonline.co.uk` returns the 2015 marketplace and "THE COPY CO (UK) LTD"; the brief that
says otherwise never arrived. **My "the brief may be over-weighting the directory" hypothesis is
superseded by a stronger one that is measured: the brief was not there at all.**
- **This dwarfs RFC_037.** That RFC is about the classifier not seeing its SIBLINGS. It cannot see its
  own site's BRIEF — the artefact the owner personally reviews and approves before release.
- ⚠ **Migration 464 licenses a regulated business model off this same block** (445's finding), so on
  those 7 sites the licence text renders and the constraint does not.
- **Class is `bugs_open/453`'s** (the 445 lane contributed the general case there rather than filing a
  new number, and that lane is deep in this class today). **The lane-specific consequence — every
  brief-writer brief, five built sites — is ours and is recorded here.**
- **Fix shape, NOT yet applied and not to be rushed** (I broke the classifier today by rushing one):
  either the brief-writer emits a rendered `text` alongside the object, or the template renders the
  object. The data-side fix is live-immediately and testable; **but copyonline is mid-build and the
  owner has said not to change a running build**, so this is his call, not mine.
- **(oo) THE BLAST RADIUS IS THE WHOLE CHAIN, and the two consumers fail DIFFERENTLY.**
  **[MEASURED, config]** exactly two live agents' prompts reference the brief, and **both reference
  `mission_brief.text`**; **zero** reference `mission_brief` any other way:
  `domain-research-classifier.classify_and_extract` and `build-site-planner.plan_site`. Those are the
  two steps that decide what a site IS and what pages it HAS.
  **[MEASURED, artefact]** `llm_call_log`, `plan_site` prompts filtered by the site they are FOR
  (`LIKE 'Plan a website for <domain>%'` — an earlier filter of mine matched prompts that merely
  MENTIONED a domain and returned gamedesign's prompt three times; corrected):
  | site | plan_site prompt has a `## Pre-Defined Mission` heading |
  |---|---|
  | advertise.co.uk (2026-09-02 13:09) | **no** |
  | designblog.co.uk (16:10) | **no** |
  | seotools.co.uk (16:13) | **no** |
  | websitepromotion.co.uk (16:15) | **no** |
  | gamedesign.uk (2026-09-02 17:33, a `.text` brief) | yes |
  **So the two failures are not the same failure.** In the CLASSIFIER the guard OPENS and the child is
  empty — heading plus "the mission is the primary source", then `<no value>`. In the PLANNER the
  heading does not appear at all, so the guard never opened, which is a different cause (the brief is
  absent from that step's template data, not merely missing a child key). **Both end at the same
  place: the brief reaches neither.**
  ⚠ **Do not use "contains `<no value>`" as the tell** — every `plan_site` prompt sampled contains one
  somewhere, including gamedesign's working one. The heading's presence is the discriminating check.
  **Net: for all four live remakes, both the classification and the plan were made without the brief.**
- **(oo) ⚠ I WAS WRONG ABOUT THE PLANNER, and caught it myself 20 minutes later.** I wrote that the
  two consumers "fail DIFFERENTLY" and that the planner's guard "never opened". **False.** I searched
  planner prompts for `## Pre-Defined Mission` — a heading only the CLASSIFIER emits. **The planner's
  heading is `## Mission`.** Re-measured with the right needle: advertise, designblog, seotools and
  websitepromotion all render `## Mission` followed by **`<no value>`**, exactly like the classifier;
  gamedesign.uk (whose brief HAS `.text`) renders its real mission text as a working control; and
  boxingonline.com, with no `mission_brief` at all, correctly renders no block — a negative control.
  `plan_site.input_fields` DOES include `site_specs`, so the brief reaches the step.
  **It is ONE failure mode in both consumers, and therefore ONE data fix — a `text` key — repairs
  both.** Corrected in the `bugs_open/453` CONTRIB too, where the false sub-case would have sent the
  fixer looking for a second mechanism that does not exist.
  **The lesson is the day's own, turned on me twice now: I encoded the wrong question.** "Absence of
  a heading" is only evidence if you searched for the heading that artefact actually emits. The check
  that saved it was reading the planner's template rather than trusting my own earlier query.


---

### (pp) 2026-09-03 ~17:30Z — the brief-visibility defect is no longer a prediction: copyonline is building a different site, and BOTH consumers that have run say so in their own words

Entry (oo) reported that the owner-approved brief reaches neither of its two consumers, evidenced at
the rendered prompt (`<no value>` where the mission should be). That was an instrument reading. This
entry records the **outcome**, which is worse and needs no instrument at all — the agents wrote their
own confession into the artefacts they produced.

**1. Both agents state, unprompted, that they had no brief.**

`site_specs.classification.reasoning`, written 16:57:10Z by `domain-research-classifier`, verbatim:

> "Confidence is moderate because **no mission brief was supplied** and the existing site content is
> sparse — the strategic direction has been inferred from the site's own stated rules and the domain."

`site_specs.tools.reasoning`, written 16:01:52Z by `tool-suggester`, verbatim:

> "**Without existing pages loaded, I'm inferring from the domain** that visitors are likely businesses
> or individuals seeking copywriting services…"

This is the strongest evidence class available and I should have looked for it first. Entry (oo)
measured the *prompt* (an input) and reasoned forward to the consequence. The `reasoning` field is the
*output*, and the model narrates its own inputs there for free. **When a defect is "an agent cannot see
X", read what the agent said about its inputs before building a prompt-render harness.** Cheaper,
earlier, and it cannot be argued with. Filed as a check in WRONG_CALLS.

**2. The classification it produced is not the site the owner approved.**

| | |
|---|---|
| `category` | `hub` |
| `industry_tags` | marketplace, directory, community-platform, interactive-platform, professional-services, b2b, creative-agency, content-platform, practitioner-platform, tool-portal |

Inferred from the **old Drupal 7 rules page** — the classifier says so. The brief describes a
copywriting authority site with one converting page. It has been classified as a two-sided marketplace.

**3. Zero of the 30 planned pages are among the 10 built. [MEASURED 2026-09-03 17:25Z]**

Built (all 10, all `active`, all deployed today 16:15–17:20Z): five library tool pages —
Website Brief Starter, SERP Snippet Previewer, Insight Injector, Keyword Intent Classifier,
Title Tag Scorer — and one companion guide each.

The brief's four tools are Headline Scorer, Readability and Clarity Checker, Call-to-Action Tester,
Length and Character Counter (plus Tone-of-Voice Consistency Checker). **Intersection with what was
built: empty.** So is the intersection with its 26 planned guides, editorials and data pages. The AI
opening angle the owner asked to lead with (three guides), the lead-route page carrying the whole
conversion aim, and the UK Copywriter Directory are all unbuilt.

`needs_content_page` items name only the ten tool/guide pages. The site is being built as an SEO tools
portal, which is a coherent site — just not this one.

**4. The three duplicate tools shipped while the owner's decision on them was open.**

`serp-snippet-previewer`, `title-tag-scorer` and `keyword-intent-classifier` duplicate seotools. I
recorded them in the handoff as "pages exist as `planned`, undeployed" and put the retire-or-keep
question to the owner. They are now `active` and deployed (16:22:43Z, 16:39:58Z, 16:34:42Z). **A
question put to a human does not pause the pipeline**, and I wrote the status line as though it would.
Any "undeployed, so the decision is still cheap" phrasing needs a re-read before it is relied on.

**5. Not caused by this lane.** copyonline carries **no `add_tool` work items**. The five tool pages
came through the tool-deployer's own page-creating path (`bugs_open/450` §7), driven by the
`tool-suggester` spec above. The 7+1 `add_tool` items this lane fired were the 450 instance, a
different site.

**6. The window is still open, narrowly.** `plan` is still absent and `resolved_composition` has not
been written, so the 30 pages are not yet foreclosed — the planner has not run. Everything above was
decided by the classifier and the tool-suggester, both of which have already run blind.

**Not acted on.** The fix is written, guarded and verified in
`SQL_2026-09-03c_make_briefs_visible_to_the_classifier_and_planner_HOLD.sql` and is NOT applied. The
owner's standing instruction is not to change a running build, and he has not answered. What has
changed since I last asked is only the evidence, not the permission.

### (qq) 2026-09-03 ~18:10Z — correction to (pp): I called the ten pages "live" and nothing is published

Entry (pp) says the ten tool/guide pages are "built, deployed and live". **"Live" is wrong and was my
word, not the system's.** [MEASURED 2026-09-03 18:03Z]

```
https://copyonline.co.uk/                                       200  Drupal 7, <title>Copywriters
https://copyonline.co.uk/tools/title-tag-scorer/index.html      404   <- the recorded pages.url
   …/tools/title-tag-scorer/  …/tools/title-tag-scorer          404   <- other URL forms
   …/guides/tool-title-tag-scorer-guide.html                    404
   …/tools/website-brief-starter/index.html                     404
https://copyonline.co.uk/this-page-does-not-exist-<epoch>.html  404   <- invented-URL control
```

Control 404s, so the domain is not parked and these 404s mean what they say. `x-generator: Drupal 7`,
`server: LiteSpeed` at the root — the previous installation, still serving, which is also what the
blind classifier read to conclude the site is a marketplace.

`sites` row: `publish_target` NULL · `publish_project` NULL · `published_at` NULL ·
`last_deployed_at` NULL · `build_status` `pending`. **The site has never been published anywhere.**

**The mistake.** `pages.deployed_at` is populated on all ten and `pages.status='active'`. That is a
page-level flag about rendering; it carries no claim about the site having a publication target. I
read a page-level status as site-level publication, and then wrote a stronger word than either
supports into an owner-facing message.

**Why it slipped past the usual guard.** The estate's rule *trust the rendered artefact, not the
status* is one I applied this morning to the tool components — the whole repair was driven by fetching
pages rather than reading `page_components`. I dropped it in the afternoon because the status I was
reading agreed with the conclusion I had already reached from six other measurements. **A status that
confirms a well-evidenced conclusion is the one least likely to be checked**, and the error ran the
safe direction (over-stating damage), which invites action rather than scrutiny.

Corrected in: `PLAN_2026-09-03_copyonline_remediation_options_and_cost.md` §7, `README_where_we_are.md`,
a WRONG_CALLS row, and a new LANDMINES entry distinguishing this from the neighbouring
deploy-window entries (there `deployed_at` is fresh and the bytes are stale, so the remedy is to wait;
here there is nothing to wait for, and the two are indistinguishable from the `pages` table alone).

**What does not change:** the brief reaches nothing, six derived specs were written blind, the site is
being designed as a marketplace, `needs_strategy` is queued and the plan has not run. What changes is
the cost of option A, which is now internal rework with no public exposure.

### (rr) 2026-09-03 ~17:55Z — MAJOR CORRECTION to (pp): a THIRD consumer reads the brief and it works, the planner gets the brief second-hand, and my recommendation to hold the build was wrong

Entry (pp) concluded that the owner-approved brief "reaches neither of its two consumers" and I widened
that to the owner as *"your approved brief reaches none of the agents that decide what the site
becomes"*. **The widening is refuted by an artefact written eleven minutes after I wrote it.**

**1. `domain-strategist` wrote `site_specs.strategy` at 17:44:34Z and it is faithful to the brief.**

```
site_type    : "authority-portal"          <- NOT the marketplace the classifier decided
growth_path  : "Launch with the AI-first editorial argument, the four core tool pages,
                … the copywriter directory seeded from initial research, and the lead route…
                Add the tone-of-voice consistency checker as the aspirational fifth tool…
                Route inbound leads from webdesign.uk and other estate siblings…"
revenue      : primary_model "lead_generation" — "the site already has a designed lead route
                as its single converting page … initially to the site owner and later to paying
                lead recipients"
sponsored    : "The randomised listing approach at launch is the right trust-building default"
```

It names the brief's **four tools** (headline scorer, readability checker, CTA tester, length
counter) and its **fifth aspirational** one, the lead route as the single converting page, the
copywriter directory, the Copy Clinic, the glossary, the AI-first opening — **and two instructions the
owner gave me in chat and I wrote into the brief: randomised directory listings at launch, and routing
webdesign.uk leads here.** Re-running (pp)'s own marker test with `strategy` included: it hits lead
route, `headline scorer`, `randomised listing` and `webdesign.uk`. The six earlier specs hit none.

**2. The mechanism, and it is at the TEMPLATE, not the data.** [MEASURED 2026-09-03 17:53Z, from the
live `agent_definitions` rows]

| agent | how its prompt names the brief | result |
|---|---|---|
| `domain-research-classifier` | `{{if .site_specs.specs.mission_brief}}` … `{{.site_specs.specs.mission_brief.text}}` | `<no value>` |
| `build-site-planner` | identical pair | `<no value>` |
| **`domain-strategist`** | **`{{.site_specs}}` — no mission-brief reference at all** | **works** |

The strategist renders the entire specs blob unfiltered, so the structured brief arrives as text and
the model simply reads it. **The data was never missing from any of the three.** `site_specs` is in
the planner's `input_fields` and reaches the classifier too. The defect is two template expressions
reaching for a `.text` child that brief-writer output does not have — and **a working reference
implementation has been sitting in the same estate the whole time.**

**3. The planner is blind to the brief but NOT to its content.** Its full render list:

```
{{toJSON .site_specs.specs.strategy}}      <- the WHOLE strategy object, and the strategy is faithful
{{.site_specs.specs.classification}}       <- the whole (wrong) classification
{{.site_specs.specs.identity}}  {{.site_specs.specs.briefing}}
{{.site_specs.specs.mission_brief.text}}   <- <no value>, the defect
{{.site_specs.specs.roadmap_brief.text}}   <- same shape, same trap, different aspect
```

So the brief's content reaches the planner **second-hand and in detail**, through a strategy spec that
is rendered whole. The planner will see the four tools, the lead route and the directory. It will also
see the wrong classification, so the risk is real — but it is "one good input against one bad one",
not "no input".

**4. What this does to my recommendation, which was wrong.** I recommended holding the build. On this
evidence **the right action is to leave it alone and judge the plan on its output.** The pipeline has
partially self-corrected: the strategist caught what the classifier missed, because it reads
generously. Intervening now would have disturbed a run that is recovering on its own, on the strength
of a claim I had over-generalised from two agents to all of them.

**5. What I got wrong, and how.** (pp)'s measurement was sound and its conclusion was scoped correctly
— *neither of its two consumers* — and then I widened it in the owner-facing message to *"none of the
agents that decide what the site becomes"*, which is a different and much larger claim I had not
tested. I had enumerated the consumers of `mission_brief.text` and then spoken about consumers of
**the brief**, which is a bigger set reached by more than one route. **Enumerating the readers of a
FIELD is not enumerating the readers of the INFORMATION.** The check that would have caught it is the
one I eventually ran by accident: grep every active agent's prompt for `{{.site_specs}}` as well as
for the specific path.

**6. What remains true and unchanged.** The classifier ran blind and said so; its classification is
wrong. The tool-suggester ran blind and five tool pages were built that are not the brief's, three of
them seotools duplicates on an open owner question. 7 of 23 current briefs lack `.text`. Nothing is
published. The template defect is real and worth fixing properly at `bugs_open/453` — and is now a
two-expression change with a proven pattern beside it.
