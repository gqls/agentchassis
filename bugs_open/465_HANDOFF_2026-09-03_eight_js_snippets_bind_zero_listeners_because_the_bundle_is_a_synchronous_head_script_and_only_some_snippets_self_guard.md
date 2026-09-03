# 465 — Eight `js_snippets` rows bind ZERO listeners in production, because the bundle is a synchronous `<head>` script and self-guarding is a per-snippet convention

**Filed** 2026-09-03 by the gripper dossier lane (session: robot hands).
**Status** OPEN. **Severity** medium — silent, fleet-wide, and invisible to every
status the platform records.
**Origin** raised as a `medium` objection by the council's `bug_historian` seat on
corr `5775dc10-c791-4285-9f4c-249a055b5aa3` (round 1, verdict REVISE), then
**confirmed by census** rather than accepted on the seat's word. The seat was right,
and the exposure is larger than its objection assumed.

---

## 1. The mechanism, in plain terms

A site's JavaScript snippets are concatenated into one bundle, `/assets/js/snippets.js`.
The site chrome template includes that bundle as a **plain synchronous `<script>` inside
`<head>`** — no `defer`, no `async`. Measured on `/gripper-report.html`: the tag is at
line 2219 and `</head>` closes at 2238.

So every snippet in the bundle **executes while `<head>` is still being parsed** —
before `<body>` exists at all. Any snippet that looks for its mount point at that moment
finds nothing.

There are two ways that fails, and **neither one raises an error**:

- `document.querySelector('...')` returns **`null`**, and the snippet's own
  `if (!el) return;` bails cleanly.
- `document.querySelectorAll('...')` returns an **empty NodeList**, and the snippet's
  `.forEach(...)` iterates **zero times**, binding zero listeners.

The second is the nastier one: there is no null to guard against, no exception, and no
log line. The snippet ran, completed successfully, and did nothing.

**Whether a snippet survives this is decided by convention, not by the platform.** Some
snippets wrap themselves in a `document.readyState === 'loading'` /
`DOMContentLoaded` guard; nothing requires it, nothing checks it, and a snippet author
gets no signal when they omit it.

## 2. The census — `[MEASURED 2026-09-03]`

```sql
SELECT name, octet_length(js_content) AS bytes,
       (js_content LIKE '%DOMContentLoaded%') AS guarded,
       (js_content LIKE '%querySelector%' OR js_content LIKE '%getElementById%') AS mounts
  FROM js_snippets ORDER BY guarded, name;
```

**18 rows. 9 carry no DOM-ready guard; 8 of those 9 query the DOM.**

| unguarded snippet | bytes | how it fails at head-parse |
|---|---|---|
| `accordion` | 560 | `querySelectorAll(".accordion-trigger").forEach` → 0 iterations |
| `copy-to-clipboard` | 372 | `querySelectorAll(".copy-btn").forEach` → 0 iterations |
| `counter-animate` | 815 | `querySelectorAll` at top level → 0 iterations |
| `form-validation` | 671 | `querySelectorAll("form[data-validate]").forEach` → 0 iterations |
| `lazy-load-images` | 509 | `querySelectorAll('img[loading="lazy"]').forEach` → 0 iterations |
| `mobile-menu-toggle` | 338 | `querySelector(".menu-toggle")` → **null**, `if (menuToggle && mobileMenu)` bails |
| `smooth-scroll` | 307 | `querySelectorAll('a[href^="#"]').forEach` → 0 iterations |
| `typing-effect` | 538 | `querySelectorAll(".typing").forEach` → 0 iterations |

(`news-date-formatter`, 506 B, is the 9th unguarded row and is the only one that does
**not** query the DOM — no exposure.)

The 9 guarded rows — `chat-input-box-loader`, `gripper-report-intake-widget`,
`hero-card-carousel`, `lobby-grid-loader`, `provocation-card-loader`,
`provocations-archive-loader`, `scroll-reveal`, `stat-band`, `teaser-reveal-panel` —
are unaffected. Note the shape of that split: **the large, recently-authored loaders
guard; the small, generic, older utility snippets do not.**

⚠ **This is a count of ROWS IN THE LIBRARY, not of broken pages.** It is the upper
bound on exposure, not the damage. See §4 for what is deliberately not measured yet.

## 3. How it was found, and why it was nearly missed twice

`gripper-report-intake-widget` had exactly this defect. It was diagnosed on 2026-09-03
and fixed with an `init()` + `readyState` guard (commit `991cf8b8b`, seed 651).

