# NOTE 2026-08-18 (~12:00Z) — two measured corrections to the JOINT handoff §3, from the webdesign_uk_build_service lane

Left as a file because your session is not reachable from this machine
(checked: 37 peers, none is this lane; `SendMessage` has no target). Nothing
here has been APPLIED — you are driving both lanes and had four rewrites in
flight against the register at 11:50Z, so I did not edit it under you.

Both items are corrections to statements in your §3 "Standing hazards". Both
are measured, and both make something you have deferred cheaper than you think.

---

## 1. "example links need an allow-list mechanism first" — THE MECHANISM ALREADY EXISTS AND IS LIVE

Your §3 says the cross_site_domain guard "blocks any other-site domain in
webdesign.uk copy — example links need an allow-list mechanism first, if the
owner ever wants them".

There is nothing to build. The opt-in allowlist has been live since
**v1.0.1146 (2026-07-21)** and a sibling site uses it in production **today**:

```sql
SELECT domain, content_data->'allowed_reference_domains' FROM sites
 WHERE content_data ? 'allowed_reference_domains';
-- fundamentallyai.com | ["leopardessconsulting.co.uk","finetuning.uk","idea.uk","relojistas.com"]
```

- Read by `loadAllowedReferenceDomains` (`validate_page_content.go:1462`) from
  `sites.content_data->'allowed_reference_domains'`, a JSON string array.
- `checkDomainContamination` skips **both** the domain and the company check for
  an allowlisted site — deliberately, so a rewrite naming the company rather
  than the bare domain cannot re-block the same approved reference.
- It is per-site and fully opt-in: key absent → nil map → today's behaviour,
  unchanged. Built for exactly this case (`bugs_open/055`): *"a portfolio /
  meta site may name another of our sites on purpose — as a case study"*.
- **Live config. No code change, no rebuild, no council round.**

**And the guard refused ONE of four, not four of four.** The hardcoded
`knownSites` list is five domains: finetuning.uk, gaswholesalers.com,
ai-agent-orchestration.com, leopardessconsulting.co.uk, dartsonline.com. Of the
owner's four attested example domains, only **dartsonline.com** is in it —
noted.co.uk, cookly.uk and vetcomparison.uk are not subject to this check at
all. The persisted detail for the faq run (11:47:41Z) lists exactly one
cross-site issue, and it is dartsonline.com.

So if the owner un-defers example links, the whole cost is one row:

```sql
UPDATE sites SET content_data = jsonb_set(content_data, '{allowed_reference_domains}',
       '["dartsonline.com"]'::jsonb) WHERE domain='webdesign.uk';
```

**This does not reopen the owner's decision** — he deferred examples on the
substantive ground that none of those sites was built by this one-shot route,
which is a copy-honesty judgement the allowlist has no bearing on. It only
means the blocker is not a reason, and "we need to build a mechanism" should
not be weighing on that decision.

---

## 2. "known coin-flip failures the gate correctly catches (just re-triage)" — MEASURED: not a coin flip, and not correct

Your §1 files the refund-mention failures as coin-flips. Measured against the
live pattern through the real scanner
(`datahelpers.ScanBannedClaims`, pattern `\brefunds?\b|\brefundable\b|\bmoney.back\b`),
over twelve natural ways to state the owner's own no-refunds position:

**8 of 12 are BLOCKED.** Only *"we do not offer refunds"* and *"we don't offer
refunds"* survive. Blocked include:

| phrasing | verdict |
|---|---|
| "Refunds are not available." | BLOCKED |
| "Refunds are not offered once work has started." | BLOCKED |
| "There are no refunds." | BLOCKED |
| "No refunds." | BLOCKED |
| "The price is non-refundable." | BLOCKED |
| "Do you offer refunds? No." | BLOCKED |

**The mechanism** (`claims.go`, `NegationGuard.NegatedAt`): the guard scans
**backwards** from the matched token, within the clause, for a negation cue. A
cue that FOLLOWS the token can never suppress — so *"refunds are **not**
available"* is read as a refund promise. Bare "no" and "non-" are excluded from
the cue vocabulary on purpose, documented and pinned by
`TestBareNoIsAKnownResidualOfTheSharedGuard`.

So the writer has two survivable phrasings out of twelve and no way to know
which. That is a systematic bias, not variance, and it costs a failed rebuild
each time it lands wrong. It also explains the flag the morning left open: the
served index carried *"We do not offer refunds"* at 10:22Z and carries **zero**
occurrences of "refund" now (cache-busted, 11:42Z). Steering the writer to
point at "the full terms" works around it; it does not stop the disclosure
being squeezed out of every page that tries to state it plainly.

**A tested replacement is ready and HELD, not applied:**
`webdesign_uk_build_service/SQL_2026-08-18d_refund_ban_promise_shapes_HELD.sql`
— bans the promise SHAPES instead of the bare word, in the lane's usual guarded
form (exact-match replace of the one pattern, ban count asserted unchanged,
facts asserted not lost, no fixed fact count).

Verified in both directions, on the exact string as written in that file:

- **24 hand cases: 0 failures** — every denial above allowed; all twelve promise
  shapes still blocked (money-back, full refund, "a refund is available",
  refundable deposit, "we will refund you", "request a refund", …).
- **26 real corpus lines** — every refund-bearing component in the fleet, 7
  sites, none written for this test: **0 newly blocked**. The five retired
  £1,200-model promises on this site ("walk away and get a full refund", "Full
  refund until you accept") stay blocked. The 11 freed lines are all other
  sites' consumer-rights prose ("a refund of the interest and charges" in the
  Ombudsman guides) — never a promise by the site itself.

**Apply it when your four rewrites are settled**, and re-read the register
first — you edit that row in place and so would this.

**I deliberately did NOT touch the shared negation guard.** Widening it to take
bare "no" would change the claims gate for every site in the estate to make one
site's copy easier, and the exclusion is a deliberate fleet-wide decision with a
test pinning it. The fix belongs in this site's own pattern.

---

## What is NOT decided here, and is the owner's

Whether the home page must carry the no-refunds disclosure at all is a
commercial / consumer-rights call, and it is still open. Item 2 only makes the
sentence **sayable** in normal English; it does not decide that it should be
said, and it writes no copy.

— webdesign_uk_build_service lane, 2026-08-18
