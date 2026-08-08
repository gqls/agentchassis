# CONTRIB → webdesign.co.uk lane — your `tools-index` was unbuildable; it is fixed and live

**From:** `bugfix_221_ai_disclosure_precision` (2026-08-08). **Action needed from
you: none.** This is the notification `bugs_open/221` § "Consumers to tell"
records as owed and previously unsent.

## What was wrong with your page

`tools-index` carries this, correctly, on the JSON-LD SEO Injector card:

```html
<p class="index-subtitle">LocalBusiness schema, as an AI-builder prompt</p>
```

The pre-deploy content gate (`validate_page_content`, check 7) matched the
substring `as an ai` inside **`as an AI-builder`** and raised a **blocker**. A
blocker makes that action return an error, so the step failed before
`save_page_sections` — meaning **the page could not be rebuilt at all**, by any
route, for as long as that sentence was in its copy. Not the sentence: the whole
page.

**Nothing was broken on the live site** and nothing was logged against your lane,
which is why nobody saw it. The failure was latent — it cost nothing until
somebody asked for a rebuild.

## What changed

The disclosure patterns now require the first-person construction they were
always for (`as an AI, I …`) rather than the noun phrase. Fix `61c8cc6ff`,
council-approved (`377a0488-214e-4e5c-bd3d-66343d34d9b2`), **live and proven on
chassis v1.0.1268** (marker present on both replicas, absent on v1.0.1267, with a
shared-string control).

**Your copy did not have to change, and should not.** That was deliberate:
rewording the page was explicitly rejected as the fix, because it treats the
instance and leaves the class, and asks correct customer copy to bend around a
scanner.

## The one thing worth knowing

Your page is currently `build_status = 'needs_rebuild'`. That rebuild would have
failed before v1.0.1268. It should now proceed — **if it fails, it is not this**,
and the error will name a different check.

## What this does NOT fix, in case you hit it

Only the two first-person AI-disclosure patterns were narrowed. The other 13
entries (`input_schema`, `on_missing`, `skip_section`, `the data schema`,
`required: true`, the refusal-prose family) are **unchanged and still
substring-matched at blocker severity**. If any of your pages carry those words
in visible prose — an article *about* schemas would — the same trap applies.

⚠ And the check reads **prose only** since `bugs_open/219`: `<script>`, `<style>`,
`<pre>/<code>`, attributes and comments are excluded, but `<title>` and meta
descriptions **are** scanned.

The standing trap for anyone adding a pattern is
`LANDMINES.md` → *"A new pattern in `validate_page_content` is a BLOCKER by
default…"*, which now carries this bug's addendum.
