package routes

import (
	"pickup/internal/scheduler/controllers"
	"pickup/internal/scheduler/middlewares"
	"pickup/internal/utils"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine, authCtl *controllers.AuthController, studentCtl *controllers.StudentController, adminCtl *controllers.AdminController, jwtUtil *utils.JWTUtil) {
	api := r.Group("/api/v1")

	auth := api.Group("/auth")
	auth.POST("/login", authCtl.Login)

	authProtected := auth.Group("")
	authProtected.Use(middlewares.JWTAuth(jwtUtil))
	authProtected.POST("/bind-phone", authCtl.BindPhone)
	authProtected.GET("/me", authCtl.Me)

	student := api.Group("/student")
	student.Use(middlewares.JWTAuth(jwtUtil), middlewares.RequireRoles("student"))
	student.POST("/requests", studentCtl.CreateRequest)
	student.GET("/requests/my", studentCtl.MyRequests)
	student.PUT("/requests/:id", studentCtl.UpdateRequest)

	admin := api.Group("/admin")
	admin.Use(middlewares.JWTAuth(jwtUtil), middlewares.RequireRoles("staff"))
	admin.GET("/drivers", adminCtl.ListDrivers)
	admin.POST("/drivers", adminCtl.CreateDriver)
	admin.PUT("/drivers/:id", adminCtl.UpdateDriver)
	admin.GET("/shifts/dashboard", adminCtl.Dashboard)
	admin.GET("/requests/pending", adminCtl.PendingRequests)
	admin.GET("/users", middlewares.RequireRoles("admin"), adminCtl.ListUsers)
	admin.POST("/users/:id/set-staff", middlewares.RequireRoles("admin"), adminCtl.SetStaff)
	admin.POST("/users/:id/unset-staff", middlewares.RequireRoles("admin"), adminCtl.UnsetStaff)
	admin.POST("/shifts", adminCtl.CreateShift)
	admin.PUT("/shifts/:id", adminCtl.UpdateShift)
	admin.POST("/shifts/:id/assign-student", adminCtl.AssignStudent)
	admin.POST("/shifts/:id/remove-student", adminCtl.RemoveStudent)
	admin.POST("/shifts/:id/assign-staff", middlewares.RequireRoles("admin"), adminCtl.AssignStaff)
	admin.POST("/shifts/:id/remove-staff", adminCtl.RemoveStaff)
	admin.POST("/shifts/:id/publish", adminCtl.PublishShift)

	api.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})
}
