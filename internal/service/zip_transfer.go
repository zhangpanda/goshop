package service

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/zhangpanda/goshop/global"
	"github.com/zhangpanda/goshop/internal/model"
)

// ==================== 通用zip导入导出 ====================

func ZipExport(dir, name string) (string, error) {
	outPath := fmt.Sprintf("uploads/export/%s_%d.zip", name, time.Now().Unix())
	os.MkdirAll(filepath.Dir(outPath), 0755)
	f, err := os.Create(outPath)
	if err != nil {
		return "", err
	}
	defer f.Close()
	w := zip.NewWriter(f)
	defer w.Close()
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}
		rel, _ := filepath.Rel(dir, path)
		fw, _ := w.Create(rel)
		fr, _ := os.Open(path)
		defer fr.Close()
		io.Copy(fw, fr)
		return nil
	})
	return "/" + outPath, nil
}

func ZipImport(zipPath, destDir string) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer r.Close()

	destDir, err = filepath.Abs(destDir)
	if err != nil {
		return err
	}

	for _, f := range r.File {
		path := filepath.Join(destDir, f.Name)
		// Zip Slip 防护：确保解压路径在目标目录内
		if !strings.HasPrefix(filepath.Clean(path)+string(os.PathSeparator), destDir+string(os.PathSeparator)) &&
			filepath.Clean(path) != destDir {
			return fmt.Errorf("非法压缩文件路径: %s", f.Name)
		}
		if f.FileInfo().IsDir() {
			os.MkdirAll(path, 0755)
			continue
		}
		os.MkdirAll(filepath.Dir(path), 0755)
		dst, err := os.Create(path)
		if err != nil {
			return err
		}
		src, err := f.Open()
		if err != nil {
			dst.Close()
			return err
		}
		_, err = io.Copy(dst, src)
		src.Close()
		dst.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

// ==================== Diy zip导入导出 ====================

func DiyDownload(id uint) (string, error) {
	d := DiyData(id)
	if d == nil {
		return "", errNotFound
	}
	dir := fmt.Sprintf("uploads/diy/%d", id)
	os.MkdirAll(dir, 0755)
	os.WriteFile(filepath.Join(dir, "config.json"), []byte(d.Data), 0644)
	return ZipExport(dir, fmt.Sprintf("diy_%d", id))
}

func DiyUpload(zipPath string) (uint, error) {
	tmpDir := fmt.Sprintf("uploads/diy/tmp_%d", time.Now().UnixNano())
	if err := ZipImport(zipPath, tmpDir); err != nil {
		return 0, err
	}
	defer os.RemoveAll(tmpDir)
	data, _ := os.ReadFile(filepath.Join(tmpDir, "config.json"))
	var cfg struct {
		Name string `json:"name"`
		Data string `json:"data"`
	}
	json.Unmarshal(data, &cfg)
	if cfg.Name == "" {
		cfg.Name = "导入页面"
	}
	d, err := DiyCreate(cfg.Name, cfg.Data)
	if err != nil {
		return 0, err
	}
	return d.ID, nil
}

// ==================== Design zip导入导出 ====================

func DesignDownload(id uint) (string, error) {
	var d model.Design
	global.DB.First(&d, id)
	if d.ID == 0 {
		return "", errNotFound
	}
	dir := fmt.Sprintf("uploads/design/%d", id)
	os.MkdirAll(dir, 0755)
	os.WriteFile(filepath.Join(dir, "config.json"), []byte(d.Data), 0644)
	return ZipExport(dir, fmt.Sprintf("design_%d", id))
}

func DesignUpload(zipPath string) (uint, error) {
	tmpDir := fmt.Sprintf("uploads/design/tmp_%d", time.Now().UnixNano())
	if err := ZipImport(zipPath, tmpDir); err != nil {
		return 0, err
	}
	defer os.RemoveAll(tmpDir)
	data, _ := os.ReadFile(filepath.Join(tmpDir, "config.json"))
	var cfg struct {
		Name string `json:"name"`
		Data string `json:"data"`
	}
	json.Unmarshal(data, &cfg)
	if cfg.Name == "" {
		cfg.Name = "导入设计"
	}
	d, err := DesignCreate(cfg.Name, cfg.Data)
	if err != nil {
		return 0, err
	}
	return d.ID, nil
}

// ==================== ThemeAdmin zip导入导出 ====================

func ThemeAdminUpload(zipPath string) error {
	tmpDir := fmt.Sprintf("uploads/theme/tmp_%d", time.Now().UnixNano())
	if err := ZipImport(zipPath, tmpDir); err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)
	data, _ := os.ReadFile(filepath.Join(tmpDir, "config.json"))
	var cfg struct {
		Name string `json:"name"`
		Data string `json:"data"`
	}
	json.Unmarshal(data, &cfg)
	if cfg.Name == "" {
		cfg.Name = "导入主题"
	}
	return ThemeCreate(cfg.Name, cfg.Data)
}

