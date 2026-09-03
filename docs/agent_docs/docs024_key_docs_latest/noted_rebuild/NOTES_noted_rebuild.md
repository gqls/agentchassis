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

---

## 2026-08-10, later — box selection, measured on the box

Owner ruling: Mythic Beasts preferred over Hetzner where the choice is easy;
**apis.uk carries the API and nothing else**. Estate is 4 boxes — idea.uk
(Hetzner CX, money-live), relojistas.com (Hetzner CPX22), tools-api island
(Mythic Beasts `vds:toolsapisuk`), webdesign.uk (Mythic Beasts `vds:webdesign`).
The ruling excludes the island outright and deprioritises both Hetzner boxes, so
exactly one candidate remains and there is no judgement left to exercise.

`[MEASURED 2026-08-10]` on `webdesign.vs.mythic-beasts.com` (176.126.243.62,
Cambridge), via `ssh -i ~/.ssh/webdesign_box_ed25519 root@…`:

```
nproc      -> 2
free -h    -> 7.8Gi total   506Mi used   6.2Gi free   7.3Gi available
df -h /    -> 50G total     2.4G used    47G free (5%)
ss -tln    -> *:8081  127.0.0.1:8080  0.0.0.0:22  127.0.0.1:20241
which psql postgres -> none
running    -> cloudflared, nginx, ssh, unattended-upgrades (+ base system)
```

**Worth recording because it contradicts the paperwork in a useful direction.**
`HANDOFF_2026-08-03_P1_shopfront.md:112` and `HANDOFF_2026-08-04_continue_here.md:188`
both specify **4 GB** ("the one number not to trim"). The provisioned machine has
**7.8 GB**. A requirements table is a statement of what to buy, not of what was
bought — read the box.

`[CORRECTED]` My earlier entry in this file said the island "has the database and
the backup habit already" and put the webdesign box at 8 GB. Both were
directionally right and imprecise: the island is 1 core / 2 GB / **20 GB** SSD,
£16.20/mo (`RUNBOOK_island.md:4-5`), and the webdesign box is 7.8 GB / 50 GB. The
island's disk in particular I had never checked and should not have implied.

### Incidental security finding — webdesign lane's, not this one's

`[MEASURED]` `webdesign-chat` (pid 105145) listens on **`*:8081`**, all
interfaces. Mitigation measured too, so this is a gap and not an exposure:

```
ufw status verbose -> Status: active
                      Default: deny (incoming), allow (outgoing)
                      22/tcp (OpenSSH) ALLOW IN  Anywhere      <- the only rule
```

and an actual outside TCP connect to `176.126.243.62:8081` **does not reach it**
(tested from this host; the disconfirming result — a successful connect — would
have made this an incident rather than a note).

Why it still matters: `webdesign.uk.nginx:32-34` states the box's posture
explicitly — *"LISTENS ON LOOPBACK ONLY, ON PURPOSE… Binding 127.0.0.1 means even
a firewall mistake exposes nothing."* `webdesign-chat` does not honour that, so
ufw is its sole control rather than its second one. `ufw allow 8081` or a
`--force reset` would publish a service that spends money against the Anthropic
API. **Not fixed here** — another lane's service, outside this task's scope, and
changing it would restart a live commercial service on somebody else's behalf.

**The transferable form:** a loopback-binding *convention* enforced by nothing is
not a control ([[a-doc-comment-is-not-an-enforcement-mechanism]]). The next
service on this box — which will be noted's — must bind `127.0.0.1` explicitly,
and that is now written into PLAN §3a-i rather than left to whoever writes it.

### Consequence for noted's design

Media (voice recordings, photos) goes to **B2**, not the box disk. Two reasons,
and the second is the load-bearing one: it is better architecture anyway, and it
removes the only shared resource on which noted could take the webdesign.uk
shopfront down. noted's storage growth is unbounded (audio and images per note,
per user, forever); the shopfront's is not. Postgres on the box then holds text,
metadata and object keys, which stays small indefinitely.

---

## 2026-08-10, later still — Postgres installed on the box (phase 1 begun)

Owner: *"please carry on and install postgres there"*. Done on
`webdesign.vs.mythic-beasts.com`. Everything below is `[MEASURED]` on the box.

### Before/after control, because this box is not ours alone

Baseline taken first, and re-taken after:

| | before | after |
|---|---|---|
| external `https://webdesign.uk/` | 200, 25970 B, sha `26293b6d…` | **byte-identical** |
| box-local `Host: webdesign.uk → :8080` | 200, 28419 B | 200, 28419 B |
| chat `:8081/health` | 200 | 200 |
| services | nginx, cloudflared, webdesign-chat active | + postgresql, all active |
| memory | 506 Mi used | 542 Mi used (postgres RSS 89 MB) |
| disk | 2.4 G used, 47 G free | 2.7 G used, 47 G free |

**The box-local check is the one that matters.** External `webdesign.uk` 302s to
`webdesign.co.uk`, a *different* site served from the bucket — so the external
check would have passed even if I had broken nginx on the box. Noted in the
RUNBOOK; anyone testing this box from outside is testing the wrong thing.

### What was installed

PostgreSQL **16.14** (Ubuntu 24.04.4), cluster `16/main`, `listen_addresses =
localhost`, bound `127.0.0.1:5432`. Verified **not reachable** from off-box by an
actual TCP connect to `176.126.243.62:5432` (the disconfirming result — a
successful connect — would have made this an incident). ufw untouched: still
default-deny inbound, port 22 the only rule.

Role `noted`, database `noted` owned by it, password generated **on the box**
with `openssl rand` and never echoed to my terminal; written to
`/etc/noted/noted.env`, mode 600 root:root, following idea.uk's
`/etc/idea/idea.env` pattern.

### Three traps, all of which produce a silent nightly failure

**1. `CONNECT` is granted to `PUBLIC` on every database by default.** My negative
control — "can the `noted` role open the `postgres` database?" — **FAILED on
first run**: it returned `1`. That is stock Postgres behaviour, not a
misconfiguration, but this box is now a *shared cluster* (webdesign + noted +
whatever comes next), so it matters. Closed with `REVOKE CONNECT ON DATABASE
postgres, template1 FROM PUBLIC`. Re-ran the control and it **changed answer** to
`FATAL: permission denied for database "postgres"`, with the positive control
(own database still reachable) still passing. A check that never changed answer
would have proved nothing.

**2. `pg_dump -f <file>` cannot write into a root-only backup directory.** The
dump dir is `700 root:root` — correct, since a dump contains every note — but
`pg_dump` runs as the `postgres` user under `sudo`, so *it* opens the output file
and gets `Permission denied`. Fix: dump to **stdout** and let the script's own
root shell do the redirect.

**This one is the reason to never trust `systemctl enable`.** I enabled the timer
*and then ran the service immediately* rather than waiting for 03:20. It failed
instantly. Had I only enabled it, the first evidence would have been an empty
`/var/backups/noted/` discovered at some point after real user notes existed —
and the timer would have reported nothing, because a failed oneshot leaves no
trace anyone is watching. **Enabling a timer is not evidence that the job works.**

**3. The same permission wall inverts on restore, and it counterfeits a corrupt
backup.** My first validity check ran `sudo -u postgres pg_restore -l <dump>` and
printed **INVALID** — which I nearly recorded as "the backup is bad". It wasn't:
`postgres` cannot *read* a 600 root-owned file, and `pg_restore` reports that as
an unreadable archive. As root, `pg_restore -l` exits 0 and lists a valid archive
(6 TOC entries, gzip, dump version 1.15-0).

The general form, and it is the same shape as this morning's diff error: **a
check that fails for its own reasons is indistinguishable from the thing it was
checking failing.** Both times the wrong answer was the alarming one, and both
times the fix was to test the claim by a different route rather than re-run the
same instrument.

### The restore is PROVEN, not assumed

A dump that has never been restored is a file, not a backup. Round trip:

```
psql "$URL" -c "CREATE TABLE probe(...); INSERT INTO probe VALUES (1,'survives a restore');"
/usr/local/sbin/noted-pg-backup.sh          -> wrote …170314Z.dump (3811 bytes)
sudo -u postgres createdb noted_restore_probe
sudo -u postgres pg_restore -d noted_restore_probe < <dump>
sudo -u postgres psql -d noted_restore_probe -tAc "SELECT v FROM probe WHERE id=1;"
  -> survives a restore
```

Note `pg_restore -U postgres -h /var/run/postgresql` as root does **not** work —
peer auth maps to root, not postgres. Redirect on stdin instead.

Timer: `noted-pg-backup.timer`, daily 03:20 UTC + up to 15 min jitter,
`Persistent=true`. `Nice=10` and `IOSchedulingClass=idle` because a live
shopfront shares the disk. Next run confirmed scheduled.

### The gap that is NOT closed, and must be before launch

`[STATED, NOT FIXED]` **The backups are on the same disk as the database.** That
covers a bad migration or a `DELETE` without a `WHERE` — the failures that
actually happen — and nothing else. If the box is lost the dumps go with it.

Deliberately not fixed here: an off-box copy needs a credential **on** the box,
and this box's whole posture is that it holds none and dials nothing in. A B2
application key scoped write-only to a single backup prefix is the right answer
(it dials OUT, like cloudflared, so the posture survives), but issuing one is the
owner's call. **This must be resolved before the first real user note lands.**
Written into the script's own header, not only here, because the script is what
the next session opens.

Also still absent: nothing on this box backs up `webdesign-chat`'s own
`state.json` / `transcripts.jsonl`. Not mine, but nobody has it.

---

## 2026-08-10 — B2 backup key probed against real B2, and the shopfront characterised

### The key capability set is TESTED, not reasoned

Created a temporary bucket-scoped key, ran every operation against it, deleted
the key and the test object. Results, all `[MEASURED]`:

| operation | `listBuckets,writeFiles` | verdict |
|---|---|---|
| `b2 account authorize` with `writeFiles` ALONE | **refused** | `listBuckets` is mandatory for the CLI |
| upload | works | the only thing the job needs |
| download | `unauthorized` | a stolen key cannot read anyone's notes |
| `b2 ls` | `unauthorized` | cannot enumerate backups |
| `b2 rm` | `unauthorized for application key with capabilities 'writeFiles,listBuckets'` | cannot delete |
| `b2 ls` on another bucket | `Application key is restricted to buckets: ['personae-noted-backups']` | scoping holds |
| **`b2 file hide`** | **SUCCEEDS** | see below |

**Two of these would have cost a silent nightly failure or a false security
claim.**

**`listBuckets` is required by the CLI itself** — `b2 account authorize` refuses
outright with *"application key has no listBuckets capability, which is required
for the b2 command-line tool"*. Reasoning from "the job only needs to write" gives
the wrong answer, and the failure would have arrived at 03:20 with nobody
watching.

**`writeFiles` includes HIDE.** After hiding the probe, `b2 ls --recursive`
returned **nothing at all** — the backup looked deleted. `--versions` showed the
truth: a 0-byte hide marker over the real 43-byte object. `b2 file unhide` then
recovered it with content intact. So the defensible claim is *"cannot delete or
read, but CAN hide"*, and a monitor that asks "is today's dump there?" cannot tell
a hidden backup from a missing one.

### Three false negatives from my own checks in one session — the pattern is now the finding

1. `pg_restore -l` as the postgres user reported a valid dump as unreadable
   (it could not read a 600 root file).
2. My bucket-inspection one-liner printed `fileLockEnabled: None` because the
   field is **top-level `isFileLockEnabled`**, not nested under
   `fileLockConfiguration.value` — the flag had worked perfectly.
3. `b2 file delete` "passed" the deny test by being **an invalid CLI subcommand**
   in v4.7 (`b2 rm` is the real one). A syntax error looked exactly like a
   permission denial, and only reading the full output instead of grepping for
   `denied|unauthorized` exposed it.

All three said "the thing is broken/absent" when the thing was fine and the
*instrument* was wrong. Combined with this morning's backwards diff, that is four.
**The through-line: I grep for the error I expect, and a check that fails for its
own reasons is indistinguishable from the thing under test failing.** The habit
that caught every one was the same — print the whole output, and verify by a
second, differently-shaped route. Worth a WRONG_CALLS entry on the pattern rather
than four on the instances.

### Object Lock verified working, including against myself

`b2 rm` on a retained file was **denied to the full admin key** (which holds
`bypassGovernance` and `deleteFiles`) and succeeded only with an explicit
`--bypass-governance`. Governance retention is real, and the escape hatch is
opt-in per command rather than implied by the capability.

### The shopfront is broken — characterised, not fixed

`[MEASURED 2026-08-10]` Owner: *"The shopfront was already broken so we will need
to fix it later as part of this thread."* Concretely:

| hostname | result |
|---|---|
| `https://webdesign.uk/` | **302 → `https://webdesign.co.uk/`** |
| `https://www.webdesign.uk/` | **302 → `https://webdesign.co.uk/`** |
| `https://preview.webdesign.uk/` | **200, 28785 B — reaches the box** |
| `POST https://webdesign.uk/api/chat` | **302 → webdesign.co.uk** |

So the box is **healthy and serving correctly**; both public hostnames are
redirected away to a different site (webdesign.co.uk, served from the bucket, not
the box). The redirect happens **in front of the tunnel** — cloudflared's ingress
maps `webdesign.uk`, `www.` and `preview.` all to `127.0.0.1:8080`, and only
`preview.` arrives. That places the cause at the Cloudflare layer (a redirect
rule / page rule on the zone), not on the box.

Consequence worth flagging to that lane: **the chat service is unreachable at its
own domain** — `POST /api/chat` never reaches the box. `HANDOFF_2026-08-10c`
attributes the owner's "nothing happened" chat test to a stale cached
`snippets.js` (`cf-cache-status: HIT`, `age: 5621`). That diagnosis may be
incomplete or superseded: a 302 on the API endpoint would produce the same
symptom and is not a caching problem. I have **not** established which came first
and am not claiming the earlier diagnosis was wrong — only that today the
endpoint 302s, and a cache fix would not address that.

**Not fixed here** (another lane's live commercial service, and a zone-level
change). Tracked in `README_where_we_are.md` as an explicit follow-through, per
the owner's instruction that we see it through in this thread. The intended route
is the framework's own checkers — `check_site_unreachable` / the experience-loop
tool checks — which is a better outcome than a one-off manual fix, because it
would have *caught* this: a site whose public hostname 302s to a different domain
is exactly what an availability check exists to find, and none is currently seated
(`site_unreachable` is code-committed but config-HELD in
`sql_for_agents/368_site_availability_driver_HOLD.sql`).

---

## 2026-08-10 — CORRECTION: the webdesign.uk 302 is INTENTIONAL

> **CORRECTED (owner, 2026-08-10): *"the 302 for webdesign.uk is intentional"*.**
> The entry above characterises `webdesign.uk` → `webdesign.co.uk` as the
> shopfront's breakage, locates its cause at the Cloudflare zone, and speculates
> that it supersedes the stale-`snippets.js` diagnosis. **The redirect is
> deliberate configuration. It is not a fault, it has no cause to find, and it
> does not supersede anything.** Everything I wrote downstream of "the box is
> healthy but its public hostnames are redirected away" was reasoning from a
> defect that does not exist.

**What I actually measured stands; what I concluded from it does not.** The
measurements were right — `webdesign.uk` and `www.` 302 to `webdesign.co.uk`,
`preview.webdesign.uk` serves the box at 200, `POST /api/chat` on the apex 302s
too. What was wrong was the inference that a redirect I did not expect must be
the breakage the owner had mentioned.

**The error is a specific and repeatable one: I had an unexplained observation
and an unlocated fault, and I joined them.** The owner said "the shopfront was
already broken"; I found something surprising; I made the surprising thing the
broken thing. Nothing in the evidence connected them — a deliberate redirect and
a real defect look identical from outside, because *intent is not observable from
a response code*. The check I never ran was the cheap one: **ask.**

This is the same failure as the backwards diff and the four false negatives
logged in `WRONG_CALLS.md`, arriving from a new direction: not a
mis-read instrument this time, but **an unexplained observation promoted to an
explanation because a vacancy existed for one.** Filed there too.

**Consequences to unpick:**
- The chat is **not** unreachable-by-defect. It is reachable at
  `preview.webdesign.uk` (verified: a live 429 from the rate limiter), and the
  apex redirect is by design.
- `HANDOFF_2026-08-10c`'s stale-`snippets.js` diagnosis is **not** superseded by
  anything I found. I should not have cast doubt on another lane's finding on the
  strength of an inference I had not checked.
- **The actual shopfront breakage remains UNCHARACTERISED.** I have not
  established what it is, and this file should not be read as though I have. The
  owner's instruction to follow it through in this thread still stands; the next
  step is to ask him what the symptom is, not to go hunting for another surprising
  observation to promote.

**Retained and still true:** no availability check is seated fleet-wide
(`site_unreachable` is code-committed, config HELD in
`sql_for_agents/368_site_availability_driver_HOLD.sql`). That remains worth doing
on its own merits — but note it would **not** have caught this, and would in fact
have raised a false positive on a deliberate redirect. A checker that cannot tell
intent from fault needs the intent recorded somewhere it can read, which is an
argument for `site_specs`/`deploy_config` carrying expected-redirect
declarations before that check is ever seated fleet-wide.

---

## 2026-08-10 — the engine is built, tested and running; public exposure is BLOCKED on DNS

### What is live on the box

`[MEASURED]` `noted-engine` — a single cross-compiled Go binary (10 MB, static),
shipped to `/usr/local/bin/noted-engine`, `systemd` unit `noted-engine.service`,
running as an unprivileged `noted` user with `ProtectSystem=strict`, an empty
`ReadWritePaths=` (all state is in Postgres, nothing on disk) and a
`@system-service` syscall filter.

- `curl 127.0.0.1:8090/api/health` → `{"status":"ok"}` — and that endpoint pings
  the database, so it reports the thing that matters rather than "the process is
  up".
- **Bound `127.0.0.1:8090` ONLY**, verified, and verified *not* reachable from
  off-box. This is the deliberate non-repetition of `webdesign-chat`'s `*:8081`.
- Schema created by the binary itself at startup (`accounts`, `sessions`,
  `notes`, `media`), so code and schema cannot drift.
- nginx site added on `127.0.0.1:8082` proxying `/api/` → 8090, with a tighter
  rate limit on `login`/`register` than on the rest.
- **The shopfront is byte-identical throughout** — box-local origin still
  `200 / 28419 B` after the nginx reload.

### The tests, and the mutation that proves they work

Seven tests, all executed against a real Postgres (they **skip loudly** without
one — a suite that silently tests nothing is worse than one that fails):

