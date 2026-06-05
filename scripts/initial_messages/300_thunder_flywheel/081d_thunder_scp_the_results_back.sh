# (confirm the exact secret name + ip if the connection fails:)
#   SELECT instance_ip, ssh_user, ssh_key_secret_name FROM thunder_instances WHERE id='fabfd7fa-ac84-4476-86f3-f7ac57862214';

# 1) pull the box's SSH private key from its k8s secret
kubectl -n ai-persona-system get secret thunder-ssh-fabfd7fa-ac84-4476-86f3-f7ac57862214 \
  -o jsonpath='{.data.private_key}' | base64 -d > /tmp/iter0_key
chmod 600 /tmp/iter0_key

# 2) copy the adapter dir off the box
scp -P 30340 -i /tmp/iter0_key \
  -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null \
  -r ubuntu@216.81.200.234:/workspace/adapter_out \
  ./iter0_adapter_out


  -- check first
  ssh -p 30340 -i /tmp/iter0_key -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null \
    ubuntu@216.81.200.234 \
    'grep -E "RUN_SH_FULL_OK|RUN_SH_DONE" /workspace/train.log; ls -la /workspace/adapter_out 2>/dev/null'


loop

while true; do
  if ssh -p 30340 -i /tmp/k -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null \
       ubuntu@216.81.200.234 'test -f /workspace/adapter_out/adapter_config.json'; then
    echo "adapter present — copying"
    scp -P 30340 -i /tmp/k -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null \
      -r ubuntu@216.81.200.234:/workspace/adapter_out ./iter0_adapter_out
    break
  fi
  echo "not ready $(date -u +%H:%M:%S) — sleeping"; sleep 300
done


The loop only breaks on success — it has no exit for a crash. If 02_train were to die, adapter_config.json would never appear and the loop would poll indefinitely. Given the run's health that's unlikely, but if it ever goes quiet far past the expected finish, a one-off grep RUN_SH_FATAL /workspace/train.log plus a tail tells you whether to keep waiting or investigate.
And once the copy lands, verify it's the real adapter before trusting it — a quick ls -la ./iter0_adapter_out should show adapter_config.json and adapter_model.safetensors (the safetensors file being a few hundred MB), not just the checkpoints/ dir. After that the box is safe to reconcile/decommission, and 1cd65dd7 can move off running — which the monitor will now do correctly once it's enabled, or you can do by hand in the meantime.

