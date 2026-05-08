#!/usr/bin/env python3
"""\
Go 工具链：**`go.mod` 的 `go major.minor.patch` 为唯一真值**（须三位版本）。`toolchain` 行可选；若存在则须为 `go` + 同版本，`go mod tidy` 常会删掉与 `go` 重复的 `toolchain`。

- `python3 scripts/sync_go_toolchain.py`：校验 Dockerfile、CI、mise、.tool-versions 及文档中的 `golang:` / `toolchain` / 常见 Go 版本表述。
- `python3 scripts/sync_go_toolchain.py --write`：改好 `go` 行后执行，回写上述文件与常见 Markdown（**不必**再手对照清单）。

CI 在 `setup-go` 之前强制执行本脚本。

排除目录见 `SKIP_DIR_PARTS`（含 `node_modules`、`static` 等），避免全仓盲扫拖慢。
"""

from __future__ import annotations

import argparse
import os
import re
import sys
from pathlib import Path


REPO_ROOT = Path(__file__).resolve().parent.parent

SKIP_DIR_PARTS = frozenset(
    {".git", "node_modules", "static", ".cursor", "bin", "uploads", "cert", "vendor"},
)

# 参与扫描的文件：常见文本；含无后缀的 .tool-versions
TEXT_NAMES = frozenset(
    {"Dockerfile", "docker-compose.yml", "Makefile", ".tool-versions"},
)


def _want_path(p: Path) -> bool:
    if not p.is_file():
        return False
    rel = p.relative_to(REPO_ROOT)
    if any(part in SKIP_DIR_PARTS for part in rel.parts):
        return False
    suf = p.suffix.lower()
    if p.name in TEXT_NAMES:
        return True
    return suf in {".md", ".yml", ".yaml", ".toml", ".mod"}


def _walk_repo_files() -> list[Path]:
    """@returns 剪枝遍历仓库内文件路径（跳过 node_modules 等大目录）"""
    out: list[Path] = []
    stack: list[Path] = [REPO_ROOT]
    while stack:
        cur = stack.pop()
        try:
            with os.scandir(cur) as it:
                for entry in it:
                    name = entry.name
                    if entry.is_dir(follow_symlinks=False):
                        if name in SKIP_DIR_PARTS:
                            continue
                        stack.append(Path(entry.path))
                    else:
                        out.append(Path(entry.path))
        except OSError:
            continue
    return out


def _iter_text_files() -> list[Path]:
    out: list[Path] = []
    for p in _walk_repo_files():
        if not _want_path(p):
            continue
        try:
            p.read_text(encoding="utf-8")
        except (OSError, UnicodeDecodeError):
            continue
        out.append(p)
    return sorted(out)


def parse_go_mod(content: str) -> tuple[str, str | None]:
    """
    @returns (go_version, toolchain_token) toolchain_token 形如 go1.25.10；无则为 None。
    """
    go_v: str | None = None
    tc: str | None = None
    for line in content.splitlines():
        s = line.strip()
        if s.startswith("go ") and not s.startswith("golang"):
            parts = s.split()
            if len(parts) == 2 and re.fullmatch(r"\d+\.\d+\.\d+", parts[1]):
                go_v = parts[1]
        if s.startswith("toolchain "):
            parts = s.split()
            if len(parts) == 2 and parts[1].startswith("go"):
                tc = parts[1]
    if not go_v:
        raise ValueError("go.mod 缺少 `go major.minor.patch`（须含补丁号）")
    return go_v, tc


def toolchain_for_go(go_version: str) -> str:
    """@returns 例如 go1.25.10"""
    if not re.fullmatch(r"\d+\.\d+\.\d+", go_version):
        raise ValueError(f"go 版本须为 major.minor.patch: {go_version!r}")
    return "go" + go_version


