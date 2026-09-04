# NOTES — theme kits (append-only, NEWEST AT THE BOTTOM)

The technical log: what was tried, what the system actually said, and every misstep.
**The missteps are not an appendix — they are the point.**

---

## 2026-09-02 — Phase 1, reconstructed from the handoff and the plan file

Written up on 2026-09-03; this lane had no NOTES file while the work was happening,
which is itself the first entry.

**Built and committed:** `theme_kits` registry, `page_archetypes`, `apply_theme_kit`,
the layout resolver rung, migrations 689 and 691. Commits `0902039c0` through
`b6039c26b` (full table in `HANDOFF_2026-09-02_continue_here.md` §4).

**Eight defects found by two Fable reviews and fixed BEFORE the migration was applied.**
The schema was still unapplied, so these cost nothing:

1. A `UNIQUE` that could never fire — every `page_archetypes` row has a NULL in the key
   and Postgres NULLs are distinct by default, so the seed's `ON CONFLICT` was dead code.
   Fixed to `UNIQUE NULLS NOT DISTINCT`. **Proven by induced fault**: re-running the seed
   returns `INSERT 0 0` where the old version would have returned `INSERT 0 1`.
2. A false structured fact — a kit-chosen layout was recorded as
   `layout_source: 'library_match'`, which asserts a weighted tag match that never
   happened. Validator enum widened in 689, **spliced from the LIVE function body** so
   its UUID checks, name-field loop and `reasoning`/`resolved_by`/`resolved_at`
   requirements are preserved verbatim. A first attempt rewrote the validator from a
   partial read and would have silently dropped all of those.
3. `fill_gaps` queued a `needs_composition` that was guaranteed to be refused, while
   reporting success.
4. A presence check that disagreed with the reader's own shape acceptance. The resolvers
   use `extractReferenceValuesFromSpec`, which accepts BOTH nested and FLAT palette
   shapes; an independent nested-only check here read a flat-shaped site as "has none",
   wrote the kit's values nested, and the resolver then preferred the kit's — silently
   overwriting the site's own palette, which is the exact guarantee `fill_gaps` exists to
   keep. Latent when found ([MEASURED 2026-09-02] 0 flat palette rows, 1 flat typography
   row) — fixed before it wasn't.
5. The typography cascade asymmetry (see PLAN). `design_intent` is rung 1 for typography
   and rung 2 for palette, so writing kit fonts into `design_intent` would have silently
   outranked a human's `mission.preferred_typography`. Fixed with
   `missionPrefersTypography()` → `typoLocked`.
6. The migration number collided with another session's 686, filed one minute later the
   same afternoon. Renumbered to 689.
7. The port carried a dead component name — `defaultSectionsForPage`'s contact case
   returns `contact-hero`, which has ZERO `content_components` rows in any state, while
   18 live contact pages render `hero-contact` and the sibling about case already uses
   `hero-about`. The port corrects the transposition rather than seeding a dead name.
8. `fill_gaps` as the default mode was a no-op on the majority case — 33 of 57 sites had
   already been touched by the classifier, so a kit deferring to existing `design_intent`
   started nothing. Superseded by the owner's ruling (default `start`).

**MISSTEPS, 2026-09-02** — each caught by someone else, none by me:

- **I wrote "deployed … live since commit `0902039c0`" into the concept register when
  neither half was true.** `to_regclass` → NULL and no roll had happened. One command
  each. Caught by a Fable review.
- **My commit message got the same-file attribution backwards.** I said I had picked up
  another session's change; in fact their pathspec commit swept MY half-written rung, and
  HEAD did not compile for 51 seconds.
- **I over-stated a tripwire** — warned that a fix would overwrite a hand-seeded row;
  `siteSpecDeepMerge` means it would not.
- **Both my leading fix candidates for `bugs_open/438` were refuted**, and candidate 2's
  disqualification was a parenthetical I wrote *in the same sentence* and ranked anyway.
- **I told gamedesign.uk that `mission.preferred_palette` was the reliable lever.** It
  was — for the composition — and had no effect on the served site.
