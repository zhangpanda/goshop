package shopxo

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/zhangpanda/goshop/internal/handler"
	"github.com/zhangpanda/goshop/internal/service"
	"github.com/zhangpanda/goshop/pkg/response"
)

func sxAddressIndex(c *gin.Context) {
	list, _ := service.GetAddressList(c.GetUint("user_id"))
	response.OK(c, list)
}
func sxAddressDetail(c *gin.Context) {
	id := getID(c)
	addr, err := service.GetAddress(c.GetUint("user_id"), id)
	if err != nil {
		response.Fail(c, http.StatusNotFound, err.Error())
		return
	}
	response.OK(c, addr)
}
func sxAddressSave(c *gin.Context) {
	var req service.AddressReq
	c.ShouldBind(&req)
	id, _ := strconv.ParseUint(c.PostForm("id"), 10, 64)
	if id > 0 {
		service.UpdateAddress(c.GetUint("user_id"), uint(id), &req)
	} else {
		service.CreateAddress(c.GetUint("user_id"), &req)
	}
	response.OK(c, nil)
}
func sxAddressDelete(c *gin.Context) {
	id := getID(c)
	service.DeleteAddress(c.GetUint("user_id"), id)
	response.OK(c, nil)
}
func sxAddressSetDefault(c *gin.Context) {
	// 设为默认
	id := getID(c)
	service.UpdateAddress(c.GetUint("user_id"), id, &service.AddressReq{IsDefault: true})
	response.OK(c, nil)
}
func sxAddressExtraction(c *gin.Context) {
	response.OK(c, service.GetSelfExtractionAddressList())
}
func sxAddressOutSystemAdd(c *gin.Context) {
	var req service.AddressReq
	c.ShouldBind(&req)
	service.CreateAddress(c.GetUint("user_id"), &req)
	response.OK(c, nil)
}

func sxPersonalIndex(c *gin.Context) {
	response.OK(c, service.LoginUserInfo(c.GetUint("user_id")))
}
func sxPersonalSave(c *gin.Context) {
	var req service.PersonalSaveReq
	c.ShouldBind(&req)
	service.PersonalSave(c.GetUint("user_id"), &req)
	response.OK(c, nil)
}

func sxLoginPwdUpdate(c *gin.Context) { handler.UpdatePassword(c) }
func sxLogout(c *gin.Context)         { response.OK(c, nil) }

func sxFavorIndex(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	list, total, _ := service.GetFavorites(c.GetUint("user_id"), page, 20)
	response.OK(c, map[string]interface{}{"total": total, "data": list})
}
func sxFavorCancel(c *gin.Context) {
	id := getID(c)
	service.ToggleFavorite(c.GetUint("user_id"), id)
	response.OK(c, nil)
}

func sxBrowseIndex(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	list, total, _ := service.GetBrowseHistory(c.GetUint("user_id"), page, 20)
	response.OK(c, map[string]interface{}{"total": total, "data": list})
}
func sxBrowseDelete(c *gin.Context) {
	ids := parseUintSlice(c.PostForm("ids"))
	service.GoodsBrowseDelete(ids, c.GetUint("user_id"))
	response.OK(c, nil)
}

func sxCommentsIndex(c *gin.Context)  { sxGoodsComments(c) }
func sxCommentsDetail(c *gin.Context) { response.OK(c, nil) }
func sxCommentsSave(c *gin.Context)   { handler.CreateReview(c) }
func sxCommentsDelete(c *gin.Context) {
	id := getID(c)
	if err := service.GoodsCommentsDeleteForUser(id, c.GetUint("user_id")); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	response.OK(c, nil)
}

func sxIntegralIndex(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	list, total, _ := service.GetPointsLog(c.GetUint("user_id"), page, 20)
	integral := service.UserIntegral(c.GetUint("user_id"))
	response.OK(c, map[string]interface{}{"total": total, "data": list, "integral": integral})
}

func sxMessageIndex(c *gin.Context) { handler.GetMessages(c) }

func sxRegionIndex(c *gin.Context) {
	pid, _ := strconv.ParseUint(c.DefaultQuery("parent_id", "0"), 10, 64)
	list, _ := service.GetRegionList(uint(pid))
	response.OK(c, list)
}
func sxRegionAll(c *gin.Context) { response.OK(c, service.RegionAll()) }
func sxRegionCodeData(c *gin.Context) {
	id := getID(c)
	response.OK(c, service.RegionCodeData(id))
}

func sxAgreementIndex(c *gin.Context) {
	name := c.Query("name")
	response.OK(c, service.AgreementGet(name))
}

func sxDiyIndex(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Query("id"), 10, 64)
	if id > 0 {
		data, _ := service.DiyApiCustomInit(uint(id))
		response.OK(c, data)
	} else {
		response.OK(c, service.AppClientHomeDiyData())
	}
}

func sxFormInputIndex(c *gin.Context) {
	id := getID(c)
	response.OK(c, service.FormInputPreview(id))
}
func sxFormInputDataIndex(c *gin.Context) {
	fid, _ := strconv.ParseUint(c.Query("form_id"), 10, 64)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	list, total, _ := service.FormInputDataList(uint(fid), page, 20)
	response.OK(c, map[string]interface{}{"total": total, "data": list})
}
func sxFormInputDataSave(c *gin.Context) { handler.FormInputDataSubmitHandler(c) }
func sxFormInputDataDelete(c *gin.Context) {
	id := getID(c)
	service.FormInputApiDelete(id, c.GetUint("user_id"))
	response.OK(c, nil)
}

func sxFormInputVerifySend(c *gin.Context) {
	account := c.PostForm("accounts")
	if account == "" {
		account = c.PostForm("mobile")
	}
	service.SendVerifyCode(account, "forminput")
	response.OK(c, nil)
}

// ===== 辅助函数 =====

func getID(c *gin.Context) uint {
	id, _ := strconv.ParseUint(c.Query("id"), 10, 64)
	if id == 0 {
		id, _ = strconv.ParseUint(c.PostForm("id"), 10, 64)
	}
	if id == 0 {
		id, _ = strconv.ParseUint(c.Query("goods_id"), 10, 64)
	}
	if id == 0 {
		id, _ = strconv.ParseUint(c.PostForm("order_id"), 10, 64)
	}
	return uint(id)
}

func parseUintSlice(s string) []uint {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	var ids []uint
	for _, p := range parts {
		id, _ := strconv.ParseUint(strings.TrimSpace(p), 10, 64)
		if id > 0 {
			ids = append(ids, uint(id))
		}
	}
	return ids
}
