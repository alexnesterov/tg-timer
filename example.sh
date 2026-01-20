#!/bin/bash

# Example script to run the Telegram Timer Bot

# Load environment variables from .env file (if it exists)
if [ -f .env ]; then
    export $(cat .env | grep -v '^#' | xargs)
fi

# Check if BOT_TOKEN is set
if [ -z "$BOT_TOKEN" ]; then
    echo "Error: BOT_TOKEN is not set!"
    echo "Please set it in .env file or export as environment variable"
    echo "Copy .env.example to .env and add your token"
    exit 1
fi

echo "Starting Telegram Timer Bot..."
echo "Available commands:"
echo "  /timer 30s - set timer for 30 seconds"
echo "  /timer 10m - set timer for 10 minutes"
echo "  /cancel - cancel active timer"
echo ""

# Build and run the bot
echo "Building bot..."
make build

echo "Running bot..."
./bin/tg-timer
