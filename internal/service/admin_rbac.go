package service

import (
	"encoding/json"
	"errors"

	"github.com/zhangpanda/goshop/global"
	"github.com/zhangpanda/goshop/internal/model"
	"github.com/zhangpanda/goshop/pkg/auth"
	"golang.org/x/crypto/bcrypt"
)

// ========== 请求/响应结构 ==========

type AdminLoginReq struct {
	Username    string `json:"username" binding:"required"`
	Password    string `json:"password" binding:"required"`
	CaptchaKey  string `json:"captcha_key"`
	CaptchaCode string `json:"captcha_code"`
}

type AdminLoginResp struct {
	Token string      `json:"token"`
	Admin model.Admin `json:"admin"`
}

type CreateAdminReq struct {
	Username string `json:"username" binding:"required,min=3,max=32"`
	Password string `json:"password" binding:"required,min=6,max=64"`
	Nickname string `json:"nickname"`
	RoleID   uint   `json:"role_id" binding:"required"`
}

type UpdateAdminStatusReq struct {
	Status int8 `json:"status"`
}

type RoleReq struct {
	Name   string   `json:"name" binding:"required"`
	Desc   string   `json:"desc"`
	Powers []string `json:"powers"`
	Status int8     `json:"status"`
}

// ========== 管理员 ==========

func AdminLogin(req *AdminLoginReq) (*AdminLoginResp, error) {
	var admin model.Admin
	if err := global.DB.Preload("Role").Where("username = ?", req.Username).First(&admin).Error; err != nil {
		return nil, errors.New("管理员不存在")
	}
	if admin.Status == 0 {
		return nil, errors.New("账号已禁用")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(admin.Password), []byte(req.Password)); err != nil {
		return nil, errors.New("密码错误")
	}
	token, err := auth.GenerateToken(admin.ID, true, global.Cfg.JWT.Secret, global.Cfg.JWT.Expire)
	if err != nil {
		return nil, err
	}
	return &AdminLoginResp{Token: token, Admin: admin}, nil
}

func CreateAdmin(req *CreateAdminReq) (*model.Admin, error) {
	var count int64
	global.DB.Model(&model.Admin{}).Where("username = ?", req.Username).Count(&count)
	if count > 0 {
		return nil, errors.New("用户名已存在")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	admin := model.Admin{
		Username: req.Username,
		Password: string(hash),
		Nickname: req.Nickname,
		RoleID:   req.RoleID,
		Status:   1,
	}
	if err := global.DB.Create(&admin).Error; err != nil {
		return nil, err
	}
	return &admin, nil
}

func GetAdminList(page, pageSize int) (int64, []model.Admin, error) {
	var total int64
	var list []model.Admin
	db := global.DB.Model(&model.Admin{})
	db.Count(&total)
	err := db.Preload("Role").Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return total, list, err
}

func UpdateAdminStatus(id uint, status int8) error {
	return global.DB.Model(&model.Admin{}).Where("id = ?", id).Update("status", status).Error
}

// ========== 角色 ==========

func CreateRole(req *RoleReq) (*model.Role, error) {
	powersJSON, _ := json.Marshal(req.Powers)
	role := model.Role{
		Name:   req.Name,
		Desc:   req.Desc,
		Powers: string(powersJSON),
		Status: 1,
	}
	if err := global.DB.Create(&role).Error; err != nil {
		return nil, err
	}
	return &role, nil
}

func GetRoleList() ([]model.Role, error) {
	var list []model.Role
	err := global.DB.Order("id ASC").Find(&list).Error
	return list, err
}

func UpdateRole(id uint, req *RoleReq) error {
	powersJSON, _ := json.Marshal(req.Powers)
	return global.DB.Model(&model.Role{}).Where("id = ?", id).Updates(map[string]interface{}{
		"name": req.Name, "desc": req.Desc, "powers": string(powersJSON), "status": req.Status,
	}).Error
}

func DeleteRole(id uint) error {
	var count int64
	global.DB.Model(&model.Admin{}).Where("role_id = ?", id).Count(&count)
	if count > 0 {
		return errors.New("该角色下还有管理员，无法删除")
	}
	return global.DB.Delete(&model.Role{}, id).Error
}

func CheckPower(roleID uint, power string) bool {
	var role model.Role
	if err := global.DB.First(&role, roleID).Error; err != nil {
		return false
	}
	if role.Status == 0 {
		return false
	}
	var powers []string
	if err := json.Unmarshal([]byte(role.Powers), &powers); err != nil {
		return false
	}
	for _, p := range powers {
		if p == "*" || p == power {
			return true
		}
	}
	return false
}

// IsSuperAdmin 判断角色是否为超级管理员（Powers包含"*"）
func IsSuperAdmin(roleID uint) bool {
	var role model.Role
	if err := global.DB.First(&role, roleID).Error; err != nil {
		return false
	}
	var powers []string
	if err := json.Unmarshal([]byte(role.Powers), &powers); err != nil {
		return false
	}
	for _, p := range powers {
		if p == "*" {
			return true
		}
	}
	return false
}
