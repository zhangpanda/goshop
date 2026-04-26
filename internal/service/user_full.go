package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/zhangpanda/goshop/global"
	"github.com/zhangpanda/goshop/internal/model"
	"github.com/zhangpanda/goshop/pkg/auth"
	"golang.org/x/crypto/bcrypt"
)

// UserInfo 获取用户完整信息
func UserInfo(userID uint) map[string]interface{} {
	user, _ := GetUserByID(userID)
	if user == nil {
		return nil
	}
	return map[string]interface{}{
		"user": user, "order_total": OrderStatusGroupTotal(userID),
		"favor_total": GoodsFavorTotal(userID), "browse_total": totalCount(&model.BrowseHistory{}, "user_id = ?", userID),
		"cart_total": GoodsCartTotal(userID), "unread_msg": UnreadCount(userID),
	}
}

// UserBaseInfo 基础信息
func UserBaseInfo(userID uint) *model.User {
	u, _ := GetUserByID(userID)
	return u
}

// GetUserViewInfo 用户展示信息（脱敏）
func GetUserViewInfo(userID uint) map[string]interface{} {
	var u model.User
	global.DB.Select("id, nickname, avatar").First(&u, userID)
	return map[string]interface{}{"id": u.ID, "nickname": u.Nickname, "avatar": u.Avatar}
}

// UserInsert 创建用户
func UserInsert(username, password, nickname string) (*model.User, error) {
	return Register(&RegisterReq{Username: username, Password: password, Nickname: nickname})
}

// UserUpdateHandle 更新用户
func UserUpdateHandle(userID uint, updates map[string]interface{}) error {
	return global.DB.Model(&model.User{}).Where("id = ?", userID).Updates(updates).Error
}

// PersonalSave 个人资料保存
type PersonalSaveReq struct {
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
	Phone    string `json:"phone"`
}

func PersonalSave(userID uint, req *PersonalSaveReq) error {
	return UserSave(userID, req.Nickname, req.Avatar, req.Phone)
}

// UserStatusCheck 用户状态检查
func UserStatusCheck(userID uint) error {
	var u model.User
	global.DB.Select("status").First(&u, userID)
	if u.Status == 0 {
		return errors.New("账号已禁用")
	}
	return nil
}

// UserLoginHandle 登录处理
func UserLoginHandle(user *model.User) (*LoginResp, error) {
	token, err := auth.GenerateToken(user.ID, false, global.Cfg.JWT.Secret, global.Cfg.JWT.Expire)
	if err != nil {
		return nil, err
	}
	return &LoginResp{Token: token, User: *user}, nil
}

// LoginUserInfo 获取当前登录用户信息
func LoginUserInfo(userID uint) *model.User {
	u, _ := GetUserByID(userID)
	return u
}

// TokenUserinfo Token获取用户信息
func TokenUserinfo(token string) (*model.User, error) {
	claims, err := auth.ParseToken(token, global.Cfg.JWT.Secret)
	if err != nil {
		return nil, err
	}
	return GetUserByID(claims.UserID)
}

// UserTokenData 用户Token数据
func UserTokenData(userID uint) (string, error) {
	return auth.GenerateToken(userID, false, global.Cfg.JWT.Secret, global.Cfg.JWT.Expire)
}

// UserTokenUpdate 更新Token
func UserTokenUpdate(userID uint) (string, error) { return UserTokenData(userID) }

// IsExistAccounts 检查账号是否存在
func IsExistAccounts(account string) bool {
	var c int64
	global.DB.Model(&model.User{}).Where("username = ? OR phone = ?", account, account).Count(&c)
	return c > 0
}

// UserLoginAccountsCheck 登录账号检查
func UserLoginAccountsCheck(account string) (*model.User, error) {
	var u model.User
	global.DB.Where("username = ? OR phone = ?", account, account).First(&u)
	if u.ID == 0 {
		return nil, errors.New("用户不存在")
	}
	if u.Status == 0 {
		return nil, errors.New("账号已禁用")
	}
	return &u, nil
}

// UserRegAccountsCheck 注册账号检查
func UserRegAccountsCheck(username string) error {
	if IsExistAccounts(username) {
		return errors.New("账号已存在")
	}
	return nil
}

// UserRegForbidCheck 注册禁止检查
func UserRegForbidCheck() error {
	if GetConfig("site_register_close") == "1" {
		return errors.New("注册已关闭")
	}
	return nil
}

// UserForgetAccountsCheck 忘记密码账号检查
func UserForgetAccountsCheck(account string) (*model.User, error) {
	return UserLoginAccountsCheck(account)
}

// LoginVerifySend 登录验证码发送
func LoginVerifySend(account string) error { return SendVerifyCode(account, "login") }

// RegVerifySend 注册验证码发送
func RegVerifySend(account string) error { return SendVerifyCode(account, "register") }

// ForgetPwdVerifySend 忘记密码验证码发送
func ForgetPwdVerifySend(account string) error { return SendVerifyCode(account, "forget") }

// AppMobileBindVerifySend 手机绑定验证码
func AppMobileBindVerifySend(mobile string) error { return SendVerifyCode(mobile, "bind") }

// AppEmailBindVerifySend 邮箱绑定验证码
func AppEmailBindVerifySend(email string) error { return SendVerifyCode(email, "email_bind") }

// AppEmailBind 邮箱绑定
func AppEmailBind(userID uint, email, code string) error {
	if err := CheckVerifyCode(email, code, "email_bind"); err != nil {
		return err
	}
	return global.DB.Model(&model.User{}).Where("id = ?", userID).Update("email", email).Error
}

