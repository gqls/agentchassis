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