def collect_violations(go_version: str, tc_need: str) -> list[str]:
    errs: list[str] = []

    def bad(path: Path, msg: str) -> None:
        errs.append(f"{path.relative_to(REPO_ROOT)}: {msg}")

    gomod_path = REPO_ROOT / "go.mod"
    raw = gomod_path.read_text(encoding="utf-8")
    try:
        go_parsed, tc_parsed = parse_go_mod(raw)
    except ValueError as e:
        return [str(e)]
    if go_parsed != go_version:
        bad(gomod_path, f"go 行解析为 {go_parsed}，与预期 {go_version} 不一致")
    if tc_parsed is not None and tc_parsed != tc_need:
        bad(
            gomod_path,
            f"若保留 `toolchain` 行则须为 {tc_need}，当前: {tc_parsed!r}（`go mod tidy` 常会删掉与 `go` 重复的 toolchain，可省略）",
        )

    for path in _iter_text_files():
        text = path.read_text(encoding="utf-8")
        rel = str(path.relative_to(REPO_ROOT))

        for m in re.finditer(r"golang:(\d+\.\d+\.\d+)-alpine", text):
            if m.group(1) != go_version:
                bad(path, f"应为 golang:{go_version}-alpine，出现 golang:{m.group(1)}-alpine")

        if ".github/workflows" in rel and path.suffix.lower() in {".yml", ".yaml"}:
            for m in re.finditer(r"go-version:\s*['\"](\d+\.\d+\.\d+)['\"]", text):
                if m.group(1) != go_version:
                    bad(path, f"go-version 应为 {go_version}")

        if path.name == "mise.toml":
            m = re.search(r'go\s*=\s*"(\d+\.\d+\.\d+)"', text)
            if m and m.group(1) != go_version:
                bad(path, f'mise 应为 go = "{go_version}"')

        if path.name == ".tool-versions":
            for line in text.splitlines():
                s = line.strip()
                if s.startswith("golang ") and len(s.split()) >= 2:
                    v = s.split()[1]
                    if v != go_version:
                        bad(path, f"golang 应为 {go_version}")

        for m in re.finditer(r"\btoolchain (go\d+\.\d+\.\d+)\b", text):
            if m.group(1) != tc_need:
                bad(path, f"toolchain 应为 {tc_need}，出现 {m.group(1)}")

    return errs


def _patch_go_mod(go_version: str) -> None:
    p = REPO_ROOT / "go.mod"
    text = p.read_text(encoding="utf-8")
    t = re.sub(
        r"^go \d+\.\d+\.\d+\s*$",
        f"go {go_version}",
        text,
        flags=re.MULTILINE,
    )
    tc = toolchain_for_go(go_version)
    if re.search(r"^toolchain ", t, re.MULTILINE):
        t = re.sub(
            r"^toolchain go[\d.]+\s*$",
            f"toolchain {tc}",
            t,
            flags=re.MULTILINE,
        )
    else:
        t = re.sub(
            r"^(go \d+\.\d+\.\d+)\s*$",
            rf"\1\n\ntoolchain {tc}\n",
            t,
            count=1,
            flags=re.MULTILINE,
        )
    p.write_text(t, encoding="utf-8")


