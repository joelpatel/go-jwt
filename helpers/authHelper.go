package helpers

import (
	"errors"

	"github.com/gin-gonic/gin"
)

/*
type == "USER" -> they can access only their data

type == "ADMIN" -> they can access info. about every users

	@param {string} userID - the user id belongs to the search space and not the person who initiated the request
*/
func MatchUserTypeToUserID(ctx *gin.Context, userID string) error {
	userType := ctx.GetString("user_type")
	uID := ctx.GetString("uid")

	/*
		User with uID is trying to access another user with userID.
		Thus, return an error.
	*/
	if userType == "USER" && uID != userID {
		return errors.New("unauthorized to access this resource")
	}

	return nil
}

func CheckUserType(c *gin.Context, role string) error {
	userType := c.GetString("user_type")
	if userType != role {
		return errors.New("unauthorized to access this resource")
	}
	return nil
}
