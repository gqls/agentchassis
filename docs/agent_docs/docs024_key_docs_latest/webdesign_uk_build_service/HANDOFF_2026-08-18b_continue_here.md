# HANDOFF 2026-08-18b — the retired £1,200 offer is out of the machinery, the briefing questionnaire serves any site type, and the owner has RULED on build quality vs delivery speed — SUPERSEDES HANDOFF_2026-08-18


> **⚠ UPDATED 2026-08-18 ~16:10Z. Two of this file's open items are CLOSED and one
> has hardened into a decision the owner must make.**
>
> - **§1 is DONE.** The chat prompt-maker is **LIVE** on the box (commit
>   `434d2b64b`, verified at the running service, smoke-tested), and the Mythic
>   Beasts deploy is now **makefile targets** at the owner's request:
>   `make box-release` / `box-status` / `box-verify` (+ `box-build`,
>   `box-build-tree`, `box-push`, `box-deploy`, `box-test`). Deliberately NOT under
>   `release` — see §1 below for why, it is unchanged reasoning.
> - **§2 now has EVIDENCE.** "Usually ready the next day" is **refuted by
>   measurement**. What is owed is the replacement figure, and it is the owner's to
>   attest. See §2.
> - **A new trap, and it invalidates a line in the RUNBOOK:** md5 cannot tell you
>   what SOURCE the box binary came from. The binary now stamps its own commit.

**Read order, cold:** this file → `NOTES_webdesign_uk_build_service.md` (the four
2026-08-18 afternoon entries) → `README_where_we_are.md` (last entry, owner-facing)
→ `../site_delivery_and_editor/HANDOFF_2026-08-18_continue_here.md` (the JOINT
cold-start; another session drives both lanes) → `bugs_open/299`.

---

## 1. ✅ DONE — the chat box deploy, now a makefile target

**The prompt-maker is LIVE**, commit `434d2b64b`, rolled 15:58Z. Verified at the
**running service**, not at git and not at a tag:

```
build provenance: git_commit=434d2b64b26d91c1861d42cd474139318441ecc8
facts: fetched 22 facts from relay   ·   facts: live mode, site=webdesign.uk
sitechat on 127.0.0.1:8081 (max_turns=20, daily_ceiling=$10.00)
```
Smoke-tested functionally: *"I run a small darts league and want a website for
it"* → *"what would you want the site to actually do for people in the league?"*
No business assumption, one question at a time. The conduct works.

**To roll it again: `make box-release`.** Also `box-status` (what is actually
running), `box-verify` (md5 **and** provenance), `box-build REF=<sha>` to pin.

**It is NOT under `release`, on purpose.** `release` is
`build-backend push-backend deploy-core deploy-agents deploy-agent-cleanup
release-dashboard` — a different machine, a different credential (ssh key, not
kubeconfig), a different blast radius. A customer-facing bot must not roll as a
side effect of a fleet deploy, and a fleet deploy must not fail because an ssh key
expired. **What was wrong was that the path was INVISIBLE, not that it was
separate** — `sitechat` appeared nowhere in the makefile, which is exactly why the
prompt-maker was committed in the belief the next release would carry it.

`box-build` builds from **committed HEAD via `git archive`**, like the backend, so
it cannot bundle another session's WIP. `box-build-tree` is the opt-in escape hatch
and stamps `<short>-tree` so a WIP binary can never be mistaken for a release.

### ⚠ NEW TRAP — md5 cannot tell you which SOURCE the box is running

Measured while proving the rollback path: rebuilding the **exact commit** behind
the then-live binary (`84202f061`) produced md5 `65da9971` / 9381552 bytes against
the box's `f07fb146` / 9381544. Same source, different digest. **These builds are
not byte-reproducible across build environments.** So the RUNBOOK's standing line
— "md5sum on the box must equal the local build" — proves only that the box holds
the file you just pushed; it cannot answer "which commit is live", and it looks
like it can. **The binary now stamps its commit** (`main.go` `buildCommit`,
linker-set) and logs the same `build provenance` line the backend uses. Ask the
service, not the digest. The RUNBOOK section still carries the old advice and
should be corrected.

