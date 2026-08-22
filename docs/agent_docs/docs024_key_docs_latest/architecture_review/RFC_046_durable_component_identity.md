# RFC_046 — a component row's identity is INFERRED **five** different ways (as of 2026-08-22) and stamped none

**Status:** OPEN — raised 2026-08-22 by the `bugfix_357_component_identity` lane, at the explicit
recommendation of the council gate's `architecture` seat (trail `62aac6c2`, round 2).
**Source bug:** `bugs_open/357`. **Lane:** `docs024_key_docs_latest/bugfix_357_component_identity/`.

> *"This is the fourth distinct heuristic-identity mechanism layered into
> `save_page_sections_action.go` (stub detection, Layer 2 slot-name carry-forward, shrink/floor
> guards, now data-component matching). Recommend an RFC scoping a durable component-instance
> identity (e.g. a stamped identity token independent of `slot_name`/`position`) rather than a
> fifth heuristic layer next time this class recurs."*
> — `architecture` seat, council round 2, 2026-08-22

## 1. The property that does not hold

A `page_components` row carries three things: an **identity** (`component_id`, `slot_name`), the
**bytes** it serves (`rendered_html`), and the **content** those bytes were made from
(`content_data`). Every page-composition path writes, carries and re-writes all three as one bundle.

> **No seam anywhere asserts that the three agree, and no writer stamps which component actually
> produced the bytes.** Identity is re-derived, by inference, at every hop.

## 2. The five inferences — **5** as of 2026-08-22, all live

> **This is a census of call sites, so it goes stale by ADDITION** (owner ruling 2026-08-22). Re-run before
> quoting the number: `git log --since=2026-08-22 --diff-filter=A -- platform/orchestration/actions/` — a
> non-empty result means a new writer may have added a sixth, and the count must be re-derived, not repeated.

| # | where | what it infers identity from |
|---|---|---|
| 1 | `saveSectionsExtractFromHTML` (:1427, :1463) | the `data-component` attribute in the HTML, when present |
| 2 | `saveSectionsExtractFromHTML` fallback (:1453–1476) | nothing — emits the sentinel `"section"` for *unknown* |
| 3 | `enrichSectionsWithPlannedNames` (:1778) | **position**: `planned[Position-1]` from `pages.sections` |
| 4 | `enrichSectionsWithComponentIDs` (:1493+) | **fuzzy name matching**: suffix-strip, underscore/hyphen, `-hero` LIKE |
| 5 | Layer 2 carry-forward (:517) | **slot-name string equality** against the incoming set |

Two of these are the whole of `bugs_open/357`: #2 produces "I don't know", #3 converts it into a
confident wrong answer, #4 resolves that answer to a shared component's UUID, and the row is
persisted claiming to be a `hero` while holding a whole interactive tool. **22** live rows as of 2026-08-22, newest born that same day on a site homepage.

**And #5 is why the obvious repair is not safe.** It matches stored rows to incoming ones on slot
name and nothing else, so *correcting* a row's identity makes the next plan-driven rebuild miss the
match, take the `default:` re-append branch, and serve the incoming hero band **beside** the tool.
A byte-preserving re-type changes what four live sites serve. That is not a bug in the repair; it is
the system having no way to say "this row is that component's output" other than by its name.

## 3. Why a sixth inference is the wrong move

The `bugfix_357` lane proposed exactly that — comparing the component template's `data-component`
against the stored HTML's. It is the best of the six (measured fleet-wide 2026-08-22: **1,550** agree, **0**
disagree, **24** genuine defects, and it drops the ~131 template-drift false positives the naive
prefix test produces). **It is still an inference**, and it inherits the class's defects:

- **it is silent for 190 of 339 components** (as of 2026-08-22), which declare no attribute at all;
- it cannot vouch for the very component the repair would introduce (a `{{.body}}` passthrough
  declares no attribute), so the fix's own output is outside its reach;
