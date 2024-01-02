package routes

import (
	controller "demo/controllers"

	"github.com/gin-gonic/gin"
)

func AuthRoutes(incomingRoutes *gin.Engine) {
	incomingRoutes.POST("users/login", controller.Login())
	incomingRoutes.POST("users/signup", controller.Signup())
}
