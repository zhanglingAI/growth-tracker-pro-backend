package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/growth-tracker-pro-backend/internal/models"
	"github.com/growth-tracker-pro-backend/internal/service"
)

// Handler 处理程序
type Handler struct {
	service service.Service
}

// NewHandler 创建处理程序
func NewHandler(svc service.Service) *Handler {
	return &Handler{service: svc}
}

// RegisterRoutes 注册路由
func (h *Handler) RegisterRoutes(r *gin.Engine) {
	// 中间件
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(corsMiddleware())

	// 健康检查
	r.GET("/health", h.HealthCheck)

	// API路由组
	api := r.Group("/api/v1")
	{
		// 认证 (不需要登录)
		auth := api.Group("/auth")
		{
			auth.POST("/login", h.Login)
		}

		// 需要登录的接口
		protected := api.Group("")
		protected.Use(authMiddleware())
		{
			// 用户
			user := protected.Group("/user")
			{
				user.GET("/info", h.GetUserInfo)
				user.PUT("/info", h.UpdateUserInfo)
			}

			// 宝宝
			children := protected.Group("/children")
			{
				children.GET("", h.GetChildren)
				children.POST("", h.CreateChild)
				children.GET("/:id", h.GetChildDetail)
				children.PUT("/:id", h.UpdateChild)
				children.DELETE("/:id", h.DeleteChild)
				children.POST("/switch", h.SwitchChild)
				children.POST("/:id/growth-stage", h.SetGrowthStage)
				children.GET("/:id/alerts", h.GetChildAlerts)

				// 环境问卷评估
				children.POST("/:id/environment-assessment", h.CreateEnvironmentAssessment)
				children.GET("/:id/environment-assessment/latest", h.GetLatestEnvironmentAssessment)
				children.GET("/:id/environment-assessment/history", h.GetEnvironmentAssessmentHistory)

				// 靶身高与生长速度
				children.GET("/:id/target-height-comparison", h.GetTargetHeightComparison)
				children.GET("/:id/growth-velocity", h.GetGrowthVelocity)
			}

			// 预警
			alerts := protected.Group("/alerts")
			{
				alerts.POST("/:alertId/read", h.MarkAlertRead)
				alerts.POST("/:alertId/dismiss", h.DismissAlert)
				alerts.GET("/summary", h.GetAlertsSummary)
			}

			// 记录
			records := protected.Group("/records")
			{
				records.GET("", h.GetRecords)
				records.POST("", h.CreateRecord)
				records.PUT("/:id", h.UpdateRecord)
				records.DELETE("/:id", h.DeleteRecord)
			}

			// 订阅
			subscription := protected.Group("/subscription")
			{
				subscription.GET("", h.GetSubscription)
				subscription.POST("/createOrder", h.CreateOrder)
			}

			// 家庭
			family := protected.Group("/family")
			{
				family.GET("", h.GetFamily)
				family.POST("", h.CreateFamily)
				family.POST("/join", h.JoinFamily)
				family.DELETE("/leave", h.LeaveFamily)
				family.PUT("/members/:id/role", h.UpdateMemberRole)
				family.POST("/inviteCode", h.GenerateInviteCode)
			}

			// AI
			ai := protected.Group("/ai")
			{
				ai.POST("/chat", h.Chat)
				ai.POST("/parseReport", h.ParseReport)
			}

			// 首页
			protected.GET("/home", h.GetHomeData)
		}
	}

	// 微信支付回调 (不需要登录验证，但需要签名验证)
	api.POST("/pay/callback", h.PayCallback)
}

// ========== 健康检查 ==========

func (h *Handler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, models.BaseResponse{
		Code: models.CodeSuccess,
		Msg:  "success",
		Data: map[string]string{
			"status":  "healthy",
			"version": "1.0.0",
		},
	})
}

// ========== 认证 ==========

func (h *Handler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.BaseResponse{
			Code: models.CodeParamError,
			Msg:  "参数错误: " + err.Error(),
		})
		return
	}

	resp, err := h.service.Login(c.Request.Context(), req.Code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.BaseResponse{
			Code: models.CodeServerError,
			Msg:  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.BaseResponse{
		Code: models.CodeSuccess,
		Msg:  "success",
		Data: resp,
	})
}

