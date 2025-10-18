#!/bin/bash
# Test script for echo.aml
# This script demonstrates how to test the echo agent interactively

cd "$(dirname "$0")"

echo "Testing echo.aml agent..."
echo "Building..."
amlx build echo.aml -o echo_test

echo ""
echo "You can now run the echo agent with:"
echo "  ./echo_test echo.aml"
echo ""
echo "Or use amlx run:"
echo "  amlx run echo.aml"
echo ""
echo "The agent will show '>' and echo back what you type."
echo "Type 'exit' to quit."
echo ""

# Run it directly
./echo_test echo.aml
