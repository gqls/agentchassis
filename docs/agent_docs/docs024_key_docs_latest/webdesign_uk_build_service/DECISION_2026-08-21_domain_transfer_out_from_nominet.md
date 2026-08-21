# Transferring a sold domain out to the customer — the mechanism, what it costs, and the one decision still owed

**Status:** mechanism VERIFIED at primary sources 2026-08-21. Owner has ruled that a
manual step per domain is acceptable for now. **One decision remains open and it is
his:** whose name the domain is registered in while the customer is renting it (§4).

**Raised by:** the owner, 2026-08-21 — *"We need to agree a transfer out from nominet.
Nominet's transfer rules are changing to be more like other tld registrars so we'll need
to keep abreast of it. It is likely to be a manual step for now for each domain and that
is ok for now."*

**Supersedes the `[UNVERIFIED]` note** left by `SQL_2026-08-21` and NOTES 2026-08-21,
which said nobody in this lane had checked who can execute the transfer. Now checked.

---

## 1. The words first, because three of them are jargon and all three matter

- **Registrar.** The company that holds a domain in its account at the registry. For
  `.uk` the registry is **Nominet**, and we are one of its registrars: we have a member
  account, a tag, and machine access to it.
- **IPS TAG.** The label identifying which registrar holds a `.uk` domain. Moving a
  `.uk` domain to a different registrar means changing this label. There is no
  customer-held password for it, which is what makes `.uk` different from `.com`.
- **Registrant.** The legal owner of the domain, recorded separately from the registrar.
  **Changing the registrar and changing the owner are two different operations**, with
  different processes and different fees. Conflating them is the trap in this whole
  subject, and it is the reason §4 exists.

## 2. How it works TODAY, and it is two operations, not one

Selling a domain outright for £200 means both of these, in this order:

1. **Registrant Transfer** — changing the recorded owner from us to the customer. This
   can only be done at Nominet itself (Online Services → "Registrant Transfer"), not
   through our own systems. `[VERIFIED 2026-08-21]` Nominet's published fee schedule:
   **£10+VAT** for a straightforward name change, **£20+VAT** for a change of
   type/company, **£35+VAT** where extra verification is needed. Corrections (spelling,
   marriage, same company number) are free.
2. **IPS TAG change** — releasing the domain to whichever registrar the customer has
   chosen. Free for us to perform as the holding registrar.

Once step 1 has happened, the customer is the registrant and **can do step 2 themselves**
through their own Nominet account for ~£10+VAT, without us. So the "they arrange it with
their new registrar" half of the attested `domain_buy_once` fact is mechanically
achievable — but only after we have done step 1, which is unavoidably ours.

**Renting at £10/month involves neither operation.** The domain simply stays where it is.

## 3. What changes on 9 FEBRUARY 2027, and it is good news for this product

`[VERIFIED 2026-08-21 at registrars.nominet.uk]` Nominet is retiring the IPS TAG transfer
process and replacing it with the mechanism every other TLD already uses: a **Transfer
Authorisation Code**. Formal notice was sent 4 June 2026; the transition is
**9 February 2027**. Portfolios migrate automatically to Nominet's "Dragon Domain
Manager", and Nominet moves to a standard EPP implementation at the same time.

The new flow, in Nominet's own words: *"Losing Registrar: Generates and provides the
Transfer Authorisation Code to the registrant."* The customer gives that code to their
new registrar, and **if the domain is unlocked and the code is correct the transfer
completes immediately.**

**Why this matters commercially rather than just operationally.** A code is a thing you
can hand over in advance. From February 2027 the transfer-out step can be *pre-issued at
handover* and put straight into the delivery email alongside the ZIP — at which point
*"nobody's time is included"* becomes true by construction rather than true because we
worded it carefully. That is the same shape as the Phase 4 token design for the ZIP link:
stop promising to do something later, hand over the thing that makes it unnecessary.

**What it does NOT change: step 1.** The Transfer Authorisation Code replaces the TAG
change, not the Registrant Transfer. Ownership changes remain a separate Nominet
operation with a separate fee.

## 4. THE DECISION STILL OWED — whose name is on the domain during the rental?

Everything above is mechanism. This is a business choice, it is the owner's, and it
changes what we do on every single sale.

