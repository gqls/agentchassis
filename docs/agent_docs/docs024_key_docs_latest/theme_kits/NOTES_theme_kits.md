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
