package service

import (
	"bytes"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"os"
	"strings"
	"time"
	"unicode"

	"golang.org/x/image/math/fixed"

	"github.com/golang/freetype"
	"github.com/jung-kurt/gofpdf"
	"github.com/skip2/go-qrcode"
	"gorm.io/gorm"

	"movingManager/database"
	"movingManager/model"
)

// TagResponse 标签响应结构
type TagResponse struct {
	MoveUid            string `json:"move_uid"`
	TagUid             string `json:"tag_uid"`
	TagName            string `json:"tag_name"`
	Remark             string `json:"remark"`
	IsVerified         int    `json:"is_verified"`
	Status             int    `json:"status"` // 标签状态
	IsDeleted          int    `json:"is_deleted"`
	DeletedAt          int64  `json:"deleted_at,omitempty"`
	CreatedAt          string `json:"created_at"`
	UpdatedAt          string `json:"updated_at,omitempty"`
	StartLocation      string `json:"start_location"`
	EndLocation        string `json:"end_location"`
	MoveTime           string `json:"move_at"`
	MoveRemark         string `json:"move_remark"`
	TagCount           int    `json:"tag_count"`
	VerifiedTagCount   int    `json:"verified_tag_count"`
	UnverifiedTagCount int    `json:"unverified_tag_count"`
	IsCompleted        int    `json:"is_completed"`
}

