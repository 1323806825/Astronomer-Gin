package middleware

import (
	"astronomer-gin/model"
	"astronomer-gin/pkg/constant"
	"astronomer-gin/pkg/database"
	"astronomer-gin/pkg/util"

	"github.com/gin-gonic/gin"
)

// AdminMiddleware 管理员权限验证中间件
// 必须在AuthMiddleware之后使用
func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取用户ID（由AuthMiddleware设置）
		phone, exists := c.Get("phone")
		if !exists {
			util.Forbidden(c, constant.ErrPermissionDenied.Message)
			c.Abort()
			return
		}

		// 查询用户角色
		var user model.User
		db := database.GetDB()
		if err := db.Where("phone = ?", phone).Select("role").First(&user).Error; err != nil {
			util.Forbidden(c, constant.ErrPermissionDenied.Message)
			c.Abort()
			return
		}

		// 验证是否为管理员或超级管理员
		if user.Role != "admin" && user.Role != "super_admin" {
			util.Forbidden(c, constant.ErrPermissionDenied.Message)
			c.Abort()
			return
		}

		// 设置管理员标识
		c.Set("is_admin", true)
		c.Set("admin_role", user.Role)

		c.Next()
	}
}

// SuperAdminMiddleware 超级管理员权限验证中间件
// 必须在AuthMiddleware之后使用
func SuperAdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取用户ID
		phone, exists := c.Get("phone")
		if !exists {
			util.Forbidden(c, constant.PermissionDenied)
			c.Abort()
			return
		}

		// 查询用户角色
		var user model.User
		db := database.GetDB()
		if err := db.Where("phone = ?", phone).Select("role").First(&user).Error; err != nil {
			util.Forbidden(c, constant.PermissionDenied)
			c.Abort()
			return
		}

		// 验证是否为超级管理员
		if user.Role != "super_admin" {
			util.Forbidden(c, constant.PermissionDenied)
			c.Abort()
			return
		}

		// 设置超级管理员标识
		c.Set("is_super_admin", true)

		c.Next()
	}
}
