# 015 — Dynamic Application Creation Guidelines

How the system creates web applications today, where it's heading, and the principles that guide both custom builds and integration with existing platforms.

---

## What We Build Today

The current pipeline produces static multipage websites: HTML pages, CSS themes, images, deployed to Cloudflare via git. Each site gets:

- Component-based page structure (header, hero, features, CTA, footer)
- LLM-generated content matched to industry and audience
- CSS theme with variable-driven colour system
- Interactive tools (calculators, converters) via fork-on-deploy
- Automated quality checks and improvement cycles

The pipeline handles the full lifecycle — planning, content writing, design, deployment, quality audit, and iterative improvement. It works. But it only produces one kind of output: static sites on Cloudflare.

---

## Where We're Going

Three tiers of increasing capability:

### Tier 1: Static Sites with Dynamic Components (now → near term)

What we have plus smarter components. Tools already run client-side JS. Next steps:

- Forms that submit to external services (Formspree, Netlify Forms, custom webhook)
- Client-side search (Pagefind, Lunr)
- Dynamic content injection (RSS feeds, API-fetched data rendered client-side)
- Analytics integration (plausible, umami — self-hosted, no cookies)
- A/B testing via client-side JS (no backend needed)

These don't change the deployment model. Still static HTML on Cloudflare. The "dynamic" part runs in the browser.

### Tier 2: Sites with Agent-Powered Backends (medium term)

Each deployed site gets a lightweight backend that agents manage:

- Newsfeed handling — agent fetches, curates, publishes content on a schedule
- Contact form processing — submissions stored, notifications sent, follow-up automated
- Affiliate integration — product links, click tracking, commission reporting
- Content research and writing — scheduled research → draft → review → publish pipeline
- Data collection — scraping, API polling, aggregation, presented on the site

The backend could be:
- A set of serverless functions (Cloudflare Workers, AWS Lambda)
- A lightweight Go/Node service per site (containerised, shared cluster)
- An agent sub-system dedicated to the site (smaller version of what we have)

### Tier 3: Full Application Generation (longer term)

The system doesn't just build sites — it builds applications:

- Admin panels (like the one we just designed for our own system)
- Dashboards with real-time data
- SaaS prototypes with auth, billing, and core features
- E-commerce storefronts with inventory and checkout
- Internal tools for specific business workflows

At this tier, the agents are writing backend code, database schemas, API endpoints, and frontend applications — then deploying and maintaining them.

---

## Architecture Principles

These apply at every tier.

### 1. Agents build what they understand

An agent that generates a React dashboard should understand React patterns, not just template-fill. This means the LLM prompts include framework conventions, component patterns, and anti-patterns specific to the target stack.

For each target framework, we need a **framework spec** stored in `content_components` or a new `framework_specs` table:
- Naming conventions (file structure, component naming)
- Available libraries and versions
- Security patterns (auth, CSRF, input validation)
- Deployment requirements (build step, env vars, runtime)

### 2. One site, one repository, one deployment

Every site or application owns its git repository. All files — frontend, backend, config, migrations — live in one repo. Deployment is triggered by git push. This keeps the model clean regardless of what the site contains.

### 3. Separation of generated and human content

Generated code and content must be clearly marked (HTML comments, file headers) so the improvement loop knows what it can safely regenerate and what a human has customised. Human edits take precedence — the system should never overwrite them without explicit instruction.

### 4. Backend complexity lives in agents, not in generated code

The generated backend should be as simple as possible. Business logic that requires judgment (content curation, lead scoring, anomaly detection) stays in the agent system. The site's backend is a thin layer that receives agent output and serves it.

Example: a newsfeed page doesn't run its own scraping. An agent scrapes, filters, summarises, and pushes the result to the site's data store. The site just renders what's there.

### 5. Incremental complexity

Start with the simplest version that works. A contact form starts as a mailto: link. Then it becomes a Formspree integration. Then a Cloudflare Worker that stores to D1. Then a full backend with CRM integration. Each step is a separate work item, not a big-bang rewrite.

---

## Frontend Creation Guidelines

### Component Model

Every frontend element is a self-contained component with:

```
<style>
  /* Layout and spacing only */
  /* Colours always via CSS variables */
  /* Mobile-first responsive */
</style>

<div class="component-name" data-component="component-name">
  <!-- Semantic HTML -->
  <!-- Accessible (labels, ARIA, focus states) -->
</div>

<script>
  /* Self-contained, no global pollution */
  /* No external dependencies unless explicitly approved */
  /* Progressive enhancement — works without JS where possible */
</script>
```

