package routes

import (
	"my-work/config"
	"my-work/controllers"
	"my-work/websocket"

	"github.com/gin-gonic/gin"
)

func UserRoutes(incomingRoutes *gin.RouterGroup, app *config.AppConfig) {
	incomingRoutes.POST("/logout", controllers.Logout(app))
	incomingRoutes.POST("/message", controllers.SendText(app))
	// incomingRoutes.POST("/watchchats", controllers.WatchChats(app))

	incomingRoutes.POST("/chatlist", controllers.GetChats(app))
	incomingRoutes.POST("/getmessages", controllers.GetMessages(app))
	incomingRoutes.POST("/myprofile", controllers.MyProfile(app))
	incomingRoutes.POST("/uploadphotoforulala", controllers.UploadUlalaImageAndReturnUrl(app))

}

func PublicRoutes(incomingRoutes *gin.Engine, app *config.AppConfig) {
	incomingRoutes.POST("/imageverification", controllers.ImageVarification(app))
	incomingRoutes.POST("/facedetect", controllers.ImageDetectFace(app))
	incomingRoutes.POST("/getbytearray", controllers.GetByteArray())
	incomingRoutes.POST("/signin", controllers.SignIn(app))
	incomingRoutes.POST("/signup", controllers.SignUp(app))
	incomingRoutes.POST("/accountvalidate", controllers.ValidateOtpAndSaveUser(app))
	incomingRoutes.POST("/refreshtoken", controllers.RefreshToken(app))
}

func WebSocketRoutes(incomingRoutes *gin.RouterGroup, app *config.AppConfig) {
	incomingRoutes.GET("/ws/chats", websocket.HandleMessageListWebSocket(app))
	incomingRoutes.GET("/ws/messages", websocket.HandleMessageListWebSocket(app))
	// incomingRoutes.GET("/ws/messages", websocket.HandleMessageWebSocket(app))
}
