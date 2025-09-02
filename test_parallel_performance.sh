#!/bin/bash

echo "Testing Phase 1 Optimizations"
echo "=============================="

BINARY="./bin/murex"

echo "Testing --parallel 0 (unlimited) with 10 items:"
time $BINARY -c 'a [1..10] -> foreach ! --parallel 0 { out "Item processed" }'

echo ""
echo "Testing --parallel 2 with 10 items:"
time $BINARY -c 'a [1..10] -> foreach ! --parallel 2 { out "Item processed" }'

echo ""
echo "Testing --parallel 4 with 10 items:"
time $BINARY -c 'a [1..10] -> foreach ! --parallel 4 { out "Item processed" }'

echo ""
echo "Testing serial processing for comparison:"
time $BINARY -c 'a [1..10] -> foreach ! { out "Item processed" }'