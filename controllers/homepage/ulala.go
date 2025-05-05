package homepage

import (
	"context"
	"my-work/config"
	"my-work/controllers"
	"my-work/models/homepage"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// func InsertUlala(app *config.AppConfig) gin.HandlerFunc {
// 	return func(ctx *gin.Context) {
// 		mctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
// 		defer cancel()

// 	}
// }

// time,
func Ulala(app *config.AppConfig) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		mctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		var ulalaSkipLimit homepage.UlalaSkipLimitWithCurrentTime
		if err := ctx.ShouldBindJSON(&ulalaSkipLimit); err != nil {
			controllers.ErrorResponse(ctx, http.StatusBadRequest, "Parsing Error", err.Error())
			return
		}

		clientToken, tokenError := controllers.GetMyToken(ctx)
		if tokenError != "" {
			controllers.ErrorResponse(ctx, http.StatusUnauthorized, "token error", tokenError)
			ctx.Abort()
			return
		}
		userDetails, idError := controllers.GetMyId(mctx, app, clientToken)
		if idError != "" {
			controllers.ErrorResponse(ctx, http.StatusUnauthorized, "Id Error", idError)
			ctx.Abort()
		}
		userMoreDetails, detailError := controllers.GetUserMoreDetails(mctx, app, userDetails.UserID)
		if detailError != nil {
			controllers.ErrorResponse(ctx, http.StatusInternalServerError, "More Detail Error", detailError.Error())
			ctx.Abort()
		}
		println(userMoreDetails.UserID)

		//get first latest updates
		// GetLatestUpdatesWithFilter(mctx, app, ulalaSkipLimit, userMoreDetails)

	}
}

// func GetLatestUpdatesWithFilter(mctx context.Context, app *config.AppConfig, ulalaSkipLimit homepage.UlalaSkipLimitWithCurrentTime, userMoreDetails *models.UserMoreDetails) {
// 	// Validate and parse date fields
// 	startDate, err := time.Parse(time.RFC3339, ulalaSkipLimit.FromDate)
// 	if err != nil {
// 		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start_date format"})
// 		return
// 	}

// 	endDate, err := time.Parse(time.RFC3339, ulalaSkipLimit.ToDate)
// 	if err != nil {
// 		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end_date format"})
// 		return
// 	}

// 	// Build the MongoDB query filter
// 	filter := bson.M{
// 		"user_id": getTransactions.User_ID,
// 		"transaction_details.transaction_date": bson.M{
// 			"$gte": startDate,
// 			"$lte": endDate,
// 		},
// 	}
// 	collection := app.Client.Database("talkmore").Collection("ulala")

// 	filter := bson.M{
// 		"date": bson.M{
// 			"$gte": params.FromDate,
// 			"$lte": params.ToDate,
// 		},
// 	}

// 	opts := options.Find().
// 		SetSkip(int64(params.Skip)).
// 		SetLimit(int64(params.Limit)).
// 		SetSort(bson.D{{Key: "date", Value: 1}}) // Optional sorting by date

// 	cursor, err := collection.Find(ctx, filter, opts)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer cursor.Close(ctx)

// 	var results []bson.M
// 	if err := cursor.All(ctx, &results); err != nil {
// 		return nil, err
// 	}

// 	return results, nil

// }