- **I quoted an all-time count from a rolling window.** Told `bugs_open/445` I had
  "verified independently — exactly 1 row" for `needs_new_layout_candidate`. It is **2**:
  `site_work_items` archives terminal rows. Worse than a miss — the check is a standing
  entry in my own auto-loaded memory index, which already noted it had failed to fire for
  me once before. The durable lesson is 445's: **independent corroboration is not
  protection when both parties inherit the same framing.** We each queried the table the
  question named. When a second check agrees, ask what BOTH checks assumed.
- **I diagnosed the chrome bottleneck from the wrong end** — measured the library and the
  pins, never the DATA. A pin SELECTS a component; nothing POPULATES it.
  `header-with-categories` needs ~12 template variables; designblog.co.uk's header
  `content_data` is empty, zero keys. An unsupplied variable renders blank. So the truer
  explanation of "36 of 37 sites render identical chrome" is not that nobody selected —
  **it is that nobody ever supplied the data for anything else.** Retracted to both lanes
  that had received the incomplete version.

---

## 2026-09-03 — liveness settled, then a council REVISE, then three false claims found

**Liveness established, all three facts** (the handoff's own §1 said this was unverified;
§0 settled it after the kubeconfig token was refreshed and the owner said go ahead).
Binary LIVE on `v1.0.1355` by `/proc/1/exe` capability probe with both controls;
migrations 689 + 691 APPLIED 2026-09-02 via a scoped `MIGRATIONS_DIR` run; adoption **0**.
691's three live sites verified appearance-neutral at the artefact before and after.

**Council round 1: `complete_revise`**, correlation
`bed139b2-f512-436a-9ba8-ff2fbfade8ef`, `decided_by: gating objection from guardian`, 4
reviewers abstained. Confirmed live today:

```
 current_step   |  status   |          updated_at
 complete_revise | COMPLETED | 2026-09-02 21:44:22.285319+00
```

**The objection was about the submission, not the code, and it was right.**
`editquality` noted the rationale cited a fixed typography cascade asymmetry while the
`apply_theme_kit_action.go` sketch showed only palette guards. The guard does exist
(`apply_theme_kit_action.go:330-341`, verified by reading it today) and my sketch made
the claim unverifiable. **A reviewer judges the submission, not the repository.**

**Then the more valuable half of the same round.** The reviewers ran their own check of
my `grounded_in` claim that `content_components.function` has "3 collisions after the
canonical predicate" and got **dozens**. Both figures are right under their own
predicate. Re-measured today with the queries written out:

| predicate | colliding functions |
|---|---|
| raw, every row | **84** as of 2026-09-03 |
| `is_active AND forked_from IS NULL` | **3** as of 2026-09-03 |

The three canonical collisions are `site-header`, `site-footer`,
`tool-agent-complexity-estimator`. **The denominator I had quoted was also stale**:
distinct functions are 425 raw / 410 canonical, not the 364 in the submission and the
register — stale by simple addition, exactly what the count-needs-a-date rule warns about.

**MISSTEP FOUND WHILE RE-MEASURING FOR AN UNRELATED REASON — two false claims, and they
were in a council submission and the concept register.** I had asserted that
`chromePinEligibleSQL` "EXCLUDES the confusingly-named `site-header`/`site-footer` rows
(`component_level='section'`) entirely" and that "the actually-eligible rows are
`header-theme-chrome`/`footer-theme-chrome`". What the database says:

```
 site-footer | section | t | t     site-header | section | t | t
 site-footer | site    | t | t     site-header | site    | t | t
 …                                 site-header | site    | t | f
```

`site-header` has active `component_level='site'` rows, so it IS chrome-eligible; and
`function LIKE '%theme-chrome%'` returns **0 rows** — I named components that do not
exist. Under the predicate as actually written, `site-header` is the **only** one of ten
chrome-eligible functions a name subquery resolves ambiguously (2 rows; the predicate
filters neither `forked_from` nor duplicates). **So the conclusion — hardcode verified
UUIDs — is unchanged and BETTER supported than the reason I gave for it.**

