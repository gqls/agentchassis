# PLAN — bug 327, the build trigger that can publish nothing and exit 0

Design, phasing, decisions **and their reasons**. Corrections live here, marked.

---

## The problem, in one line

The customer-facing entry point of the one-shot site build can print every identifier, exit
0, and send nothing — and the operator cannot tell that from success.

## Decisions, and why

**1. A library, not another documented recipe.** The remedy has been in `LANDMINES.md` since
2026-08-16. `[MEASURED 2026-08-23]` 25 of 218 publishers print a `PUBLISH_OK` receipt and
**2** assert on it. Twenty-three authors read the guidance, followed it, and still exit 0 when
the receipt is absent. **Guidance-to-copy is not a delivery mechanism**; the remedy had to
become callable. Modelled on `scripts/council-scope.sh`, the only other sourced fragment in
this repo, and born from the same drift class.

**2. The receipt must be ASSERTED, not printed.** This is the whole finding. On the failure
mode that actually bites, the exit code carries **no information** — so a printed receipt is
not belt-and-braces, it is the only instrument, and an uninspected instrument reads nothing.

**3. Two exit-code families, because (a) and (b) demand opposite responses.** *Never
published* → retry at once. *Published, not consumed* → wait; a duplicate costs a round, and
that state is also ordinary latency's signature, which CLAUDE.md tells sessions not to retry
on. Encoding the distinction in the exit code is the point of the change, not a nicety.

**4. Codes start at 10.** Callers already own 1 and 2 — `097`'s documented contract is
"1 = hard error, 2 = REFUSED out of scope". A publish outcome must never collide with a
caller's own vocabulary.

**5. No automatic retry.** On an indeterminate receipt an auto-retry is a double-publish
engine: the one case where you cannot tell whether the message landed is the case where
resending is most dangerous.

**6. Phase 0 + 1 only** (owner's choice). The class is not swept. Mass-editing one-shot lane
scripts risks **re-publishing live triggers**, and 23 of them cannot even run.

**7. The Go submit path is filed, not built.** `platform/kafka.Producer` is already
`RequiredAcks: RequireAll, Async: false`, so an in-cluster submit path would make a silent
drop **unrepresentable** rather than merely detected, and could close the ~2-day forensic
window. It needs an image build, a council round and a fleet roll; it protects nobody this
week. Recorded in **OPP-009**'s verify-later as proposed-and-unbuilt.

## Corrections to this plan, made while executing it

> **CORRECTED — the two "Phase 1 siblings" are not scripts.** The plan named
> `077_submit_domain.sh` and `076_trigger_build_pipeline.sh` **from their filenames, without
> reading them**. 076 is a 1,101-line notes file with no shebang; 077 fails `bash -n` at line
> 117 and hardcodes `DOMAIN="idea.uk"` after the argument parse. Neither can run; neither was
> migrated. "Fixing" a file that cannot run would make it *look* runnable — a new trap, not a
> closed one.

> **CORRECTED — my census over-counted by ~2×.** "200 scripts using the racing form" counts
> *files containing the pattern*. `[MEASURED 2026-08-23]` of 201 such files, **178 parse** and
> 105 are also executable; **23 cannot run at all**. The honest exposure is **178**, and I
> quoted the flattering number.

> **CORRECTED — the bug file's own verification recipe cannot fail.** It proposes an
> unreachable broker and a non-zero exit; the **unfixed** script already exits 1 that way. The
> silent arm is empty stdin against a *healthy* broker: zero messages, exit 0, no output.

> **NOT AS EXPECTED — the race did not reproduce.** 0 of 10 old-form publishes lost on
> 2026-08-23, which excludes the historical 4-in-5 (p≈1e-7) but bounds today's rate only below
> ~26%. So the new form is **structurally immune**, not *demonstrated to beat the race*. In the
> same run the old form published a **duplicate** (11 delivered for 10 sends) — a pathology
> nothing in the bug file or the landmine anticipated. One observation, not a rate.

## What ships

| layer | artefact | status |
|---|---|---|
| 1 | `scripts/kafka-publish-lib.sh` (+ `--self-test`) | built, 11/11 offline, every runtime arm proven live |
| 1 | `082_submit_domain_unified.sh` migrated | done; induced-failure verified |
| 2 | `check_kcat_stdin_race` in `pattern-check.py` | done; 1.7% over 300 commits, 5/5 true positives |
| 2 | ratchet test in `cmd/config-key-audit` | **not built** — nothing runs `go test` automatically here (no CI), so its only real value was council visibility |
| 3 | in-cluster Go submit path | **filed, unbuilt** (OPP-009) |

## Known gaps, stated

- **The load-bearing artefact ships unreviewed.** The council was asked and REFUSED on scope
  (exit 2): everything is under `scripts/`. Not forced — that is the owner's call.
- **Adoption is one caller.** 178 runnable racing publishers remain. Top of the queue is
  `scripts/trigger-landmine-verifier.sh` — the tool that verifies landmines publishes with the
  racing form and reports "0 failed to publish" from the one signal absent on the silent arm.
- **Exit 11 is verified at the classifier only**, not end-to-end.
- **V6 (healthy end-to-end submit) deliberately not run** — it creates a site and consumes
  every stage `item_key` for ever (`bugs_open/326`).
