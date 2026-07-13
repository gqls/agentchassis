#!/usr/bin/env python3
"""
isolation_test_phase_a.py — box-free Tier-1 test of the Phase A upload/resume
mechanics from 02_train_llama_3_3_70b.py, with NO GPU and NO training.

It presigns PUT/GET URLs directly against B2 (the same creds the thunder-adapter
uses), then exercises faithful copies of 02_train's _tar_dir / _put_file /
_download_and_extract against dummy directories. This validates the riskiest
mechanics — the presigned-PUT signature + Content-Type, the tar round-trip, the
checkpoints/ exclusion, and GET+extract — before spending any GPU time.

The helper bodies below are deliberate COPIES of the ones in 02_train: this harness
must not import 02_train, because that triggers `from unsloth import ...`, which needs
the GPU stack. If those helpers are later split into an importable module, point this
harness at that module instead of copying.

Env (pull from the thunder-adapter deployment/secret — see the chat for kubectl):
  B2_APPLICATION_KEY_ID, B2_APPLICATION_KEY, S3_ENDPOINT
  TRAINING_BUCKET   (optional; default personae-model-training)
  S3_REGION         (optional; default parsed from S3_ENDPOINT, else us-west-004)

Requires:  pip install boto3 requests
Run:       python3 isolation_test_phase_a.py
"""
from __future__ import annotations

import json
import os
import re
import shutil
import sys
import tarfile
import time
from pathlib import Path

import requests
import boto3
from botocore.client import Config


# ── copies of 02_train's Phase A helpers (keep in sync) ──────────────────────
_OCTET = "application/octet-stream"
_HTTP_TIMEOUT = (30, 1800)


def _tar_dir(src_dir: Path, dest_tar: Path, exclude_top: set[str] | None = None) -> None:
    exclude_top = exclude_top or set()

    def _filter(ti: tarfile.TarInfo):
        parts = ti.name.split("/")
        if len(parts) >= 2 and parts[1] in exclude_top:
            return None
        return ti

    with tarfile.open(dest_tar, "w:gz") as tar:
        tar.add(str(src_dir), arcname=src_dir.name, filter=_filter)


def _put_file(url: str, path: Path) -> None:
    size = path.stat().st_size
    with open(path, "rb") as fh:
        resp = requests.put(url, data=fh, headers={"Content-Type": _OCTET}, timeout=_HTTP_TIMEOUT)
    if resp.status_code not in (200, 201):
        raise RuntimeError(f"PUT {resp.status_code}: {resp.text[:500]}")
    print(f"    uploaded {path.name} ({size / 1e6:.2f}MB) -> HTTP {resp.status_code}")


def _download_and_extract(url: str, into_dir: Path, workdir: Path) -> None:
    into_dir.mkdir(parents=True, exist_ok=True)
    tmp = workdir / "resume.tar.gz"
    with requests.get(url, stream=True, timeout=_HTTP_TIMEOUT) as resp:
        if resp.status_code != 200:
            raise RuntimeError(f"GET {resp.status_code}: {resp.text[:500]}")
        with open(tmp, "wb") as fh:
            for chunk in resp.iter_content(chunk_size=8 << 20):
                fh.write(chunk)
    with tarfile.open(tmp, "r:gz") as tar:
        try:
            tar.extractall(str(into_dir), filter="data")
        except TypeError:
            tar.extractall(str(into_dir))
    tmp.unlink(missing_ok=True)


# ── presign (boto3 against B2's S3-compatible endpoint) ──────────────────────
def _region_from_endpoint(ep: str) -> str:
    m = re.search(r"s3\.([a-z0-9-]+)\.backblazeb2\.com", ep or "")
    return m.group(1) if m else os.environ.get("S3_REGION", "us-west-004")


def _s3():
    ep = os.environ["S3_ENDPOINT"]
    return boto3.client(
        "s3",
        endpoint_url=ep,
        aws_access_key_id=os.environ["B2_APPLICATION_KEY_ID"],
        aws_secret_access_key=os.environ["B2_APPLICATION_KEY"],
        region_name=_region_from_endpoint(ep),
        # s3v4 + path-style is what B2's S3 endpoint expects for presigned URLs.
        config=Config(signature_version="s3v4", s3={"addressing_style": "path"}),
    )


def _presign_put(s3, bucket: str, key: str, expires: int = 3600) -> str:
    # NOTE: no ContentType in Params — this mirrors the adapter's GetPresignedPutURL,
    # so the signature is NOT content-type-bound and _put_file's octet-stream header is
    # accepted. If you presign WITH a ContentType, the PUT must send that exact value.
    return s3.generate_presigned_url(
        "put_object", Params={"Bucket": bucket, "Key": key}, ExpiresIn=expires, HttpMethod="PUT"
    )


