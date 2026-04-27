#!/bin/bash
cd "$(dirname "$0")"
npx playwright test e2e/ --reporter=line > /tmp/e2e-deep.txt 2>&1
tail -5 /tmp/e2e-deep.txt
