# Spaceship: 44 new domains since the 09-02 snapshot

Delivered alongside `spaceship_domains_2026-09-03.csv` in response to the owner's
report ("They have someone else's account details in the listing and some of
them have only just been added") relayed by the domain_valuation lane 2026-09-03.
This file covers the Spaceship-registrar side only — it does **not** address the
".co.uk / just added" half of that report, which the valuation lane already
resolved separately via the Nominet registry walk (commit `72ad40c6f`: 14 genuinely
new .co.uk domains, registry-verified as the owner's via the DESIGNCONSULT tag,
some still in the post-registration grace period — that explains "only just been
added" for THAT set and is a closed loop).

## What changed at Spaceship

`[MEASURED 2026-09-03]` domain count went 203 → 247 (+44) between the two
snapshots. SellerHub listings are **byte-identical**, 831 rows both days — no
SellerHub-side change at all, and none of the 44 new domains appear in SellerHub.

All 44 arrivals are `.com` (the whole Spaceship account is 100% `.com`, zero
`.co.uk` — so the owner's 50 `.co.uk` names cannot be at Spaceship regardless of
this finding).

## Why this is worth a second look, and why it is probably NOT the "someone else's
## details" issue

Registration dates run 2019-10-25 to 2026-09-03 (today) — most are aged domains,
not new registrations, and several names read as a different portfolio style
entirely from the rest of the account (which is UK-trade-keyword: e.g.
`electriccarchargingpoints.com`, `floatingpontoons.com`). The new arrivals include
generic multi-language business names: `blockchain-consultant.com`,
`king-transport.com`, `arbeitagentur.com` (DE), `urlaubsverwaltung.com` (DE),
`trykkeriet.com` (NO), `bouwvakkers.com` (NL), `keukenrol.com` (NL),
`lesrouges.com` (FR), `acomedido.com` (ES). Full list + reg dates in
`spaceship_domains_2026-09-03.csv` (diff against the 09-02 file for the set).

**But every technical signal checked is consistent with the rest of the account**,
not with a foreign injection:
- `registrant` contact ID is identical across all 247 domains (old and new) —
  one value, no exceptions. It resolves (via `GET /v1/contacts/{id}`) to the
  account holder's own name/email on file, matching this repo's own git author
  identity. **No contact-ID discrepancy anywhere in the account.**
- `verificationStatus: success` and `privacyProtection.level: high` on all 44,
  same as the rest of the account.
- Nameserver parking on the 44: 37 atom.com, 6 spaceship.net (Spaceship's own
  default), 1 afternic.com — atom.com is already the account's dominant
  secondary-parking pattern (58 of the original 203), so this is not a new or
  unusual NS pattern.

Read together, this looks like a bulk acquisition (buying a batch of aged,
brandable domains from an aftermarket source) landing in the account under the
owner's own identity — not evidence of account compromise or a listing carrying
someone else's details. **Still worth the owner confirming he made this purchase**,
given it is +22% of the account's domain count in one day and the naming style is
a clean break from everything else here.

## The part that could NOT be checked — and is the more important one

The owner's phrase was "someone else's account details **in the listing**" — that
reads as a marketplace/SellerHub listing carrying wrong seller or payout
information, not a WHOIS/registrant mismatch. Checked the full Spaceship API
surface (docs.spaceship.dev, 2026-09-02+03): **the SellerHub domain object
(`GET /v1/sellerhub/domains/{domain}`) exposes no seller identity, payee,
bank/payout, or account-linkage field of any kind** — only `name`, `status`,
`description`, and the BIN/min price pair. There is no API surface to check for
this at all.

**This means the risk the owner described is unverified, not cleared.** If a
SellerHub (or Afternic/Atom, which this estate holds no API keys for) listing
really is carrying another party's payout details, nothing here would have
caught it — it can only be seen in the Spaceship dashboard's own SellerHub UI
(account/seller settings, wherever payout method lives), or the equivalent
screen at Afternic/Atom if that is what he meant. **He should check that
directly and treat "no API finding" as "not checkable," not "all clear."**
