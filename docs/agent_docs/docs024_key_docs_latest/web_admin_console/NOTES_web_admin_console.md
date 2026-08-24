# NOTES — web admin console

Running record, append-only, newest at the bottom. Missteps included on purpose.

---

## 2026-08-23 — a build-break I attributed wrongly, twice, in opposite directions

While committing the `/c/` prefetch guard, `go build ./internal/core-manager/...` failed with
three `not enough arguments in call to emitRequiredFieldsMissing` errors in
`render_site_components_action.go` and `section_editor_actions.go`. I put a line in the commit
message: *"do NOT build/roll core-manager until HEAD compiles again."*

**That line is now FALSE, and a stale blocker is worse than no note** — this estate has already
been bitten by an "inert until the roll" line that left a detector switched off for nine days
after its blocker cleared (`LANDMINES.md`). So, plainly: **the build is clean. core-manager can
be built and rolled.**

### What actually happened, and both of my readings were wrong

1. **First I said it was "another session's in-flight work".** Reasonable, but unchecked.
2. **Then I checked and said the opposite** — the two failing callers were not dirty and neither
   was the definition file, which pointed at committed breakage on HEAD. Also wrong.
3. **The decisive check settled it: `git archive HEAD | tar -x` into a temp dir and build there.**
   HEAD compiled fine. So the fault was in the working tree after all, and reading (1) was
   right — but I only knew that after building a tree with no working-tree changes in it, which
   is the only way to separate the two on a shared checkout.

The owner is the **`bugs_open/342`** lane (an absent required field rendering empty and silent
at 13 of 15 render call sites). It was mid-refactor: `emitRequiredFieldsMissing` had gained a
`pageContext` parameter in `work_items_common.go` while its callers had not caught up in that
session's tree. **It fixed itself within the hour** — `eb918bd58` and the commits after it — so
there was nothing to chase and nobody to nudge.

### The transferable bit

**On a shared working tree, "the build is broken" is not a fact about the repository until you
have built a tree with no working-tree changes in it.** `git status` on the failing files is not
enough: the file whose *signature* moved can be committed while the callers that need updating
sit in someone else's uncommitted edit, or the reverse, and both look identical from a status
line. The one-liner:

```bash
T=$(mktemp -d); git archive HEAD | tar -x -C "$T"; (cd "$T" && go build ./...) ; rm -rf "$T"
```

**And a transient break needs a re-check before it goes in a commit message**, because the
message cannot be amended (forward-only) and the claim outlives the condition by days.

---

## 2026-08-24 — sent to correspond, not to build; and a pattern language that ate my evidence

The owner was asked (by me, from a two-day-stale context) whether to expose the console as
`admin.apis.uk`, keep it VPN-only, or gate it. He answered neither: *"Please correspond with the
webdesign.uk live webdesign thread that has built a web facing console."* **The console was
already live.** My session had been handed `PLAN_2026-08-22` and had not re-read the lane before
forming a question — the lane had produced six commits and a new handoff in between.

**The check that would have caught it before I asked: `ls -lt` the lane directory.** It takes
one second and would have shown `HANDOFF_2026-08-24_continue_here.md` written 20 minutes
earlier. I ran it only after the owner's answer. On this tree a plan handed to you names the
state it was written in, not the state you are in.

### Re-measured three of the 08-24 handoff's own falsifiers (2026-08-24 ~11:15Z)

- `https://admin.apis.uk/` → **302** to `billowing-smoke-5ed4.cloudflareaccess.com`; the meta
  JWT decodes to `auth_status: NONE`, `hostname: admin.apis.uk`. Live and gated. CONFIRMED.
- `https://www.apis.uk/` → **301**, `location: https://apis.uk/`. §3's redirect rule **is
  applied** — the handoff still lists it as owner-pending. One falsifier closed.
- `https://links.webdesign.uk/c/x` → **could not resolve host**. §2's box steps are **not**
  applied, so `/c/` has not moved off the shopfront and the parking-page-rule landmine is still
  the only thing holding it. Still owner-pending, correctly listed.

### The misstep worth the entry: `LIKE '%__step_error%'` is not a literal search

Sizing the `bugs_open/099` landmine for the build-steps screen, I counted COMPLETED rows whose
`collected_data` mentions `__step_error` with `LIKE '%__step_error%'` and got **315**.

**In SQL `LIKE`, `_` is a single-character wildcard.** The pattern actually asked for "any two
characters followed by `step_error`". The honest count via `strpos` is **176**. The key whose
distinguishing feature is a double underscore is exactly the key `LIKE` cannot be trusted with —
the wildcards sit precisely where the evidence is.

Then I compounded it: I assumed the 176 − 67 gap between "literal anywhere" and "top-level key"
was **nesting**, and wrote that the top-level test misses 109 real errors. It does not. One query
extracting 320 characters around the literal showed the gap is **workflow configuration naming
the field** — `"note_body_field": "__step_error.message"` inside an `append_doc_note` step. The
top-level jsonb test is exact: 67 real `"__step_error":` keys, all top-level, 67 = 67.

So I twice reported a fabricated defect in someone else's correct design. **Both errors share one
cause: I believed a count produced by a pattern without reading a single row it matched.** That
is now the check — read one matching row before quoting any pattern-derived count. Recorded in
`WRONG_CALLS.md`; the finding itself is in `PLAN_2026-08-24_build_steps_screen.md` §6b, written
up in the direction that survived.

### What that measuring pass did produce, all in `PLAN_2026-08-24_build_steps_screen.md` §6

The plan's §2 states `orchestration_states` has no site column. It has one, with **three**
indexes on it, populated on 2,136 of 4,410 rows — and of the two JSON paths §2 proposed instead,
one matches **zero** rows. That is its only backend change, made smaller and safer. §6d is the
one that may reshape the screen: `execution_path` is empty on **100%** of rows, so there is no
recorded step sequence, and a site's `site_id`-tagged orchestrations are mostly its periodic
sweeps rather than its build — the thing the owner calls a build is a `site_work_items` chain
(`082_submit_domain_unified.sh`). I did **not** act on that; it is this lane's call.