This is the **third** entry of the same shape from this lane: a right conclusion resting
on a wrong reason. Logged in `WRONG_CALLS.md`. The lesson I want the next session to
take: **a conclusion I already believe is exactly where I stop checking the evidence for
it**, and nothing external caught this — a reviewer objected to a *different* sentence
and the re-measurement swept it up.

**NEW FINDING, and it is the most decision-relevant thing here.** Re-verifying the
register entry end to end found a third thing wrong, this time about the world rather
than about a document:

```
           kit           | header_function | header_level | footer_function | footer_level
 brochure-formal-classic | site-header     | site         | site-footer     | site
 docs-technical          | site-header     | site         | site-footer     | site
 soft-editorial          | site-header     | site         | site-footer     | site
 tool-portal-light       | site-header     | site         | site-footer     | site
```

**All four seeded kits pin the chrome the default already picks.**
`ChromeSlotFunction()` (`component_library.go:386`) hardcodes `header → "site-header"`
and `footer → "site-footer"`, so an unpinned site gets exactly this. **The pins are
no-ops.** Same indistinguishability already recorded for the six pre-existing
`style_collections` pins, which all point at the default's own pick; the kits add four
more.

**Which completes a pattern worth stating plainly: three of the four dimensions a kit
bundles cannot change what a site looks like.** Palette cannot reach the stylesheet
(overlay is spec-wins on all 8 core slots, measured at the artefact);
`page_archetypes` governs at most 1 live page in 18; chrome is the no-op above. **Layout
is the only one left**, and two of the four kits pick a layout the matcher would have
picked anyway.

**Fixed and committed today** (`28aeb4ca0`): the layout rung recorded
`candidates: ["<kit layout>"]`, which `install_site_composition_action.go` writes through
as `lineage.layout_candidates`, reading as "one candidate was considered and won" when
the matcher never ran. Now `[]string{}`; the consumer guards on `len(cands) > 0` so the
field is omitted entirely. Verified: no test asserted the old value, the package builds,
gofmt clean with a malformed control file proving the check discriminates. **Not
emittable until a site adopts a kit** (adoption 0), so nothing in flight changes.

**Council round 2 resubmitted** on the same correlation with the typography guard
sketched, every predicate-dependent figure carrying a runnable query, the false chrome
claim retracted in `grounded_in` rather than quietly dropped, and the `candidates` fix
folded into the layout edit (same file — the sanctioned way to stay inside the 8-edit
cap). Dry run passed admission first, so the validation cost nothing.

**Two gofmt traps worth the line.** `gofmt -l <file>` exits 0 whether or not it lists
anything, so `gofmt -l f && echo BAD || echo OK` prints BAD forever — I read that as a
real failure for one turn before checking. Gate on `| wc -l`. And a comment placed
INSIDE a Go map literal breaks the alignment group, so gofmt rewrites every neighbouring
key; putting it above the `return` kept the diff to the one line that changed.

---

## 2026-09-03 (later the same day) — the retraction above was WRONG. `name` is not `function`.

**Correcting my own entry from an hour earlier.** The block above headed "MISSTEP FOUND
WHILE RE-MEASURING" says `header-theme-chrome`/`footer-theme-chrome` "do not exist" and
that the register's claim was false. **That is wrong. The register's original claim was
TRUE in both halves, and my retraction was the error.**

`content_components` has **both** a `name` and a `function` column, holding near-identical
vocabularies by design. I ran `WHERE function LIKE '%theme-chrome%'`, got 0 rows, and
concluded the components did not exist. They are `name` values whose `function` values are
`site-header`/`site-footer` — which is precisely the distinction the original claim was
drawing.

Resolved by id, which is what I should have done first (migration 689 names both UUIDs
three lines above the comment I was calling wrong):

```
 58fde68f-…-ea21cf27a9af | header-theme-chrome | site-header | site | t | unforked
 e6347680-…-1cea509159d1 | footer-theme-chrome | site-footer | site | t | unforked
```

