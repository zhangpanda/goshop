package service

import (
	"github.com/zhangpanda/goshop/internal/app"
	"github.com/zhangpanda/goshop/internal/model"
)

type AnswerCreateReq struct {
	GoodsID uint   `json:"goods_id"`
	Title   string `json:"title" binding:"required"`
	Content string `json:"content"`
}

func CreateAnswer(userID uint, req *AnswerCreateReq) (*model.Answer, error) {
	a := model.Answer{UserID: userID, GoodsID: req.GoodsID, Title: req.Title, Content: req.Content}
	return &a, app.Must().DB.Create(&a).Error
}

func GetAnswerList(goodsID uint, page, pageSize int) ([]model.Answer, int64, error) {
	var total int64
	db := app.Must().DB.Model(&model.Answer{})
	if goodsID > 0 {
		db = db.Where("goods_id = ?", goodsID)
	}
	db.Count(&total)
	var list []model.Answer
	err := db.Preload("User").Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}

func ReplyAnswer(id uint, reply string) error {
	return app.Must().DB.Model(&model.Answer{}).Where("id = ?", id).Updates(map[string]interface{}{"reply": reply, "status": 1}).Error
}

func DeleteAnswer(id uint) error {
	return app.Must().DB.Delete(&model.Answer{}, id).Error
}