def _presign_get(s3, bucket: str, key: str, expires: int = 3600) -> str:
    return s3.generate_presigned_url(
        "get_object", Params={"Bucket": bucket, "Key": key}, ExpiresIn=expires
    )


# ── dummy fixtures ───────────────────────────────────────────────────────────
def _write_dummy_checkpoint(d: Path) -> None:
    d.mkdir(parents=True, exist_ok=True)
    (d / "optimizer.pt").write_bytes(os.urandom(2 * 1024 * 1024))      # ~2MB stand-in
    (d / "trainer_state.json").write_text(json.dumps({"global_step": 50}))
    (d / "config.json").write_text(json.dumps({"dummy": True}))


def _write_dummy_adapter(d: Path) -> None:
    d.mkdir(parents=True, exist_ok=True)
    (d / "adapter_config.json").write_text(json.dumps({"r": 16}))
    (d / "adapter_model.safetensors").write_bytes(os.urandom(1024 * 1024))  # ~1MB stand-in
    (d / "manifest.json").write_text(json.dumps({"final_loss": 0.21}))
    # a checkpoints/ subdir that MUST be excluded from the final tarball:
    ck = d / "checkpoints" / "checkpoint-50"
    ck.mkdir(parents=True, exist_ok=True)
    (ck / "optimizer.pt").write_bytes(os.urandom(512 * 1024))


def main() -> int:
    for v in ("B2_APPLICATION_KEY_ID", "B2_APPLICATION_KEY", "S3_ENDPOINT"):
        if not os.environ.get(v):
            print(f"FAIL: env {v} not set. Pull it from the thunder-adapter (see header).")
            return 2

    bucket = os.environ.get("TRAINING_BUCKET", "personae-model-training")
    prefix = f"finetuning/_isolation_test/{int(time.time())}"
    work = Path("./_iso")
    if work.exists():
        shutil.rmtree(work)
    work.mkdir(parents=True)

    print(f"bucket={bucket}  endpoint={os.environ['S3_ENDPOINT']}")
    print(f"region={_region_from_endpoint(os.environ['S3_ENDPOINT'])}  prefix={prefix}")
    s3 = _s3()
    ckpt_key = f"{prefix}/ckpt-0.tar.gz"
    final_key = f"{prefix}/adapter.tar.gz"
    ckpt_put = _presign_put(s3, bucket, ckpt_key)
    final_put = _presign_put(s3, bucket, final_key)
    ckpt_get = _presign_get(s3, bucket, ckpt_key)
    print("presigned 2 PUT + 1 GET URLs\n")

    # ── Stage 1: checkpoint tar + PUT (the signature / Content-Type check) ──
    print("STAGE 1  checkpoint tar + PUT")
    ck_src = work / "ckpt_src" / "checkpoint-50"
    _write_dummy_checkpoint(ck_src)
    ck_tar = work / "ckpt-0.tar.gz"
    _tar_dir(ck_src, ck_tar)
    _put_file(ckpt_put, ck_tar)
    print("  STAGE 1 PASS\n")

    # ── Stage 2: final adapter tar with checkpoints/ EXCLUDED + PUT ──
    print("STAGE 2  final adapter tar (exclude checkpoints/) + PUT")
    ad_src = work / "adapter_out"
    _write_dummy_adapter(ad_src)
    ad_tar = work / "adapter.tar.gz"
    _tar_dir(ad_src, ad_tar, exclude_top={"checkpoints"})
    with tarfile.open(ad_tar, "r:gz") as t:
        names = t.getnames()
    leaked = [n for n in names if n.split("/")[1:2] == ["checkpoints"]]
    if leaked:
        print(f"  STAGE 2 FAIL: checkpoints/ leaked into final tar: {leaked[:3]}")
        return 1
    print(f"  exclusion OK ({len(names)} members, none under checkpoints/)")
    _put_file(final_put, ad_tar)
    print("  STAGE 2 PASS\n")

    # ── Stage 3: resume GET + extract round-trips the checkpoint back ──
    print("STAGE 3  resume GET + extract")
    dst = work / "resume_dst"
    _download_and_extract(ckpt_get, dst, work)
    got = dst / "checkpoint-50" / "optimizer.pt"
    if not got.is_file():
        print(f"  STAGE 3 FAIL: {got} missing after extract")
        return 1
    if got.read_bytes() != (ck_src / "optimizer.pt").read_bytes():
        print("  STAGE 3 FAIL: extracted optimizer.pt bytes differ from source")
        return 1
    print("  round-trip byte-identical")
    print("  STAGE 3 PASS\n")

    print("ALL STAGES PASS — Phase A upload/resume mechanics are sound against B2.")
    print(f"(test objects left under s3://{bucket}/{prefix}/ — delete when done)")
    return 0


if __name__ == "__main__":
    sys.exit(main())