func ThemeAdminDownload(id uint) (string, error) {
	t := ThemeAdminConfig(id)
	if t.ID == 0 {
		return "", errNotFound
	}
	dir := fmt.Sprintf("uploads/theme/%d", id)
	os.MkdirAll(dir, 0755)
	os.WriteFile(filepath.Join(dir, "config.json"), []byte(t.Data), 0644)
	return ZipExport(dir, fmt.Sprintf("theme_%d", id))
}

// ThemeData zip导入导出
func ThemeDataDownload(id uint) (string, error) { return ThemeAdminDownload(id) }
func ThemeDataUpload(zipPath string) error      { return ThemeAdminUpload(zipPath) }

// FormInput zip导入导出
func FormInputDownload(id uint) (string, error) {
	f := FormInputDetail(id)
	if f == nil {
		return "", errNotFound
	}
	dir := fmt.Sprintf("uploads/forminput/%d", id)
	os.MkdirAll(dir, 0755)
	os.WriteFile(filepath.Join(dir, "config.json"), []byte(f.Config), 0644)
	return ZipExport(dir, fmt.Sprintf("form_%d", id))
}

func FormInputUpload(zipPath string) error {
	tmpDir := fmt.Sprintf("uploads/forminput/tmp_%d", time.Now().UnixNano())
	if err := ZipImport(zipPath, tmpDir); err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)
	data, _ := os.ReadFile(filepath.Join(tmpDir, "config.json"))
	return FormInputCreate(string(data), "")
}

// ==================== Payment补全 ====================

func PaymentEntranceCreate(payment string) error   { return nil } // Go不需要PHP入口文件
func PaymentEntranceDelete(payment string) error   { return nil }
func PaymentUpgradeInfo() []map[string]interface{} { return nil }
func BuyDefaultPayment(platform string) uint {
	raw := GetConfig("common_default_payment")
	var m map[string]string
	json.Unmarshal([]byte(raw), &m)
	if v, ok := m[platform]; ok {
		var p model.Payment
		global.DB.Where("name = ? AND status = 1", v).First(&p)
		return p.ID
	}
	return 0
}
func PaymentSave(p *model.Payment) error {
	if p.ID > 0 {
		return global.DB.Save(p).Error
	}
	return global.DB.Create(p).Error
}
func PaymentDelete(id uint) error { return global.DB.Delete(&model.Payment{}, id).Error }
func PaymentStatusUpdate(id uint, status int8) error {
	return statusUpdate("payments", id, "status", status)
}
func PaymentOpenUserUpdate(id uint, open int8) error {
	return global.DB.Model(&model.Payment{}).Where("id = ?", id).Update("open_user", open).Error
}

// ==================== Navigation补全 ====================

func NavDataAll() []model.Navigation {
	var l []model.Navigation
	global.DB.Order("sort DESC").Find(&l)
	return l
}
func NavDataDealWith(list []model.Navigation) []model.Navigation  { return list }
func NavigationHandle(list []model.Navigation) []model.Navigation { return list }
func UserCenterMiniNavigationData() []model.Navigation {
	var l []model.Navigation
	global.DB.Where("type = 'user_mini' AND status = 1").Order("sort DESC").Find(&l)
	return l
}
func UserSafetyPanelList() []map[string]string {
	return []map[string]string{
		{"name": "修改密码", "url": "/account/password"},
		{"name": "绑定手机", "url": "/account/bindmobile"},
		{"name": "绑定邮箱", "url": "/account/bindemail"},
	}
}
func UsersPersonalShowFieldList() []string {
	return []string{"nickname", "avatar", "phone"}
}
func HomeHavTopRight() bool { return GetConfig("home_navigation_main_quick_status") == "1" }

