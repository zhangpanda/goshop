package shopxo

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/zhangpanda/goshop/internal/handler"
	"github.com/zhangpanda/goshop/internal/service"
	"github.com/zhangpanda/goshop/pkg/response"
)

func sxCartIndex(c *gin.Context) {
	list, _ := service.GetCartList(c.GetUint("user_id"))
	response.OK(c, list)
}
func sxCartSave(c *gin.Context) {
	var req service.AddCartReq
	c.ShouldBind(&req)
	cart, err := service.AddCart(c.GetUint("user_id"), &req)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	response.OK(c, cart)
}
func sxCartDelete(c *gin.Context) {
	ids := parseUintSlice(c.PostForm("ids"))
	service.DeleteCart(c.GetUint("user_id"), ids)
	response.OK(c, nil)
}
func sxCartStock(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Query("id"), 10, 64)
	stock, _ := service.GoodsCartStock(uint(id))
	response.OK(c, map[string]int{"stock": stock})
}

func sxBuyIndex(c *gin.Context) {
	ids := parseUintSlice(c.Query("ids"))
	buyType := c.DefaultQuery("buy_type", "cart")
	resp, err := service.BuyOrderInit(c.GetUint("user_id"), ids, buyType)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	response.OK(c, resp)
}
func sxBuyAdd(c *gin.Context) {
	// buy/add：goods_id+sku_id 即时购（兼容层先加购物车再下单）
	var req struct {
		GoodsID   uint `json:"goods_id" form:"goods_id"`
		SKUID     uint `json:"sku_id" form:"sku_id"`
		Stock     int  `json:"stock" form:"stock"`
		AddressID uint `json:"address_id" form:"address_id"`
	}
	c.ShouldBind(&req)
	if req.Stock == 0 {
		req.Stock = 1
	}
	userID := c.GetUint("user_id")
	cart, err := service.AddCart(userID, &service.AddCartReq{GoodsID: req.GoodsID, SKUID: req.SKUID, Quantity: req.Stock})
	if err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	addrID := req.AddressID
	order, err := service.CreateOrder(userID, &service.CreateOrderReq{AddressID: &addrID, CartIDs: []uint{cart.ID}})
	if err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	response.OK(c, order)
}

func sxOrderIndex(c *gin.Context) {
	userID := c.GetUint("user_id")
	var req service.OrderListReq
	if err := c.ShouldBind(&req); err != nil {
		req.Page = 1
		req.PageSize = 20
	}
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}
	payload, err := ShopXOOrderIndexPayload(userID, &req)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.OK(c, payload)
}

func sxOrderDetail(c *gin.Context) {
	id := getID(c)
	order, err := service.GetOrderDetail(c.GetUint("user_id"), id)
	if err != nil {
		response.Fail(c, http.StatusNotFound, err.Error())
		return
	}
	payRows, _ := ShopXOUserPaymentRows()
	if payRows == nil {
		payRows = []map[string]interface{}{}
	}
	detail := ShopXOOrderDetailView(order)
	response.OK(c, map[string]interface{}{
		"data":               detail,
		"operate":            service.OrderOperateButtons(order),
		"steps":              service.OrderStepData(order),
		"payment_list":       payRows,
		"default_payment_id": DefaultPaymentIDForShopXO(),
		"status_tips":        "",
		"site_fictitious":    nil,
	})
}

