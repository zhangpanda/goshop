#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""对 **同一请求** 分别访问 ShopXO PHP 与 GoShop，比较 JSON 响应中 `data` 的 **键路径形状**（不比较数值）。

用于「逐字段对拍」的一阶段：**结构/字段名**是否对齐；数值与业务差异需结合人工或更严的 golden。

前提：两个站点均可访问，且对匿名接口返回 `{"code":0,"msg":...,"data":...}`（与 ShopXO ApiService 习惯一致）。

用法:
  SHOPXO_PHP_BASE=https://your-php.shop/ SHOPXO_GO_BASE=http://127.0.0.1:8080 \\
    python3 scripts/shopxo_dual_json_diff.py

  --samples path   默认 `scripts/data/shopxo_json_diff_samples.json`
  --fail-on-shape  若 PHP 的 `data` 中出现 Go 侧不存在的键路径，则退出码 1

环境变量:
  SHOPXO_PHP_BASE   PHP 站点根（尾可无 `/`）
  SHOPXO_GO_BASE      Go 站点根

exit 0/1/2
"""

from __future__ import annotations

import argparse
import json
import os
import sys
import urllib.error
import urllib.parse
import urllib.request
from pathlib import Path
from typing import Any


def _paths(obj: Any, prefix: str = "") -> set[str]:
    """对象内所有 dict 键路径（用 '.' 连接）；list 用 `[*]` 表示一层。"""
    out: set[str] = set()
    if isinstance(obj, dict):
        for k, v in obj.items():
            p = f"{prefix}.{k}" if prefix else k
            out.add(p)
            out |= _paths(v, p)
    elif isinstance(obj, list) and obj:
        p = f"{prefix}[*]" if prefix else "[*]"
        out.add(p)
        out |= _paths(obj[0], p)
    return out


def _fetch_json(base: str, path_with_q: str) -> tuple[int, Any | None]:
    url = base.rstrip("/") + path_with_q
    req = urllib.request.Request(url, headers={"User-Agent": "shopxo-dual-json-diff/1"})
    try:
        with urllib.request.urlopen(req, timeout=60) as r:
            raw = r.read().decode("utf-8", errors="replace")
            return r.status, json.loads(raw)
    except urllib.error.HTTPError as e:
        raw = e.read().decode("utf-8", errors="replace")
        try:
            return e.code, json.loads(raw)
        except json.JSONDecodeError:
            return e.code, None
    except Exception as e:
        print(f"请求失败 {url}: {e}", file=sys.stderr)
        return -1, None


def main() -> int:
    ap = argparse.ArgumentParser()
    ap.add_argument(
        "--samples",
        default=str(Path(__file__).resolve().parent.parent / "scripts/data/shopxo_json_diff_samples.json"),
    )
    ap.add_argument("--fail-on-shape", action="store_true")
    args = ap.parse_args()

    php_base = os.environ.get("SHOPXO_PHP_BASE", "").rstrip("/")
    go_base = os.environ.get("SHOPXO_GO_BASE", "").rstrip("/")
    if not php_base or not go_base:
        print("请设置 SHOPXO_PHP_BASE 与 SHOPXO_GO_BASE", file=sys.stderr)
        return 2

    samples_path = Path(args.samples)
    samples = json.loads(samples_path.read_text(encoding="utf-8"))

    failed = False
    for spec in samples:
        route = spec["route"]
        method = spec.get("method", "GET").upper()
        q = spec.get("query", "").strip()
        if q and not q.startswith("&"):
            q = "&" + q
        path = "/api.php?s=" + urllib.parse.quote(route) + q
        if method != "GET":
            print(f"跳过非 GET 样本: {route}", file=sys.stderr)
            continue

        _, j_php = _fetch_json(php_base, path)
        _, j_go = _fetch_json(go_base, path)
        if not isinstance(j_php, dict) or not isinstance(j_go, dict):
            print(f"{route}: 非 JSON 或解析失败", file=sys.stderr)
            failed = True
            continue

        d_p = j_php.get("data")
        d_g = j_go.get("data")
        paths_p = _paths(d_p, "data") if d_p is not None else set()
        paths_g = _paths(d_g, "data") if d_g is not None else set()
        only_p = sorted(paths_p - paths_g)
        only_g = sorted(paths_g - paths_p)
        print(f"\n=== {route} ===")
        print(f"  PHP code={j_php.get('code')} Go code={j_go.get('code')}")
        if only_p:
            print("  仅 PHP data 形状有:", only_p)
            if args.fail_on_shape:
                failed = True
        if only_g:
            print("  仅 Go data 形状有:", only_g)

    return 1 if failed else 0


if __name__ == "__main__":
    raise SystemExit(main())
