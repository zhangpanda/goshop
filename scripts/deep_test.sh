#!/bin/bash
# 兼容入口（旧文档仍可能写 deep_test）：
#   默认 —— 与 scripts/quick_test.sh 相同（vet + test，无 -race）。
#   GOSHOP_TEST_RACE=1 —— 与 scripts/ci_test.sh 相同（再加 -race）。
#
# 新习惯：日常 bash scripts/quick_test.sh ；发版 bash scripts/ci_test.sh
#
# 集成/沙盒仍需服务：scripts/integration_test.sh、scripts/sandbox_pay_test.sh

set -euo pipefail
HERE="$(cd "$(dirname "$0")" && pwd)"
if [ "${GOSHOP_TEST_RACE:-}" = "1" ]; then
  exec bash "$HERE/ci_test.sh"
else
  exec bash "$HERE/quick_test.sh"
fi
