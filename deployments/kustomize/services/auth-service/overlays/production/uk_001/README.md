We can now apply this entire production configuration for the auth-service with a single command:

kubectl apply -k deployments/kustomize/services/auth-service/overlays/production/uk_001