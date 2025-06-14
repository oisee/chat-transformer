#!/bin/bash

echo "=== Testing Media Functionality ==="

# Test default behavior (no copying)
echo "Testing default behavior (relative paths only)..."
TEST_OUTPUT="/tmp/wisdom-media-test-$(date +%s)"
/usr/bin/timeout 15s ./chat-transformer -i "/mnt/c/Users/avinogradova/OneDrive/@wisdom/raw/" -o "$TEST_OUTPUT"

echo ""
echo "=== Media Info with Relative Paths ==="
head -20 "$TEST_OUTPUT/chatgpt/media/media_info.json"

echo ""
echo "=== Media Directory Contents (default) ==="
ls -la "$TEST_OUTPUT/chatgpt/media/"

echo ""
echo "=== Testing --copy-media Flag ==="
TEST_OUTPUT_COPY="/tmp/wisdom-media-copy-test-$(date +%s)"
/usr/bin/timeout 10s ./chat-transformer -i "/mnt/c/Users/avinogradova/OneDrive/@wisdom/raw/" -o "$TEST_OUTPUT_COPY" --copy-media

echo ""
echo "=== Media Directory Contents (with copy) ==="
ls -la "$TEST_OUTPUT_COPY/chatgpt/media/" 2>/dev/null || echo "Copy test timed out"

echo ""
echo "=== README Files Created ==="
find "$TEST_OUTPUT/chatgpt/media" -name "README.md" 2>/dev/null || echo "No READMEs found"

# Clean up
rm -rf "$TEST_OUTPUT" "$TEST_OUTPUT_COPY"