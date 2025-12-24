package upload

import (
	"astronomer-gin/pkg/constant"
	"astronomer-gin/pkg/util"
	"astronomer-gin/service"

	"github.com/gin-gonic/gin"
)

// UploadHandler 文件上传处理器
type UploadHandler struct {
	uploadService service.UploadServiceV2
}

// NewUploadHandler 创建上传Handler实例
func NewUploadHandler(uploadService service.UploadServiceV2) *UploadHandler {
	return &UploadHandler{
		uploadService: uploadService,
	}
}

// UploadImage 上传图片
// @Summary 上传图片
// @Description 上传图片文件（支持jpg, jpeg, png, gif, webp），最大5MB
// @Tags 文件上传
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "图片文件"
// @Success 200 {object} util.Response{data=map[string]string} "成功返回图片URL"
// @Failure 400 {object} util.Response "参数错误"
// @Failure 500 {object} util.Response "服务器错误"
// @Router /api/v1/upload/image [post]
func (h *UploadHandler) UploadImage(c *gin.Context) {
	// 获取上传的文件
	file, err := c.FormFile("file")
	if err != nil {
		util.Error(c, constant.ErrParamInvalid.Code, "请选择要上传的图片")
		return
	}

	// 调用服务层上传图片
	fileURL, err := h.uploadService.UploadImage(c.Request.Context(), file)
	if err != nil {
		util.Error(c, constant.ErrFileUploadFailed.Code, err.Error())
		return
	}

	// 返回成功响应
	util.SuccessWithMessage(c, "图片上传成功", gin.H{
		"file_url":  fileURL,
		"file_name": file.Filename,
		"file_size": file.Size,
	})
}

// UploadImageWithVersions 上传图片并生成多个版本
// @Summary 上传图片（多版本）
// @Description 上传图片并自动生成缩略图、中图、大图等多个版本
// @Tags 文件上传
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "图片文件"
// @Success 200 {object} util.Response{data=service.ImageVersions} "成功返回多版本图片URL"
// @Failure 400 {object} util.Response "参数错误"
// @Failure 500 {object} util.Response "服务器错误"
// @Router /api/v1/upload/image-versions [post]
func (h *UploadHandler) UploadImageWithVersions(c *gin.Context) {
	// 获取上传的文件
	file, err := c.FormFile("file")
	if err != nil {
		util.Error(c, constant.ErrParamInvalid.Code, "请选择要上传的图片")
		return
	}

	// 调用服务层上传图片（多版本）
	versions, err := h.uploadService.UploadImageWithVersions(c.Request.Context(), file)
	if err != nil {
		util.Error(c, constant.ErrFileUploadFailed.Code, err.Error())
		return
	}

	// 返回成功响应
	util.SuccessWithMessage(c, "图片上传成功", gin.H{
		"file_name": file.Filename,
		"file_size": file.Size,
		"versions":  versions,
	})
}

// UploadFile 上传文件
// @Summary 上传文件
// @Description 上传任意类型文件，最大20MB
// @Tags 文件上传
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "文件"
// @Param category formData string false "文件分类" default(files)
// @Success 200 {object} util.Response{data=map[string]string} "成功返回文件URL"
// @Failure 400 {object} util.Response "参数错误"
// @Failure 500 {object} util.Response "服务器错误"
// @Router /api/v1/upload/file [post]
func (h *UploadHandler) UploadFile(c *gin.Context) {
	// 获取上传的文件
	file, err := c.FormFile("file")
	if err != nil {
		util.Error(c, constant.ErrParamInvalid.Code, "请选择要上传的文件")
		return
	}

	// 获取文件分类（可选）
	category := c.DefaultPostForm("category", "files")

	// 调用服务层上传文件
	fileURL, err := h.uploadService.UploadFile(c.Request.Context(), file, category)
	if err != nil {
		util.Error(c, constant.ErrFileUploadFailed.Code, err.Error())
		return
	}

	// 返回成功响应
	util.SuccessWithMessage(c, "文件上传成功", gin.H{
		"file_url":  fileURL,
		"file_name": file.Filename,
		"file_size": file.Size,
		"category":  category,
	})
}

// UploadMultiple 批量上传文件
// @Summary 批量上传文件
// @Description 一次上传多个文件，每个文件最大20MB
// @Tags 文件上传
// @Accept multipart/form-data
// @Produce json
// @Param files formData file true "文件列表"
// @Param category formData string false "文件分类" default(files)
// @Success 200 {object} util.Response{data=[]map[string]string} "成功返回文件URL列表"
// @Failure 400 {object} util.Response "参数错误"
// @Failure 500 {object} util.Response "服务器错误"
// @Router /api/v1/upload/multiple [post]
func (h *UploadHandler) UploadMultiple(c *gin.Context) {
	// 获取表单
	form, err := c.MultipartForm()
	if err != nil {
		util.Error(c, constant.ErrParamInvalid.Code, "获取文件失败")
		return
	}

	// 获取所有文件
	files := form.File["files"]
	if len(files) == 0 {
		util.Error(c, constant.ErrParamInvalid.Code, "请选择要上传的文件")
		return
	}

	// 限制最多上传10个文件
	if len(files) > 10 {
		util.Error(c, constant.ErrParamInvalid.Code, "最多只能上传10个文件")
		return
	}

	// 获取文件分类
	category := c.DefaultPostForm("category", "files")

	// 上传所有文件
	var results []gin.H
	var failedFiles []string

	for _, file := range files {
		fileURL, err := h.uploadService.UploadFile(c.Request.Context(), file, category)
		if err != nil {
			failedFiles = append(failedFiles, file.Filename)
			continue
		}

		results = append(results, gin.H{
			"file_url":  fileURL,
			"file_name": file.Filename,
			"file_size": file.Size,
		})
	}

	// 返回结果
	if len(failedFiles) > 0 {
		util.SuccessWithMessage(c, "部分文件上传失败", gin.H{
			"success_count": len(results),
			"failed_count":  len(failedFiles),
			"failed_files":  failedFiles,
			"files":         results,
		})
	} else {
		util.SuccessWithMessage(c, "所有文件上传成功", gin.H{
			"success_count": len(results),
			"files":         results,
		})
	}
}

// DeleteFile 删除文件
// @Summary 删除文件
// @Description 根据文件URL删除文件
// @Tags 文件上传
// @Accept json
// @Produce json
// @Param request body map[string]string true "请求体"
// @Success 200 {object} util.Response "删除成功"
// @Failure 400 {object} util.Response "参数错误"
// @Failure 500 {object} util.Response "服务器错误"
// @Router /api/v1/upload/delete [delete]
func (h *UploadHandler) DeleteFile(c *gin.Context) {
	// 解析请求参数
	var req struct {
		FileURL string `json:"file_url" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		util.Error(c, constant.ErrParamInvalid.Code, "请提供有效的文件URL")
		return
	}

	// 调用服务层删除文件
	err := h.uploadService.DeleteFile(c.Request.Context(), req.FileURL)
	if err != nil {
		util.Error(c, constant.ErrFileDeleteFailed.Code, err.Error())
		return
	}

	// 返回成功响应
	util.SuccessWithMessage(c, "文件删除成功", nil)
}
