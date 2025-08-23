-- FILE: platform/database/migrations/003_storage_env_references.sql
-- Migration: Define environment variable references for storage-enabled agents
-- Date: 2025-08-21
-- Description: Updates agent definitions to reference storage environment variables
--              that will be injected by Kubernetes at runtime

BEGIN;

-- Update agent definitions to reference environment variables from Kubernetes secrets/configmaps
-- These are REFERENCES, not values - actual values come from Kubernetes
UPDATE agent_definitions
SET env_vars = COALESCE(env_vars, '[]'::jsonb) || '[
  {
    "name": "AWS_ACCESS_KEY_ID",
    "valueFrom": {
      "secretKeyRef": {
        "name": "personae-storage-secrets",
        "key": "AWS_ACCESS_KEY_ID"
      }
    }
  },
  {
    "name": "AWS_SECRET_ACCESS_KEY",
    "valueFrom": {
      "secretKeyRef": {
        "name": "personae-storage-secrets",
        "key": "AWS_SECRET_ACCESS_KEY"
      }
    }
  },
  {
    "name": "S3_ENDPOINT",
    "valueFrom": {
      "configMapKeyRef": {
        "name": "storage-config",
        "key": "S3-ENDPOINT"
      }
    }
  },
  {
    "name": "S3_REGION",
    "valueFrom": {
      "configMapKeyRef": {
        "name": "storage-config",
        "key": "S3-REGION"
      }
    }
  },
  {
    "name": "IMAGE_BUCKET",
    "valueFrom": {
      "configMapKeyRef": {
        "name": "storage-config",
        "key": "image_bucket"
      }
    }
  },
  {
    "name": "ASSETS_BUCKET",
    "valueFrom": {
      "configMapKeyRef": {
        "name": "storage-config",
        "key": "assets_bucket"
      }
    }
  }
]'::jsonb
WHERE type IN (
    'site-publisher',
    'html-developer',
    'visual-designer',
    'image-generator',
    'content-creator'
    )
  AND (env_vars IS NULL OR NOT env_vars::text LIKE '%AWS_ACCESS_KEY_ID%');

-- Update default_config to indicate storage capability (no hardcoded values)
UPDATE agent_definitions
SET default_config = COALESCE(default_config, '{}'::jsonb) || '{
  "capabilities": {
    "storage": {
      "enabled": true,
      "provider": "s3",
      "access_key_env_var": "AWS_ACCESS_KEY_ID",
      "secret_key_env_var": "AWS_SECRET_ACCESS_KEY",
      "endpoint_env_var": "S3_ENDPOINT",
      "region_env_var": "S3_REGION",
      "bucket_env_var": "ASSETS_BUCKET"
    }
  }
}'::jsonb
WHERE type IN (
    'site-publisher',
    'html-developer',
    'visual-designer',
    'image-generator',
    'content-creator'
    )
  AND (
    default_config IS NULL
   OR default_config->'capabilities'->>'storage' IS NULL
    );

-- Site publisher specific configuration
UPDATE agent_definitions
SET default_config = default_config || '{
  "publishing": {
    "index_file": "index.html",
    "enable_versioning": true,
    "cache_control": "public, max-age=3600"
  }
}'::jsonb
WHERE type = 'site-publisher'
  AND (default_config->>'publishing' IS NULL);

-- Update timestamps
UPDATE agent_definitions
SET updated_at = NOW()
WHERE type IN (
               'site-publisher',
               'html-developer',
               'visual-designer',
               'image-generator',
               'content-creator'
    );

COMMIT;