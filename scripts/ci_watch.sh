#!/usr/bin/env bash
# scripts/ci_watch.sh — 用 GitHub CLI 监听 Actions 运行，结束码与流水线结论一致（失败为 1）。
#
# 用法:
#   ./scripts/ci_watch.sh                 # 监听当前分支上名为 CI 的 workflow 最近一次运行
#   ./scripts/ci_watch.sh 25545473309     # 监听指定 run database id
#   WORKFLOW=Release ./scripts/ci_watch.sh
#
# 环境变量:
#   WORKFLOW — Workflow 名称（与 YAML 中 `name:` 一致），默认 CI
#   BRANCH   — 未传 run id 时选用哪条分支，默认当前 git 分支（失败则回退 main）
#   GH_REPO  — 可选，传入 gh 的 owner/repo（等同 gh -R）

set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

WORKFLOW="${WORKFLOW:-CI}"

repo_flags=()
if [[ -n "${GH_REPO:-}" ]]; then
	repo_flags=(--repo "$GH_REPO")
fi

need_gh() {
	if ! command -v gh >/dev/null 2>&1; then
		echo "error: 未找到 gh，请先安装 GitHub CLI 并执行 gh auth login。" >&2
		exit 127
	fi
}

current_branch() {
	local b
	b="$(git rev-parse --abbrev-ref HEAD 2>/dev/null || true)"
	if [[ -z "$b" || "$b" == "HEAD" ]]; then
		echo "main"
	else
		echo "$b"
	fi
}

latest_run_id() {
	local branch="${BRANCH:-$(current_branch)}"
	local id
	id="$(gh run list ${repo_flags[@]+"${repo_flags[@]}"} --workflow "$WORKFLOW" --branch "$branch" --limit 1 --json databaseId --jq '.[0].databaseId')"
	if [[ -z "$id" || "$id" == "null" ]]; then
		echo "error: 未找到分支 «${branch}» 上 workflow «${WORKFLOW}» 的运行记录（可先 push 或检查 WORKFLOW/BRANCH/GH_REPO）。" >&2
		return 1
	fi
	echo "$id"
}

usage() {
	cat <<'EOF'
用法:
  scripts/ci_watch.sh [run_database_id]

未传 run id 时，会取当前仓库、当前分支（或 BRANCH）、WORKFLOW（默认 CI）的最近一次运行并 watch。
失败时进程退出码为 1；未安装 gh 为 127。
EOF
}

main() {
	need_gh

	if [[ "${1:-}" == "-h" || "${1:-}" == "--help" ]]; then
		usage
		exit 0
	fi

	local run_id="${1:-}"
	if [[ -z "$run_id" ]]; then
		run_id="$(latest_run_id)"
	fi

	echo "Watching GitHub Actions run ${run_id} (workflow=${WORKFLOW}) …"
	if ! gh run watch "${repo_flags[@]}" "$run_id" --exit-status; then
		echo "error: run ${run_id} 未通过或 gh run watch 失败。" >&2
		echo "  建议: gh run view ${run_id} --log-failed${GH_REPO:+ -R ${GH_REPO}}" >&2
		exit 1
	fi
}

main "$@"