func sxOrderPay(c *gin.Context) {
	userID := c.GetUint("user_id")
	var body struct {
		IDs       string `form:"ids" json:"ids"`
		ID        string `form:"id" json:"id"`
		PaymentID uint   `form:"payment_id" json:"payment_id"`
		OpenID    string `form:"openid" json:"openid"`
		ReturnURL string `form:"return_url" json:"return_url"`
	}
	_ = c.ShouldBind(&body)
	idsStr := strings.TrimSpace(body.IDs)
	if idsStr == "" {
		idsStr = strings.TrimSpace(body.ID)
	}
	ids := parseUintSlice(idsStr)
	if len(ids) == 0 {
		response.Fail(c, http.StatusBadRequest, "请选择订单")
		return
	}
	if body.PaymentID == 0 {
		response.Fail(c, http.StatusBadRequest, "请选择支付方式")
		return
	}
	payload, outerMsg, err := ShopXOCompatUnifiedPay(userID, ids, body.PaymentID, body.OpenID, body.ReturnURL, c.ClientIP())
	if err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	if outerMsg != nil {
		c.JSON(http.StatusOK, gin.H{"code": 0, "msg": *outerMsg, "data": payload})
		return
	}
	response.OK(c, payload)
}
func sxOrderPayCheck(c *gin.Context) {
	id := getID(c)
	_, err := service.OrderPayCheck(c.GetUint("user_id"), id)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	response.OK(c, nil)
}

// sxCashierPayData 对应 cashier/paydata：小程序内 wx.login 的 code 换 openid 后拉 JSAPI 参数（需先 order/pay 创建 PayLog）。
func sxCashierPayData(c *gin.Context) {
	var body struct {
		AuthCode string `form:"authcode" json:"authcode"`
		OrderNo  string `form:"order_no" json:"order_no"`
	}
	_ = c.ShouldBind(&body)
	if body.AuthCode == "" {
		body.AuthCode = c.Query("authcode")
	}
	if body.OrderNo == "" {
		body.OrderNo = c.Query("order_no")
	}
	payload, err := ShopXOCashierPayData(body.AuthCode, body.OrderNo)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	response.OK(c, payload)
}

func sxOrderCancel(c *gin.Context) {
	id := getID(c)
	if err := service.CancelOrder(c.GetUint("user_id"), id); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	response.OK(c, nil)
}
func sxOrderCollect(c *gin.Context) {
	id := getID(c)
	if err := service.ConfirmReceive(c.GetUint("user_id"), id); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	response.OK(c, nil)
}
func sxOrderDelete(c *gin.Context) {
	id := getID(c)
	service.DeleteOrder(c.GetUint("user_id"), id)
	response.OK(c, nil)
}
func sxOrderComments(c *gin.Context) {
	id := getID(c)
	items := service.OrderItemList(id)
	rows := make([]map[string]interface{}, 0, len(items))
	for i := range items {
		it := &items[i]
		rows = append(rows, map[string]interface{}{
			"goods_id":  it.GoodsID,
			"images":    it.Image,
			"goods_url": fmt.Sprintf("/pages/goods-detail/goods-detail?id=%d", it.GoodsID),
		})
	}
	response.OK(c, map[string]interface{}{
		"data": map[string]interface{}{
			"id":    id,
			"items": rows,
		},
		"editor_path_type": "common",
	})
}

func sxAftersaleIndex(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	list, total, _ := service.GetAftersaleList(c.GetUint("user_id"), page, 20)
	response.OK(c, map[string]interface{}{"total": total, "data": list})
}
func sxAftersaleDetail(c *gin.Context) {
	id := getID(c)
	as, err := service.GetAftersaleDetail(c.GetUint("user_id"), id)
	if err != nil {
		response.Fail(c, http.StatusNotFound, err.Error())
		return
	}
	response.OK(c, service.OrderAftersaleDetailData(as.ID))
}
func sxAftersaleCreate(c *gin.Context) { handler.AftersaleCreate(c) }
func sxAftersaleDelivery(c *gin.Context) {
	id := getID(c)
	var req service.AftersaleDeliveryReq
	c.ShouldBind(&req)
	if err := service.AftersaleDelivery(c.GetUint("user_id"), id, &req); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	response.OK(c, nil)
}
func sxAftersaleCancel(c *gin.Context) {
	id := getID(c)
	service.AftersaleCancel(c.GetUint("user_id"), id)
	response.OK(c, nil)
}
