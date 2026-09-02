# CONTRIB 2026-09-02 — DECISION 1 IS RULED (YES), AND IT IS RE-HOMED TO THIS LANE

**From the `routing_capability_guard` lane (`bugs_open/395`), which is closing.** This is not a request
for work and not a proposal. It is a **ruling plus a measurement**, handed to the lane that owns the
subject matter, because holding it in my file was an accident of where the blockage was found.

**Read `bugs_open/320` §15 first** — it is the ruling this one supersedes, and the reason this needs
your judgement rather than mine.

---

## 1. The ruling, in the owner's own words

Asked whether an **automated finding may cause a published page description to be REWRITTEN**, the
owner ruled (2026-08-26): **YES, but only where the description was machine-written, never
human-authored copy.** He added, unprompted: ***"I haven't yet written any manually."***

Re-confirmed 2026-09-02, on the question of building it: ***"I'd say yes."***

**Why this is yours and not mine:** `320` §15 records him granting `overwrite_existing: true` for a
**one-off** 681-page pass and **explicitly withholding it for the standing mechanism** — then that
being verified afterwards, with the seeded agent left unarmed. This ruling reverses that withholding.
The column, the backfiller and the one-off precedent are all yours.

## 2. ⚠ THE FINDING THAT MATTERS: the authority you are being granted ALREADY EXISTS

**This is the thing worth carrying across, and it makes the build far smaller than the ruling sounds.**

`[MEASURED 2026-09-02]` `save_page_meta_description_action.go:211` is **not** an unconditional UPDATE.
It is guarded **twice**, in series:

```go
overwrite := datahelpers.GetBoolField(config, "overwrite_existing", false)
// $3 = overwrite. When false the row is only touched if it is currently blank.
UPDATE pages SET meta_description = $2, updated_at = NOW()
 WHERE id = $1 AND ($3::bool OR COALESCE(meta_description, '') = '')
RETURNING id
```

1. **`overwrite_existing`, an opt-in config field, DEFAULT FALSE**, enforced **inside the WHERE
   clause** rather than by a read-then-write in Go — so it cannot race a concurrent writer.
2. **AND** the backfiller's scheduled `pre_query` selects `COALESCE(p.meta_description,'')=''`.

`sql.ErrNoRows` is handled as a **refusal**, not an error — `{"updated": false, "reason":
"already_has_description"}` — so a caller that forgets the flag gets a clean no-op, not a surprise.

**Consequence for scoping your build:** this is **not** "build a path that can overwrite published
copy". That capability is built, and built in exactly the shape the 2026-08-02 owner ruling prescribes
for new authority on a shared seam (opt-in field, unsafe default OFF, decision visible at the caller).
**What is missing is only a work-item-driven ROUTE that sets the flag.**

⚠ My own lane's roster described this write as *"the only unconditional UPDATE"* for a week, and so
did a shipped code comment. Both are now corrected. If any of your docs carry that phrasing, it is
wrong in the direction that makes this job look bigger and riskier than it is.

## 3. ⚠ The condition the ruling rests on, which the system CANNOT currently enforce

The owner's ruling distinguishes machine-written from human-authored descriptions. **`pages` carries
no provenance of any kind** — no source column, no author, no stamp — so that distinction exists only
in his sentence.

`[MEASURED 2026-09-02]` **40 of 1,153** active pages have a blank description; none is hand-written,
and here is the stronger form of that: **no human-facing path can write the column at all.**
`internal/core-manager/admin/page_admin_handlers.go` has four `UPDATE pages` sites (writing
`suppressed_sections` and `page_spec`), so the admin surface exists and reaches this table —
**`meta_description` is not among the columns it writes, and there are ZERO mentions of the column in
the whole of `internal/` and `frontends/`.**

**So today the ruling's narrow option covers everything, because a human COULD NOT have written one.**
That is a real guarantee resting on a **dated absence**, and it goes stale by ADDITION.

**The condition to carry forward — not a mechanism to build now:** whoever builds a human-facing
editor for `pages.meta_description` must add the provenance mark **in the same change that ships the
editor**. Building the marker before an editor exists produces a permanently-NULL column, and a
guard reading `authored_by IS NULL` then permits everything for ever while looking like a control.
The marker design is `bugs_open/403`'s (leopardess lane) — three-state: `__authored` FORBIDS, minted
LICENSES, neither = today's behaviour. **Cite `403`'s own commits (`0049b10d9`, `24cc44ed1`), not this
file.**

## 4. What is NOT settled, and is yours to decide

The owner ruled the authority. He did **not** rule the shape. Open:

- **Which producer files the repair item**, and under what `item_type`.
- **Per-site opt-in, or fleet-wide?** `320` §15's one-off used a full backup table
  (`meta_description_pre_regen_20260821`) and a reversible single UPDATE. `scripts/regen-meta-descriptions.sh`
  is the worked precedent for how the authority was handled last time.
- **Whether a canary site comes first.** My recommendation, offered only because I was asked to
  recommend: yes.

**A council round is warranted** — it is a standing path that rewrites published copy, which is
precisely the authority `320` §15 withheld.

## 5. Provenance of this contrib

- Ruling: owner, 2026-08-26 and 2026-09-02, via the `routing_capability_guard` lane.
- The overwrite-flag finding: that lane's handoff §12; the no-human-writer census: §2.1a.
- ⚠ **That lane is CLOSING.** Do not route questions at it. Everything above is in
  `docs/agent_docs/docs024_key_docs_latest/routing_capability_guard/HANDOFF_2026-08-26_continue_here.md`
  (§2, §12, §19, §21) and in `bugs_open/395`, both of which outlive the session.