// GenerateTagPDF 生成标签PDF业务处理
func GenerateTagPDF(userUid, moveUid string) ([]byte, error) {
	// 获取该搬运下的所有未删除标签
	tagModel := model.TagModel{}
	tags, err := tagModel.GetTagsByMove(userUid, moveUid, true)
	if err != nil {
		return nil, fmt.Errorf("查询标签失败: %v", err)
	}

	if len(tags) == 0 {
		return nil, fmt.Errorf("该搬运下没有标签")
	}

	// 创建PDF
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	// 添加中文字体支持
	pdf.AddUTF8Font("Alibaba", "", "fonts/AlibabaPuHuiTi-3-95-ExtraBold.ttf")
	pdf.SetFont("Alibaba", "", 11)

	// 初始化容器布局参数（前端式容器布局）
	var (
		pageWidth       = 210.0 // A4页面宽度(mm)
		margin          = 5.0   // 页面左右边距
		columns         = 4     // 每行固定4列
		padding         = 1.0   // 容器内边距
		verticalSpacing = 2.0   // 行垂直间距
		pageHeight      = 297.0 // A4页面高度(mm)
		yStart          = 2.0   // 页面顶部起始Y坐标
	)
	containerWidth := (pageWidth - margin*2) / float64(columns) // 自动计算容器宽度（均分4列）
	currentX := margin                                          // 当前X坐标
	currentY := yStart                                          // 当前Y坐标
	currentRowHeight := 0.0                                     // 当前行高度（动态计算）

	i := 0

	// 容器布局主循环
	for _, tag := range tags {
		// 计算标签内容高度（文本高度+二维码高度+内边距）
		textHeight := calculateTextHeight(tag.TagName, containerWidth-padding*2, 11) // 动态计算文本高度
		qrSize := containerWidth - padding*2 - 10                                    // 二维码宽度=容器宽度-内边距-额外间距
		contentHeight := textHeight + qrSize + padding*3                             // 总内容高度（增加额外内边距）

		// 检查是否需要换行
		if currentX+containerWidth > pageWidth-margin {
			// 换行逻辑：更新Y坐标并重置行宽度
			currentY += currentRowHeight + verticalSpacing
			currentX = margin
			currentRowHeight = 0
		}

		// 检查是否需要分页
		if currentY+contentHeight > pageHeight-margin*2 {
			pdf.AddPage()
			currentY = yStart
			currentX = margin
			currentRowHeight = 0
		}

		// 绘制容器边框
		pdf.Rect(currentX, currentY, containerWidth, contentHeight, "D")

		// 绘制标签文本（自适应字体）
		pdf.SetXY(currentX+padding, currentY+padding)
		adjustFontSize(pdf, tag.TagName, containerWidth-padding*2, 11, 8)
		pdf.MultiCell(containerWidth-padding*2, 5, fmt.Sprintf("标签 %d\n%s", i+1, tag.TagName), "", "CM", false)
		pdf.SetFontSize(11) // 恢复默认字体

		// 绘制二维码（容器内底部居中）
		qrImgPath, err := generateQRCodeImage(fmt.Sprintf("http://192.168.2.17:5173/tag/%s", tag.TagUid))
		if err != nil {
			return nil, fmt.Errorf("生成二维码失败: %v", err)
		}
		defer os.Remove(qrImgPath)
		pdf.ImageOptions(qrImgPath,
			currentX+padding+(containerWidth-padding*2-qrSize)/2, // 水平居中
			currentY+contentHeight-padding-qrSize,                // 垂直底部对齐
			qrSize, qrSize, true, gofpdf.ImageOptions{ImageType: "png", ReadDpi: true}, 0, "")

		// 更新当前行状态
		currentX += containerWidth
		if contentHeight > currentRowHeight {
			currentRowHeight = contentHeight // 记录当前行最大高度
		}
		i++
	}

	// 将PDF输出到字节缓冲区
	var buf bytes.Buffer
	if err = pdf.Output(&buf); err != nil {
		return nil, fmt.Errorf("生成PDF失败: %v", err)
	}

	// 验证PDF头部和尾部
	pdfBytes := buf.Bytes()
	if len(pdfBytes) < 10 || string(pdfBytes[:5]) != "%PDF-" {
		return nil, fmt.Errorf("生成的PDF文件格式无效（头部错误）")
	}
	// 放宽PDF文件尾验证
	if !bytes.Contains(pdfBytes[len(pdfBytes)-100:], []byte("%%EOF")) {
		return nil, fmt.Errorf("生成的PDF文件格式无效（尾部缺失）")
	}

	// 生成二维码图片
	qrImgPath, err := generateQRCodeImage("https://example.com/tag/verify")
	if err != nil {
		return nil, fmt.Errorf("生成二维码失败: %v", err)
	}
	defer os.Remove(qrImgPath)

	// 在新页面添加二维码（A4页面尺寸210x297mm）
	pdf.AddPage()
	pdf.ImageOptions(qrImgPath, (210-40)/2, 100, 40, 0, true, gofpdf.ImageOptions{ImageType: "png", ReadDpi: true}, 0, "")

	return buf.Bytes(), nil
}

// CreateTag 创建标签业务处理
// CreateTagRequest 创建标签请求参数
type CreateTagRequest struct {
	MoveUid string `json:"move_uid"` // 搬运UID
	TagName string `json:"tag_name"` // 标签名称
	Remark  string `json:"remark"`   // 标签备注
}

func CreateTag(userUid string, req CreateTagRequest) (*TagResponse, error) {
	var response *TagResponse
	err := database.DB.Transaction(func(tx *gorm.DB) error {
		// 验证搬运记录是否存在且属于当前用户
		tagModel := model.TagModel{}
		move, err := tagModel.GetMoveByUserAndMoveUidTx(tx, userUid, req.MoveUid, true)
		if err != nil {
			return err
		}

		// 创建标签
		tag := model.TagModel{
			UserUid:    userUid,
			MoveUid:    req.MoveUid,
			TagName:    req.TagName,
			Remark:     req.Remark,
			IsVerified: 0, // 默认未核销

		}
		if err := tag.CreateTx(tx); err != nil {
			return fmt.Errorf("创建标签失败: %v", err)
		}

		// 更新搬运记录的标签统计
		if err := tagModel.UpdateMoveTagCountTx(tx, move, 1, 0, 1); err != nil {
			return fmt.Errorf("更新搬运标签统计失败: %v", err)
		}

		// 转换为响应格式
		response = convertTagToResponse(&tag)
		return nil
	})
	return response, err
}

