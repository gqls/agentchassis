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

**Fresh chassis v1.0.1305 rolled** (both replicas, 2026-08-16 22:07/22:08Z). Phase B is
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
