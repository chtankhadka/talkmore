package routes

import (
	"my-work/config"
	"my-work/controllers"

	"github.com/gin-gonic/gin"
)

func UserRoutes(incomingRoutes *gin.RouterGroup, app *config.AppConfig) {
	incomingRoutes.POST("/logout", controllers.Logout(app))

}

func PublicRoutes(incomingRoutes *gin.Engine, app *config.AppConfig) {
	incomingRoutes.POST("/signin", controllers.SignIn(app))
	incomingRoutes.POST("/signup", controllers.SignUp(app))
	incomingRoutes.POST("/accountvalidate", controllers.ValidateOtpAndSaveUser(app))
	incomingRoutes.POST("/refreshtoken", controllers.RefreshToken(app))
}
