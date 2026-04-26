package service

import (
	"errors"

	"github.com/zhangpanda/goshop/global"
	"github.com/zhangpanda/goshop/internal/model"
	"github.com/zhangpanda/goshop/pkg/auth"
	"golang.org/x/crypto/bcrypt"
)

type RegisterReq struct {
	Username string `json:"username" form:"username" binding:"required,min=3,max=32"`
	Password string `json:"password" form:"password" binding:"required,min=6,max=64"`
	Nickname string `json:"nickname" form:"nickname"`
}

type LoginReq struct {
	Username string `json:"username" form:"username" binding:"required"`
	Password string `json:"password" form:"password" binding:"required"`
}

type LoginResp struct {
	Token string     `json:"token"`
	User  model.User `json:"user"`
}

func Register(req *RegisterReq) (*model.User, error) {
	var count int64
	global.DB.Model(&model.User{}).Where("username = ?", req.Username).Count(&count)
	if count > 0 {
		return nil, errors.New("用户名已存在")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := model.User{
		Username: req.Username,
		Password: string(hash),
		Nickname: req.Nickname,
	}
	if err := global.DB.Create(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func Login(req *LoginReq) (*LoginResp, error) {
	var user model.User
	if err := global.DB.Where("username = ?", req.Username).First(&user).Error; err != nil {
		return nil, errors.New("用户不存在")
	}
	if user.Status == 0 {
		return nil, errors.New("账号已禁用")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, errors.New("密码错误")
	}

	token, err := auth.GenerateToken(user.ID, false, global.Cfg.JWT.Secret, global.Cfg.JWT.Expire)
	if err != nil {
		return nil, err
	}
	return &LoginResp{Token: token, User: user}, nil
}

func GetUserByID(id uint) (*model.User, error) {
	var user model.User
	if err := global.DB.First(&user, id).Error; err != nil {
		return nil, errors.New("用户不存在")
	}
	return &user, nil
}