**The full picture, and the query to use — select BOTH columns:**

```sql
SELECT name, function, component_level, is_active, forked_from IS NULL AS unforked,
       (is_active AND component_level IN ('site','header','footer','head')) AS chrome_eligible
  FROM content_components WHERE function IN ('site-header','site-footer')
 ORDER BY function, chrome_eligible DESC, name;
```

11 rows as of 2026-09-03, exactly **3** chrome-eligible: `header-theme-chrome`,
`footer-theme-chrome`, and **`header-leopardess` — an ACTIVE FORK of one client's
header**, eligible because `chromePinEligibleSQL` has no `forked_from` filter. The rows
named `site-header`/`site-footer` are `section`-level and ineligible, exactly as
originally written.

**So the reason to hardcode UUIDs in the seed is SHARPER than either version of the claim
said:** a function-name subquery for `site-header` returns two eligible rows and the
extra one is one client's fork. `bugfix_118`'s PLAN had already flagged `header-leopardess`
as this hazard.

**What caught it — not a query and not a reviewer.** Before pinging another lane about
"my" false claim, I grepped the tree for its propagation and found **70 files naming these
components, with UUIDs, and migration 339 carrying `RAISE EXCEPTION` drift guards on
updating them.** A component that does not exist does not have drift guards. The
corroboration was sitting in the repo I was about to correct.

**The chrome no-op finding SURVIVES, and its reason needed restating too.** All four kits
pin `header-theme-chrome`/`footer-theme-chrome`, and that is exactly the row
`ResolveChromeComponent` already returns for an unpinned site — established independently
by `bugfix_118`, whose register note records that after its fleet repoint
`GetComponentByFunction` and `ResolveChromeComponent` "already returned the same row for
both chrome functions". My first write-up said the pins "resolve to
`site-header`/`site-footer`", conflating `name` with `function` in the same way as the bad
retraction. **The finding is about row identity, not about a function string.**

**The pattern, stated for whoever picks this up.** Four write-ups today, four times a
conclusion that survived while the reason for it failed — and this one inverted into a
wrong conclusion from a right suspicion. A retraction reads as "someone checked", so it
carries more authority than the assertion it replaces, which makes a wrong retraction the
more expensive direction to fail in. **When the conclusion keeps surviving while the
reasons keep failing, the conclusion is coming from somewhere other than the evidence
being cited for it. Go and find where.**

**Owed as a result:** council round 2 was already in flight carrying the false retraction
in its `grounded_in`, so a **round 3 on the same correlation** is owed to withdraw the
withdrawal. Round 2's verdict had not landed at the time of writing (the only
`council_report` row is still round 1's, `revise`, 2026-09-02 21:43Z).

### Same afternoon, the fourth predicate lesson in one day — and this time I checked BEFORE asserting

Having established that `header-leopardess` is chrome-eligible and sorts alphabetically
before `header-theme-chrome`, I was one step from filing "every unpinned site resolves to
a client's forked header" as a live fleet defect. **I read the resolver instead of
inferring from the data, and it is already handled and documented in full**
(`component_library.go:336-378`):

| predicate | clauses | eligible rows for `site-header`/`site-footer` |
|---|---|---|
| `chromeEligibleSQL` — **pool SELECTION** | `is_active AND forked_from IS NULL AND component_level IN (…)` | **2** |
| `chromePinEligibleSQL` — **a PIN** | `is_active AND component_level IN (…)`, omits `forked_from` **on purpose** | **3** (incl. the fork) |

`forked_from IS NULL` is load-bearing in the pool predicate, and the source comment names
this exact row as why: *"an ACTIVE fork of one site's header is a candidate to become
every other site's header … `header-leopardess` sorts first among active `site-header`
rows and is what `link_site_components` would have assigned."* A pin omits it because
naming a site's own fork is precisely what a pin is for. `chrome_pin_test.go` pins the
asymmetry and goes red if the two predicates are made equal.

