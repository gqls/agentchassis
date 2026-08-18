# Cloudflare Worker: `portfolio-sites-router`

`worker.js` is the **B2 proxy fronting every B2-hosted site** (36 zones, one
account). It signs each GET with SigV4 against
`s3.us-east-005.backblazeb2.com`, bucket `portfolio-sites`, object key
`<hostname><path>`.

**Provenance:** exported from the LIVE worker via the Workers API on
2026-08-06 (`bugs_open/132` — the previous copy here was a stale Nov-2025
version that materially differed from production; the live source was under no
version control until this export). Keep this file identical to what is
deployed: deploy FROM this file, and re-export after any dashboard-side edit.

**Deploy** (needs `~/.cloudflare/404-token.env`, account
`13044f178ae0b156961065f55c8fada8`):

```sh
. ~/.cloudflare/404-token.env
curl -s -X PUT \
  "https://api.cloudflare.com/client/v4/accounts/13044f178ae0b156961065f55c8fada8/workers/scripts/portfolio-sites-router" \
  -H "Authorization: Bearer $CLOUDFLARE_API_404_TOKEN" \
  -F 'metadata={"main_module":"worker.js","compatibility_date":"2025-11-24","bindings":[{"type":"plain_text","name":"B2_APP_KEY","text":"<from ~/.cloudflare/portfolio-sites-router.settings.json>"},{"type":"plain_text","name":"B2_KEY_ID","text":"<same>"}]};type=application/json' \
  -F 'worker.js=@scripts/cloudflare/worker.js;type=application/javascript+module'
```

⚠ **BUILD THE METADATA FROM THE LIVE SETTINGS, do not hand-type it** (added 2026-08-18,
DGH-012). The recipe above is minimal and therefore lossy: this worker also carries
`observability` (enabled, with logs persisted) and a `compatibility_date`, and a PUT whose
metadata omits them resets them silently. Generate the block from
`~/.cloudflare/portfolio-sites-router.settings.json` — bindings, `compatibility_date`,
`compatibility_flags`, `observability` — and assert BOTH B2 bindings are present and
non-empty before the PUT. Worked implementation used for the DGH-012 deploy:
`deploy_worker.sh` in that lane's scratchpad; copy it here if you need it twice.

⚠ **The PUT response's `result.bindings` is `[]` even on a fully successful deploy.** It is
the API not echoing them, and it looks EXACTLY like the credential-stripping outage this
file warns about. Confirm from `…/workers/scripts/portfolio-sites-router/settings` and from
a live page fetch — never from the PUT response.

⚠ **`node --check` PASSES a syntactically broken `worker.js`.** ESM syntax in a `.js` file
makes the check a no-op: a copy with one `)` removed exited 0. Check it as `.mjs`
(`docker run --rm -v "$PWD:/w:ro" node:20-alpine node --check /w/worker.mjs`) **and prove the
checker fails on a deliberately broken copy in the same session** — a syntax error here is a
36-zone outage.

⚠ **The metadata MUST carry the two B2 bindings** (`B2_KEY_ID`, `B2_APP_KEY`,
type `plain_text`) — an update without them strips the worker's credentials
and takes every site down. Their values are NOT in this repo: read them from
`~/.cloudflare/portfolio-sites-router.settings.json` (or the dashboard).
They are `plain_text` vars, not encrypted secrets, so the settings API returns
them — a hardening candidate (convert to `secret_text`), noted in
`bugs_open/132`, owner's call.

**Verify after deploy** — request a genuinely missing path and read the BODY
(`curl -s https://<domain>/no-such-page-$RANDOM`): expect the site's own 404
markup with HTTP status 404, never the B2 JSON, never a 200. Fetching
`/404.html` directly proves nothing (it returned 200 throughout the bug). Then
confirm a real page still serves 200 and `/worker-health` still answers.
