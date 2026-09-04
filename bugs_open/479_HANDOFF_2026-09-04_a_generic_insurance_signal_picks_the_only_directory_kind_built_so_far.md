# BUG 479 — a generic "insurance" signal picks the ONLY directory kind built so far, and a farm insurance hub ships a private-medical-insurer directory

**Filed** 2026-09-04 by the `farmerinsurance_uk` lane.
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