// ========== 用户 ==========

func (h *Handler) GetUserInfo(c *gin.Context) {
	userID := c.GetString("user_id")

	user, err := h.service.GetUserInfo(c.Request.Context(), userID)
	if err != nil || user == nil {
		c.JSON(http.StatusNotFound, models.BaseResponse{
			Code: models.CodeNotFound,
			Msg:  "用户不存在",
		})
		return
	}

	c.JSON(http.StatusOK, models.BaseResponse{
		Code: models.CodeSuccess,
		Msg:  "success",
		Data: user,
	})
}

func (h *Handler) UpdateUserInfo(c *gin.Context) {
	userID := c.GetString("user_id")

	var req models.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.BaseResponse{
			Code: models.CodeParamError,
			Msg:  "参数错误: " + err.Error(),
		})
		return
	}

	if err := h.service.UpdateUser(c.Request.Context(), userID, &req); err != nil {
		c.JSON(http.StatusInternalServerError, models.BaseResponse{
			Code: models.CodeServerError,
			Msg:  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.BaseResponse{
		Code: models.CodeSuccess,
		Msg:  "更新成功",
	})
}

// ========== 宝宝 ==========

func (h *Handler) GetChildren(c *gin.Context) {
	userID := c.GetString("user_id")

	children, err := h.service.GetChildren(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.BaseResponse{
			Code: models.CodeServerError,
			Msg:  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.BaseResponse{
		Code: models.CodeSuccess,
		Msg:  "success",
		Data: children,
	})
}

func (h *Handler) CreateChild(c *gin.Context) {
	userID := c.GetString("user_id")

	var req models.CreateChildRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.BaseResponse{
			Code: models.CodeParamError,
			Msg:  "参数错误: " + err.Error(),
		})
		return
	}

	child, err := h.service.CreateChild(c.Request.Context(), userID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.BaseResponse{
			Code: models.CodeServerError,
			Msg:  err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, models.BaseResponse{
		Code: models.CodeSuccess,
		Msg:  "创建成功",
		Data: child,
	})
}

func (h *Handler) GetChildDetail(c *gin.Context) {
	userID := c.GetString("user_id")
	childID := c.Param("id")

	detail, err := h.service.GetChildDetail(c.Request.Context(), userID, childID)
	if err != nil {
		c.JSON(http.StatusNotFound, models.BaseResponse{
			Code: models.CodeNotFound,
			Msg:  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.BaseResponse{
		Code: models.CodeSuccess,
		Msg:  "success",
		Data: detail,
	})
}

func (h *Handler) UpdateChild(c *gin.Context) {
	userID := c.GetString("user_id")
	childID := c.Param("id")

	var req models.UpdateChildRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.BaseResponse{
			Code: models.CodeParamError,
			Msg:  "参数错误: " + err.Error(),
		})
		return
	}

	if err := h.service.UpdateChild(c.Request.Context(), userID, childID, &req); err != nil {
		c.JSON(http.StatusInternalServerError, models.BaseResponse{
			Code: models.CodeServerError,
			Msg:  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.BaseResponse{
		Code: models.CodeSuccess,
		Msg:  "更新成功",
	})
}

func (h *Handler) DeleteChild(c *gin.Context) {
	userID := c.GetString("user_id")
	childID := c.Param("id")

	if err := h.service.DeleteChild(c.Request.Context(), userID, childID); err != nil {
		c.JSON(http.StatusInternalServerError, models.BaseResponse{
			Code: models.CodeServerError,
			Msg:  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.BaseResponse{
		Code: models.CodeSuccess,
		Msg:  "删除成功",
	})
}

func (h *Handler) SwitchChild(c *gin.Context) {
	userID := c.GetString("user_id")

	var req models.SwitchChildRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.BaseResponse{
			Code: models.CodeParamError,
			Msg:  "参数错误: " + err.Error(),
		})
		return
	}

	if err := h.service.SwitchChild(c.Request.Context(), userID, req.ChildID); err != nil {
		c.JSON(http.StatusInternalServerError, models.BaseResponse{
			Code: models.CodeServerError,
			Msg:  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.BaseResponse{
		Code: models.CodeSuccess,
		Msg:  "切换成功",
	})
}

// ========== 记录 ==========

func (h *Handler) GetRecords(c *gin.Context) {
	childID := c.Query("child_id")
	if childID == "" {
		c.JSON(http.StatusBadRequest, models.BaseResponse{
			Code: models.CodeParamError,
			Msg:  "缺少child_id参数",
		})
		return
	}

	var req models.RecordListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.BaseResponse{
			Code: models.CodeParamError,
			Msg:  "参数错误: " + err.Error(),
		})
		return
	}
	req.ChildID = childID

	resp, err := h.service.GetRecords(c.Request.Context(), childID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.BaseResponse{
			Code: models.CodeServerError,
			Msg:  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.BaseResponse{
		Code: models.CodeSuccess,
		Msg:  "success",
		Data: resp,
	})
}

func (h *Handler) CreateRecord(c *gin.Context) {
	userID := c.GetString("user_id")

	var req models.CreateRecordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.BaseResponse{
			Code: models.CodeParamError,
			Msg:  "参数错误: " + err.Error(),
		})
		return
	}

	record, err := h.service.CreateRecord(c.Request.Context(), userID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.BaseResponse{
			Code: models.CodeServerError,
			Msg:  err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, models.BaseResponse{
		Code: models.CodeSuccess,
		Msg:  "创建成功",
		Data: record,
	})
}

func (h *Handler) UpdateRecord(c *gin.Context) {
	userID := c.GetString("user_id")
	recordID := c.Param("id")

	var req models.UpdateRecordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.BaseResponse{
			Code: models.CodeParamError,
			Msg:  "参数错误: " + err.Error(),
		})
		return
	}

	if err := h.service.UpdateRecord(c.Request.Context(), userID, recordID, &req); err != nil {
		c.JSON(http.StatusInternalServerError, models.BaseResponse{
			Code: models.CodeServerError,
			Msg:  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.BaseResponse{
		Code: models.CodeSuccess,
		Msg:  "更新成功",
	})
}

func (h *Handler) DeleteRecord(c *gin.Context) {
	userID := c.GetString("user_id")
	recordID := c.Param("id")

	if err := h.service.DeleteRecord(c.Request.Context(), userID, recordID); err != nil {
		c.JSON(http.StatusInternalServerError, models.BaseResponse{
			Code: models.CodeServerError,
			Msg:  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.BaseResponse{
		Code: models.CodeSuccess,
		Msg:  "删除成功",
	})
}

// ========== 订阅 ==========

func (h *Handler) GetSubscription(c *gin.Context) {
	userID := c.GetString("user_id")

	resp, err := h.service.GetSubscription(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.BaseResponse{
			Code: models.CodeServerError,
			Msg:  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.BaseResponse{
		Code: models.CodeSuccess,
		Msg:  "success",
		Data: resp,
	})
}

func (h *Handler) CreateOrder(c *gin.Context) {
	userID := c.GetString("user_id")

	var req models.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.BaseResponse{
			Code: models.CodeParamError,
			Msg:  "参数错误: " + err.Error(),
		})
		return
	}

	order, err := h.service.CreateOrder(c.Request.Context(), userID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.BaseResponse{
			Code: models.CodeServerError,
			Msg:  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.BaseResponse{
		Code: models.CodeSuccess,
		Msg:  "success",
		Data: order,
	})
}

func (h *Handler) PayCallback(c *gin.Context) {
	// 解析微信支付回调XML
	var xmlData map[string]string
	if err := c.ShouldBindXML(&xmlData); err != nil {
		c.XML(http.StatusBadRequest, map[string]string{
			"return_code": "FAIL",
			"return_msg":  "参数错误",
		})
		return
	}

	if err := h.service.ProcessPayCallback(c.Request.Context(), xmlData); err != nil {
		c.XML(http.StatusOK, map[string]string{
			"return_code": "FAIL",
			"return_msg":  err.Error(),
		})
		return
	}

	c.XML(http.StatusOK, map[string]string{
		"return_code": "SUCCESS",
		"return_msg":  "OK",
	})
}

// ========== 家庭 ==========

func (h *Handler) GetFamily(c *gin.Context) {
	userID := c.GetString("user_id")

	resp, err := h.service.GetFamily(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, models.BaseResponse{
			Code: models.CodeNotFound,
			Msg:  "未找到家庭",
		})
		return
	}

	c.JSON(http.StatusOK, models.BaseResponse{
		Code: models.CodeSuccess,
		Msg:  "success",
		Data: resp,
	})
}

func (h *Handler) CreateFamily(c *gin.Context) {
	userID := c.GetString("user_id")

	var req models.CreateFamilyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.BaseResponse{
			Code: models.CodeParamError,
			Msg:  "参数错误: " + err.Error(),
		})
		return
	}

	family, err := h.service.CreateFamily(c.Request.Context(), userID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.BaseResponse{
			Code: models.CodeServerError,
			Msg:  err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, models.BaseResponse{
		Code: models.CodeSuccess,
		Msg:  "创建成功",
		Data: family,
	})
}

func (h *Handler) JoinFamily(c *gin.Context) {
	userID := c.GetString("user_id")

	var req models.JoinFamilyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.BaseResponse{
			Code: models.CodeParamError,
			Msg:  "参数错误: " + err.Error(),
		})
		return
	}

	if err := h.service.JoinFamily(c.Request.Context(), userID, &req); err != nil {
		statusCode := http.StatusInternalServerError
		code := models.CodeServerError
		if err.Error() == "邀请码无效" {
			code = models.CodeInvalidInviteCode
		} else if err.Error() == "家庭成员已满" {
			code = models.CodeFamilyFull
		}
		c.JSON(statusCode, models.BaseResponse{
			Code: code,
			Msg:  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.BaseResponse{
		Code: models.CodeSuccess,
		Msg:  "加入成功",
	})
}

func (h *Handler) LeaveFamily(c *gin.Context) {
	userID := c.GetString("user_id")

	if err := h.service.LeaveFamily(c.Request.Context(), userID); err != nil {
		c.JSON(http.StatusInternalServerError, models.BaseResponse{
			Code: models.CodeServerError,
			Msg:  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.BaseResponse{
		Code: models.CodeSuccess,
		Msg:  "退出成功",
	})
}

func (h *Handler) UpdateMemberRole(c *gin.Context) {
	userID := c.GetString("user_id")
	memberID := c.Param("id")

	var req models.UpdateMemberRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.BaseResponse{
			Code: models.CodeParamError,
			Msg:  "参数错误: " + err.Error(),
		})
		return
	}

	if err := h.service.UpdateMemberRole(c.Request.Context(), userID, memberID, req.Role); err != nil {
		c.JSON(http.StatusInternalServerError, models.BaseResponse{
			Code: models.CodeServerError,
			Msg:  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.BaseResponse{
		Code: models.CodeSuccess,
		Msg:  "更新成功",
	})
}

func (h *Handler) GenerateInviteCode(c *gin.Context) {
	userID := c.GetString("user_id")

	resp, err := h.service.GenerateInviteCode(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, models.BaseResponse{
			Code: models.CodeNotFound,
			Msg:  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.BaseResponse{
		Code: models.CodeSuccess,
		Msg:  "success",
		Data: resp,
	})
}

// ========== AI ==========

func (h *Handler) Chat(c *gin.Context) {
	userID := c.GetString("user_id")

	var req models.AIChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.BaseResponse{
			Code: models.CodeParamError,
			Msg:  "参数错误: " + err.Error(),
		})
		return
	}

	resp, err := h.service.Chat(c.Request.Context(), userID, &req)
	if err != nil {
		statusCode := http.StatusInternalServerError
		code := models.CodeServerError
		if err.Error() == "AI额度已用完，请升级会员" {
			code = models.CodeQuotaExhausted
		}
		c.JSON(statusCode, models.BaseResponse{
			Code: code,
			Msg:  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.BaseResponse{
		Code: models.CodeSuccess,
		Msg:  "success",
		Data: resp,
	})
}

func (h *Handler) ParseReport(c *gin.Context) {
	userID := c.GetString("user_id")

	var req models.ParseReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.BaseResponse{
			Code: models.CodeParamError,
			Msg:  "参数错误: " + err.Error(),
		})
		return
	}

	resp, err := h.service.ParseReport(c.Request.Context(), userID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.BaseResponse{
			Code: models.CodeServerError,
			Msg:  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.BaseResponse{
		Code: models.CodeSuccess,
		Msg:  "success",
		Data: resp,
	})
}

// ========== 首页 ==========

func (h *Handler) GetHomeData(c *gin.Context) {
	userID := c.GetString("user_id")

	resp, err := h.service.GetHomeData(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.BaseResponse{
			Code: models.CodeServerError,
			Msg:  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.BaseResponse{
		Code: models.CodeSuccess,
		Msg:  "success",
		Data: resp,
	})
}

// ========== 中间件 ==========

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Header("Access-Control-Max-Age", "86400")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

func authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token == "" {
			c.JSON(http.StatusUnauthorized, models.BaseResponse{
				Code: models.CodeUnauthorized,
				Msg:  "未登录",
			})
			c.Abort()
			return
		}

		// 简化处理：Bearer token格式
		if len(token) > 7 && token[:7] == "Bearer " {
			token = token[7:]
		}

		// 实际实现需要验证JWT token
		// 这里简化处理，假设token格式为: jwt_token_{user_id}_{timestamp}
		parts := splitToken(token)
		if len(parts) < 2 {
			c.JSON(http.StatusUnauthorized, models.BaseResponse{
				Code: models.CodeUnauthorized,
				Msg:  "无效的token",
			})
			c.Abort()
			return
		}

		// 解析user_id: jwt_token_{user_id}_{timestamp} -> parts[2]
		userID := parts[2]
		if userID == "" || len(parts) < 3 {
			c.JSON(http.StatusUnauthorized, models.BaseResponse{
				Code: models.CodeUnauthorized,
				Msg:  "无效的token",
			})
			c.Abort()
			return
		}

		c.Set("user_id", userID)
		c.Next()
	}
}

func splitToken(token string) []string {
	var parts []string
	start := 0
	for i := 0; i < len(token); i++ {
		if token[i] == '_' && i > start {
			parts = append(parts, token[start:i])
			start = i + 1
		}
	}
	if start < len(token) {
		parts = append(parts, token[start:])
	}
	return parts
}

// ========== 预警 ==========

func (h *Handler) SetGrowthStage(c *gin.Context) {
	userID := c.GetString("user_id")
	childID := c.Param("id")

	var req models.SetGrowthStageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.BaseResponse{
			Code: models.CodeParamError,
			Msg:  "参数错误: " + err.Error(),
		})
		return
	}

	if err := h.service.SetGrowthStage(c.Request.Context(), userID, childID, &req); err != nil {
		c.JSON(http.StatusInternalServerError, models.BaseResponse{
			Code: models.CodeServerError,
			Msg:  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.BaseResponse{
		Code: models.CodeSuccess,
		Msg:  "设置成功",
	})
}

func (h *Handler) GetChildAlerts(c *gin.Context) {
	userID := c.GetString("user_id")
	childID := c.Param("id")

	var req models.AlertListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.BaseResponse{
			Code: models.CodeParamError,
			Msg:  "参数错误: " + err.Error(),
		})
		return
	}
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}

	resp, err := h.service.GetChildAlerts(c.Request.Context(), userID, childID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.BaseResponse{
			Code: models.CodeServerError,
			Msg:  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.BaseResponse{
		Code: models.CodeSuccess,
		Msg:  "success",
		Data: resp,
	})
}

func (h *Handler) MarkAlertRead(c *gin.Context) {
	userID := c.GetString("user_id")
	alertID := c.Param("alertId")

	if err := h.service.MarkAlertRead(c.Request.Context(), userID, alertID); err != nil {
		c.JSON(http.StatusNotFound, models.BaseResponse{
			Code: models.CodeNotFound,
			Msg:  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.BaseResponse{
		Code: models.CodeSuccess,
		Msg:  "标记成功",
	})
}

func (h *Handler) DismissAlert(c *gin.Context) {
	userID := c.GetString("user_id")

	var req models.DismissAlertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.BaseResponse{
			Code: models.CodeParamError,
			Msg:  "参数错误: " + err.Error(),
		})
		return
	}

	if err := h.service.DismissAlert(c.Request.Context(), userID, &req); err != nil {
		c.JSON(http.StatusNotFound, models.BaseResponse{
			Code: models.CodeNotFound,
			Msg:  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.BaseResponse{
		Code: models.CodeSuccess,
		Msg:  "已忽略",
	})
}

func (h *Handler) GetAlertsSummary(c *gin.Context) {
	userID := c.GetString("user_id")

	resp, err := h.service.GetAlertsSummary(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.BaseResponse{
			Code: models.CodeServerError,
			Msg:  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.BaseResponse{
		Code: models.CodeSuccess,
		Msg:  "success",
		Data: resp,
	})
}

// ========== 环境问卷评估 ==========

func (h *Handler) CreateEnvironmentAssessment(c *gin.Context) {
	userID := c.GetString("user_id")
	childID := c.Param("id")

	var req models.CreateEnvironmentAssessmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.BaseResponse{
			Code: models.CodeParamError,
			Msg:  "参数错误: " + err.Error(),
		})
		return
	}

	resp, err := h.service.CreateEnvironmentAssessment(c.Request.Context(), userID, childID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.BaseResponse{
			Code: models.CodeServerError,
			Msg:  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.BaseResponse{
		Code: models.CodeSuccess,
		Msg:  "success",
		Data: resp,
	})
}

func (h *Handler) GetLatestEnvironmentAssessment(c *gin.Context) {
	userID := c.GetString("user_id")
	childID := c.Param("id")

	resp, err := h.service.GetLatestEnvironmentAssessment(c.Request.Context(), userID, childID)
	if err != nil {
		if err.Error() == "暂无评估记录" {
			c.JSON(http.StatusNotFound, models.BaseResponse{
				Code: models.CodeNotFound,
				Msg:  err.Error(),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, models.BaseResponse{
			Code: models.CodeServerError,
			Msg:  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.BaseResponse{
		Code: models.CodeSuccess,
		Msg:  "success",
		Data: resp,
	})
}

func (h *Handler) GetEnvironmentAssessmentHistory(c *gin.Context) {
	userID := c.GetString("user_id")
	childID := c.Param("id")

	page := parseInt(c.Query("page"), 1)
	pageSize := parseInt(c.Query("page_size"), 20)

	resp, err := h.service.GetEnvironmentAssessmentHistory(c.Request.Context(), userID, childID, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.BaseResponse{
			Code: models.CodeServerError,
			Msg:  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.BaseResponse{
		Code: models.CodeSuccess,
		Msg:  "success",
		Data: resp,
	})
}

// ========== 靶身高与生长速度 ==========

func (h *Handler) GetTargetHeightComparison(c *gin.Context) {
	userID := c.GetString("user_id")
	childID := c.Param("id")

	resp, err := h.service.GetTargetHeightComparison(c.Request.Context(), userID, childID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.BaseResponse{
			Code: models.CodeServerError,
			Msg:  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.BaseResponse{
		Code: models.CodeSuccess,
		Msg:  "success",
		Data: resp,
	})
}

func (h *Handler) GetGrowthVelocity(c *gin.Context) {
	userID := c.GetString("user_id")
	childID := c.Param("id")
	monthsBack := parseInt(c.Query("months_back"), 12)

	resp, err := h.service.GetGrowthVelocity(c.Request.Context(), userID, childID, monthsBack)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.BaseResponse{
			Code: models.CodeServerError,
			Msg:  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.BaseResponse{
		Code: models.CodeSuccess,
		Msg:  "success",
		Data: resp,
	})
}

// 辅助函数
func parseInt(s string, defaultVal int) int {
	if val, err := strconv.Atoi(s); err == nil {
		return val
	}
	return defaultVal
}