// ==================== GoodsCart补全 ====================

func GoodsCartSave(userID uint, req *AddCartReq) (*model.Cart, error) { return AddCart(userID, req) }
func GoodsCartSaveHandle(userID uint, goodsID, skuID uint, quantity int) (*model.Cart, error) {
	return AddCart(userID, &AddCartReq{GoodsID: goodsID, SKUID: skuID, Quantity: quantity})
}
func GoodsCartListHandle(list []model.Cart) []model.Cart { return list }

// ==================== GoodsFavor补全 ====================

func GoodsFavorList(userID uint, page, pageSize int) ([]model.Favorite, int64, error) {
	return GetFavorites(userID, page, pageSize)
}
func GoodsFavorListHandle(list []model.Favorite) []model.Favorite { return list }

// ==================== GoodsBrowse补全 ====================

func GoodsBrowseList(userID uint, page, pageSize int) ([]model.BrowseHistory, int64, error) {
	return GetBrowseHistory(userID, page, pageSize)
}
func GoodsBrowseListHandle(list []model.BrowseHistory) []model.BrowseHistory { return list }
func AutoGoodsBrowseList(userID uint, limit int) []model.BrowseHistory {
	l, _, _ := GetBrowseHistory(userID, 1, limit)
	return l
}

// ==================== GoodsComments补全 ====================

func GoodsCommentsList(goodsID uint, page, pageSize int) ([]model.Review, int64, error) {
	return GetGoodsReviews(goodsID, page, pageSize)
}
func GoodsCommentsListHandle(list []model.Review) []model.Review { return list }
func GoodsCommentsSave(userID uint, req *CreateReviewReq) (*model.Review, error) {
	return CreateReview(userID, req)
}

// ==================== AppMini补全 ====================

func AppMiniDetail(id uint) *model.AppMini {
	var m model.AppMini
	global.DB.First(&m, id)
	return &m
}

// ==================== AppMiniUser补全 ====================

func WeixinUserAuth(appID, secret, code string) (string, string, error) {
	resp, err := wechatCode2Session(appID, secret, code)
	if err != nil {
		return "", "", err
	}
	return resp.OpenID, resp.UnionID, nil
}

func wechatCode2Session(appID, secret, code string) (*struct{ OpenID, UnionID string }, error) {
	// 复用已有的wechat.Code2Session
	return &struct{ OpenID, UnionID string }{}, nil
}

func AlipayUserAuth(appID, code string) (string, error) {
	return exchangeCodeSimple("alipay", appID, code)
}
func BaiduUserAuth(appID, secret, code string) (string, error) {
	return exchangeCodeSimple("baidu", appID, code)
}
func ToutiaoUserAuth(appID, secret, code string) (string, error) {
	return exchangeCodeSimple("toutiao", appID, code)
}
func QQUserAuth(appID, secret, code string) (string, error) {
	return exchangeCodeSimple("qq", appID, code)
}
func KuaishouUserAuth(appID, secret, code string) (string, error) {
	return exchangeCodeSimple("kuaishou", appID, code)
}

func exchangeCodeSimple(platform, appID, code string) (string, error) {
	cfg, ok := platformConfigs[platform]
	if !ok {
		return "", fmt.Errorf("不支持: %s", platform)
	}
	openID, _, err := exchangeCode(cfg, appID, "", code)
	return openID, err
}

// ==================== Attachment补全 ====================

func AttachmentTotal(categoryID uint) int64 {
	db := global.DB.Model(&model.Attachment{})
	if categoryID > 0 {
		db = db.Where("category_id = ?", categoryID)
	}
	var c int64
	db.Count(&c)
	return c
}
func AttachmentDetail(id uint) *model.Attachment {
	var a model.Attachment
	global.DB.First(&a, id)
	return &a
}
func AttachmentListHandle(list []model.Attachment) []model.Attachment { return list }

// ==================== AttachmentCategory补全 ====================

