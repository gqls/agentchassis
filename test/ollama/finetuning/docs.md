kubectl create configmap finetune-script --from-file=train.py

# 1. Apply the ConfigMap
kubectl apply -f configmap.yaml

# 2. Run the Job
kubectl apply -f job.yaml

# 3. Check progress
kubectl get pods -l job-name=cpu-finetune-job
kubectl logs -f $(kubectl get pod -l job-name=cpu-finetune-job -o name)


📦 (Optional) Save trained model to a PersistentVolumeClaim

If you want to keep your trained model between runs, add a PVC:

apiVersion: v1
kind: PersistentVolumeClaim
metadata:
name: finetune-storage
spec:
accessModes: ["ReadWriteOnce"]
resources:
requests:
storage: 10Gi

Then in job.yaml, add:
volumeMounts:
- name: script
mountPath: /workspace
- name: storage
mountPath: /workspace/output
volumes:
- name: script
configMap:
name: finetune-script
- name: storage
persistentVolumeClaim:
claimName: finetune-storage