// AppAccountsBindhHandle 账号绑定处理
func AppAccountsBindhHandle(userID uint, platform, openID string) error {
	var p model.UserPlatform
	global.DB.Where("user_id = ? AND platform = ?", userID, platform).First(&p)
	if p.ID > 0 {
		return global.DB.Model(&p).Update("openid", openID).Error
	}
	return global.DB.Create(&model.UserPlatform{UserID: userID, Platform: platform, OpenID: openID}).Error
}

// UserOpenidBind OpenID绑定
func UserOpenidBind(userID uint, platform, openID, unionID string) error {
	return global.DB.Create(&model.UserPlatform{UserID: userID, Platform: platform, OpenID: openID, UnionID: unionID}).Error
}

// UserOpenidHandle OpenID处理
func UserOpenidHandle(platform, openID string) *model.UserPlatform {
	var p model.UserPlatform
	global.DB.Where("platform = ? AND openid = ?", platform, openID).First(&p)
	if p.ID == 0 {
		return nil
	}
	return &p
}

// UserUnionidHandle UnionID处理
func UserUnionidHandle(platform, unionID string) *model.UserPlatform {
	var p model.UserPlatform
	global.DB.Where("platform = ? AND unionid = ?", platform, unionID).First(&p)
	if p.ID == 0 {
		return nil
	}
	return &p
}

// MatchingUserPlatformData 匹配用户平台数据
func MatchingUserPlatformData(userID uint) []model.UserPlatform {
	var list []model.UserPlatform
	global.DB.Where("user_id = ?", userID).Find(&list)
	return list
}

// UserPlatformInfo 平台信息
func UserPlatformInfo(userID uint, platform string) *model.UserPlatform {
	var p model.UserPlatform
	global.DB.Where("user_id = ? AND platform = ?", userID, platform).First(&p)
	if p.ID == 0 {
		return nil
	}
	return &p
}

// UserPlatformInsert 平台插入
func UserPlatformInsert(p *model.UserPlatform) error { return global.DB.Create(p).Error }

// UserPlatformUpdate 平台更新
func UserPlatformUpdate(id uint, updates map[string]interface{}) error {
	return global.DB.Model(&model.UserPlatform{}).Where("id = ?", id).Updates(updates).Error
}

// UserReferrerEncryption 推荐人加密
func UserReferrerEncryption(userID uint) string { return fmt.Sprintf("ref_%d", userID) }

// UserReferrerDecrypt 推荐人解密
func UserReferrerDecrypt(ref string) uint {
	var id uint
	fmt.Sscanf(ref, "ref_%d", &id)
	return id
}

// UserNumberCodeCreatedHandle 用户编号生成
func UserNumberCodeCreatedHandle() string {
	return fmt.Sprintf("U%d", time.Now().UnixNano()%1000000000)
}

// UserUniqueMethod 用户唯一性方法
func UserUniqueMethod() string { return "username" }

// UserEntranceLeftData 用户中心左侧数据
func UserEntranceLeftData(userID uint) []map[string]interface{} {
	return []map[string]interface{}{
		{"name": "订单管理", "url": "/account/orders", "count": OrderStatusGroupTotal(userID)["pending"]},
		{"name": "我的收藏", "url": "/account/favorites", "count": GoodsFavorTotal(userID)},
		{"name": "我的消息", "url": "/messages", "count": UnreadCount(userID)},
	}
}

// UserSystemInfo 用户系统信息
func UserSystemInfo(userID uint) map[string]interface{} {
	platforms := MatchingUserPlatformData(userID)
	return map[string]interface{}{"platforms": platforms, "platform_count": len(platforms)}
}

// IsImaVerify 是否需要图形验证码
func IsImaVerify() bool { return GetConfig("site_verify_ima") == "1" }

// AuthUserProgram 用户授权程序
func AuthUserProgram(userID uint) bool { return UserStatusCheck(userID) == nil }

// CacheLoginUserInfo 缓存登录用户信息（Go用JWT无需session缓存）
func CacheLoginUserInfo(userID uint) *model.User { return LoginUserInfo(userID) }

// CacheUserTokenData 缓存Token数据
func CacheUserTokenData(userID uint) (string, error) { return UserTokenData(userID) }

// UserListHandle 用户列表数据处理
func UserListHandle(list []model.User) []model.User {
	for i := range list {
		list[i].Password = ""
	}
	return list
}

// AppUserInfoHandle APP用户信息处理
func AppUserInfoHandle(user *model.User) map[string]interface{} {
	return map[string]interface{}{
		"id": user.ID, "nickname": user.Nickname, "avatar": user.Avatar,
		"phone": user.Phone, "points": user.Points, "wallet": user.WalletBalance,
	}
}

// UserLoginOrRegBackRefererUrl 登录注册后跳转
func UserLoginOrRegBackRefererUrl() string { return "/" }

// Reg 注册（别名）
func Reg(req *RegisterReq) (*model.User, error) { return Register(req) }

// Logout 退出（别名）
func Logout(userID uint) error { return nil } // JWT无状态，客户端删token即可

// ForgetPwd 忘记密码（别名）
func ForgetPwd(req *ForgetPwdReq) error { return ForgetPassword(req) }

// AppMobileBind 手机绑定（别名）
func AppMobileBind(userID uint, req *BindMobileReq) error { return BindMobile(userID, req) }

// AppAccountsBindVerifySendHandle 绑定验证码发送处理
func AppAccountsBindVerifySendHandle(account, typ string) error { return SendVerifyCode(account, typ) }

// UserHandle 用户数据处理（通用）
func UserHandle(user *model.User) {
	user.Password = ""
}

// UserBaseHandle 用户基础数据处理
func UserBaseHandle(user *model.User) {
	UserHandle(user)
}

func init() {
	// 确保bcrypt包被引用
	_ = bcrypt.MinCost
}
