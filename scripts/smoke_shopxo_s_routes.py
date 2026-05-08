#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""对 running GoShop 做 ShopXO 风格 `/api.php` 的 **全量 s= 冒烟**（与 `scripts/data/shopxo_uniapp_normal_routes.txt` 一致）。

前提：BASE 指向已启动的站点（与 `integration_test.sh` 相同）。

校验规则（替换 PHP 部署的最低门禁：接口已挂接，非 5xx，且 JSON 响应不出现「接口不存在」）：
  - JSON：`code` 存在则为业务/参数错误亦可（-1 等），但不得为路由缺失；
  - 验证码入口：HTTP 200 且 `Content-Type` 含 `image/png`。

用法:
  BASE=http://127.0.0.1:8080 python3 scripts/smoke_shopxo_s_routes.py

环境变量:
  TOKEN   若已登录，可传入 JWT（query `token=`），跳过本脚本内注册；
  GOSHOP_SMOKE_SKIP_SETUP=1  假定 TOKEN 已具备，且仍尝试 REST 建单（失败则部分带单号的路由随缘）。

exit 0 全通过；1 有失败。
"""

from __future__ import annotations

import json
import os
import random
import re
import string
import sys
import urllib.error
import urllib.parse
import urllib.request
from pathlib import Path


REPO_ROOT = Path(__file__).resolve().parent.parent
ROUTES_FILE = REPO_ROOT / "scripts" / "data" / "shopxo_uniapp_normal_routes.txt"
COMPAT_FILE = REPO_ROOT / "internal" / "compat" / "shopxo" / "compat.go"

UNKNOWN_ROUTE_MARK = "接口不存在"

# 仅验证码：返回 PNG，不校验 JSON
PNG_ROUTES = frozenset({"user/userverifyentry", "forminput/verifyentry"})


def _ascii_safe_request_path(path_qs: str) -> str:
    """
    将 path 中 `?` 后的 query 重新 urlencode，避免 http.client 对 request-target 使用 ASCII 编码时，
    因 wd=手机 等非 ASCII 查询值触发 UnicodeEncodeError。
    """
    if "?" not in path_qs:
        return path_qs
    path_only, q = path_qs.split("?", 1)
    pairs = urllib.parse.parse_qsl(q, keep_blank_values=True)
    encoded = urllib.parse.urlencode(pairs, quote_via=urllib.parse.quote)
    return f"{path_only}?{encoded}"


def _load_auth_required() -> set[str]:
    text = COMPAT_FILE.read_text(encoding="utf-8", errors="replace")
    i = text.index("var authRequiredRoutes")
    j = text.index("var routeMap", i)
    block = text[i:j]
    return set(re.findall(r'"([a-z]+/[a-z_]+)":\s*true', block))


def _load_routes() -> list[str]:
    if not ROUTES_FILE.is_file():
        print(f"缺少路由清单: {ROUTES_FILE}（请从官方 uni-app 生成并提交）", file=sys.stderr)
        raise SystemExit(2)
    lines = [ln.strip() for ln in ROUTES_FILE.read_text(encoding="utf-8").splitlines()]
    return sorted({ln for ln in lines if ln and not ln.startswith("#")})


def _req(
    base: str,
    method: str,
    path_qs: str,
    *,
    data: bytes | None = None,
    headers: dict[str, str] | None = None,
) -> tuple[int, str, dict[str, str]]:
    url = base.rstrip("/") + _ascii_safe_request_path(path_qs)
    h = {"User-Agent": "goshop-smoke-shopxo/1.0"}
    if headers:
        h.update(headers)
    r = urllib.request.Request(url, data=data, headers=h, method=method)
    try:
        with urllib.request.urlopen(r, timeout=60) as resp:
            body = resp.read()
            ct = resp.headers.get("Content-Type", "")
            return resp.status, body.decode("utf-8", errors="replace"), {"Content-Type": ct}
    except urllib.error.HTTPError as e:
        body = e.read().decode("utf-8", errors="replace")
        return e.code, body, {"Content-Type": e.headers.get("Content-Type", "")}


def _rest_json(base: str, method: str, path: str, body: dict | None, bearer: str) -> dict:
    data = None
    hdr = {"Authorization": f"Bearer {bearer}", "Content-Type": "application/json"}
    if body is not None:
        data = json.dumps(body).encode("utf-8")
    code, raw, _ = _req(base, method, path, data=data, headers=hdr)
    try:
        out = json.loads(raw)
    except json.JSONDecodeError as ex:
        raise RuntimeError(f"REST {method} {path} HTTP {code} 非 JSON: {raw[:200]}") from ex
    if out.get("code") != 0:
        raise RuntimeError(f"REST {method} {path} fail: {out}")
    return out["data"]


def _order_id(odata) -> int:
    if isinstance(odata, dict):
        return int(odata["id"])
    return int(odata[0]["id"])


def _cart_row_id(cdata) -> int:
    if isinstance(cdata, dict):
        return int(cdata["id"])
    return int(cdata[0]["id"])


def _setup_rest(base: str) -> tuple[str, str, str, str, str, str, str]:
    """注册、登录、地址、两笔待付款订单、两条未下单购物车行（stock / delete 分离）；返回 token, addr_id, cart_stock_id, cart_del_id, order_a, order_b, tag。"""
    suf = "".join(random.choices(string.ascii_lowercase + string.digits, k=8))
    user = f"smoke{suf}"
    pwd = "smoke123456"
    _rest_json(
        base,
        "POST",
        "/api/register",
        {"username": user, "password": pwd},
        "",
    )
    login = _rest_json(base, "POST", "/api/login", {"username": user, "password": pwd}, "")
    token = login["token"]
    addr = _rest_json(
        base,
        "POST",
        "/api/address",
        {
            "name": "冒烟",
            "phone": "13800000000",
            "province": "北京市",
            "city": "北京市",
            "district": "朝阳区",
            "detail": "测试地址",
            "is_default": True,
        },
        token,
    )
    addr_id = str(addr["id"])
    order_ids: list[str] = []
    for _ in range(2):
        cdata = _rest_json(
            base,
            "POST",
            "/api/cart",
            {"goods_id": 1, "sku_id": 1, "quantity": 1},
            token,
        )
        cid = _cart_row_id(cdata)
        odata = _rest_json(
            base,
            "POST",
            "/api/orders",
            {"address_id": int(addr_id), "cart_ids": [cid]},
            token,
        )
        oid = _order_id(odata)
        order_ids.append(str(oid))
    fresh_stock = _rest_json(
        base,
        "POST",
        "/api/cart",
        {"goods_id": 1, "sku_id": 1, "quantity": 1},
        token,
    )
    id_stock = str(_cart_row_id(fresh_stock))
    fresh_del = _rest_json(
        base,
        "POST",
        "/api/cart",
        {"goods_id": 1, "sku_id": 1, "quantity": 1},
        token,
    )
    id_del = str(_cart_row_id(fresh_del))
    return token, addr_id, id_stock, id_del, order_ids[0], order_ids[1], suf


def _payments_offline_id(base: str) -> str:
    code, raw, _ = _req(base, "GET", "/api/payments", data=None)
    if code != 200:
        raise RuntimeError(f"GET /api/payments HTTP {code}")
    d = json.loads(raw)
    if d.get("code") != 0:
        raise RuntimeError(d)
    for p in d.get("data") or []:
        cfg = p.get("config") or ""
        if "offline" in cfg or '"payment_key":"offline"' in cfg:
            return str(p["id"])
    raise RuntimeError("无 offline 支付方式")


def _shopxo(
    base: str,
    token: str,
    route: str,
    *,
    method: str = "GET",
    extra_qs: str = "",
    form_body: str | None = None,
    json_body: dict | None = None,
) -> tuple[int, str, dict[str, str]]:
    path = "/api.php?s=" + urllib.parse.quote(route)
    if token:
        path += "&token=" + urllib.parse.quote(token)
    if extra_qs:
        path += extra_qs if extra_qs.startswith("&") else "&" + extra_qs
    data = None
    hdr: dict[str, str] = {}
    if json_body is not None:
        data = json.dumps(json_body).encode("utf-8")
        hdr["Content-Type"] = "application/json"
    elif form_body is not None:
        data = form_body.encode("utf-8")
        hdr["Content-Type"] = "application/x-www-form-urlencoded"
    status, body, heads = _req(base, method, path, data=data, headers=hdr or None)
    return status, body, heads


def _shopxo_avatar_png_upload(base: str, token: str, route: str) -> tuple[int, str, dict[str, str]]:
    """personal/useravatarupload：multipart 1×1 PNG。"""
    boundary = "----GoshopSmokeAvatar"
    png = bytes.fromhex(
        "89504e470d0a1a0a0000000d49484452000000010000000108060000001f15c489"
        "0000000a49444154789c6300010000050000270000000049454e44ae426082"
    )
    body = b"".join(
        [
            f"--{boundary}\r\n".encode(),
            b'Content-Disposition: form-data; name="file"; filename="a.png"\r\n',
            b"Content-Type: image/png\r\n\r\n",
            png,
            f"\r\n--{boundary}--\r\n".encode(),
        ]
    )
    path = "/api.php?s=" + urllib.parse.quote(route) + "&token=" + urllib.parse.quote(token)
    hdr = {"Content-Type": f"multipart/form-data; boundary={boundary}"}
    return _req(base, "POST", path, data=body, headers=hdr)


def _classify_request(route: str, ctx: dict[str, str]) -> tuple[str, str, str | None, dict | None]:
    """返回 method, extra_qs, form_body, json_body。"""
    t = ctx["token"]
    tag = ctx["tag"]
    a = ctx["order_a"]
    b = ctx["order_b"]
    addr = ctx["addr_id"]
    cart = ctx["cart_id"]  # buy_index / cart_stock
    cart_del = ctx["cart_del_id"]
    pay = ctx["payment_id"]

    if route in PNG_ROUTES:
        return "GET", "&type=smoke", None, None

    json_post: dict[str, dict] = {
        "orderaftersale/create": {},
        "forminputdata/save": {"form_id": 1, "data": "{}"},
        "safety/loginpwdupdate": {
            "old_password": "smoke123456",
            "new_password": "smoke123457",
        },
        "usergoodscomments/save": {"order_item_id": 999999, "rating": 5, "content": "smoke"},
    }
    if route in json_post:
        return "POST", "", None, json_post[route]

    post_form: dict[str, str] = {
        "article/datalist": "id=1&page=1",
        "cart/save": "goods_id=1&sku_id=1&quantity=1",
        "cart/delete": f"ids={cart_del}",
        "user/login": "accounts=nobody&pwd=badsecret",
        "user/reg": f"accounts=nobody_{tag}&pwd=badsecret&verify=nocode",
        "user/forgetpwd": "accounts=nobody&verify=nocode&pwd=newbadsecret",
        "user/appmobilebind": "mobile=13800000000&verify=000000",
        "user/appemailbind": "email=test@t.t&verify=000000",
        "user/onekeyusermobilebind": "encrypted_data=x&iv=x",
        "user/userbasereg": "field=test",
        "user/loginverifysend": "accounts=nobody@test.com",
        "user/regverifysend": "accounts=nobody@test.com",
        "user/forgetpwdverifysend": "accounts=nobody@test.com",
        "user/appmobilebindverifysend": "mobile=13800000000",
        "user/appemailbindverifysend": "email=test@t.t",
        "forminput/verifysend": "accounts=13800000000",
        "order/commentssave": f"id={a}&goods_id=[1]&rating=[5]&content=[\"smoke\"]&images=[]",
        "buy/add": f"goods_id=1&sku_id=1&stock=1&address_id={addr}",
        "order/pay": f"ids={a}&payment_id={pay}",
        "personal/save": "nickname=smoke",
        "useraddress/save": "id=0&name=测&phone=13800000001&province=北京市&city=北京市&district=朝阳区&detail=x",
        "useraddress/delete": "id=999999",
        "useraddress/outsystemadd": "name=测&phone=13800000002&province=北京市&city=北京市&district=朝阳区&detail=y",
        "orderaftersale/delivery": "id=999999&express_name=sf&express_no=1",
        "ueditor/index": "action=noop",
        "cashier/paydata": "authcode=fake&order_no=NONE",
    }
    if route in post_form:
        return "POST", "", post_form[route], None

    get_extra: dict[str, str] = {
        "goods/detail": "&goods_id=1",
        "goods/favor": "&id=1",
        "goods/specdetail": "&id=1&goods_id=1",
        "goods/spectype": "&id=1",
        "goods/stock": "&id=1",
        "goods/goodsscore": "&id=1",
        "goods/comments": "&id=1",
        "order/detail": f"&id={a}",
        "order/paycheck": f"&id={a}",
        "order/cancel": f"&id={b}",
        "order/collect": f"&id={a}",
        "order/delete": f"&id={a}",
        "order/comments": f"&id={a}",
        "buy/index": f"&ids={cart}",
        "cart/stock": f"&id={cart}",
        "search/index": "&wd=手机",
        "search/datalist": "&wd=手机&page=1",
        "search/start": "&wd=手机",
        "orderaftersale/detail": "&id=999999",
        "orderaftersale/cancel": "&id=999999",
        "useraddress/detail": f"&id={addr}",
        "userintegral/index": "&page=1",
        "usergoodscomments/detail": "&id=1",
        "usergoodscomments/delete": "&id=999999",
        "usergoodsfavor/cancel": "&id=1",
        "usergoodsbrowse/delete": "&id=999999",
        "message/index": "&page=1",
        "forminput/index": "&id=1",
        "forminputdata/index": "&form_id=1&page=1",
        "forminputdata/delete": "&id=999999",
        "forminputdata/detail": "&id=999999",
        "diy/index": "&id=1",
        "design/index": "&id=0",
        "customview/index": "&id=999999",
        "paylog/detail": "&id=1",
        "article/detail": "&id=1",
        "agreement/index": "&document=userregister",
        "region/index": "&pid=0",
        "region/codedata": "&id=110000",
        "user/appminiuserauth": "&code=fake",
        "user/appminiuserinfo": "&openid=fake",
        "user/onekeyusermobiledecrypt": "&encryptedData=x&iv=x",
        "user/tokenuserinfo": "",
        "user/center": "",
    }
    return "GET", get_extra.get(route, ""), None, None


def main() -> int:
    base = os.environ.get("BASE", "http://127.0.0.1:8080").rstrip("/")
    auth = _load_auth_required()
    routes = _load_routes()
    skip_setup = os.environ.get("GOSHOP_SMOKE_SKIP_SETUP") == "1"
    existing = os.environ.get("TOKEN", "").strip()

    try:
        payment_id = _payments_offline_id(base)
    except Exception as e:
        print(f"无法取得支付方式: {e}", file=sys.stderr)
        return 2

    if existing and skip_setup:
        token = existing
        ctx = {
            "token": token,
            "addr_id": "1",
            "cart_id": "1",
            "cart_del_id": "1",
            "order_a": "1",
            "order_b": "1",
            "payment_id": payment_id,
            "tag": "skip",
        }
        print("# 使用环境变量 TOKEN + GOSHOP_SMOKE_SKIP_SETUP=1（单号占位可能不准）")
    else:
        try:
            token, addr_id, id_stock, id_del, oa, ob, tag = _setup_rest(base)
        except Exception as e:
            print(f"REST 准备数据失败: {e}", file=sys.stderr)
            return 2
        ctx = {
            "token": token,
            "addr_id": addr_id,
            "cart_id": id_stock,
            "cart_del_id": id_del,
            "order_a": oa,
            "order_b": ob,
            "payment_id": payment_id,
            "tag": tag,
        }

    fails: list[str] = []
    for route in routes:
        need_auth = route in auth
        tok = ctx["token"] if need_auth else ""
        meth, xq, form, jsn = _classify_request(route, ctx)
        if route == "personal/useravatarupload":
            status, body, hdrs = _shopxo_avatar_png_upload(base, tok, route)
        else:
            status, body, hdrs = _shopxo(
                base, tok, route, method=meth, extra_qs=xq, form_body=form, json_body=jsn
            )

        if route in PNG_ROUTES:
            ct = hdrs.get("Content-Type", "")
            if status != 200 or "image/png" not in ct.lower():
                fails.append(f"{route}: 期望 PNG 200, got HTTP {status} ct={ct!r}")
            continue

        if UNKNOWN_ROUTE_MARK in body:
            fails.append(f"{route}: 仍返回「{UNKNOWN_ROUTE_MARK}」")
            continue
        if status >= 500:
            fails.append(f"{route}: HTTP {status}")
            continue

    if fails:
        print(f"\n失败 {len(fails)} / {len(routes)}:\n", file=sys.stderr)
        for ln in fails:
            print(ln, file=sys.stderr)
        return 1
    print(f"OK ShopXO s= 全量冒烟通过: {len(routes)} 条（BASE={base}）")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
