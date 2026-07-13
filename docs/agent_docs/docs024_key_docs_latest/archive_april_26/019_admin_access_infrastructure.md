# Admin Access Infrastructure

How to expose the admin dashboard securely while keeping the agent system isolated.

---

## Architecture

```
Internet
    │
    ▼
┌──────────────────┐
│  WireGuard VPN   │  ← Only VPN clients can reach admin services
│  (K8s pod or     │
│   cloud VM)      │
└────────┬─────────┘
         │  10.8.0.0/24 (VPN subnet)
         ▼
┌──────────────────┐
│  nginx ingress   │  ← Routes admin.yourdomain.com to core-manager
│  (internal only) │
└────────┬─────────┘
         │
    ┌────┴────────────────────────┐
    │                             │
    ▼                             ▼
┌──────────┐             ┌──────────────┐
│core-mgr  │             │ auth-service  │
│:8082     │             │ :8081         │
│/admin/*  │             │ /api/v1/auth  │
│/api/v1/* │             │               │
└──────────┘             └──────────────┘
```

No public ingress for admin. VPN clients get a private IP that can reach the cluster's internal services. Everyone else sees nothing.

---

## Option A: WireGuard in the Cluster (simplest)

Deploy WireGuard as a pod in the cluster. Your laptop connects to it, gets routed into the cluster network.

### 1. Generate keys

```bash
# On your machine
wg genkey | tee admin-private.key | wg pubkey > admin-public.key
wg genkey | tee server-private.key | wg pubkey > server-public.key
```

### 2. Deploy WireGuard server

```yaml
# deployments/kustomize/services/wireguard/base/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: wireguard
  namespace: ai-persona-system
spec:
  replicas: 1
  selector:
    matchLabels:
      app: wireguard
  template:
    metadata:
      labels:
        app: wireguard
    spec:
      containers:
        - name: wireguard
          image: linuxserver/wireguard:latest
          securityContext:
            capabilities:
              add: ["NET_ADMIN", "SYS_MODULE"]
            privileged: true
          ports:
            - containerPort: 51820
              protocol: UDP
          env:
            - name: PUID
              value: "1000"
            - name: PGID
              value: "1000"
            - name: TZ
              value: "Europe/London"
          volumeMounts:
            - name: wg-config
              mountPath: /config
      volumes:
        - name: wg-config
          secret:
            secretName: wireguard-config
---
apiVersion: v1
kind: Service
metadata:
  name: wireguard
  namespace: ai-persona-system
spec:
  type: LoadBalancer  # or NodePort if no cloud LB
  ports:
    - port: 51820
      protocol: UDP
      targetPort: 51820
  selector:
    app: wireguard
```

### 3. WireGuard server config

```ini
# Create as a K8s secret
[Interface]
Address = 10.8.0.1/24
ListenPort = 51820
PrivateKey = <server-private-key>

# Allow routing to cluster service network
PostUp = iptables -A FORWARD -i wg0 -j ACCEPT; iptables -t nat -A POSTROUTING -o eth0 -j MASQUERADE
PostDown = iptables -D FORWARD -i wg0 -j ACCEPT; iptables -t nat -D POSTROUTING -o eth0 -j MASQUERADE

[Peer]
# Your laptop
PublicKey = <admin-public-key>
AllowedIPs = 10.8.0.2/32
```

```bash
kubectl -n ai-persona-system create secret generic wireguard-config \
  --from-file=wg0.conf=server-wg0.conf
```

### 4. Client config (your laptop)

```ini
# /etc/wireguard/admin.conf
[Interface]
Address = 10.8.0.2/24
PrivateKey = <admin-private-key>
DNS = 10.96.0.10  # cluster DNS (coredns) — lets you resolve *.svc.cluster.local

[Peer]
PublicKey = <server-public-key>
Endpoint = <wireguard-service-external-ip>:51820
AllowedIPs = 10.8.0.0/24, 10.96.0.0/12  # VPN subnet + cluster service CIDR
PersistentKeepalive = 25
```

```bash
sudo wg-quick up admin
# Now you can reach cluster services directly:
curl http://core-manager.ai-persona-system.svc.cluster.local:8082/health
```

### 5. Access the admin dashboard

With VPN connected:

```
http://core-manager.ai-persona-system.svc.cluster.local:8082/admin/
```

Or add an nginx ingress inside the cluster that only listens on the VPN subnet.

---

## Option B: External VM Bastion (more flexible)