// UpdateTag 更新标签业务处理
// UpdateTagRequest 编辑标签请求参数
type UpdateTagRequest struct {
	TagUid     string `json:"tag_uid"`     // 标签UID
	TagName    string `json:"tag_name"`    // 标签名称
	Remark     string `json:"remark"`      // 标签备注
	IsVerified int    `json:"is_verified"` // 是否核销(0-未核销,1-已核销)
}

// GetTagDetail 获取标签详情业务处理
func GetTagDetail(userUid, tagUid string) (*TagResponse, error) {
	var tag model.TagModel
	if err := tag.GetByUID(userUid, tagUid, true); err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("用户无此标签记录")
		}
		return nil, fmt.Errorf("查询标签记录失败: %v", err)
	}

	// 查询关联的搬运记录
	var move model.MoveModel
	if err := move.GetByUID(userUid, tag.MoveUid, true); err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("关联的搬运记录不存在")
		}
		return nil, fmt.Errorf("查询搬运记录失败: %v", err)
	}

	// 转换为响应格式
	return &TagResponse{
		MoveUid:            tag.MoveUid,
		TagUid:             tag.TagUid,
		TagName:            tag.TagName,
		Remark:             tag.Remark,
		IsVerified:         tag.IsVerified,
		IsDeleted:          tag.IsDeleted,
		DeletedAt:          tag.DeletedAt,
		CreatedAt:          time.Unix(tag.CreatedAt, 0).Format("2006-01-02 15:04:05"),
		UpdatedAt:          time.Unix(tag.UpdatedAt, 0).Format("2006-01-02 15:04:05"),
		StartLocation:      move.StartLocation,
		EndLocation:        move.EndLocation,
		MoveTime:           time.Unix(move.MoveAt, 0).Format("2006-01-02 15:04:05"),
		MoveRemark:         move.Remark,
		TagCount:           move.TagCount,
		VerifiedTagCount:   move.VerifiedTagCount,
		UnverifiedTagCount: move.UnverifiedTagCount,
		IsCompleted:        move.IsCompleted,
	}, nil
}

func UpdateTag(userUid string, req UpdateTagRequest) (*TagResponse, error) {
	var response *TagResponse
	err := database.DB.Transaction(func(tx *gorm.DB) error {
		// 查询标签并验证所有权
		var tag model.TagModel
		if err := tag.GetByUserAndTagUidTx(tx, userUid, req.TagUid, true); err != nil {
			return err
		}

		// 记录原始核销状态
		oldIsVerified := tag.IsVerified

		// 更新标签信息
		tag.TagName = req.TagName
		tag.Remark = req.Remark

		// 如果核销状态变更，需要更新搬运记录的统计
		if oldIsVerified != req.IsVerified {
			tag.IsVerified = req.IsVerified

			// 查询关联的搬运记录
			tagModel := model.TagModel{}
			move, err := tagModel.GetMoveByMoveUidTx(tx, tag.MoveUid, true)
			if err != nil {
				return fmt.Errorf("查询搬运记录失败: %v", err)
			}

			// 使用原子操作更新搬运标签统计（确保非负）
			var deltaVerified, deltaUnverified int
			if req.IsVerified == 1 {
				// 从未核销变为已核销：已核销+1，未核销-1
				deltaVerified = 1
				deltaUnverified = -1
			} else {
				// 从已核销变为未核销：已核销-1，未核销+1
				deltaVerified = -1
				deltaUnverified = 1
			}

			// 执行原子更新并自动计算完成状态
			if err := tx.Model(move).Updates(map[string]interface{}{
				"verified_tag_count":   gorm.Expr("GREATEST(verified_tag_count + ?, 0)", deltaVerified),
				"unverified_tag_count": gorm.Expr("GREATEST(unverified_tag_count + ?, 0)", deltaUnverified),
				"is_completed":         gorm.Expr("CASE WHEN unverified_tag_count + ? = 0 THEN 1 ELSE 0 END", deltaUnverified),
			}).Error; err != nil {
				return fmt.Errorf("更新搬运标签统计失败: %v", err)
			}
		}

		if err := tag.UpdateTx(tx); err != nil {
			return fmt.Errorf("更新标签失败: %v", err)
		}

		// 转换为响应格式
		response = convertTagToResponse(&tag)
		return nil
	})
	return response, err
}

