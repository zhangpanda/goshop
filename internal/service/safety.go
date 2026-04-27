package service

import (
	crand "crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/zhangpanda/goshop/global"
	"github.com/zhangpanda/goshop/internal/model"
	"golang.org/x/crypto/bcrypt"
)

func cryptoRandInt(max int) int {
	n, _ := crand.Int(crand.Reader, big.NewInt(int64(max)))
	return int(n.Int64())
}

// SendVerifyCode 发送验证码（短信/邮件）
func SendVerifyCode(account, typ string) error {
	code := fmt.Sprintf("%06d", cryptoRandInt(1000000))
	vc := model.VerifyCode{
		Account:  account,
		Code:     code,
		Type:     typ,
		ExpireAt: time.Now().Add(5 * time.Minute),
	}
	global.DB.Create(&vc)
	// 发送短信或邮件
	tpl := SmsTemplateValue(typ)
	if tpl == "" {
		tpl = typ
	}
	param := fmt.Sprintf(`{"code":"%s"}`, code)
	if strings.Contains(account, "@") {
		SendEmail(account, "验证码", fmt.Sprintf("您的验证码是：<b>%s</b>，5分钟内有效。", code))
	} else {
		SendSms(account, tpl, param)
	}
	return nil
}

// CheckVerifyCode 校验验证码
func CheckVerifyCode(account, code, typ string) error {
	var vc model.VerifyCode
	global.DB.Where("account = ? AND type = ? AND status = 0 AND expire_at > ?", account, typ, time.Now()).
		Order("id DESC").Find(&vc)
	if vc.ID == 0 {
		return errors.New("验证码不存在或已过期")
	}
	if vc.Code != code {
		return errors.New("验证码错误")
	}
	global.DB.Model(&vc).Update("status", 1)
	return nil
}

// UpdatePassword 修改密码
type UpdatePwdReq struct {
	OldPassword string `json:"old_password" form:"old_password" binding:"required"`
	NewPassword string `json:"new_password" form:"new_password" binding:"required,min=6,max=64"`
}

func UpdatePassword(userID uint, req *UpdatePwdReq) error {
	var user model.User
	if err := global.DB.First(&user, userID).Error; err != nil {
		return errors.New("用户不存在")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.OldPassword)); err != nil {
		return errors.New("原密码错误")
	}
	hash, _ := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	return global.DB.Model(&user).Update("password", string(hash)).Error
}

// ForgetPassword 忘记密码（通过验证码重置）
type ForgetPwdReq struct {
	Account  string `json:"account" form:"account" binding:"required"`
	Code     string `json:"code" form:"code" binding:"required"`
	Password string `json:"password" form:"password" binding:"required,min=6,max=64"`
}

func ForgetPassword(req *ForgetPwdReq) error {
	if err := CheckVerifyCode(req.Account, req.Code, "forget"); err != nil {
		return err
	}
	var user model.User
	global.DB.Where("phone = ? OR username = ?", req.Account, req.Account).Find(&user)
	if user.ID == 0 {
		return errors.New("用户不存在")
	}
	hash, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	return global.DB.Model(&user).Update("password", string(hash)).Error
}

// BindMobile 绑定手机号
type BindMobileReq struct {
	Mobile string `json:"mobile" form:"mobile" binding:"required"`
	Code   string `json:"code" form:"code" binding:"required"`
}

func BindMobile(userID uint, req *BindMobileReq) error {
	if err := CheckVerifyCode(req.Mobile, req.Code, "bind"); err != nil {
		return err
	}
	return global.DB.Model(&model.User{}).Where("id = ?", userID).Update("phone", req.Mobile).Error
}

// GetUserPlatforms 获取用户绑定的平台
func GetUserPlatforms(userID uint) ([]model.UserPlatform, error) {
	var list []model.UserPlatform
	return list, global.DB.Where("user_id = ?", userID).Find(&list).Error
}
