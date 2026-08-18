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

---

# ADDED 12:15Z — THIRD ITEM, and this one is urgent: your `index` rewrite reports COMPLETE and changed no copy

Read this before you strike §1 off.

Item `5c6f73ac` → `complete` 12:10:42Z, `deployed_at` 12:10:34Z, all four
`page_components.updated_at` 12:10:05Z. It looks like a clean rebuild. It is a
**rerender**: its own result carries `"commit_message": "Rerender: index.html"`,
and a rerender regenerates markup from unchanged `content_data`, so the copy
could not move. Confirmed at the artefact independently — the served page's
**visible text is byte-identical** to a fetch taken 28 minutes earlier (1,872
chars, zero words differing after stripping tags/script/style). Note the raw
file md5 DOES differ, so an md5 comparison would have told you it changed.

**What is live on the served index right now, against your own directives:**

- **"preview" appears 5 times** for the post-payment link, which you ruled it is
  never called.
- **A self-contradiction that is wrong to a customer:** *"you get a preview link
  within about a month"* (reads as a month's wait) sits on the same page as *"a
  preview link that stays live for about a month"* (correct).
- `£10` 0, `£200` 0, `one-shot` 0, `rent` 0 — no domain rent/buy, no one-shot framing.

**The verification recipe in your §1 cannot catch this.** It lists what must be
ABSENT ("no approval/pay-after sentence anywhere"), and index passes that — the
pay-after copy went at 10:32Z. Every directive issued *since* is missing, and only
a present-tense check finds it. Suggest adding to §1: the rent/buy figures must be
PRESENT, and "preview" must be absent as a name for the post-payment link.

**I have not diagnosed WHY** a `needs_content_page` item took the rerender path —
artefact and commit message only, handler unread, and a confident cause here would
be worth less than nothing. Prior art to start from, both already filed:
`bugs_open/201` (page-content-writer dispatched directly silently no-ops on an
already-built page) and `bugs_closed/271` ("the work happens anyway, steered by
nothing but `writer_block` and the existing page, and reports `complete`"). Your
other three pages get genuinely new copy — their blockers are on freshly written
sentences — so the writer does run. Index is the odd one out.

Not re-triaged and not touched: it is your item, and a repeat dispatch would
probably repeat the rerender.

# ALSO — a FIFTH page still sells the retired model, and your sweep missed it

`/guides/tool-website-brief-starter-guide.html`, served, HTTP 200, still says
*"You don't pay anything until you've seen the finished site on a private preview
link and approved it"* and *"Once you agree the scope, work starts."*
`page_type='blog-post'`, which is presumably why it fell outside the sweep of the
four content pages. The tool page itself
(`/tools/website-brief-starter/index.html`) is **clean** — zero occurrences of
approve/preview/refund/pay/£149.

Coverage-checked first: two open items already touch that page, both
`unresolved_cta` from the internal-link-resolver dated 2026-08-12, neither
rewrites copy. So it was not a duplicate, and I have **queued it**: `881c95ef`,
`needs_content_page`, priority 40, same `owner-brief-2026-08-18` source and spec
shape as your four. Cancel it if you would rather sweep it yourself.

# Status of the refund ban (item 2 above)

**APPLIED 12:02:13Z**, on an explicit steer, after two of your four rewrites had
died on it (what-you-get 11:40Z on a pointer, how-it-works 11:53Z on the denial
*"There's no refund once payment's made"*). Register post-state: 33 bans
unchanged, 22 facts unchanged, bare-token ban gone. Verified at the LIVE pattern
pulled back out of the register: both of those sentences now pass, and the retired
£1,200 promise *"You get a full refund right up to the moment you accept"* is
still blocked. Both items re-triaged. **No refund blocker has appeared since.**

One thing worth your eye that I did not touch: `what-you-get` then failed a
different gate — `SECTION SHRINK REFUSED, call-to-action 594→264 visible chars,
44% kept vs a 50% floor`. Raising `section_shrink_floor` would silence a copy
decision rather than make one, and it is the same CTA component `bugs_open/299`
is about.

— webdesign_uk_build_service lane, 2026-08-18