// DeleteTag 删除标签业务处理
func DeleteTag(userUid, tagUid string, isDeleted int) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		// 恢复操作时需要查找已删除记录（isDeleted=0表示恢复）
		onlyUndeleted := isDeleted != 0
		// 查询标签并验证所有权
		var tag model.TagModel
		if err := tag.GetByUserAndTagUidTx(tx, userUid, tagUid, onlyUndeleted); err != nil {
			return err
		}

		// 如果是删除操作(不是恢复)
		if isDeleted == 1 {
			// 查询关联的搬运记录
			tagModel := model.TagModel{}
			move, err := tagModel.GetMoveByMoveUidTx(tx, tag.MoveUid, true)
			if err != nil {
				return fmt.Errorf("查询搬运记录失败: %v", err)
			}

			// 更新搬运记录的标签统计
			var verifiedDelta, unverifiedDelta int
			if tag.IsVerified == 1 {
				verifiedDelta = -1
			} else {
				unverifiedDelta = -1
			}
			if err := tagModel.UpdateMoveTagCountTx(tx, move, -1, verifiedDelta, unverifiedDelta); err != nil {
				return fmt.Errorf("更新搬运标签统计失败: %v", err)
			}
		}

		// 更新标签删除状态
		// 恢复标签时增加总数和未核销数
		if isDeleted == 0 {
			// 查询关联的搬运记录
			tagModel := model.TagModel{}
			move, err := tagModel.GetMoveByMoveUidTx(tx, tag.MoveUid, true)
			if err != nil {
				return fmt.Errorf("查询搬运记录失败: %v", err)
			}

			// 根据标签原核销状态更新对应计数
			var unverifiedDelta, verifiedDelta int
			if tag.IsVerified == 1 {
				verifiedDelta = 1
			} else {
				unverifiedDelta = 1
			}

			// 更新搬运记录的标签统计（原子操作确保非负）
			if err := tx.Model(move).Updates(map[string]interface{}{
				"tag_count":            gorm.Expr("GREATEST(tag_count + 1, 0)"),
				"unverified_tag_count": gorm.Expr("GREATEST(unverified_tag_count + ?, 0)", unverifiedDelta),
				"verified_tag_count":   gorm.Expr("GREATEST(verified_tag_count + ?, 0)", verifiedDelta),
				"is_completed":         0, // 恢复后可能有未核销标签，设置为未完成
			}).Error; err != nil {
				return fmt.Errorf("更新搬运标签统计失败: %v", err)
			}
		}

		return tag.UpdateDeleteStatusTx(tx, isDeleted)
	})
}

