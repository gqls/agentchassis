# The owner's voice corpus — anonymised, awaiting his approval of the removals

**Supplied by the owner 2026-08-26** after granting permission to use his writing
(`../PROVENANCE.md` §"The corpus measurement" records why nothing usable existed before this).
**6,595 words**, eleven pieces: nine blog/site articles and one file of ~30 emails.

**Status: APPROVED 2026-08-26** — *"I approve the removals"*. Cleared for use in the voice datasets.

## What was removed, and the rule applied

His instruction: *"strip all the names and domain names and the html and my name from the copy"*.
Applied slightly wider than asked, because two categories are stronger identifiers than either
names or domains and would have been odd to leave in:

| category | replaced with | note |
|---|---|---|
| his own name (three variants, incl. two bylines) | `[NAME]` | |
| every correspondent's first name | `[NAME]` | not distinguished from each other — voice learning does not need to tell them apart, and distinguishing them re-identifies |
| domain names | `[DOMAIN]` | including ones he owned and ones being offered to him |
| his email address | `[EMAIL]` | |
| his username on a domain platform | `[USERNAME]` | |
| **a street address and postcode** | `[ADDRESS]` | ⚠ wider than asked. A meeting address is stronger PII than a name |
| **towns and cities** | `[LOCATION]` | ⚠ wider than asked |
| **a named former employer, and a government department** | `[EMPLOYER]` / `[ORGANISATION]` | ⚠ wider than asked |
| a coffee chain, a domain platform, an AI product | `[BRAND]` / `[PLATFORM]` / `[PRODUCT]` | |
| all HTML markup, nav menus, image galleries, affiliate links, script tags | deleted | no prose value |
| author bio blocks and their phone number | deleted whole | they were credentials, not voice |
| outbound reference URLs and citation lists | deleted | |

## What was deliberately KEPT

Subject-matter product names that carry no personal information and are part of what the writing
is *about* — Jira, Redmine, Drupal, Amazon, Trac, Mantis, GIMP, Golang, Kubernetes. Removing those
would gut the technical pieces without protecting anyone.

## ⚠ The raw text was never written to disk

This directory holds only the cleaned version. The original correspondence — third parties' names,
an address, a phone number — was anonymised in flight and never committed, because this repository
is version-controlled and a git history is exactly the wrong place for other people's personal
correspondence. The mapping from real value to placeholder exists only in the conversation where
he supplied it.

## What this unblocks

Datasets **1** (email voice), **3** (copy style) and **4** (support-reply tone) — blocked until now
on material, not permission. See `../PROVENANCE.md`.