Two earlier readings of the same widget said it was **working**, and both were true at
the wrong altitude:

- the widget code **was** in the served bundle (`grep` confirmed it);
- the mount div **was** in the served page (`grep` confirmed that too).

Neither fact implies a rendered button. That misreading is already logged in
`WRONG_CALLS.md` under 2026-09-03 — *"in bundle + mount div ≠ rendered button"*.

The generalisation was then almost missed a third time: the fix shipped as a one-row
patch, and it took a council seat objecting on **scope** to ask whether the convention
itself was the defect. It was.

## 4. What is NOT established, and must be before anyone sizes this

**Filed honestly rather than dressed up:** the census counts library rows. It does
**not** establish that eight features are broken on live sites. Three things stand
between a row and live damage, none of them measured yet:

1. **Bundle membership.** A snippet is only exposed if it is actually rendered into
   some site's bundle. `render_js_snippets_for_site` decides that; a row in the library
   that reaches no bundle is latent, not broken.
2. **Markup presence.** `smooth-scroll` only matters on a page that has `a[href^="#"]`;
   `accordion` only on a page with `.accordion-trigger`. A snippet bundled onto a site
   whose pages carry none of its selectors was never doing anything anyway.
3. **Head placement is per-template.** The `<head>`-synchronous include was measured on
   robot-hands.com's chrome. Another site's chrome may include the bundle before
   `</body>`, in which case its snippets work by luck of placement.

**The discriminating query for (1)+(3) is the next step and it is cheap.** Do it before
quoting "8 broken features" anywhere — that number is not yet earned.

## 5. Fix candidates, ordered by what closes the door

1. **Emit `defer` on the bundle `<script>` tag** (or move the include to just before
   `</body>`). One change at the renderer, and it makes the failure **unrepresentable**
   for every snippet, existing and future, without touching a single snippet body. A
   deferred script runs after parsing, so `querySelector` always sees `<body>`.
   *This is the one that ends the class.* Cost: needs a check that no snippet depends on
   running before first paint.
2. **Wrap every snippet at render time.** `render_js_snippets_for_site` emits each body
   already wrapped in a ready-check, so an unguarded author cannot reintroduce the bug.
   Also class-closing, but it rewrites content at render, which is harder to debug and
   costs bytes against per-snippet budgets (the gripper widget lives under a hard 8192 B
   verify bound with 8 B of headroom today).
3. **Add the guard to the 8 rows individually.** Lowest risk per row, and the only
   option that needs no platform change — but it is exactly the "fix one call site at a
   time" shape this platform has repeatedly paid for, and it leaves the ninth author to
   rediscover it. **Do not stop here.**
4. **A detector only** (fail a check when a `js_snippets` row queries the DOM without a
   guard). Worth having alongside 1 or 2; on its own it documents the trap rather than
   removing it, and "authors must remember" is the defect.

**Recommendation: (1), with (4) as the ratchet.** (3) is a stopgap for a specific site
with a specific broken feature, not the fix.

## 6. Verifying any fix — the trap is that grep passes while the page stays broken

`grep`ping the bundle proves the code shipped. It does **not** prove a listener bound.
The gripper widget was verified three ways, in increasing strength:

- served bundle greps (`DOMContentLoaded` count ≥ 2, `function init()` = 1) — necessary,
  **not sufficient**, and this is the check that gave two false passes;
- **execution** under a DOM stub reproducing the real load order (`querySelector`
  returns null while `readyState === 'loading'`), running the **old** widget as a
  negative control — old: 0 listeners, no button; new: 1 listener, button after
  `DOMContentLoaded`. A harness that only ever runs the new code asserts its own
  bookkeeping;
- a **human loading the page in a browser** — the only place the proof actually exists.

Insist on the second at minimum. The first is what failed here.

## 7. Related

- `WRONG_CALLS.md`, 2026-09-03 — the "in bundle + mount div ≠ rendered button" entry.
- `LANDMINES.md` — the `js_snippets` entry filed alongside this bug.
- Council corr `5775dc10-c791-4285-9f4c-249a055b5aa3`, seat `bug_historian`, round 1 —
  the objection that started this, and the reason a per-instance fix was not the end of it.
- `docs024_key_docs_latest/robot_hands_gripper_dossier/` — the lane that hit it, its
  NOTES entry for 2026-09-03, and seed 651 for a worked example of the guard shape.
