#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""从官方 shopxo-uniapp 源码抽取 get_request_url 对应的 s=controller/action。

App.vue 约定：get_request_url(a, c, plugins, ...) 最终为 api.php?s=c/a；
若带 plugins，则为 s=plugins/index&pluginsname=...&pluginscontrol=c&pluginsaction=a（与 routeMap 单列不同）。

用法:
  python3 scripts/extract_shopxo_uniapp_routes.py /path/to/shopxo-uniapp
  python3 scripts/extract_shopxo_uniapp_routes.py /path/to/shopxo-uniapp --fail-on-missing

  --fail-on-missing  若 uni-app 静态扫描得到的普通 s= 路由在 GoShop routeMap 中有缺失，则退出码 1（供 CI 锁死「可替换」基线）。

输出: stdout 为 JSON，含 normal_routes、plugin_calls、goshop_only 等。
"""

from __future__ import annotations

import json
import re
import sys
from pathlib import Path

# get_request_url( action, control, plugin?, ... ) — 参数多为单引号或双引号字符串
_RE_CALL = re.compile(
    r"""get_request_url\s*\(\s*"""
    r"""(['"])(?P<a>.*?)\1\s*,\s*"""
    r"""(['"])(?P<c>.*?)\3"""
    r"""(?:\s*,\s*(['"])(?P<p>.*?)\5)?""",
    re.DOTALL,
)

_SKIP_DIRS = {"node_modules", "unpackage", "dist", ".git"}


def iter_source_files(root: Path):
    for pat in ("*.vue", "*.js", "*.nvue", "*.ts"):
        for p in root.rglob(pat):
            if any(part in _SKIP_DIRS for part in p.parts):
                continue
            yield p


def extract_calls(root: Path) -> tuple[set[tuple[str, str, str | None]], list[str]]:
    """返回 {(action, control, plugin_or_none)}, 以及解析警告。"""
    out: set[tuple[str, str, str | None]] = set()
    warns: list[str] = []
    for path in iter_source_files(root):
        try:
            text = path.read_text(encoding="utf-8", errors="ignore")
        except OSError as e:
            warns.append(f"read {path}: {e}")
            continue
        for m in _RE_CALL.finditer(text):
            a = (m.group("a") or "").strip()
            c = (m.group("c") or "").strip()
            plug = m.group("p")
            if plug is not None:
                plug = plug.strip()
                if plug in ("params", "group") or "=" in plug:
                    # 误判：第四参是 URL 拼接而非插件名
                    continue
            if not a or not c or "\n" in a or "\n" in c:
                continue
            if plug:
                out.add((a, c, plug))
            else:
                out.add((a, c, None))
    return out, warns


def goshop_route_keys(compat_go: Path) -> set[str]:
    keys: set[str] = set()
    text = compat_go.read_text(encoding="utf-8", errors="ignore")
    for m in re.finditer(r'^\s+"([a-z]+/[a-z_]+)":\s*(?:sx[a-zA-Z0-9]+|handler\.)', text, re.MULTILINE):
        keys.add(m.group(1))
    return keys


def main() -> int:
    argv = [a for a in sys.argv[1:] if not a.startswith("-")]
    flags = {a for a in sys.argv[1:] if a.startswith("-")}
    fail_on_missing = "--fail-on-missing" in flags

    if len(argv) < 1:
        print(
            "用法: python3 extract_shopxo_uniapp_routes.py /path/to/shopxo-uniapp [--fail-on-missing]",
            file=sys.stderr,
        )
        return 2
    root = Path(argv[0]).resolve()
    if not root.is_dir():
        print(f"不是目录: {root}", file=sys.stderr)
        return 2

    calls, warns = extract_calls(root)
    normal_s = sorted({f"{c}/{a}" for a, c, p in calls if p is None})
    plugin_s = sorted(
        {
            f"plugins/index|name={p}|ctl={c}|act={a}"
            for a, c, p in calls
            if p is not None
        }
    )

    repo = Path(__file__).resolve().parent.parent
    compat = repo / "internal" / "compat" / "shopxo" / "compat.go"
    goshop = goshop_route_keys(compat) if compat.is_file() else set()

    uni_normal = set(normal_s)
    missing = sorted(uni_normal - goshop)
    extra = sorted(goshop - uni_normal)

    report = {
        "uniapp_root": str(root),
        "normal_route_count": len(normal_s),
        "normal_routes": normal_s,
        "plugin_style_call_count": len(plugin_s),
        "plugin_calls_sample": plugin_s[:80],
        "goshop_route_map_count": len(goshop),
        "missing_in_goshop": missing,
        "in_goshop_not_seen_in_uniapp_scan": extra,
        "parse_warnings": warns[:20],
    }
    print(json.dumps(report, ensure_ascii=False, indent=2))
    if warns:
        print(f"\n# {len(warns)} parse/file warnings (showing first 20 in JSON)", file=sys.stderr)

    if fail_on_missing and missing:
        print(
            "\n--fail-on-missing: routeMap 缺少下列 uni-app 普通 s= 路由:\n  "
            + "\n  ".join(missing),
            file=sys.stderr,
        )
        return 1
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
