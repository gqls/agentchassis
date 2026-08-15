#!/usr/bin/env python3
"""Rebuild of the 08-12 session's presign.py (concierge equivalent of
prepare_object_url). Mints presigned GET/PUT URLs against the B2 bucket the
thunder-adapter uses, with creds pulled live from the pod env by the caller.

Usage: presign.py <GET|PUT> <key> [expiry_minutes]
Env:   B2_KEY_ID, B2_KEY, S3_ENDPOINT (all required), TRAINING_BUCKET
"""
import os, sys
import boto3
from botocore.config import Config

method, key = sys.argv[1].upper(), sys.argv[2]
expiry = int(sys.argv[3]) if len(sys.argv) > 3 else 240  # minutes

s3 = boto3.client(
    "s3",
    endpoint_url=os.environ["S3_ENDPOINT"],
    aws_access_key_id=os.environ["B2_KEY_ID"],
    aws_secret_access_key=os.environ["B2_KEY"],
    config=Config(signature_version="s3v4"),
    region_name=os.environ.get("S3_REGION", "us-east-005"),
)
bucket = os.environ.get("TRAINING_BUCKET", "personae-model-training")
op = "get_object" if method == "GET" else "put_object"
url = s3.generate_presigned_url(op, Params={"Bucket": bucket, "Key": key}, ExpiresIn=expiry * 60)
print(url)
