make build - Build all test components
make test-all - Run all tests locally
make test-all-k8s - Run all tests in Kubernetes
make help - See all available commands

If you want to debug issues, use these GODEBUG options instead:
# For debugging specific issues (add only when needed):
# ENV GODEBUG=http2debug=2  # Debug HTTP/2 issues
# ENV GODEBUG=netdns=go     # Use Go's DNS resolver
# ENV GODEBUG=schedtrace=1000  # Print scheduler trace every second