## 2. ⏳ OWNER RULING + EVIDENCE — the delivery promise is refuted; the new figure is the owner's

The owner ruled 2026-08-18:

> *"If we can improve pageflow builder to include all the checks and balances that
> the other flow has then that's great but I don't think it can be done. I'd rather
> change the estimated delivery time if it means a better product for the customer."*

**Measured since, and it backs him up. "Usually ready the next day" is REFUTED.**

First, a metric that lied and was discarded: "first page created → last page
deployed" made older sites look progressively slower (relojistas 795h). That is an
artefact — `pages.deployed_at` is **overwritten by every later rerender**, so it
measures the last time anything deployed, not time-to-build.

Page *creation* is fast and is not the cost: loanzy.uk 20 pages, adversecredit 19
pages, remortgagecalculator 6 pages, all **0.0h** span (one batch).

**The triage tail is the cost and it runs past a day:**

| site | built | items | closed | STILL OPEN | elapsed |
|---|---|---|---|---|---|
| remortgagecalculator.uk | 08-17 11:44 | 48 | 26 | **22** | 25.3h |
| loanzy.uk | 08-18 12:53 | 77 | 62 | **15** | 3.1h |
| adversecreditmortgage.co.uk | 08-18 12:35 | 47 | 5 | 42 | 1.2h (too young) |

Not cosmetic: remortgagecalculator at **25 hours** still has `needs_page` ×4,
`needs_new_component` ×3, `needs_imagery` (one high) and 10 `unresolved_cta`. A
site missing four pages is not deliverable. loanzy at 3.1h has 9 HIGH items open
including `site_unreachable`.

**WHAT IS OWED, and it is the owner's call:** `build_duration` attests `value: 1`,
claim *"From having what is needed from the customer, the site is usually ready the
next day"*, and **the live chat bot renders that claim verbatim to customers**.

**Do NOT pick the number from this data.** n=2, and neither site has reached
"done", so the evidence refutes 1 day without establishing the right figure.
Recommendation: **attest a deliberately safe figure now** (under-promise, tighten
later once "done" is instrumented) rather than leave a refuted promise live while a
better measurement is built. When it changes, `value`, `claim`, `writer_line` and
`context_terms` move together, **and the pages need a rebuild to pick it up**.

**Do not re-plumb the builder.** The owner ruled on the TRADE-OFF, not the
mechanism. `pageflow-builder` is still `recommended_builder` on **20 of 21** sites
and owns the fleet's ONLY `briefing_questionnaire`. Swapping the route is a
programme nobody has scoped.

## 3. What LANDED today (all committed, all verified live)

| what | state | proof |
|---|---|---|
| Refund ban narrowed to promise shapes | **LIVE** 12:02:13Z | `SQL_2026-08-18d`. Verified at the LIVE pattern: both sentences that actually blocked a rebuild now pass, retired £1,200 promise still blocked |
| £1,200 + retired terms swept from **9 specs** | **LIVE** | `SQL_2026-08-18e` + `f`. Seven phrases asserted nowhere; guard proven able to fail against unswept data |
| Briefing questionnaire → any site type | **LIVE** | `SQL_2026-08-18g`. 11 → 15 fields, backup taken, guard proven able to fail |
| Chat prompt-maker | **LIVE 15:58Z** | `434d2b64b`; provenance line at the running service + functional smoke test |
| Box deploy in the makefile | **LIVE** | `make box-release`; 8 targets, build-from-committed-HEAD |
| Chat binary provenance stamp | **LIVE** | `buildCommit` in `main.go`; md5 proven insufficient |
| Brief-starter GUIDE rewrite | **queued** `881c95ef` | still served with pay-after-approval copy |

**Answers to the two flags the morning handoff left open:** `index.rebuild_policy` is
`generic` after the chat placement, so generic rebuilds are NOT refused (checked, it
could have read `owned`). The no-refunds sentence had gone from the served index
because the claims gate blocked **8 of 12** natural phrasings — now fixed.

