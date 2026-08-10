# NOTES — noted.co.uk rebuild

Append-only, newest at the bottom. Technical log: evidence, commands, what the
system actually said, and every misstep.

---

## 2026-08-10 — session 1, opening survey

### What noted.co.uk actually is today

`[MEASURED]` A hand-built vanilla-JS single-page note app. No build step, no
framework, no server. Source of truth on disk is
`~/projects/domains/noted.co.uk/`, which holds **three** generations:

| dir | mtime | what it adds |
|---|---|---|
| `01` | 2026-01-27 | text notes only |
| `02 - voice notes` | 2026-01-29 | voice recording + camera |
| `03-sharing` | 2026-03-20 | share button, og: tags, favicon, manifest |

> **CORRECTED 2026-08-10, same session — the paragraph below was WRONG and the
> error was mine, not the system's.** I first wrote that "the live site is `02`,
> not `03`" and that the live page was *missing* the share button. I had read the
> direction of a `diff` backwards: in `diff live 02`, a `<` line is present in
> **live** and absent from `02`. The share button is a `<` line. Live **has** it.
>
> What caught it: a real browser. I had gone on to reason that live `app.js`
> bound a listener to a `#btn-share` that live `index.html` lacked, concluded the
> app must be crashing during init, and was about to file that as a
> site-breaking bug. A Playwright probe against `https://noted.co.uk/` returned
> **0 uncaught page errors, 0 console messages**, a rendered note list, a working
> `+ New`, and a real download from the backup button. Then `sha256sum` settled
> it: the bytes I had downloaded from B2 are **identical** to the bytes the
> browser receives (`index.html` `95e70d99…`, `app.js` `8dec41f8…`), and
> `grep -c btn-share live_noted/index.html` = **1**.
>
> The cheap check that would have caught it before I built a theory on top:
> `grep -c btn-share` on the downloaded file — one command, and it directly tests
> the claim instead of re-reading the diff that produced it. Logged to
> `WRONG_CALLS.md`. Two things worth keeping: **confidence tracked the size of
> the conclusion, not the strength of the evidence** — "the live site is broken"
> felt more certain than "the live site is fine", because it was a more
> interesting story. And a diff's direction is exactly the kind of detail that
> reads as settled while being a coin-flip.

`[MEASURED, browser-verified]` **The live site sits BETWEEN local `02` and local
`03-sharing`.** It has the share feature in both `index.html` and `app.js`. What
`03-sharing` adds on top of live is:

- the four `og:` meta tags, `favicon.ico` and `apple-touch-icon.png` links
  (`index.html`, 6 lines);
- a reworked share flow in `app.js` (138 diff lines) — live shares via
  `navigator.share` only, whereas `03` adds a desktop fallback that generates a
  self-contained portable `.html` with images embedded as data URIs.

So `03-sharing` is still **local work that was never deployed**, but the gap is
narrower than I first said, and it is a *refinement* of a feature that is already
live rather than the feature's introduction.

`[MEASURED]` Two dangling references in `03-sharing` that must not be shipped as
they stand: `og:image` points at `https://noted.co.uk/icon-512.png`, and the
`<link rel="icon">`/`apple-touch-icon` point at `/favicon.ico` and
`/apple-touch-icon.png`. **`icon-512.png` does not exist anywhere** — not in the
bucket, not in the local dir. `favicon.ico` and `apple-touch-icon.png` exist in
`03-sharing/` on disk but are **not in the bucket**. Deploying `03` unchanged
would ship a broken social preview and two 404s.

`[MEASURED, browser-verified]` **The live app is healthy.** Playwright against
`https://noted.co.uk/`: 0 uncaught errors, 0 console output, `#save-status` reads
`Saved`, the note list renders, `+ New` adds a second note, and `#btn-download`
produces `noted-backup-2026-08-10.json`. Whatever else is wrong with this site,
it is not broken for its users today. Do not write a wind-down notice that
implies otherwise.

### Where it is hosted — it is already on the framework's own bucket

`[MEASURED]` This was the surprise. Response headers from `https://noted.co.uk/`:

```
server: cloudflare
x-amz-version-id: 4_ze273722284f8b56e9ead0917_f113bc69b73b9e38d_d20260209_...
last-modified: Mon, 09 Feb 2026 14:28:55 GMT
```

That bucket id `e273722284f8b56e9ead0917` is **`portfolio-sites`** — confirmed
against `b2 bucket list`. And the files are there under the domain prefix:

```
b2 ls --recursive b2://portfolio-sites/noted.co.uk/
  noted.co.uk/index.html          3161  2026-02-09
  noted.co.uk/js/app.js          20516  2026-02-09
  noted.co.uk/css/main.css        8046  2026-02-09
  noted.co.uk/guides/about.html   3513  2026-02-09
  ...
```

**noted.co.uk is a hand-built site sitting in the framework's deploy bucket**,
served through Cloudflare. It is the same shape of problem as the webdesign.uk
shopfront that produced the OWNER RULING of 2026-08-04 — except this one predates
that ruling by six months and is an *application*, not a brochure page.

`[MEASURED]` DNS is Cloudflare-proxied (`172.67.171.14`, `104.21.63.162`), so
the bucket is not addressed directly by visitors.

### Whether anyone uses it — UNMEASURABLE from here, do not claim either way

`[UNMEASURED]` **I could not establish a user count and nobody should quote one.**
Two independent reasons, both fatal:

1. **B2 access logging stopped on 2026-05-10.** `personae-access-logs/portfolio-sites/`
   holds 6,752 objects, first `2026-03-14`, last `2026-05-10-12-40-10`. Three
   months of silence, and today is 2026-08-10. So there is no recent data at all.
