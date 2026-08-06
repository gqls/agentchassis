# NOTES — bugfix 203 cleanup (append-only, newest at the bottom)

## 2026-08-06 — session start, claim, and the two cheap wins

- Triage: 206/207/205 (the newest opens) all actively claimed by other live sessions —
  verified in their transcripts (edits + task lists), not just `who-owns.py`, which is
  lagging by design. 203 unclaimed; took it. Claim committed `dfdcdecd2`.
- Council verdict for the source fix (`880a405a6`, corr `42eda9a5`): **APPROVED r1**,
  3 objectors, 5 advisory objections, none high. Read in full from
  `diagnosis_artifacts.body`. First query failed — guessed a `content` column instead
  of reading the schema; `\d diagnosis_artifacts` first, then it was `body`.
- Liveness: proven by ancestry against 197's pod-proven commit (see RUNBOOK). Then a
  fresh roll to **v1.0.1261** landed mid-session (confirmed both pods) — built from
  later HEAD, so carries the fix a fortiori.
- The census re-run and everything downstream still pending at this note.

Key discovery, pre-empting the council's own class-audit ask: `contextToMap`'s
DEFAULT VALUES map (component_library.go ~1136–1147 at `880a405a6`) still fabricates
`primary_cta_url: /contact.html` / `secondary_cta_url: /about.html`, and the alias
block copies `cta_url → primary_cta_url`. The 203 fix removed the `cta_url` default
but the `primary_*` family keeps the class alive on the regex-fallback path.
[UNMEASURED at this note]: whether any live template consumes `primary_cta_url`, and
what the regex renderer ships for an absent key (bug_historian M2 warns: possibly
literal `{{.field}}` text, which is WHY those defaults exist).
