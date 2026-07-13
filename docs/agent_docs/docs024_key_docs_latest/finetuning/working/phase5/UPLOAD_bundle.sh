# Upload the training scripts bundle to B2
#
# Target key MUST match the migration's presign_scripts step:
#   bucket: personae-model-training
#   key:    finetuning/scripts/bundle.tar.gz
#
# The adapter presigns a GET for this exact object; the launcher's ssh_exec
# curls that presigned URL on the VM. The VM holds NO B2 creds — only the
# time-limited presigned URL.
#
# Use the REAL b2 CLI (pip), NOT the snap (which is a BBC Micro emulator).

# ── 0. Get the bundle onto a machine with b2 access ────────────────────────
# (download bundle.tar.gz from this session's outputs first, then:)

# ── 1. Install the real b2 CLI in a throwaway venv ─────────────────────────
python3 -m venv /tmp/b2venv
source /tmp/b2venv/bin/activate
pip install --quiet b2

# ── 2. Authorise with the storage creds (from k8s secret personae-storage-secrets) ──
# These are the B2 application key id / key used by the adapter + preparer.
export B2_APPLICATION_KEY_ID='0052322485eed970000000008'
export B2_APPLICATION_KEY='K005b+FUQnTk1qjSSsEnBgN/Bdfpyrs'
b2 account authorize "${B2_APPLICATION_KEY_ID}" "${B2_APPLICATION_KEY}"

# ── 3. Upload the bundle to the exact key the migration expects ────────────
b2 file upload \
  personae-model-training \
  ./bundle.tar.gz \
  finetuning/scripts/bundle.tar.gz

# ── 4. Verify it's there ───────────────────────────────────────────────────
b2 ls --long personae-model-training finetuning/scripts/
# Expect a row for finetuning/scripts/bundle.tar.gz

deactivate

# ── Notes ──────────────────────────────────────────────────────────────────
# - To update the launch chain later (e.g. tweak smoke params, add a step),
#   edit run.sh, rebuild the tarball, re-upload to the SAME key. No DB
#   migration and no chassis redeploy needed — the workflow just re-presigns
#   and re-fetches the new bundle on the next run.
# - If your installed b2 CLI is an older major version, the subcommands differ:
#     v3+:  b2 account authorize / b2 file upload / b2 ls
#     v2:   b2 authorize-account / b2 upload-file / b2 ls
#   Adjust to match `b2 version`.
