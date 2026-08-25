# CONTRIB — your §9 is now enforced in code, and `bugs_open/320` has a live recurrence

**From:** the `bugs_open/395` session, 2026-08-25.
**For:** whoever picks up `bugs_open/320` (`meta_description_is_never_asked_for`).
**You do not need to do anything.** This is a notification plus one open question that is yours.

---

## 1. Your §9 was broken five days after you wrote it, by a different producer

`bugs_open/320` §9 says, plainly:

> **"Do not file `content_rewrite` items for them.** §5: measured, completes, does nothing."

On **2026-08-24** the offer-analyser filed exactly that item — webdesign.co.uk `index`, about the
meta description, routed at `page-build-handler`. It completed `complete` on 08-24 22:25Z with a
commit sha and a deploy result, and the column did not change. It became the worked case of
`bugs_open/395`.

**This is the same page and the same field as your §5 second row** (`13522562-…`, completed
2026-08-15 19:15Z, column 0 chars). Your table now has a third entry, nine days later.

⚠ **Nobody was careless.** Your §9 is clear, well-evidenced and came out of a diagnosis-loop run. The
395 lane grepped `bugs_open/` before filing — for the MECHANISM (completion gating, acceptance tests,
verifiers) — and correctly found nothing, because your file is about descriptions being *missing*.
Two bugs about the same rows shared no vocabulary. Both lanes have logged that in `WRONG_CALLS.md`;
the transferable rule is **grep for the mechanism AND for the column**.

## 2. It is enforced now — §9 is code

Committed `af3194204` (`Council-Submitted: 021cb965-…`), register **WII-035**. Go, so inert until an
image rolls.

`classifyFinding` in `write_audit_findings_action.go` is now a wrapper: it routes as before, then
asks whether the handler it chose can actually write the field the finding's own acceptance criterion
names. If not, the finding becomes a **`capability_gap`** — `deferred`, empty `handler_agent`,
priority 200, deduped on the **field** — instead of dispatchable work.

**It parks rather than drops, and your bug is the reason.** Dropping the finding would replace a
false green with silence, and silence is precisely what `320` was filed about. `capability_gap` is a
read surface, not a bin: `[MEASURED 2026-08-25]` 50 rows and two live readers
(`diagnose_triage_action.go:361`, `fixloop_digest_action.go:358`).

**The roster carries your census as its evidence**, verbatim enough to check —
`save_page_meta_description_action.go:211` is the only unconditional writer, reachable from one agent
whose `pre_query` selects `COALESCE(meta_description,'')=''`; `site_db_actions.go:1235` and
`apply_adoption_plan_action.go:84` are both `COALESCE(NULLIF(EXCLUDED,''), existing)`.

⚠ **One correction to your §4 for the record:** it says *"No UPDATE path exists. All seven writers of
the column are create-or-upsert."* That was true when written and is now **stale by addition** — your
own fix added one (`save_page_meta_description_action.go:211`, migrations 488/493). The unconditional
UPDATE exists; what does not exist is a *work-item-driven route to it*. Worth a dated correction in
place, since the sentence reads as "nothing can ever overwrite" and a future lane may act on it.

## 3. What is still yours, and it needs the owner

`320` stays **OPEN** and this change does not close it. It stops the estate lying about these
findings; it does not repair a single description.

Making them repairable means a standing path that rewrites the **published** description of a live
page in response to an automated finding. Your §15 records the owner granting `overwrite_existing:
true` for the one-off 681-page regeneration and **explicitly withholding it for the standing
mechanism** — and you verified afterwards that the seeded agent was left unarmed.

So the open question, which neither the 395 lane nor this session will take:

> **Should a finding be able to route at a writer that can overwrite a non-empty
> `pages.meta_description`?** If yes, it is an owner decision about standing authority, not a lane's.
> If no, these findings correctly live on the capability-gap list for ever, and that list is the
> honest record of a thing we can see and have chosen not to fix automatically.

The 395 lane has written the same question into `bugs_open/395` §9c. It would be better answered
once, in your file, since the column is yours.

## 4. Pointers

- `bugs_open/395` §9 (the recurrence + the four writers) and §10 (gate 1c is live on `v1.0.1339`)
- register `WII-035` in `docs/agent_docs/docs026_concept_register/register/work-item-integrity.md`
- `LANDMINES.md` — *"A REJECTED acceptance predicate is a WRAPPER"* (only relevant if you consume predicates)
- `WRONG_CALLS.md`, 2026-08-25 — both lanes' entries on grepping the mechanism and not the field
