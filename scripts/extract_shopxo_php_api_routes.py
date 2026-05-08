#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""扫描 ShopXO **PHP** `app/api/controller/*.php` 中对外可调 action，映射为 `s=controller/action`（全小写）。

ThinkPHP / ShopXO 惯例：`User::LoginVerifySend` → `user/loginverifysend`（大写字母分词后全小写拼接）。

用法:
  python3 scripts/extract_shopxo_php_api_routes.py /path/to/shopxo
  python3 scripts/extract_shopxo_php_api_routes.py /path/to/shopxo --fail-uniapp-contract

  --fail-uniapp-contract  读取仓库内 `scripts/data/shopxo_uniapp_normal_routes.txt`，要求每条路由在
    (1) 本脚本扫出的 PHP API、(2) GoShop `compat.go` routeMap 中均存在；缺一即退出码 1。

exit 0 成功；1 合约失败；2 用法/路径错误。
"""

from __future__ import annotations

import json
import re
import sys
from pathlib import Path

_RE_PUBLIC_METHOD = re.compile(r"^\s*public\s+function\s+(\w+)\s*\(", re.MULTILINE)

# 不参与 API 契约的 Controller（基类、运维、纯内部等）
_SKIP_CONTROLLER_FILES = frozenset({"Common.php"})

_SKIP_METHODS = frozenset(
    {
        "__construct",
        "__destruct",
        "__call",
        "FormTableInit",
    }
)


def php_method_to_action(name: str) -> str:
    """PascalCase 方法名 → shopxo-uniapp 风格 action 字符串（全小写、无分隔符）。"""
    parts = re.findall(r"[A-Z][a-z0-9]*", name)
    if not parts:
        return name.lower()
    return "".join(p.lower() for p in parts)


def extract_php_routes(controller_dir: Path) -> tuple[set[str], list[str]]:
    routes_set: set[str] = set()
    warnings: list[str] = []
    for php in sorted(controller_dir.glob("*.php")):
        if php.name in _SKIP_CONTROLLER_FILES:
            continue
        ctrl = php.stem.lower()
        try:
            text = php.read_text(encoding="utf-8", errors="replace")
        except OSError as e:
            warnings.append(f"read {php}: {e}")
            continue
        for m in _RE_PUBLIC_METHOD.finditer(text):
            meth = m.group(1)
            if meth in _SKIP_METHODS:
                continue
            action = php_method_to_action(meth)
            key = f"{ctrl}/{action}"
            routes_set.add(key)
    return routes_set, warnings


def goshop_route_keys(compat_go: Path) -> set[str]:
    keys: set[str] = set()
    text = compat_go.read_text(encoding="utf-8", errors="ignore")
    for m in re.finditer(
        r'^\s+"([a-z]+/[a-z_]+)":\s*(?:sx[a-zA-Z0-9]+|handler\.)', text, re.MULTILINE
    ):
        keys.add(m.group(1))
    return keys


def load_uniapp_contract(repo: Path) -> list[str]:
    p = repo / "scripts" / "data" / "shopxo_uniapp_normal_routes.txt"
    if not p.is_file():
        raise SystemExit(f"缺少合约文件: {p}")
    lines = [ln.strip() for ln in p.read_text(encoding="utf-8").splitlines()]
    return sorted({ln for ln in lines if ln and not ln.startswith("#")})


def main() -> int:
    argv = [a for a in sys.argv[1:] if not str(a).startswith("-")]
    flags = {str(a) for a in sys.argv[1:] if str(a).startswith("-")}
    fail_contract = "--fail-uniapp-contract" in flags

    if len(argv) < 1:
        print("用法见脚本顶部 docstring。", file=sys.stderr)
        return 2
    root = Path(argv[0]).resolve()
    ctrl = root / "app" / "api" / "controller"
    if not ctrl.is_dir():
        print(f"不是 ShopXO 根目录或缺少 {ctrl}", file=sys.stderr)
        return 2

    php_routes, warns = extract_php_routes(ctrl)
    repo = Path(__file__).resolve().parent.parent
    compat = repo / "internal" / "compat" / "shopxo" / "compat.go"
    goshop = goshop_route_keys(compat) if compat.is_file() else set()

    report: dict = {
        "shopxo_root": str(root),
        "php_api_route_count": len(php_routes),
        "php_routes": sorted(php_routes),
        "goshop_route_map_count": len(goshop),
        "in_php_not_in_goshop": sorted(php_routes - goshop),
        "in_goshop_not_in_php_scan": sorted(goshop - php_routes),
        "parse_warnings": warns[:30],
    }

    if fail_contract:
        uni = load_uniapp_contract(repo)
        missing_php = sorted(set(uni) - php_routes)
        missing_go = sorted(set(uni) - goshop)
        report["uniapp_contract_count"] = len(uni)
        report["contract_missing_in_php"] = missing_php
        report["contract_missing_in_goshop"] = missing_go

    print(json.dumps(report, ensure_ascii=False, indent=2))

    if fail_contract:
        bad = False
        if missing_php:
            bad = True
            print(
                "\n--fail-uniapp-contract: 下列 uni-app 冻结路由在 **PHP app/api/controller 扫描结果中不存在**（请核对方法命名映射）:\n  "
                + "\n  ".join(missing_php),
                file=sys.stderr,
            )
        if missing_go:
            bad = True
            print(
                "\n--fail-uniapp-contract: 下列路由在 **Go routeMap 中不存在**:\n  "
                + "\n  ".join(missing_go),
                file=sys.stderr,
            )
        if bad:
            return 1
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