| | **(a) Registered in OUR name** | **(b) Registered in the CUSTOMER'S name** |
|---|---|---|
| Renting (£10/mo) | Safe. Stop paying, we keep it. | **Exposed.** They are the legal owner from day one and can move it away without ever paying the £200. |
| Buying (£200) | Two operations: Registrant Transfer (£10–35+VAT) then release. Both manual, both ours. | One operation: release the tag. From Feb 2027, a code we can pre-issue. |
| Our per-sale cost | A Nominet fee plus a manual step | Near zero |
| Fits "nobody's time is included" | Requires one mechanical step of ours per sale | Yes, almost entirely |

**Recommendation: (a), register in our name.** The rental exposure in (b) is a real
commercial hole and the £200 is the whole point of the option; the cost of (a) is one
manual step and £10–35+VAT against a £200 sale, which the owner has already said is
acceptable. **Not applied — this is recorded as a recommendation, not a decision, and
nothing is encoded at the register until the owner rules.**

## 5. What the copy says now, and the one line that is slightly ahead of reality

The attested `domain_buy_once` reads: *"…it must be moved to their own registrar: we do
not stay their registrar. We give them what they need to move it; arranging the transfer
with their new registrar is theirs to do, and no support time is included in the price."*

- *"we do not stay their registrar"* — correct under both models.
- *"arranging the transfer with their new registrar is theirs to do"* — correct.
- *"We give them what they need to move it"* — **true from 9 February 2027** (we hand
  them a code) and **not literally true today** (there is nothing to hand over; we
  perform a tag change on request). Not damaging and not false in substance, but a
  customer today could reasonably ask "where is my code?" and there would not be one.

**No copy has been changed for this.** It is one line, it depends on §4, and it becomes
correct on its own in February 2027.

**And note what is NOT a contradiction, because the lane has been round this once
already.** `no_presales_service` says nobody's time is included. Performing the
paperwork to hand over a thing somebody has bought is **delivery, not support** — the
same category as building the site. The 2026-08-19 collision was about hand-holding a
customer through *their* registrar's process, and that stays excluded.

## 6. Keeping abreast, which the owner asked for specifically

There is **no scheduled-review mechanism in this estate** — no diary, no tickler, no
review-due field anywhere in the concept register (checked 2026-08-21). Rather than
invent one for a single date, the checkpoints are written where the cold-start read order
actually goes: this file, the RUNBOOK procedure, and the handoff's open list.

| when | what to check | where |
|---|---|---|
| Before the FIRST domain sale, whenever that is | That §2 still describes reality, and that our tag and EPP access still work | `registrars.nominet.uk` fee schedule + a live EPP login |
| **By 2026-12-01** | Nominet's transition detail: whether the TAC flow, locks or timings have moved since the 4 June 2026 notice | https://registrars.nominet.uk/registry/dot-uk/faq/ |
| **9 February 2027** | Transition day. Re-read §3, rewrite §2, and revisit the `domain_buy_once` line in §5 — it becomes literally true that day | same |

**The failure mode this table is written against:** the transition is far enough away
that every session between now and then will correctly decide it is not their problem,
and the lane will arrive at February 2027 with copy describing a mechanism that no longer
exists. A date with no owner is not a plan, so it is in the handoff's STILL OPEN list too.

## 7. What we already have, so nobody re-derives it

- **Nominet member account with EPP access.** `epp.nominet.org.uk:700`, TLS, RFC 5734
  framing. Password at `~/.config/nominet/epp-password`.
- **A working stdlib-only EPP client in this repo**:
  `idea_uk_vm_site/box/nominet-epp-ns-change.py` (dry-run by default; changes
  nameservers, not tags, but the connection/login/framing half is exactly what a tag
  release would reuse).
- **Traps already paid for**, from `domains_cloudflare_rollout/RUNBOOK`: pin to IPv4
  (IPv6 gets a 94-byte brush-off where IPv4 gets the full greeting); the egress IP must
  be allowlisted in Online Services; **the greeting is served to ANY IP, so only a
  successful login tests the allowlist** — a greeting is not a connectivity proof.

**Automation is explicitly NOT being built.** The owner ruled manual-per-domain is fine
for now, and at current volume (zero sales) an automated release path would be a
mechanism rotting unexercised — a cost this platform has been bitten by before. The EPP
client is noted so that *when* volume justifies it, nobody starts from scratch.

---

**Sources** (fetched 2026-08-21):
- Nominet registrar FAQ, transition and TAC flow — https://registrars.nominet.uk/registry/dot-uk/faq/
- Nominet UK fee schedule — https://registrars.nominet.uk/uk-namespace/managing-account/payments/fee-schedule/
- Nominet tag types (self-managed tags) — https://registrars.nominet.uk/uk-namespace/registrar-agreement/selecting-tag-type/
