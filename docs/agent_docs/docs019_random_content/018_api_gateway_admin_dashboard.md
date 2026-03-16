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