- and every inference added makes the next author's model of "how does this system know what a
  component is?" strictly harder, which is how five became five.

The council approved that predicate's *measurement* and still gated the round twice — first because
the change was unsafe, then because the safe version stopped nothing. **That oscillation is the
signal.** There is no version of an inference-based fix that is both safe and effective, because the
thing being inferred is knowable and simply is not recorded.

## 4. The shape of the answer

**Stamp identity at the point of production, and make every later hop read the stamp rather than
re-derive it.** When a renderer produces bytes from a component, it knows exactly which component
and which instance; that fact should be written with the bytes and never inferred again.

Sketch, deliberately not a design:

- a stamped token on the row — component id **and** instance — written by whatever produced the
  bytes, independent of `slot_name` and `position`;
- carry paths (Layer 2, `carryStoredSection`, `extractSectionsFromMetadata`) move the stamp with the
  bytes instead of re-deriving a name, which also makes #5's matching exact rather than string-based;
- the fallback for genuinely unidentifiable input stays *honestly unknown* — the sentinel is correct
  and it is the conversion of the sentinel into a plan-derived name (#3) that is not;
- inferences #1–#4 become a **birth-time** concern only, for input that arrives unstamped, and their
  output is recorded as inferred rather than as fact.

**Existing machinery to reuse rather than reinvent:** `page_component_history` (**26,965** rows as of 2026-08-22, from
2026-03-16, already carries `slot_name`, `rendered_html_digest`, `op`, `application_name`);
`rendered_html_digest` (the same-statement stamp from `bugs_open/229` / IMP-052, which already
asserts "reproducible from content_data" and is written only by the render/save seam); and
`CLC-014`'s per-instance component scoping (`{{.InstanceID}}`), which is the closest thing to an
instance identity the estate already has.

## 5. What this RFC is asking for

A decision on **whether the estate wants stamped component identity at all**, because the honest
alternative is to accept inference and stop pretending otherwise:

1. **Stamp it** (the seat's recommendation). Highest cost, and it is the only option under which the
   357 repair is provably safe rather than argued to be.
2. **Accept inference, and consolidate.** One inference function, called by every hop, replacing
   five — no new guarantee, but the next author has one place to read. Cheaper, and it does not make
   the repair safe.
3. **Neither.** Then `bugs_open/357`'s 22 rows stay as they are, correctly parked, and the class
   recurs — which is a defensible call as long as it is a call, and not the default that arrives by
   nobody deciding.

**What this RFC does NOT ask for:** approval of any repair to the 22 existing rows. That changes
what four live sites serve and is the owner's decision, separately, whichever option is chosen here.

## 6. Evidence

All measured 2026-08-22 against the live database and HEAD; queries and their traps are in
`docs024_key_docs_latest/bugfix_357_component_identity/RUNBOOK_component_identity.md`.

- The population: **22** rows as of 2026-08-22, 4 sites, newest `2026-08-22 08:50:12` (`vetcomparison.uk` `index`, a
  homepage). `hero` is planned first on all 22.
- The writer, settled by fingerprint: all 22 carry `position=1` and
  `content_brief.section_guidance='hero section'`; `save_page_sections_action.go` is the only
  `page_components` writer that writes `content_brief` at all.
- The conservation loop: `vetcomparison.uk` `index` has six completed rerenders 08-19 → 08-22 with
  the tool serving throughout, and its rows were re-created **inside** the 08-22 rerender window
  (`08:44:51` → `08:50:19`, rows at `08:50:12`).
- No tool has been destroyed by this: searched `page_component_history` with a control — **182**
  slots interactive-and-still-interactive as of 2026-08-22, and of the 17 that changed, **15 GREW**. None is in the
  357 population.
- Council trail `62aac6c2`: round 1 REVISE (gated by `bug_historian`), round 2 REVISE (gated by
  `editquality`, seconded by `bug_historian`), six seats approving cleanly in round 2.
