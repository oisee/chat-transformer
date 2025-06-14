#!/bin/bash

# Create a test output directory
TEST_OUTPUT="/tmp/wisdom-test-$(date +%s)"
echo "Testing new structure in: $TEST_OUTPUT"

# Run with timeout to test the new structure
/usr/bin/timeout 20s ./chat-transformer -i "/mnt/c/Users/avinogradova/OneDrive/@wisdom/raw/" -o "$TEST_OUTPUT"

echo ""
echo "=== Directory Structure ==="
echo "Claude structure:"
find "$TEST_OUTPUT/claude" -type d | sort | head -20

echo ""
echo "ChatGPT structure:"
find "$TEST_OUTPUT/chatgpt" -type d | sort | head -20

echo ""
echo "=== README Files ==="
find "$TEST_OUTPUT" -name "README.md" | sort

echo ""
echo "=== Media Files ==="
ls -la "$TEST_OUTPUT/claude/media/" 2>/dev/null || echo "Claude media dir not found"
ls -la "$TEST_OUTPUT/chatgpt/media/" 2>/dev/null || echo "ChatGPT media dir not found"

echo ""
echo "=== Sample Chat Files ==="
echo "Claude chats:"
find "$TEST_OUTPUT/claude/chats" -name "*.json" | head -3
echo "ChatGPT chats:"
find "$TEST_OUTPUT/chatgpt/chats" -name "*.json" | head -3

# Clean up
echo ""
echo "Cleaning up test directory..."
rm -rf "$TEST_OUTPUT"