**So my "3 chrome-eligible rows" figure is right for PINS and wrong as a general
statement** — the fourth time today that a number was true only under a predicate I had
not named. Corrected in the register and the RUNBOOK with both predicates in a table.

**And the no-op finding is now airtight rather than inferred:** under the pool predicate
the only eligible row for each chrome function is `header-theme-chrome` /
`footer-theme-chrome`, which is exactly what all four kits pin. No tiebreak, no
alphabetical accident, no inference — `ResolveChromeComponent` returns the kits' pick for
a site with no pin at all.

**The thing I did right, recorded because the rest of this file is misses:** the
disconfirming read was cheap and I took it before writing the claim down. Three earlier
errors today came from asserting a mechanism from row data; this one came out differently
only because I opened the function.

### Council round 2 came back REVISE with a REAL DEFECT — the best thing this review produced

Verdict `revise` at 2026-09-03 15:32:59Z, `editquality` objecting on edit 2, 3 abstained.
**Accepted in full.** The objection cited landmines keyed to my own file: a design
preference pre-seeded into `design_intent` before a FRESH build is silently superseded by
`domain-research-classifier`, and `apply_theme_kit` writes
`design_intent.{palette,typography}`.

**Confirmed by reading the file. There is no guard** — `grep -n "classifier\|domain-research"`
in `apply_theme_kit_action.go` finds only comments about the owner's ruling, never a
predicate. `write_site_spec` supersedes the current row after a deep merge in which
**scalar keys are overwritten by the incoming value**, so the classifier discards the
kit's `reference_values`. Measured first-hand by the gamedesign.uk lane: manual
`design_intent` at 17:04:35 with `pinned=t` was `is_current=f` by 17:11:32, carrying a
different hex.

| dimension | fresh path | already-classified site |
|---|---|---|
| layout | **survives** — aspect `theme_kit_adoption`, which the classifier does not write | survives |
| palette | discarded — **moot for appearance**, no `design_intent` palette reaches the 8 core slots anyway | kit wins |
| **typography** | **discarded, and typography IS what renders** | kit wins, and renders |
| chrome | unaffected (own columns) | unaffected |

**`design_intent.<dim>.locked` does NOT cover this and it is important not to believe it
does.** `locked` is read when `apply_theme_kit` writes; nothing makes the classifier
respect it. The key survives the deep merge while the values do not, so the row ends up
**claiming a human pin over a classifier's values** — the most misleading end state
available.

**So the defect inverts the owner's framing.** He ruled *"by default it can start with a
theme"* — a kit as the starting point for a NEW site. As built, a kit works on a site that
has already been classified and is silently defeated on a new one.

**Recorded, not fixed, and deliberately so.** CONTRIB into `bugs_open/438` §6d, which
already owns this mechanism (§6a-bis) — a CONTRIB rather than a new bug file. Three
candidates, all costed there: make the classifier respect `locked` (architecture-scope, it
changes the classifier's write authority over a shared aspect, and belongs on 438's fix
list); also write `mission.preferred_typography` (builds on the defect — 438 is explicit
that mission's durability on that path is an *accident* caused by 438 itself, and it
collides with my own typography guard); or refuse/warn when no classifier-written
`design_intent` exists yet (cheapest, honest, my preference, and a behaviour change to a
live shared action that deserves its own round).

**A stale doc comment in my own file, found on the way.** The header said
`"fill_gaps" (default)` and never mentioned `start` — the pre-ruling behaviour, so the
file's own comment stated the **opposite** of the shipped default. Fixed, along with the
ordering hazard now documented there. Comment-only; verified first that no source-scanning
test makes that comment load-bearing.

### And a FIFTH instance of the name/function trap, in the same verdict, found by the reviewer

Round 2's reviewer verified my `grounded_in` claim that *"`contact-hero` has ZERO
`content_components` rows in any state"* and their query returned **1**. They were right.

```
name=contact-hero | function=hero-contact | component_level=section | is_active
-- one row, as of 2026-09-03
```