// VerifyTag 核销标签业务处理
func VerifyTag(userUid, tagUid string, isVerified int) error {
	db := database.DB

	// 开启事务
	return db.Transaction(func(tx *gorm.DB) error {
		// 查询标签并验证所有权
		var tag model.TagModel
		if err := tag.GetByUserAndTagUidTx(tx, userUid, tagUid, true); err != nil {
			if err == gorm.ErrRecordNotFound {
				return fmt.Errorf("用户无此标签记录")
			}
			return fmt.Errorf("查询标签失败: %v", err)
		}

		// 如果状态没变，直接返回
		if tag.IsVerified == isVerified {
			return nil
		}

		// 查询关联的搬运记录
		tagModel := model.TagModel{}
		move, err := tagModel.GetMoveByMoveUidTx(tx, tag.MoveUid, true)
		if err != nil {
			return fmt.Errorf("查询搬运记录失败: %v", err)
		}

		// 计算标签统计变化量
		var verifiedDelta, unverifiedDelta int
		if isVerified == 1 {
			verifiedDelta = 1
			unverifiedDelta = -1
		} else {
			verifiedDelta = -1
			unverifiedDelta = 1
		}

		// 更新搬运记录的标签统计
		// 核销标签时更新计数并检查完成状态
		if err := tx.Model(&move).Updates(map[string]interface{}{
			"verified_tag_count":   gorm.Expr("GREATEST(verified_tag_count + ?, 0)", verifiedDelta),
			"unverified_tag_count": gorm.Expr("GREATEST(unverified_tag_count + ?, 0)", unverifiedDelta),
			"is_completed":         gorm.Expr("CASE WHEN unverified_tag_count + ? = 0 THEN 1 ELSE is_completed END", unverifiedDelta),
		}).Error; err != nil {
			return fmt.Errorf("更新搬运标签统计失败: %v", err)
		}

		// 更新标签核销状态
		tag.IsVerified = isVerified
		if err := tag.UpdateTx(tx); err != nil {
			return fmt.Errorf("更新标签核销状态失败: %v", err)
		}

		return nil
	})
}

// GetTagList 获取标签列表业务处理
func GetTagList(userUid, moveUid string, page, pageSize int) ([]TagResponse, int64, error) {
	// 验证搬运记录是否存在且属于当前用户
	var move model.MoveModel
	if err := move.GetByUID(userUid, moveUid, true); err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, 0, fmt.Errorf("搬运记录不存在")
		}
		return nil, 0, fmt.Errorf("查询搬运记录失败: %v", err)
	}

	// 调用model层查询方法
	var tag model.TagModel
	tags, total, err := tag.ListByMove(userUid, moveUid, page, pageSize, true)
	if err != nil {
		return nil, 0, fmt.Errorf("查询标签列表失败: %v", err)
	}

	// 转换为响应格式
	var responses []TagResponse
	for _, tag := range tags {
		responses = append(responses, *convertTagToResponse(&tag))
	}

	return responses, total, nil
}

// calculateTextHeight 计算文本高度
func calculateTextHeight(text string, width float64, fontSize float64) float64 {
	// 使用freetype测量文本高度
	fontBytes, err := os.ReadFile("fonts/AlibabaPuHuiTi-3-95-ExtraBold.ttf")
	if err != nil {
		return 0
	}
	font, err := freetype.ParseFont(fontBytes)
	if err != nil {
		return 0
	}

	ctx := freetype.NewContext()
	ctx.SetFont(font)
	ctx.SetFontSize(fontSize)

	// 计算文本宽度和高度
	words := strings.Fields(text)
	currentWidth := 0.0
	lines := 1
	lineHeight := fontSize * 1.2 // 行高为字号的1.2倍

	for _, word := range words {
		wordWidth := 0.0
		for _, r := range word + " " {
			index := font.Index(r)
			hMetric := font.HMetric(fixed.Int26_6(fontSize*64), index) // 添加字体大小参数
			advance := hMetric.AdvanceWidth                            // 获取字符前进宽度
			wordWidth += float64(advance) / 64 * 0.3527                // 26.6定点转毫米
		}
		if currentWidth+wordWidth > width {
			lines++
			currentWidth = wordWidth
		} else {
			currentWidth += wordWidth
		}
	}

	return float64(lines) * lineHeight
}

// adjustFontSize 调整字体大小以适应宽度
func adjustFontSize(pdf *gofpdf.Fpdf, text string, maxWidth float64, maxSize, minSize float64) {
	// 保存原始字体大小

	found := false

	// 从最大字体大小开始尝试
	for size := maxSize; size >= minSize; size -= 0.5 {
		pdf.SetFontSize(size)
		width := pdf.GetStringWidth(text)
		if width <= maxWidth {
			found = true
			break
		}
	}

	// 如果没有找到合适的大小，使用最小字体大小
	if !found {
		pdf.SetFontSize(minSize)
	}
}

