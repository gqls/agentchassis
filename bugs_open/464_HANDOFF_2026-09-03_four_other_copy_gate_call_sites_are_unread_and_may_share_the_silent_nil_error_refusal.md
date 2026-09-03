# 464 — FOUR other copy-gate call sites are UNREAD, and may share the silent `(map, nil)` refusal that `bugs_open/442` just fixed at one of them

**Filed** 2026-09-03 by the `bugsweep_2026_08_26` lane, **at the council's explicit request**.
**Status: OPEN. Unowned.**

> **Resolve by SLUG** (`four_copy_gate_call_sites_unread_silent_refusal`) — bug numbers collide on
> this tree; `git log` the FILE PATH, not the number.

**This bug exists because a reviewer refused to let an unfinished audit be written up as
finished.** The `bug_historian` seat, reviewing `bugs_open/442`'s fix (corr
`76288ff9-3cde-46e6-b65a-22564fac8f6d`, round 2), objected [medium] and filed a `MISSING` that
named the remedy:

> "The plan's own audit found 5 files calling the copy gates (`ScanVoice*`/`checkBannedClaims`)
> but only checked, **by grep intersection**, whether they share this action's silent `(map, nil)`
> refusal shape — and admits it did not read the other four … recommend that follow-up be **filed
> as a numbered bug** before this closes the register entry as done."

That is this file.

---

## 1. What the mechanism is, in plain terms

The platform has **copy gates**: a per-site voice gate (banned phrases, style rules) and a
fleet-wide banned-claims sweep. Several actions run them over text before saving it.

`bugs_open/442` found that when the gates refuse in `save_page_meta_description`, the action
returns `(map, nil)` — **a nil error**. So the step SUCCEEDS, the orchestration COMPLETEs, the
scheduled task stamps a clean run, and the page silently keeps whatever it had (nothing). That is
fixed **for that one action**, live on `v1.0.1359`.

**The question this bug exists to answer: do the other four callers do the same thing?**

## 2. The population, and exactly how far the check got

`[MEASURED 2026-09-03]` files calling `ScanVoiceSingleValue`, `ScanVoice(` or `checkBannedClaims`:

| file | read? | status |
|---|---|---|
| `save_page_meta_description_action.go` | **read in full** | the 442 case — **FIXED**, live `v1.0.1359` |
| `section_editor_regulated_guard.go` | **NOT read** | unknown |
| `save_sections_claims_guard.go` | **NOT read** | unknown |
| `rewrite_negations_action.go` | **NOT read** | unknown |
| `validate_page_content.go` | **NOT read** | unknown |

**What WAS done, and why it is not enough.** A grep intersection: files containing
`"(updated|saved|applied|written)": *false` cross-referenced against the copy-gate callers. Only
`save_page_meta_description_action.go` appeared in both sets.

⚠ **That is a grep intersection, not a read, and it is unsound in a specific way:** it keys on
four result-key SPELLINGS. **A caller using any other key name — `refused`, `blocked`, `skipped`,
`ok`, a bare `status` string — is invisible to it and returns a clean-looking absence.** The
original 442 defect was itself found by reading, not by grep, and the same file's §4 records a
grep that failed to reproduce its own author's number.

**So: the four are UNREAD. This file does not claim they are affected, and does not claim they are
clear.**

## 3. Why it is worth reading them rather than assuming

- **The refusal shape travels with the GATE, not with the field** (`442` §7). Whatever a caller
  saves, if it consults the same gates it faces the same "what does a refusal look like from
  outside" question.
- **Two of the four have shapes that hint either way** and neither hint is evidence:
  `save_sections_claims_guard.go` logs `CLAIMS FLOOR BLOCKED` (a Warn — the same surface 442 was
  about), and `rewrite_negations_action.go` returns `map[string]interface{}{"status": …}, nil`
  markers, which is a *different key spelling* and therefore exactly what §2's grep cannot see.
- The estate's most-repeated failure family is 016b §9's *"a silent fallback deploys a hollow
  section as success"*. Four unexamined members of a family with a known instance is the shape
  worth spending an hour on.

## 4. How to work it

Read each of the four and answer three questions per file, **in the code, not by grep**:
1. When the gate fires, does the caller **error**, **block**, or **return a value with a nil
   error**?
2. If it returns, **does anything downstream assert on that value?** (442's finding was that a
   nil-error refusal is only half the defect; the other half is that no consumer reads the result.)
3. If it is silent, is the damage **durable** (a page ships wrong / stays blank) or transient
   (the next run retries)? Only the durable ones are 442's class.

**Do NOT re-run the grep intersection and call it done** — that is the check this bug exists
because of.

## 5. If one is affected, the remedy already exists and is proven

`bugs_open/442` §10 is the worked pattern, live since `v1.0.1359`: file a work item **at an actor
that can repair it**, not into a handler-less review queue.
⚠ **The number that decides where to file** `[MEASURED 2026-09-03, live UNION archive]`: items
**with** a `handler_agent` complete at **83%** (56,315); items with **none** at **17%** (6,699,
989 parked). A flag-only `needs_human_review` row looks like a fix and is the 17%.

## 6. Provenance

**Not run through `090`, and no cross-cutting root cause is asserted** — this file's whole content
is a *stated absence of knowledge* about four named files plus the method for closing it, which is
the opposite of a confident structural claim. The one measured claim (the five-file population) is
a grep over the tree, reproducible above. Raised by an independent reviewer with no access to this
tree, which is corroboration that the gap is real rather than my own scruple.

**Related:** `bugs_open/442` (the instance, FIXED and live) · `bugs_open/320` (the owner
requirement that added the gates 2026-08-19) · `bugs_open/338` / **CQ-035** (single-value gating) ·
016b §9 silent-fallback family · concept register **SEO-004** / **SEO-008**
