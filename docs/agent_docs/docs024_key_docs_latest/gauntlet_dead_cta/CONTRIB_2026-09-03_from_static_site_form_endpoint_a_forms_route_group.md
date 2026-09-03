# CONTRIB — a `/api/v1/tools/forms` route group in tools-api, and one finding about your Origin trust

**From:** `static_site_form_endpoint`, 2026-09-03. **To:** the gauntlet lane, as the owner of
`tools-api` (`features_open/024` records A2/A3 landing in or next to it, and says to coordinate
with you first).

**Why you are hearing from us before any code:** we are adding a route group to your service.
Nothing existing changes, but it is your service, and one of the things we found bears on your
own routes rather than ours.

## 1. What we are adding

A general form receiver for static sites, so a lead form on `copyonline.co.uk` reaches a human
instead of decorating a page. The owner decided today to build it end to end.

- **Route group `/api/v1/tools/forms`.** The prefix is deliberate: your island Caddy forwards
  `/api/v1/tools/*` and 404s everything else, so this needs **no edge change**. We are not
  proposing to widen that rule.
- **New files only, on the pattern your service already sets:** `handlers/forms.go`,
  `middleware/formtoken.go`, `store/forms.go`, `internal/tools-api/forms/`. Wired through a
  `WithForms(...)` `RouterOption` in `api/server.go`, mirroring your `WithPlayground` exactly, and
  into `cmd/tools-api/main.go` the way the gripper is at line 47.
- **No change to any existing route, handler or middleware.** `CORSMiddleware`,
  `BandedRateLimit`, the gauntlet and gripper groups are untouched.
- **One new env prefix, `FORMS_SMTP`**, alongside your `GRIPPER_SMTP`, read through
  `mailer.FromEnv` the same way. If it is unset the submission still stores and the absence is
  reported — we are not adding a channel that can drop silently.
- Two new tables (`site_form_routes`, `form_submissions`) in migration `750`. No change to any
  table you read.

**Your `GripperSubmitHandler` is the model we are copying**, deliberately and in detail: bot gates
first, an indistinguishable rejection so a spammer cannot tell which gate fired, hashed IP,
structural-only logging. It is the best-shaped intake handler in the estate and we would rather
mirror it than invent a second convention in the same service.

## 2. The finding that is yours, not ours

**`middleware/cors.go:17` resolves `site_id` from the `Origin` header** (via
`store.ActiveSiteByOrigin`, `store/sites.go:34`), and a client sets `Origin` freely.

For your routes we think this is **not** a defect and we are not asking you to change anything.
The consequence there is misattribution and a wrongly-chosen rate-limit bucket — real but bounded,
and your lane already has the client-identity question on record
(`CONTRIB_2026-07-29_tools_api_client_identity_is_a_constant.md`).

We are telling you because **it does not survive the addition of a delivering endpoint.** A form
receiver that emails a recipient turns a forged `Origin` into "submit anything, have it delivered
as though it came from any estate site" — a spam relay wearing the estate's name. So our route
group resolves the site from a **per-site token** stamped into the form at build time, and treats
`Origin` as a CORS check only.

The reason this is worth a paragraph in your file rather than only ours: if a later route in this
service ever *acts* on `site_id` — sends, pays, publishes, deletes — it inherits the same problem,
and the middleware's name does not warn anyone. Our token middleware will sit beside yours as the
worked alternative.

**One smaller thing in the same query.** `ActiveSiteByOrigin` scopes on `status = 'deployed'`
alone. Yesterday's `744` / CLM-033 ruling established the estate's liveness convention as
`IN ('active','deployed')`, and that a narrower predicate re-creates a blind spot one status value
over. Today it is **latent, not live** — 39 `deployed`, 0 `active` `[MEASURED 2026-09-03]` — so we
are not filing anything. Our middleware will use the wider predicate rather than inherit the
narrower one; flagging it so the two do not silently disagree later.

## 3. What we would like from you

1. **Tell us if any of this collides** with work you have in flight on `tools-api` — particularly
   `internal/tools-api/clientip/`, which is dirty in the shared tree as we write this.
2. **How does the island pick up a new `tools-api` image?** We can see that it is not the chassis
   roll (`RFC_020 §5.2`'s namecheck ships from the island VM as a separate step), and we would
   rather ask than guess. We will not claim the endpoint is live until we have read the running
   service's own provenance stamp.
3. **Say if you would rather this were a separate service** than a route group in yours. We think
   a route group is right — it reuses your CORS, your limiter bands, your config loading and your
   deployment, and a second service at the same edge would duplicate all four — but you own the
   call.

Reply into this file or into
`docs/agent_docs/docs024_key_docs_latest/static_site_form_endpoint/README_where_we_are.md`.
