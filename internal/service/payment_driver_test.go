package service

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"sort"
	"strings"
	"testing"
)

func TestGetPaymentDriver(t *testing.T) {
	for _, name := range []string{"wechat_jsapi", "alipay_pc", "alipay_h5", "offline", "wallet"} {
		if _, err := GetPaymentDriver(name); err != nil {
			t.Errorf("GetPaymentDriver(%q) = %v", name, err)
		}
	}
	if _, err := GetPaymentDriver("nonexistent"); err == nil {
		t.Error("GetPaymentDriver(nonexistent) should fail")
	}
}

func TestAlipayVerifySign(t *testing.T) {
	// Generate test RSA key pair
	privKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	pubDER, _ := x509.MarshalPKIXPublicKey(&privKey.PublicKey)
	pubPEM := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubDER})

	// Build params and sign
	params := map[string]string{
		"out_trade_no": "TEST001",
		"trade_no":     "2026042600001",
		"trade_status": "TRADE_SUCCESS",
		"total_amount": "99.00",
	}
	// Build sign string
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var buf strings.Builder
	for i, k := range keys {
		if i > 0 {
			buf.WriteByte('&')
		}
		buf.WriteString(k + "=" + params[k])
	}
	h := sha256.New()
	h.Write([]byte(buf.String()))
	sig, _ := rsa.SignPKCS1v15(rand.Reader, privKey, crypto.SHA256, h.Sum(nil))
	params["sign"] = base64.StdEncoding.EncodeToString(sig)
	params["sign_type"] = "RSA2"

	if !AlipayVerifySign(params, string(pubPEM)) {
		t.Error("AlipayVerifySign should return true for valid signature")
	}

	// Tamper with data
	params["total_amount"] = "1.00"
	if AlipayVerifySign(params, string(pubPEM)) {
		t.Error("AlipayVerifySign should return false for tampered data")
	}
}

func TestAlipayVerifySign_EmptyKey(t *testing.T) {
	if AlipayVerifySign(map[string]string{"sign": "abc"}, "") {
		t.Error("should fail with empty public key")
	}
}

func TestAlipayVerifySign_NoSign(t *testing.T) {
	if AlipayVerifySign(map[string]string{"foo": "bar"}, "some-key") {
		t.Error("should fail with no sign param")
	}
}

func TestAlipayDriverMethod(t *testing.T) {
	tests := []struct {
		clientType string
		wantMethod string
	}{
		{"pc", "alipay.trade.page.pay"},
		{"h5", "alipay.trade.wap.pay"},
		{"app", "alipay.trade.app.pay"},
	}
	for _, tt := range tests {
		d := &AlipayDriver{ClientType: tt.clientType}
		if got := d.method(); got != tt.wantMethod {
			t.Errorf("AlipayDriver{%s}.method() = %s; want %s", tt.clientType, got, tt.wantMethod)
		}
	}
}

func TestAlipayDriverProductCode(t *testing.T) {
	tests := []struct {
		clientType string
		wantCode   string
	}{
		{"pc", "FAST_INSTANT_TRADE_PAY"},
		{"h5", "QUICK_WAP_WAY"},
		{"app", "QUICK_MSECURITY_PAY"},
	}
	for _, tt := range tests {
		d := &AlipayDriver{ClientType: tt.clientType}
		if got := d.productCode(); got != tt.wantCode {
			t.Errorf("AlipayDriver{%s}.productCode() = %s; want %s", tt.clientType, got, tt.wantCode)
		}
	}
}

func TestOfflineDriverPay(t *testing.T) {
	d := &OfflineDriver{}
	resp, err := d.Pay(nil, &PayDriverReq{OrderNo: "T001"})
	if err != nil {
		t.Fatal(err)
	}
	if resp.TradeNo != "OFFLINE_T001" {
		t.Errorf("TradeNo = %s; want OFFLINE_T001", resp.TradeNo)
	}
}

func TestSandboxDriverPay(t *testing.T) {
	d := &SandboxDriver{Name: "alipay_pc", Real: &AlipayDriver{ClientType: "pc"}}
	resp, err := d.Pay(nil, &PayDriverReq{OrderNo: "T999", Amount: 100, Description: "test"})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(resp.TradeNo, "SANDBOX_") {
		t.Errorf("TradeNo = %s; want SANDBOX_ prefix", resp.TradeNo)
	}
	if !strings.Contains(resp.PayURL, "order_no=T999") {
		t.Errorf("PayURL = %s; want callback URL with order_no", resp.PayURL)
	}
}

func TestSandboxDriverRefund(t *testing.T) {
	d := &SandboxDriver{Name: "wechat_jsapi", Real: &WechatJSAPIDriver{}}
	if err := d.Refund(nil, &RefundDriverReq{OrderNo: "T999"}); err != nil {
		t.Errorf("sandbox refund should succeed, got: %v", err)
	}
}

func TestSandboxDriverWrapsReal(t *testing.T) {
	// Offline driver always succeeds — sandbox should preserve its response data
	d := &SandboxDriver{Name: "offline", Real: &OfflineDriver{}}
	resp, err := d.Pay(nil, &PayDriverReq{OrderNo: "T888", Amount: 50})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(resp.TradeNo, "SANDBOX_") {
		t.Error("should have sandbox trade no")
	}
	// PayURL should be sandbox callback, not empty
	if resp.PayURL == "" {
		t.Error("PayURL should be sandbox callback URL")
	}
}
