#!/bin/bash

set -e

# Get the directory of this script
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# Change to project root
cd "$PROJECT_ROOT"

echo "Updating README.md with current scenarios..."

# Check if README.md exists
if [[ ! -f "README.md" ]]; then
    echo "Error: README.md not found!"
    exit 1
fi

# Check if markers exist
if ! grep -q "<!-- AUTO-GENERATED-SCENARIOS-START -->" README.md; then
    echo "Error: AUTO-GENERATED-SCENARIOS-START marker not found in README.md!"
    echo "Please add the following markers around the scenarios list:"
    echo "<!-- AUTO-GENERATED-SCENARIOS-START -->"
    echo "<!-- AUTO-GENERATED-SCENARIOS-END -->"
    exit 1
fi

if ! grep -q "<!-- AUTO-GENERATED-SCENARIOS-END -->" README.md; then
    echo "Error: AUTO-GENERATED-SCENARIOS-END marker not found in README.md!"
    echo "Please add the following markers around the scenarios list:"
    echo "<!-- AUTO-GENERATED-SCENARIOS-START -->"
    echo "<!-- AUTO-GENERATED-SCENARIOS-END -->"
    exit 1
fi

# Get current scenarios
SCENARIOS=$(go run main.go list | sort)

# Create the new scenarios section
SCENARIOS_SECTION=""
while IFS= read -r scenario; do
    if [[ -n "$scenario" ]]; then
        SCENARIOS_SECTION="${SCENARIOS_SECTION}- \`$scenario\`"$'\n'
    fi
done <<< "$SCENARIOS"

# Remove trailing newline
SCENARIOS_SECTION=$(echo -n "$SCENARIOS_SECTION")

# Create a temporary file to work with
cp README.md README.tmp

# Use awk to replace content between markers
awk -v scenarios="$SCENARIOS_SECTION" '
BEGIN { in_section = 0 }
/<!-- AUTO-GENERATED-SCENARIOS-START -->/ { 
    print $0
    print scenarios
    in_section = 1
    next
}
/<!-- AUTO-GENERATED-SCENARIOS-END -->/ { 
    in_section = 0
    print $0
    next
}
!in_section { print $0 }
' README.tmp > README.new

# Replace the original file
mv README.new README.md
rm README.tmp

echo "README.md updated successfully!"
echo "Current scenarios:"
echo "$SCENARIOS"
