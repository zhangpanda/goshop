package initialize

import (
	"fmt"
	"log"
	"time"

	"github.com/zhangpanda/goshop/global"
	"github.com/zhangpanda/goshop/internal/model"
)

// InitDefaultSeedData 初始化展示数据（分类、品牌、商品+SKU、文章、优惠券）
func InitDefaultSeedData() {
	var count int64
	global.DB.Model(&model.Goods{}).Count(&count)
	if count > 0 {
		return
	}

	// ========== 商品分类 ==========
	categories := []model.Category{
		{ID: 1, ParentID: 0, Name: "手机数码", Sort: 100, Status: 1},
		{ID: 2, ParentID: 0, Name: "电脑办公", Sort: 90, Status: 1},
		{ID: 3, ParentID: 0, Name: "时尚服饰", Sort: 80, Status: 1},
		{ID: 4, ParentID: 0, Name: "名品箱包", Sort: 70, Status: 1},
		{ID: 5, ParentID: 0, Name: "运动户外", Sort: 60, Status: 1},
		{ID: 6, ParentID: 0, Name: "家用电器", Sort: 50, Status: 1},
		// 二级分类
		{ID: 11, ParentID: 1, Name: "手机", Sort: 50, Status: 1},
		{ID: 12, ParentID: 1, Name: "耳机", Sort: 40, Status: 1},
		{ID: 13, ParentID: 1, Name: "智能手表", Sort: 30, Status: 1},
		{ID: 14, ParentID: 1, Name: "平板电脑", Sort: 20, Status: 1},
		{ID: 21, ParentID: 2, Name: "笔记本", Sort: 50, Status: 1},
		{ID: 22, ParentID: 2, Name: "台式机", Sort: 40, Status: 1},
		{ID: 23, ParentID: 2, Name: "显示器", Sort: 30, Status: 1},
		{ID: 31, ParentID: 3, Name: "男装", Sort: 50, Status: 1},
		{ID: 32, ParentID: 3, Name: "女装", Sort: 40, Status: 1},
		{ID: 33, ParentID: 3, Name: "鞋靴", Sort: 30, Status: 1},
		{ID: 41, ParentID: 4, Name: "女包", Sort: 50, Status: 1},
		{ID: 42, ParentID: 4, Name: "男包", Sort: 40, Status: 1},
		{ID: 43, ParentID: 4, Name: "旅行箱", Sort: 30, Status: 1},
		{ID: 51, ParentID: 5, Name: "运动鞋", Sort: 50, Status: 1},
		{ID: 52, ParentID: 5, Name: "运动服", Sort: 40, Status: 1},
		{ID: 53, ParentID: 5, Name: "户外装备", Sort: 30, Status: 1},
		{ID: 61, ParentID: 6, Name: "空调", Sort: 50, Status: 1},
		{ID: 62, ParentID: 6, Name: "冰箱", Sort: 40, Status: 1},
		{ID: 63, ParentID: 6, Name: "洗衣机", Sort: 30, Status: 1},
	}
	global.DB.Create(&categories)

	// ========== 品牌 ==========
	brands := []model.Brand{
		{ID: 1, Name: "Apple", Desc: "Apple 设计并打造了 iPhone、iPad、Mac、Apple Watch 等产品", Sort: 100, Status: 1},
		{ID: 2, Name: "华为", Desc: "全球领先的 ICT 基础设施和智能终端提供商", Sort: 90, Status: 1},
		{ID: 3, Name: "小米", Desc: "创新科技企业，智能手机和智能硬件", Sort: 80, Status: 1},
		{ID: 4, Name: "联想", Desc: "全球领先的个人电脑及智能设备制造商", Sort: 70, Status: 1},
		{ID: 5, Name: "Nike", Desc: "全球著名的体育运动品牌", Sort: 60, Status: 1},
		{ID: 6, Name: "GUCCI", Desc: "意大利奢侈品牌，创立于1921年", Sort: 50, Status: 1},
		{ID: 7, Name: "COACH", Desc: "美国纽约轻奢时尚品牌", Sort: 40, Status: 1},
		{ID: 8, Name: "海尔", Desc: "全球领先的美好生活和数字化转型解决方案服务商", Sort: 30, Status: 1},
	}
	global.DB.Create(&brands)

	// ========== 商品 + SKU ==========
	type goodsSeed struct {
		goods model.Goods
		skus  []model.GoodsSKU
	}

	now := time.Now()
	img := func(n int) string { return fmt.Sprintf("/uploads/seed/product-%d.jpg", n) }
	detail := func(title, desc string) string {
		return fmt.Sprintf(`<h2>%s</h2><p>%s</p>`, title, desc)
	}
	seeds := []goodsSeed{
		// --- 手机数码 ---
		{
			goods: model.Goods{CategoryID: 11, BrandID: 1, Title: "iPhone 16 Pro Max", Subtitle: "强大超乎想象", MainImage: img(1), Detail: detail("iPhone 16 Pro Max", "搭载全新 A18 Pro 芯片，配备 4800 万像素四合一超级长焦镜头，支持 5 倍光学变焦。全新钛金属设计，6.9 英寸超视网膜 XDR 显示屏，峰值亮度高达 2000 尼特。"), Status: 1, Sort: 100, SalesCount: 2680, AccessCount: 58320},
			skus: []model.GoodsSKU{
				{Name: "256GB 沙漠色", Price: 999900, Stock: 200, Coding: "IP16PM-256-DT", Status: 1},
				{Name: "512GB 沙漠色", Price: 1149900, Stock: 150, Coding: "IP16PM-512-DT", Status: 1},
				{Name: "1TB 黑色钛金属", Price: 1349900, Stock: 80, Coding: "IP16PM-1T-BT", Status: 1},
			},
		},
		{
			goods: model.Goods{CategoryID: 11, BrandID: 2, Title: "HUAWEI Mate 70 Pro+", Subtitle: "鸿蒙智慧旗舰", MainImage: img(2), Detail: detail("HUAWEI Mate 70 Pro+", "搭载全新麒麟芯片，运行 HarmonyOS NEXT，超感知 XMAGE 影像系统，支持十档可变光圈。全系标配卫星通信功能。"), Status: 1, Sort: 95, SalesCount: 1890, AccessCount: 42100},
			skus: []model.GoodsSKU{
				{Name: "256GB 雅金", Price: 849900, Stock: 180, Coding: "HW-M70PP-256", Status: 1},
				{Name: "512GB 雅黑", Price: 949900, Stock: 120, Coding: "HW-M70PP-512", Status: 1},
			},
		},
		{
			goods: model.Goods{CategoryID: 12, BrandID: 1, Title: "AirPods Pro 3", Subtitle: "入耳式主动降噪，入主新声代", MainImage: img(3), Detail: detail("AirPods Pro 3", "全新 H3 芯片，自适应音频智能调节，个性化空间音频，USB-C 充电盒支持精确查找。"), Status: 1, Sort: 88, SalesCount: 5230, AccessCount: 31200},
			skus: []model.GoodsSKU{
				{Name: "标准版", Price: 199900, Stock: 500, Coding: "APP3-STD", Status: 1},
			},
		},
		{
			goods: model.Goods{CategoryID: 13, BrandID: 1, Title: "Apple Watch Ultra 3", Subtitle: "为极限而生", MainImage: img(4), Detail: detail("Apple Watch Ultra 3", "49mm 钛金属表壳，双频 GPS，水深仪和水温传感器，最长 72 小时续航。"), Status: 1, Sort: 85, SalesCount: 890, AccessCount: 18700},
			skus: []model.GoodsSKU{
				{Name: "49mm 钛金属 越野表带", Price: 599900, Stock: 100, Coding: "AWU3-49-TI", Status: 1},
				{Name: "49mm 钛金属 海洋表带", Price: 649900, Stock: 80, Coding: "AWU3-49-OC", Status: 1},
			},
		},
		{
			goods: model.Goods{CategoryID: 14, BrandID: 1, Title: "iPad Pro M4", Subtitle: "薄出新境界", MainImage: img(5), Detail: detail("iPad Pro M4", "搭载 M4 芯片，Ultra Retina XDR 显示屏，支持 Apple Pencil Pro 和妙控键盘。"), Status: 1, Sort: 82, SalesCount: 1560, AccessCount: 27800},
			skus: []model.GoodsSKU{
				{Name: "11英寸 256GB 深空黑", Price: 849900, Stock: 120, Coding: "IPADP-M4-11-256", Status: 1},
				{Name: "13英寸 512GB 深空黑", Price: 1299900, Stock: 60, Coding: "IPADP-M4-13-512", Status: 1},
			},
		},
		{
			goods: model.Goods{CategoryID: 11, BrandID: 3, Title: "小米15 Ultra", Subtitle: "影像旗舰，徕卡光学", MainImage: img(6), Detail: detail("小米15 Ultra", "徕卡 Summilux 光学镜头，5000 万像素四摄系统，第三代骁龙 8 移动平台。"), Status: 1, Sort: 80, SalesCount: 3210, AccessCount: 39500},
			skus: []model.GoodsSKU{
				{Name: "16GB+512GB 白色", Price: 599900, Stock: 300, Coding: "MI15U-512-W", Status: 1},
				{Name: "16GB+1TB 黑色", Price: 699900, Stock: 150, Coding: "MI15U-1T-B", Status: 1},
			},
		},
		// --- 电脑办公 ---
		{
			goods: model.Goods{CategoryID: 21, BrandID: 1, Title: "MacBook Pro 16 M4 Max", Subtitle: "为专业而生的强大实力", MainImage: img(7), Detail: detail("MacBook Pro 16 M4 Max", "M4 Max 芯片，最高 48GB 统一内存，Liquid Retina XDR 显示屏，长达 24 小时续航。"), Status: 1, Sort: 78, SalesCount: 1120, AccessCount: 24600},
			skus: []model.GoodsSKU{
				{Name: "36GB+1TB 深空黑", Price: 2799900, Stock: 50, Coding: "MBP16-M4MAX-1T", Status: 1},
				{Name: "48GB+2TB 深空黑", Price: 3499900, Stock: 30, Coding: "MBP16-M4MAX-2T", Status: 1},
			},
		},
		{
			goods: model.Goods{CategoryID: 21, BrandID: 4, Title: "联想小新Pro 16 2025", Subtitle: "轻薄高性能，创作无压力", MainImage: img(8), Detail: detail("联想小新Pro 16 2025", "AMD Ryzen 8000 系列处理器，2.5K 120Hz 高刷屏，70Wh 大电池，轻至 1.95kg。"), Status: 1, Sort: 72, SalesCount: 2340, AccessCount: 19800},
			skus: []model.GoodsSKU{
				{Name: "R7-8845H 16G+512G", Price: 499900, Stock: 200, Coding: "LN-XP16-R7", Status: 1},
				{Name: "R9-8945H 32G+1T", Price: 649900, Stock: 100, Coding: "LN-XP16-R9", Status: 1},
			},
		},
		// --- 时尚服饰 ---
		{
			goods: model.Goods{CategoryID: 31, BrandID: 5, Title: "Nike Sportswear Tech Fleece 男子连帽衫", Subtitle: "轻盈保暖，运动风尚", MainImage: img(9), Detail: detail("Nike Tech Fleece", "标志性 Tech Fleece 面料，轻盈保暖，修身剪裁，拉链口袋设计。"), Status: 1, Sort: 68, SalesCount: 4560, AccessCount: 22100},
			skus: []model.GoodsSKU{
				{Name: "M 黑色", Price: 79900, Stock: 300, Coding: "NK-TF-M-BK", Status: 1},
				{Name: "L 黑色", Price: 79900, Stock: 250, Coding: "NK-TF-L-BK", Status: 1},
				{Name: "XL 灰色", Price: 79900, Stock: 200, Coding: "NK-TF-XL-GY", Status: 1},
			},
		},
		{
			goods: model.Goods{CategoryID: 32, BrandID: 0, Title: "法式复古碎花连衣裙", Subtitle: "优雅气质，浪漫花语", MainImage: img(10), Detail: detail("法式复古碎花连衣裙", "法式方领设计，复古碎花印花，高腰收腰版型，飘逸雪纺面料。"), Status: 1, Sort: 65, SalesCount: 6780, AccessCount: 35200},
			skus: []model.GoodsSKU{
				{Name: "S 碎花蓝", Price: 29900, Stock: 400, Coding: "FR-DRESS-S", Status: 1},
				{Name: "M 碎花蓝", Price: 29900, Stock: 350, Coding: "FR-DRESS-M", Status: 1},
				{Name: "L 碎花粉", Price: 29900, Stock: 300, Coding: "FR-DRESS-L", Status: 1},
			},
		},
		{
			goods: model.Goods{CategoryID: 33, BrandID: 5, Title: "Nike Air Force 1 '07", Subtitle: "经典传承，百搭之选", MainImage: img(11), Detail: detail("Nike Air Force 1", "经典 Air Force 1 设计，Air 缓震科技，耐磨橡胶外底，百搭纯白配色。"), Status: 1, Sort: 62, SalesCount: 8920, AccessCount: 51300},
			skus: []model.GoodsSKU{
				{Name: "40 白色", Price: 89900, Stock: 200, Coding: "NK-AF1-40-W", Status: 1},
				{Name: "42 白色", Price: 89900, Stock: 180, Coding: "NK-AF1-42-W", Status: 1},
				{Name: "43 白色", Price: 89900, Stock: 150, Coding: "NK-AF1-43-W", Status: 1},
			},
		},
		// --- 名品箱包 ---
		{
			goods: model.Goods{CategoryID: 41, BrandID: 6, Title: "GUCCI GG Marmont 小号肩背包", Subtitle: "经典双G，优雅之选", MainImage: img(12), Detail: detail("GUCCI GG Marmont", "标志性双 G 金属配件，绗缝人字纹皮革，可调节链条肩带，内置多隔层。"), Status: 1, Sort: 58, SalesCount: 320, AccessCount: 18900},
			skus: []model.GoodsSKU{
				{Name: "黑色", Price: 1690000, Stock: 30, Coding: "GC-GGM-BK", Status: 1},
				{Name: "红色", Price: 1690000, Stock: 20, Coding: "GC-GGM-RD", Status: 1},
			},
		},
		{
			goods: model.Goods{CategoryID: 41, BrandID: 7, Title: "COACH Tabby 中号手提包", Subtitle: "纽约轻奢，都市风尚", MainImage: img(13), Detail: detail("COACH Tabby", "标志性 Tabby 搭扣设计，精选鹅卵石纹皮革，可拆卸肩带，多种背法。"), Status: 1, Sort: 55, SalesCount: 890, AccessCount: 12300},
			skus: []model.GoodsSKU{
				{Name: "黑色", Price: 399500, Stock: 60, Coding: "CO-TABBY-BK", Status: 1},
				{Name: "棕色", Price: 399500, Stock: 50, Coding: "CO-TABBY-BR", Status: 1},
			},
		},
		{
			goods: model.Goods{CategoryID: 43, BrandID: 0, Title: "新秀丽拉杆箱 20寸登机箱", Subtitle: "轻量耐用，商旅之选", MainImage: img(14), Detail: detail("新秀丽拉杆箱", "PC 材质轻量箱体，TSA 海关锁，万向静音飞机轮，扩展层设计。"), Status: 1, Sort: 50, SalesCount: 2100, AccessCount: 15600},
			skus: []model.GoodsSKU{
				{Name: "20寸 黑色", Price: 159900, Stock: 200, Coding: "SM-20-BK", Status: 1},
				{Name: "24寸 银色", Price: 199900, Stock: 150, Coding: "SM-24-SV", Status: 1},
			},
		},
		// --- 运动户外 ---
		{
			goods: model.Goods{CategoryID: 51, BrandID: 5, Title: "Nike Pegasus 41 男子跑步鞋", Subtitle: "飞马传奇，为跑而生", MainImage: img(15), Detail: detail("Nike Pegasus 41", "React 泡棉中底，Zoom Air 气垫，Flywire 飞线技术，透气网眼鞋面。"), Status: 1, Sort: 48, SalesCount: 5670, AccessCount: 28900},
			skus: []model.GoodsSKU{
				{Name: "41 黑白", Price: 89900, Stock: 200, Coding: "NK-PG41-41", Status: 1},
				{Name: "42 黑白", Price: 89900, Stock: 180, Coding: "NK-PG41-42", Status: 1},
				{Name: "43 蓝色", Price: 89900, Stock: 150, Coding: "NK-PG41-43", Status: 1},
			},
		},
		// --- 家用电器 ---
		{
			goods: model.Goods{CategoryID: 61, BrandID: 8, Title: "海尔新风空调 大1.5匹 一级能效", Subtitle: "新风净化，健康呼吸", MainImage: img(16), Detail: detail("海尔新风空调", "双动力新风系统，HEPA 滤网净化，一级能效变频，56°C 高温自清洁。"), Status: 1, Sort: 45, SalesCount: 1230, AccessCount: 9800},
			skus: []model.GoodsSKU{
				{Name: "大1.5匹 白色", Price: 459900, Stock: 100, Coding: "HR-AC-15P", Status: 1},
			},
		},
		{
			goods: model.Goods{CategoryID: 62, BrandID: 8, Title: "海尔冰箱 501升 十字对开门", Subtitle: "大容量，全空间保鲜", MainImage: img(17), Detail: detail("海尔冰箱 501升", "全空间保鲜科技，干湿分储，DEO 净味系统，一级能效变频压缩机。"), Status: 1, Sort: 42, SalesCount: 980, AccessCount: 8700},
			skus: []model.GoodsSKU{
				{Name: "501升 星蕴银", Price: 399900, Stock: 80, Coding: "HR-FR-501", Status: 1},
			},
		},
	}

	for _, s := range seeds {
		g := s.goods
		g.CreatedAt = now
		g.UpdatedAt = now
		global.DB.Create(&g)
		for _, sku := range s.skus {
			sku.GoodsID = g.ID
			sku.CreatedAt = now
			sku.UpdatedAt = now
			global.DB.Create(&sku)
		}
	}
	log.Println("default goods seeded:", len(seeds), "products")

	// ========== 文章分类 + 文章 ==========
	articleCats := []model.ArticleCategory{
		{ID: 1, Name: "新品资讯", Sort: 50, Status: 1},
		{ID: 2, Name: "选购指南", Sort: 40, Status: 1},
		{ID: 3, Name: "帮助中心", Sort: 30, Status: 1},
		{ID: 4, Name: "关于我们", Sort: 20, Status: 1},
	}
	global.DB.Create(&articleCats)

	articles := []model.Article{
		{CategoryID: 1, Title: "iPhone 16 Pro Max 全新发布", Cover: "/uploads/seed/article-1.jpg", Content: "<p>Apple 今日正式发布 iPhone 16 Pro Max，搭载全新 A18 Pro 芯片，配备 4800 万像素四合一超级长焦镜头，支持 5 倍光学变焦。全新钛金属设计，提供沙漠色、黑色钛金属等多种配色。</p><p>iPhone 16 Pro Max 拥有 6.9 英寸超视网膜 XDR 显示屏，峰值亮度高达 2000 尼特，支持 ProMotion 自适应刷新率技术。全天候显示功能让你随时掌握重要信息。</p>", Author: "GoShop", Sort: 50, Status: 1},
		{CategoryID: 1, Title: "HUAWEI Mate 70 系列正式亮相", Cover: "/uploads/seed/article-2.jpg", Content: "<p>华为 Mate 70 系列搭载全新麒麟芯片，运行 HarmonyOS NEXT 操作系统，带来更流畅的智慧体验。Mate 70 Pro+ 配备超感知 XMAGE 影像系统，支持十档可变光圈。</p><p>全系标配卫星通信功能，支持天通卫星通话和北斗卫星消息，让你在无网络环境下也能保持联络。</p>", Author: "GoShop", Sort: 48, Status: 1},
		{CategoryID: 1, Title: "MacBook Pro M4 Max 性能怪兽来袭", Cover: "/uploads/seed/article-3.jpg", Content: "<p>全新 MacBook Pro 搭载 M4 Max 芯片，最高可配 48GB 统一内存和 2TB 固态硬盘。CPU 性能较上代提升最高 30%，GPU 性能提升最高 40%，为专业创作者提供前所未有的强大性能。</p><p>Liquid Retina XDR 显示屏支持 1000 尼特持续亮度和 1600 尼特 HDR 峰值亮度，P3 广色域让每一个像素都栩栩如生。</p>", Author: "GoShop", Sort: 46, Status: 1},
		{CategoryID: 2, Title: "2026 年手机选购指南：旗舰机怎么选", Cover: "/uploads/seed/article-4.jpg", Content: "<p>选购旗舰手机时，建议从以下几个维度考虑：</p><p><strong>处理器性能：</strong>A18 Pro、麒麟、骁龙 8 Gen 4 是目前三大顶级移动芯片，日常使用差异不大，重度游戏用户建议关注 GPU 性能。</p><p><strong>影像系统：</strong>如果你热爱摄影，iPhone 16 Pro Max 的长焦和 Mate 70 Pro+ 的可变光圈都是不错的选择。</p><p><strong>续航与充电：</strong>5000mAh 以上电池容量已成标配，快充功率建议选择 67W 以上。</p>", Author: "GoShop", Sort: 40, Status: 1},
		{CategoryID: 2, Title: "笔记本电脑选购：创作者 vs 商务办公", Cover: "/uploads/seed/article-5.jpg", Content: "<p>创作者用户推荐 MacBook Pro M4 Max 或联想小新 Pro 16，前者适合 macOS 生态的视频剪辑和音乐制作，后者性价比更高适合 Windows 用户。</p><p>商务办公用户更看重便携性和续航，MacBook Air M3 和联想 ThinkPad X1 Carbon 都是不错的选择，重量控制在 1.3kg 以内，续航超过 10 小时。</p>", Author: "GoShop", Sort: 38, Status: 1},
		{CategoryID: 3, Title: "如何注册成为会员", Cover: "/uploads/seed/article-6.jpg", Content: "<p>点击页面右上角「账户」进入登录页面，选择「注册」标签，填写用户名和密码即可完成注册。注册成功后自动登录。</p><p>注册会员可享受：收藏商品、查看订单、领取优惠券、积分签到等专属权益。</p>", Author: "GoShop", Sort: 30, Status: 1},
		{CategoryID: 3, Title: "退换货政策说明", Cover: "/uploads/seed/article-7.jpg", Content: "<p>自签收之日起 7 天内，商品未经使用且包装完好，可申请无理由退换货。</p><p><strong>退货流程：</strong>进入「我的订单」→ 选择需要退货的订单 → 点击「申请售后」→ 选择退货原因 → 提交申请。</p><p>审核通过后，请在 7 个工作日内将商品寄回，我们收到商品并确认无误后，将在 3 个工作日内完成退款。</p>", Author: "GoShop", Sort: 28, Status: 1},
		{CategoryID: 3, Title: "支付方式与配送说明", Cover: "/uploads/seed/article-8.jpg", Content: "<p><strong>支付方式：</strong>支持微信支付、支付宝等主流支付方式。</p><p><strong>配送说明：</strong>默认使用顺丰快递，下单后 48 小时内发货（节假日顺延）。大部分地区 2-3 天可送达，偏远地区 5-7 天。</p><p>订单满 99 元包邮，不满 99 元收取 10 元运费。</p>", Author: "GoShop", Sort: 26, Status: 1},
		{CategoryID: 4, Title: "关于 GoShop", Cover: "/uploads/seed/article-9.jpg", Content: "<p>GoShop 是一个追求品质与体验的电商平台。我们精选全球优质商品，致力于为用户提供简洁、高效、愉悦的购物体验。</p><p>我们相信，好的产品值得被更多人发现。无论是前沿科技产品，还是经典时尚单品，GoShop 都以严格的品质标准为你把关。</p>", Author: "GoShop", Sort: 20, Status: 1},
		{CategoryID: 4, Title: "联系我们", Cover: "/uploads/seed/article-10.jpg", Content: "<p>如有任何问题，欢迎通过以下方式联系我们：</p><p>客服邮箱：hi@zhangpanda.com</p><p>工作时间：周一至周五 9:00-18:00</p><p>您也可以访问「支持」页面查看常见问题解答。</p>", Author: "GoShop", Sort: 18, Status: 1},
	}
	global.DB.Create(&articles)
	log.Println("default articles seeded:", len(articles), "articles")

	// ========== 优惠券 ==========
	coupons := []model.Coupon{
		{Name: "新人专享 ¥50 优惠券", Type: 1, MinAmount: 29900, Value: 5000, Total: 10000, PerLimit: 1, StartTime: now, EndTime: now.AddDate(0, 3, 0), Status: 1},
		{Name: "满 500 减 80", Type: 1, MinAmount: 50000, Value: 8000, Total: 5000, PerLimit: 2, StartTime: now, EndTime: now.AddDate(0, 1, 0), Status: 1},
		{Name: "数码专区 9 折券", Type: 2, MinAmount: 100000, Value: 90, Total: 3000, PerLimit: 1, StartTime: now, EndTime: now.AddDate(0, 2, 0), Status: 1},
		{Name: "无门槛 ¥10 优惠券", Type: 3, MinAmount: 0, Value: 1000, Total: 20000, PerLimit: 3, StartTime: now, EndTime: now.AddDate(0, 6, 0), Status: 1},
	}
	global.DB.Create(&coupons)
	log.Println("default coupons seeded:", len(coupons), "coupons")

	// ========== 筛选价格区间 ==========
	prices := []model.ScreeningPrice{
		{ID: 1, Name: "0-299", MinPrice: 0, MaxPrice: 29900, Sort: 50},
		{ID: 2, Name: "300-999", MinPrice: 30000, MaxPrice: 99900, Sort: 40},
		{ID: 3, Name: "1000-4999", MinPrice: 100000, MaxPrice: 499900, Sort: 30},
		{ID: 4, Name: "5000-9999", MinPrice: 500000, MaxPrice: 999900, Sort: 20},
		{ID: 5, Name: "10000以上", MinPrice: 1000000, MaxPrice: 99999900, Sort: 10},
	}
	global.DB.Create(&prices)
	log.Println("default screening prices seeded")

	// ========== 轮播图 ==========
	slides := []model.Slide{
		{Name: "探索全新 iPhone", Images: `["/uploads/seed/slide-1.jpg"]`, Sort: 30, Status: 1},
		{Name: "MacBook Pro", Images: `["/uploads/seed/slide-2.jpg"]`, Sort: 20, Status: 1},
		{Name: "春季新品上市", Images: `["/uploads/seed/slide-3.jpg"]`, Sort: 10, Status: 1},
	}
	global.DB.Create(&slides)
	log.Println("default slides seeded:", len(slides), "slides")
}

