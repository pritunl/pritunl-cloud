#!/bin/bash
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
INSTALL_DIR="$SCRIPT_DIR/install"

if [ ! -d "$INSTALL_DIR" ]; then
    echo "Error: install directory not found at: $INSTALL_DIR"
    exit 1
fi

if ! command -v wget &> /dev/null; then
    echo "Error: wget is not installed"
    exit 1
fi

processed=0
downloaded=0

for script in "$INSTALL_DIR"/*.sh; do
    if [ ! -f "$script" ]; then
        echo "No .sh files found in $INSTALL_DIR"
        exit 0
    fi

    echo "Processing: $script"
    ((processed++))

    iso_url=$(grep -E "^[[:space:]]*ISO_URL=" "$script" | tail -1 | cut -d'=' -f2- | sed 's/^["'\'']//' | sed 's/["'\'']$//')

    if [ -z "$iso_url" ]; then
        echo "  No ISO_URL found in $script"
        continue
    fi

    iso_url=$(echo "$iso_url" | sed 's/\${\([^}]*\)}/\1/g' | sed 's/\$\([A-Za-z_][A-Za-z0-9_]*\)/\1/g')

    echo "  Downloading: $iso_url"

    if sudo wget -P /var/lib/virt/iso "$iso_url"; then
        echo "  Successfully downloaded: $filename"
        ((downloaded++))
    else
        echo "  Failed to download from: $iso_url"
    fi

    echo ""
done

echo "Summary:"
echo "  Processed $processed .sh files"
echo "  Successfully downloaded $downloaded files"
