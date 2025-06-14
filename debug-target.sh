#!/bin/bash

# Clear the target conversation to force reprocessing
rm -f "/mnt/c/Users/avinogradova/OneDrive/@wisdom/expanded/chatgpt/conversations/2025/06/2025-06-11_AI Ethics and Copyright.json"

# Run just to process ChatGPT and capture the debug output for our target conversation
./chat-transformer -i "/mnt/c/Users/avinogradova/OneDrive/@wisdom/raw/" -o "/mnt/c/Users/avinogradova/OneDrive/@wisdom/expanded/" 2>&1 | grep -A 20 -B 5 "DEBUG.*68490016-358c-800c-a8e7-a0965ab83993"