# Should locks expire? — investigation and recommendation

You asked whether time-based lock expiry is worth implementing. Below is what I
found in code/docs, the case for and against based on actual failure modes,
and a proposed sequencing.

Short answer: **yes, time-based expiry is worth doing**, but as a deliberate
project paired with auto-lock policy, not a one-line schema add. Doc 004 v4
designed it correctly; only half landed in code, and the half that did land
makes the missing half more important.

---

## What's in code, what's in docs

### Code (production)

`platform/orchestration/actions/check_component_lock.go`:

```go
// Hard locks: set by humans, only humans can remove
switch lockedBy.String {
case "admin", "admin-removed", "checkpoint":
    status.IsHard = true
default:
    // "deploy" or anything else is a soft lock
    status.IsHard = false
}
```

Discovery and audit data-loading queries: all use `pc.locked_at IS NULL` as
the filter. No time comparison anywhere. Confirmed across
`check_unlinked_page_components`, `check_undeployed_assets`, the empty-section
finder, and the visual/content auditor data queries.

Every lock is permanent until something explicitly clears it.

### Docs (design intent)

Doc 004 (improvement loop) v4 changelog dated 2026-03-31 explicitly lists
"lock types with expiry" as a v4 change. The body of doc 004 specifies:

| Lock type | Behaviour |
|---|---|
| `permanent` | Never expires; manual unlock only |
| `timed` | Expires after N days (default 90) |
| `review` | Creates HITL review item on expiry |

with implementation as `lock_type` and `lock_expires_at` columns and the
filter expansion:

```sql
AND (locked_at IS NULL OR (lock_expires_at IS NOT NULL AND lock_expires_at < NOW()))
```

Doc 007 (adoption pipeline) v3/v4 repeats the same design.

### The gap

The columns don't exist in schema. The filter expansion isn't in any query.
The expiry mechanism was specced and never built.

But — critically — **the pass-counter reset that doc 004 paired with lock
expiry IS implemented**. Look at the improvement-loop workflow:

```
load_pass_count → check_audit_pass_limit
  if pass_count >= 3 → notify_scheduler_clean → complete_clean
```

And the auto-reset triggers (time-based, direction change, major rebuild,
manual re-audit) are wired up. So the system already has the "pass counter
resets after 60 days" half of the rhythm. It just doesn't have the matching
"locks release after the same window" half.

---

## The case for expiry — what breaks without it

Three failure modes the current "permanent unless manually cleared" model
produces.

### Failure mode 1: lock proliferation from auto-lock-on-deploy

Doc 013 confirms: every direct edit + deploy auto-locks the affected
component. This is a deliberate UX choice — humans shouldn't have to
remember to lock things they've worked on. But it's an aggressive policy:

- Fix a typo → component locked, permanently.
- Adjust a single value in a form field → component locked, permanently.
- Tweak the wording of a CTA → component locked, permanently.

After six months of this, a typical site has many components locked under
`'admin'` for reasons that no longer matter. The improvement loop's surface
area shrinks monotonically. Eventually the loop has nothing meaningful to
work on.