```
TestOneAccountCannotSeeAnothersNotes      PASS
TestMediaIsAccountScoped                  PASS
TestUnauthenticatedIsRefused              PASS
TestQuotaIsEnforced                       PASS
TestImportOfTheRealBackupFormat           PASS
TestLoginDoesNotRevealWhetherAnAccountExists PASS
TestPasswordHashingRoundTrips             PASS
```

`[MEASURED]` **Mutation test, because a green suite proves nothing until it has
been shown able to go red.** Changing `ListNotes`'s
`WHERE account_id=$1` to `WHERE (account_id=$1 OR TRUE)` — the exact defect the
suite exists for — produced:

```
--- FAIL: TestOneAccountCannotSeeAnothersNotes
    LEAK: bob's note list contains alice's note: {"notes":[{"id":1,"title":"Alice private","content":"her diary"...
```

Reverted; green again. The isolation guarantee is therefore tested, not asserted.

### A design decision that REVISES the earlier plan — media is in Postgres

`PLAN_2026-08-10_noted_rebuild.md` §3a-i said media would go to B2 so noted's
unbounded growth could not fill the box's 50 GB disk and take the shopfront down.
**The concern was right; the remedy is now different**, and the reasoning is in
`schema.sql`'s header:

- the coupling is closed by a **hard per-account quota enforced in the same
  transaction as the insert** (with `SELECT … FOR UPDATE`, so two concurrent
  uploads cannot both see room). A quota *bounds* growth; a different storage
  backend only *relocates* it.
- media in Postgres is media inside the nightly encrypted dump. Media in B2 needs
  its own backup story, its own credential on the box, and a restore that
  reunites two stores at one point in time — at launch volumes that is more ways
  to lose a recording, not fewer.

Migration path left open deliberately: a `storage_key` column beside `bytes`,
filled for new rows, `bytes` drained in the background. Revisit past a few GB.

### ⚠ BLOCKED, and I made a mess on the way

`cloudflared tunnel route dns <tunnel> app.noted.co.uk` **created
`app.noted.co.uk.webdesign.uk`**, not `app.noted.co.uk`. The tunnel's
`cert.pem` is authorised for the **webdesign.uk zone only**, so cloudflared
treated the hostname as relative to it and silently appended the zone. It
reported success:

```
INF Added CNAME app.noted.co.uk.webdesign.uk which will route to this tunnel
```

`[MEASURED]` The stray record is real and proxied —
`app.noted.co.uk.webdesign.uk. 284 IN A 104.21.54.51 / 172.67.223.216` — and
`app.noted.co.uk` does **not** exist. It is harmless (no ingress rule matches it,
so it falls to the tunnel's `http_status:404`) but it is clutter in someone
else's zone and **I cannot remove it**: there is no `route dns` delete, and no
Cloudflare API token is available here.

**The lesson, and it is the same one as the `--bucket` scoping earlier today:** a
tool that silently reinterprets your argument against its own authorised scope
will report success for the thing it did instead. `route dns` never errors on a
cross-zone hostname; it just makes a different record. **Check what a DNS command
created, never that it exited 0.**

**Two things the owner must do**, neither of which I can:
1. Delete the stray `app.noted.co.uk.webdesign.uk` CNAME.
2. Authorise routing for the `noted.co.uk` zone — either
   `cloudflared tunnel login` selecting that zone (browser), or a CF API token
   with DNS edit on it, after which the CNAME can be created properly.

Until then the engine is running and correct but reachable only from the box.
**noted.co.uk itself is deliberately untouched** and still serving the legacy app
with its wind-down notice — `200 / 4229 B`, re-checked after all of the above.

### Full-stack smoke test — the promise, demonstrated

`[MEASURED]` Run on the box through nginx (`box/smoke-test.sh`):

```
1. register            : session issued
   cookie flags        : HttpOnly Secure SameSite=Lax
2. save a note         : {"id":1,"title":"Shopping",...}
3. import a backup     : {"notes":1,"photos":0,"recordings":1,"skipped":0}
4. sign in AGAIN (as a different browser would) — are the notes there?
     - From my phone | recordings: 1
     - Shopping | recordings: 0
5. another account sees NOTHING of theirs:
     notes visible to the other account: 0
```

Step 4 is the whole reason for the rebuild — a note and its recording, reached
from a *second, independent session*, which the old app could never do. Step 5 is
the isolation guarantee holding against a live second account, not just in tests.

**A first attempt at this failed and the failure was the security control
working.** Over plain HTTP the follow-up calls returned `not signed in`, because
the session cookie is `Secure` and curl correctly refuses to send it over
`http://`. Real traffic is HTTPS via Cloudflare, so this only ever affects a
box-local test. The fix was to send the header explicitly — **not** to unset
`Secure`, which is exactly the shape of "fix the checker to agree with the
system" that this estate has been bitten by before. Recorded because the
tempting move was one flag away.

---

## 2026-08-11 — the migration story is much better than we thought

`[MEASURED, browser-verified]` **A different page on the same origin can read the
existing app's notes, recordings and all.** Planted a note plus a 2 KB audio blob
via `https://noted.co.uk/`, then navigated to `https://noted.co.uk/guides/about.html`
— a different document — and opened `NotedDB` cold from there:

```
{'stores': ['audio','history','images','notes'],
 'title': 'Old note', 'content': 'written before the rebuild', 'audio': 1}
```

The disconfirming result would have been an empty database or a failed open.
It did not happen.

**Why this matters more than it looks.** IndexedDB is keyed by ORIGIN, and the
origin is `https://noted.co.uk` regardless of whether the bytes come from the B2
bucket (today) or the VM (after cutover). So **at relaunch, every existing user's
notes are still sitting in their browser**, and the rebuilt site can read them
directly.

**Consequence: the manual export→import path is a FALLBACK, not the migration.**
The migration can be a page on the new site that reads `NotedDB`, shows the person
what it found, and offers to move it into their new account. Nobody needs to have
exported anything. This answers the owner's question — a "legacy notes page" is
not only possible, it is the better primary path.

**Design for it** (`/legacy`, to be built as part of the framework front end):
- opens `NotedDB` **without a version number**, so it never triggers an
  `onupgradeneeded` and never migrates or damages the old data;
- **read-only against the old stores.** It must not delete anything, ever — a
  person who does not finish signing up must still find their notes next time;
- shows counts first (notes, recordings, photos) so the person sees their data
  before being asked to do anything;
- offers two routes: sign in and copy everything up, or download the same
  full-backup file the current app produces;
- survives the old database being absent — a new visitor must not see an error.

**What this does NOT change: keep urging the backup.** The notes survive a site
change; they do not survive the person clearing browsing data, switching device or
browser, or a browser evicting storage. The notice is still right, and the reason
it is right is unchanged — *we* hold no copy.

**Do not let this finding decay into an assumption.** It was true on 2026-08-11
against the live origin. Re-run the probe above before relying on it at cutover,
because it is exactly the kind of fact that a change of hostname, a redirect to
`www.`, or a move to a different scheme would silently invalidate — origin is
scheme + host + port, and all three must stay identical.

---

## 2026-08-11 — the engine is PUBLIC, and two traps on the way

Owner authorised the `noted.co.uk` zone for the tunnel (backing up the old
`cert.pem` to `cert.pem.webdesign` first — the new cert is confirmed *different*,
so a different zone really was selected). `app.noted.co.uk` is now live over
HTTPS and the full flow works end to end from the public internet:

```
1. registered           : ok (cookie HttpOnly; Secure; SameSite=Lax)
2. save a note          : {"id":3,"title":"Shopping",...}
3. import real backup   : {"notes":1,"recordings":1,"skipped":0}
4. NEW session (a different browser) sees:
     - From my phone | recordings: 1
     - Shopping | recordings: 0
```

### TRAP 1 — `systemctl kill -s HUP cloudflared` TERMINATES it. It does not reload.

`[MEASURED]` I sent SIGHUP expecting an ingress reload, because the unit has no
`ExecReload` and a restart interrupts a tunnel that also carries the **live
webdesign.uk shopfront**. The journal:

```
Sent signal SIGHUP to main process 42569 (cloudflared) on client request.
cloudflared.service: Deactivated successfully.
```

**The tunnel went down and stayed down** — the unit has no `Restart=`, so nothing
brought it back. `systemctl start` restored it; `preview.webdesign.uk` returned to
its exact baseline (`200 / 29234 B`). **Outage ≈ 40 seconds on a live commercial
site, caused by me.**

Two things worth carrying: **a "gentler" action is only gentler if you have
checked what it does** — SIGHUP is a reload convention, not a guarantee, and this
binary treats it as shutdown. And **`cloudflared.service` here has no `Restart=`**,
so any signal that stops it is an outage until a human notices. Worth adding
`Restart=on-failure` to that unit — it is the webdesign lane's file, so raised
rather than changed.

### TRAP 2 — the Worker was eating the hostname, and the tunnel was never reached

After the restart, `https://app.noted.co.uk/api/health` still returned
**404 `text/plain` "Not found"** while:

- `cloudflared ingress validate` → **OK**
- `cloudflared ingress rule https://app.noted.co.uk/api/health` → **"Matched rule
  #3 → http://127.0.0.1:8082"**
- the origin answered locally: `Host: app.noted.co.uk` → `127.0.0.1:8082` → **200**

Everything on the box was right. **The discriminating test** was to compare the
404 against a path the Worker certainly handles:

```
https://noted.co.uk/definitely-not-a-real-file-$$   -> 404 text/plain "Not found"
https://app.noted.co.uk/definitely-not-a-real-file  -> 404 text/plain "Not found"
cmp -> IDENTICAL
```

Same handler ⇒ the **`portfolio-sites-router` Worker**, not cloudflared. Confirmed
against the API — the noted.co.uk zone carries:

```
noted.co.uk/*    -> portfolio-sites-router
*.noted.co.uk/*  -> portfolio-sites-router     <-- this caught app.
```

**Fix:** a more specific route with `script: null` ("run no worker here"), which
beats the wildcard and cannot affect `noted.co.uk` itself:

```
app.noted.co.uk/* -> (no worker)
```

Applied with the token in `~/.cloudflare/404-token.env`, which turned out to have
zone **write** scope. Verified immediately afterwards that `noted.co.uk` still
serves the app and the notice (`200`, `rebuild-notice-head` present, `js/app.js`
200) and that the shopfront was unchanged.

**The transferable form: a wildcard Worker route silently owns every subdomain of
a zone.** Adding a tunnel hostname on a B2-fronted zone will *always* hit this,
for all 36 zones on this account. The tell is a 404 whose body is byte-identical
to the Worker's, while every check *on the box* passes — and the box is where a
person naturally looks.

**And the meta-lesson, again:** three of my checks (`ingress validate`, `ingress
rule`, the local origin curl) all passed and all were true, and the thing was
still broken, because **every one of them tested the box and the fault was in
front of it.** Passing checks bound the fault; they do not locate it. What found
it was asking "what *else* returns exactly this response?"

---

## 2026-08-11 — framework build dispatched; experience patterns written first

### Dispatched

`082_submit_domain_unified.sh noted.co.uk --email hello@noted.co.uk
--mission-file MISSION_2026-08-11_noted.txt`, correlation
**`59397ca9-c1c4-4938-8d2a-e78ffd7e045b`**.

Brief validated before dispatch rather than after a failure: single-lined, no
double quotes, and **zero digits** (a number in a spec is a given and outranks
every writer-side rule). Through `claimscan` with noted's own evidence base:
**0 findings**.

`[MEASURED 17:20]` Cascade so far — `submission` and `mission_brief` written by
`domain-submitter` at 17:00:38. `needs_domain_research` sits **`triaged`** and
has not been claimed after ~20 minutes.

**That is queue depth, not a fault.** The pump is demonstrably alive fleet-wide:
`607 triaged / 3 claimed / 70 completed in the last 30 minutes`. Our item is
behind a large backlog. **Do not resubmit** — a missing row is latency, and a
retry costs a duplicate round.

Note three items dated `16:39:33`, i.e. **before** this dispatch —
`needs_composition`, `needs_design`, `evaluate_tools`, all `detected`. They are
from the discovery rotation picking the site up now that a `sites` row exists,
not from this build. `detected` is not dispatchable, so nothing acts on them.

### The experience patterns, written BEFORE the app

| pattern | contract clauses | degraded states | checks |
|---|---|---|---|
| `authenticated-note-sync` | 3 | 2 | 12 |
| `legacy-local-data-adoption` | 3 | 2 | 6 |

Every check type verified against the runner's own table — **an unsupported type
is INERT, not an error**, so a typo would produce a check that silently never
runs and a template that looks thorough. The verify query in the SQL file is
what catches it.

**Written honestly about what cannot be checked, at the check it affects:**
- `sign_in_round_trip` is **flaky by construction** — there is no
  `expect_within_ms` and the runner asserts 300 ms after the step, while a
  sign-in is a network round trip. A failure is not proof of a broken sign-in.
  Do not promote it to a gate until the runner can wait.
- `notes_survive_a_reload` **cannot be ordered** after it, so it only means
  anything when the runner happens to arrive signed in. The real assertion —
  the same notes in a second, independent session — is not expressible today and
  lives in `box/smoke-test.sh` instead.
- The legacy page's actual behaviour (a browser holding legacy data is shown the
  right counts) is **not expressible at all**: the runner cannot seed IndexedDB
  before a check. Covered by a Playwright probe, and flagged so a green criteria
  run is never read as covering it.

### A vocabulary that does not fit an app

`experience_patterns.funnel_stage` is CHECK-constrained to
`awareness|consideration|conversion`. Neither `retention` (what
`authenticated-note-sync` actually is) nor `onboarding` (what
`legacy-local-data-adoption` is) exists — the vocabulary was built for marketing
funnels, and **a product behaviour that happens after someone converts has
nowhere to sit**. Used `conversion` and said in the file that it is the least
wrong rather than the right answer. If more app-shaped patterns land, that column
needs widening; recorded rather than quietly fudged.

### Delivery path extended

`sitesync` on the box generalised from one hardcoded domain to a list, so
`noted.co.uk` is pulled from `gqls/vm-sites` alongside `webdesign.uk`. The loop
is the whole change — adding a third domain is one line, and no domain can now be
added by editing another's `rsync` and getting its `--delete` target wrong. Old
version kept at `/usr/local/bin/sitesync.bak-20260811`. Verified running as
`www-data` exactly as the timer does, with the shopfront's web root untouched
(6 files) and its served page byte-identical.

A second guard was added while generalising: **rsync into a web root that does
not exist would create it and serve an empty site**, so a missing web root now
means "this site is not set up here" rather than "make it".

---

## 2026-08-11 18:30 — why the build has not started, measured

The build has **not** progressed 90 minutes after dispatch: still 0 pages, the
`needs_domain_research` item still `triaged`, `attempt_count = 0`, never claimed.

The previous session's note (handoff §2) read this as fleet queue depth —
"the pump is alive and this is queue depth, a missing row is latency". The pump
being alive is right; the mechanism is not queue depth in the sense implied, and
the difference decides whether waiting is rational. **The reason it has not
started is the shape of the dispatch selector.**

`build-pipeline-trigger`'s `find_dispatchable_site` step (live agent config, read
from `agent_definitions`, not from a seed) ends:

```sql
ORDER BY wi.created_at ASC, wi.priority ASC, wi.id ASC LIMIT 1
```

It picks **one site at a time, globally oldest work item first**, across every
site in the estate. So this is not a per-site queue that noted is at the front
of — it is one estate-wide line, and noted's item joined it at 17:00 today.

`[MEASURED 2026-08-11 18:26]`

| quantity | value |
|---|---|
| pending items older than ours (the true backlog ahead) | **589**, across **19 sites** |
| drain rate, items older than ours completing | **~95/hour**, steady 13:00–17:00 (81/106/86/101/91) |
| implied ETA | **~6 hours** |
| trigger cadence | every ~2.5 min (`build-pipeline-trigger`, 195 runs) |

Two things that make waiting the correct action rather than merely the passive one:

- **The set ahead of us is closed to new arrivals.** Ordering is by `created_at`,
  so every work item the estate files from now on sorts *behind* ours. The 589 can
  only shrink. (Caveat, stated because it is not guaranteed: an item created
  before 17:00 that is currently `deferred`/`blocked`/`detected` — 433 of those
  exist — would join the queue *ahead* of us if it were moved to `triaged`.)
- **Resubmitting cannot help, and this is now checked rather than assumed.**
  Selection is on `created_at`; `refreshOpenWorkItemSQL()`
  (`load_work_item_actions.go:1417`) sets `updated_at = NOW()` and **never touches
  `created_at`**. So a resubmit either refreshes the existing row — no change to
  position — or files a newer row that sorts further back. It cannot advance us
  either way, and costs a duplicate round. The handoff's "DO NOT RESUBMIT" was the
  right instruction with the wrong reason; it is now the right instruction with a
  measured one.

**Also correcting handoff §2's remedy:** "if the cascade has stalled rather than
queued, the manual pump heartbeat is `076_trigger_build_pipeline.sh`". Firing that
by hand does **nothing** for noted. It runs the same trigger with the same
`ORDER BY`, so it selects the same globally-oldest site it would have selected
anyway. It is a remedy for a pump that is not firing; ours fires every 2.5
minutes. Reserve it for the case where `build-pipeline-trigger` shows no recent
rows in `orchestration_states`.

### A wrong turn on the way, recorded

I first checked `kubectl get cronjobs`, found **no build-pipeline heartbeat
CronJob** among the ten that exist, and had the mechanism half-written up as
"nothing drives this type — the pump is undriven" (the trigger script's own header
says "the CronJob heartbeat that *would* normally fire every 30 minutes", which
reads as confirmation). That was wrong. `orchestration_states` shows
`build-pipeline-trigger` with 195 runs, the most recent two minutes before I
looked — it is driven by something that is not a k8s CronJob. **An absence in the
CronJob list is not an absence of scheduling**, and the script comment describing
a CronJob is documentation, not a live fact. What caught it: querying the run
history instead of stopping at the config that should have explained it.

Second, smaller: my first drain measurement counted *eligible* items (564 → 501 →
419 in ten minutes, which looked like a ~20/min drain and a 21-minute ETA). That
number oscillates for a reason unrelated to progress — the selector excludes a
site entirely while any of its items is `claimed`, so a single claim drops that
site's whole backlog out of the count and a release puts it back. The honest
figure is the one above: pending items older than ours, which does not move when
a claim is taken.

---

## 2026-08-12 12:40 — the build landed; delivery to the box is silently NOT happening

**The queue reached us, and the ~6h estimate held.** `needs_domain_research` was
claimed and completed at **23:55:19** on 08-11 — **6h55m** after filing at 17:00,
against the `[MEASURED]` prediction of ~6h from 589 items at ~95/hour. Recorded
because the estimate was falsifiable and could have been wrong: no bypass was
fired, so the queue drained on its own and the arithmetic is what it was tested on.

