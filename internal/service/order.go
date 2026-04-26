package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/zhangpanda/goshop/global"
	"github.com/zhangpanda/goshop/internal/model"
)

type CreateOrderReq struct {
	AddressID    *uint  `json:"address_id"` // 快递/同城必填，自提可选，虚拟不需要
	CartIDs      []uint `json:"cart_ids" form:"cart_ids" binding:"required,min=1"`
	UserCouponID *uint  `json:"user_coupon_id"`
	OrderModel   int8   `json:"order_model"` // 0快递 1同城 2自提 3虚拟
	Remark       string `json:"remark"`
}

type OrderListReq struct {
	Status   *int8 `form:"status"`
	Page     int   `form:"page,default=1"`
	PageSize int   `form:"page_size,default=20"`
}

type OrderListResp struct {
	Total int64         `json:"total"`
	List  []model.Order `json:"list"`
}

func generateOrderNo() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

func CreateOrder(userID uint, req *CreateOrderReq) (*model.Order, error) {
	// 地址处理：快递/同城必须有地址，自提可选，虚拟不需要
	var addrJSON []byte
	if req.OrderModel == model.OrderModelVirtual {
		addrJSON = []byte("{}")
	} else {
		if req.AddressID == nil || *req.AddressID == 0 {
			if req.OrderModel == model.OrderModelExpress || req.OrderModel == model.OrderModelLocal {
				return nil, errors.New("请选择收货地址")
			}
			addrJSON = []byte("{}")
		} else {
			var addr model.Address
			if err := global.DB.Where("id = ? AND user_id = ?", *req.AddressID, userID).First(&addr).Error; err != nil {
				return nil, errors.New("地址不存在")
			}
			addrJSON, _ = json.Marshal(addr)
		}
	}

	// 查购物车
	var carts []model.Cart
	if err := global.DB.Where("id IN ? AND user_id = ?", req.CartIDs, userID).
		Preload("Goods").Preload("SKU").Find(&carts).Error; err != nil {
		return nil, err
	}
	if len(carts) == 0 {
		return nil, errors.New("购物车为空")
	}

	// 构建订单，优先使用促销价
	var totalAmount int64
	var items []model.OrderItem
	for _, c := range carts {
		if c.SKU == nil || c.Goods == nil {
			return nil, errors.New("商品或SKU不存在")
		}
		if c.SKU.Stock < c.Quantity {
			return nil, fmt.Errorf("商品 %s 库存不足", c.Goods.Title)
		}
		price := c.SKU.Price
		if promoPrice, err := GetPromoPrice(c.SKUID); err == nil {
			price = promoPrice
		}
		totalAmount += price * int64(c.Quantity)
		items = append(items, model.OrderItem{
			GoodsID:  c.GoodsID,
			SKUID:    c.SKUID,
			Title:    c.Goods.Title,
			Image:    c.Goods.MainImage,
			SkuName:  c.SKU.Name,
			Price:    price,
			Quantity: c.Quantity,
		})
	}

	// 优惠券抵扣
	payAmount := totalAmount
	var usedCoupon *model.UserCoupon
	if req.UserCouponID != nil && *req.UserCouponID > 0 {
		var uc model.UserCoupon
		if err := global.DB.Preload("Coupon").Where("id = ? AND user_id = ? AND status = 0", *req.UserCouponID, userID).First(&uc).Error; err != nil {
			return nil, errors.New("优惠券不可用")
		}
		if uc.Coupon == nil || time.Now().After(uc.Coupon.EndTime) {
			return nil, errors.New("优惠券已过期")
		}
		discount, err := CalcCouponDiscount(uc.Coupon, totalAmount)
		if err != nil {
			return nil, err
		}
		payAmount = totalAmount - discount
		usedCoupon = &uc
	}

	order := model.Order{
		OrderNo:     generateOrderNo(),
		UserID:      userID,
		TotalAmount: totalAmount,
		PayAmount:   payAmount,
		Status:      model.OrderStatusPending,
		OrderModel:  req.OrderModel,
		Remark:      req.Remark,
		Address:     string(addrJSON),
	}

	// 预约模式：状态为待确认
	if IsBookingMode() {
		order.Status = model.OrderStatusBooking
	}

	// 自提模式生成6位自提码
	if req.OrderModel == model.OrderModelPickup {
		order.ExtractionCode = fmt.Sprintf("%06d", time.Now().UnixNano()%1000000)
	}

	// 虚拟商品：从商品的 FictitiousValue 字段获取
	if req.OrderModel == model.OrderModelVirtual {
		if len(carts) > 0 && carts[0].Goods != nil {
			order.FictitiousValue = carts[0].Goods.Detail // 可扩展为独立字段
		}
	}

	tx := global.DB.Begin()

	if err := tx.Create(&order).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	for i := range items {
		items[i].OrderID = order.ID
	}
	if err := tx.Create(&items).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// 扣库存
	for _, c := range carts {
		result := tx.Model(&model.GoodsSKU{}).Where("id = ? AND stock >= ?", c.SKUID, c.Quantity).
			Update("stock", global.DB.Raw("stock - ?", c.Quantity))
		if result.RowsAffected == 0 {
			tx.Rollback()
			return nil, fmt.Errorf("商品 %s 库存不足", c.Goods.Title)
		}
	}

	// 清购物车
	tx.Where("id IN ? AND user_id = ?", req.CartIDs, userID).Delete(&model.Cart{})

	// 标记优惠券已使用
	if usedCoupon != nil {
		now := time.Now()
		tx.Model(usedCoupon).Updates(map[string]interface{}{
			"status":   1,
			"order_id": order.ID,
			"used_at":  &now,
		})
	}

	tx.Commit()

	// 记录订单状态历史
	AddOrderStatusHistory(order.ID, -1, order.Status, "订单创建", "系统")

	// 保存订单货币信息
	SaveOrderCurrency(order.ID, order.PayAmount)

	order.Items = items
	return &order, nil
}

