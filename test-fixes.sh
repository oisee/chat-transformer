#!/bin/bash

# Backup existing output
if [ -d "/mnt/c/Users/avinogradova/OneDrive/@wisdom/expanded_backup" ]; then
    rm -rf "/mnt/c/Users/avinogradova/OneDrive/@wisdom/expanded_backup"
fi
mv "/mnt/c/Users/avinogradova/OneDrive/@wisdom/expanded" "/mnt/c/Users/avinogradova/OneDrive/@wisdom/expanded_backup"

# Run with timeout to test fixes
/usr/bin/timeout 30s ./chat-transformer -i "/mnt/c/Users/avinogradova/OneDrive/@wisdom/raw/" -o "/mnt/c/Users/avinogradova/OneDrive/@wisdom/expanded/"

echo "Testing completed. Check the fixed conversation:"
echo "Looking for the specific conversation..."
find "/mnt/c/Users/avinogradova/OneDrive/@wisdom/expanded/chatgpt" -name "*AI Ethics*" -type f