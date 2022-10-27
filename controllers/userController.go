package controllers

import (
	"context"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/joelpatel/go-jwt/database"
	"github.com/joelpatel/go-jwt/helpers"
	"github.com/joelpatel/go-jwt/models"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

var userCollection *mongo.Collection = database.OpenCollection(database.Client, "users")
var validate = validator.New()

var PasswordHashCost string

func loadPasswordHashCost() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading the dotenv file.\nerror: %v\n", err.Error())
	}
	PasswordHashCost = os.Getenv("PASSWORD_HASH_COST")
}

func HashPassword(password *string) string {
	if PasswordHashCost == "" {
		loadPasswordHashCost()
	}

	cost, _ := strconv.Atoi(PasswordHashCost)
	hashedPasswordBytes, err := bcrypt.GenerateFromPassword([]byte(*password), cost)
	if err != nil {
		log.Fatalf("Error hasing the password.\nerror: %v\n", err)
	}
	return string(hashedPasswordBytes)
}

func VerifyPassword(requestUserPassword *string, dbUserPassword *string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(*dbUserPassword), []byte(*requestUserPassword))
	return err == nil
}

func Signup() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var c, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var user models.User

		if err := ctx.BindJSON(&user); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "error parsing the request body\n" + err.Error()})
			return
		}

		validationErr := validate.Struct(user)
		if validationErr != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "validation failed\n" + validationErr.Error()})
			return
		}

		emailCount, err := userCollection.CountDocuments(c, bson.M{"email": user.Email})
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if emailCount != 0 {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Email already in use. Please enter a different email or login using the already signed email."})
			return
		}

		phoneCount, err := userCollection.CountDocuments(c, bson.M{"phone": user.Phone})
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if phoneCount != 0 {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Phone already in use. Please enter a different phone number."})
			return
		}

		// hashedPassword := HashPassword(&user.Password)
		// user.Password = hashedPassword
		user.Password = HashPassword(&user.Password)

		user.CreatedAt, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		user.UpdatedAt = user.CreatedAt
		user.ID = primitive.NewObjectID()
		user.UserID = user.ID.Hex()
		token, refreshToken := helpers.GenerateAllTokens(&user.Email, &user.FirstName, &user.UserType)
		user.Token = token
		user.RefreshToken = refreshToken

		insertionNumber, insertError := userCollection.InsertOne(c, &user)
		if insertError != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": insertError.Error()})
			return
		}

		ctx.JSON(http.StatusOK, insertionNumber)
	}
}

func Login() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		c, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var requestUser models.User
		var databaseUser models.User

		if err := ctx.BindJSON(&requestUser); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "error parsing the request body\n" + err.Error()})
			return
		}

		err := userCollection.FindOne(c, bson.M{"email": requestUser.Email}).Decode(&databaseUser)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		passwordIsValid := VerifyPassword(&requestUser.Password, &databaseUser.Password)

		if !passwordIsValid {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Email or password is incorrect."})
			return
		}

		token, refreshToken := helpers.GenerateAllTokens(&(databaseUser.Email), &(databaseUser.FirstName), &(databaseUser.UserType))
		helpers.UpdateAllTokens(token, refreshToken, &(databaseUser.UserID))

		// TODO: replace the following line to code which updates local databaseUser directly
		err = userCollection.FindOne(c, bson.M{"user_id": databaseUser.UserID}).Decode(&databaseUser)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		ctx.JSON(http.StatusOK, databaseUser)
	}
}

func GetUsers() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if err := helpers.CheckUserType(ctx, "ADMIN"); err != nil {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}

		c, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		recordPerPageString, exists := ctx.GetQuery("recordPerPage")
		recordPerPage, err := strconv.Atoi(recordPerPageString)
		if err != nil || recordPerPage < 1 || !exists {
			recordPerPage = 10
		}

		page, err := strconv.Atoi(ctx.Query("page"))
		if err != nil || page < 1 {
			page = 1
		}

		// startIndex, _ := strconv.Atoi(ctx.Query("startIndex"))
		// matchStage := bson.D{{Key: "$match", Value: bson.D{{}}}}
		// groupStage := bson.D{{Key: "$group", Value: bson.D{
		// 	{Key: "_id", Value: bson.D{{Key: "_id", Value: "null"}}},
		// 	{Key: "total_count", Value: bson.D{{Key: "$sum", Value: 1}}},
		// 	{Key: "data", Value: bson.D{{Key: "$push", Value: "$$ROOT"}}},
		// }}}
		// projectStage := bson.D{ // which all datapoints should go the the user and which all shouldn't
		// 	{Key: "$project", Value: bson.D{
		// 		{Key: "_id", Value: 0},
		// 		{Key: "total_count", Value: 1},
		// 		{Key: "user_items", Value: bson.D{
		// 			{Key: "#slice", Value: "null"},
		// 		}},
		// 	}},
		// }
		// result, err := userCollection.Aggregate(c, mongo.Pipeline{
		// 	matchStage,
		// 	groupStage,
		// 	projectStage,
		// })

		startIndex := (page - 1) * recordPerPage

		filter := bson.D{{}}
		option := new(options.FindOptions)

		option.SetSkip(int64(startIndex))
		option.SetLimit(int64(recordPerPage))

		result, err := userCollection.Find(c, filter, option)

		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Error occured while aggregating users.\n" + err.Error()})
			return
		}
		var allUsers []bson.M
		if err = result.All(c, &allUsers); err != nil {
			log.Println("error: " + err.Error())
			return
		}
		ctx.JSON(http.StatusOK, allUsers)
	}
}

func GetUser() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		userID := ctx.Param("user_id")
		if err := helpers.MatchUserTypeToUserID(ctx, userID); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var user models.User
		err := userCollection.FindOne(c, bson.M{"user_id": userID}).Decode(&user)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "error occured when trying to find one user from db\n" + err.Error()})
			return
		}
		ctx.JSON(http.StatusOK, user)
	}
}
