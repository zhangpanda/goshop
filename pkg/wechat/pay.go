package wechat

import (
	"context"
	"crypto/rsa"
	"fmt"
	"os"

	"github.com/wechatpay-apiv3/wechatpay-go/core"
	"github.com/wechatpay-apiv3/wechatpay-go/core/option"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments/app"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments/h5"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments/jsapi"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments/native"
	"github.com/wechatpay-apiv3/wechatpay-go/services/refunddomestic"
	"github.com/wechatpay-apiv3/wechatpay-go/utils"
)

type Client struct {
	AppID     string
	MchID     string
	NotifyURL string
	client    *core.Client
	privateKey *rsa.PrivateKey
}

func NewClient(appID, mchID, mchAPIKey, serialNo, privateKeyPath, notifyURL string) (*Client, error) {
	keyData, err := os.ReadFile(privateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("read private key: %w", err)
	}
	privateKey, err := utils.LoadPrivateKey(string(keyData))
	if err != nil {
		return nil, fmt.Errorf("load private key: %w", err)
	}

	ctx := context.Background()
	opts := []core.ClientOption{
		option.WithWechatPayAutoAuthCipher(mchID, serialNo, privateKey, mchAPIKey),
	}
	client, err := core.NewClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("new wechat client: %w", err)
	}

	return &Client{
		AppID:      appID,
		MchID:      mchID,
		NotifyURL:  notifyURL,
		client:     client,
		privateKey: privateKey,
	}, nil
}

// PrepayRequest 小程序下单参数
type PrepayRequest struct {
	OrderNo     string
	Description string
	Amount      int64  // 分
	OpenID      string // 用户 openid
}

// Prepay 小程序预下单，返回 prepay_id
func (c *Client) Prepay(ctx context.Context, req *PrepayRequest) (*jsapi.PrepayWithRequestPaymentResponse, error) {
	svc := jsapi.JsapiApiService{Client: c.client}
	resp, _, err := svc.PrepayWithRequestPayment(ctx, jsapi.PrepayRequest{
		Appid:       core.String(c.AppID),
		Mchid:       core.String(c.MchID),
		Description: core.String(req.Description),
		OutTradeNo:  core.String(req.OrderNo),
		NotifyUrl:   core.String(c.NotifyURL),
		Amount:      &jsapi.Amount{Total: core.Int64(req.Amount)},
		Payer:       &jsapi.Payer{Openid: core.String(req.OpenID)},
	})
	return resp, err
}

// PrepayH5 H5支付，返回跳转URL
func (c *Client) PrepayH5(ctx context.Context, req *PrepayRequest, clientIP string) (string, error) {
	svc := h5.H5ApiService{Client: c.client}
	resp, _, err := svc.Prepay(ctx, h5.PrepayRequest{
		Appid:       core.String(c.AppID),
		Mchid:       core.String(c.MchID),
		Description: core.String(req.Description),
		OutTradeNo:  core.String(req.OrderNo),
		NotifyUrl:   core.String(c.NotifyURL),
		Amount:      &h5.Amount{Total: core.Int64(req.Amount)},
		SceneInfo:   &h5.SceneInfo{PayerClientIp: core.String(clientIP), H5Info: &h5.H5Info{Type: core.String("Wap")}},
	})
	if err != nil {
		return "", err
	}
	return *resp.H5Url, nil
}

// PrepayApp APP支付，返回调起支付参数
func (c *Client) PrepayApp(ctx context.Context, req *PrepayRequest) (*app.PrepayWithRequestPaymentResponse, error) {
	svc := app.AppApiService{Client: c.client}
	resp, _, err := svc.PrepayWithRequestPayment(ctx, app.PrepayRequest{
		Appid:       core.String(c.AppID),
		Mchid:       core.String(c.MchID),
		Description: core.String(req.Description),
		OutTradeNo:  core.String(req.OrderNo),
		NotifyUrl:   core.String(c.NotifyURL),
		Amount:      &app.Amount{Total: core.Int64(req.Amount)},
	})
	return resp, err
}

// PrepayNative Native扫码支付，返回二维码链接
func (c *Client) PrepayNative(ctx context.Context, req *PrepayRequest) (string, error) {
	svc := native.NativeApiService{Client: c.client}
	resp, _, err := svc.Prepay(ctx, native.PrepayRequest{
		Appid:       core.String(c.AppID),
		Mchid:       core.String(c.MchID),
		Description: core.String(req.Description),
		OutTradeNo:  core.String(req.OrderNo),
		NotifyUrl:   core.String(c.NotifyURL),
		Amount:      &native.Amount{Total: core.Int64(req.Amount)},
	})
	if err != nil {
		return "", err
	}
	return *resp.CodeUrl, nil
}

// Refund 退款
type RefundRequest struct {
	OrderNo  string
	RefundNo string
	Total    int64 // 原订单金额(分)
	Refund   int64 // 退款金额(分)
	Reason   string
}

func (c *Client) Refund(ctx context.Context, req *RefundRequest) (*refunddomestic.Refund, error) {
	svc := refunddomestic.RefundsApiService{Client: c.client}
	resp, _, err := svc.Create(ctx, refunddomestic.CreateRequest{
		OutTradeNo:  core.String(req.OrderNo),
		OutRefundNo: core.String(req.RefundNo),
		Reason:      core.String(req.Reason),
		Amount: &refunddomestic.AmountReq{
			Total:   core.Int64(req.Total),
			Refund:  core.Int64(req.Refund),
			Currency: core.String("CNY"),
		},
	})
	return resp, err
}
