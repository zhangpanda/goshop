package service

import (
	"errors"
	"fmt"

	"github.com/zhangpanda/goshop/global"
	"github.com/zhangpanda/goshop/internal/model"
	"github.com/zhangpanda/goshop/pkg/auth"
	"github.com/zhangpanda/goshop/pkg/wechat"
)

type WxLoginReq struct {
	Code     string `json:"code" binding:"required"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
}

type WxLoginResp struct {
	Token  string     `json:"token"`
	User   model.User `json:"user"`
	IsNew  bool       `json:"is_new"`
}

func WxLogin(req *WxLoginReq) (*WxLoginResp, error) {
	cfg := global.Cfg.Wechat
	if cfg.AppID == "" || cfg.AppSecret == "" {
		return nil, errors.New("微信小程序未配置")
	}

	session, err := wechat.Code2Session(cfg.AppID, cfg.AppSecret, req.Code)
	if err != nil {
		return nil, fmt.Errorf("微信登录失败: %w", err)
	}

	var user model.User
	isNew := false
	global.DB.Where("open_id = ?", session.OpenID).Find(&user)
	if user.ID == 0 {
		// 新用户自动注册
		user = model.User{
			OpenID:   session.OpenID,
			UnionID:  session.UnionID,
			Nickname: req.Nickname,
			Avatar:   req.Avatar,
			Status:   1,
		}
		if err := global.DB.Create(&user).Error; err != nil {
			return nil, err
		}
		isNew = true
	} else {
		// 更新昵称头像
		if req.Nickname != "" || req.Avatar != "" {
			updates := map[string]interface{}{}
			if req.Nickname != "" {
				updates["nickname"] = req.Nickname
			}
			if req.Avatar != "" {
				updates["avatar"] = req.Avatar
			}
			global.DB.Model(&user).Updates(updates)
		}
	}

	token, err := auth.GenerateToken(user.ID, false, global.Cfg.JWT.Secret, global.Cfg.JWT.Expire)
	if err != nil {
		return nil, err
	}

	return &WxLoginResp{Token: token, User: user, IsNew: isNew}, nil
}