The whole cascade then ran unattended overnight:

| spec / stage | by | at |
|---|---|---|
| identity, classification, content_direction, design_intent | domain-research-classifier | 23:54 |
| vertical_landscape | vertical-exemplar-researcher | 01:40 |
| strategy | domain-strategist | 02:22 |
| briefing | build-briefing-agent | 03:04 |
| site plan → pages → imagery → rerender | build-site-planner, page-build-handler, image-build-handler | 03:22–04:37 |

**5 pages, 16 components, zero empty renders**: `index` (landing), `how-it-works`,
`migrate` (landing), `about`, `contact`. `rebuild_policy` is not a `pages` column
on this schema — the decomposition check (§4.4, "no `owned` pages") still needs
doing wherever it does live.

One `needs_page` shows `failed` ("Re-render index after its image asset landed",
attempt 1 of 3). **It self-resolved** — five `page_rerender` items filed at
04:12 all completed 04:17–04:24, index included, so the artefact is current. Left
`failed` rather than retried; `failed` is not in the dispatcher's
`status IN ('triaged','approved')`, so it will never be re-picked. Harmless here,
but it is a dead row that looks like an open failure.

### The pages ARE in the repo and are NOT on the box

`[MEASURED 2026-08-12 12:35]` Framework → repo works: `gqls/vm-sites/noted.co.uk/`
holds all five files (17.9–26.4 KB) plus `assets/` and `tools/`, committed
04:19–04:37. The box's checkout is at the right commit (`15d11f095`, "Rerender:
migrate.html", branch `main`, correct remote).

`/var/www/noted.co.uk/` is **empty**, and box-local nginx returns **403 / 162 B**
for `Host: noted.co.uk`.

**Root cause, verified first-hand end to end** — the box's clone is a **cone-mode
sparse checkout whose cone contains only `webdesign.uk`**:

```
$ git -C /var/lib/sitesync/repo sparse-checkout list
webdesign.uk
$ cat .git/info/sparse-checkout        # /*  !/*/  /webdesign.uk/
$ git -C … ls-tree --name-only origin/main   # noted.co.uk IS in the tree
$ ls /var/lib/sitesync/repo/           # …but no noted.co.uk/ directory
```

So `git fetch`/`reset --hard origin/main` succeed and materialise nothing for
noted; the folder never appears; and **sitesync's own safety guard then skips it
silently**:

```bash
if [ -d "$folder" ] && [ -d "$webroot" ]; then rsync -a --delete "$folder/" "$webroot/"; fi
```

That guard is correct for the case it was written for ("the box is provisioned
before the site's first page deploys — absent is nothing to sync, not an error"),
but it **cannot distinguish "not deployed yet" from "sparse cone excludes it"**.
Result: `sitesync.service` exits **0/SUCCESS** every 5 minutes, the timer looks
healthy, and nothing is ever delivered. Verified: last run 12:33:20, `status=0`.

**Where the drift comes from.** The 2026-08-11 generalisation added
`noted.co.uk:/var/www/noted.co.uk` to `DOMAINS` in `/usr/local/bin/sitesync`. The
sparse cone is set in a *different* file that only ever runs once —
`webdesign_uk_build_service/box/setup-webdesignbox.sh:70`,
`git sparse-checkout set webdesign.uk` (idea.uk's original does the same at
`provision-pullsync.sh:119`, commented "this box fetches ONLY idea.uk/"). **Two
lists that must agree, in two scripts, one of which is never run again.** The
NOTES entry of 08-11 called the loop "the whole change" — it was not; that claim
is corrected here.

### Why fixing it is safe, checked rather than assumed

`noted.co.uk` apex is served **from B2**, not from the box: response carries
`x-amz-*` + `server: cloudflare`, and the body still contains "being refreshed"
and "Save everything" — the legacy app and its wind-down notice, intact. So
populating `/var/www/noted.co.uk` takes the box from 403 to serving the framework
build and **cannot change what any user sees**; cutover remains the separate
deliberate step (§4.5). `rsync --delete` is scoped to that web root, so
`webdesign.uk` is not a target.

Residual risk to state plainly: `sparse-checkout add` re-runs a checkout in the
repo that also feeds the **live commercial shopfront**. If `webdesign.uk`
materialised *partially*, the next `rsync --delete` would damage it. Before/after
control on the shopfront is mandatory (§8), not optional.

### Two things NOT actioned, flagged

1. **Six `unresolved_cta` items in `needs_human_review`** — and confirmed at the
   artefact, not just from the summaries: **every `hero` and `call-to-action`
   slot renders zero `<a>` anchors** (index, how-it-works, migrate; plus the
   about/contact heroes), and **no component anywhere links to
   `app.noted.co.uk`**. Body content does link internally (index
   `info-card-grid` 6 anchors, `about-content` 3, both referencing `/migrate`).
   The items' own fix note: *"No real page exists to serve as this CTA's
   destination (no eligible content hub). The gated template renders no button."*
   So the site has no call to action and no route to the product at all.
   `[INFERRED, not verified]` the resolver looks for an internal content hub and
   this product's CTA target is a different origin (`app.noted.co.uk`), which is
   why nothing resolved — that is a guess about the resolver, not a read of it.
   **Do not "fix" this by writing `rendered_html`**: a CTA fix written only there
   is invisible to the template and dies on the next content change.
2. **Shopfront byte drift, unattributed.** Handoff §8 records the box-local
   baseline as `200 / 28419 B`; it now reads `200 / 28015 B` (−404). This session
   ran read-only commands only and did not touch it; the webdesign lane updates
   that site continuously, so the likeliest explanation is a legitimate change
   and a stale baseline. Recorded, **not** joined to anything else — handoff §9
   family 3 is exactly this mistake. Whoever owns that baseline should re-pin it.

### Attribution note on commit `23f1229f0`

That commit carries a **same-file passenger**: the `mortgagecalculator_couk_adoption`
lane's in-place correction of its own `undeployed_asset` LANDMINES entry (an
expanded `source:` line plus three bullets routing to
`bugs_open/248_…undeployed_asset_repair_deploys_every_asset_as_a_hero…`). It
arrived under my commit message, which describes only the noted lane's work.

Nothing was lost and forward-only holds — their content is committed intact; the
pre-commit pattern check flagged `1 line removed from an append-only ledger`,
which on inspection was their old `source:` line superseded by their new one, not
a deleted entry. Recorded here because a pathspec commit cannot exclude a
same-file passenger (CLAUDE.md says so explicitly), my `git diff --numstat` check
ran *before* their write landed, and the commit message is therefore the only
place a bisecting reader would look and not find them.

---

## 2026-08-12 13:15 — delivery fixed and hardened; CTA destinations set

Both on the owner's explicit decisions this session ("fix and harden";
"primary CTAs → app.noted.co.uk").

### 1. Delivery: sparse cone fixed, then the drift class closed

`sudo -u www-data git -C /var/lib/sitesync/repo sparse-checkout add noted.co.uk`
(needs `GIT_SSH_COMMAND` — it is a `--filter=blob:none` clone, so newly-included
paths must fetch their blobs).

`[MEASURED]` before → after, every figure taken in the same breath:

| check | before | after |
|---|---|---|
| shopfront box-local (`Host: webdesign.uk`) | `200 / 28015` | **`200 / 28015`** |
| `webdesign.uk` webroot files | 18 | **18** |
| `webdesign.uk` checkout files | 18 | **18** |
| `noted.co.uk` webroot files | **0** | **17** |
| `noted.co.uk` box-local (`Host: noted.co.uk`) | **403 / 162** | **200 / 26402** |

26402 B is exactly `index.html`'s size in the repo. The box now serves the
framework build — `<title>Noted — Put a thought down quickly and find it again
later</title>`, 54 sections, 3 `data-component`s, and no `NotedDB` (i.e. not the
legacy app).

**The live site is untouched, verified not assumed:** `https://noted.co.uk/` still
returns the legacy app from B2 with the **same** `x-amz-version-id` as before the
change, and still contains "being refreshed" and "Save everything".
`app.noted.co.uk/api/health` still `{"status":"ok"}`. Cutover remains a separate
deliberate step.

