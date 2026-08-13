# Database user passwords (not root)
variable "auth_db_user_password" {
  description = "Password for auth service database user"
  type        = string
  sensitive   = true
}

variable "templates_db_user_password" {
  description = "Password for templates database user"
  type        = string
  sensitive   = true
}

variable "clients_db_user_password" {
  description = "Password for clients database user"
  type        = string
  sensitive   = true
}

# PgBouncer admin console user (bugs_open/246, owner decision D1 2026-08-11).
#
# NOT a PostgreSQL user. `pgbouncer_admin` exists only inside PgBouncer, as its
# `admin_users`/`stats_users` (pgbouncer-configmap.yaml:73-74), and it is what
# `SHOW POOLS` / `SHOW CLIENTS` require. Without it, client-side pool queueing
# (`cl_waiting`, `maxwait`) is unmeasurable by anything: `pg_stat_activity` cannot
# substitute, because every row's client_addr is PgBouncer itself, so Postgres
# sees the server pool and never the client queue.
#
# ⚠ THIS VALUE IS ONE HALF OF A PAIR THAT MUST AGREE.
# PgBouncer authenticates the console against `/etc/pgbouncer/userlist.txt`
# (`auth_type = md5`), mounted from the `pgbouncer-userlist` secret, which
# Terraform does NOT manage — it is the hand-applied kustomize manifest
# `deployments/kustomize/services/pgbouncer/pgbouncer-secret.yaml`, whose repo
# copy carries the literal placeholder `PGBOUNCER_ADMIN_PASSWORD_HERE`.
# Setting this variable alone does NOT make `SHOW POOLS` work; the userlist entry
# must carry the same password. Activation runbook:
#   docs/agent_docs/docs024_key_docs_latest/bugfix_246_shared_pool_ownership/RUNBOOK_shared_pool_ownership.md §9
#
# That two-owners-of-one-value shape is the same defect class as bugs_open/246
# itself. Registered as a follow-up rather than fixed here: having Terraform
# render the whole userlist would make the pair unrepresentable, but it needs the
# clients_user/templates_user passwords moved into the same resource and is a
# bigger change than this decision authorised.
# ⚠⚠ THE VALUE CURRENTLY IN terraform.tfvars.secret DOES NOT WORK. Corrected 2026-08-13.
#
# When this variable was added (2026-08-12) I GENERATED a fresh 32-character password.
# That was the wrong move: `pgbouncer_admin` **already had a working 20-character password**
# in the `pgbouncer-userlist` secret, and pgbouncer authenticates against that file alone
# (`auth_type = md5`, `auth_file` only — there is no `auth_query`/`auth_user`). So the two
# halves now hold different strings and the console refuses:
#
#     psql -U pgbouncer_admin -d pgbouncer -c "SHOW POOLS;"
#     FATAL:  password authentication failed
#
# (Distinguish that from `FATAL: not allowed`, which is what a NON-admin such as
# clients_user gets. "not allowed" = wrong user; "password authentication failed" = right
# user, wrong password. Only the value is out of step; the admin roster is correct.)
#
# THE FIX IS TO RECORD, NOT TO IMPOSE — replace the generated value in
# terraform.tfvars.secret with the EXISTING userlist value and re-apply. That needs no
# pgbouncer restart (the userlist is untouched, and pgbouncer already accepts it) and puts
# nothing at risk for clients_user/templates_user, whose lines are never rewritten. The
# alternative — writing this value INTO the userlist — costs a restart that drops every
# pooled connection fleet-wide, to no benefit.
#
# UNTIL THAT HAPPENS: the `PGBOUNCER_ADMIN_PASSWORD` key in `personae-platform-secrets` is
# LIVE AND WRONG. Do not read it and conclude you hold the admin password — you hold a
# string that authenticates nothing. Full account and both wrong calls:
#   docs/agent_docs/docs024_key_docs_latest/bugfix_246_shared_pool_ownership/RUNBOOK_shared_pool_ownership.md §9
variable "pgbouncer_admin_password" {
  description = "Password for the pgbouncer_admin console user (SHOW POOLS/SHOW CLIENTS). MUST be the EXISTING value from the pgbouncer-userlist secret — record it, do not generate one. As of 2026-08-13 the value in tfvars is a generated string that does NOT authenticate; see RUNBOOK §9."
  type        = string
  sensitive   = true
}

# Platform keys
variable "jwt_secret_key" {
  description = "JWT signing key for auth-service"
  type        = string
  sensitive   = true
}

variable "agent_bootstrap_key" {
  description = "Bootstrap key for platform agents"
  type        = string
  sensitive   = true
}

# Default API keys (temporary until per-user keys)
variable "default_anthropic_api_key" {
  description = "Default Anthropic API key"
  type        = string
  sensitive   = true
}

variable "default_stability_api_key" {
  description = "Default Stability API key"
  type        = string
  sensitive   = true
}

variable "default_banana_api_key" {
  description = "Default Gemini Banana API key"
  type        = string
  sensitive   = true
}

variable "default_gemini_content_api_key" {
  description = "Default Gemini Content API key"
  type        = string
  sensitive   = true
}

variable "default_serp_api_key" {
  description = "Default SERP API key"
  type        = string
  sensitive   = true
}

variable "default_scraping_bee_api_key" {
  description = "Default SCRAPING BEE API key"
  type        = string
  sensitive   = true
}

variable "default_firecrawl_api_key" {
  description = "Default FIRECRAWL API key"
  type        = string
  sensitive   = true
}

variable "default_perplexity_api_key" {
  description = "Default PERPLEXITY API key"
  type        = string
  sensitive   = true
}

variable "default_companies_house_api_key" {
  description = "Default COMPANIES HOUSE API key"
  type        = string
  sensitive   = true
}

variable "default_thunder_compute_api_key" {
  description = "Default THUNDER_COMPUTE_API_KEY key"
  type        = string
  sensitive   = true
}

variable "default_grok_api_key" {
  description = "Default GROK_API_KEY key"
  type        = string
  sensitive   = true
}

# Docker
variable "docker_password" {
  description = "Docker Hub password"
  type        = string
  sensitive   = true
}

variable "docker_email" {
  description = "Docker Hub email"
  type        = string
}

variable "github_personal_access_token" {
    description = "Github personal access token"
    type = string
    sensitive = true
}

variable "github_read_token" {
    description = " Github personal access token for doc analyser READ ONLY"
    type = string
    sensitive = true
}

variable "auth_db_host" {
  description = "The endpoint for the external MySQL database."
  type        = string
  default     = "rs17.uk-noc.com"
}

variable "auth_db_name" {
  description = "The name of the external MySQL database."
  type        = string
  default     = "catalogu_vectordb_chassis"
}

variable "auth_db_user" {
  description = "Username for the external MySQL database."
  type        = string
  default     = "catalogu_personae"
}

variable "agent_image_tag" {
  description = "Docker image tag for agent chassis"
  type        = string
  default     = "v1.0.44"
}