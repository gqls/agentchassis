The chain of things that hold the password:

1. **PostgreSQL** itself (the user's password)
2. **`personae-platform-secrets`** K8s secret (what agents read)
3. **`pgbouncer-userlist`** K8s secret (what PgBouncer uses to authenticate clients)

Order matters. Safest sequence:

```bash
# 1. Generate new passwords
NEW_CLIENTS_PW=$(openssl rand -base64 24 | tr -d '=/+' | head -c 32)
NEW_TEMPLATES_PW=$(openssl rand -base64 24 | tr -d '=/+' | head -c 32)
echo "clients: $NEW_CLIENTS_PW"
echo "templates: $NEW_TEMPLATES_PW"
# Save these somewhere safe before proceeding

# 2. Change passwords in PostgreSQL (existing connections keep working)
kubectl -n ai-persona-system exec -it postgres-clients-0 -- \
  psql -U postgres -d clients_db -c "ALTER USER clients_user PASSWORD '$NEW_CLIENTS_PW';"

kubectl -n ai-persona-system exec -it postgres-templates-0 -- \
  psql -U postgres -d templates_db -c "ALTER USER templates_user PASSWORD '$NEW_TEMPLATES_PW';"

# 3. Update the K8s secret that agents and PgBouncer read from
kubectl -n ai-persona-system create secret generic personae-platform-secrets \
  --from-literal=CLIENTS_DB_PASSWORD="$NEW_CLIENTS_PW" \
  --from-literal=TEMPLATES_DB_PASSWORD="$NEW_TEMPLATES_PW" \
  --from-literal=AUTH_DB_PASSWORD="$(kubectl -n ai-persona-system get secret personae-platform-secrets -o jsonpath='{.data.AUTH_DB_PASSWORD}' | base64 -d)" \
  --dry-run=client -o yaml | kubectl apply -f -

# 4. Rebuild PgBouncer userlist (reads from personae-platform-secrets)
make deploy-065-pgbouncer ENVIRONMENT=production REGION=uk001

# 5. Restart PgBouncer to pick up new userlist
make pgbouncer-restart ENVIRONMENT=production REGION=uk001

# 6. Test PgBouncer
make pgbouncer-test ENVIRONMENT=production REGION=uk001

# 7. Restart agent pods to pick up new password from secret
kubectl -n ai-persona-system rollout restart deployment/agent-chassis
```

Step 2 changes what PostgreSQL accepts. Existing connections (using the old password) stay alive until they disconnect. Steps 3-5 update PgBouncer. Step 7 makes agents reconnect with the new password.

One thing to watch — step 3 only preserves the three `_DB_PASSWORD` keys. If `personae-platform-secrets` has other keys, you'll need to include them too. Check first with:

```bash
kubectl -n ai-persona-system get secret personae-platform-secrets -o json | jq '.data | keys'
```