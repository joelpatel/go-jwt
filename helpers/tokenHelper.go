package helpers

import (
	"context"
	"log"
	"os"
	"time"

	jwt "github.com/golang-jwt/jwt"
	"github.com/joelpatel/go-jwt/database"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/joho/godotenv"
)

type SignedDetails struct {
	Email     string
	FirstName string
	UserType  string
	jwt.StandardClaims
}

var userCollection *mongo.Collection = database.OpenCollection(database.Client, "users")

var JWT_Secret_Key string

func loadSecretKey() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading the dotenv file.\nerror: %v\n", err.Error())
	}
	JWT_Secret_Key = os.Getenv("JWT_PRIVATE_KEY")
}

func GenerateAllTokens(email *string, firstName *string, userType *string) (string, string) {
	if JWT_Secret_Key == "" {
		loadSecretKey()
	}

	claims := &SignedDetails{
		Email:     *email,
		FirstName: *firstName,
		UserType:  *userType,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Local().Add(time.Hour * 24).Unix(),
		},
	}

	refreshClaims := &SignedDetails{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Local().Add(time.Hour * 24 * 365).Unix(),
		},
	}

	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(JWT_Secret_Key))
	if err != nil {
		log.Fatalf("Error creating jwt token.\nerror: %v\n", err.Error())
	}

	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString([]byte(JWT_Secret_Key))
	if err != nil {
		log.Fatalf("Error creating jwt refreshToken.\nerror: %v\n", err.Error())
	}

	return token, refreshToken
}

func UpdateAllTokens(token string, refreshToken string, userID *string) {
	c, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	var updateObject primitive.D
	updateObject = append(updateObject, bson.E{Key: "token", Value: token})
	updateObject = append(updateObject, bson.E{Key: "refresh_token", Value: refreshToken})
	updatedAt, _ := time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
	updateObject = append(updateObject, bson.E{Key: "updated_at", Value: updatedAt})

	upsert := true
	filter := bson.M{"user_id": userID}
	opt := options.UpdateOptions{
		Upsert: &upsert,
	}

	_, err := userCollection.UpdateOne(
		c,
		filter,
		bson.D{
			{Key: "$set", Value: updateObject},
		},
		&opt,
	)

	if err != nil {
		log.Panic(err.Error())
	}
}

func ValidateToken(token string) (*SignedDetails, string) {
	if JWT_Secret_Key == "" {
		loadSecretKey()
	}
	parsedToken, err := jwt.ParseWithClaims(
		token,
		&SignedDetails{},
		func(t *jwt.Token) (interface{}, error) {
			return []byte(JWT_Secret_Key), nil
		},
	)

	if err != nil {
		return nil, err.Error()
	}

	claims, ok := parsedToken.Claims.(*SignedDetails)
	if !ok {
		return nil, "token is invalid"
	}

	if claims.ExpiresAt < time.Now().Local().Unix() {
		return nil, "token is expired"
	}

	return claims, ""
}
