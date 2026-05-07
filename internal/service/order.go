package service

import (
	crypto_rand "crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/zhangpanda/goshop/global"
	"github.com/zhangpanda/goshop/internal/model"
	"github.com/zhangpanda/goshop/internal/repository"
	"gorm.io/gorm"
)

// OrderService encapsulates order business logic with injected dependencies.
type OrderService struct {
	db      *gorm.DB
	orders  repository.OrderRepo
	carts   repository.CartRepo
	address repository.AddressRepo
	sku     repository.SKURepo
}

// NewOrderService creates an OrderService with explicit dependencies.
func NewOrderService(db *gorm.DB, orders repository.OrderRepo, carts repository.CartRepo, addr repository.AddressRepo, sku repository.SKURepo) *OrderService {
	return &OrderService{db: db, orders: orders, carts: carts, address: addr, sku: sku}
}

// DefaultOrderService returns an OrderService using the global DB and repository registry.
// orderSvc is the singleton instance, initialized via InitOrderService from main.
var orderSvc *OrderService

// InitOrderService creates the singleton OrderService. Call once at startup.
func InitOrderService(svc *OrderService) {
	orderSvc = svc
}

// DefaultOrderService returns the singleton. Falls back to creating one from globals if not initialized.
func DefaultOrderService() *OrderService {
	if orderSvc != nil {
		return orderSvc
	}
	return NewOrderService(global.DB, repository.Repos.Order, repository.Repos.Cart, repository.Repos.Address, repository.Repos.SKU)
}

type CreateOrderReq struct {
	AddressID    *uint  `json:"address_id"`
	CartIDs      []uint `json:"cart_ids" form:"cart_ids" binding:"required,min=1"`
	UserCouponID *uint  `json:"user_coupon_id"`
	OrderModel   int8   `json:"order_model"`
	Remark       string `json:"remark"`
}

type OrderListReq struct {
	Status   *int8 `form:"status" json:"status"`
	Page     int   `form:"page" json:"page"`
	PageSize int   `form:"page_size" json:"page_size"`
}

type OrderListResp struct {
	Total int64         `json:"total"`
	List  []model.Order `json:"list"`
}

func generateOrderNo() string {
	b := make([]byte, 4)
	crypto_rand.Read(b)
	return fmt.Sprintf("%s%08x", time.Now().Format("20060102150405"), b)
}

// --- Package-level functions delegate to DefaultOrderService for backward compat ---

func CreateOrder(userID uint, req *CreateOrderReq) (*model.Order, error) {
	return DefaultOrderService().CreateOrder(userID, req)
}

func GetOrderList(userID uint, req *OrderListReq) (*OrderListResp, error) {
	return DefaultOrderService().GetOrderList(userID, req)
}

func GetOrderDetail(userID, orderID uint) (*model.Order, error) {
	return DefaultOrderService().GetOrderDetail(userID, orderID)
}

func CancelOrder(userID, orderID uint) error {
	return DefaultOrderService().CancelOrder(userID, orderID)
}

// --- Methods on OrderService ---

func (s *OrderService) CreateOrder(userID uint, req *CreateOrderReq) (*model.Order, error) {
	// Address
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
			addr, err := s.address.GetByIDAndUser(*req.AddressID, userID)
			if err != nil {
				return nil, errors.New("地址不存在")
			}
			addrJSON, _ = json.Marshal(addr)
		}
	}

	// Cart
	carts, err := s.carts.FindByIDsAndUser(req.CartIDs, userID)
	if err != nil {
		return nil, err
	}
	if len(carts) == 0 {
		return nil, errors.New("购物车为空")
	}

	// Build items
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
			GoodsID: c.GoodsID, SKUID: c.SKUID,
			Title: c.Goods.Title, Image: c.Goods.MainImage,
			SkuName: c.SKU.Name, Price: price, Quantity: c.Quantity,
		})
	}

	// Coupon
	payAmount := totalAmount
	var usedCoupon *model.UserCoupon
	if req.UserCouponID != nil && *req.UserCouponID > 0 {
		var uc model.UserCoupon
		if err := s.db.Preload("Coupon").Where("id = ? AND user_id = ? AND status = 0", *req.UserCouponID, userID).First(&uc).Error; err != nil {
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
		OrderNo: generateOrderNo(), UserID: userID,
		TotalAmount: totalAmount, PayAmount: payAmount,
		Status: model.OrderStatusPending, OrderModel: req.OrderModel,
		Remark: req.Remark, Address: string(addrJSON),
	}

	if IsBookingMode() {
		order.Status = model.OrderStatusBooking
	}
	if req.OrderModel == model.OrderModelPickup {
		order.ExtractionCode = fmt.Sprintf("%06d", time.Now().UnixNano()%1000000)
	}
	if req.OrderModel == model.OrderModelVirtual && len(carts) > 0 && carts[0].Goods != nil {
		order.FictitiousValue = carts[0].Goods.Detail
	}

	// Transaction
	tx := s.db.Begin()

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

	// Deduct stock via repo
	for _, c := range carts {
		if err := s.sku.DeductStock(tx, c.SKUID, c.Quantity); err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("商品 %s 库存不足", c.Goods.Title)
		}
	}

	// Clear cart
	if err := s.carts.DeleteByIDsAndUser(tx, req.CartIDs, userID); err != nil {
		tx.Rollback()
		return nil, err
	}

	// Mark coupon used
	if usedCoupon != nil {
		now := time.Now()
		tx.Model(usedCoupon).Updates(map[string]interface{}{
			"status": 1, "order_id": order.ID, "used_at": &now,
		})
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	AddOrderStatusHistory(order.ID, -1, order.Status, "订单创建", "系统")
	SaveOrderCurrency(order.ID, order.PayAmount)

	order.Items = items
	return &order, nil
}

func (s *OrderService) GetOrderList(userID uint, req *OrderListReq) (*OrderListResp, error) {
	offset := (req.Page - 1) * req.PageSize
	list, total, err := s.orders.List(userID, req.Status, offset, req.PageSize)
	if err != nil {
		return nil, err
	}
	return &OrderListResp{Total: total, List: list}, nil
}

func (s *OrderService) GetOrderDetail(userID, orderID uint) (*model.Order, error) {
	return s.orders.GetByUserAndID(userID, orderID)
}

func (s *OrderService) CancelOrder(userID, orderID uint) error {
	order, err := s.orders.GetByUserAndID(userID, orderID)
	if err != nil {
		return errors.New("订单不存在")
	}
	if order.Status != model.OrderStatusPending {
		return errors.New("只有待付款订单可以取消")
	}

	tx := s.db.Begin()

	if err := s.orders.UpdateStatus(tx, orderID, model.OrderStatusCancelled); err != nil {
		tx.Rollback()
		return err
	}

	// Restore stock
	var items []model.OrderItem
	tx.Where("order_id = ?", orderID).Find(&items)
	for _, item := range items {
		if err := s.sku.RestoreStock(tx, item.SKUID, item.Quantity); err != nil {
			tx.Rollback()
			return err
		}
	}

	if err := tx.Commit().Error; err != nil {
		return err
	}

	AddOrderStatusHistory(orderID, model.OrderStatusPending, model.OrderStatusCancelled, "用户取消订单", "用户")
	return nil
}