---

## 4. STILL OPEN

1. **The chat box roll** (§1) and **the delivery-time figure** (§2). Both above.
2. **`index` rewrite reported COMPLETE and changed no copy** — it was a *rerender*
   (`"commit_message": "Rerender: index.html"`; served visible text byte-identical).
   So the post-payment link is still called a **"preview" 5 times** on the served
   index, against the owner's directive, and the page contradicts itself: *"you get a
   preview link within about a month"* (reads as a month's wait) alongside *"a preview
   link that stays live for about a month"*. **Cause NOT diagnosed** — artefact and
   commit message only, handler unread. Start at `bugs_open/201` and
   `bugs_closed/271`. Handed to the joint-driving session in their directory.
3. **`what-you-get` fails a SHRINK gate**, not a claims gate: `SECTION SHRINK REFUSED,
   call-to-action 594→264 visible chars (44% kept, floor 50%)`. Raising
   `section_shrink_floor` would silence a copy decision rather than make one, and it
   is the same CTA `bugs_open/299` is about.
4. **§3 of the old handoff**: `bugs_open/299` (home CTA dials the phone), the contact
   email domain mismatch (`webdesign@contactforsales.com`, item `a8d6f440`), Stripe
   webhook hostname and keys.
5. **The apex `webdesign.uk` 302s to `webdesign.co.uk`** — a different site in the
   estate. The chat API answers on `preview.webdesign.uk` only. A customer typing
   the apex lands on another brand. Found in passing while smoke-testing; whether
   the redirect is intended is an owner question, and nobody owns it.
6. **HITL as a briefing step** — owner accepted the ordering: questions first (DONE),
   then HITL, and route it through the **work-item** queue, which has a working
   screen. The orchestration HITL path has never fired: `collect_via_hitl` 0,
   `brief_answers` 0, `hitl_mode` 0 across 369 briefing orchestrations, while the LLM
   path's `briefing_answers` reads 3 as the control. No consumer found for
   `system.notifications.ui`.

---

## 5. Traps this lane paid for TODAY (all cheap to re-hit)

- **A bare-token `banned_claims` pattern bans the DENIAL too.** `\brefunds?\b` blocked
  8 of 12 ways of stating the owner's own no-refunds position, because the negation
  guard scans **backwards only** and bare "no"/"non-" are excluded cues by design.
  Filed in `LANDMINES.md`.
- **A guard that has only seen the state it was written for proves nothing.** Both
  SQL guards were run against the OLD data first and made to fail. Do this.
- **A hardcoded fact list in a register migration goes stale within hours** — the
  08-18b list would have aborted on a *correct* register because another lane had
  legitimately retired two facts. Compare against the row your transaction
  supersedes instead.
- **Set operators associate left-to-right**: `A EXCEPT B UNION ALL C EXCEPT D` is not
  a symmetric difference. My first version could never have failed.
- **`submission` embeds its own differing copies** of `mission_brief` and
  `roadmap_brief`; **`content_direction` carries a rendered `formatted` duplicate** of
  its structured fields. Fix one copy and the other stays stale and authoritative.
- **Mirroring edits into a rendered duplicate over-matches on short anchors** —
  replacing `"refund"` produced *"Never describe the no refunds or revision right"*.
  Anchors ≥25 chars only.
- **A `complete` work item is not a repaired artefact** — see §4.2.

---

## 6. Falsifiers

- A newer handoff in this dir or in `site_delivery_and_editor/`.
- The four/five queued rewrites' statuses, and whether the served pages actually
  changed — **check the served page, never the item status** (§4.2 is exactly that).
- Whether the chat box was rolled after all (`md5sum /usr/local/bin/sitechat` on the
  box vs a local build).
- Whether `build_duration` has been re-attested by the owner (§2).
- The register's `updated_at`: two lanes write it, and the other edits IN PLACE.
