package service

import (
	"errors"

	"github.com/zhangpanda/goshop/global"
	"github.com/zhangpanda/goshop/internal/model"
)

type AddressReq struct {
	Name        string  `json:"name" form:"name" binding:"required"`
	Phone       string  `json:"phone" form:"phone" binding:"required"`
	Province    string  `json:"province" form:"province"`
	City        string  `json:"city" form:"city"`
	District    string  `json:"district" form:"district"`
	Detail      string  `json:"detail" form:"detail" binding:"required"`
	Lng         float64 `json:"lng" form:"lng"`
	Lat         float64 `json:"lat" form:"lat"`
	IDCard      string  `json:"id_card" form:"id_card"`
	IDCardFront string  `json:"id_card_front" form:"id_card_front"`
	IDCardBack  string  `json:"id_card_back" form:"id_card_back"`
	IsDefault   bool    `json:"is_default" form:"is_default"`
}

func CreateAddress(userID uint, req *AddressReq) (*model.Address, error) {
	addr := model.Address{
		UserID:      userID,
		Name:        req.Name,
		Phone:       req.Phone,
		Province:    req.Province,
		City:        req.City,
		District:    req.District,
		Detail:      req.Detail,
		Lng:         req.Lng,
		Lat:         req.Lat,
		IDCard:      req.IDCard,
		IDCardFront: req.IDCardFront,
		IDCardBack:  req.IDCardBack,
		IsDefault:   req.IsDefault,
	}
	// 如果设为默认，先取消其他默认
	if req.IsDefault {
		global.DB.Model(&model.Address{}).Where("user_id = ?", userID).Update("is_default", false)
	}
	if err := global.DB.Create(&addr).Error; err != nil {
		return nil, err
	}
	return &addr, nil
}

func GetAddressList(userID uint) ([]model.Address, error) {
	var list []model.Address
	err := global.DB.Where("user_id = ?", userID).Order("is_default DESC, id DESC").Find(&list).Error
	return list, err
}

func UpdateAddress(userID, addrID uint, req *AddressReq) error {
	var addr model.Address
	if err := global.DB.Where("id = ? AND user_id = ?", addrID, userID).First(&addr).Error; err != nil {
		return errors.New("地址不存在")
	}
	if req.IsDefault {
		global.DB.Model(&model.Address{}).Where("user_id = ? AND id != ?", userID, addrID).Update("is_default", false)
	}
	return global.DB.Model(&addr).Updates(map[string]interface{}{
		"name": req.Name, "phone": req.Phone, "province": req.Province,
		"city": req.City, "district": req.District, "detail": req.Detail,
		"lng": req.Lng, "lat": req.Lat, "id_card": req.IDCard,
		"id_card_front": req.IDCardFront, "id_card_back": req.IDCardBack,
		"is_default": req.IsDefault,
	}).Error
}

func DeleteAddress(userID, addrID uint) error {
	return global.DB.Where("id = ? AND user_id = ?", addrID, userID).Delete(&model.Address{}).Error
}
