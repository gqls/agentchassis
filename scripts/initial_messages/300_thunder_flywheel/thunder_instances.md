ant@ant-XPS-15-9500:~/projects/agentchassis/scripts/utils$ THUNDER_COMPUTE_API_KEY=a7383d1885fa832cc6f30674af5e327ddc9fd1ca0c912eb5e0ac1ce925b12596
ant@ant-XPS-15-9500:~/projects/agentchassis/scripts/utils$ cd /path/to/thunder_probe   # wherever you save it

# Your shell already has a working token (tnr works), but the CLI may store it
# in a config file rather than env. Confirm it's in env, or pull it from the secret:
echo "len: ${#THUNDER_COMPUTE_API_KEY}"
# If 0, get it from the cluster:
export THUNDER_COMPUTE_API_KEY=$(kubectl -n ai-persona-system get secret personae-default-secrets \
-o jsonpath='{.data.THUNDER_COMPUTE_API_KEY}' | base64 -d)

go run thunder_probe.go
bash: cd: /path/to/thunder_probe: No such file or directory
len: 64
==================== GET /specs ====================
STATUS: 200
RESPONSE:
{
"specs": {
"a100xl_x1_production": {
"displayName": "NVIDIA A100 (80GB)",
"vramGB": 80,
"gpuCount": 1,
"mode": "production",
"vcpuOptions": [
15
],
"ramPerVCPUGiB": 8,
"storageGB": {
"min": 100,
"max": 300
},
"ephemeralStorageGB": {
"min": 0,
"max": 500
}
},
"a100xl_x1_prototyping": {
"displayName": "NVIDIA A100 (80GB)",
"vramGB": 80,
"gpuCount": 1,
"mode": "prototyping",
"vcpuOptions": [
4,
8,
12
],
"ramPerVCPUGiB": 8,
"storageGB": {
"min": 100,
"max": 300
},
"ephemeralStorageGB": {
"min": 0,
"max": 200
}
},
"a100xl_x2_production": {
"displayName": "NVIDIA A100 (80GB)",
"vramGB": 80,
"gpuCount": 2,
"mode": "production",
"vcpuOptions": [
30
],
"ramPerVCPUGiB": 8,
"storageGB": {
"min": 100,
"max": 300
},
"ephemeralStorageGB": {
"min": 0,
"max": 1000
}
},
"a100xl_x2_prototyping": {
"displayName": "NVIDIA A100 (80GB)",
"vramGB": 80,
"gpuCount": 2,
"mode": "prototyping",
"vcpuOptions": [
8,
12,
16,
20,
24
],
"ramPerVCPUGiB": 8,
"storageGB": {
"min": 100,
"max": 300
},
"ephemeralStorageGB": {
"min": 0,
"max": 400
}
},
"a100xl_x4_production": {
"displayName": "NVIDIA A100 (80GB)",
"vramGB": 80,
"gpuCount": 4,
"mode": "production",
"vcpuOptions": [
60
],
"ramPerVCPUGiB": 8,
"storageGB": {
"min": 100,
"max": 500
},
"ephemeralStorageGB": {
"min": 0,
"max": 2000
}
},
"a100xl_x8_production": {
"displayName": "NVIDIA A100 (80GB)",
"vramGB": 80,
"gpuCount": 8,
"mode": "production",
"vcpuOptions": [
120
],
"ramPerVCPUGiB": 8,
"storageGB": {
"min": 100,
"max": 500
},
"ephemeralStorageGB": {
"min": 0,
"max": 4000
}
},
"a6000_x1_prototyping": {
"displayName": "RTX A6000",
"vramGB": 48,
"gpuCount": 1,
"mode": "prototyping",
"vcpuOptions": [
4,
6
],
"ramPerVCPUGiB": 8,
"storageGB": {
"min": 100,
"max": 300
},
"ephemeralStorageGB": {
"min": 0,
"max": 0
}
},
"h100_x1_production": {
"displayName": "NVIDIA H100",
"vramGB": 80,
"gpuCount": 1,
"mode": "production",
"vcpuOptions": [
15
],
"ramPerVCPUGiB": 8,
"storageGB": {
"min": 100,
"max": 300
},
"ephemeralStorageGB": {
"min": 0,
"max": 500
}
},
"h100_x1_prototyping": {
"displayName": "NVIDIA H100",
"vramGB": 80,
"gpuCount": 1,
"mode": "prototyping",
"vcpuOptions": [
4,
8,
12,
16
],
"ramPerVCPUGiB": 8,
"storageGB": {
"min": 100,
"max": 300
},
"ephemeralStorageGB": {
"min": 0,
"max": 200
}
},
"h100_x2_production": {
"displayName": "NVIDIA H100",
"vramGB": 80,
"gpuCount": 2,
"mode": "production",
"vcpuOptions": [
30
],
"ramPerVCPUGiB": 8,
"storageGB": {
"min": 100,
"max": 300
},
"ephemeralStorageGB": {
"min": 0,
"max": 1000
}
},
"h100_x2_prototyping": {
"displayName": "NVIDIA H100",
"vramGB": 80,
"gpuCount": 2,
"mode": "prototyping",
"vcpuOptions": [
8,
12,
16,
20,
24
],
"ramPerVCPUGiB": 8,
"storageGB": {
"min": 100,
"max": 300
},
"ephemeralStorageGB": {
"min": 0,
"max": 400
}
},
"h100_x4_production": {
"displayName": "NVIDIA H100",
"vramGB": 80,
"gpuCount": 4,
"mode": "production",
"vcpuOptions": [
60
],
"ramPerVCPUGiB": 8,
"storageGB": {
"min": 100,
"max": 500
},
"ephemeralStorageGB": {
"min": 0,
"max": 2000
}
},
"h100_x8_production": {
"displayName": "NVIDIA H100",
"vramGB": 80,
"gpuCount": 8,
"mode": "production",
"vcpuOptions": [
120
],
"ramPerVCPUGiB": 8,
"storageGB": {
"min": 100,
"max": 500
},
"ephemeralStorageGB": {
"min": 0,
"max": 4000
}
},
"l40_x1_production": {
"displayName": "NVIDIA L40",
"vramGB": 48,
"gpuCount": 1,
"mode": "production",
"vcpuOptions": [
10
],
"ramPerVCPUGiB": 8,
"storageGB": {
"min": 100,
"max": 300
},
"ephemeralStorageGB": {
"min": 0,
"max": 500
}
},
"l40_x1_prototyping": {
"displayName": "NVIDIA L40",
"vramGB": 48,
"gpuCount": 1,
"mode": "prototyping",
"vcpuOptions": [
4,
8
],
"ramPerVCPUGiB": 8,
"storageGB": {
"min": 100,
"max": 300
},
"ephemeralStorageGB": {
"min": 0,
"max": 200
}
},
"l40_x2_production": {
"displayName": "NVIDIA L40",
"vramGB": 48,
"gpuCount": 2,
"mode": "production",
"vcpuOptions": [
20
],
"ramPerVCPUGiB": 8,
"storageGB": {
"min": 100,
"max": 300
},
"ephemeralStorageGB": {
"min": 0,
"max": 1000
}
},
"l40_x2_prototyping": {
"displayName": "NVIDIA L40",
"vramGB": 48,
"gpuCount": 2,
"mode": "prototyping",
"vcpuOptions": [
8,
12,
16
],
"ramPerVCPUGiB": 8,
"storageGB": {
"min": 100,
"max": 300
},
"ephemeralStorageGB": {
"min": 0,
"max": 400
}
},
"l40_x4_production": {
"displayName": "NVIDIA L40",
"vramGB": 48,
"gpuCount": 4,
"mode": "production",
"vcpuOptions": [
40
],
"ramPerVCPUGiB": 8,
"storageGB": {
"min": 100,
"max": 500
},
"ephemeralStorageGB": {
"min": 0,
"max": 2000
}
},
"l40_x8_production": {
"displayName": "NVIDIA L40",
"vramGB": 48,
"gpuCount": 8,
"mode": "production",
"vcpuOptions": [
80
],
"ramPerVCPUGiB": 8,
"storageGB": {
"min": 100,
"max": 500
},
"ephemeralStorageGB": {
"min": 0,
"max": 4000
}
},
"l40s_x1_production": {
"displayName": "NVIDIA L40S",
"vramGB": 48,
"gpuCount": 1,
"mode": "production",
"vcpuOptions": [
10
],
"ramPerVCPUGiB": 8,
"storageGB": {
"min": 100,
"max": 300
},
"ephemeralStorageGB": {
"min": 0,
"max": 500
}
},
"l40s_x1_prototyping": {
"displayName": "NVIDIA L40S",
"vramGB": 48,
"gpuCount": 1,
"mode": "prototyping",
"vcpuOptions": [
4,
8
],
"ramPerVCPUGiB": 8,
"storageGB": {
"min": 100,
"max": 300
},
"ephemeralStorageGB": {
"min": 0,
"max": 200
}
},
"l40s_x2_production": {
"displayName": "NVIDIA L40S",
"vramGB": 48,
"gpuCount": 2,
"mode": "production",
"vcpuOptions": [
20
],
"ramPerVCPUGiB": 8,
"storageGB": {
"min": 100,
"max": 300
},
"ephemeralStorageGB": {
"min": 0,
"max": 1000
}
},
"l40s_x2_prototyping": {
"displayName": "NVIDIA L40S",
"vramGB": 48,
"gpuCount": 2,
"mode": "prototyping",
"vcpuOptions": [
8,
12,
16
],
"ramPerVCPUGiB": 8,
"storageGB": {
"min": 100,
"max": 300
},
"ephemeralStorageGB": {
"min": 0,
"max": 400
}
},
"l40s_x4_production": {
"displayName": "NVIDIA L40S",
"vramGB": 48,
"gpuCount": 4,
"mode": "production",
"vcpuOptions": [
40
],
"ramPerVCPUGiB": 8,
"storageGB": {
"min": 100,
"max": 500
},
"ephemeralStorageGB": {
"min": 0,
"max": 2000
}
},
"l40s_x8_production": {
"displayName": "NVIDIA L40S",
"vramGB": 48,
"gpuCount": 8,
"mode": "production",
"vcpuOptions": [
80
],
"ramPerVCPUGiB": 8,
"storageGB": {
"min": 100,
"max": 500
},
"ephemeralStorageGB": {
"min": 0,
"max": 4000
}
}
}
}

==================== GET /instances/list ====================
STATUS: 200
RESPONSE:
{}