// EnsureDefaultPayments 在无支付方式时补全最小配置（兼容层 order/pay 与列表展示回退依赖）
func EnsureDefaultPayments() {
	var c int64
	global.DB.Model(&model.Payment{}).Count(&c)
	if c > 0 {
		return
	}
	// payment_key 须与 internal/service/payment_driver.go 中 GetPaymentDriver 注册名一致
	payments := []model.Payment{
		{Name: "线下支付", Logo: "", Config: `{"payment_key":"offline"}`, Sort: 100, Status: 1},
		{Name: "钱包余额", Logo: "", Config: `{"payment_key":"wallet"}`, Sort: 99, Status: 1},
		{Name: "微信JSAPI", Logo: "", Config: `{"payment_key":"wechat_jsapi"}`, Sort: 96, Status: 1},
		{Name: "微信H5", Logo: "", Config: `{"payment_key":"wechat_h5"}`, Sort: 95, Status: 1},
		{Name: "微信APP", Logo: "", Config: `{"payment_key":"wechat_app"}`, Sort: 94, Status: 1},
		{Name: "微信扫码", Logo: "", Config: `{"payment_key":"wechat_native"}`, Sort: 93, Status: 1},
		{Name: "支付宝手机网站", Logo: "", Config: `{"payment_key":"alipay_h5"}`, Sort: 89, Status: 1},
		{Name: "支付宝电脑网站", Logo: "", Config: `{"payment_key":"alipay_pc"}`, Sort: 88, Status: 1},
		{Name: "支付宝APP", Logo: "", Config: `{"payment_key":"alipay_app"}`, Sort: 87, Status: 1},
		{Name: "支付宝小程序", Logo: "", Config: `{"payment_key":"alipay_mini"}`, Sort: 86, Status: 1},
		{Name: "当面付", Logo: "", Config: `{"payment_key":"alipay_face"}`, Sort: 85, Status: 1},
		{Name: "PayPal", Logo: "", Config: `{"payment_key":"paypal"}`, Sort: 70, Status: 1},
	}
	global.DB.Create(&payments)
	log.Println("default payments ensured:", len(payments))
}
