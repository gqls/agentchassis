find . -type f -exec sh -c '
  echo "===== $1 ====="
  head -c 30000 "$1"
  echo
  echo
' sh {} \; > repo_summary.txt
