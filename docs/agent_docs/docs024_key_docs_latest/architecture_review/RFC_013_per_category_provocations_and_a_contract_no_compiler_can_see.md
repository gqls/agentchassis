# RFC 013 — Per-category provocations, and the feed contract no compiler can see

**Status: OPEN — filed 2026-08-05, awaiting an owner ruling.** Blocks
`provocation_pipeline` PLAN §9.2 (categories). Nothing is built and nothing is
committed against it; this RFC exists so that the design is agreed *before* code,
which is what PLAN §9.2 itself instructs:

> **Do not design multi-category rotation without agreeing this with them first** —
> it is not a vonc-side-only change, and discovering it late means a live mismatch
> between the page and the engine.

**Filed by** the `provocation_pipeline` lane. **The other lane this binds is
`gauntlet_dead_cta`**, which owns `internal/tools-api` (PLAN §8). This RFC is also
the notification that owner ruling 2026-07-29 §3 requires — *"a shared mechanism's
OTHER consumers must be told, not merely measured"* — and it is written to be read
by them, not only by the owner.

---

## 1. Problem + evidence

### 1.1 What is being asked for

Owner, 2026-07-31: *"I want it to be hugely popular so I think we will need to
have different categories sooner rather than later, ranging from current political
opinions to pets etc each with a different target audience."*

Today the platform publishes **exactly one** live provocation per domain. Several
simultaneous category dailies cannot be expressed in the shape the engine reads.

### 1.2 The contract as it actually is [MEASURED 2026-08-05, first-hand]

PLAN §9.2 states the constraint correctly but not sharply enough to design
against, so it was re-derived by reading the code end to end. **The engine's
entire validation of the feed is three checks**, all in
`internal/tools-api/handlers/round.go` `FetchProvocation`:

| line | check |
|---|---|
| `round.go:73-76` | the key `today` exists at the top level |
| `round.go:78` | its value is not the four bytes `null` |
| `round.go:78` | its value is not zero-length |

That is the complete list. What happens to the value afterwards:

- `round.go:44` returns it as `json.RawMessage` — **opaque bytes, never parsed**
- `round.go:114` → `store.CreateRound` → `rounds.go:39` `INSERT INTO
  gauntlet_rounds (…, provocation, …)`, column typed `jsonb`
  (`sql_for_agents/198_tools_api_gauntlet_rounds.sql:28`)
- `position.go:67` and `defend.go:67` interpolate `string(round.Provocation)`
  **directly into the AI prompt as text**

Two negatives, checked rather than assumed, because the whole argument below
depends on them:

- **Nothing in `internal/tools-api` unmarshals the provocation.** The only
  `Unmarshal` in the path is `round.go:70`, which parses the *envelope*
  (`map[string]json.RawMessage`) to find the `today` key — never the value.
- **No field of `today` is named anywhere server-side.** `headline`, `teaser`,
  `slug` and `detail_body` return zero hits across `internal/tools-api/**.go`
  excluding tests. The field names exist only in the publisher, in three
  front-end loaders, and in `209_experience_gauntlet_round5_gaps.sql`, which
  records the contract as prose: *"provocation is the verbatim 'today' object
  from /data/provocations.json — do not invent extra fields"*.

> **⚠ THE CONSEQUENCE, AND IT INVERTS THE OBVIOUS CHOICE.** Changing `today` from
> an object into a map of categories **passes all three checks**. It is not null,
> it is not empty, the key is present. The engine would persist the whole map as
> the round's provocation and paste it into the model's prompt. **There is no code
> path on which that returns an error.** The failure is silent, and it surfaces as
> an AI arguing against a blob — degraded round quality, which is the hardest
> possible symptom to attribute.

For contrast, this platform has already been bitten by the *reverse* of this and
fixed it deliberately: `FetchProvocation`'s doc comment says it fails loud on an
absent `today` *"per bug_historian pattern #7 rather than returning a blank
provocation"*. The loud path was designed. **A shape change slips underneath it.**

### 1.3 The coupling is invisible to every static tool [MEASURED]

```
go list -deps ./platform/orchestration/actions | grep tools-api   →  (no rows)
```

The chassis publisher imports **nothing** from `internal/tools-api`. The two
sides share `platform/httpguard` and `platform/aiservice`, and **no type that
describes the feed**. The contract is a JSON file fetched over HTTPS plus two
prose comment blocks that must be kept in agreement by hand.

