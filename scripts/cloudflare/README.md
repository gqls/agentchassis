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