**One component, whose `name` and `function` are the same two words reversed.** So
`contact-hero` is not a dead name — it is that component's actual name, and `hero-contact`
is its function. Both strings resolve, because the section loader indexes by **both**
(`loadComponentSchemas`: *"Index by both name and function for fast lookup in the section
loop"*). **The seed is harmless either way and the port changes nothing functionally — but
the reason given for it was false.**

⚠ **This exact pair is the FOUNDING CASE of the LANDMINES entry about these two columns —
an entry this lane wrote on 2026-09-02.** I then made the error it warns about twice in one
submission, the next day, and a reviewer found the second one. The entry is now marked with
its fourth sighting and the new direction it did not cover (a false zero licenses a
*retraction*, not just a build).

**Round 3 submitted** on the same correlation: objection accepted with what I did and why
the remedy is deferred, the false retraction withdrawn, and the `contact-hero` claim
corrected. Run correlation `8e6f2aa8-ceae-4d22-a543-a47196f57193`; trail id stays
`bed139b2-f512-436a-9ba8-ff2fbfade8ef`.

### Two small missteps worth the line, both mechanical

**The backtick trap fired on a commit message.** Committing the workstreams-index entry I
used a plain double-quoted `-m "…"` containing `` `locked` ``, and bash **executed it** —
`/bin/bash: line 23: locked: command not found`, and the word is simply missing from the
stored message ("the  key does NOT protect against that path"). The file content is
correct, because that was written with a quoted heredoc. Forward-only, so no amend; the
meaning survives from context.

**The difference is exactly one habit:** every repo commit this session used
`-m "$(cat <<'EOF' … EOF)"`, whose **quoted** delimiter suppresses expansion, and the one
message I typed as a plain double-quoted string is the one that lost a word. This is a
standing LANDMINE ("backticks in `-m` execute") and it still caught me on the one commit
where I dropped the pattern. **Use the quoted-heredoc form for every message, especially
the short ones** — a long message gets the careful treatment and a one-paragraph one gets
typed.

**`gofmt -l` exits 0 whether or not it lists a file**, so `gofmt -l f && echo BAD || echo OK`
prints BAD for ever. I read that as a real failure for one turn. **Gate on `| wc -l`**, and
prove the gate discriminates with a deliberately malformed control file — which is what
finally distinguished "the file is dirty" from "my check cannot fail".

### Round 3: `revise` again — and the objection was that I had DEFERRED the fix

Verdict 15:56:39Z, gated by **`bug_historian`** this time, a different seat from rounds 1
and 2. The architecture seat also ran, which it rarely does. The objection, in substance:
the rationale itself confirms the fresh-path supersede for both palette and typography,
with typography the one dimension that renders; it is a diagnosed, accepted mechanism with
no guard anywhere in the file; and the plan explicitly ships without one.

**That is fair, and two rounds gating on the same missing guard is the council saying the
deferral was itself the defect.** My round-3 position — "the remedy owes its own council
round" — was defensible in isolation and looks like avoidance across three rounds. **So I
built it.**

**`b18091066`.** `classifierDesignIntentState()` asks whether
`domain-research-classifier` has EVER written this site's `design_intent` (current row or
superseded). If not, a classifier write is still ahead of us, and the apply records
`design_intent_supersede_risk` in **three** places: a WARN naming the mechanism, what
survives and what is lost and stating that `locked` is not a remedy; the
`theme_kit_adoption` spec, so it is durable and queryable rather than scrolling; and the
action's result, for a caller that reads only the result.

**It reports and does not refuse**, and I flagged that in the submission as the judgement I
most want challenged. Layout rides a different aspect, survives, and is the only dimension
a kit moves — refusing the whole apply would throw away the working part to protect the
broken one.

**A three-state string, never a bool.** `at_risk_no_classifier_write_yet` /
`classifier_already_wrote` / `unknown`. A read failure must not be recorded as "no risk".
Inventing a `false` for an unknown is the **same false-structured-fact class this very
submission already fixed twice** — a kit layout recorded as `library_match`, and a
candidate list of one that was never scored. Third time in one lane; at least this time it
was designed out rather than found.

