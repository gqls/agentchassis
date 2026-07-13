Deployment:
cd frontends/admin-dashboard
npm install
npm run build        # verify it builds locally
cd ../..
make build-dashboard DASHBOARD_TAG=v1.0.1
make push-dashboard DASHBOARD_TAG=v1.0.1
make deploy-dashboard
make dashboard-port-forward   # then open http://localhost:8080


local dev (no Docker needed)
# Terminal 1: port-forward core-manager
kubectl -n ai-persona-system port-forward svc/core-manager 8088:8088

# Terminal 2: run Vite dev server
cd frontends/admin-dashboard && npm run dev
# Opens http://localhost:5173, proxies API calls to core-manager


VITE.js

Vite is a build tool for frontend projects. It does two things:
In development, it runs a local server that serves your source files directly to the browser using native ES modules. When you change a file, the browser gets only that file's update — not a full rebuild. This makes the dev feedback loop near-instant even on large projects. It also proxies API calls (our /api → localhost:8088 config), so the SPA can talk to the backend without CORS issues.
For production, it bundles everything into optimised static files — minified JS, CSS, hashed filenames for caching — using Rollup under the hood. The output is a dist/ folder with index.html and a few .js/.css files. That's what nginx serves.
Why Vite over alternatives: Webpack is the older standard but its config is verbose and builds are slower. Create React App is deprecated. Vite has become the default for new React/Vue/Svelte projects — zero config for common cases, fast builds, and the proxy setup is a few lines.
For our use case it's straightforward: npm run dev for local development with hot reload, npm run build to produce the static files that go into the Docker image. The vite.config.js is 15 lines. There's nothing to maintain.
The existing frontends/admin-dashboard/ already had a vite.config.js placeholder, so this was the intended tool for the project.


Two new endpoints added:
PATCH /admin/sites/:site_id — updates site-level fields (company_name, tagline, email, phone, contact_address, logo_text). Only sends fields you want to change.
POST /admin/work-items — creates a new work item from the admin dashboard. Required fields: site_id, item_type, summary.
Once deployed, here's the flow for fixing gaswholesalers.com. First, what real contact data do you have for them? The steps will be:
Step 1: Update the site's contact details:

curl -s -X PATCH "http://localhost:8080/api/v1/admin/sites/5fe15466-4e2e-4ff2-981e-98c1b7074002" \
-H "Authorization: Bearer $TOKEN" \
-H "Content-Type: application/json" \
-d '{
"email": "REAL_EMAIL@gaswholesalers.com",
"phone": "+44 REAL PHONE",
"contact_address": "REAL ADDRESS"
}' | python3 -m json.tool

Step 2: Create a rebuild work item for the contact page:

curl -s -X POST "http://localhost:8080/api/v1/admin/work-items" \
-H "Authorization: Bearer $TOKEN" \
-H "Content-Type: application/json" \
-d '{
"site_id": "5fe15466-4e2e-4ff2-981e-98c1b7074002",
"item_type": "content_rewrite",
"summary": "Rebuild contact page with real contact data",
"severity": "high",
"handler_agent": "page-build-handler",
"page_id": "40bda1d8-33e6-40e2-93ae-d718ee27d84f",
"priority": 10
}' | python3 -m json.tool

Step 3: Resolve the placeholder item via the dashboard Resolve button, with note "Real contact data provided, rebuild queued".
The improvement loop will pick up the new content_rewrite item, the page-build-handler will rebuild the contact page pulling the updated email/phone/address from the sites table, and the deployer will push it live.