def write_toolchain_files(go_version: str) -> None:
    _patch_go_mod(go_version)
    tc_need = toolchain_for_go(go_version)

    df = REPO_ROOT / "Dockerfile"
    dtxt = df.read_text(encoding="utf-8")
    dtxt = re.sub(
        r"^FROM golang:\d+\.\d+\.\d+-alpine AS backend\s*$",
        f"FROM golang:{go_version}-alpine AS backend",
        dtxt,
        flags=re.MULTILINE,
    )
    df.write_text(dtxt, encoding="utf-8")

    ci = REPO_ROOT / ".github" / "workflows" / "ci.yml"
    ctxt = ci.read_text(encoding="utf-8")
    ctxt = re.sub(
        r"(go-version:\s*)['\"]\d+\.\d+\.\d+['\"]",
        rf"\g<1>'{go_version}'",
        ctxt,
    )
    ci.write_text(ctxt, encoding="utf-8")

    mise = REPO_ROOT / "mise.toml"
    mtxt = mise.read_text(encoding="utf-8")
    mtxt = re.sub(
        r'go\s*=\s*"\d+\.\d+\.\d+"',
        f'go = "{go_version}"',
        mtxt,
    )
    mise.write_text(mtxt, encoding="utf-8")

    (REPO_ROOT / ".tool-versions").write_text(f"golang {go_version}\n", encoding="utf-8")

    for path in _iter_text_files():
        if path in {df, ci, mise, REPO_ROOT / ".tool-versions", REPO_ROOT / "go.mod"}:
            continue
        suf = path.suffix.lower()
        if suf not in {".md", ".yml", ".yaml"} and path.name not in {"docker-compose.yml"}:
            continue
        t = path.read_text(encoding="utf-8")
        orig = t
        t = re.sub(
            r"golang:\d+\.\d+\.\d+-alpine",
            f"golang:{go_version}-alpine",
            t,
        )
        t = re.sub(
            r"\btoolchain go\d+\.\d+\.\d+\b",
            f"toolchain {tc_need}",
            t,
        )
        t = re.sub(
            r"`go \d+\.\d+\.\d+`",
            f"`go {go_version}`",
            t,
        )
        t = re.sub(
            r"^(-\s*)\*\*Go (\d+\.\d+\.\d+)\*\*",
            rf"\g<1>**Go {go_version}**",
            t,
            flags=re.MULTILINE,
        )
        t = re.sub(
            r"(-\s*Go )\*\*(\d+\.\d+\.\d+)\*\*",
            rf"\g<1>**{go_version}**",
            t,
        )
        t = re.sub(
            r"\*\*`setup-go`\s+(\d+\.\d+\.\d+)\*\*",
            f"**`setup-go` {go_version}**",
            t,
        )
        t = re.sub(
            r"→\s*\*\*`(\d+\.\d+\.\d+)`\*\*）",
            f"→ **`{go_version}`**）",
            t,
        )
        t = re.sub(
            r"(后端工具链为 \*\*Go )\d+\.\d+\.\d+(\*\*)",
            rf"\g<1>{go_version}\g<2>",
            t,
        )
        t = re.sub(
            r"（Go (\d+\.\d+\.\d+)，",
            f"（Go {go_version}，",
            t,
        )
        t = re.sub(
            r"(后端:\s*Go )(\d+\.\d+\.\d+)",
            rf"\g<1>{go_version}",
            t,
        )
        if t != orig:
            path.write_text(t, encoding="utf-8")


def main() -> None:
    ap = argparse.ArgumentParser(description=__doc__)
    ap.add_argument("--check", action="store_true", help="仅校验（默认行为）")
    ap.add_argument("--write", action="store_true", help="按 go 行回写并校验")
    args = ap.parse_args()
    if args.write and args.check:
        print("勿同时使用 --write 与 --check", file=sys.stderr)
        sys.exit(2)

    raw = (REPO_ROOT / "go.mod").read_text(encoding="utf-8")
    go_version, _ = parse_go_mod(raw)
    tc_need = toolchain_for_go(go_version)

    if args.write:
        write_toolchain_files(go_version)
        errs = collect_violations(go_version, tc_need)
        if errs:
            for e in errs:
                print(e, file=sys.stderr)
            sys.exit(1)
        print(f"已对齐 go {go_version} / toolchain {tc_need}。")
        return

    errs = collect_violations(go_version, tc_need)
    if errs:
        for e in errs:
            print(e, file=sys.stderr)
        print(
            "\n在 go.mod 改好 `go` 行后执行: python3 scripts/sync_go_toolchain.py --write && go mod tidy",
            file=sys.stderr,
        )
        sys.exit(1)
    print(f"OK: go {go_version} / {tc_need}")


if __name__ == "__main__":
    main()
