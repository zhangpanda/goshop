package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/zhangpanda/goshop/global"
	"github.com/zhangpanda/goshop/internal/model"
)

// BuyOrderInit 购买页初始化（返回地址+商品+可用优惠券+自提点）
type BuyInitResp struct {
	Address    *model.Address           `json:"address"`
	Goods      []BuyGoodsItem           `json:"goods"`
	Coupons    []model.UserCoupon       `json:"coupons"`
	Extraction []map[string]interface{} `json:"extraction"`
	SiteModel  int8                     `json:"site_model"`
	Total      int64                    `json:"total"`
}
type BuyGoodsItem struct {
	GoodsID  uint   `json:"goods_id"`
	Title    string `json:"title"`
	Image    string `json:"image"`
	SKUName  string `json:"sku_name"`
	Price    int64  `json:"price"`
	Quantity int    `json:"quantity"`
}

func BuyOrderInit(userID uint, cartIDs []uint, buyType string) (*BuyInitResp, error) {
	resp := &BuyInitResp{}
	// 默认地址
	var addr model.Address
	global.DB.Where("user_id = ? AND is_default = true", userID).First(&addr)
	if addr.ID == 0 {
		global.DB.Where("user_id = ?", userID).Order("id DESC").First(&addr)
	}
	if addr.ID > 0 {
		resp.Address = &addr
	}

	// 商品列表
	goods, err := BuyTypeGoodsList(userID, cartIDs, buyType)
	if err != nil {
		return nil, err
	}
	resp.Goods = goods

	// 计算总价
	for _, g := range goods {
		resp.Total += g.Price * int64(g.Quantity)
	}

	// 可用优惠券
	var status int8 = 0
	resp.Coupons, _ = GetMyCoupons(userID, &status)

	// 自提点
	resp.Extraction = GetSelfExtractionAddressList()

	// 站点模式
	siteType := GetConfig("common_site_type")
	switch siteType {
	case "1":
		resp.SiteModel = 1
	case "2":
		resp.SiteModel = 2
	case "3":
		resp.SiteModel = 3
	default:
		resp.SiteModel = 0
	}
	return resp, nil
}

// BuyTypeGoodsList 按购买类型获取商品列表
func BuyTypeGoodsList(userID uint, ids []uint, buyType string) ([]BuyGoodsItem, error) {
	switch buyType {
	case "cart":
		var carts []model.Cart
		global.DB.Where("id IN ? AND user_id = ?", ids, userID).Preload("Goods").Preload("SKU").Find(&carts)
		items := make([]BuyGoodsItem, len(carts))
		for i, c := range carts {
			items[i] = BuyGoodsItem{GoodsID: c.GoodsID, Quantity: c.Quantity}
			if c.Goods != nil {
				items[i].Title = c.Goods.Title
				items[i].Image = c.Goods.MainImage
			}
			if c.SKU != nil {
				items[i].SKUName = c.SKU.Name
				items[i].Price = c.SKU.Price
			}
		}
		return items, nil
	case "goods":
		if len(ids) < 2 {
			return nil, errors.New("参数错误")
		}
		goodsID, skuID := ids[0], ids[1]
		var goods model.Goods
		global.DB.First(&goods, goodsID)
		var sku model.GoodsSKU
		global.DB.First(&sku, skuID)
		return []BuyGoodsItem{{GoodsID: goodsID, Title: goods.Title, Image: goods.MainImage, SKUName: sku.Name, Price: sku.Price, Quantity: 1}}, nil
	}
	return nil, errors.New("不支持的购买类型")
}

// BuyGoodsCheck 购买前商品校验
func BuyGoodsCheck(goodsID, skuID uint, quantity int) error {
	var goods model.Goods
	if err := global.DB.First(&goods, goodsID).Error; err != nil {
		return errors.New("商品不存在")
	}
	if goods.Status != 1 {
		return errors.New("商品已下架")
	}
	var sku model.GoodsSKU
	if err := global.DB.First(&sku, skuID).Error; err != nil {
		return errors.New("规格不存在")
	}
	if sku.Stock < quantity {
		return fmt.Errorf("库存不足(剩余%d)", sku.Stock)
	}
	return nil
}

// BuyDataStorage 购买数据临时存储（Redis）
func BuyDataStorage(userID uint, data interface{}) error {
	if global.RDB == nil {
		return nil
	}
	b, _ := json.Marshal(data)
	return global.RDB.Set(context.Background(), fmt.Sprintf("buy_data_%d", userID), b, 30*time.Minute).Err()
}

func BuyDataRead(userID uint) (map[string]interface{}, error) {
	if global.RDB == nil {
		return nil, nil
	}
	val, err := global.RDB.Get(context.Background(), fmt.Sprintf("buy_data_%d", userID)).Result()
	if err != nil {
		return nil, err
	}
	var data map[string]interface{}
	json.Unmarshal([]byte(val), &data)
	return data, nil
}

func BuyDataDelete(userID uint) {
	if global.RDB == nil {
		return
	}
	global.RDB.Del(context.Background(), fmt.Sprintf("buy_data_%d", userID))
}

// SingleOrderPayBeginCheck 单订单支付前校验
func SingleOrderPayBeginCheck(userID, orderID uint) error {
	var order model.Order
	if err := global.DB.Where("id = ? AND user_id = ?", orderID, userID).First(&order).Error; err != nil {
		return errors.New("订单不存在")
	}
	if order.Status != model.OrderStatusPending {
		return errors.New("订单状态不允许支付")
	}
	return nil
}

// MoreOrderPayBeginCheck 多订单合并支付前校验
func MoreOrderPayBeginCheck(userID uint, orderIDs []uint) error {
	for _, id := range orderIDs {
		if err := SingleOrderPayBeginCheck(userID, id); err != nil {
			return fmt.Errorf("订单%d: %s", id, err.Error())
		}
	}
	return nil
}
