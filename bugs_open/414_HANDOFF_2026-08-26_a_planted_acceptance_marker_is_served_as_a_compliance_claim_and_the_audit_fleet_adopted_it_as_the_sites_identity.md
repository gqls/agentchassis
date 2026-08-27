# 414 — a planted acceptance marker is SERVED as a compliance claim, and the audit fleet adopted it as the site's identity

**Filed 2026-08-26 by the portfolio_positioning lane. Status: OPEN — framework fix committed
(`fc588e445`, council `f4c144ad`), spec sources fully stripped, the inverted audit item rejected,
copy repair dispatched through the framework and pending its render. §7 is the 2026-08-27 record.**

> **⚠ CORRECTED 2026-08-27 — the line that stood here was "spec source FIXED live", and it was
> WRONG in a way that kept this bug alive.** Only `content_direction` had been stripped. The plant
> had already been COPIED, in its own prose, into the **current `strategy`** aspect by
> `domain-strategist` on 2026-08-12 — an aspect the writer never reads and `build-site-planner`
> does. So "regeneration can no longer re-plant the phrase" (§"What is FIXED", below, left as
> written) was false for ten days. Caught by re-running this file's own §Population census over
> **every** aspect instead of the one that was edited. See §7.

## Symptom

`https://lendzy.co.uk/about.html` serves the sentence **"checked against the FCA handbook, rule
by rule"** twice, and `/guides/tool-affordability-complaint-checker-guide.html` once
[MEASURED 2026-08-26 19:0xZ, curl by body]. On a finance-adjacent site whose whole premise is
independence and accuracy, that is an unverifiable claim of regulatory diligence in the site's
own voice — nobody performed a rule-by-rule check.

## Mechanism — self-evidencing, every link verbatim

