#!/bin/bash
set -e
cd "$(dirname "$0")/.."

kill $(lsof -ti :8080) 2>/dev/null || true
kill $(lsof -ti :3010) 2>/dev/null || true
sleep 1

GOSHOP_E2E=1 nohup ./bin/goshop > /tmp/goshop-e2e.log 2>&1 &
echo "Backend PID=$!"

cd admin
nohup npm run dev > /tmp/admin-dev.log 2>&1 &
echo "Frontend PID=$!"

for i in $(seq 1 15); do
  sleep 2
  RES=$(curl -s -X POST http://127.0.0.1:8080/api/admin/login \
    -H 'Content-Type: application/json' \
    -d '{"username":"admin","password":"admin123","captcha_key":"t","captcha_code":"000000"}' 2>/dev/null)
  if echo "$RES" | grep -q '"code":0'; then echo "Backend OK"; break; fi
  [ "$i" = "15" ] && echo "Backend FAIL: $RES"
done

for i in $(seq 1 15); do
  sleep 2
  CODE=$(curl -s -o /dev/null -w "%{http_code}" http://127.0.0.1:3010/login 2>/dev/null)
  [ "$CODE" = "200" ] && echo "Frontend OK" && break
  [ "$i" = "15" ] && echo "Frontend FAIL"
done

npx playwright test e2e/full-flow.spec.ts --reporter=line > /tmp/e2e-result.txt 2>&1
echo "Exit: $?"
