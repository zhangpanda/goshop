package service

import (
	"crypto/hmac"
	"crypto/sha1"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net"
	"net/http"
	"net/smtp"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/zhangpanda/goshop/internal/app"
	"github.com/zhangpanda/goshop/internal/model"
)

// ========== 短信（阿里云 SMS） ==========

func SendSms(phone, templateCode, templateParam string) error {
	accessKey := GetConfig("common_sms_apikey")
	accessSecret := GetConfig("common_sms_secret")
	signName := GetConfig("common_sms_sign")
	if accessKey == "" || accessSecret == "" {
		app.Must().DB.Create(&model.SmsLog{Phone: phone, Content: templateParam, Type: templateCode, Status: 0})
		return nil
	}

	params := map[string]string{
		"AccessKeyId": accessKey, "Action": "SendSms", "Format": "JSON",
		"PhoneNumbers": phone, "SignName": signName,
		"SignatureMethod": "HMAC-SHA1", "SignatureVersion": "1.0",
		"SignatureNonce": fmt.Sprintf("%d%d", time.Now().UnixNano(), rand.Intn(9999)),
		"TemplateCode":   templateCode, "TemplateParam": templateParam,
		"Timestamp": time.Now().UTC().Format("2006-01-02T15:04:05Z"),
		"Version":   "2017-05-25",
	}
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var qs strings.Builder
	for i, k := range keys {
		if i > 0 {
			qs.WriteByte('&')
		}
		qs.WriteString(smsEncode(k) + "=" + smsEncode(params[k]))
	}
	mac := hmac.New(sha1.New, []byte(accessSecret+"&"))
	mac.Write([]byte("GET&%2F&" + smsEncode(qs.String())))
	params["Signature"] = base64.StdEncoding.EncodeToString(mac.Sum(nil))

	v := url.Values{}
	for k, val := range params {
		v.Set(k, val)
	}
	resp, err := (&http.Client{Timeout: 10 * time.Second}).Get("https://dysmsapi.aliyuncs.com/?" + v.Encode())
	if err != nil {
		app.Must().DB.Create(&model.SmsLog{Phone: phone, Content: templateParam, Type: templateCode, Status: 0})
		return err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var res struct {
		Code    string `json:"Code"`
		Message string `json:"Message"`
	}
	json.Unmarshal(body, &res)

	ok := res.Code == "OK"
	app.Must().DB.Create(&model.SmsLog{Phone: phone, Content: templateParam, Type: templateCode, Status: boolToInt8(ok)})
	if !ok {
		return fmt.Errorf("短信发送失败: %s", res.Message)
	}
	return nil
}

func smsEncode(s string) string {
	e := url.QueryEscape(s)
	e = strings.ReplaceAll(e, "+", "%20")
	e = strings.ReplaceAll(e, "*", "%2A")
	e = strings.ReplaceAll(e, "%7E", "~")
	return e
}

func boolToInt8(b bool) int8 {
	if b {
		return 1
	}
	return 0
}

// ========== 邮件（SMTP） ==========

func SendEmail(to, subject, body string) error {
	host := GetConfig("common_email_smtp_host")
	port := GetConfig("common_email_smtp_port")
	account := GetConfig("common_email_smtp_account")
	password := GetConfig("common_email_smtp_pwd")
	fromName := GetConfig("common_email_smtp_name")
	if host == "" || account == "" {
		app.Must().DB.Create(&model.EmailLog{Email: to, Title: subject, Content: body, Status: 0})
		return nil
	}
	if port == "" {
		port = "465"
	}
	if fromName == "" {
		fromName = "GoShop"
	}

	msg := fmt.Sprintf("From: %s <%s>\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n%s",
		fromName, account, to, subject, body)

	addr := net.JoinHostPort(host, port)
	var err error
	if port == "465" {
		err = sendMailSSL(addr, host, account, password, to, msg)
	} else {
		auth := smtp.PlainAuth("", account, password, host)
		err = smtp.SendMail(addr, auth, account, []string{to}, []byte(msg))
	}

	app.Must().DB.Create(&model.EmailLog{Email: to, Title: subject, Content: body, Status: boolToInt8(err == nil)})
	return err
}

func sendMailSSL(addr, host, user, pass, to, msg string) error {
	conn, err := tls.Dial("tcp", addr, &tls.Config{ServerName: host})
	if err != nil {
		return err
	}
	c, err := smtp.NewClient(conn, host)
	if err != nil {
		return err
	}
	defer c.Close()
	if err = c.Auth(smtp.PlainAuth("", user, pass, host)); err != nil {
		return err
	}
	if err = c.Mail(user); err != nil {
		return err
	}
	if err = c.Rcpt(to); err != nil {
		return err
	}
	w, err := c.Data()
	if err != nil {
		return err
	}
	w.Write([]byte(msg))
	w.Close()
	return c.Quit()
}

func errStr(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}