### CSS Variable Contract

All components use the same CSS variable names. This is the interface between components and themes:

```css
/* Colours */
--color-primary, --color-primary-hover, --color-primary-text
--color-secondary, --color-secondary-hover, --color-secondary-text
--color-accent
--color-text, --color-text-muted, --color-heading
--color-background, --color-surface, --color-card-bg
--color-border
--color-header-bg, --color-header-text
--color-footer-bg, --color-footer-text

/* Dark section overrides */
--section-text, --section-text-muted, --section-heading
--section-surface, --section-border

/* Layout */
--container-max-width (default: 1200px)
--spacing-section (default: 5rem 2rem)
--border-radius
--shadow
```

Components set dark section variables on their own container when `is_dark_section = true`. Child elements inherit automatically.

### Quality Rules

Content generated by LLM must pass validation before deployment:

- No placeholder text (NEEDS HUMAN REVIEW, Lorem ipsum, [INSERT], TODO:)
- No unrendered template variables ({{.field}})
- No hallucinated contact info (only use verified email/phone from site record)
- No fabricated testimonials, case studies, or statistics
- No cross-site contamination (wrong company name)
- Internal links only to pages that exist
- All sections must have real content or be hidden (not deployed with placeholder)

### Interactive Tools

Tools are self-contained HTML/JS/CSS components stored in `content_components` with `component_level = 'tool'`. Rules:

- All calculation/logic runs client-side
- No external API calls or CDN dependencies
- CSS variables for colours (never hardcoded hex)
- Mobile-responsive (touch targets ≥ 44px)
- Wrapped in `<div class="tool-container">`
- Include clear heading and instruction text

Tools can be library originals (shared across sites) or site forks (customised per site). The fork-on-deploy model means each site owns its copy independently.

---

## Backend Creation Guidelines

### For Tier 1 (static + client-side)

No backend code generated. Dynamic behaviour via:
- Client-side JS in component `<script>` blocks
- Third-party service integrations (forms, analytics, search)
- Embedded widgets (calendly, stripe payment links)

### For Tier 2 (agent-powered backends)

**Cloudflare Workers** (preferred for simple backends):
- Single JS/TS file per function
- D1 (SQLite) for data storage
- KV for caching
- R2 for file storage
- Deployed via wrangler in the same git repo

**Agent-managed services** (for complex backends):
- Go or Node service, containerised
- Deployed to shared K8s cluster or dedicated namespace
- Agent handles: schema migrations, endpoint creation, data pipeline config
- The site's backend is a consumer of agent output, not a producer of logic

### For Tier 3 (full application generation)

Target frameworks, in order of priority:

**React + Tailwind** (admin panels, dashboards, SaaS frontends)
- Component-based, familiar to most developers
- Shadcn/ui for consistent component library
- Vite for build tooling
- Good LLM generation support (lots of training data)

**Next.js** (full-stack applications)
- Server-side rendering for SEO-sensitive pages
- API routes for backend logic
- Auth via NextAuth
- Database via Prisma or Drizzle

**Laravel** (traditional web applications, especially for ISP hosting)
- Blade templates for server-rendered pages
- Eloquent ORM for database
- Built-in auth, sessions, queues
- Deploys to any shared hosting with PHP
- Strong ecosystem (Nova for admin, Cashier for billing)

**WordPress** (content-heavy sites, client-managed)
- Custom theme generation from our component model
- ACF (Advanced Custom Fields) for structured content
- WooCommerce for e-commerce
- Deploys to any WordPress host
- Client can manage content without developer involvement

---

## Integration with Existing Platforms

### The Publishing Model

Instead of "we deploy to Cloudflare", think "we publish to a target". The deployment step becomes pluggable:

```
Agent pipeline → content/code ready → publish adapter → target platform
```

Current adapters:
- `git-adapter` → GitHub → Cloudflare Pages (static sites)

Future adapters:
- `cpanel-adapter` → FTP/SSH → shared hosting (PHP/WordPress sites)
- `wordpress-adapter` → WP REST API → WordPress instance (content + theme)
- `laravel-forge-adapter` → Forge API → Laravel app on DigitalOcean/AWS
- `vercel-adapter` → Vercel API → Next.js/React apps
- `shopify-adapter` → Shopify API → e-commerce stores
- `cloudflare-workers-adapter` → Wrangler → serverless functions

Each adapter handles:
1. Authentication with the target platform
2. File upload / code push
3. Build trigger (if needed)
4. DNS/domain configuration
5. SSL provisioning
6. Health check after deployment

### WordPress Integration Path

