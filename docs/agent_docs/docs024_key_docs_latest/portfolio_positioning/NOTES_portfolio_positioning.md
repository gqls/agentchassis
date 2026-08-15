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