This isn't hypothetical. The auto-lock-on-deploy trigger fires `BEFORE
UPDATE` on every page_components write (per the trigger declaration in
the schema), so anything that flows through the dashboard editor gets
locked.

### Failure mode 2: forgotten human locks

Six months after a manual edit, the user doesn't remember:
- Which components they edited.
- Why they locked any particular one.
- Whether the lock is still warranted.

Without expiry or a reminder mechanism, the user has to actively audit
their own locks to discover what's stale. Most users won't.

### Failure mode 3: evergreen-but-stale (your specific case)

A page that's "good enough" today might be suboptimal in twelve months
because:

- Design conventions evolve (typography, spacing, layout patterns).
- Site direction has shifted (new competitor, new audience, refined positioning).
- Brand voice has matured.
- Content gaps that were acceptable now look like genuine omissions.

The lock that protected the page from rewrite in 2026 actively prevents
the loop from reconsidering it in 2027. The user has to know to revisit.

### What 004 v4 saw

Doc 004's "natural rhythm" passage describes the intended dynamic:

> Build → audit × 3 → cap reached → site quiet
>   ... 60 days ...
>   → pass counter resets, expired locks release
>   → improvement loop runs fresh
>   → finds new issues (content aged, design dated, new opportunities)
>   → audit × 3 → quiet again
>
> Sites breathe and improve rather than constant churn or permanent stasis.

This is the design. **Half of it is wired up, half isn't.** The pass
counter resets, but with no locks expiring, the freshly-reset audit has
nothing to act on for any component that was ever locked.

---

## The case against (or for not doing it the obvious way)

Three reasons to be careful, not three reasons to skip it.

### 1. Some content really shouldn't expire

Brand identity, legal disclaimers, contact details, product names — these
should remain locked indefinitely once approved. The naive "everything
expires after 90 days" rule would produce embarrassing regressions
(brand voice drift, legal copy auto-rewritten, etc.).

This is exactly what `lock_type = 'permanent'` is for in 004's design.
The right default for routine edits is `'timed'`; the right opt-in for
critical content is `'permanent'`.

### 2. The "review" type creates HITL work that may not be wanted

A `'review'` lock that creates a `needs_lock_review` work item on expiry
is a lovely idea, but it adds operational load. If the user doesn't want
to see "we're about to revisit your work" notifications, they have to
mute the type. Workable but not free.

For the imagery work specifically, `'review'` is probably more than we
need today.

### 3. "Auto-release after N days" needs the right N

Different content has different natural lifespans:
- Hero copy: maybe 6-12 months.
- Tool calculator output: indefinite (logic is logic).
- Blog headlines: very short (news ages fast).
- Image of a robotic gripper: ageing slowly, maybe 2-3 years before
  visual style trends make it obsolete.

A single `default_expiry = 90 days` is wrong for half the content. The
duration should depend on content type or be configurable per-row.

004's default of 90 days is a reasonable starting point but the system
should support overrides.

---

## Recommendation

### For imagery (immediate work)

**Don't add timed expiry to assets in Phase 2A.** What I shipped is correct
for the assets table today — `locked_at` + `locked_by`. Hard vs soft
distinction is enough for the imagery loop:

- `'manual'` → hard, never auto-revisit.
- `'visual-design-auditor'`, `'imagery-quality-auditor'` → soft, agents may
  revisit on next audit cycle.
- `'audit-pending'` → transient, the agent that set it should clear it on
  completion. (Mechanism: write the lock at audit-start, clear at
  audit-finish. Doesn't need expiry.)

This is consistent with what 013 already documents for page_components
and site_components.

### For the broader system (deferred, separate project)

**Decision (2026-05-08):** Implement timed expiry as a future focused
project, after the imagery loop work is bedded in. Approved policy:

| Source | Default lock_type | Default expiry |
|---|---|---|
| `'admin'` (human edit via dashboard) | **`permanent`** | — |
| `'admin-removed'` | `permanent` | — |
| `'checkpoint'` (explicit human approval) | `permanent` | — |
| `'deploy'` (auto-lock-on-deploy) | `timed` | NOW() + 30 days |
| `'manual'` (asset upload) | `permanent` | — |
| Auditor approvals (visual-design, imagery-quality) | `timed` | NOW() + 90 days |
| `'audit-pending'` | not a lock — clear immediately on completion | — |

**The substantive change:** only auto-locks (`'deploy'`) and auditor
approvals get timed expiry. **Human-set locks remain permanent.** This is
deliberately more conservative than my initial draft — the original
proposed `'admin'` going to 90-day-timed, which would have silently
expired typo fixes and small edits along with deliberate hand-crafted
content.

Under this policy, 004 v4's "natural rhythm" still works for the
machine-generated parts of the site (auto-locks and auditor approvals
release after 60-90 days, the loop revisits, sites stay improvable),
while every human edit stays protected indefinitely. The user retains
control via manual unlock when they want to revisit a hand-crafted
component.

The "lock proliferation" concern doesn't disappear under this policy —
human edits still accumulate permanent locks. But that's a deliberate
user choice ("I want every edit to be sticky") rather than an inadvertent
side effect. A dashboard "show locks older than X" view becomes more
important under this policy because users need a way to find and review
their old locks. That's downstream UX work, not part of the schema project.

### Implementation sketch (future project)

When the project is taken up:

1. **Schema** — add `lock_type text` and `lock_expires_at timestamptz` to
   all four lock-bearing tables: `page_components`, `site_components`,
   `site_plan_directives`, `assets`. Single migration, single transaction.

2. **Auto-lock writers** set `lock_type` per the policy table above.
   Most writers are setting `'deploy'` or auditor names — those default
   to `timed` with the appropriate expiry.

3. **Filter update** — single sweep replacing `locked_at IS NULL` with
   `(locked_at IS NULL OR (lock_type = 'timed' AND lock_expires_at < NOW()))`.
   Catalogued from grep — about 8-10 callsites, all mechanical.

4. **Helper update** — `CheckComponentLock` returns `LockType` and
   `LockExpiresAt`. Hard locks ignore expiry. Soft locks honour it.

5. **Discovery check** — new `expired_review_locks` check creates
   `needs_lock_review` work items for `review`-type locks past expiry.
   This is the HITL hook 004 v4 described.

6. **Dashboard view** — show locks with their classification (hard /
   soft / timed / permanent) and expiry status. Most of this UI work
   is downstream and can ship later.

### Ordering

- **Now**: keep Phase 2A as shipped. `locked_at` + `locked_by` for assets
  is the right shape; it's forward-compatible with adding `lock_type` +
  `lock_expires_at` later.
- **Next**: complete the imagery loop (Phase 2B/C/D for multi-image,
  Phase 4/6 for auditor work).
- **Then**: the lock-expiry project as a single coherent piece across all
  four lock-bearing tables. Mirror 004 v4's design exactly, with the
  conservative policy above (human edits stay permanent).

Doing it as a unified change avoids the situation where one table has
`lock_expires_at` and others don't, and consumers have to know which is
which. Coherence beats incremental delivery here.

---

## Why this matters more than it sounds

Without expiry, the system has a slow-motion problem: every auto-lock is
permanent. Sites become progressively less improvable over time. The
improvement loop's value proposition — "agents keep your site good as
the world changes around it" — is undermined by the very locking
mechanism designed to protect human work.

Doc 004 v4's author saw this and designed the right answer. The answer
just didn't get all the way to the schema. Worth finishing.

**Note (2026-05-08):** This is anticipatory rather than empirical. The
platform is in early stages of publishing, no real-world cases of
unimprovable sites observed yet. The case for the project rests on the
design having a clear gap that will manifest as the system scales, not
on observed pain. That's a perfectly reasonable basis for the work — it's
cheaper to fix a known gap before it bites than to clean up afterwards —
but it does mean the project shouldn't displace work that addresses
observed problems.

---

## Updates needed to existing docs after this investigation

### `031_locks.md`

Three additions:
1. **Hard vs soft locks subsection** — currently missing. Pattern A
   consumers have no idea the `locked_by` value matters.
2. **"Time-based expiry" subsection** — replace the misleadingly simple
   "today's locks are all effectively `permanent`" with: "today's locks
   have no time-based expiry. Hard vs soft is encoded in `locked_by`.
   The richer `lock_type` / `lock_expires_at` design from 004 v4 / 007 v4
   is specced but not implemented; see [LOCKS_findings_and_proposed_corrections.md
   or wherever this lands]."
3. **Add `assets` to the "Where locks live" table** — already in
   PROPOSED_031_locks_addendum.md from earlier, folds in here.

### `phase_2a_assets_locking.sql`

Rewrite the docstring for the third (and final) time. Describe:
- Detection: `locked_at IS NULL` (unlocked) vs `IS NOT NULL` (locked).
- Classification: by `locked_by` (hard = `admin`/`admin-removed`/
  `checkpoint`; soft = `deploy` / `'manual'` / others).
- No time-based expiry today. `lock_type` and `lock_expires_at` are
  documented in 004 v4 / 007 v4 as future work; will be added uniformly
  across all four lock-bearing tables when implemented.
- Reference `actions/check_component_lock.go` as the canonical Go-side
  implementation.

### `PLAN_imagery_loop_closure.md`

Phase 2.1 already has the correct columns in my prior pass. Add a note
that timed expiry is a separate project; imagery's needs are met by
hard-vs-soft (encoded in `locked_by`) for now.

---

## What I'm not changing

- `auto_lock_on_deploy` trigger semantics. Leave as-is until the timed
  expiry project picks them up.
- Phase 2A migration. It's correct as shipped.
- The improvement-loop workflow. It already has the audit-pass-reset
  half of the rhythm; the lock-expiry half is added in the future
  project, not in imagery work.
