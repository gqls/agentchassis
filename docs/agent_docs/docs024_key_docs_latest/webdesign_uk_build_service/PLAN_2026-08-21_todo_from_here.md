# The plan from here — webdesign.uk build service + site delivery

**Written 2026-08-21** at the owner's request ("show the todo list plan from here").
Supersedes nothing; it is the ordered read of the two lanes' open work as it stands
today. Cold-start context: `HANDOFF_2026-08-20_continue_here.md`.

**The one-line state:** the commercial position is settled, attested and live at the
served pages and the chat bot. **Nothing can be sold yet**, because there is no way to
take money and no way to deliver what was bought. Everything below is in service of
those two sentences.

---

## The critical path to first revenue, in order

Each of these blocks the next. Nothing else on this page matters until they are done.

| # | What | Who | Blocked on |
|---|---|---|---|
| 1 | **Stripe keys** into terraform; the webhook edge exception; the webhook hostname | **OWNER** | nothing — this is the top of the tree |
| 2 | **Phase 4 HTTP surface** — `/d/<token>` mints a clamped presign and redirects; `/c/<token>` records the transfer confirmation | build | nothing; schema is live (migration 511, DGH-014) |
| 3 | **The delivery email** through `platform/mailer` — ZIP link, live-site link, Netlify invite, both domain links, Stripe portal | build | 2 |
| 4 | **The weekly chase email** — two subjects: site off our hosting, bought domain off our registrar account | build | 3 |
| 5 | **The retraction job** — gives `live_link_expires_at` teeth; today nothing expires and serving is unbounded | build | 2 |

**Note on 2:** this is also what delivers the owner's "longest link we have" ruling. It
cannot be delivered as a number — a presigned URL is capped at 7 days by the signing
protocol and the code already sits exactly on that ceiling. A token of ours, redeemed for
a fresh short presign, makes the window ours. Same mechanism as `/c/`, so build them
together.

## Decisions owed by the owner (each unblocks work, none is urgent except #1 above)

| | Decision | Why it matters | Recommendation |
|---|---|---|---|
| D1 | **Whose name is on a domain during the rental** | Changes every sale: ours = renting is safe but a sale is two Nominet operations; theirs = a sale is nearly free but a renter can leave with the domain unpaid | **Ours.** Decision doc §4 |
| D2 | **Second Nominet TAG** (domain programme only) | Separate programme, not this lane's critical path | — |
| D3 | **Contact email** `webdesign@contactforsales.com` — domain mismatch (item `a8d6f440`) | Small, visible on every page | pick an address on a domain we own |

## Domain transfer-out — settled, with dated checkpoints

Mechanism verified 2026-08-21:
`DECISION_2026-08-21_domain_transfer_out_from_nominet.md`. Manual per domain, by owner
ruling. **Not to be automated.**

- [ ] **Before the first sale** — prove Nominet EPP access still works. A greeting is not
      a login; only a completed login tests the IP allowlist. RUNBOOK has the commands.
- [ ] **By 2026-12-01** — re-read Nominet's transition detail at
      `registrars.nominet.uk/registry/dot-uk/faq/`. Has the TAC flow, locking or timing
      moved since the 4 June 2026 notice?
- [ ] **2027-02-09** — transition day. Rewrite RUNBOOK step 2, revisit
      `domain_buy_once`'s "we give them what they need to move it" (it becomes literally
      true), and **fold a pre-issued transfer code into the delivery email** — that is
      the product win, not the process one.
- [ ] **After the first real transfer** — correct the RUNBOOK from what actually
      happened. It is written UNTESTED on purpose.

## Copy and register — small, all unblocked

- [ ] **`bugs_open/299`** — the home-page CTA names the Website Brief Starter and its
      href **dials the phone**. Filed, deliberately unpatched: the section was written
      after the 268 fleet fix, so something still generates it. The producer question is
      the point, not the patch.
- [ ] **`what-you-get` shrink gate** — `call-to-action 594→264 visible chars (44% kept,
      floor 50%)`. Same CTA as 299. Raising the floor would silence a copy decision
      rather than make one.
- [ ] **The prompt-maker pointer is DUE.** The chat conduct deliberately does not name
      the Website Brief Starter because that tool's guide page was selling the retired
      model. It no longer is (guide landed 2026-08-19). Read the live conduct before
      editing.
- [ ] **`writer_block` breaks its own first rule** — "never use an em dash", and it
      contains six (paragraphs 12, 14, 20, 21). Measured 6 before and 6 after the 08-21
      edits, so none were introduced then. Five-minute fix; the rule itself prescribes
      the replacements (full stop, comma, colon, brackets).

## Parked — do not develop

- **Reseller market supposition.** Owner-parked. Discussion only; written to no fact,
  writer_block, mission or spec. Do not encode it, do not reopen it unasked.
- **HITL as a briefing step.** Ordering accepted (questions first — done; then HITL via
  the **work-item** queue, which has a working screen). The orchestration HITL path has
  never fired: `collect_via_hitl` 0, `brief_answers` 0, `hitl_mode` 0 across 369 briefing
  orchestrations, against `briefing_answers` = 3 as the control. Not next.

## Standing rules that constrain all of the above

- **Every site goes through the framework.** Never hand-build a page (owner, 2026-08-04).
- **A better product beats a faster promise** (owner, 2026-08-18). Not a licence to
  re-plumb the builder.
- **No example sites** until this route has produced some (owner, 2026-08-18).
- **The register is the wire.** Never steer via item-spec prose, never hand-edit HTML.
  And when you edit a fact, check `writer_block` still agrees with it — asserting
  `writer_block` unchanged is exactly how a retired value keeps steering the writer
  (LANDMINES, 2026-08-21).
