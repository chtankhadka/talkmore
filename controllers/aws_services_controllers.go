package controllers

import (
	"encoding/base64"
	"encoding/json"
	"my-work/config"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

func ImageVarification(app *config.AppConfig) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		match, err := compareFaces("ss.jpg", "ss.jpg")
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, bson.M{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusOK, bson.M{"success": match})
	}
}
func ImageDetectFace(app *config.AppConfig) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req FaceRequest

		// 🔹 Parse JSON body
		if err := json.NewDecoder(ctx.Request.Body).Decode(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format"})
			return
		}

		// 🔹 Decode Base64 to byte arrays
		sourceBytes, err := base64.StdEncoding.DecodeString(req.Source)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid source image encoding"})
			return
		}
		faceCount, err := detectFaces(sourceBytes)

		if err != nil {
			ctx.JSON(http.StatusInternalServerError, bson.M{"error": err.Error()})
			return
		}
		response := map[string]interface{}{
			"face_count": faceCount,
		}
		ctx.JSON(http.StatusOK, bson.M{"success": response})

	}
}