func GetOrderList(userID uint, req *OrderListReq) (*OrderListResp, error) {
	db := global.DB.Model(&model.Order{}).Where("user_id = ?", userID)
	if req.Status != nil {
		db = db.Where("status = ?", *req.Status)
	}

	var total int64
	db.Count(&total)

	var list []model.Order
	offset := (req.Page - 1) * req.PageSize
	err := db.Preload("Items").Order("id DESC").
		Offset(offset).Limit(req.PageSize).Find(&list).Error

	return &OrderListResp{Total: total, List: list}, err
}

func GetOrderDetail(userID, orderID uint) (*model.Order, error) {
	var order model.Order
	err := global.DB.Where("id = ? AND user_id = ?", orderID, userID).
		Preload("Items").First(&order).Error
	if err != nil {
		return nil, errors.New("订单不存在")
	}
	return &order, nil
}

func CancelOrder(userID, orderID uint) error {
	var order model.Order
	if err := global.DB.Where("id = ? AND user_id = ?", orderID, userID).First(&order).Error; err != nil {
		return errors.New("订单不存在")
	}
	if order.Status != model.OrderStatusPending {
		return errors.New("只有待付款订单可以取消")
	}

	tx := global.DB.Begin()

	tx.Model(&order).Update("status", model.OrderStatusCancelled)

	// 恢复库存
	var items []model.OrderItem
	tx.Where("order_id = ?", orderID).Find(&items)
	for _, item := range items {
		tx.Model(&model.GoodsSKU{}).Where("id = ?", item.SKUID).
			Update("stock", global.DB.Raw("stock + ?", item.Quantity))
	}

	tx.Commit()

	AddOrderStatusHistory(orderID, model.OrderStatusPending, model.OrderStatusCancelled, "用户取消订单", "用户")

	return nil
}