The most pragmatic near-term expansion. Many ISPs include WordPress hosting. The agent system would:

1. **Generate theme** — convert our component model (header, hero, features, CTA, footer) into a WordPress theme with `functions.php`, template files, and `style.css`
2. **Create content** — use WP REST API to create pages, posts, menus
3. **Configure plugins** — install and configure ACF, Yoast SEO, contact form plugin
4. **Deploy theme** — push via FTP or Git (depending on host)
5. **Maintain** — content updates via REST API, theme updates via same deploy path

The advantage: clients get a site they can log into and edit themselves. The agent system handles the initial build and ongoing improvements, but the client isn't locked in.

### Laravel Integration Path

For clients needing custom backend logic:

1. **Scaffold application** — generate Laravel project with routes, controllers, models, migrations
2. **Generate views** — Blade templates from our component library
3. **Configure services** — mail, queue, cache, storage
4. **Deploy via Forge** — Laravel Forge handles server provisioning, SSL, deployments
5. **Agent backends** — background jobs that run agent-managed tasks (content refresh, data sync)

---

## Agent Sub-Systems for Sites

The longer-term model: each site gets its own miniature agent system.

### What a Site Agent System Looks Like

```
Site: gaswholesalers.com
├── content-agent     — monitors industry news, drafts blog posts
├── newsfeed-agent    — curates and publishes relevant articles
├── affiliate-agent   — manages product links, tracks commissions
├── analytics-agent   — processes visitor data, generates reports
├── seo-agent         — monitors rankings, suggests improvements
└── maintenance-agent — checks for broken links, stale content
```

These are lighter than our current agents — possibly just scheduled LLM calls with structured output, not full orchestration workflows. They could run as:

- Cloudflare Workers (scheduled CRON triggers)
- Lightweight containers in a shared cluster
- Functions within the main agent system, scoped to one site

### Shared vs Dedicated Infrastructure

| Aspect | Shared (current) | Dedicated (future) |
|--------|-----------------|-------------------|
| Agents | Run in shared chassis pods | Run in site-specific namespace or Workers |
| Data | All in clients_db | Per-site database (D1, dedicated PG schema) |
| Deployment | Shared git-adapter | Per-site adapter instance |
| Cost | Low (shared resources) | Higher but isolated (per-client billing possible) |
| Customisation | Limited to config | Full code customisation per site |

The migration path: start shared (what we have), add per-site scheduling (Tier 2), then per-site isolation (Tier 3) for premium clients.

---

## Build Order

### Phase 1: Expand the static pipeline (now)

- Deploy the admin API and review UI
- Tools: `tool-generator` (LLM-created tools), `tool-improver`
- Better content validation (placeholder detection, contact info verification)
- Fix the topic lifecycle and idle timeout issues

### Phase 2: Dynamic components (next)

- Contact form handling (Formspree → Cloudflare Worker)
- Client-side search (Pagefind integration)
- Analytics (Plausible/Umami snippet injection)
- Blog/article publishing workflow (markdown → HTML → deploy)

### Phase 3: Platform adapters (after phase 2)

- WordPress theme generation from component model
- WP REST API content publishing
- cPanel FTP deployment adapter
- Test with one real client on shared hosting

### Phase 4: Agent-powered backends (after phase 3)

- Newsfeed agent (scrape → curate → publish)
- Affiliate link management
- Content research and writing pipeline (scheduled)
- Per-site Cloudflare Workers for backend logic

### Phase 5: Full application generation (future)

- React/Next.js application scaffolding
- Database schema generation and migration
- Auth and billing integration
- Admin panel generation (like our site-admin-dashboard)
- Laravel application generation for traditional hosting

---

## Key Decisions Still Ahead

**Generated code ownership:** When a site owner edits generated code, how do we track what's theirs vs ours? Git diff tracking? Marker comments? A "lock" flag per file?

**Multi-tenancy model:** Shared database with site_id scoping (now) vs per-site schemas vs per-site databases. Affects cost, isolation, and complexity.

**Agent billing:** If each site gets its own agents, how is LLM usage tracked and billed? Per-call metering? Monthly token budgets? Fixed-price tiers?

**Framework selection:** Should the system choose the framework based on the site's requirements, or should the client specify? A "build me an e-commerce site" request could produce WordPress+WooCommerce, Shopify, or a custom Next.js app depending on constraints.

**Plugin/extension model:** Can third parties create agents or components that plug into the pipeline? A "Mailchimp integration agent" that someone else writes and we host?

These don't need answers now, but they shape the architecture. Every decision we make today should keep these options open rather than closing them off.