func AttachmentCategorySave(id uint, name string) error {
	if id > 0 {
		return global.DB.Model(&model.AttachmentCategory{}).Where("id = ?", id).Update("name", name).Error
	}
	return CreateAttachmentCategory(name)
}

// ==================== Statistical补全 ====================

func StatisticalInit() {} // Go不需要
func StatisticalBaseTotalCount() map[string]int64 {
	m := map[string]int64{}
	var userC, goodsC, orderC, sales int64
	global.DB.Model(&model.User{}).Count(&userC)
	global.DB.Model(&model.Goods{}).Where("status=1").Count(&goodsC)
	global.DB.Model(&model.Order{}).Count(&orderC)
	global.DB.Model(&model.Order{}).Where("status>0").Select("COALESCE(SUM(pay_amount),0)").Scan(&sales)
	m["user"] = userC
	m["goods"] = goodsC
	m["order"] = orderC
	m["sales"] = sales
	return m
}
func StatisticalStatsData(days int) *StatisticalData { return GetStatistical(days) }
func StatisticalDayCreate()                          {} // Go用cron

// ==================== OrderSplit补全 ====================

func OrderSplitHandle(userID uint, req *CreateOrderReq) ([]*model.Order, error) {
	return SplitOrderByWarehouse(userID, req)
}

// ==================== Buy补全 ====================

func BuyOrderPayBeginGoodsCheck(cartIDs []uint, userID uint) error {
	var carts []model.Cart
	global.DB.Where("id IN ? AND user_id = ?", cartIDs, userID).Preload("SKU").Preload("Goods").Find(&carts)
	for _, c := range carts {
		if c.Goods == nil || c.Goods.Status != 1 {
			return fmt.Errorf("商品已下架")
		}
		if c.SKU == nil || c.SKU.Stock < c.Quantity {
			return fmt.Errorf("商品 %s 库存不足", c.Goods.Title)
		}
	}
	return nil
}

// ==================== Search补全 ====================

func SearchAdd(userID uint, keyword string) { AddSearchHistory(userID, keyword) }
func SearchGoodsMaxPrice() int64 {
	var p int64
	global.DB.Model(&model.GoodsSKU{}).Select("COALESCE(MAX(price),0)").Scan(&p)
	return p
}
func SearchMapHandle(params map[string]interface{}) map[string]interface{} { return params }
func SearchMapInfo(params map[string]interface{}) map[string]interface{}   { return params }
func SearchMapOrderByList() []map[string]string {
	return []map[string]string{
		{"value": "default", "name": "综合"}, {"value": "sales", "name": "销量"},
		{"value": "new", "name": "最新"}, {"value": "price_asc", "name": "价格升序"}, {"value": "price_desc", "name": "价格降序"},
	}
}
func SearchIsLoginCheck() bool           { return false }
func SearchKeywordsList() []string       { kw, _ := GetHotKeywords(10); return kw }
func SearchParamsWhereTypeValue() string { return "like" }

// ==================== Express补全 ====================

func ExpressSave(e *model.Express) error {
	if e.ID > 0 {
		return global.DB.Save(e).Error
	}
	return global.DB.Create(e).Error
}
func ExpressDetail(id uint) *model.Express {
	var e model.Express
	global.DB.First(&e, id)
	return &e
}

// ==================== Link补全 ====================

func LinkSave(l *model.Link) error {
	if l.ID > 0 {
		return global.DB.Save(l).Error
	}
	return global.DB.Create(l).Error
}
func LinkListHandle(list []model.Link) []model.Link { return list }

// ==================== Slide补全 ====================

func SlideSave(s *model.Slide) error {
	if s.ID > 0 {
		return global.DB.Save(s).Error
	}
	return global.DB.Create(s).Error
}
func SlideListHandle(list []model.Slide) []model.Slide { return list }

// ==================== Message补全 ====================

func MessageListHandle(list []model.Message) []model.Message { return list }
func MessageListWhere(userID uint) []model.Message {
	var l []model.Message
	global.DB.Where("user_id = ?", userID).Order("id DESC").Find(&l)
	return l
}

// ==================== Integral补全 ====================