**Then the hardening**, because the one-command fix leaves the defect in place:
`sitesync` now **derives the sparse cone from `DOMAINS` on every run** and only
calls `sparse-checkout set` when they differ (an unconditional set would re-run a
checkout in the live shopfront's clone 288 times a day for nothing). The combined
`if [ -d folder ] && [ -d webroot ]` guard is split into two, each with a **loud
warning on stderr**, because the two skips mean opposite things and the silent
version cost a day. Repo copy and box copy are identical; box backup at
`/usr/local/bin/sitesync.bak-20260812`.

Logic tested **both ways before shipping** — the no-op arm is the one that rots
unnoticed, so it was exercised explicitly: matching cone (either sort order) →
no-op; the actual bug state (`webdesign.uk` only) → sets both; empty cone → sets;
a stray extra domain → converges back to `DOMAINS`. The manual run as `www-data`
was silent and exited 0, which is now the *correct* meaning of silence.

### 2. CTA destinations — `CTA_2026-08-12_noted_cta_destinations.sql`

Read the **templates**, not the work-item summaries, for the field names
(`content_components` `23f95f00…` hero, `0197e8d7…` call-to-action): both gate each
button on `{{if and .cta_text .cta_url}}` / `{{if and .primary_cta .primary_cta_url}}`
/ `{{if and .secondary_cta .secondary_cta_url}}`. The build wrote every CTA *text*
and **no URL at all**, so all six slots rendered zero anchors.

Destinations were chosen to match the copy the framework already wrote — no
wording changed (owner ruling 2026-08-06):

| page · slot | primary | secondary |
|---|---|---|
| index · hero / call-to-action | "Sign in" → `https://app.noted.co.uk/` | "See how it works" → `/how-it-works.html` |
| how-it-works · hero | "Open Noted" → `https://app.noted.co.uk/` | "Bring your notes with you" → `/migrate.html` |
| how-it-works · call-to-action | "Sign in" → `https://app.noted.co.uk/` | "Already have notes somewhere else? …" → `/migrate.html` |
| migrate · hero / call-to-action | **"Save everything" → deliberately none** | "See how it works" → `/how-it-works.html` |

**Why migrate's primary is left unresolved.** "Save everything" is a *local-data
rescue*, not a sign-in: its true destination is the `/legacy` page that PLAN §4
step 3 has not built yet. `app.noted.co.uk` would misdescribe the action, and
`/legacy.html` would ship a 404 into a platform that actively detects unbuilt
internal links. The template's designed degraded state — render no button — is the
honest option. **Those two work items stay `needs_human_review` on purpose; set
them when `/legacy` exists.** Four items closed, two open, asserted in-SQL.

### A silent no-op I nearly shipped

My first draft closed the work items by joining on `wi.spec->>'page_id'`. **That
key does not exist** — the spec keys are `component`, `fix`, `missing`,
`page_name`, `section_name`, `source`. It would have matched **zero rows, updated
nothing, and committed successfully**, and the only visible result would have been
four items still sitting in `needs_human_review` looking like the queue's fault
rather than mine. Caught by enumerating the keys (`jsonb_object_keys`) instead of
reading the one path I expected to be there — the standing rule about a jsonb path
read being blind to the shape underneath it, which I had to apply to my own SQL.
The file now asserts `ROW_COUNT = 4` and raises otherwise: a hand-written status
change with no assertion is indistinguishable from a no-op.

### 3. Rerender filed — and a correction to my own script's warning

`scripts/initial_messages/001_assemble_all_pages_rerender/082_trigger_rerender_site_noted.sh`
(new, adapted from the gaswholesalers one). A **rerender** is correct here and a
regeneration would be actively wrong: a rerender merges `content_data`, so the
hand-set `*_cta_url` keys survive, while a regeneration replaces it and would drop
them (memory `bugfix 238`). `refresh_site_components=false` — only page bodies
changed.

Dispatch verified at the artefact, not at kcat's exit code: orchestration
`fe51e8d7-8dff-4f0a-9ab8-d2ebf49ee644` COMPLETED 13:10:49 and **5 `page_rerender`
items** are in `triaged`.

> **CORRECTED within minutes of writing it:** that script's header warns "expect
> hours, not minutes" from the estate-queue starvation. `[MEASURED 13:12]` the
> backlog ahead of these items is **3 items across 1 site** — the 589-item queue of
> yesterday has drained. The warning is right as a standing caution and wrong as a
> description of now, which is exactly why the position is worth measuring rather
> than inheriting from a doc. Left in the script as a caution, dated here.

---

## 2026-08-12 13:30 — the CTAs are live in rendered_html, and two of my claims were wrong

### `page_rerender` does NOT re-render a component from `content_data`

`[MEASURED]` After `CTA_…destinations.sql` set the URLs, I fired `rerender-pages`.
All **5 `page_rerender` items completed with `"success": true`**, each carrying a
`deploy_result` naming a commit ("Rerender: about.html", files_count 1) — and
`page_components.rendered_html` did not change **by one byte**: same `updated_at`
(04:15/04:33/04:37), still zero anchors. The repo has **no commit after 04:37**
either, because the re-assembled page was byte-identical so git had nothing to
commit and the adapter reported success anyway.

**`page_rerender` re-assembles a page from its EXISTING component HTML.** It does
not re-render a component from `content_data`. This corrects the assumption behind
the 082 trigger I wrote an hour earlier, and it is a textbook case of "a `complete`
work item is not a repaired artefact" — five of them, all green, all inert.

The action that *does* re-render from `content_data` is the **section editor's
`content_edit`** (`section_editor_actions.go:215`): update `content_data` →
re-render the component template with site context → `UPDATE
page_components.rendered_html` → reassemble the page → commit. New trigger:
`scripts/initial_messages/130_section_editor/074_section_editor_noted_cta_urls.sh`,
sequential by design (two slots on one page would otherwise race two page
assemblies and two commits for the same file), canaried on one component before
firing all six because the spawn→call handshake on that path is known-flaky.

Result, verified at the artefact — all six re-rendered 13:22–13:27:

| page · slot | anchors | destinations |
|---|---|---|
| index · hero, call-to-action | 2, 2 | `app.noted.co.uk` + `/how-it-works.html` |
| how-it-works · hero, call-to-action | 2, 2 | `app.noted.co.uk` + `/migrate.html` |
| migrate · hero | 2 | **`/contact.html`** + `/how-it-works.html` |
| migrate · call-to-action | 1 | `/how-it-works.html` |

### "Leave `cta_url` unset and the template renders no button" — WRONG, and it
### produced exactly the misleading thing I was trying to avoid

`migrate · hero` now renders:

```html
<a href="/contact.html" class="btn btn-primary">Save everything</a>
```

**I never set that.** `content_data` for that component still has **no `cta_url`**
— my deliberate omission persisted exactly as written. The `/contact.html` is
supplied by a **RENDER-TIME resolver**, so the gate `{{if and .cta_text .cta_url}}`
is satisfied by a value that never appears in the source of truth. My reasoning in
`CTA_…destinations.sql`'s header — "the template's designed degraded state is
render no button, which is honest" — **is false**, and it failed in the one
direction I said I was protecting against: a button labelled "Save everything",
the action that rescues someone's only copy of their notes, pointing at a contact
form.

Note the earlier evidence pointed the other way and I over-read it: at the 04:15
build the same component rendered **zero** anchors and the work item said "no
eligible content hub". So the resolver's behaviour differs between the build path
and the section-editor render path (or its eligibility changed once pages existed)
— `[UNMEASURED]` which, and I am not asserting a cause without reading it.

**Consequence for the fix candidates:** "leave it absent" is not available as a way
to suppress a CTA on this platform. Suppressing one requires either removing the
`cta_text` (a content change, and content is the framework's job) or setting an
explicit destination. Raised with the owner rather than guessed, because the
product answer is "this button needs `/legacy`", and `/legacy` is PLAN §4 step 3.

**No live impact:** noted.co.uk still serves the legacy app from B2, so nothing
user-facing carries the wrong link. It must not survive cutover.

---

## 2026-08-12 14:05 — /legacy is BUILT: the rescue tool exists, tested against real data

Owner decision: build `/legacy` (PLAN §4 step 3) and point migrate's button at it.

### The facts I read before writing a line of it

Every one of these came from source, not from the design doc, and two of them would
have caused silent data loss if guessed:

- **Legacy schema** (`gqls/sites` `noted.co.uk/js/storage.js`): `NotedDB` v4 —
  `notes` (keyPath `id`), `history` (keyPath `revId`, autoIncrement, index
  `noteId`), `audio` and `images` **both keyed by `noteId`**, one record per note.
- **⚠ `audio`/`images` have TWO record shapes.** Current is
  `{noteId, items:[Blob]}`; an earlier one is `{noteId, blob:Blob}`, and
  `Storage._getMediaArray` handles both (`storage.js:55`). **A reader that knows
  only the current shape silently drops every recording saved by the older version
  and looks perfectly healthy.** I would not have invented this.
- **Export shape** (`js/app.js downloadFullBackup`): `format:
  "noted.co.uk/full-backup"`, `version: 1`, `exportedAt`, **whole note records**
  (not id/title/content), plus `history`/`audio`/`images` maps keyed by note id
  with media inline as base64 data URLs; filename
  `noted-full-backup-YYYY-MM-DD.json`. The engine's `backupFile` (server.go:321)
  decodes a subset of this and ignores `history` — so the file is a superset by
  design and both halves are fixed by what is already on people's disks.

### The tool

`docs/.../noted_rebuild/legacy_tool/noted-legacy-rescue.html` — markup + inline JS.
Three invariants, stated in its `tool-doc` header because this touches what is for
many people the only copy of their notes:

1. **Open with no version number** — naming one runs `onupgradeneeded`, a schema
   migration against that only copy.
2. **Never write, never delete** — every transaction is `readonly`; there is no
   `deleteDatabase` anywhere.
3. **Leave nothing behind** — opening a database that does not exist CREATES one,
   so when `onupgradeneeded` fires we **abort the versionchange transaction**.

### The probe, and the mutations that prove it can fail

`legacy_tool/test_legacy_rescue.py` — Playwright, 24 checks, three cases (no
database; a real database with notes + both media shapes + history; a same-named
database without a `notes` store). This exists because **the platform's check runner
cannot seed IndexedDB** (HANDOFF §5), so it is the only thing that exercises this
code against data.

**All 24 passed — which on its own is worth nothing, so both load-bearing guards
were mutation-tested:**

| mutation | result |
|---|---|
| remove the `abort()` in `onupgradeneeded` | **FAIL** "LEAVES NO DATABASE BEHIND — databases now: ['NotedDB']" |
| make `mediaItems()` ignore the legacy `{blob}` shape | **FAIL** ×2: recording count 2 not 3, and `def-456` missing from the audio map |

Both then passed again on restore. The stray-database one is the useful surprise:
it confirms that opening a non-existent `NotedDB` really does create it, so guard 3
is load-bearing rather than defensive decoration.

### Into the framework, and what it refused first

`scripts/initial_messages/140_tool_suggester/075_create_noted_legacy_rescue_tool.sh`
dispatches `create_tool_component` — the framework path, not a hand-built page
(owner ruling 2026-08-04), and it canonicalises the page identity, which is the
whole point of `bugs_open/080`.

**First attempt REFUSED**, correctly: *"generated tool HTML lacks the tool-doc
header"*. Every new tool's `<script>` must open with a sentinel-delimited block
(`/* === tool-doc ===` … `=== /tool-doc === */`, `platform/content/tool_doc_header.go`)
carrying purpose + behavioural invariants; it is read by the tool auditor and
**stripped at deploy** so it never ships. Nothing was created by the failed run.
Added it in the house style and it went through.

Also worth recording: the first dispatch never reached Kafka at all because an
**unquoted heredoc** let bash try to execute the Python. Fixed by passing values
through the environment and quoting the heredoc (`<<'PYEOF'`).

### Result

| thing | value |
|---|---|
| component | `tool-legacy-rescue`, `component_level='tool'`, 12,526 chars incl. the JS |
| page | **`/tools/legacy-rescue/index.html`** — the framework's canonical shape |
| companion guide | `/guides/tool-legacy-rescue-guide.html`, `needs_content_page` queued |
| migrate's CTAs | now render **both** buttons, "Save everything" → the tool |

**The URL is NOT `/legacy.html`.** `CanonicalisePage(role="tool")` yields
`/tools/<bare>/index.html`, and nested rather than flat because `siteUsesFlatURLs`
reads `site_specs` aspect `structure` — which **this site does not have** — and
defaults to nested. So noted's content pages are flat (`/about.html`) while its
tool page is nested. That is the framework's decision and I took it rather than
hand-roll an identity; recorded because it looks like an inconsistency and is not
one to "tidy".

### Still open

- **The tool page is `build_status='planned'`** — its component is rendered and
  `deployed`, but the page has not been assembled and committed to `gqls/vm-sites`
  yet, so `/tools/legacy-rescue/index.html` **404s today**. `page_rerender` will not
  pick it up: that path files one item per page where `build_status='deployed'`.
  `[UNMEASURED]` which mechanism promotes a `planned` tool page — find it before
  assuming the queue will do it on its own.
- The probe tests the file, **not the deployed page**. Re-run it against the real
  URL once the page ships, and re-run the origin probe at cutover (§6) — origin is
  scheme+host+port, so a `www.` redirect silently invalidates all of this.
- The "sign in and copy everything up" route is **not** built; the download route
  is, and it is the one §6 makes mandatory because it needs no account.

---

## 2026-08-12 17:10 — a hand-filed work item deployed the WRONG PAGE and reported success

A fresh chassis build rolled at 14:55 (pods 0 restarts, well past the ~300s
no-dispatch window by the time anything was sent). It changed nothing here: both new
pages were still `planned` and the guide still `needs_human_review` at 16:16.

To promote the tool page I hand-filed a `page_rerender` item, building it by copying
an existing row's whole shape — the trick I'd just used to avoid guessing NOT NULL
columns (`source`, `created_by`, … discovered one constraint violation at a time,
which is the wrong method and cost three round trips).

**`site_work_items` has a real `page_id` COLUMN** (plus `component_id`, and
`idx_swi_page` over it). My template row was the **contact** page's. I overrode
`spec` — including `spec->>'page_id'` — and never touched the column. The handler
reads the **column**.

`[MEASURED]` The item ran, reported `"success": true`, and deployed:

```
"files": ["/contact.html", "/tools/assets/contact-form.js"], "file_path": "/contact.html"
```

**Nothing anywhere said the target was wrong.** The summary said "Assemble and deploy
the legacy rescue tool page". The spec said `page_name: tool-legacy-rescue`. The
status said `complete`. Only the deploy result named the file, and only the join
`LEFT JOIN pages p ON p.id = wi.page_id` shows it plainly:

| item | page the COLUMN points at |
|---|---|
| `…_assemble` (16:18) | **contact** |
| `…_assemble_v2` (17:07) | tool-legacy-rescue |

No damage: contact.html re-deployed byte-identically. On a destructive item type the
same mistake would not have been free.

**The general shape, which is the bit worth keeping:** copying a row to inherit its
column shape also inherits its *targeting*. The fields that make a work item point
at something are not all in `spec` — and a spec-only override produces an item that
reads correctly to every human check and executes against a different object. Reset
the FK columns explicitly, and assert the target with a join before believing the
summary.

This is the third time this lane has been caught by a `complete` item that did not
do what it said: five `page_rerender`s that changed zero bytes, the CTA re-render
that needed the section editor instead, and now this. **`complete` is a statement
about the runner, not about the artefact.**

Filed `…_assemble_v2` at 17:07 with the column set; outcome pending at the time of
writing. Handoff `HANDOFF_2026-08-12_continue_here.md` §2 and §6 carry both the
state and the check.

### 17:12 — it worked, and the fix confirms the diagnosis

`…_assemble_v2` completed **17:10:09** and `tool-legacy-rescue` flipped to
`deployed`. Same item type, same shape, one field different — the `page_id`
**column** — so this is a confirmation of the cause rather than a workaround that
happened to succeed.

`[MEASURED 17:12]`, at the artefact: the page is in `gqls/vm-sites`, landed on the
box at **17:11:46** (one sitesync tick after the commit — the delivery chain fixed
this morning working unattended), serves **`200 / 23425`** box-locally, contains the
rescue markup (12) and `indexedDB.open` (1), and `migrate.html` links to it twice.
Shopfront `200 / 28075` unharmed; the live apex still serves the legacy app.

**The `tool-doc` header is absent from the served page (0 occurrences) while present
in the DB template** — the strip-at-deploy half of the contract, working. Worth
recording because it is the kind of thing that looks like a bug from either end if
you only check one.

**Deviation from PLAN §4.2, stated rather than quietly accepted:** the plan says a
tool's JS deploys as `/tools/assets/{fn}.js`. It did not — `tools/assets/` holds only
`contact-form.js` and the rescue JS is inline in the page. Cause is mine: I passed
the JS inside `html_content`, so `content_components.js_content` is empty and the
asset emitter had nothing to extract. Works as served; if the separate asset is
wanted, the JS belongs in `js_content`.

Shopfront baseline drift, again unattributed: `28419` (previous handoff) → `28015`
(this morning) → `28075` (now). Read-only session throughout; that lane ships
continuously. This is why the handoff says to take the baseline yourself immediately
before touching anything rather than trusting a recorded figure.

---

## 2026-08-12 — privacy copy DRAFTED at the owner's direct request

This copy was reserved to the owner (`evidence_base.writer_block`: *"that copy is
the owner's and will be supplied"*), and a 2026-08-10 proposal was rejected and
scrubbed. **He asked for it directly today, which supersedes the reservation** — a
delegation is his to make. The scrub did its job: I could not see the rejected
wording, so this is not a reheat of it. The surviving guidance is the 08-11 steer —
**avoid where anything is stored, focus on what the site does for people.**

Draft: `COPY_2026-08-12_privacy_DRAFT_for_owner.md`. **Not published.** No page, no
seed, `evidence_base` untouched.

**The `writer_block` and the owner's steer pull against each other** — the block
says *describe the mechanism plainly (where something is stored…)*, the steer says
*avoid where anything is stored*. Threaded by describing what a person
**experiences** ("there when you sign in on a computer you have borrowed for ten
minutes") and letting the reader draw the conclusion, which is what the block
actually asks for. Recorded because the next writer will hit the same tension.

`[MEASURED]` Checked with `COPY_2026-08-12_privacy_check.py`: **clean against all 7
live patterns**, no `writer_block` style words, no figures. Three properties make it
a check rather than a gesture:

1. patterns read from the **live DB**, not copied into the script — a ban added
   tomorrow is enforced tomorrow with no edit;
2. it tests **only** the section after `## THE DRAFT`, because the commentary
   deliberately quotes banned phrases while explaining why they are banned, and a
   whole-file checker would fail on its own explanation;
3. a **positive control** runs first — the old site's real sentence, which MUST be
   caught. It was, by pattern 1. A zero from a checker that has never fired is not
   evidence, and this file exists partly so that stays true after edits.

**Three commitments are flagged at the top of the draft** rather than buried in it —
don't sell, don't advertise, won't train on it, plus "we will say what changed". They
are promises about future behaviour that only the owner can make, so they are listed
where he must consciously accept or strike them.

**One question I refused to answer for him:** deletion vs backups. The draft says
deleting takes a note out of your account, which is true; it does not mention that
encrypted nightly backups carry a 30-day object lock, so a deleted note is not gone
from every copy immediately. How much to disclose is a product decision. Written up
as an open question, not silently papered over — the alternative was a sentence that
reads as a clean-erasure promise this product cannot honour, which is the exact
class of claim the `evidence_base` exists to stop.

### Owner approved the copy with one edit; it is now REGISTERED, and the page is not

Owner: *"just change the word 'plainly' to something else otherwise go ahead."*
Changed to **"we will spell it out"** — concrete, and it adds detail rather than
restating the promise. Re-checked: still clean against all 7 bans, control still
fires.

**Registered in `evidence_base`** (`apply_privacy_copy.py`), which is the step that
turns *"that copy is the owner's and will be supplied"* into supplied. Two
properties of that script are the point:

- the copy is **extracted from the draft file**, so doc and database cannot drift;
- the new spec row is **derived from the live one** (`data || {...}`), so the seven
  `banned_claims`, `facts`, `governing_rule`, `audit_doc` and `schema_notes` carry
  across untouched. **Retyping the blob is how a ban silently disappears**, and
  nothing downstream would report it.

`[MEASURED]` after the write: **7 bans still present**; `supplied_copy.privacy`
body **1582 chars, exactly matching the local draft**; `writer_block` now says the
copy HAS BEEN SUPPLIED and must be used **verbatim**, while keeping the prohibition
on inventing privacy wording anywhere else.

Two guards fired on the way and both were right:

1. The first run refused because `"spell it out"` was not found — the markdown is
   hard-wrapped, so the owner's edit sits as `"spell it\nout"`. The fix was to
   normalise whitespace **in the guard**, not to drop the check.
2. Python `%`-formatting broke on the SQL's `RAISE ... %` placeholders. Switched to
   an explicit token replace.

### The page itself is NOT created, deliberately

`[UNMEASURED]` there is no framework path I could find that adds ONE content page on
demand. What exists: `build-site-planner` (created all five pages from a plan, and a
re-run's blast radius on those five is unknown to me), `create_tool_component`
(tool pages only, and forces the `tool-` prefix), `create_report_page` (forces
`rebuild_policy='owned'`, which this lane forbids), and `needs_content_page` (writes
content for a page that **already exists**). No `site_plan` or `structure` spec is
stored for this site, so there is no planned-but-unbuilt privacy page to promote.
The only discovery check that mentions privacy is `check_misdirected_cta`, which
merely recognises the name.

So the two available routes both have a real cost: re-run the planner and risk five
working pages, or hand-create the `pages` row — **which is precisely the
hand-rolled-identity mistake `bugs_open/080` is about and which I warned about in
this same session when the tool page landed at `/tools/legacy-rescue/index.html`
rather than `/legacy.html`.** Doing it now because the session is nearly over would
be the shortcut this lane's docs exist to prevent. Left for the owner to choose.

> ## ⚠ CORRECTED 2026-08-16 by the `dartsonline_traffic` lane — THERE IS A THIRD ROUTE, and it is the framework's own
>
> **Contributing rather than rewriting: the section above was honest and its `[UNMEASURED]`
> marker was correct.** What follows is the measurement it was waiting for. Added here because
> this section reads as a settled dead end to anyone who arrives later — and it did exactly
> that to me. I repeated *"there is no framework path that adds one content page on demand"*
> to the owner as established fact, **dropping the `[UNMEASURED]` marker in the process**,
> and wrote it into a handoff. That is logged in `WRONG_CALLS.md` (2026-08-16) as my error,
> not this lane's: the marker did its job and I laundered it away.
>
> **The route: `apply_gap_plan`'s `new_page` approach.** `[MEASURED] 2026-08-16`
>
> - `platform/orchestration/actions/apply_gap_plan_action.go:8` — its own header lists the
>   approaches, including *"`new_page`: creates page record + needs_content_page work item"*.
>   Branch at `:140`, implementation `applyNewPage` at `:289`, plus a dedicated test file
>   `apply_gap_plan_new_page_test.go`.
> - **This resolves the section's fourth bullet.** *"`needs_content_page` (writes content for
>   a page that already exists)"* is exactly right about that item type **in isolation** —
>   and `new_page` is the step that creates the page and *then* files it. The two compose;
>   the chain is not blocked by the constraint on its second half. That single reading is the
>   whole of the difference between this section's conclusion and the working route.
> - It runs in the live **`content-gap-planner`** agent, fed by `needs_content_planning`
>   work items which are **actively draining — 27 `complete`, newest 2026-08-15** `[MEASURED]`.
> - **Proven in production**, on a different site in this same estate:
>   `docs024_key_docs_latest/dartsonline_traffic/SQL_2026-07-29n_news_page.sql` raised a
>   planner item with `'approach','new_page'` asserted in the spec; `dartsonline.com/news/index.html`
>   serves **200** today.
>
> **Two caveats, so this correction does not become the next over-claim:**
> 1. `approach` is read from the **plan the LLM produces** (`apply_gap_plan_action.go:127`),
>    so the planner may still choose `add_to_page` or `not_actionable`. Asserting it in the
>    item spec is a strong steer, not a guarantee — check which branch actually ran.
> 2. I have **not** re-checked this against noted.co.uk's specifics. This lane's note that
>    *"no `site_plan` or `structure` spec is stored for this site"* may still bear on whether
>    the planner can act here — that is the thing to verify before using this route on noted,
>    and it is why nothing on this site has been touched by me.
>
> **So the owner's choice framed above is now a choice of three, and the third does not carry
> either of the costs that made the other two unattractive.**
>
> **⚠ And this lane had ALREADY caught part of it — two hours later, further down this very
> file.** See *"First, a correction to what I wrote two hours ago"* below: the site plan is a
> set of TABLES (`site_plans`, `site_plan_pages`, …), not a `site_specs` aspect, noted **did**
> have a current plan all along, and that session wrote in terms as plain as mine —
> *"That error is why I told the owner there was no framework path to add one content page.
> There was; I was looking in the wrong store."*
>
> **So there were two independent routes and this section was wrong about both** — and the
> self-correction, though prompt and honest, sits ~65 lines below the claim under a heading
> that does not mention pages. I read the claim, did not read on, and repeated it. **That is
> the reusable lesson for everyone, not just me: a correction is only load-bearing if it is
> attached to the thing it corrects.** Hence this block being placed here rather than
> appended at the end of the file, which is where the append-only convention would otherwise
> have put it.

**What is true today:** the wording is approved, checked, and canonical in the
framework — any agent writing a page that needs it is now instructed to use it
verbatim and forbidden from inventing a rival version. That was the blocking half.

---

## 2026-08-12 18:45 — structure spec + site plan backfilled, and it produced the build path I said did not exist

### First, a correction to what I wrote two hours ago

I recorded that "no `site_plan` or `structure` spec is stored for this site". Half
right, and the wrong half mattered: **the site plan is not a spec aspect, it is a
set of TABLES** — `site_plans`, `site_plan_pages`, `site_plan_sections`,
`site_plan_imagery`, `site_plan_directives`. I had queried `site_specs` aspects and
concluded from an absence there. noted **did** have a current plan all along
(`185149a7…`, `build-site-planner`, 03:22) listing the five content pages.

That error is why I told the owner there was no framework path to add one content
page. There was; I was looking in the wrong store.

### What was actually missing

| gap | consequence, measured |
|---|---|
| no `structure` spec | `siteUsesFlatURLs` treats absent spec / absent key / any non-`"flat"` value all as nested, so this site's URL shape **was never a decision**. It bit us the same day: the rescue tool canonicalised to `/tools/legacy-rescue/index.html` while every content page is flat |
| plan missing 2 live pages | `create_tool_component` wrote `tool-legacy-rescue` and its guide **directly**, outside the plan |
| no privacy page anywhere | the owner-approved copy had nowhere to live |

### The backfill records reality; it does not change it

`BACKFILL_2026-08-12_structure_spec_and_site_plan.sql`. `url_shape` is written as
**`"nested"`** — what the site already does — because **a backfill that moves a live
URL is not a backfill.** `site_url_shape.go` carries the measured warning for why
that restraint matters: *"upsertPage overwrites pages.url unconditionally and the
deployer takes the file path from it. Measured on loancalculator.co.uk 2026-08-10:
24 of 26 live URLs would have moved."*

Blast radius here is enumerated rather than asserted: `CanonicalisePage` routes only
**tool/guide/game** roles through `nestedOrFlatURL`; content, landing and blog-post
are flat on every shape. noted has exactly one tool page and its guide is
`role=blog-post`, so flipping to `"flat"` would move **one** page. That migration is
costed at the foot of the SQL file, unapplied, with the CTA re-point and the orphan
cleanup it would require — and the note that it is cheapest **now**, while the build
is not public.

Applied clean: 8 planned pages, `url_shape=nested`, and an assertion that the tool
page URL did not move.

### Then the reconcile — the path that was there the whole time

`reconcile_site_plan` (pure read-decide-write, **no LLM**) diffs `site_plan_pages`
against realised pages and emits `needs_page` for the delta. It lives inside
`build-site-planner`, which re-derives the plan first — so I called the action alone
via an inline workflow (`081_reconcile_plan_noted.sh`) rather than re-planning five
working pages to add one.

`[MEASURED]` `pages_total 8, pages_emitted 2, pages_skipped_built 6,
pages_review_emitted 0, rerender_emitted true` → `needs_page` for **privacy** and
for **tool-legacy-rescue-guide**, plus the terminal `needs_rerender`.

**And my prediction was wrong in an instructive way.** I expected
`tool-legacy-rescue` to raise an `owned_page_review` (rule 3, role=tool, the guard
against the generic builder clobbering an owned page). It did not: **rule 1 wins
first** — deployed at the current plan version, so it is skipped as built before the
role is even considered. Rule 3 fires on a tool page that is missing or not
deployed, not on one that is fine. The script's comment now records the measurement
instead of my guess, because the guess would have sent the next session hunting a
human-review item that is never coming.

### Also settled: the retraction risk I flagged was NOT real

I wrote in the backfill header that a plan which does not list a live page is a
retraction risk, marked `[INFERRED, not measured]`. Reading
`reconcile_site_plan_action.go` settles it: **every decision is per PLAN page.** A
page that exists but is absent from the plan is never considered, so it cannot be
retired by this path. The inference was wrong and the marker is why it was cheap.
Recording the two tool pages in the plan is still right — it is what makes them
visible to the diff — but not for the reason I gave.

### 18:48–19:05 — the reconcile emitted two builds; BOTH failed, in different ways, and my part-1 backfill was incomplete

**`privacy` — `complete` work item, no page, total silence.** `[MEASURED]` the
build reached `complete_error` at 18:48:32 with **no `__step_error` at all**, no
`pages` row, nothing in `agent_error_log` — and the `needs_page` item was marked
**`complete`**. Cause: **part 1 of my backfill added the page to
`site_plan_pages` and not to `site_plan_sections`.** A planned page needs BOTH.
Without sections the handler reaches `check_has_ready_sections` and stops, before
the validator that would have logged anything. Fixed in
`BACKFILL_2026-08-12b_privacy_page_sections.sql` (hero + generic-text-block,
component names taken from this site's existing plan rather than invented).

That failure signature is worth keeping: **a `complete` item, no page, and silence
in every log.** It looks nothing like a validation failure, and "complete" is once
again a statement about the runner.

**`tool-legacy-rescue-guide` — a REAL blocker, and my inference about it was
wrong.** I had told the owner the guide was very likely blocked on the missing
privacy copy. It was not. `agent_error_log` (which is where
`validate_page_content` persists its issue list, precisely so post-mortems do not
need pod logs) names it exactly:

```
value:    "no server"
location: "The old version of Noted had no account and no server. Every note you
           wrote there stayed in that one browser, on"
severity: blocker
```

The writer described the **old** app **truthfully** — it genuinely had no server —
and the ban `(no|zero|without a)[ -]?(server|servers|cloud|backend)` cannot see
tense or subject. **A migration guide's whole job is to describe the old
architecture, so on this one page the ban fires on a true sentence.** That is the
cost of a regex ban, not a defect in the writer.

Registering the approved copy did **not** unblock it, and would not have: the
blocker was never about privacy wording. Corrected here and in the handoff.

**Two things I had also missed and should record:**

- `validate_page_content` runs banned-claim patterns **FLEET-WIDE plus per-site**
  (`datahelpers/claims_global.go`, `bugs_open/104`), on every site whether or not
  it has an `evidence_base`. My `COPY_2026-08-12_privacy_check.py` tests only
  noted's **seven per-site** patterns, so a clean run there is necessary and
  **not sufficient**. The check is still worth having; its scope now needs saying
  out loud.
- The blocker detail is in **`agent_error_log`**, keyed by `domain` + `action`,
  not in `orchestration_states`. I spent several queries hunting `collected_data`
  for something the platform had already written down in the place designed for it.

### 19:06 — the sections were NOT the cause either. The real chain, mapped

Second attempt, after part 2 added the section plan: **same signature** — item
`complete` at 19:07, no `pages` row, `complete_error`, no `__step_error`.

`[MEASURED]` the branch that decides it:

```
check_page_found = {"condition_met": false, "next_step_override": "complete_error"}
```

**`page-build-handler` LOADS a page record. It never creates one.** A plan page
with no `pages` row cannot be built by it, and it says so only through a condition
flag — no error, no log, and the work item still finishes `complete`.

So the chain is longer than reconcile:

```
site_plan_pages + site_plan_sections   (the plan — what I backfilled)
        │
        ▼  sync_pages_to_db  ← THE STEP I SKIPPED; creates the `pages` rows
        │                       (SyncPagesToDBAction, site_db_actions.go:218;
        │                        run by build-site-planner, pageflow-builder,
        │                        site-work-orchestrator)
        ▼  reconcile_site_plan  → emits needs_page for the delta
        ▼  page-build-handler   → requires the pages row to already exist
```

`reconcile_site_plan` emits `needs_page` for a plan page with no `pages` row, and
the handler then refuses it for exactly that reason. **The two disagree about whose
job it is to create the row**, and the disagreement is silent. That is worth a
`bugs_open/` entry on its own; not filed here because this lane should not assert a
platform-wide claim without the `090` loop (CLAUDE.md), and the claim is
cross-cutting.

**Why I stopped rather than pushing on.** `sync_pages_to_db` reads its pages from
`page_plan`/`site_plan` **in collected_data**, not from `site_plan_pages` — it is
built to run immediately after the planner, with the plan still in memory. Calling
it standalone means synthesising that payload. It is an **upsert loop, no DELETE**
(checked — a one-page payload would touch only that page and leave the other seven),
so it is not dangerous, but constructing the input shape is guesswork and this
session has already spent two wrong attempts on this page. The honest next step is
`pageflow-builder` or `site-work-orchestrator` — both already run `sync_pages_to_db`
— read one of their workflows and drive the whole plan→pages→build sequence the way
it was designed, instead of hand-assembling the middle of it.

**Corrected, again:** part 2's header says the missing sections were the cause of
the 18:48 failure. They were **a** gap — a planned page does need both tables — but
they were not the cause. The cause at 18:48 and at 19:06 is the same missing
`pages` row.

---

## 2026-08-13 — the owner's question answered: how the two real routes create `pages` rows, and the adoption route is reproducible here

Owner declined `pageflow-builder` and asked what routes `domain-submitter` and the
adoption orchestrator use. Read from source, not recalled:

| route | how `pages` rows are born |
|---|---|
| **domain-submitter** → … → `build-site-planner` | `plan_site` (LLM) → `write_site_plan` → **`sync_pages` (`sync_pages_to_db`)** → `populate_nav` → `reconcile_site_plan`. The sync runs with the plan **still in collected data**, which is why the action reads memory, not tables, and cannot be cleanly called standalone |
| **adoption** (`site-adoption-agent`, `apply_adoption_plan_action.go:541`) | **plain SQL, itself**: `INSERT INTO pages (…) VALUES (…, 'planned', …) ON CONFLICT (site_id, name) DO UPDATE …` — no `sync_pages_to_db` anywhere, then one relay item |

**This reframes yesterday's caution.** I had treated a hand-created `pages` row as
"the bugs_open/080 mistake". Reading adoption shows the platform itself creates
pages by direct upsert; 080's defect is hand-rolling the **identity**, not the
insert. Our identity was canonicalised when the plan row was written
(`role=content` → `/privacy.html` on any url_shape).

So: mirrored adoption's upsert for the one page, **every value read from the
current plan** (`site_plan_pages` for name/url/title/role/meta/nav flags;
`site_plan_sections` in plan order for `pages.sections`, which is rebuild
membership). `ON CONFLICT (site_id, name)` — the same conflict target adoption
relies on — makes it idempotent. Verified:
`/privacy.html · content · planned · ["hero","generic-text-block"] · footer-only`.

Reconcile re-run (corr `829f3e7a…`). If the handler builds it this time, the chain
mapped at 19:06 yesterday is confirmed end to end: plan tables + pages row →
reconcile emits → handler builds. Poll running.

### 10:21–10:30 — the page BUILT, the writer ignored the copy, and both halves are now fixed

**The chain is confirmed end to end.** With the `pages` row present (adoption-style
upsert), reconcile emitted `needs_page`, the handler built it, and
`/privacy.html` went `deployed` at 10:21. So yesterday's map holds: plan tables +
pages row → reconcile → handler.

**Then the verbatim check failed completely: 0 of 22 sentences.** The writer wrote
its own privacy prose — honest, but not the owner's, and it names where things are
stored, the exact thing his steer avoided.

Root cause, read from the live prompt template, not guessed:
`page-content-writer` injects **only the `writer_block` STRING**
(`{{.site_specs.specs.evidence_base.writer_block}}`). My 08-12 registration told
the writer to use copy "under `supplied_copy.privacy`" — **a JSON path that never
travels to the prompt.** The writer was instructed to use copy it was never given.
An instruction pointing at data outside the reader's context is wired to nothing —
the same class as "a doc comment is not an enforcement mechanism", and I wrote one
the day after quoting that rule.

**Fix, both halves:**

1. **Immediate** — `074b_section_editor_noted_privacy_copy.sh`: section-editor
   `content_edit` sets hero subheadline (= the copy's intro) and the text block
   (= the rest, `<strong>` lead-ins, mailto link; heading "What that means in
   practice", a fragment of the copy's own sentence, because the template's
   `<h2>` is unconditional and an empty one draws audits). The copy is EXTRACTED
   FROM THE DRAFT at run time — no second copy to drift. `[MEASURED]` **22/22
   sentences verbatim** in rendered_html; on the box at 10:29 (`200/18060`,
   "spell it out" present). ⚠ Re-run 074b after any REGENERATION of the page —
   rerender merges, regeneration replaces (bugfix 238).
2. **Durable** — `embed_privacy_copy_in_writer_block.py`: the copy now travels
   **inline in writer_block itself**, so any future regeneration has the text in
   its prompt, not a pointer. Verified: 7 bans intact, copy present in the block,
   `supplied_copy` kept as the canonical store.

Controls at the end: shopfront `200/28286` (that lane ships continuously — take
baselines fresh), apex still the legacy app. Nothing here is public.

**What remains open on this thread:** the guide (`tool-legacy-rescue-guide`) is
still `needs_human_review` on the true-sentence "no server" blocker — the ban is
the owner's to narrow or not. And the first `unresolved_cta` pair on the guide
page will resolve when its content builds.

### Owner ruling: "we don't want to say no server" — the ban stays; the INSTRUCTION was the defect

The decision was already implicit in the ruling: nothing gets narrowed. What
remained was mechanical, plus one finding worth the entry on its own:

**The writer's instructions modelled the banned phrase.** `writer_block` said *"The
old site had no server at all"* — so a writer describing the old version in the
migration guide echoed its own instructions' framing and was blocked by the
validator. The gate and the guidance disagreed, and the writer obeyed the guidance.
The general shape: **a rule that quotes the forbidden phrasing while forbidding it
teaches the phrasing** — prompt text is training data for the very model reading it
(kin to memory `prompt-text-poisons-its-own-detector`, different failure surface:
there it poisoned the detector, here it poisoned the writer).

Fixed by `reword_old_app_instruction.py` (exact-sentinel replace, derived row):
the clause now reads "describe it by what it DID: the old Noted kept everything in
the browser you wrote it in, on that one device, and nowhere else. Never describe
it by what it lacked…" `[MEASURED]` after the write: 7 bans intact, approved copy
still inline, and the banned-shape regex finds **zero** matches anywhere in the
new writer_block — the check that could have come out otherwise.

Guide item requeued (`needs_human_review` → `triaged`) with the ruling recorded in
its spec. Poll running. If the writer still produces a banned shape, the validator
will catch it again and the fallback is a section-editor `content_edit` on the
offending sentence — not a ban change.

### 16:50–22:07 — guide rebuilt CLEAN under the corrected instruction; queue emptied

The requeued guide deployed at 16:50. `[MEASURED]` at the artefact: **zero
banned-shape hits** in the rendered text (the same regex the ban uses), 1,150
words, and the old app described affirmatively throughout — the lead sentence is
"The old Noted kept everything in the browser you wrote it in, on that one device,
and nowhere else", which is the corrected writer_block's framing almost verbatim.
So the instruction was the defect: same writer, same page, same ban; only the
guidance changed.

Guide CTAs then set via section editor (texts chose their own destinations: hero
"Find your notes from the old Noted" and primary "Save everything" → the tool;
secondary → /how-it-works.html). Verified rendered (hero 1 anchor, cta 2, both
linking the tool) and on the box at 22:07 (`200/24164`, 3 tool links).

Closed with asserted row counts: the guide's 2 `unresolved_cta` (resolved), the
privacy `unresolved_cta` (**no CTA by design** — the hero carries no cta_text and
the approved copy has its own contact route), and the original 08-12
`needs_content_page` (superseded by the reconcile-emitted rebuild). **The work
queue for this site is now empty** except the three inert `detected`
discovery-rotation rows from 08-10, which nothing dispatches.

---

## 2026-08-14 — three owner answers, and the discovery that there was NO EDITOR

Owner: (1) the privacy page does NOT mention the 30-day backup retention — ruled,
recorded in `evidence_base.supplied_copy.privacy.open_question`, closed; (2) test
the degraded states headlessly then hand over a manual walkthrough; (3) cutover
when we're ready.

### The discovery that reframed (2)

`[MEASURED]` **The product had no editor.** The engine is API-only (grepped
`Routes()`: `/api/*` handlers, nothing serving HTML); `app.noted.co.uk/` serves
the framework's marketing index through nginx `location /` (grep: zero
`textarea`/`/api/notes` in any served page; only contact.html touches `/api/`).
So "Sign in" → `https://app.noted.co.uk/` → the marketing page → its own
"Sign in" — **a loop with no product in it.** Every prior verification was
API-level (curl smoke tests), which is exactly how a missing UI stays invisible:
each layer proved its own half and nothing walked the person's path.

The degraded-state clause protected a surface that did not exist. So the editor
was built — it is PLAN §4.2's own item ("the editor becomes a
`component_level='tool'` component; its JS deploys calling /api/ same-origin"),
not new scope.

### The editor (`editor_tool/noted-write.html`, function `tool-write`)

Sign in / register (engine messages surfaced verbatim), notes list, title +
textarea, Save. The degraded contract is the spine, not a feature:

1. **"Saved ✓" appears ONLY on a 2xx** — sequence is dirty → "Unsaved changes"
   → "Saving…" → server answer decides. Never optimistic.
2. **Failure is loud and text is sacred** — bordered `role=alert` banner
   ("…nothing has been lost"), status "NOT saved", content never touched, Try
   again re-sends the same text.
3. **beforeunload asks** while dirty, and only while dirty.

Media capture is deliberately absent from v1 (the engine stores it; this UI does
not record it — the tool-doc says so).

### The headless proof (`editor_tool/test_editor_degraded.py`)

21 checks, six cases, against a local stub faithful to `server.go`'s shapes with
a deliberately SLOW save — the delay is what makes the ordering claim *provable*
(mid-flight the status must read "Saving…", not "Saved"). Outage = Playwright
`route.abort` on POST /api/notes; server error = fulfilled 500 with the engine's
own message. **This is the induced-failure test the platform runner cannot do.**

**All three contract guards mutation-verified:** optimistic-Saved → the ordering
check fails ("was 'Saved ✓'"); banner suppressed → the suite dies on the missing
banner (a TimeoutError, caught-by-crash — legible enough, noted); clear-text-on-
failure → both "text untouched" checks fail. Clean run restored: ALL CASES PASSED.

Two instrument notes: Playwright's headless shell **suppresses the real
beforeunload prompt** (tried real activation — click+type — then
`close(run_before_unload=True)`; no dialog ever arrives), so Case F tests the
guard's arm/disarm via a cancelable synthetic event and the REAL prompt belongs
to the manual walkthrough. And an `exit=$?` after a `| tail` pipe reads tail's
status, not python's — the mutation-2 "exit=0" was that trap, already in memory,
stepped in anyway.

### Deploy + the owner's half

Born via `076_create_noted_write_tool.sh` (create_tool_component; tool-doc header
in place first this time). Pages `tool-write` (/tools/write/index.html) +
companion guide; assemble item filed with the **page_id column** set (the proven
v2 shape); plan + structure spec updated to 10 pages. Manual walkthrough for the
owner: `WALKTHROUGH_2026-08-14_degraded_states_by_hand.md` — four parts (normal
save; devtools-Offline outage; beforeunload; cross-browser proof), what pass
looks like, and the one honest limit (a mid-request death can save without
telling; retry converges because a retried save updates, not duplicates).

Next when the page lands: sitesync tick → live smoke on the real page →
re-point the site's "Sign in" CTAs at /tools/write/ → bind
`authenticated-note-sync` to the built component (§4.1, still open).

---

## 2026-08-15 — the editor is LIVE, smoke-tested against the real service; the Sign in loop is closed

The assemble item completed overnight via the estate queue (no bypass — the owner
did not answer the offer, and waiting was free). Page `deployed`, publicly serving
at `https://app.noted.co.uk/tools/write/` (200/22947, editor markup present,
tool-doc stripped).

### The live smoke (`editor_tool/smoke_live_editor.py`) — ALL 11 CHECKS PASSED

Real page, real nginx, real engine, open internet. Registered a throwaway
(`noted-smoke-<epoch>@example.invalid` — unroutable by design; the engine has no
account-deletion endpoint so the account remains, holding one test note), then:

- **outage first**: route.abort on POST /api/notes → loud banner, "NOT saved",
  text untouched, and — the clause itself — **"Saved ✓" never appeared**;
- **recovery**: unroute → Try again → the SAME text saved by the real engine;
- **the product's promise**: a SECOND independent browser context signed in and
  got the note back, same title, same text.

The script takes a base URL argument precisely so it can be re-run **on the apex
at cutover** — origin is scheme+host+port and §6's warning applies to the editor
now too.

### The Sign in loop closed

`index` and `how-it-works` primaries re-pointed from `https://app.noted.co.uk/`
(which looped to the marketing page) to **`…/tools/write/`** via 074 (EDITS
updated; its Usage comment line, mangled by an earlier sed, repaired). Verified in
`rendered_html` (all four slots) and on the box — index at 14:41, how-it-works one
sitesync tick later at 14:45. Shopfront 200 throughout; apex still the legacy app.

### State: the walkthrough is ready for the owner's hands

`WALKTHROUGH_2026-08-14_degraded_states_by_hand.md`, at
`https://app.noted.co.uk/tools/write/`. Blocker 4 closes on the owner's word, not
mine — the loud-enough/obvious-enough judgement is the part that was always his.

Still open after that: bind `authenticated-note-sync` + `legacy-local-data-adoption`
to the built components (§4.1, `site_experiences`, proposed → bound → verified) —
the editor now EXISTS to bind to; then cutover (§4.5, owner's word) with the §6
origin probe and the smoke re-run against the apex.

---

## 2026-08-15 — owner: "that all works" — BLOCKER 4 CLOSED

The owner walked the four parts by hand at `https://app.noted.co.uk/tools/write/`
(normal save; devtools-Offline outage; close-tab prompt; cross-browser proof) and
confirmed. The judgement that was always his — loud enough, obvious enough for a
person mid-thought — is made. Degraded states verified by hand; handoff §5
blocker 4 closed. Launch conditions remaining: bind the experience patterns
(§4.1), then cutover (§4.5) on his word.

### The experience patterns are BOUND (§4.1) — through the action, and its doors earned their keep

`bind_site_experience` (bind_site_experience_action.go) refuses four things —
unclosed binding, empty value, anchorless selector, dead page — and I hit two of
its refusals before it accepted, both correct:

1. **"no bindings supplied"** — shape, not content: `bindings_field` defaults to
   `experience_bindings` at the ROOT of collected data; mine sat under
   `input_data.` Fixed with the config key the spec provides
   (`bindings_field: input_data.experience_bindings`), not by moving the payload.
2. **"`empty_state` is supplied but the entry never reads it"** — door 1's second
   half, the one the header says rots: a value nothing uses misrepresents what the
   fork asserts. Dropped it.

**One schema change, stated in the SQL and here:** `legacy-local-data-adoption`
declared `adopt_control` as REQUIRED and used it in **zero** checks or contract
clauses. It was declared on 08-11 for the sign-in-and-copy-up route that was
deliberately not built (the download route is the mandatory one — §6, no account
needed). Binding a required key to a selector that does not exist on the page
would be exactly the "check that cannot fail" the action's own header warns of.
Made optional with the reason written INTO the schema; required again when the
adopt route ships. `EXPERIENCES_2026-08-15_bind_noted_patterns.sql`.

**A correction to my own read, cheap because I checked before asserting:** I first
took door 4 ("a page role that resolves to nothing") to mean
`destination_roles=["account"]` would refuse — this site has no `account` page,
the editor is `role=tool`. Reading `experiencePageBindings` shows door 4 checks
bindings whose SCHEMA type is `page`, and both schemas type everything `string`.
Door 4 was never engaged.

`[MEASURED]` `site_experiences`: both rows **`proposed`**, every selector a real
element id on a deployed page (`#nw-*` from noted-write.html; `#lr-*` from
noted-legacy-rescue.html), `api_base` `/api` (same-origin, correct on both hosts
and after cutover). `proposed` not `bound` because both patterns are still
`draft` — the action records a draft's fork as a proposal, which is §4.1's first
rung, not a step skipped. Promoting the patterns out of draft, and the green run
that writes `verified`, are the experience loop's job; three checks stay
known-unrunnable by the platform and are covered by the two Playwright probes.

**Everything before cutover is now done.** §4.5 remains, on the owner's word:
repoint the apex from B2 to the box, keep the legacy app reachable for a grace
period, re-run `test_legacy_rescue.py`'s origin premise and
`smoke_live_editor.py https://noted.co.uk` on the new origin the moment it flips.

---

## 2026-08-16 — CUTOVER DONE. noted.co.uk now serves the framework build.

Owner: "carry on and repoint the apex."

### The insight that made it safe

`portfolio-sites-router` carries **B2 key bindings** (`B2_APP_KEY`, `B2_KEY_ID`)
and fetches the bucket **directly** — it never consults the zone origin. So while a
worker owned `noted.co.uk/*`, the apex served the legacy app from B2 **regardless
of DNS**. That made the whole cutover decomposable:

- DNS and tunnel ingress could be prepared with **zero user impact** (verified —
  after pointing the apex DNS at the tunnel, the apex still served B2, notice
  intact, `x-amz-*` present);
- **the worker route is the switch**, and rollback is one PUT restoring it.

The rollback was written to `RUNBOOK_cutover.md` and **committed before the flip**,
not after.

### Order, all verified before the switch

1. **Legacy app preserved at `https://noted.co.uk/legacy-app/`** — md5-identical to
   what B2 served (5 files compared). Placed at `/var/www/noted-legacy/`, **outside
   `/var/www/noted.co.uk`**, because sitesync rsyncs `--delete` into that root every
   5 minutes and would erase it. **It must stay on THIS origin**: IndexedDB is
   origin-keyed, so on a subdomain it would show every visitor an empty database and
   look exactly like data loss.
2. Tunnel ingress for apex + `www`; `ingress validate` OK; applied with
   `systemctl restart` (never `kill -HUP` — it terminates, and the unit has no
   `Restart=`). Shopfront checked either side.
3. Apex DNS → CNAME to the tunnel (`--overwrite-dns`; the record it replaced was
   unreadable — the token is workers-routes scope only — which was acceptable
   ONLY because the worker, not DNS, is the rollback lever).
4. **The flip**: `noted.co.uk/*` → `script: null`.

### Verified after, at the artefact

Ten public paths 200 with correct sizes — index, privacy, migrate, how-it-works,
about, contact, both tools, the guide, and `/legacy-app/`. Engine healthy;
webdesign shopfront untouched.

**`smoke_live_editor.py https://noted.co.uk` — 11/11**, including the induced
outage and a second browser session retrieving the note: the editor works on the
new origin.

**The migration premise, re-proven on the LIVE site** (§6's requirement, now
`legacy_tool/probe_origin_after_cutover.py`): seeded `NotedDB` through the
preserved legacy app at `https://noted.co.uk/legacy-app/`, then loaded the rescue
tool at `https://noted.co.uk/tools/legacy-rescue/` — it found the notes and the
recording, and **the seeded note came back in the downloaded file with its text
verbatim and its recording as a data URL**, engine-compatible format string. The
origin is byte-identical to before (`https://noted.co.uk`); only the server behind
it changed, which is exactly why everyone's notes are still reachable.

### Three things that misled, recorded

- **Immediately after the flip, the apex read 404s and 4393 bytes** — worker-route
  propagation across the edge, not a fault. Cache-busted requests showed the
  origin correct throughout. *A control taken seconds after a routing change can
  read the OLD world.*
- **The same trap on nginx**: controls taken immediately after `systemctl reload
  nginx` returned the OLD config's answers (three different paths all reporting
  26889 bytes — the fallback index). Settled moments later.
- **`nginx -t` broke because I left a `.bak` in `/etc/nginx/sites-enabled/`** —
  nginx glob-includes that directory, so the backup was parsed as a second config
  (duplicate `limit_req_zone`). The running nginx was unaffected, but the next
  reload would have failed. Backups now live in `/root/nginx-backups/`.

Also: my first origin-probe assertion (`"1 note"`) failed because the legacy app
creates its own note on load — **my assertion was wrong, not the product**. The
strengthened version asserts the SEEDED note is in the rescued payload, which is
the claim that actually matters.

`www.noted.co.uk` has never existed on this zone (no record, no links, canonical
is the bare apex), so no regression and nothing that could shift the origin. The
ingress rule added for it is inert and correct if it is ever added.

---

## 2026-08-17 — the old app is RETIRED and the notice is gone

Owner: "let's retire the old app and remove the notice."

**Measured before touching anything:** every one of the 22 hits `/legacy-app/` ever
took was our own cutover probing (16:03–16:04 on the 16th, curl + headless Chrome
via the tunnel loopback) — no real person visited it in the grace period. Nothing
on the live site links to it, and the wind-down notice existed only inside the old
app's own index.html, so retiring the app removes the notice with it.

**What "retired" means, precisely:** serving stopped; **no bytes destroyed
anywhere.** The copy remains at `/var/www/noted-legacy/` on the box, in
`gqls/sites` master, and in B2. `/legacy-app/` now **302s to
`/tools/legacy-rescue/`** — anyone arriving there almost certainly wants their
notes, and the rescue tool is the page that returns them, on the same origin, so
it reads the same IndexedDB the old app wrote.

Two small traps on the way, both caught by verifying at the artefact:

- nginx's default absolute redirect built the Location from the LISTEN address —
  `http://noted.co.uk:8082/...` — fine box-locally, broken publicly. Fixed with
  `absolute_redirect off;` (relative Location, resolved by the browser against the
  real origin). Verified publicly: 302 → `https://noted.co.uk/tools/legacy-rescue/`,
  followed to a 200.
- The origin probe seeded its test notes THROUGH `/legacy-app/`, so retirement
  broke the probe, not the product. Re-pointed it to seed from the rescue page
  itself (origin is what matters, not which page writes) and re-ran: **9/9**, and
  cleaner — "1 note", exactly the seeded one, since the old app no longer runs to
  create its own.

The notice ("being refreshed") now greps **zero** on every public path. Rollback,
should the owner ever want the old app back at a URL:
`/root/nginx-backups/noted.co.uk.pre-retire-20260817` restores the alias block.

---

## 2026-08-17 — owner: header CTA → sign-in page; email → noted@contactforsales.com

Two content changes, one platform gap, one shell trap.

### The email (noted@contactforsales.com), everywhere it lives

Inventoried first: `sites.email` + `sites.content_data.email`; the approved privacy
copy (draft file → `supplied_copy` → `writer_block` inline → rendered page); 6
other components (migrate ×2, contact ×3, and `tool-storage-risk-explainer-guide`
— **a page another loop built since the 15th**, swept the same); the chrome
footer; and the two scripts that carry the address (074b, checker). Order
mattered: `sites.email` first, because the hallucinated-email validator and the
chrome renderer both read it.

- draft + scripts: sed; checker re-run **clean, control fires**
- evidence_base: `update_email_in_evidence_base.py` — derived row, 7 bans intact,
  old address `f`, body regenerated FROM the draft
- components: `074c_section_editor_noted_email_sweep.sh` — discovers targets at
  run time (so the loop-built page was swept without knowing its name), replaces
  recursively, fires content_edit per slot with only the changed fields
- privacy: 074b re-run → **22/22 verbatim, new address, old address absent**
- footer: chrome re-render from the updated sites row

### The sweep's first run did 1 of 6 — the stdin-drain trap

`while IFS= read -r t; do fire_one …` — and `fire_one`'s status poll runs
`kubectl exec -i`, which **inherits the loop's stdin and drains the remaining
five targets as its own input**. Exit 0, no error, "swept 1 slot(s)" — and the
standalone reproduction parsed all six perfectly, which is what pointed at the
loop rather than the pipeline. Fix: feed the loop on **fd 3**
(`read -u 3 … done 3<<<`). The general form joins the estate's silent-shell-trap
family: **a `-i` flag inside a stdin-fed loop is a consumer of the loop.**

### The header CTA — derived, not configured; two-layer fix

`render_site_components_action.go` DERIVES the header CTA (contact nav item,
else resolver-ranked fallback; text hardcoded "Get Started") — no owner control.

1. **Durable**: opt-in `sites.content_data->>'header_cta_url'` / `header_cta_text`
   override, applied after the derivation (absence = today's behaviour on every
   site), gated by the SAME ChromeLinkPolicy so a stale override degrades with a
   Warn instead of dangling (191's class). Package tests pass. **Inert until an
   image rolls.** Council-Submitted `89f3331e…`; commit `229e14e74`. Key set for
   noted (`/tools/write/index.html`) so the first post-roll render locks it in.
2. **Interim**: header `rendered_html` patched to the same target. **Known gap:**
   a pre-roll chrome re-render regenerates `/contact.html`; the next post-roll
   chrome render converges to the override. Recorded, accepted.

A GitHub 503 failed one deploy_page mid-way (transient; retry clean). Full-site
re-assembly queued (corr `8cfeb9c1…`, ~139 ahead) — verify at the box when it
lands: zero `hello@noted.co.uk` on every served page, header CTA
`/tools/write/index.html` everywhere, footer carrying the new address.

### Verified end to end — with one instrument lesson

Re-assembly (12 items) completed; box synced. Final state, measured at all three
layers:

| layer | old address | new address | header CTA |
|---|---|---|---|
| DB components + chrome | 0 | present | `/tools/write/index.html` |
| box files | 0 | present (index 1, migrate 3, privacy 2, contact 5) | ditto |
| live | 0 | **edge-obfuscated** | ditto, all 10 pages |

The live grep for the new address returned 0 on most pages while the box files
carried it — because **Cloudflare email obfuscation rewrites every address into a
`data-cfemail` span at the edge**. Decoding the span (XOR with its first byte)
yields exactly `noted@contactforsales.com`, and browsers run the decoder. So: **a
grep for an email on a Cloudflare-proxied page tests the obfuscator, not the
page** — measure at the box (pre-edge) or decode the span. My earlier race
explanation (sitesync tick) was wrong; the obfuscation was the whole story, and
contact's `new=1` live was the one occurrence sitting where the rewriter does not
reach (a form attribute).

Everything the owner asked for is live: Get Started → the sign-in page on all ten
pages; the email changed in every store, page, and the canonical privacy copy
(22/22 verbatim). The durable CTA override rides the next fleet roll
(`229e14e74`, Council-Submitted `89f3331e…`).

---

## 2026-08-18 — "a person will answer" removed; the CTA override is LIVE and derived; council round 2

### The phrase (owner edit)

Removed through the full chain: draft ("write to noted@contactforsales.com." —
checker clean, 308 words) → evidence_base (derived row, 7 bans intact, phrase
absent from the whole spec) → 074b re-render → the storage-risk guide's
article-body carried the same sentence and was edited the same way. Verified at
the box (privacy 0, guide 0) and live (0), header CTA intact.

**Instrument note:** the verbatim check read 21/22 — the missing "sentence" was my
tag-stripper inserting a space where the mailto anchor is now sentence-final
(`…com</a>.` → `com .`). The raw HTML carries the anchor followed directly by the
full stop; with ` .`→`.` normalisation, **22/22**. Second instrument lesson in two
days from the same checker: the diff tests the strip, then the copy.

### The fresh build carries the override — the interim patch is retired

`build provenance: 0b185bad2…`; `git merge-base --is-ancestor 229e14e74 <stamp>`
→ ancestor. Re-rendered chrome under the new binary: the header CTA came out
**`/tools/write/index.html` BY DERIVATION** from `header_cta_url` (updated
18:13:59). The interim rendered_html patch is now unreachable state — any future
chrome re-render produces the owner's CTA from the key. The pre-roll reversion
window recorded on the 17th is closed.

### Council: round 1 REVISE was RIGHT about the plan, already false about the world

editquality's gating objection: the plan's rationale said "noted sets the key in
the same change" but the edits array held only the Go edit — "ships as a dead
feature for the one site that requested it". The key had in fact been set BEFORE
round 1 (the round's own reviewer read-back shows it), so the defect was the
PLAN's under-description, not the estate's state. Per practice: revised, not
defended — round 2 under the same correlation adds the executed config change as
an explicit edit, and answers the second reviewer signal (a `site_config` aspect
existed and was unevaluated): `sites.content_data` chosen because
`loadSiteDataFull` already reads the sites row (zero new query paths) and the map
already carries sibling presentation knobs; migrating later is one UPDATE + one
COALESCE. Also: the trigger refuses a pseudo-path in `edits[].file` ("ONE EDIT =
ONE FILE, repo-relative") — a DB config change is pointed at the doc that records
it.

---

## 2026-08-19 — noted lost its CSS on the 18th; diagnosed, repaired, contributed to bug 198

Full mechanism, evidence and repair are in `bugs_open/198` (second-incident
section, appended today — that lane owns the defect) and the new LANDMINE
("A `css_themes` row can be BORN EMPTY…"). One-paragraph record: theme row born
empty 08-15 (git got the real 17,475 bytes, DB got nothing, no check compares);
21 contrast patches 08-18 appended correctly to the empty base under 198's
candidate-3 guard and `deploy_css` committed the DB value wholesale over the
file (17,475 → 91 → 2,381); the contrast loop amplified its own damage. Repaired
v23 = base + provenance + patches; DB==repo==box==live (20,367 B), 98 base
selectors serving. The git-writer shrink guard 198 named on 08-04 is still the
open door. New standalone handoff: `HANDOFF_2026-08-19_continue_here.md`.

---

## 2026-08-24 — pasteboard stage 1 BUILT AND PROVEN LOCALLY: any media in a note (engine + editor)

Owner (08-22): images, videos, GIFs, audio all in one note; a pasteboard to
paste them onto, move them, edit them; "we can get there in stages". The staged
design and its decisions: `PLAN_2026-08-24_media_pasteboard.md`. Stage 1 = the
plumbing end to end, keeping the linear editor; stage 2 = the board; stage 3 =
editing in place.

### Engine (`box/noted-engine/`, commit 2e6b04aa8)

`video` joins the kinds — the CHECK lives in an already-created table, so
`CREATE TABLE IF NOT EXISTS` cannot change it; the idempotent
`DROP CONSTRAINT IF EXISTS` + re-add pair in schema.sql is the real migration.
`DELETE /api/media/{id}` (account-scoped in the SQL; the existing trigger frees
the quota). `GET /api/media/{id}` now serves via `http.ServeContent` so Range
works (a `<video>` could play but never seek before). `/api/me` reports
`max_upload`. Each note's list entry gains a unified `media` array (all kinds,
upload order); the grouped `audio`/`images` fields stay.

Tests 10/10 against a throwaway postgres:16 container; the two new guards were
shown able to go RED (unscoped delete → LEAK caught; dropped video kind →
refused-upload caught).

**Misstep 11 — a `$`-anchored sed cannot restore a line that continues past the
mutation site.** My mutation-restore seds anchored on end-of-line; both target
lines continue (a backtick SQL argument list, a brace), so both "restores"
matched NOTHING and left the mutations in place — while printing success. Caught
in the same breath because the full suite is re-run AFTER restoring, which went
red. The cheap check, used for the editor mutations after this: keep a pristine
copy and `cmp` byte-identical after restore, never trust the sed.

### Editor (`editor_tool/noted-write.html`, commit 908970add) — NOT yet deployed

Paste / drag-drop / picker; kind from MIME (GIF = image/gif, animates natively);
strip renders images inline and video/audio as players from `/api/media/{id}`
same-origin (cookies ride, Range works); remove asks first and the item leaves
the strip only on the server's 2xx; storage meter + size/type refusals BEFORE
any bytes travel, from `/api/me`. Contract 4 joins the tool-doc: an item shows
as stored ONLY on a 2xx; a failed upload is held loudly with Try again; the
beforeunload guard covers held media. A brand-new note saves itself first
through the SAME save path (the upload needs a note to belong to). Existing
`#nw-*` ids untouched (experience-pattern selectors bind to them).

Harness: 4 new cases, 38 checks green; three new guards mutation-verified —
silent-vanish-on-failure (crash-red: the suite times out waiting for "NOT
stored", exit 1), optimistic removal, guard ignoring media. Restore proven by
byte-compare against a pristine copy, then a clean run.

### Deploy state, and the order that binds it

**ENGINE FIRST**: the new UI calls `kind=video`, `DELETE /api/media/{id}` and
reads `max_upload` — all absent from the running binary (built 08-10). The
editor degrades gracefully against the old engine (meter hidden, no pre-checks)
but video uploads would be refused and removes would 405. So the tool update
(`scripts/initial_messages/140_tool_suggester/077_update_noted_write_tool.sh`,
replace_existing/TL-047 — supplied html verbatim, no LLM, shrink-guarded) is
HELD until the engine has rolled, then the page_rerender is hand-filed (RESET
the `page_id` COLUMN — the wrong-page trap of 08-12) and the live smoke re-run
(now includes a media round-trip: upload in session 1, DECODE in session 2 —
naturalWidth > 0, not just the tag — then live DELETE there).

**The engine deploy itself was REFUSED by this session's permission classifier**
(scp + install + restart on the box). Binary built (static linux/amd64,
15,028,354 B, at `box/noted-engine/noted-engine`), commands in the RUNBOOK; the
owner can run them, or approve the attempt.

Quota note for the owner: 50 MB/account and 25 MB/file will pinch for video —
both env-tunable (`NOTED_MEDIA_QUOTA_MB`, `NOTED_MAX_UPLOAD_MB`), and the quota
is the valve protecting the shared 50 GB disk, so raising it is his call, not
ours. Account deletion (open thread 3) grows more urgent with media: the smoke
probes now leave media-bearing throwaway accounts (this run's own deletes them).

### Open thread 1 CLOSED — the CTA override's round 2 verdict

`[MEASURED]` 2026-08-24, `doc_notes`: **COUNCIL GATE — APPROVED** at 2026-08-18
18:28 UTC under the same correlation `89f3331e-57f4-4f8f-8f58-de6222d17337`,
"approved with 5 advisory objection(s) — none high-severity". Commit `229e14e74`
carries `Council-Submitted:` and 098 resolves it at report time — no amend,
nothing owed. The five advisories live in the full report
(`diagnosis_artifacts` kind=council_report, that correlation, latest row) and
are undispositioned — a small next-session item, not a gate.

---

## 2026-08-24 (later) — stage 1 DEPLOYING: engine LIVE, tool regenerated, page in the queue

The owner ran the two RUNBOOK blocks by hand (the session's classifier refused
box writes; checksum compared before the overwrite: `83c895ce…` both sides).

**Engine LIVE + verified at the artefact, 15:3x UTC:**
- `DELETE /api/media/1` unauthenticated via the public edge → **401** (route
  registered; the pre-media binary answers 405 — no DELETE pattern), with the
  GET control also 401 (so the 401 is not vacuous);
- live constraint: `kind = ANY (ARRAY['audio','image','video'])` — the startup
  migration ran;
- shopfront box-local control 200 (33641 → 34061 B across the ~40 min between
  baselines — that lane ships continuously; nothing on this path touches :8080);
- backup of the old binary at `/root/noted-engine.pre-20260824`.

**Tool regenerated in place, 15:45:29 UTC** (077, replace_existing; publish via
kafka_publish_checked — receipt PUBLISHED, corr `f20de677-a836-40ab-8905-39c4ae5a7d15`;
orchestration COMPLETED). Verified at the artefact, positive AND negative:
placement rendered_html (23,174 B) contains `nw-add-media` = t, contains the old
"Voice recordings and photographs are NOT" doc line = f.

**page_rerender hand-filed** for tool-write (`c50da088-c07a-4dd6-9096-08c5618bfd37`,
item_key `…_assemble_media_v1`): temp-table copy of today's 13:57 tool-write row,
completion fields nulled, status `triaged` (the claim gates filter
`IN ('triaged','approved')` — claim_work_item_action.go:102), and the target
asserted by the LEFT JOIN before believing anything: the `page_id` COLUMN points
at **tool-write**. NOTE: rerenders for this site ran at 13:57 TODAY (another
lane's or a sweep's) — they predate the 15:45 regeneration, so the served page
is still pre-media until this item runs.

`[MEASURED]` queue position at filing: **492 claimable items older than mine**,
drain ~152/h (completions in the last hour) ⇒ ~3 h. Queue alive, so the
single-page bypass (whose condition is a DEAD queue) is not invoked; a watcher
polls the row every 5 min. Remaining when it lands: bytes on the box, live page,
`smoke_live_editor.py https://noted.co.uk` (now with the media round-trip).

### Stage 1 IS LIVE — the queue landed, the chain verified, the smoke passed

The rerender item completed 18:13 UTC (~2.5 h queued — inside the ~3 h
estimate). The watcher's own record carries an instrument note: 13 consecutive
polls 17:08–18:08 returned EMPTY (stderr was discarded, so the cause is lost —
likely transient exec/API failures), then recovered unaided. An empty poll is
"instrument down", not "status unchanged" — a watcher that asserts on emptiness
would have died an hour early. Keep stderr next time.

Chain verified at every artefact: item result deployed exactly
`tools/write/index.html` to **vm-sites** (commit `2ac2eed8cd…`); box file
updated 17:41, `nw-add-media` present, old doc line absent; live page carries
the UI. Then `smoke_live_editor.py https://noted.co.uk` — **ALL 15 PASSED**,
incl. the new media round-trip: upload via the real picker → storage meter live
→ SECOND independent browser session decodes the image (`naturalWidth > 0`,
bytes not tags) → live DELETE removes it there. The throwaway account holds one
text note; its media was deleted by the smoke's own final step.

**Stage 1 (PLAN_2026-08-24) is DONE: built, deployed, live-proven.** Next:
stage 2, the pasteboard proper (notes.layout JSONB + the board view).

---

## 2026-08-25 — the CTA-override advisories DISPOSITIONED (open thread 4 closed)

Nine advisory points across 5 objecting reviewers (7 abstained), three themes:

1. **Un-ledgered config write** (editquality m/guardian l/debug_historian m —
   the same point three ways): ACCEPTED. The hand-run `content_data` UPDATE is
   now `sql_for_agents/608_noted_header_cta_override_ledger_backfill.sql` —
   idempotent no-op against the live row, `DO`/`RAISE` verify (not bare
   SELECTs), carrying the approval corr. Committed `b02c5955b`.
2. **Fait accompli / build-id-not-artefact / claims-outside-schema**
   (editquality m, debug_historian l, prior_art l): ANSWERED BY STANDING
   RULINGS — review here is after the fact by design (owner ruling 2026-07-29
   §2, shared HEAD), and the liveness claim rested on behavioural proof (the
   re-render DERIVED the URL), which the objection itself grants is stronger
   than a symbol grep.
3. **Mechanism follow-ups, each CHECKED 2026-08-25**:
   - guardian m — "confirm override-after-policy on ALL render paths": the
     override has exactly ONE application site, inside the shared render
     function immediately after the derivation, same `chromeLinks.Allows`
     gate (render_site_components_action.go:198–212) — ordering is a property
     of the function, not the calling pipeline. Confirmed by grep: no other
     .go file applies `header_cta_url`.
   - bug_historian l — sibling call sites: no footer CTA derivation exists in
     the file (footer builds legal/social links only). No sibling gap.
   - prior_art m — "site-header schema may already carry cta_url": it does,
     but `"source": "renderer"` — it is the plumbing the derivation WRITES
     INTO, not a pre-existing owner control. The new key is the control.
   - bug_historian m — a REFUSED override logs only a Warn; the owner's
     explicit request can silently degrade. REAL and ACCEPTED AS FOLLOW-UP,
     not built now: emit an owner-visible signal (work item) on override
     rejection. Recorded here so it is not folklore; small platform change,
     own council round when picked up.

---

## 2026-08-25 — B2 storage + stage 2 board BUILT AND PROVEN; deploy staged for the owner

Owner rulings executed (PLAN addendum): media → B2; stage 2 mobile-first;
deletion planned (`PLAN_2026-08-25_account_deletion.md`); advisories done
(entry above).

### B2 (engine)

Bucket `personae-noted-media` (private) + app key `noted-media-engine`
(listFiles/readFiles/writeFiles/deleteFiles, bucket-scoped) created via the
CLI. **The live service refuses `/b2api/v2/` AND `/b2api/v3/`
`b2_authorize_account`** ("not currently supported on API version number N")
— found because the credential was probed against the REAL service before any
Go was trusted; v4 nests everything under `apiInfo.storageApi` and makes
`allowed` a bucket LIST. Client is v4, stdlib-only, refuses a non-bucket-scoped
key. Full live round-trip through the engine's OWN client: upload, 206 ranged
download, delete, delete-again-converges — `TestB2LiveRoundTrip`, gated on the
env keys so CI without credentials skips loudly.

Suite 13/13 + live test; mutation-proven: quota-refusal-leaks-B2-object and
layout-clobbered-by-layoutless-save both went red.

### Stage 2 board (editor)

List|Board toggle; text + media as items; Pointer Events only; 44px handles,
`touch-action:none` on handles alone; fractional-of-width coordinates;
arrangement rides Save. Harness 51 checks over cases A–L incl. the 390px
touch-viewport case (no overflow, touch drag, computed touch-action). Three
board mutants red: drag-not-dirty, layout-dropped-from-save,
pixels-not-fractions.

### Misstep 12 — a directory-pathspec commit cannot take an untracked file

Two commits described the B2 client; neither contained it (`WRONG_CALLS.md`
2026-08-25 has the tell and the cheap check). Repaired forward; the shipped
binary was always built from the real sources.

### Deploy state

Binary (v4 B2 + layout) built and committed. Owner runs: binary install +
`/etc/noted/noted.env` gains the two `NOTED_B2_*` lines (snippet staged in the
session scratchpad — the secret never entered the repo or scrollback) +
restart. Then: startup line must read `media storage: B2 bucket
personae-noted-media` (LogState — the greppable honest signal), 077 fires the
stage-2 editor, page_rerender (page_id COLUMN), live smoke. Existing media
rows keep serving from Postgres (`storage_key IS NULL` path) — no migration of
old bytes required, drain optional later.

### 2026-08-25, later — B2 + board DEPLOYING: engine LIVE on B2, editor queued

The owner copied binary + env snippet (checksums compared: `e5aca46d…` both
sides); the install ran from this session this time. Engine verified at the
artefact: startup line **`media storage: B2 bucket personae-noted-media`**
(LogState — the greppable positive), health ok, shopfront control steady
(200/34061, unchanged from yesterday's post-change baseline), and all three
new columns (`media.storage_key`, `media.b2_file_id`, `notes.layout`) present
on the box DB — the startup migration ran. Old binary backed up at
`/root/noted-engine.pre-20260825-b2`; the env snippet was consumed and
removed from /root.

Editor stage 2 fired via 077 (corr `be9854e7-5083-4437-9c3b-77ee90826ca0`,
receipt PUBLISHED, orchestration COMPLETED); placement verified carrying
`nw-board` AND `nw-view-toggle` (33,081 B). page_rerender hand-filed
(`516689bd…`, item_key `…_assemble_board_v2`, column asserted → tool-write),
**55 items ahead** at filing. Watcher armed — WITH stderr this time
(yesterday's instrument lesson). When it lands: box/live bytes, then the live
smoke, then `b2 ls` as the after-control (the smoke's own delete should leave
the bucket empty again).

### 2026-08-25, evening — B2 + THE BOARD ARE LIVE, proven end to end

Queue took ~5 min (55 ahead). Chain verified at every artefact: deploy commit
`7356af8c…` (one file, vm-sites, repo HEAD carries both board markers); box
synced 11:35 (83,713 B, `nw-board` present); live page serves it. **Live smoke
15/15 on the apex** — and the media leg now runs the REAL B2 path, proven
three ways: the engine's own journal (`b2 upload begin key=media/acct_11/…
bytes=73` — the smoke's PNG), the second session DECODING it back through the
proxy, and `b2 ls` EMPTY afterwards (the delete removed the object — no
orphan). An empty bucket alone cannot discriminate the path; the journal line
is the positive.

One instrument note: the first box check after the item completed read the
OLD file (07:48) — the deploy commits to git and sitesync ticks 5-minutely, so
a check inside the tick window reads yesterday's world (§6 of the cutover
runbook's lesson, again).

**Stage 2 (PLAN_2026-08-24 §Stage 2, mobile-first per the 08-25 ruling) and
stage 1b (B2) are DONE: built, deployed, live-proven.** Remaining on this
programme: account deletion (planned, two owner decisions), stage 3 (edit in
place), the quota's paid tier when the owner wants it.

### 2026-08-25, addendum — a re-run of the install recipe clobbered its own backup (repaired)

The owner re-ran the (already-completed) install block three times; each run's
FIRST step — `cp` live binary → backup — succeeded before the chain died on the
consumed `.new` file, so `/root/noted-engine.pre-20260825-b2` became a copy of
the NEW binary (`e5aca46d…` == live, measured). The live service was never
touched (the restart sat behind the failing step). **Repaired**: stage-1 binary
extracted from git (`git show 2e6b04aa8:…/noted-engine`, sha `83c895ce…`
matching the 08-24 install record) and scp'd back over the clobbered copy;
verified on the box. **Door closed**: the RUNBOOK recipe now runs `chmod` on
`.new` FIRST, so a re-run fails before the backup step. The general shape: in
an `&&` chain, every step BEFORE the guard runs on every retry — put the
existence check first, or a "safe to re-run" script quietly is not.

---

## 2026-08-25, second session — deletion + captions + image editing BUILT; deploying

Owner rulings: immediate deletion GO; stage 3 = crop/rotate/captions as the
SEED of an editing surface (his different-note-type vision recorded in the
PLAN); quota stays 50 MB; tier copy drafted, NOT shipped (no mechanism —
bugfix-161's class).

**Engine (commit a63a6e21b, binary `4ab02cf6…`)**: `DELETE /api/account`
(password re-typed; B2 objects before the row cascade; refuses wholesale if any
object cannot go; goodbye only after everything is gone) + `PATCH
/api/media/{id}` captions (on the ROW, 500 chars). 16/16 tests incl.
B2-refusal-aborts-everything + retry-converges-to-empty-bucket; password gate
and B2 cleanup mutation-verified red. **DEPLOYED + verified**: sha matched,
active, still `media storage: B2`, DELETE /api/account and PATCH /api/media
both answer 401 unauthenticated (old binary: 405), caption column present,
shopfront 200. Rollback: `/root/noted-engine.pre-20260825-deletion`.

**Editor (commit db5df07ba)**: Account panel (loud failures, goodbye on 2xx);
captions in strip + board; rotate/drag-crop editing that never mutates in place
(copy first, original after, board items RETARGET preserving position; GIFs
refused with the reason). Harness 61 checks (A–O); MUT-G refusal-as-success
crash-red, MUT-H silent-edit-failure crash-red, MUT-I optimistic caption red.
Smoke now SELF-CLEANS: deletes its throwaway through the live panel and proves
the credentials die (commit fdb4fb502).

**Deploying**: 077 corr `39c50b3c…` COMPLETED, placement carries
`nw-account-panel` + `nw-edit-canvas` (47,444 B); page_rerender `909e6c3d…`
(`…_delete_edit_v3`, column asserted → tool-write), 136 ahead at filing,
watcher armed. On landing: box/live bytes → live smoke (now 19 checks incl.
live deletion self-cleanup).

`[MEASURED]` box DB at 15:40 UTC: media rows 0 Postgres / 0 B2 (the smokes
deleted their own), accounts 5 (throwaways + owner tests) — so the proposed
backup sentence ("media is not in the backups") is exactly true with no pre-B2
residue, and the 5 throwaways can be drained through the live endpoint once
this lands.

### 2026-08-25, close — deletion + captions + editing LIVE; the estate is clean

v3 rerender complete (commit `a1556e04…`); box + live page carry the account
panel and edit canvas after the sitesync tick (checked AFTER the tick this
time). **Live smoke 17/17** — the run's finale deleted its own account through
the live panel and proved the credentials dead. Then the four historical smoke
throwaways (ids 7/9/10/11) were drained through the LIVE endpoint with their
reconstructed credentials (login 200 → delete 200, all four); after-control:
**exactly 1 account remains, the owner's**. The privacy page's "close your
account" promise now has a mechanism behind it, exercised in production five
times today.

Awaiting the owner's word (presented in chat): the backup sentence for the
privacy page (ships through draft → checker → evidence_base → 074b on his
yes) and the paid-tier copy (HELD until a purchase mechanism exists).

---

## 2026-08-25, third session — the backup sentence is LIVE; tier copy approved and HELD

Owner approved both drafts (tier copy minus its last sentence). The backup
sentence went through the full chain and is live: draft edited (08-12 open
question RESOLVED in place with the measured facts) → checker (whose figure
regex knew only marketing shapes — "30 days" read as "none"; widened, and the
flagged figure then GROUNDED as evidence_base fact #1, registered by derived
supersede-insert, 7 bans carried) → NEW re-runnable
`refresh_privacy_copy_from_draft.py` (the two 08-12 one-shots cannot re-run:
apply FATALs on its long-gone clause, embed's "already embedded" is a
presence check, not currency) → 074b (both content_edits COMPLETED) → box +
live.

**Verified 26/26 sentences verbatim**: 25 on the live page + the email
sentence at the BOX (the edge obfuscator rewrites it — the documented
instrument case, measured where the rewriter does not reach).

Three instrument lessons in one chain, each caught before it could mislead:
the checker's shape-blind figure regex; the refresher's own verify probe
straddling a hard-wrap (false absence — normalised both sides); urllib's
default UA drawing a Cloudflare 403 that read as a page failure (curl fetch
substituted). The habit held: suspect the instrument, then PROVE it.

Paid-tier copy (approved wording, minus the 50 MB-free sentence) recorded in
PLAN_2026-08-24 — ships the day a purchase mechanism exists.

---

## 2026-08-25 (late) / new session — CTA-override follow-up BUILT (handoff §4 item 1)

The 08-25 disposition's accepted follow-up (bug_historian m: a REFUSED
`header_cta_url` override logs only a Warn) is now code: a refused override
files an owner-visible **`cta_override_rejected`** work item beside the Warn
(`emitCTAOverrideRejectedItem`, render_site_components_action.go).
`needs_human_review`, no handler, severity medium (nothing dead ships — the
derived CTA serves; the defect is a stated intent silently ignored). Key
`cta_override_rejected:<site_id>` with **refreshOnConflict**, not the sibling
emitters' drop: the open row must name the CURRENT refused value (bugs_open/184
class). Fire-and-forget — a refused override must never fail the render.

Proof: two call-site-driving sqlmock tests (hitl_refresh_adoption_test.go
pattern); the policy pin **mutation-verified red** (flipped to dropOnConflict →
FAIL on the unmet UPDATE expectation, reverted → PASS). Tested against
COMMITTED HEAD via `verify-head-builds.sh --with` (both files overlaid),
because the shared working tree's package test-compile is broken by another
lane's untracked WIP (`write_audit_findings_field_capability.go`), and HEAD's
own findingcodes scan separately fails on the 396 lane's undeclared
`WORK_ITEM_STATUS_OVERRIDE_REFUSED` — both pre-existing, neither mine, neither
touched.

Register: **STY-058** (styling-render-pipeline.md + index row) — which also
closes a standing gap found on the way: the 08-18 content_data override itself
had NO register entry (the 07-28 §2 same-commit condition was missed a week
ago; stated in the entry rather than backdated). The entry carries the
two-carriers landmine: a grep for `header_cta_url` hits BOTH STY-053
(config.chrome, policy-BYPASSING) and this one (content_data, policy-vetted) —
"one application site" is true only PER CARRIER.

Council: submitted, corr `10b84432-eef8-4462-b451-7d791ee78d33` (DRY_RUN
admission passed first); committing with `Council-Submitted:` per the 07-30
rule. Inert until the next chassis roll either way.

### Council verdict on the CTA follow-up — APPROVED round 1, advisories dispositioned

Corr `10b84432`: APPROVED, 2 advisories, none high. Run start was immediate
(not the measured 29-min queue), verdict ~20 min. Dispositions live in the
STY-058 register entry (the durable home); the ones that changed anything:

- **bug_historian l (key coarseness) — ACTED ON**: item_key gained the slot,
  `cta_override_rejected:<site_id>:header` (STY-054's shape). A per-site key
  made "one override slot per site" a silent assumption; closed, not documented.
- **bug_historian m / reuse_agent m (other silent-degrade sites) — CENSUSED**:
  5 `.Allows(` call sites in non-test Go as of 2026-08-25; exactly ONE vets an
  owner-set value (this one). The rest vet derived values — refusal there is
  the mechanism working. A second owner-set carrier is the trigger to move the
  emit to the policy boundary.
- editquality m (register "promised but not in plan"): satisfied in fact —
  the gate refuses prose paths as plan edits; commit `169ac5e1b` carried both
  register files. editquality l (anti-churn × refresh): read at the door,
  behaviour correct (open row → refresh; recent terminal → Deferred; two
  strikes → `[unresolved]`, which for this type means the human twice closed
  it and the refusal recurred — the visibility intended). guidelines l
  (ON CONFLICT worry): the shared door's idx_swi_dedup-matched ON CONFLICT is
  the sanctioned lockstep contract; reuse IS the compliance.

Tests re-run green with the new key (mocks don't bind the key, so the earlier
mutation proof stands unchanged). Still inert until the next chassis roll.

---

## 2026-08-26 — design-discovery rotation resumes; noted's exposure checked BEFORE its turn

Cross-session heads-up (webdesign-tool-rebuilds lane): `site-discovery-rotation-design`
re-enabled 2026-08-26 09:20Z after 15 days paused (the 08-11 cost-scare pause was never
unwound; `bugs_open/401`). ~1 site/3h, least-recently-visited first — noted's turn
inside ~2–3 days. Findings are born `detected`; `detected-item-promoter` (15-min
cadence) can auto-dispatch known-good (item_type, handler) pairs. **A surprise design
item or auto-repair on noted in the next few days is THIS, not a stray thread.**

Exposure checked today rather than after a surprise:

- **generic_theme (the colour-churn landmine) will NOT fire on noted.**
  `[MEASURED 2026-08-26]` the live check requires `content_data->'color_scheme'` to be
  a non-empty OBJECT and noted's is (`jsonb_typeof`='object', `<> '{}'` true). Verified
  at the ARTEFACT: the current check's SQL literal is present in the live chassis
  binary (`agent-chassis-6dd68888dc-8h8b7`, /proc/1/exe grep, positive + nonsense
  controls both behaved). No `design_intent.palette.reference_values` block exists on
  noted — advisory-only anyway (see the corrected memory landmine) — and
  `enforceLayoutScheme` (live since v1.0.1140) bounds background/text damage if a
  design pass ever does run.
- **Probe near-miss worth keeping:** grepping the binary for the FIX COMMIT's literal
  (`content_data ? 'color_scheme'`, commit 3437f2212) read ABSENT with working
  controls — because a later hardening rewrote the query. A capability probe must use
  the CURRENT source's literal, not the diff's; the first read would have "proven" a
  five-week-old fix undeployed.
- The other rotation checks (palette_contrast, image_url_404, tool_health,
  missing_css) have no known noted-specific trap; anything they file will be visible
  as work items on the site and should be read against this entry first.

---

## 2026-08-26 (cont.) — the sweep wave hit TODAY, not in 2-3 days; site verified intact; two peer lanes engaged

The "~2-3 day ramp" in the rotation heads-up was wrong for noted: LRV-first put
us near the front, and tool/audit sweeps were ALREADY active overnight (first
completions 03:35Z, i.e. BEFORE the 09:20Z design-rotation re-enable — separate
mechanisms). ~40 new work items on noted this morning; 67 open rows total.

**Verified intact after the wave** (all at the artefact, 2026-08-26):
- Live smoke **17/17 PASSED** (editor, engine, media, deletion).
- Privacy "30 days" spot-check present (full 26/26 not re-run).
- **Stored tool-write component (2f24b506) IS our shipped editor**: 47,444
  chars vs repo file 47,531 B, `nw-` ids + "Saved ✓" present, updated_at =
  our 08-25 15:33Z ship. The TWO failed instance_scope_conversion attempts
  wrote nothing. So the 12 queued page_rerender items redeploy the same
  editor — not a loaded gun.

**What already changed today (03:34–03:36Z + 07:42Z):**
- Head chrome rebuilt twice ("deactivated Document Head" ×2) — this TOOK THE
  GTM TAG (analytics lane's 08-24 backfill was artefact-only; their
  cross-session note + `bugs_open/397` §9; they are applying the durable fix
  with owner go). The open `chrome_divergence_overwritten` (head, 49,431 B) is
  THEIR backfill being archived out — told them to claim it; also pointed
  them at the 08-18 header sibling (2,385 B, unknown provenance).
- `undeployed_asset` 'icon' "deployed=true" **three times** — and the
  referenced paths STILL serve the 68,972 B HTML fallback. ⚠ noted 200s
  every path INCLUDING an invented control, so status-code probes are
  useless here — discriminate by content_type (text/html = broken).
- page_rerender on tool-legacy-rescue-guide ran, then went
  `[unresolved after 2 attempts]` on its misdirected-CTA reason.

**Standing findings:**
- **favicon.png / logo.png / og-card.png references in the head are broken on
  the wire and PRE-EXISTING** — both the archived (old) and current head carry
  them; today's rebuild did not introduce them. No hero.jpg on index/about
  served HTML. Damage class: no favicon, broken link previews. The framework
  is queuing LLM-generated brand assets (logo, icon, hero, og-card) at
  asset-deployer/image-build-handler — **a branding decision on the owner's
  product; flagged to him in README rather than allowed silently**.
- improve_tool "Write has no @media": fact true of the tool's own CSS
  (0 by design, fluid mobile-first), inference refuted by the 61-check
  harness incl. 390px. improve_tool "legacy-rescue has no <style>": same
  heuristic class, unassessed. add_tool ESCALATION spec embeds the full
  incumbent contract ("preserve these invariants") — careful, but
  replace_existing:true on the product's core surface for a multi-instance
  convention a single-instance tool can't benefit from. **Asked the
  webdesign-tool-rebuilds lane (msg 12db35e6) for the sanctioned exemption
  route instead of wont_fixing standing-sweep items that would re-detect.**
- Watchlist: `evaluate_tools` (tool-suggester could ADD tools),
  `prerequisite_missing` ×2 (research_results / content_sources absent —
  could pull noted into content-research flows; detected, no handler yet),
  contrast_failure ×7 unresolved at css-patch-agent (08-24, pre-wave).

### The escalation's premise is a WRITE CONFLICT — and the three editor items are now HELD

Digging into WHY tool-write was "still unconverted after 2 terminal attempts":
**both attempts COMPLETED, fixed:true** (08-19 repair_instance_scope_bindings,
08-25 scope_component_instance, both on 2f24b506). The stored component today
is OUR unconverted source at 15:33Z 08-25 — i.e. **each 077 `replace_existing`
ship reverts the sweep's conversion, and each side's result honestly says it
succeeded.** The add_tool regeneration escalation is built on that miscount.
Durable fix if conversion is ever ruled necessary: convert the SOURCE file
(editor_tool/noted-write.html), not the stored row our ships overwrite.

Actions taken (per the webdesign-tool-rebuilds lane's code-read: no exemption
mechanism exists; wont_fix frees the dedup slot and re-files; our items were
at `triaged` — INSIDE the claim gate, sharper than their detected-framing):
- improve_tool 5701a96d, add_tool 8812d35a, instance_scope_conversion
  f8721913 → **`deferred` + empty handler + spec.not_dispatchable/hold_reason
  inline** (capability_gap precedent). Each UPDATE guarded `AND
  status='triaged'`; all three took (nothing had claimed them).
- Question + evidence + offered resolution filed as
  `bugfix_283_component_instance_scope/CONTRIB_2026-08-26_from_noted_lane_tool_write_single_instance_and_a_write_conflict.md`
  — engaging the 08-22 owner rulings (inline-with-literal-ids makes tool-write
  a "genuine target" by their seam; our three differences argued; ruling
  requested). Release = set the items back to `triaged`.

Analytics lane follow-up: head divergence item efa20f84 confirmed as 397's
mechanism; their session's permission layer blocked closing the 11 such items
— queued for the OWNER, answer written in bugs_open/397 §10-addendum. **We do
not do that write for them** (cross-session permission laundering). The 08-18
header sibling ab9afa54 (2,385 B) is confirmed NOT theirs — provenance
unknown, stays on this lane's watchlist.

---

## 2026-08-26 (later) — the GitHub Actions "error" diagnosed: transient, GitHub-side, and inert for noted either way

Owner raised a workflow failure on gqls/vm-sites (run 32984485832, push
e94e207, `Rerender: guides/tool-write-guide.html`, "Startup failure" with
GitHub's own "we've been automatically notified" annotation). Diagnosis, all
verified 2026-08-26 ~15:30Z:

- **Not this thread's push and not any thread's defect.** e94e207 touched ONE
  file (`noted.co.uk/guides/tool-write-guide.html` — the rerender wave rolling
  the GTM restore + css fixes through, as the analytics lane forecast). The
  workflow file last changed 2026-08-05, so the pushes could not have broken
  startup. The failure window is 15:07–15:14 bracketed by SUCCESS at 14:58
  and again at 15:23:37 — a transient GitHub-side incident, exactly what the
  annotation says. One run stuck `queued` 12m from the window; expect GitHub
  to reap it.
- **Actions is not in noted's deploy path at all.** The box is pull-model:
  `sitesync` hard-resets to vm-sites main every 5 minutes. Proven end-to-end:
  the GTM tag is live on the guide page (2 refs) despite that page's two
  workflow runs both failing. The "Deploy Sites to VM" workflow serves the
  push-model domains, not this box.

### Found while looking: ONE stray `*/` survives on the live page — currently harmless, precisely measured

Live `guides/tool-write-guide.html` line 777:
`p.info-card-grid__subtitle { color: #5a5a4e; } */` — the trailing `*/` is
outside any comment. Under CSS error recovery a stray `*/` swallows the next
`{}` block. The swallowed victim here is the line-796 `:root` — which is a
BYTE-IDENTICAL duplicate of the line-362 `:root` (the page carries the
tool-portal-light layout twice; `--color-primary`/`--color-primary-text` each
defined exactly 2×, same values). **Net damage today: zero.** e94e207's diff
shows the sibling stray on `__card-body` was already cleaned; this one
remains in the stored stylesheet source, so every rerender reproduces it.
Residue of the documented 08-15→08-19 empty-theme-row accumulation (the
in-artefact "BASE RESTORED 2026-08-19" note, base git abe1b617a7).
**Forward risk, [INFERRED]:** anything later appended directly after line 777
would be swallowed. One-character fix in the stored stylesheet row — left
with the evidence rather than hand-applied, because the css-patch surface has
open fleet items (7 `contrast_failure` unresolved) and a documented history
of well-intentioned direct fixes propagating the wrong thing.

---

## 2026-08-26 (evening) — LARGE UPLOADS BUILT AND PROVEN (the 25 MB blocker; owner ruling: tier holds on this)

The whole chunked-upload programme, per PLAN_2026-08-26_large_uploads.md.
Constraint stack measured first: engine buffered whole files ([]byte to B2),
nginx capped 30m (ours), **Cloudflare ~100 MB/request (NOT ours) — so this was
a protocol design, not a bigger env var.**

**Engine** (commit "noted engine: chunked uploads", binary rebuilt+committed):
begin/parts/finish/abort on B2's large-file API (v4, stdlib, part sha1s
verified both sides); quota RESERVED at begin via pending_uploads and counted
beside media_bytes in EVERY quota check; per-account
media_quota/max_upload_override_bytes columns = the tier lever (NULL = env
default, existing accounts byte-identical); part-size chosen per upload
(>=2 parts for B2, <100 MB for Cloudflare), persisted on the row; reaper
hourly with B2 as truth (age-guarded rowless orphans); account deletion
cancels pending uploads. Suite green incl. 7 new tests; **4 guards
mutation-red on buildable mutants** — the first mutation PASSED because it hit
the dead arm (AddMedia vs AddMediaB2), re-aimed; a later mutant was a compile
error and re-shaped, because a mutant that does not build proves nothing.

**Editor** (commit "noted editor: chunked upload client", shipped via 077 at
15:38Z, stored row verified has_chunked=t): parts with per-part retry x3,
progress in the strip, STORED only on finish's 2xx, hard failure loud + held
+ reservation released; split point and cap read from /api/me — **inert
against the old engine by construction** (no small_upload_max key → old
single-shot path, so ship order is free, and default accounts can never reach
chunking even after the engine lands: max_upload = small cap). Harness 61→**75
checks** (cases P–S); 3 editor mutations red INDIVIDUALLY — the combined
no-retry+no-abort run crashed in case Q before case R ran, so the abort
mutation was re-proven alone (guard-in-series, third time this week).

**Box half staged for the owner** (RUNBOOK "Large uploads" section): engine
install (schema go:embedded — binary swap is the whole migration), nginx
30m→100m (repo copy already edited), self-cleaning >30 MB acceptance probe
(override on a throwaway via box psql; the tier lever has no HTTP surface,
deliberately).

**Pending tail:** the wave's own `page_rerender: tool-write` (97b5fdcc,
triaged) carries the editor to the wire — queue HEALTHY but busy fleet-wide
(noted drained 15 rerenders today; loop rotated to sites with 275+ waiting),
so it had not landed after 50 min. Watcher re-armed; **live smoke (17 checks)
owed the moment it lands** — until then the wire is byte-unchanged and the
last smoke (this morning, 17/17) stands.

Escape-trap note: an Edit mixing the file's literal em-dash with its `…`
escape was unplaceable — my channel rewrites emitted `\uXXXX` into the
literal char (the standing landmine) — python patcher with the escape built
programmatically is the working route.

### The editor ship LANDED — via the 413 bypass, smoke 17/17

The wave's `page_rerender: tool-write` (97b5fdcc) never got claimed: noted's
9 eligible rows (created ~15:25Z) sat 2h while the loop ran ~440 claims/2h
into sites with hours-older backlogs — `bugs_open/413`'s ordering class,
filed today by the dispatch_throughput lane (measurement contributed to
their dir). Resolution per the documented bypass
(single-page-deploy memory, all its dos observed): CLAIMED the item first,
fired 049b direct page-rerender (corr `3f179092`, assemble-only — correct,
since 077 had already written the chunked placement render), orchestration
COMPLETED deploy_success=true in seconds, repo commit `321bb6ebf` 17:23:33Z,
**on the wire one sitesync tick later (17:27:39Z, 3 chunkedUpload refs)**,
item completed honestly with the route+corr in result. **Live smoke 17/17**
against the new editor page on the OLD engine — the chunked path inert
exactly as designed. Remaining: the owner's two box blocks (RUNBOOK "Large
uploads") and the >30 MB acceptance probe after them.

---

## 2026-09-02 — LARGE-UPLOADS ENGINE DEPLOYED by the owner; verified at the artefact; smoke 17/17

Owner ran the two box blocks (2026-08-31→09-02 session): binary installed
(rollback at `/root/noted-engine.pre-20260831`), unit active, health ok;
nginx 30m→100m, `nginx -t` clean, reloaded. Verification from this side:

- **Capability probe, decisive by construction**: a self-cleaning throwaway's
  `/api/me` answered `small_upload_max: 26214400` — a KEY the old binary
  cannot emit (no control needed; absence-vs-presence is the discriminator).
  Defaults intact: max_upload 25 MiB, quota 50 MiB, overrides NULL.
- **Shopfront control byte-identical** across the whole touch (200, 69,304 B
  before and after — same-day baseline per the RUNBOOK gotcha).
- **Live smoke 17/17** against the new engine (media, board, deletion — the
  throwaway self-cleaned).

The chunked path is now REACHABLE but closed by default everywhere (default
max_upload = small cap). **Remaining, optional timing:** the >30 MB
end-to-end probe (RUNBOOK "Large uploads" — throwaway + box-psql override +
a real ~35 MB file through the editor + self-delete). The paid tier
discussion resumes on the owner's word after that proof.

---

## 2026-09-03 — owner's three findings: migrate CTA + nav FIXED on the wire; 50 MB ruling handed over; a silent save-skip found and closed on the way

Owner tested and reported: (1) migrate's "Save everything" does nothing,
(2) the nav "Save my notes" tab should save their notes, (3) the per-file cap
message says 25 MB, should be 50.

**(1)+(2) — both were SELF-LINKS in authored data** (the 08-26 sweep's item
6a087e4f had filed exactly this, awaiting a product decision the owner has
now given). Fix, deterministic end to end: hero `content_data.cta_url` and
`site_nav_items` "Save my notes" both → `/tools/legacy-rescue/index.html`
(the real export surface; `__cta_minted` left as honest history); chrome
rerendered via STY-055 (corr 7843d3a7, asserted at the STORED header, not
the status); migrate rerendered FROM DATA (corr b5191169, sections_saved=4,
rows dated today); remaining 11 pages re-assembled (assemble-only 049b ×11).
**Wire-verified: hero button and index nav both carry the new href; editor
intact (3 chunk refs); live smoke 17/17.** Item 6a087e4f completed with the
decision + trail.

**The finding on the way (LANDMINES 2026-09-03 + 049b patched, 25648df04):**
the FIRST rerender (corr 7487f8fa) went COMPLETED/deploy-success with
`rerendered: 4` — and saved NOTHING: `save_page_sections` skips
`{"success":true,"sections_saved":0,"reason":"no page name"}` when
`input_data.spec.page_name` is absent, and the pipeline then deploys the
STALE rows all-green. 049b now derives page_name from page_id and refuses a
reasoned dispatch without it. Scope honestly bounded by measurement: the
work-item route carries page_name — the wave's `[unresolved after 2
attempts]` CTA-rerenders are NOT this class (a rerender cannot fix a CTA
whose DATA still names the wrong target; fixing the data first, as here, is
the correct order — which their fixer should note).

**(3) — the 25 MB cap is env `NOTED_MAX_UPLOAD_MB=25`;** the 50 MB block is
staged with the owner (message of 2026-09-03) along with the interplay he
should rule on: cap=50 with quota=50 means one max file fills a free
account, and his 52.8 MB test file needs either a quota raise or his
account's overrides (which would also first-exercise the chunked path).
Awaiting his run/word; /api/me remains the truth either way.