So: no compile error, no type check, no shared package, and — per §1.2 — no
runtime check either. **This is RFC 007's pattern again** ("chrome eligibility
needs a package both sides can import"), and it is why this is an RFC rather
than a patch.

---

## 2. The question this RFC actually asks

**2.1 Which shape carries N live provocations per domain?**
(a) one feed file per category, or (b) `today` becomes a keyed map.

**2.2 Who changes `tools-api`, and when?** It is the `gauntlet_dead_cta` lane's
code. This RFC does not propose that the provocation lane edit it.

**2.3 Should a round record WHICH category it argued?** `gauntlet_rounds` has no
category column. Rounds are already being published publicly
(`/tools/gauntlet/round.html?r=<slug>`), so this is a decision about a durable
record, not a transient one — cheap now, a backfill later.

**2.4 Should the contract become a shared Go type** (RFC 007's remedy), or stay
prose? This is separable from 2.1 and can be ruled on independently.

---

## 3. Blast radius, stated properly (bar 2)

**Consumers of the `today` key** — the complete set, by grep for
`provocations.json` across Go, JS, HTML and templates:

| consumer | owner | reads | breaks how, if `today` changes shape |
|---|---|---|---|
| `round.go` `FetchProvocation` | gauntlet_dead_cta | envelope only | **silently** — §1.2 |
| `position.go:67`, `defend.go:67` | gauntlet_dead_cta | raw string → prompt | **silently** — degraded prompt |
| `provocation_card_loader.js` | social001 / vonc | `today.headline`, `.body` | blank card, no console error |
| `lobby_grid_loader.js` | social001 / vonc | `arena.cards` | unaffected (not `today`) |
| `provocations_archive_loader.js` | social001 / vonc | `archive.entries` | unaffected (not `today`) |
| `provocation_feed_action.go` `checkFeed` | provocation_pipeline | all four `today` fields | **loudly** — refuses to emit |

The only component that enforces `today`'s shape is the publisher's own
`checkFeed`, i.e. the writer. **The readers enforce nothing.** A writer-side
invariant cannot protect a reader in a different binary on a different host.

**Not derivable from `go list`**, and that is the finding, not a gap in the
method: the coupling is over HTTP, so the mechanical derivation available is the
grep above plus the import proof in §1.3.

---

## 4. What is already done, so the RFC asks only for the judgement

- **`provocations.category` exists** (`sql_for_agents/282_provocation_pool.sql:62`),
  `NOT NULL DEFAULT 'general'`, and migration 282's own comment already states
  the constraint this RFC asks about. It was added early on purpose (PLAN §9.2:
  "nearly free now and expensive to retrofit").
- **It is inert by design.** `loadProvocations` does not select it and
  `selectForDate` does not read it, so nine rows all sit in `general` and no
  behaviour depends on the column today.
- **Nothing needs undoing.** Whichever shape is ruled, the pool is already
  category-aware and no data migration is implied.
- **The partial unique index is per-domain, not per-category**
  (`idx_provocations_one_per_day ON provocations (domain, publish_on)`). Under
  *any* multi-category design this must become `(domain, category, publish_on)`,
  or two categories cannot both publish on the same day. That is a real
  consequence and it is the same one-line change under both options.

---

## 5. Recommendation

**Option (a): one feed file per category.** `provocations.json` stays exactly as
it is and becomes the `general` category's file for ever; a new category is
`provocations-<category>.json`. `FetchProvocation` gains a category parameter,
which also joins its in-memory cache key (`round.go:29`, `provocStore` is keyed
by domain alone with a 5-minute TTL — a category dimension is mandatory there or
categories would serve each other's provocations for up to five minutes).

Three reasons, in order of weight:

1. **It fails loudly, and (b) fails silently.** A missing category file is a
   non-200, which `round.go:59-61` already turns into an error and a 503 — a path
   the front end already handles as its documented degraded mode
   (`bugs_closed/083`: 503 specifically, because Cloudflare replaces a 502's
   body). Option (b)'s failure mode is a model arguing against a JSON blob, with
   no error anywhere. Given §1.2, **choosing (b) means choosing the option the
   system cannot detect.**
2. **It needs no coordinated deploy.** Every existing caller keeps working
   unchanged because the file it reads is unchanged, so the stages are
   independently valuable (bar 3) and the rollback needs no migration (bar 4):
   the old binary tolerates the new files by ignoring them.
3. **The client change is common to both options, so it is not a
   differentiator.** Either way the Gauntlet page must tell the engine which
   category a round is for — a map does not remove that, it only moves the choice
   from a URL to a key. So the comparison really is just the failure mode.

**The honest weakness of (a):** N categories mean N files and N publisher runs,
and a visitor loading a multi-category page fetches N files instead of one. At
the scale in question (single digits) that is not a real cost, but it becomes one
if categories ever reach dozens, and the map would then be the better shape.
Recorded here so a future reader can see the trade rather than rediscover it.

**On 2.3, recommend yes** — add `category` to `gauntlet_rounds` at the same time,
defaulted to `'general'`. Rounds are published to durable public URLs, and a
backfill of already-published rounds cannot recover a category that was never
recorded. [INFERRED — I have not read the publish path's schema in detail; the
`gauntlet_dead_cta` lane should confirm the column is cheap to add there.]

**On 2.4, recommend deferring**, and *not* bundling it with 2.1. It is the right
long-term answer and RFC 007 already argues the general case, but a shared
package between the chassis and `tools-api` is a second architecture change, and
tying the two makes the cheap decision wait on the expensive one.

---

## 6. What this RFC does NOT ask

Per owner ruling 2026-07-31 (§7.1 of the PLAN) the no-human-approval decision is
**settled and not re-opened here**. Categories change *which* provocation a round
argues, not *whether* a human approved it. Per-category gate configuration —
PLAN §9.2's first argument for the column, that "pets" and "current political
opinions" cannot share a safety threshold — is a matter for the gate, which is
being built in parallel by another session and is out of scope for this file.

---

## 7. OWNER RULING

*Not yet given — filed 2026-08-05.*

---

## 8. CONTRIB 2026-08-05 (gate/generator session) — the index you plan to change now has a SECOND consumer, and it shipped the same hour as this RFC

Appended, not edited — §§1–7 are the filing session's and are untouched. §6 already
anticipates this session ("the gate ... being built in parallel by another
session"); this is the concrete follow-up, because that work landed while this file
was being written and it lands **on your §5 index change**.

**What now exists** (approved by the council round 1, corr `28056723`, commits
`9e5e1f909` / `e3ac4e15d` / `b042fae66`): `gate_provocation`,
`generate_provocations` and — the one that matters here — **`schedule_provocations`,
which assigns `publish_on`**. Nothing is wired; there is no `agent_definitions` row.

**The collision, stated precisely.** Your §5 correctly says
`idx_provocations_one_per_day (domain, publish_on)` must become
`(domain, category, publish_on)` under any multi-category design. That index is now
also the concurrency guard for a writer:

- `nextPublishDates(latestInPool, today, n)` in
  `platform/orchestration/actions/provocation_generator_action.go` computes dates
  **one per calendar day per DOMAIN**, forward only, starting after
  `max(publish_on)` for the domain.
- `ScheduleProvocationsAction` reads that `max(publish_on)` **per domain** and
  relies on the partial unique index to make a concurrent double-booking fail
  rather than silently overwrite.

So when the index becomes per-category, **both halves have to become
category-aware in the same change**, or the scheduler will hand two categories the
same date and read the resulting constraint violation as "another session got
there first" — its existing, deliberate skip path. The failure would be a category
silently never getting scheduled, which is the same shape as the silent contract
failure your title is about.

**This is not an objection to your RFC.** Your recommendation is unaffected and I
am not asking you to widen scope: the scheduler is unwired, so there is no
migration ordering problem today, and whoever implements the index change simply
needs to know `nextPublishDates` is a caller. I have recorded the reciprocal note
in `NOTES_provocation_pipeline.md` and in the register (VONC-012) so it cannot be
found only from this end.

**One thing from my side that may be useful to §4.** The `category` column is
populated today only by its default (`'general'`) — the generator does not set it,
deliberately, because a per-category *gate threshold* (your §6's out-of-scope item)
should decide the vocabulary before a writer starts minting values into it. If your
RFC is ruled on before that, the generator is a one-line change and should follow
your decision rather than pre-empt it.

**Also measured, in case it bears on §4's currency argument:** vonc.com has **0
`content_sources` and 0 `content_feed_items`**, so the feed-ingester the PLAN names
as the currency source has never run for this site. Categories that depend on
topicality would need that configured first; the gate/generator treats currency as
optional for exactly this reason.