2. **Even the data that exists cannot separate humans from bots.** Cloudflare is
   the client from B2's point of view, so the user-agent field is `-` on every
   row. Sampled the last 30 log files: 10 hits on `noted.co.uk/index.html`, all
   `200`, all UA-stripped — against a much larger volume of obvious scanner
   noise on the same prefix (`wp-signup.php`, `wp-admin/...`, `secrets.json`).

The disconfirming result I looked for and did **not** get: a populated UA column
that would let a human request be told from a probe. So "noted has no users" is
**not** a finding — it is an absence of instrumentation. Design the wind-down as
though there are users, because nothing here can show there are not.

### The data at risk, and the trap in the backup button

`[MEASURED]` Read `03-sharing/js/storage.js`. IndexedDB `NotedDB` v4, four object
stores: `notes`, `history` (last 5 revisions per note), `audio`, `images`.
Voice clips and photos are `Blob`s keyed by note id, capped at 5 each.

`[MEASURED]` **The "Backup text notes" button exports one store of the four.**
`03-sharing/js/app.js:111-119`:

```js
const notes = await Storage.getAllNotes();
const blob = new Blob([JSON.stringify(notes, null, 2)], { type: 'application/json' });
a.download = `noted-backup-${...}.json`;
```

`getAllNotes()` reads the `notes` store only. **Voice recordings, photos and
version history are not in the export and there is no other bulk export path.**
The single-note share does embed images in a portable HTML file
(`app.js:356 downloadPortableHtml`), but that is one note at a time and drops
audio.

This is the single most important operational fact in this document: **telling
users to "back up" using the button that exists would still lose their voice
notes and photos.** The button's label is honest as far as it goes — it says
"text notes" — but a wind-down notice that says "back up your notes" while
pointing at it would not be.

Consequence for the notice: it must either say plainly that media is not
included, or we ship a full export first. Preference is to ship the full export,
because a wind-down notice whose advice is "your recordings cannot be saved" is
not an acceptable thing to tell a user about their own data.

### The framework does not know this domain exists

`[MEASURED]` `SELECT ... FROM sites` — 39 rows, no `noted.co.uk`. Nothing in
`bugs_open/`, `bugs_closed/`, or any `.md`/`.sql`/`.go`/`.sh` in the repo mentions
the domain (`grep -rl "noted\.co\.uk"` → no hits). No existing workstream dir.
So this is greenfield adoption, not a takeover of someone else's in-flight work.

### The product promise inverts, and that is a checking problem

The current site's whole pitch is that there is no server. `guides/about.html:34`:

> "This means we can't see your notes, read your text, or listen to your
> recordings."

and `index.html:33`: **"Local & Private.** Notes are stored in your browser."

The rebuild the owner asked for — server-side storage, log in from another
browser — **contradicts every one of those sentences.** That is not merely a copy
edit. The framework has claims gating (`evidence_base`, `banned_claims`), and any
rebuilt page that keeps the old privacy language while shipping a server would be
making a false claim about its own behaviour. Flagging it here because it is the
kind of thing that gets copy-pasted forward from the old site without anyone
noticing the guarantee changed underneath it.

### Framework machinery found so far that bears on an *app* (not a brochure)

`[MEASURED]` `site_specs.aspect` values currently in use across the fleet, by
count — this is the vocabulary a fully-specified site is written in:

```
audience 29 · identity 21 · content_direction 21 · design_intent 21 ·
strategy 19 · classification 18 · briefing 18 · site_config 15 ·
resolved_composition 13 · tools 13 · evidence_base 11 · vertical_landscape 10 ·
submission 9 · mission_brief 8 · imagery_style_guide 8 · site_plan 6 ·
structure 6 · site_archetype 6 · design_reference 6 · voice 5 · ...
```

`[MEASURED]` `experience_patterns` — 9 rows. The one that matters:
**`timed-remote-challenge-loop`** (`kind=micro-journey`, `status=draft`), which is
"a visitor starts a session against a remote engine, submits, receives generated
content ... and receives an outcome". **This is existing precedent for a framework
page that talks to a backend over HTTP**, and it carries `contract`,
`degraded_states`, `data_contract` and a `criteria_template`.

Its criteria template tells us what the checker can actually assert today:
`page_status_ok`, `selector_exists`, `selector_count`, `no_horizontal_overflow`,
`interaction` (fill/click then expect a selector to match text), `degraded_state`.

**And, valuably, what it cannot** — the pattern author left `_unsupported` notes
inline:

- `expect_within_ms` **does not exist**; the runner asserts 300 ms after the step.
- `retries` does not exist.
- checks "cannot be ordered or made conditional on another check today; this one
  is the whole honesty rule and is **currently unexpressible**".
- "no way to induce or simulate a failing dependency in a check run — the
  honest-degradation clause **cannot be tested by the platform at all**".

Those four gaps are load-bearing for a notes app, because the things most worth
checking about it ("the note is still there after reload", "a failed save does
not silently lose text") are exactly ordered, stateful, and dependency-dependent.
Do not plan checks that assume those gaps are closed without re-reading the
runner first — this is a `draft`-status pattern and the notes are dated.

### Open questions at end of survey (answers pending, do not guess)

1. Which VM the owner means by "an existing vm", and what is already on it.
2. Whether the framework's build pipeline can produce an app shell at all, or
   only content pages.
3. Whether `auth-service` is usable for end-user login or is internal-only.

</content>
</invoke>
