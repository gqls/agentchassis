# HANDOFF 2026-08-15 — continue here (silent_hero_logo_readers)

**Supersedes `HANDOFF_2026-08-13_continue_here.md` for state** (that file remains the record of the
269 close and the 08-14 banners). Written after 273 closed, so a fresh chat can pick up with no
other context.

**Read in this order:** this file → whatever §2 row you are acting on. Everything else is history:
`NOTES_…` (technical log), `README_where_we_are.md` (owner's plain-prose log),
`SUMMARY_2026-08-14_…` (last milestone read-out).

## 1. One-paragraph state

Every tooling defect this lane found is now FIXED AND LIVE: 261 (symbol spellings), 267
(impossible whole-file advice), 269 (bare method handles → wrong body), and 273 (the dead-end
marker that demanded names it withheld — closed 2026-08-15 on `v1.0.1300`, build point
`a2a691213`, both replicas probed with controls, council APPROVED `ba3f6047` first round). The
original bug, 236 (hero/logo lose `image_url`), has a CONFIRMED mechanism — the awaiting-park
copies only three fields and discards collected state — and is parked on an owner decision
(RFC_012). Nothing in this lane is blocked on further investigation.

## 2. The whole remaining surface, in one table

| item | state | next action | who |
|---|---|---|---|
| `bugs_closed/273` live-behaviour witness | fix LIVE, **zero bundles of any demand since the roll** — the behaviour zero is unreadable | when any diagnosis touches an over-budget file (`coordinator.go` natural), run 273 §5: the sibling section must carry `The elided handles:`. Also re-check 267 §4b's trend then | any session, opportunistic |
| `RFC_012` — awaiting-park discards collected state (= 236's fix) | OPEN, owner decision `(a)`/`(a′)` | surface, don't work | **owner** |
| `RFC_027` — `path:Symbol` handle grammar has no owner (4 bugs: 189/261/267/269) | OPEN, owner decision; "acceptable as is" is a legitimate ruling | surface, don't work | **owner** |
| 261 §8 follow-up 3 — `knownScopeIdentities` omits `values` (`diagnose_route_action.go:541`) | open, cosmetic (wastes an embedding call per package-level value; loses no evidence) | small fix + council, any time | any session |
| `bugs_open/236` | mechanism confirmed 08-14 (see its final contribution); file stays OPEN pending RFC_012 | nothing until the ruling | — |
| 269's collision-halves live witness | still unwitnessed (no collision file scoped by any live run yet); tests cover it | opportunistic, alongside the 273 witness | any session |

## 3. Proving a deploy on this service — the recipe that actually worked on 08-15

The §2 recipe in the 08-13 handoff still holds (precheck first; stamp is a startup line and
rotates). What's new: **when the stamp is out of log range, the safe binary discovery is to
extract every 40-hex string and intersect with git objects** — junk from Go's digit table is not
a git object, so this dodges the `strings`/discovery-grep landmine:

```bash
kubectl -n ai-persona-system exec <pod> -- sh -c "grep -aoE '[0-9a-f]{40}' /proc/1/exe | sort -u" > cands
while read h; do git cat-file -e "$h^{commit}" 2>/dev/null && echo "STAMP: $h"; done < cands
git merge-base --is-ancestor <your-commit> <STAMP> && echo IN
# control: a post-roll commit must be absent from the binary and NOT an ancestor
```
Measured 08-15: 78 distinct hex strings, exactly 1 a real commit. Probe **both** replicas.
⚠ Do NOT loop 60 `grep -aq` execs against `/proc/1/exe` — it times out; one `grep -aoE` pass is
the shape that works.

## 4. Traps carried forward (the ones still worth a fresh chat's attention)

1. **A zero needs its demand control named in the same sentence** — 273 §9 records a perfectly
   healthy zero that proves nothing (0 bundles since the roll). Every count over
   `diagnosis_artifacts` also needs its retention window printed (30-day clock).
2. **Sketch elision draws council blood** — three rounds running (`REVISE` once, medium twice).
   Show disputed hunks verbatim in submissions.
3. **The bundle-census phrases are load-bearing text**: never introduce "did not fit" /
   "could not be read" / "read it whole" / "NO next_scope can render this path" into new bundle
   wording — the 267 §4b trend queries discriminate on them.
4. **A fixture's arithmetic is an empirical claim** — assert it inside the test; an estimated
   fixture failed against correct code once already this lane (NOTES 08-14 late).
5. **`git mv` needs BOTH paths on the commit**, and verify with
   `git ls-tree -r --name-only HEAD | grep <slug>` → exactly one line, the new path.
6. Council submissions: `operation` must be `modify|add|remove|config_change` — `create` is
   refused client-side.

## 5. Files this lane owns

```
bugs_closed/261_… §8 = the follow-up ledger        bugs_closed/267_… §9 = live proof
bugs_closed/269_… §11 = live proof                 bugs_closed/273_… §8 = council round, §9 = live proof
bugs_open/236_…  (parked on RFC_012)               architecture_review/RFC_027_…
platform/orchestration/actions/diagnose_assemble_bundle_action.go   (+ its 3 test files incl. deadend_tail)
internal/analysis/symbolbody.go                    docs026 register: diagnosis-loop.md (DIAG-043)
```

Council correlations, newest first: 273 `ba3f6047-a2e5-4ce6-ac0e-edf0bb88c4e3` (APPROVED) ·
269 `e5809ca9-d718-44f6-8d27-6d8cd656dd28` (APPROVED) · 267 `ac23f2f7` (APPROVED).
