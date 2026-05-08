package service

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"testing"

	"github.com/zhangpanda/goshop/config"
	"github.com/zhangpanda/goshop/internal/app"
)

func TestAlipayGatewayDoURL(t *testing.T) {
	app.Must().Cfg = &config.Config{}
	if g := AlipayGatewayDoURL(); g != alipayGatewayDoDefault {
		t.Fatalf("default: %s", g)
	}
	app.Must().Cfg.Alipay.GatewayURL = "https://openapi.alipaydev.com"
	if g := AlipayGatewayDoURL(); g != "https://openapi.alipaydev.com/gateway.do" {
		t.Fatalf("append path: %s", g)
	}
	app.Must().Cfg.Alipay.GatewayURL = "https://x.test/gateway.do/"
	if g := AlipayGatewayDoURL(); g != "https://x.test/gateway.do" {
		t.Fatalf("trim slash: %s", g)
	}
}

func TestAlipayVerifyGatewayJSON(t *testing.T) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	pubDER, err := x509.MarshalPKIXPublicKey(&priv.PublicKey)
	if err != nil {
		t.Fatal(err)
	}
	pubPEM := string(pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubDER}))

	inner := `{"code":"10000","msg":"Success"}`
	sum := sha256.Sum256([]byte(inner))
	sig, err := rsa.SignPKCS1v15(rand.Reader, priv, crypto.SHA256, sum[:])
	if err != nil {
		t.Fatal(err)
	}
	signB64 := base64.StdEncoding.EncodeToString(sig)
	body := `{"alipay_trade_query_response":` + inner + `,"sign":"` + signB64 + `","sign_type":"RSA2"}`

	key, raw, err := alipayVerifyGatewayJSON([]byte(body), pubPEM)
	if err != nil || key != "alipay_trade_query_response" || string(raw) != inner {
		t.Fatalf("verify: %v key=%q raw=%s", err, key, raw)
	}

	if _, _, err := alipayVerifyGatewayJSON([]byte(`{"alipay_trade_query_response":`+inner+`,"sign":"wrong"}`), pubPEM); err == nil {
		t.Fatal("expected verify error for bad sign")
	}
}
