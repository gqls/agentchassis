# NOTE from the web_admin_console lane, 2026-08-24

Four things this lane changed that the 08-21 handoff's readers must know:

1. **D-A is ANSWERED** (owner 2026-08-22): `DECISION_2026-08-22_D-A_admin_console_public_via_access.md`.
   Narrow: admin-dashboard's gateway only, one hostname, behind Cloudflare Access. It does
   NOT license `/c/` or `/d/`.
2. **D-B is ANSWERED AND APPLIED** (owner 2026-08-22): build_duration is "three or four
   days, usually sooner", live at the bot. The retired figure's ban is written but HELD —
   `SQL_2026-08-22b_…_HOLD.sql` refuses until the 4 affected pages re-render (10 components,
   census in the file).
3. **`/c/` is MOVED off the shopfront** (owner 2026-08-24): `box/links.webdesign.uk.nginx`,
   apex vhost edited, box apply pending. **`links.webdesign.uk` is the canonical emailed-links
   host — the delivery email must mint on it.** Prefetch guard live in v1.0.1332
   (`0e9cb31ee`, Council-Submitted 6b1726ab); mail-scanner residual still the owner's call.
4. The parking-page-rule trap and the box's real hostname inventory (7 hosts, 3 ports, incl.
   noted.co.uk) are in LANDMINES.md, 2026-08-23 entries.

Continue-here for the console work: `../web_admin_console/HANDOFF_2026-08-24_continue_here.md`
