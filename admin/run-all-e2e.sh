#!/bin/bash
cd "$(dirname "$0")/.."
kill $(lsof -ti :8080) 2>/dev/null || true
sleep 1
GOSHOP_E2E=1 nohup ./bin/goshop > /tmp/goshop-e2e.log 2>&1 &
sleep 3
cd admin
npx playwright test e2e/ --reporter=line > /tmp/e2e-all.txt 2>&1
tail -5 /tmp/e2e-all.txt
