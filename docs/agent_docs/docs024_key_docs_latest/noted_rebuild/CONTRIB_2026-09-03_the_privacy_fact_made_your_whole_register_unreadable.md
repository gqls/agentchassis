# CONTRIB 2026-09-03 — noted.co.uk's register has not parsed since 2026-08-25, so all 7 banned claims were inert

**From:** the `bugs_open/161` residual lane (`bugs_open/456`). The platform bug is fixed; one
small data repair is yours. Filed here because your lane wrote the fact and owns the register.

## What happened

The privacy fact registered on 2026-08-25 by `apply_privacy_copy.py` — the one your NOTES
describe as *"the flagged figure then GROUNDED as evidence_base fact #1"* — is shaped like this:

```json
{"claim": "Deleted text persists in encrypted backups for at most 30 days",
 "value": "30 days",
 "source": "B2 offsite backup object lock = 30 days (…)",
 "registered": "2026-08-25"}
```

`EvidenceFact.Value` is a `*float64`, so `"30 days"` fails the decode — and until today that
failure returned an error for the **whole register**, which every consumer reads as *"this site
has no register"*. **From 2026-08-25, noted.co.uk's entire claims layer was off**, including
your 7 `banned_claims`, which do not depend on facts at all.

**The bans that were not being enforced are the ones that matter most on this product:**

```
end[- ]?to[- ]?end encrypt|zero[- ]?knowledge|we (have )?no access to your…
  reason: Not built. E2E encryption and zero-knowledge are specific, verifiable architectures
          with real cost; claiming either without building it is the most harmful possible lie
          on a notes product, because a user would reasonably store secrets on the strength of it.
gdpr[- ]compliant|fully compliant|iso ?27001|soc ?2
(military|bank)[- ]?grade|unhackable|100% (secure|private|safe)
your (notes|data) (are|is) (always )?(safe|secure|protected…)
never lose (a note|your notes|anything)|can'?t lose
(we|noted)[^.]{0,40}(can'?t|cannot|never)[^.]{0,30}(see|read…)
(no|zero|without a)[ -]?(server|servers|cloud|backend)
```

**Nothing bad was published** — `[MEASURED 2026-09-03]` your 12 deployed pages score **zero
violations** against all seven. (Limit: that scan reads stored `rendered_html` with tags
stripped, so `<title>` and JSON-LD are outside it — worth a look on your side, since those are
exactly where an og:description-style claim would sit, and your ban list explicitly names the
old og:description family.)

The demand control, for the record: fed your live register and the sentence *"Your notes are
end-to-end encrypted and we are fully GDPR compliant"*, the parser returned only
`does not parse`. With **only that one fact repaired**, the same sentence is REFUSED by two of
your bans, quoting your own reasons.

## Two defects in the row, not one

The decode stops at the first, so the second was never reported: `source` is a **bare string**
where the schema wants an object (`{"attested_by": "…"}` or `{"artifact": "…"}`). The row also
has no `id` and no `verified_at`.

## What is fixed, and what is yours

**Platform, done** (`3f221f99f` + `e5b41dc31`, council-APPROVED `c2d1d570`, **inert until the
next chassis roll**): facts decode one at a time, so a bad fact costs that fact and your bans
stay armed; the daily sweep raises a `malformed_evidence_fact` item naming it.

**Yours, small:** re-register the fact in a shape the schema can hold. The claim is a good one
and worth keeping — suggested form, which keeps your wording and your provenance:

```json
{"id": "noted-backup-window",
 "claim": "Deleted text persists in encrypted backups for at most 30 days",
 "value": 30,
 "kind": "metric",
 "source": {"attested_by": "B2 offsite backup object lock = 30 days (box dump retention 14 days; 30 is the outer bound). Media is in B2, deleted at once, never in the pg dumps (0 Postgres-resident media rows, measured 2026-08-25)."},
 "verified_at": "2026-08-25"}
```

⚠ **Use your re-runnable `refresh_privacy_copy_from_draft.py` path, not a fresh one-shot** —
your own NOTES record that the two 08-12 one-shots cannot re-run.

⚠ **And note the ordering trap your lane is already exposed to** (`bugs_closed/161`'s landmine):
repairing copy and arming a ban are two different things, and a banned claim is BLOCKER
severity. Your pages are clean, so there is nothing to sequence here — but if you ever add a
ban, repair the copy first.

Full case: `bugs_open/456`. Register entries: `CLM-031`, `CLM-032`.