**Proven by mutation, because a passing test is not evidence until it is shown it can
fail.** Two mutations, both red with the right message, restored green, results in the test
header:

| mutation | result |
|---|---|
| the at-risk arm always returns `already_wrote` (guard goes blind) | FAIL, "want `at_risk_…`" |
| the `err != nil` arm returns `already_wrote` not `unknown` (confident wrong answer) | FAIL, "want `unknown`" |
| restored | PASS |

A second test pins the three constants distinct and self-describing — a copy-paste
collapsing two of them leaves every state assertion green, so nothing else could catch it.

**And the predicate was chosen so it could come out otherwise:** `[MEASURED 2026-09-03]`
of **39** sites carrying a `design_intent`, **38** have one written by the classifier — true
on an established site, false on a fresh one. Had it been 39 or 0 it would not discriminate
and the guard would be decoration. The same query settled the column question:
`source_agent` is nullable and eleven other writers appear, hence
`COALESCE(source_agent, source)`.

**Round 4 submitted** carrying the guard. Run correlation
`425d4365-3342-490a-b32f-cbd1ec5d014c`.

**The thing to carry forward about this gate.** Rounds 1, 2 and 3 all returned `revise`,
from three different seats, and **every one was right**: a claim my sketch did not support,
then a real defect in live code, then my deferral of its fix. CLAUDE.md's line that "a
REVISE round is cheaper than the defect it finds" was paid out three times on one
correlation. **Do not read a repeated revise as the loop spinning without reading it.**

### Round 4: APPROVED — with 27 objections, and three of them were answerable in minutes

Verdict `approved` at 2026-09-03 16:19:31Z, *"approved with 7 advisory objection(s) — none
high-severity"*, 3 abstained. **Four rounds on one correlation: revise, revise, revise,
approved.** Every revise found something real, which is the whole argument for the gate.

**The trailer is now legitimate** — `Council-Reviewed: bed139b2-f512-436a-9ba8-ff2fbfade8ef`
— because an approved verdict has been READ. The earlier commits carry
`Council-Submitted:` and are credited automatically by 098 now the correlation has
approved; forward-only forbids amending them and none is needed.

**Three objections were cheap to answer and I answered them rather than filing them.**

**(1) `debug_historian`, medium — "verified only by unit-test mutation, no pod
verification".** They were right, and the answer is decisive against me:

| needle | in the running binary |
|---|---|
| `classifierDesignIntentState` (the round-4 guard) | **absent** |
| `at_risk_no_classifier_write_yet` | **absent** |
| `apply_theme_kit` (positive control — Phase 1) | **PRESENT** |
| `zzz_not_a_real_symbol_zzz` (negative control) | absent — *the probe discriminates* |

Pods are **174 minutes old**, started well before the 16:03 commit. **So the guard is
COMMITTED AND NOT LIVE.** Phase 1 is live; the guard rides the next roll. I had written
"inert today because adoption is 0" — true, and it was inert for a second and stronger
reason I had not checked. **This lane's own memory index carries "live and committed are
independent facts"; the seat had to remind me anyway.**

**(2) `prior_art_librarian`, medium — "per the Choice-B precedent" is an uncited
existence claim.** Right, and **the claim is false**: there are **two** Go writers of
`sites.style_collection_id`, not one — `install_site_composition_action.go` and
`SelectStyleCollectionAction` in `v3_site_actions.go`. Logged in `WRONG_CALLS.md`; the
narrower claim that was the actual point (a kit installs nothing, it queues
`needs_composition`) is unaffected. **The seat predicted the class of error from my own
track record inside this correlation** and was right.

**(3) `prior_art_librarian`, low — "nothing makes the classifier respect `locked`" is an
asserted absence.** Also right that it was asserted rather than shown. **Now shown:**
`designIntentLocked` has readers at exactly three call sites, all in
`apply_theme_kit_action.go` (lines 464, 465, 469), and nothing else in `platform/`,
`internal/` or `cmd/` reads a `design_intent.<dim>.locked` key. The other `"locked"` hits
are unrelated mechanisms — the section-editor lock, the component lock guard, adopt
fidelity. The claim holds; it just had no evidence attached.

