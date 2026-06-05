Find the adapter's B2 settings. First get the deployment name, then pull its three B2 env values in one shot:

kubectl -n ai-persona-system get deploy | grep -i thunder
kubectl -n ai-persona-system exec deploy/<NAME> -- \
printenv S3_ENDPOINT B2_APPLICATION_KEY_ID B2_APPLICATION_KEY TRAINING_BUCKET
(If exec won't cooperate, the values come from a secret — kubectl -n ai-persona-system get deploy <NAME> -o yaml | grep -i secretKeyRef -A2 tells you which secret/keys, then ... get secret <secret> -o jsonpath='{.data.<KEY>}' | base64 -d.)

Export them locally, in the shell where you'll run the harness:

export S3_ENDPOINT='https://s3.<region>.backblazeb2.com'
export B2_APPLICATION_KEY_ID='...'
export B2_APPLICATION_KEY='...'
# TRAINING_BUCKET is optional — defaults to personae-model-training

Run it (install the two deps if needed):

pip install boto3 requests
python3 isolation_test_phase_a.py

Paste me the output. It prints three stages: STAGE 1 tars a dummy checkpoint and PUTs it (this is the real signature/Content-Type check), STAGE 2 tars a dummy adapter while proving the checkpoints/ dir is excluded, and STAGE 3 GETs the checkpoint back and confirms it's byte-identical. "ALL STAGES PASS" means the upload/resume plumbing is sound.

