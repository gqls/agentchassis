# HANDOFF 2026-08-18b — the retired £1,200 offer is out of the machinery, the briefing questionnaire serves any site type, and the owner has RULED on build quality vs delivery speed — SUPERSEDES HANDOFF_2026-08-18

**Read order, cold:** this file → `NOTES_webdesign_uk_build_service.md` (the four
2026-08-18 afternoon entries) → `README_where_we_are.md` (last entry, owner-facing)
→ `../site_delivery_and_editor/HANDOFF_2026-08-18_continue_here.md` (the JOINT
cold-start; another session drives both lanes) → `bugs_open/299`.

---

## ⚠ 1. THE ONE THING THAT WILL BITE: the chat prompt-maker is NOT in `make release`

The owner said, 2026-08-18: *"commit the chat prompt-maker and it can be included in
my next release."* **It is committed (`5777ac945`) and the release will NOT carry it.**

`make release` is `build-backend push-backend deploy-core deploy-agents
deploy-agent-cleanup release-dashboard` (makefile:2499). **`sitechat` /
`chat-service` appear NOWHERE in the makefile** — checked, zero matches. The chat box
is a separate Go binary on a Mythic Beasts VM, rolled by scp + systemctl:

```bash
cd docs/agent_docs/docs024_key_docs_latest/webdesign_uk_build_service/box/chat-service
GOPROXY=off GOTOOLCHAIN=local go test . -count=1 && \
GOPROXY=off GOTOOLCHAIN=local GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o sitechat .
scp -i ~/.ssh/webdesign_box_ed25519 sitechat root@webdesign.vs.mythic-beasts.com:/root/
ssh -i ~/.ssh/webdesign_box_ed25519 root@webdesign.vs.mythic-beasts.com \
  'install -m 755 /root/sitechat /usr/local/bin/sitechat && systemctl restart webdesign-chat.service "sitechat@*.service" 2>/dev/null; \
   sleep 2; journalctl -u webdesign-chat -n 3 --no-pager -o short-iso; ss -ltnp | grep sitechat'
```
Expected: `facts: fetched N facts from relay` → `facts: live mode, site=<domain>` →
`sitechat on 127.0.0.1:<port>`. **`ss` must show `127.0.0.1:` binds only.**
Full recipe: `RUNBOOK_…md`, "Rolling the shared binary".

**It was NOT deployed by the 08-18 session** — it is a live customer-facing bot and
the owner had not been told the release would miss it. **Tell him, then roll it.**

---

## 2. OWNER RULING 2026-08-18 — better product beats faster promise (NOT YET ACTIONED)

Asked whether to drop `pageflow-builder` for the submit-domain route or improve
pageflow, the owner ruled:

> *"If we can improve pageflow builder to include all the checks and balances that
> the other flow has then that's great but I don't think it can be done. I'd rather
> change the estimated delivery time if it means a better product for the customer."*

**So the triage-based build cycle wins, and the delivery estimate gives way.** What
this means concretely, and none of it is done:

1. **`build_duration` must change.** It currently attests `value: 1`, claim *"From
   having what is needed from the customer, the site is usually ready the next day."*
   That figure exists to support the one-pass pageflow build. Under a triage-driven
   build it is not honourable.
   **Do NOT invent the replacement number.** It is a live customer promise: the chat
   bot renders the claim verbatim and the pages publish `writer_line` "usually ready
   the next day". Measure the real triage-flow duration first (site row → all pages
   `deployed_at`, over recent builds), propose a figure, and get the owner to attest
   it. Value AND claim AND writer_line all move together, plus `context_terms`.
2. **The claim is load-bearing in more than one place.** Grep before changing:
   `build_duration` is cited in `evidence_base`, and "next day"/"ready the next day"
   appears in served page copy. A register change alone leaves the pages stale until
   they rebuild.
3. **Do not re-plumb the builder yet.** The owner ruled on the TRADE-OFF, not on the
   mechanism. `pageflow-builder` is still `recommended_builder` on **20 of 21** sites
   and owns the fleet's ONLY `briefing_questionnaire`. Swapping the route is a
   programme, not a patch, and nobody has scoped it.

---

## 3. What LANDED today (all committed, all verified live)

| what | state | proof |
|---|---|---|
| Refund ban narrowed to promise shapes | **LIVE** 12:02:13Z | `SQL_2026-08-18d`. Verified at the LIVE pattern: both sentences that actually blocked a rebuild now pass, retired £1,200 promise still blocked |
| £1,200 + retired terms swept from **9 specs** | **LIVE** | `SQL_2026-08-18e` + `f`. Seven phrases asserted nowhere; guard proven able to fail against unswept data |
| Briefing questionnaire → any site type | **LIVE** | `SQL_2026-08-18g`. 11 → 15 fields, backup taken, guard proven able to fail |
| Chat prompt-maker | **committed, NOT deployed** | `5777ac945`; two mutation-checked tests |
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
5. **HITL as a briefing step** — owner accepted the ordering: questions first (DONE),
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