// generateTextImage 生成中文文本图片
// generateQRCodeImage 生成二维码图片文件路径
func generateQRCodeImage(url string) (string, error) {
	qrCode, err := qrcode.Encode(url, qrcode.Medium, 186)
	if err != nil {
		return "", err
	}

	// 创建临时文件
	tmpFile, err := os.CreateTemp("", "qrcode-*.png")
	if err != nil {
		return "", err
	}
	defer tmpFile.Close()

	if _, err := tmpFile.Write(qrCode); err != nil {
		return "", err
	}
	return tmpFile.Name(), nil
}

// wrapText 自动换行函数
// wrapText 智能换行（按汉字字符数）
// wrapText 智能换行（中文字符按2单位计算）
func wrapText(text string, maxLen int) string {
	count := 0
	// 由于缺少 strings 包导入，需先添加对应的导入语句。当前仅重写选择部分，需在文件开头导入 "strings" 包
	// 此注释仅为说明，实际在文件开头添加导入语句即可
	// 移除 textHeight 定义
	var result strings.Builder

	for _, r := range text {
		if count >= maxLen {
			result.WriteString("\n")
			count = 0
		}

		if unicode.Is(unicode.Han, r) {
			count += 2
		} else {
			count += 1
		}

		result.WriteRune(r)
	}

	return result.String()
}

func generateTextImage(text string, fontSize float64) ([]byte, error) {
	img := image.NewRGBA(image.Rect(0, 0, 400, 30)) // 预设图片尺寸
	draw.Draw(img, img.Bounds(), image.White, image.Point{}, draw.Src)

	// 使用项目中已有的中文字体文件
	fontBytes, err := os.ReadFile("fonts/AlibabaPuHuiTi-3-95-ExtraBold.ttf")
	if err != nil {
		return nil, fmt.Errorf("读取内置字体失败: %v", err)
	}
	font, err := freetype.ParseFont(fontBytes)
	if err != nil {
		return nil, fmt.Errorf("解析字体失败: %v", err)
	}

	c := freetype.NewContext()
	c.SetDPI(72)
	c.SetFont(font)
	c.SetFontSize(float64(fontSize))
	c.SetClip(img.Bounds())
	c.SetDst(img)
	c.SetSrc(image.Black)

	pt := freetype.Pt(10, 20) // 文本位置
	if _, err := c.DrawString(text, pt); err != nil {
		return nil, fmt.Errorf("绘制文本失败: %v", err)
	}

	// 编码为PNG
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, fmt.Errorf("编码图片失败: %v", err)
	}
	return buf.Bytes(), nil
}

// generateQRCode 生成二维码图片
func generateQRCode(data string, size int) ([]byte, error) {
	var buf bytes.Buffer
	qr, err := qrcode.New(data, qrcode.Medium) // 使用中等纠错等级
	if err != nil {
		return nil, err
	}
	if err := qr.Write(size, &buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// convertTagToResponse 将模型转换为响应格式
func convertTagToResponse(tag *model.TagModel) *TagResponse {
	updatedAt := ""
	if tag.UpdatedAt > 0 {
		updatedAt = time.Unix(tag.UpdatedAt, 0).Format("2006-01-02 15:04:05")
	}

	deleteTime := int64(0)
	if tag.IsDeleted == 1 {
		deleteTime = tag.DeletedAt
	}

	return &TagResponse{
		MoveUid:    tag.MoveUid,
		TagUid:     tag.TagUid,
		TagName:    tag.TagName,
		Remark:     tag.Remark,
		IsVerified: tag.IsVerified,
		IsDeleted:  tag.IsDeleted,
		DeletedAt:  deleteTime,
		CreatedAt:  time.Unix(tag.CreatedAt, 0).Format("2006-01-02 15:04:05"),
		UpdatedAt:  updatedAt,
	}
}
