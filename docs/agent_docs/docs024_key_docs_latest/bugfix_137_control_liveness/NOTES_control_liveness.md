# NOTES — control-liveness / runtime-fill scope (`bugs_open/137`)

Append-only, newest at the bottom. Missteps are the point, not an appendix.

---

## 2026-07-31 — picking the bug

Scanned `bugs_open/` (67 files) against the 34 `.jsonl` transcripts active since
16:00. Bug-number mentions are **very** noisy — most are `ls bugs_open/` output
scrolling past in someone else's session — so a bare mention count would have
ruled out almost everything. Filtered to sessions with >2 hits, which left six
open bugs with genuinely zero engagement: **113, 114, 115, 118, 132, 137**.

Then checked by **code symbol**, not by number, per the standing rule that every
ownership check is lagging. Three sessions had hits on
`check_tool_acceptance|evaluateStaticCriteria|DeadControlAnchors|IsNoopHref|attribute_absent`:

- `1917ffde` — finishing the vonc provocation seal; cites the 137 disagreement
  in passing, does not work it;
- `631baa00` — a component/arena rename, reading `attribute_absent`'s landmines;
- `9cb23339` — loancalculator tool values, citing the file's header.

None was reconciling the judges. Took 137.

Rejected the others for cause, not by preference:
- **113** is already fixed in code (`3096a55a6`) and awaiting post-roll
  verification — not a fix task.
- **132**'s remaining fix is a Cloudflare Worker whose source is in neither
  repo; it cannot be done from here.

## The bug's premise, re-checked before touching anything

```
curl -s https://vonc.com/provocations/index.html | grep -o '<a[^>]*href="#"[^>]*>'
→ <a class="provocations-archive__item" data-archive-template hidden href="#">
```
Exactly one hit, as the bug says. Still valid three days on.

## What reading the code added

The bug locates the disagreement in one function. The grep says the *exemption*
is inlined in eight places, all as `strings.Contains` over whatever the caller
passed — so the exemption's blast radius is set by **caller chunking**, not by
the markup. `save_sections_link_repair.go:67-71` had already worked this out and
fixed it **at its own call site**, recording the reasoning. That is the strongest
possible support for moving it into the predicate: the right answer was known
and unenforceable.

---

## MISSTEP 1 — I measured the wrong population and nearly wrote it down

Found `vonc.com/index` has two dead `href="#"` CTAs in `brief-explanation` (not a
shell) while `lobby-grid`/`provocation-card` are shells, and started to write
that up as "masked from `RepairPageLinks`".

**Wrong, and the code says so.** `RepairPageLinks` only touches `LinkScopePage`
and `LinkScopeEmpty`; `#` classifies as `LinkScopeAnchor` and that path never
looked at it. The masking is real but it belongs to a *different* consumer
(`check_dead_controls`, and the tool-acceptance sweep) — and `check_dead_controls`
reads one component at a time, so it was never masked at all.

**Caught by:** reading `ClassifyLinkScope` before writing the claim, rather than
after. Re-measured the population that `RepairPageLinks` actually governs —
**empty** hrefs in a non-shell component on a page carrying a shell — which gave
one live row: `vonc.com / index / gauntlet-cta`, 2 empty hrefs.

**The transferable bit:** "component X is masked" is not one fact. Each consumer
masks a different *class*, and the class is decided by a function the symptom
does not name.

---

## MISSTEP 2 — a SQL join on a non-unique key, and the output nearly read as a finding

To list the components on pages carrying a shell I wrote
`WHERE page IN (SELECT page FROM pc WHERE is_shell)` — joining on `p.name`.
Page names are **not unique across sites**, so it returned 75 rows: every page
called `index` in the fleet.

**Caught by:** the output being obviously too big and spanning 16 domains. The
data I needed was still visible in it, which is exactly why this was dangerous —
a smaller wrong answer would have looked right. Rewritten to join on `page_id`.

---

## MISSTEP 3 — a toy fixture produced a repair that does not exist

First live probe ran `RepairPageLinks` against an index I hand-built with two
URLs, and reported **3** repairs — the two real empty hrefs plus an unlink of
`/provocations/index.html`.

That third one is an artefact of **my own two-row index**: the URL resolves fine
against the real 18-row `pages` set. Had it reached the bug file or the council
submission it would have read as a finding, and it is not one.

**Caught by:** noticing the third repair named a page I knew existed. Re-ran
against the real page set: **2** repairs, output 48,956 → 48,866 bytes.

**This is the narrow-filter family again** — the fixture I invented defined the
answer. The check that would have prevented it is the one I eventually did:
build the index from the same table production reads.

---

## The measurement that stands (live artefacts, 2026-07-31)

| page | bytes | old exemption | new exemption | dead controls | link repairs |
|---|---|---|---|---|---|
| `vonc.com/index` (assembled) | 48,956 | **100%** | 2 spans, 6,172 B (**12.6%**) | **2** — "Get Started", "Learn More" | **2** — the two empty hrefs, 48,956→48,866 B |
| `vonc.com/provocations-index` | 7,684 | **100%** | 1 span, 1,400 B (18.2%) | 0 | 0, byte-identical |

The second row is the reconciliation: the 137 element is exempt under **both**
judges, and the page is otherwise untouched.

The two newly-visible repairs are `gauntlet-cta`'s "Enter the Gauntlet" and
"Find Your Archetype" — **the same controls `check_dead_controls.go` names in
its own header as the case that check was built for.** They had migrated from
`href="#"` to `href=""` at some point, which moved them from the sweep's class
into the repair path's class, where the page-wide skip was hiding them.

## Fleet-wide blast radius, measured

Across all deployed `page_components`: **exactly one** page has a non-shell
component holding empty hrefs alongside a page-mate shell. Small — and stated
plainly rather than dressed up, because the case for this fix is structural, not
volumetric. Separately, **zero** served *tool* pages currently carry either a
shell or a no-op href, so the tool-acceptance sweep's exemption masks nothing
there **today** — it is the mechanism 137 names, and it was unguarded.

## Tests proven load-bearing by mutation

- Mutant 1 (restore the whole-document span): **8** failures, including both
  consumers' "neighbour" cases.
- Mutant 2 (delete the exemption): **10** failures, including the *pre-existing*
  `TestRepairPageLinks_RuntimeFillShellIsExempt`.

Both directions matter: a suite asserting only mutant 1's failures would pass
against a change that simply removed the exemption.

## One pinned expectation inverted

`TestEvaluateStaticCriteria_AttributeChecksFlowThrough` asserted FAILED for the
shell-enclosed template row, and its comment noted the sweep was suppressed on
the same element — **the contradiction was pinned as an expectation**. Inverted
to SKIPPED with the date and reason in the test. Flagging it explicitly in the
council submission rather than letting a reviewer find it in the diff.

## Incidental finding, not mine to fix

`cmd/reasoningset` **does not compile at HEAD** — committed and clean, three
`declared and not used` at `main.go:504`, last touched `b82b3d8b4` (07-28
"v1.0.1188 prior to merge in main"). `go build ./platform/... ./internal/...
./pkg/...` is clean, so no service is affected. Recorded here so the next thread
that runs a broad build does not spend time on it believing it caused it.
