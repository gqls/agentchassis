#!/usr/bin/env bash
# Creates 085_rerender-provocations-vonc.sh by copying the proven index rerender
# script and swapping page identifiers (reuse-first; eyeball the diff after).
set -euo pipefail
SRC=$(ls scripts/initial_messages/210_vonc_trigger/*rerender-index-vonc.sh | head -1)
DST=scripts/initial_messages/210_vonc_trigger/085_rerender-provocations-vonc.sh
sed -e 's/b4d24f8e-fccd-49df-9dad-aa56a0b20a68/e4b3b195-919f-45ad-854e-201d3e846ea8/g' \
    -e 's#"filename":"index.html"#"filename":"provocations/index.html"#g' \
    -e 's#"page_name":"index"#"page_name":"provocations-index"#g' \
    -e 's#Rerender: index.html#Rerender: provocations/index.html#g' \
    "$SRC" > "$DST"
chmod +x "$DST"
echo "created $DST from $SRC — diff to eyeball:"
diff "$SRC" "$DST" || true