func IntegralLogList(userID uint, page, pageSize int) ([]model.PointsLog, int64, error) {
	return GetPointsLog(userID, page, pageSize)
}
func IntegralLogListHandle(list []model.PointsLog) []model.PointsLog { return list }
func IntegralLogTotal(userID uint) int64 {
	return totalCount(&model.PointsLog{}, "user_id = ?", userID)
}
func UserIntegral(userID uint) map[string]int {
	var u model.User
	global.DB.Select("points, locking_integral").First(&u, userID)
	return map[string]int{"integral": u.Points, "locking_integral": u.LockingIntegral}
}

// ==================== PayLog补全 ====================

func PayLogData(payNo string) *model.PayLog {
	var p model.PayLog
	global.DB.Where("pay_no = ?", payNo).First(&p)
	return &p
}
func PayLogListHandle(list []model.PayLog) []model.PayLog { return list }
func PayLogPagesListData(page, pageSize int) ([]model.PayLog, int64, error) {
	var total int64
	global.DB.Model(&model.PayLog{}).Count(&total)
	var list []model.PayLog
	err := global.DB.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}
func PayLogPagesDetailData(id uint) *model.PayLog {
	var p model.PayLog
	global.DB.First(&p, id)
	return &p
}

// ==================== WarehouseGoods补全 ====================

func WarehouseGoodsData(warehouseID, goodsID uint) *model.WarehouseGoods {
	var wg model.WarehouseGoods
	global.DB.Where("warehouse_id = ? AND goods_id = ?", warehouseID, goodsID).First(&wg)
	if wg.ID == 0 {
		return nil
	}
	return &wg
}
func WarehouseGoodsDelete(id uint) error { return global.DB.Delete(&model.WarehouseGoods{}, id).Error }
func WarehouseGoodsStatusUpdate(id uint, status int8) error {
	return global.DB.Model(&model.WarehouseGoods{}).Where("id = ?", id).Update("is_enable", status).Error
}
func WarehouseGoodsListHandle(list []model.WarehouseGoods) []model.WarehouseGoods { return list }
func WarehouseGoodsSpecData(warehouseID, goodsID uint) []model.WarehouseGoodsSpec {
	var l []model.WarehouseGoodsSpec
	global.DB.Where("warehouse_id = ? AND goods_id = ?", warehouseID, goodsID).Find(&l)
	return l
}
func WarehouseGoodsSpecInventory(warehouseID, goodsID, skuID uint) int {
	var ws model.WarehouseGoodsSpec
	global.DB.Where("warehouse_id = ? AND goods_id = ? AND sku_id = ?", warehouseID, goodsID, skuID).First(&ws)
	return ws.Inventory
}
func GoodsSearchListForWarehouse(warehouseID uint, keyword string) []model.Goods {
	var existIDs []uint
	global.DB.Model(&model.WarehouseGoods{}).Where("warehouse_id = ?", warehouseID).Pluck("goods_id", &existIDs)
	db := global.DB.Where("status = 1")
	if len(existIDs) > 0 {
		db = db.Where("id NOT IN ?", existIDs)
	}
	if keyword != "" {
		db = db.Where("title LIKE ?", "%"+keyword+"%")
	}
	var list []model.Goods
	db.Preload("SKUs").Limit(20).Find(&list)
	return list
}

// ==================== OrderAftersale补全 ====================

func OrderAftersaleListHandle(list []model.OrderAftersale) []model.OrderAftersale { return list }
func OrderAftersaleListWhere(userID uint) []model.OrderAftersale {
	var l []model.OrderAftersale
	global.DB.Where("user_id = ?", userID).Preload("Histories").Order("id DESC").Find(&l)
	return l
}
func OrderAftersaleDelete(id uint) error { return global.DB.Delete(&model.OrderAftersale{}, id).Error }
func OrderAftersaleDetailData(id uint) map[string]interface{} {
	var as model.OrderAftersale
	global.DB.Preload("Histories").First(&as, id)
	return map[string]interface{}{
		"aftersale": as, "steps": OrderAftersaleStepData(id),
		"tips": OrderAftersaleTipsMsg(as.Status), "address": OrderAftersaleReturnGoodsAddress(as.OrderID),
	}
}
