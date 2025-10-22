#!/bin/bash

echo "Starting Classius server..."

# Set environment variables if needed
export CLASSIUS_SKIP_MIGRATIONS=true

# Try to run the server, but continue even if migrations fail
./server 2>&1 | while IFS= read -r line; do
    echo "$line"
    if [[ "$line" == *"Failed to run migrations"* ]]; then
        echo "Migration error detected, but continuing anyway..."
        # Don't exit, let the server continue
    fi
done