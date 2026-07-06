#!/bin/bash

output_file="resulttextoutput.txt"

# Clear or create the output file
> "$output_file"

find . -type f \
    -not \( -name "go.mod" -o -name "go.sum" -o -name "*.hcl" -o -name "*.tar" -o -name "*.log" -o -name "*.tfstate" -o -name "$output_file" \) \
    -not \( -name "*.tfstate.backup" -o -name "terraform.tfstate.*prod-cluster" \) \
		-not \( -name "persona-cli" -o -name "README.md" -o -name "create_persona.sql" -o -name "*.secret" -o -name "*-lock.json" -o -name "*.txt" \) \
    -not -path "*/\\.*/*" \
		-not -path "*/backup/*" \
		-not -path "*/images/*" \
		-not -path "*/project3/*" \
		-not -path "*/docs/*" \
		-not -path "./projects/*" \
		-not -path "*/gateway/templates/*" \
		-not -path "*/strimzi-0.47.0/*" \
		-not -path "*/production/sydney/*" \
    -print0 | \
while IFS= read -r -d $'\0' file; do
    echo "filepath = $file" >> "$output_file" || { echo "Failed to write to $output_file"; exit 1; }
    cat "$file" >> "$output_file" || { echo "Failed to write to $output_file"; exit 1; }
    echo "-------------------------------------------------" >> "$output_file" || { echo "Failed to write to $output_file"; exit 1; }
done
