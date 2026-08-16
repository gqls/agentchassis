#!/usr/bin/env python3
"""Deploy the scripts bundle to B2 (FTW-031: re-uploading the object IS the
deploy). PUTs the tarball to the exact key the launcher presigns, then reads
it straight back and compares md5 — the round-trip is the proof, not the 200."""
import hashlib, os, sys
import boto3
from botocore.config import Config

src = sys.argv[1]
key = "finetuning/scripts/bundle.tar.gz"
bucket = os.environ.get("TRAINING_BUCKET", "personae-model-training")
s3 = boto3.client("s3", endpoint_url=os.environ["S3_ENDPOINT"],
    aws_access_key_id=os.environ["B2_KEY_ID"], aws_secret_access_key=os.environ["B2_KEY"],
    config=Config(signature_version="s3v4"), region_name="us-east-005")

local = hashlib.md5(open(src, "rb").read()).hexdigest()
s3.upload_file(src, bucket, key)
body = s3.get_object(Bucket=bucket, Key=key)["Body"].read()
remote = hashlib.md5(body).hexdigest()
print(f"local  {local}\nremote {remote}\n{'ROUND-TRIP OK' if local == remote else 'MISMATCH'}  ({len(body)} bytes at {bucket}/{key})")
sys.exit(0 if local == remote else 1)
