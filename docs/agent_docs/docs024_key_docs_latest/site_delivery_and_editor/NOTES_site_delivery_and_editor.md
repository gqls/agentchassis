# NOTES — site_delivery_and_editor — append-only, newest at the bottom

## 2026-08-14 — workstream created; plan approved; Phase 1 landed same evening

- Owner-approved plan (verbatim) in `PLAN_2026-08-14_site_delivery_and_editor.md`.
  Origin: the owner's Netlify+editor idea + a reference PDF
  (`~/Downloads/Automated_CMS_Architecture.pdf`), reviewed critically. Key
  rulings taken 2026-08-14: provider bar = COMPLETELY automated (Netlify not
  presumed → plan recommends Cloudflare Pages Direct Upload, seam keeps
  Netlify-OAuth available); ONE multi-tenant editor service; the framework
  KEEPS write access (locks are the two-writer referee); editor after
  handover only.
- Phase 1 (the "usually next day" rider) executed by the ai_site_selling
  lane the same evening: register supersede `SQL_2026-08-14b` (fact + wire +
  ban, ban proven non-inert 13 findings), three locked heroes moved
  surgically (href-asserted), five `nextday_*` content_rewrite items
  dispatched. Acceptance = served grep, not item status.
- Exploration facts the plan rests on (do not re-derive): zero Netlify code;
  zero CF-API automation; `*.ugg2.com` live via the portfolio-sites worker;
  `S3Client` binds one bucket and nothing binds portfolio-sites; no
  archive/zip anywhere; customer auth does not exist; no Ingress; the
  CTS-003 PATCH endpoint + auto-lock + history is the write path to extract;
  `apply_section_edit`'s build_status='approved' defect disqualifies it for
  customer edits (file that defect separately).
- Sharpest risks, ranked: (1) tenant scoping must be structural
  (session→site_id once; cross-tenant probe is the acceptance), (2) CF Pages
  partial-upload — acceptance is served-hash equality never API 200,
  (3) truncated ZIP = silent contractual failure — stream + alert.

## Coordination

- webdesign.uk's CTA components are permanently locked (SQL_2026-08-12k);
  the 268 fleet-fix thread owns unlocking. Phase 2's canary must be a quiet
  portfolio site, NOT webdesign.uk.
- Council: one run per phase (2–6), submit before/alongside each shipping
  commit; register entries listed in the PLAN roll-up.
