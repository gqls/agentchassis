# BUG 481 — a generic "insurance" signal picks the ONLY directory kind built so far, and a farm insurance hub ships a private-medical-insurer directory

**Filed** 2026-09-04 by the `farmerinsurance_uk` lane.
**Renumbered from 479 on the day of filing.** Another lane created its own `bugs_open/479` (the
Layer-2 tool-orphan case) at 12:23:57Z, four minutes after this one at 12:20:07Z. This file moved
even though it was first, because THAT 479 already carries inbound pointers from `bugs_open/385`,
`LANDMINES.md`, the portfolio_positioning and bugfix_450 lanes, and Go commits in the v1.0.1361
roll — every existing reference to "479" means theirs, and this one had none outside its own lane.
**Status** OPEN. Live damage on one site; the mechanism fires on any insurance vertical that is
not health insurance.
**Diagnosis loop** run rather than asserted (CLAUDE.md's 2026-07-31 ruling): intake corr
`bc8d399f-da3b-4408-ad0e-985bf2f7cd7a`, RUN corr `c705263c-9b07-40fb-800f-6ebe7e1ce4a8`.
Verdict to be appended below when it lands.

## 1. Symptom, at the artefact

`farmerinsurance.uk` — a UK **farm** insurance information hub (buildings, livestock, machinery,
crop, liability) — serves `/health-insurers.html`:

- `<title>` = "UK Health Insurer Directory | Farmer Insurance UK"
- `<h1>` = "A plain list of UK health insurers."
- entries: Bupa, AXA Health, VitalityHealth, WPA, Saga Health Insurance, The Exeter, Freedom
  Health Insurance, National Friendly, General & Medical, Aviva, Drewberry

and links it from the homepage under the heading "A directory of UK health insurers"
`[MEASURED 2026-09-04, curl, with a 404 control on an invented path]`.

The site's own `identity` spec gives industry "Insurance Information & Education" and services
that are entirely agricultural. Its `classification` spec's `industry_tags` lead with
`agricultural-insurance`. Nothing in the site's brief asks for private medical insurance.

## 2. Root cause — one map entry, contradicted by its own file

`platform/orchestration/actions/feed_directory_recommendation_action.go`, `verticalDirectoryMap`:

```go
"insurance": {
    Recommended:  true,
    Reason:       "Insurance sites gain authority from a cited, verified directory of UK health insurers (the one insurer kind built so far; more kinds follow)",
    Kind:         "health-insurer",
    SpecKey:      "health_insurer_directory",
    SeparatePage: true,
},
```

Two entries below, the same map states the rule that this violates:

> `"finance"` alone is deliberately NOT recommended: it is too generic to pick a single provider
> class, and **a wrong directory on a site is worse than none.** A site classified merely
> "finance" gets no directory until a sharper vertical signal lands.

`matchVerticalDirectory` (same file) lowercases `industry`, `site_type`, `category`, appends ONE
domain-derived signal, and for each takes the longest map key CONTAINED in it. Farmer reaches
`"insurance"` by two independent paths — industry "Insurance Information & Education", and the
domain string `farmerinsurance.uk` — so removing either would not have saved it.

The recommendation is then deep-merged into the `classification` spec and the flag is what
`discovery_checks/check_directory.go`'s `health-insurer` profile turns into a page
(`PageType`/`PageName` = `health-insurers`) and a homepage section.

**The `Reason` string is the tell.** It ships into the site's spec row and reads to any later
reviewer as a justification — "the one insurer kind built so far; more kinds follow" is a
statement that the *platform's* build order, not the *site's* subject, chose this directory.

## 3. Blast radius, measured

`[MEASURED 2026-09-04]` Exactly **one** site in the estate currently carries
`content_features.health_insurer_directory`: farmerinsurance.uk. So the damage today is
site-local. The mechanism is not: any site whose industry/site_type/category/domain contains
"insurance" without containing "health insurance" gets the health-insurer kind — pet, car,
travel, van, farm, business insurance. `"finance"`-class genericity is guarded; `"insurance"`-class
genericity is not.

```sql
-- the control, re-runnable
SELECT s.domain, ss.data->'content_features'->'health_insurer_directory'->>'kind'
FROM site_specs ss JOIN sites s ON s.id = ss.site_id
WHERE ss.is_current AND ss.aspect='classification'
  AND ss.data->'content_features' ? 'health_insurer_directory';
```

## 4. Related, already closed
`bugs_closed/292` fixed a DETERMINISM defect in this same matcher (a domain containing two
opposite keywords flipped run to run). That fix hardened *how* the map is searched; nobody
questioned *what* the generic `insurance` key returns. A second reason to distrust "this file
was reviewed once".

## 5. Fix candidates, ordered by what makes the bad state unrepresentable

1. **Make `"insurance"` not-recommended, exactly as `"finance"` is.** Four lines, symmetrical
   with the rule already written in the file, and it makes "the platform's build order chose
   this site's directory" impossible to express. Cost: an insurance site gets no directory until
   a sharper signal (`health insurance`, or a new kind) exists — which is what the `finance`
   entry already accepts as the right trade.
2. **Add the kinds** (`agricultural-insurer` / `insurance-broker`) so a sharper signal exists.
   Bigger: a directory kind is a register, a researcher prompt, two components, a check profile
   and a publish profile. Right destination, not a same-day fix.
3. Gate on `industry_tags` rather than substring-matching free text. Narrower than (1) and adds
   a second matching mechanism to an action whose whole design is "one deterministic matcher".
4. Leave and rely on review. Rejected: the review that would catch it is the one whose evidence
   is the spec row carrying the misleading `Reason` string.

**(1) is the fix; (2) is the follow-up.** They are independent — (1) can ship today without
waiting for (2), and (2) does not require (1) to be reverted.

## 6. Site-level remedy for farmer (separate from the platform fix)
Deleting the map entry does NOT retract the page: farmer's spec row already carries the flag and
the page is built and deployed. The site needs its own decision — retract `/health-insurers.html`
and its homepage section, or replace it with the agricultural-broker directory the growth_path
specified. Owner decision, recorded in `docs024_key_docs_latest/farmerinsurance_uk/`.

## 7. Second, cosmetic defect on the same page (do not lose it in the above)
The directory renders provider headings duplicated — "Drewberry Drewberry", "WPA WPA",
"Freedom Health Insurance Freedom Health Insurance" — or name+underwriter run together without
separation: "Saga Health Insurance Bupa", "VitalityHealth Discovery Holdings". That is the
`health-insurer-directory` component's rendering, not the recommendation action, and it will
survive whatever is decided about §6 if the kind is ever legitimately used.

## 8. How to verify a fix
- Unit: a site with industry "Insurance Information & Education" and domain `*insurance*.uk`
  yields **no** directory recommendation, while "health insurance" still yields `health-insurer`
  (the existing `TestMatchVerticalDirectory*` table is the place; it currently PINS the wrong
  behaviour — see `feed_directory_recommendation_action_test.go`'s
  `{"insurance beats finance by length", "insurance finance", …, "health-insurer"}`, which
  asserts precisely the outcome this bug is about).
- Live: re-run the §3 census; a newly classified insurance site must not gain the key.

---

## 8. THE 090 VERDICT: **UNVERIFIABLE** (stopped: iteration-cap) — what it confirmed, what it asked for, and the answers

Run corr `cdcb2981-36ce-4d02-8d37-6ab302aede12`, completed 2026-09-04 16:30:54Z after five
bundles. **It neither confirmed nor refuted the mechanism**, and this file says so rather than
quoting the run as support. Under CLAUDE.md's 2026-07-31 ruling this bug therefore rests on the
filing session's own first-hand verification, which is stated in §2 and itemised below.

**What the run established independently of me, and it is a better fingerprint than anything in
§2 above:**

> "the site's `content_features.health_insurer_directory` carries **the exact Reason string of the
> map's generic `'insurance'` entry** (not the differently-worded `'health insurance'` entry), and
> the site's `identity.industry` is the generic 'Insurance Information & Education' rather than a
> health-insurance-specific string — consistent with the hypothesis."

That is a stronger tell than the industry string alone. The two map entries produce *different*
`Reason` text, and the stored spec carries the generic one verbatim. Whatever else is true, the
row on this site was written by the `"insurance"` branch and not the `"health insurance"` branch.

**Its "still needed" list, and the answers — each read first-hand this session:**

1. *"matchVerticalDirectory's partial-match arm … has not been read"* — it is at
   `feed_directory_recommendation_action.go`, the `for _, signal := range signals` loop: exact map
   lookup first, then longest-contained-key with a lexicographic tie-break. **The run's own
   evidence trail quotes that arm verbatim as citation 2**, so this item was already in hand when
   the conclusion listed it as outstanding — noted as an observation about the run, not a
   complaint.
2. *"EvaluateDirectoryFeaturesAction … has not been read"* — read. It loads the current
   `classification` spec, takes `industry`/`site_type`/`category`, calls `matchVerticalDirectory`,
   and on a recommending match deep-merges `content_features.<SpecKey>` and supersedes-then-inserts
   the spec row. The only intervening logic is the `config == nil || !config.Recommended` guard,
   which is the no-write path. **No allowlist, no override, no per-kind gate.**
3. *"directoryCheckProfiles … could contain an intervening check"* — read
   (`discovery_checks/check_directory.go`). The `health-insurer` profile is a static struct:
   `SpecKey: "health_insurer_directory"`, `PageType`/`PageName` `health-insurers`, snippet and
   listing component names. It consumes the flag; it does not re-decide the kind.

So the answer to the question the symptom asked — *does anything other than the substring match
determine which provider class a site's directory carries?* — is **no**, on a first-hand reading
of all three bodies. That is an assertion by this lane, corroborated but not confirmed by the run.

**If anyone re-files this**, the run's stop reason says how: it was starved of function BODIES
while holding their signatures, so a re-file should name fewer symbols and ask a narrower question
about one of them (the estate's other UNVERIFIABLE today, on the vetcomparison lane, stopped for a
different reason — `scope-not-narrowing` — with the same underlying shape: the bundle never
reached the deciding artefact).