A small VM (DigitalOcean $4/mo, Hetzner €3/mo) running WireGuard + nginx. The VM has a public IP and a VPN tunnel to the cluster.

```
Your laptop → WireGuard → Bastion VM → kubectl port-forward (or direct K8s network)
                              │
                              └→ nginx: admin.yourdomain.com → core-manager:8082
```

Advantages over Option A:
- Static public IP for DNS (A record for admin.yourdomain.com)
- Let's Encrypt TLS for the admin domain
- nginx rate limiting, IP allowlisting as extra layers
- Doesn't require LoadBalancer service in cluster (cheaper)

### Bastion setup (Ubuntu)

```bash
# Install WireGuard
apt install wireguard

# Generate keys
wg genkey | tee /etc/wireguard/private.key | wg pubkey > /etc/wireguard/public.key

# Configure
cat > /etc/wireguard/wg0.conf << 'EOF'
[Interface]
Address = 10.8.0.1/24
ListenPort = 51820
PrivateKey = <bastion-private-key>

[Peer]
# Your laptop
PublicKey = <admin-public-key>
AllowedIPs = 10.8.0.2/32
EOF

systemctl enable --now wg-quick@wg0

# Install nginx + certbot
apt install nginx certbot python3-certbot-nginx

# Connect to cluster via kubectl from bastion
# (install kubectl, copy kubeconfig)
# Set up a persistent port-forward or use a K8s service tunnel
```

### nginx config on bastion

```nginx
server {
    listen 443 ssl http2;
    server_name admin.yourdomain.com;

    ssl_certificate /etc/letsencrypt/live/admin.yourdomain.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/admin.yourdomain.com/privkey.pem;

    # Only allow VPN clients
    allow 10.8.0.0/24;
    deny all;

    # Rate limit
    limit_req_zone $binary_remote_addr zone=admin:10m rate=30r/m;
    limit_req zone=admin burst=10;

    # Proxy to core-manager (via kubectl port-forward or direct cluster IP)
    location /api/ {
        proxy_pass http://127.0.0.1:8082;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    location /admin/ {
        root /var/www/admin-dashboard/dist;
        try_files $uri $uri/ /admin/index.html;
    }
}

# Redirect HTTP to HTTPS
server {
    listen 80;
    server_name admin.yourdomain.com;
    return 301 https://$host$request_uri;
}
```

---

## Option C: Start with port-forward, add VPN later

For immediate use while building:

```bash
# Terminal 1: port-forward core-manager
kubectl -n ai-persona-system port-forward svc/core-manager 8082:8082

# Terminal 2: port-forward auth-service (if needed for JWT)
kubectl -n ai-persona-system port-forward svc/auth-service 8081:8081

# Access:
# API: http://localhost:8082/api/v1/admin/sites
# Dashboard: http://localhost:8082/admin/  (once static files are served)
```

This works now, needs no infrastructure, and is secure (only your machine). Move to Option A or B when you need:
- Remote access (not at your desk)
- Team access (multiple people)
- Stable URL for bookmarks/CI

---

## Security Layers

Regardless of which option you choose:

| Layer | What | Status |
|-------|------|--------|
| Network | VPN or port-forward — no public ingress | Choose above |
| Auth | JWT via auth-service RequireAuth middleware | Already built |
| Role | RequireRole("admin") on all admin routes | Already built |
| TLS | Let's Encrypt on bastion nginx (Option B) or N/A for port-forward | Option B only |
| Rate limit | nginx rate limiting (Option B) | Option B only |
| Audit | Admin actions logged via agent_error_log + standard Go logging | Already built |
| IP allow | nginx allow/deny (Option B) or WireGuard AllowedIPs | Built into VPN |

The auth system is already functional — JWT login, admin role check, gateway proxy with enriched headers. The VPN adds the network isolation layer on top.

---

## Recommended Path

1. **Now:** port-forward (Option C) while building the API integration
2. **This week:** Option A (WireGuard in cluster) for stable access
3. **When team grows:** Option B (bastion VM) for TLS + stable domain + multi-user

The admin dashboard serves from core-manager at `/admin/`. No separate deployment needed. Build the React app, copy `dist/` to the core-manager container image, serve via Gin's StaticFS.

```go
// In server.go setupRoutes, before the catch-all
router.StaticFS("/admin", http.Dir("./admin-dashboard/dist"))
```

Or mount as a ConfigMap/volume in the deployment if you want to update the dashboard without rebuilding core-manager.