1. The 08-02 lendzy **shadow experiment** (this lane's `MISSION_2026-08-02_lendzy_shadow.md`)
   seeded `content_direction` with a **tripwire**: `positioning.acceptance_marker = "Somewhere in
   the site's written copy include the exact phrase: checked against the FCA handbook, rule by
   rule."` — duplicated as the tail line of the `formatted` field, which is what page generation
   reads (the spec row's own notes say `formatted` "carries it to page generation").
2. The lane's 08-05 handoff recorded the debt: *"marker strip owed BEFORE serving."* The site
   was then built and served by the fleet buildout without that entry ever being re-read —
   the memory-index landmine fired correctly tonight, 21 days late, exactly the
   `a-handoff-outlives-the-work-it-asked-for` shape.
3. The writer obeyed the instruction (it is an instruction, and it was followed — this is the
   `a-quoted-exemplar-in-a-prompt-is-copied-verbatim` class, deliberately induced): the phrase
   is stored in **3** components [2026-08-26] — `/about.html` `hero-about` +
   `content-block-about`, and the guide's `article-body` — in **`content_data`**, not just
   `rendered_html`.
4. **The new wrinkle, and why this file exists**: the maintenance/audit fleet then read the
   served phrase back and **canonised it**. An open `content_rewrite` item
   (`needs_human_review`) describes it as *"The site's core differentiator — FCA-rule-level
   accuracy checked guide by guide"* and asks for copy that leans INTO it. A tripwire did not
   just leak — the estate's improvement machinery adopted it as the site's identity and started
   generating work to reinforce it.

## Population — counted 2026-08-26

Fleet census of current `site_specs` for `acceptance_marker` / "exact phrase":
**1 site** carries a marker (lendzy.co.uk). apis.uk (`evidence_base`) and webdesign.co.uk
(`strategy`) match "exact phrase" innocently (a ban-pattern comment; descriptive prose).
lendzy's brief/strategy/submission also mandate a second exact phrase — *"know the rules before
you borrow"* — which is a benign brand slogan, functions as intended, and is NOT this defect.

## What is FIXED (live, DB config)

`content_direction` revised 2026-08-26 ~19:20Z: current row `81ddcc40-b1e2-426a-b4d2-ef68e949d1c8`
(`created_by='portfolio-positioning-2026-08-26'`) = the 08-02 row minus the
`positioning.acceptance_marker` key and the `formatted` tail line, applied server-side inside a
guard that asserted the exact tail before trimming. History preserved (`61ef7033…` superseded,
residue intact for audit). ~~**Regeneration can no longer re-plant the phrase.**~~

> **REFUTED 2026-08-27.** The final sentence was true of the WRITER's path and false of the
> estate's. `strategy` row `96eaff0b` (`is_current`, `domain-strategist`, 2026-08-12) still read:
> *"The acceptance marker 'checked against the FCA handbook, rule by rule' should appear in the
> site's written copy to anchor the editorial credibility claim."* Stripped 2026-08-27 08:30Z under
> the same tail-assert guard (new row `0326a892-0a82-4929-889f-84de6df83d55`, `96eaff0b`
> superseded, history intact). **Fleet re-census after the strip: 0 current specs, any aspect, any
> site, carry the phrase or an acceptance-marker instruction** — escaping the `_`, which is a SQL
> wildcard and which over-matched on the first attempt (WRONG_CALLS 2026-08-27).

## What REMAINS — and the trap in the obvious fix

- The 3 components still carry the phrase in `content_data`, and 2 pages serve it.
- ⚠ **The queued `page_rerender: Rerender page: about` (triaged) will NOT fix this** — a
  rerender regenerates from `content_data`, where the phrase lives. The repair is a **content
  rewrite** of the 3 components (framework, not hand-editing — owner ruling 08-06).
- The held `content_rewrite` item quoted above now has an **inverted premise** (the
  "differentiator" it wants amplified is a planted tripwire): whoever triages it should reject
  or rewrite it against this file, not release it as-is.

## Verify (after the copy repair)

```sql
SELECT count(*) FROM page_components pc JOIN pages p ON p.id=pc.page_id JOIN sites s ON s.id=p.site_id
WHERE s.domain='lendzy.co.uk' AND (pc.content_data::text LIKE '%checked against the FCA handbook%'
   OR pc.rendered_html LIKE '%checked against the FCA handbook%');  -- expect 0
```
plus curl both pages by body (expect 0 matches; `rm` the temp file first).

## Why no 090 run (owner ruling 2026-07-31 escape hatch, stated plainly)

The causal chain is verbatim string identity at every hop — instruction in the spec, phrase in
the stored components, phrase in the served bodies — each independently queried/fetched this
session, and the fleet census bounds the population at one site. There is no inference in the
chain for a diagnosis loop to refute; the one judgement call (that the audit item canonised the
marker) quotes the item's own text.

---

## 7. What happened on 2026-08-27 — the framework fix, and the two hops nobody had joined up

Resumed by a fresh session; the filing lane had handed the copy repair on and was idle. Everything
below was re-measured this session rather than inherited, because the one inherited claim in this
file was the one that was wrong.

### 7a. The bug was still live, and bigger than one aspect

Re-measured 2026-08-27 08:1x–08:4xZ: served `/about.html` ×2 and the affordability guide ×1 (curl
by body; control phrase "PRA handbook" = 0). Stored in **both** `content_data` and `rendered_html`
on 3 components — `hero-about`→`subheadline` (301 B), `content-block-about`→`body_text` (422 B),
`article-body`→`content` (5,697 B). None locked, none tombstoned, site not locked.

**The chain has TWO hops, and the second is what this file missed.** The plant arrived as a
**manual** row (`source='manual'`, `created_by='cqls'`, 2026-08-02 18:41) — so no guard on
`WriteSiteSpecAction` would ever have caught it. Ten days later an **agent** read that row and
restated the instruction in its own words in a **different aspect** (`domain-strategist` →
`strategy.content_strategy`, 2026-08-12 18:49). Stripping the origin did not retract the
instruction, and nothing in the estate joins a retraction to its copies.

**And the phrase had been more widespread than the census showed.** `page_component_history` holds
**14** archived rows carrying it, across `about` (3 slots on 08-11, including a `differentiators`
slot that no longer exists) and the guide's `article-body` (**4 versions, 08-15 → 08-24**). The
guide re-emitted the phrase on every single regeneration while the spec still mandated it — which
is the direct evidence that the spec drove it, and the reason a framework rewrite against a CLEAN
spec is the right repair rather than a hand edit.

### 7b. The audit item was worse than described, and is now rejected

Item `052d01b0` (`design-audit`, 2026-08-11, `needs_human_review`, handler `page-build-handler`)
carried a `current_value` that is a **fourth instance** of the claim — *"Our guides are checked
against the FCA handbook, rule by rule. We name the exact rules so you can read them yourself"* —
attributed to `page_name: index`, and a `suggestion` asking for a **"How we verify our guides"
methodology section** with named CONC citations. Its `acceptance_test` was satisfiable innocuously
(name any CONC rule), which is what made it dangerous: a handler could pass the test while building
the methodology section. The homepage instance is already gone (index regenerated 08-24), so the
item was quoting copy that no longer exists.

**REJECTED 2026-08-27 08:33Z** (`status='rejected'`, reason on the row and in `result`), under a
guard that aborted unless the row was still `needs_human_review`. Not confirmed, not retried: at
`needs_human_review` it is undispatchable, but one Retry click sets it `triaged` and regenerates the
page from `spec.suggestion`.

### 7c. The framework fix — three narrow changes, all committed in `fc588e445`

Council `f4c144ad` (submitted; trailer `Council-Submitted:`, so `098` credits it on approval).

1. **`claims_global.go` — the completeness shapes the refusing set missed by a hair.** The family
   already owned `every (claim|figure|…)s?[^.]{0,30}(is|are) (verified|checked|…)`. The guide's
   sentence puts **38** characters between "every figure" and "is checked", and "Everything" is not
   "every"+a listed noun. Window 30→60 plus an indefinite-subject sibling.
   `[MEASURED 2026-08-27, 2,405 live components]`: at window 30 that pattern fires **0** times
   fleet-wide — it was inert; at 60, **1**, exactly the missed sentence. The new entry: **1**,
   exactly the other. The 2026-07-28 NOUN narrowing is untouched, pinned by a fixture.
2. **`claims_practice.go` P6 — the diligence conjunction, at WARNING, attestation-exemptible.**
   Owner decision 2026-08-27: split by the refusing set's own bar. A compliance-services client
   could truthfully say it, and — decisively — `negationCueRe` has no bare `nothing` cue, so
   *"Nothing here has been checked against the FCA handbook, rule by rule"* reads as un-negated and
   at blocker this layer would refuse the correcting disclosure it exists to encourage.
   The conjunction is the precision, not the words `[MEASURED 2026-08-27]`: exhaustive idiom alone
   **22** components, verb+`against`+rulebook alone **13** (including lendzy's own *correct*
   imperative "Check your loan against the FCA rules"), both together **3** — the planted ones, 0 of
   the other 2,402.
3. **`cmd/brief-negation-check` gains a second detector: the claim rules applied to the
   INSTRUCTION.** Surface unioned over **every live agent prompt**, not the writer's alone — a
   writer-only surface is exactly what would have missed hop two. Scanned with the **practice family
   only**, which is measured, not preference: the fleet-wide+regulated set over 522 current spec
   rows gives **21** hits, effectively all false — **15** are the estate's own honesty instructions
   ("Never invent a person, company, scheme…") matching the never-invents pattern, which the
   negation guard cannot save because the match *starts* at "never"; and `evidence_base` rows store
   each site's `banned_claims` **as data**, quoting the sentences they forbid, so a generic scan
   convicts every site's own immune system daily. The practice family over the same text: **0 of
   532** current rows, **2 of 2,782** all-history — *exactly the two hops of this chain*. So the
   detector would have caught the plant the day it was made and the propagation ten days later.
   Files `spec_supplies_claim` at `needs_human_review` with **no handler**, because an automated
   spec-rewriter is precisely how the marker got canonised.

### 7d. Ordering — why the repair went FIRST, against the obvious instinct

⚠ **With the pattern live and the phrase still in `content_data`, a plain `page_rerender` of
`/about.html` regenerates HTML carrying the phrase, the persistence floor
(`save_sections_claims_guard.go`) refuses the save, and the OLD `rendered_html` — with the claim —
keeps serving.** The item lands `unresolved`; nothing a visitor sees changes. The gate would have
turned a useless rerender into a stranded one. So: data and dispatch first (immediate, no build),
Go committed alongside (inert until a roll). The gate then stands as the acceptance test for every
*future* rebuild, and this repair is proved at the artefact instead.

### 7e. The repair, dispatched through the framework

Two `content_rewrite` items at `spec.mode='edit_live'` (`load_current_section_content` then hands
the writer the live prose — without that flag a page-build-handler rewrite is a from-scratch
regeneration, measured at `bugs_open/178` as 4,439 → 1,806 chars on one page), priority 10, handler
`page-build-handler`, `item_key` `content_rewrite:bug414:<page>`:
`5fd36ec2` (about) and `4e364317` (the guide). The `rewrite_guidance` says what to remove, why it is
false, **and what is true from the site's own recorded brief** (name the rule beside the figure and
link it) — so the replacement is grounded in the brief rather than authored by a session — and
forbids replacing it with any other verification/completeness claim. Both claimed by the dispatcher
within ~15 minutes. A before-snapshot of all 6 components on the two pages is held so `edit_live`'s
preservation of the untouched slots can be checked rather than assumed.

### 7f. Verify (unchanged in substance, sharper in method)

```sql
SELECT count(*) FROM page_components pc JOIN pages p ON p.id=pc.page_id JOIN sites s ON s.id=p.site_id
WHERE s.domain='lendzy.co.uk' AND (pc.content_data::text LIKE '%checked against the FCA handbook%'
   OR pc.rendered_html LIKE '%checked against the FCA handbook%');  -- expect 0
```
Then `scripts/probe-page-url.sh lendzy.co.uk about tool-affordability-complaint-checker-guide`
(it reads the recorded `pages.url` and enforces an invented-URL control, so a catch-all cannot read
as healthy), plus curl-and-grep by body with the same control and a byte-delta check.
**`complete` is not proof** — a lock- or decision-gated refusal completes too.

**The retraction sweep, which is the transferable half.** For any retracted phrase, one query over
the three surfaces at retraction time — and over **every** aspect, not the one you edited:

```sql
SELECT 'spec' AS surface, s.domain, ss.aspect FROM site_specs ss JOIN sites s ON s.id=ss.site_id
 WHERE ss.is_current AND ss.data::text LIKE '%<phrase>%'
UNION ALL SELECT 'component', s.domain, COALESCE(pc.slot_name,'') FROM page_components pc
 JOIN pages p ON p.id=pc.page_id JOIN sites s ON s.id=p.site_id
 WHERE pc.content_data::text LIKE '%<phrase>%' OR pc.rendered_html LIKE '%<phrase>%'
UNION ALL SELECT 'work_item', s.domain, w.item_type FROM site_work_items w JOIN sites s ON s.id=w.site_id
 WHERE w.status NOT IN ('complete','cancelled','rejected')
   AND (w.summary LIKE '%<phrase>%' OR w.spec::text LIKE '%<phrase>%');
```

### 7g. Residuals, stated plainly

- **The completeness shape in a SPEC is not covered by the spec detector** (it scans with the
  practice family). The page gate refuses such output at blocker, so it cannot be served — but the
  instruction would survive in the spec and refuse every rebuild until a human reads the refusal.
- **P6 has no post-deploy path.** The practice family is consumed only by the build gate and
  `cmd/claimscan`, not by `check_unverified_claims`, so it fires on rebuild rather than over the
  installed base. Giving it a reporting (never refusing) consumer there is a live council question.
- **Two P6 false-positive shapes are pinned by a test that asserts they FIRE**, so they are on the
  record rather than hidden: a third-party subject ("The FCA checks firms against the handbook, rule
  by rule") and the bare-"nothing" disclosure. The general fix for the second is cues on a guard
  **shared by every claims family** — a guarantee change, i.e. architecture scope.
- **An `operating_history` attestation exempts P6 too**, which is about different work from physical
  practice. `[MEASURED 2026-08-27: 0 sites carry either attestation.]`
  `TestPracticeDiligenceHonoursTheAttestation` pins the coupling.
- **Paraphrase drift is real and a literal family cannot win that race.** The audit fleet had
  already restated the claim as "FCA-rule-level accuracy checked **guide by guide**" — no `against`,
  no rulebook noun. The durable control for the canonisation loop is the one that already shipped:
  the `model_opinion` origin stamp and record-mode filing (migration 629). Every lendzy audit item
  filed 08-25/08-26 is `deferred`, not dispatched. **The residual is the pre-door backlog** — items
  like `052d01b0` that predate it and still have a one-click Retry.

### 7h. The absence claim, checked rather than asserted (council round 1 gated on exactly this)

§7c says "nothing in the estate reads spec CONTENT for planted instructions outside
`cmd/brief-negation-check`". That is a load-bearing negative over a large surface, and the council's
`prior_art_librarian` seat gated round 1 on it having no cited existence check — correctly, given
this estate's history of dormant and duplicate detectors. Run 2026-08-27:

**87** non-test Go files reference `site_specs`; **19** of those also compile a regex or call a
claims scanner; reading each, exactly **one** judges the spec's own text — this binary, on
define-by-negation only. The near misses, named so the next reader does not have to re-derive them:
`check_literal_markdown` mentions `site_specs` only in a comment; `check_revenue_shape` reads
`strategy.revenue_models.primary_model` for **shape**; `check_build_prerequisites` and
`check_premise_incomplete` ask only whether aspects **exist**; `refresh_evidence_base_action`
maintains the register; `cmd/regcheck` scans one string handed to it on the command line.

**And there is one mechanism that already reads this exact text — to EXEMPT it.**
`rewrite_negations_action.go`'s `defaultBriefFields` includes
`site_specs.specs.content_direction.formatted`, and reads it in order to *exclude* brief-supplied
phrases from the writer-seam gate, on the stated ground that *"a site's own voice specification
outranks these rules wherever the two disagree"*. So the estate's one existing reader of the
planted text is the one that decided not to judge it. **That is not a duplicate detector; it is the
gap 414 fell through, stated in the code that creates it.**

Also run, because the same seat asked and round 1 had not cited it: this council has reviewed the
claims families four times in the preceding 20 days (`c48b7612` 08-20, `aac38d5b` 08-20 and 08-22,
`1d87615f` 08-24, `6cfaa8f0` 08-25). **No prior round covers spec-side scanning**, and none contains
an equivalent to either pattern added here.

### 7i. Round 1's other findings, kept because two of them were real

Round 1 came back **REVISE**. Recorded here rather than only in the council trail, because a REVISE
round that finds real defects is evidence about the change, not an obstacle to it:

- **Two seats independently asked whether the fleet spec-surface census descends into loop
  substeps** — the documented failure mode of hand-rolled walks over `agent_definitions`. It cannot:
  `fleetSurface` never walks steps, it regexes `default_config::text` as one document. But the
  measurement makes the question sharper than the objection: `[MEASURED 2026-08-27]` exactly **one**
  live agent carries a `site_specs` ref inside a `sub_workflow` — **`page-content-writer`**, whose
  refs all live inside its process-sections loop. A step-walking implementation would have gone
  blind to the writer's entire surface while reporting a clean fleet. Now pinned by
  `TestFleetSurfaceSeesRefsNestedInsideASubWorkflow` (`04dddb699`), whose failure message names
  `platform/validation.WalkSteps` for whoever converts it.
- **The writer set of `content_data`/`rendered_html` is two families, not six consumers**
  (`cta_label_audit.go:36-49`): `SavePageSectionsAction` (behind the floor) and
  `ApplySectionEditAction`, which "never passes through SavePageSectionsAction". **Neither new
  pattern guards a section edit** — `section_editor_regulated_guard.go` carries the regulated family
  only, by its own header's reasoning. Stated, not fixed: widening it changes refusal behaviour on
  another lane's seam.
- **A refused save already leaves a durable record** — `writeClaimsFloorLog` →
  `agent_error_log` code `CONTENT_CLAIMS_FLOOR_DETAIL`, on both the refusal and the
  record-and-allow path — and deliberately writes a RECORD not an ITEM (`bugs_open/083`: detections
  filed as items, fixed zero times; `bugs_open/077`: never file an item whose handler has no remit).
  The item-raising path exists separately as `claims_unverified`. What the objection is right about
  is that a refused rerender lands its work item at `unresolved` with nothing filed — a queue-drain
  question for `bugs_open/033`, not something this change should invent.
- **Owed after the roll**: per-SERVICE artefact verification (`git merge-base --is-ancestor` against
  the service's own build-provenance stamp), then re-run `claimscan` over the same corpus expecting
  the same three findings. A pre-merge dry run proves the source, not the running binary.

### 7j. Council: APPROVED at round 2 — and the five advisories, including one that fixed my own verification plan

`Council-Reviewed: f4c144ad-7799-4cd0-b792-d97016f3d77e` (round 2, 09:52Z, approved with 5 advisory
objections, none high-severity; round 1 REVISE at 09:24Z, gating objection answered in §7h). The
earlier commits carry `Council-Submitted:` and `098` credits them automatically on this approval — no
amend, which forward-only forbids anyway.

**⚠ The advisory that changes an owed action: my post-roll verification recipe was wrong for this
service.** §7i said "per-service artefact verification via `git merge-base --is-ancestor` against the
service's own build-provenance stamp". The `debug_historian` seat flagged that as the documented
unreliable method for `agent-chassis`, and reading the landmine it names, the truth is subtler than
either of us said: the stamp recipe is **time-limited, not inoperative** — it is a STARTUP line the
chassis rotates away in minutes (measured: still readable at 7 minutes on `v1.0.1295`). So the
corrected recipe, in order:

1. **Prove the window first**: `kubectl -n ai-persona-system logs <pod> | head -1` — a startup line
   means the log still reaches back and the stamp is in range. Anything else means it is not, and an
   empty stamp grep then means "out of range", never "unstamped".
2. `kubectl logs <pod> --tail=100000 | grep -m1 'build provenance'`, then
   `git merge-base --is-ancestor fc588e445 <stamp>` — **and a DESCENDANT commit as a must-be-absent
   control**, because the stamp is one commit and never yours, so an ancestry test with no negative
   arm can only come out true.
3. **If the window has closed**, do NOT reach for `strings` or a discovery grep for "some 40-hex
   string" (both documented to return the same wrong answer on every service). Probe a **known
   symbol** with both arms: `grep -aq 'everything (on this site' /proc/1/exe` must be PRESENT, and a
   string this change did not add must be ABSENT. The new pattern literals are pure ASCII, so the
   non-ASCII byte trap does not apply to them.

**The other four, recorded rather than actioned, each with why:**

- **`guardian`: "if the binary is not scheduled, the detector exists but never fires, silently."**
  The right question, and the honest answer is that it is scheduled (`brief-negation-check`, daily
  `40 7 * * *`, existing CronJob) **but its image is TAG-PINNED**, so the detector is inert until the
  owning lane's next image cycle — agreed with them explicitly, since there is no live positive case
  left for it to catch today. **The first live run must report the whole fleet as `M` with a non-zero
  `scanned_fields` count**; a zero from a blind scan and a zero from a clean fleet are otherwise
  identical in the report.
- **`reuse_agent`: `cmd/config-key-audit` already houses agent-config reference censuses**
  (`relaygaps.go`, `sharedoutputs.go`, `livedeclarations.go`) and I never cited it. Fair: I searched
  for prior art on the *claims* side thoroughly and on the *census* side not at all. It would not
  have changed the placement — that binary writes `doc_notes` only and a planted instruction needs a
  queue row — but the surface-derivation FUNCTION is duplicated logic, and that is a real unification
  debt, recorded here rather than argued away.
- **`compliance`: P6 at warning is too soft for this class on a finance site where the fabrication
  already shipped 24 days.** A direct challenge to the owner's 2026-08-27 decision, and it belongs in
  the record as a disagreement rather than being smoothed over: the seat whose remit is overclaimed
  reliability would have refused it. The counter-argument stands (a compliance-services client could
  say it truthfully; and the bare-"nothing" disclosure would be refused at blocker) — but the owner
  should know a seat dissented.
- **`compliance` (low), and this one is a genuinely new residual**: `evidence_base` is excluded from
  the spec scan, so **a poisoned register — a fabricated `source` or fact — sails through both the
  writer-side gate and the new spec-side detector**, because the register is treated as ground truth
  by every layer rather than scanned itself. Nothing in this change addresses that, and it is not a
  variant of 414; it is a sibling worth its own filing if anyone finds an instance.
- **`bug_historian`: filing into `bugs_open/033`'s queue (no working surface, ~1,079 items) risks
  making the fix invisible.** True, and deliberately accepted: the alternative is an automated
  handler for spec-content findings, which is precisely the mechanism that canonised the marker.
- **`editquality` (bookkeeping): `specclaims_test.go` was described but not listed as its own edit**,
  so its guards were unverifiable from the plan. It exists and is committed (`fc588e445`), and the
  sub_workflow test is in it (`04dddb699`).

### 7k. THE ABOUT PAGE IS REPAIRED AND VERIFIED AT THE ARTEFACT (2026-08-27 10:19Z); the guide is blocked on a fleet LLM outage

**`/about.html` is clean.** All three of its components were rewritten by the framework writer at
10:19:03Z and the phrase is gone from **both** `content_data` and `rendered_html`. Verified at the
served body, which is the only thing that counts:

- `curl` of the recorded `pages.url`: **0** occurrences (was 2), HTTP 200;
- **invented-URL control on the same domain returns 404**, so the 200 is a real page and not a
  catch-all — without that control a parked domain would have read as healthy;
- and the strongest evidence, because it could have come out otherwise: **`cmd/claimscan` over the
  whole repaired site now reports the about page CLEAN and still convicts the guide's `article-body`
  on both shapes, in one run over one corpus.** The detector discriminating between a repaired and an
  unrepaired component in the same pass is a better demand control than any before/after count.

**What the framework wrote**, since the gate that would have checked it is inert until the next roll
and therefore nothing but a human checked this output:

> hero: "…We explain what the FCA rulebook says a lender can and cannot do, **and every figure we
> quote comes with the named rule it's from and a pointer to check it for yourself**, so you can hold
> your lender to it."
> content-block: "…**Every regulatory figure on this site is quoted together with the named rule it
> comes from and a pointer to where you can check that rule for yourself.**"

That is a description of what the site DOES, verifiable by any reader, and it is what the recorded
brief actually requires — so the replacement is grounded in the brief rather than authored by a
session. It also does **not** trip either new pattern (the verb is "quoted", not
verified/checked/confirmed/validated, and there is no `against`+rulebook+idiom conjunction), which is
why the claimscan run above is silent on it. Had it tripped, I would have made that page unbuildable.

⚠ **`edit_live` rewrote all three slots, including `differentiators`, which never carried the
phrase** (2,245 → 2,293 chars, so it grew — no content loss). The plan predicted this: `plan_sections`
has no slot filter, so a `content_rewrite` regenerates every ready section on the page. The
before-snapshot is what made the check possible rather than a hope.

**The guide is NOT repaired, and the blocker is no longer the queue.** Its two attempts died to
infrastructure — first `"Claim timed out — handler pod likely died"`, then nothing — and since
**11:30Z there is a fleet-wide, account-level LLM outage**: measured first-hand in `llm_call_log`,
**11:00Z 36 of 132 calls failed, 12:00Z 61 of 61, 13:00Z 23 of 23, every failure a "usage limit"
error.** No LLM-bearing rewrite can complete until that clears. The about page got through at 10:19,
before it started. So: the item is queued and correct; **do not read further failures as a payload
problem, and do not burn its last attempt while the outage stands.**

**Correction to §7e's account of the claim I released.** I said no reaper covered a claim whose spawn
was dropped. **False** — `claimed-item-timeout` (enabled, 120 s) covers it, and the thresholds are the
part worth knowing because they are NOT uniform: the two auto-complete stages key on **15 minutes**
and both need completion evidence, so a dropped spawn can never match them; the `reset` stage keys on
**40 minutes** and needs no orchestration. My item was untouched at 34 minutes, which is therefore
*correct behaviour*, and my hand-release at ~09:26 preempted the mechanism by about five minutes
rather than rescuing anything. Established by the `bugs_open/413` lane reading the `reset` stage's own
WHERE after I had asserted otherwise twice, and re-verified here by extracting the clause directly.
The whole-site-darkness finding survives the correction but is **bounded at ~40 minutes per
incident**, not unbounded — and that lane measured **≥89 claim-timeout resets today across 27 sites**,
so it is a real amplifier at a real frequency. How I reached the false version is in `WRONG_CALLS.md`:
the right query, piped through `head -14`, three of twelve matching rows visible.