**The objections that are NOT answerable and should shape Phase 2** — the architecture
seat, both medium, and they read as a gate rather than a note:

- *"All four seeded kits pin chrome identical to the unpinned default — the chrome
  dimension of a kit is currently a no-op. **Shipping more kits or adopters before this is
  addressed overstates what a kit does.**"*
- *"Palette cannot reach the served stylesheet under the current render-overlay precedence
  — `theme_kits.palette_id` is **structurally decorative**. This is an architecture gap,
  not a bug in this plan, but it **should block further palette-bearing kit adoption**
  until the precedence is fixed or the capability is explicitly dropped from the
  contract."*

That is the council independently reaching this handoff's §2 conclusion and going one step
further: **do not adopt kits until the contract says what a kit actually delivers.** Taken
into the handoff as a gate on Phase 2.

Others worth carrying, not acted on: `reuse_agent` asks why a new bundling table rather
than extending `style_collections` (fair — `style_collections` already bundles per site);
`constitution` objects that `apply_theme_kit` re-implements supersede-then-insert instead
of reusing `WriteSiteSpecAction`, which my own risk (5) admitted; `bug_historian` notes the
guard detects and does not prevent, *"a recorded user decision with no enforcement point is
decorative"*; and `constitution` also objects to my rationale's tone — all-caps headers and
"dramatized process narrative" where plain engineering prose was called for. That last one
is fair and I would write it flatter next time.

---

## 2026-09-04 — the round-4 guard went LIVE overnight, and the check that proved it is the one the handoff prescribed

First action of the next session, exactly as `HANDOFF_2026-09-03` §7 asked: roll, then
re-probe. The roll had already happened, and **the probe flipped the answer**.

```
pod agent-chassis-ffc9ddff9-jvw92   Running   13h      (was 85c4984f77-* — a different ReplicaSet)
  classifierDesignIntentState   PRESENT      <-- was ABSENT on 2026-09-03
  apply_theme_kit               PRESENT      <-- positive control, Phase 1
  zzz_not_a_real_symbol_zzz     absent       <-- negative control, the probe discriminates
```

**So the guard is live.** Everything else re-verified and unchanged: `to_regclass` returns
both tables, 4 kits and 14 archetypes seeded, **adoption still 0**, and **0** adoption rows
carry `design_intent_supersede_risk` — consistent with adoption 0, since nothing has
applied a kit and the guard has therefore never had occasion to fire.

**The distinction that survives:** the guard is now LIVE AND UNEXERCISED, not inert. Two
reasons for inertness existed on 09-03 (not in the image, and nothing adopting); one has
cleared. **The first kit ever applied is the test** — read
`design_intent_supersede_risk` on the resulting `theme_kit_adoption` row.

**Worth recording as method, not just state.** On 2026-09-03 this lane wrote "THE GUARD IS
COMMITTED AND NOT LIVE" with a dated pod probe and both controls. It was true when written
and false within hours. **A `[VERIFIED]` marker with a date and controls is still a
snapshot** — it makes the claim checkable, not durable. The handoff now carries the
correction visibly rather than a silent overwrite, and its §0 says the file has been stale
in *both* directions inside 48 hours.

**A landmine from another lane that lands on this one's work** (`a7352e2ca`, written the
same day): a resubmitted council round **shares its correlation with the round before it**,
so the verdict query the trigger itself prints — `ORDER BY created_at DESC LIMIT 1` — can
hand back the OLD verdict and read as your revision being rejected. This lane happened to
read its four verdicts correctly by **counting rows** (`reports=3` → `reports=4`) rather
than taking the latest, but that was not deliberate rigour so much as wanting the whole
trail. **The rule is now in the handoff §4: count the rows against rounds submitted; never
`LIMIT 1` across a resubmission.**
