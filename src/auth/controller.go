package auth

import (
	"main/src/common"
	"main/src/user"
	"net/http"

	"github.com/gin-gonic/gin"
)

type UserController struct {
	UserRepo *user.Repository
}

func NewController(repo *user.Repository) *UserController {
	return &UserController{UserRepo: repo}
}

func (ctrl *UserController) Login(c *gin.Context) {
	var input LoginInput

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid input!"})
		return
	}

	u, err := ctrl.UserRepo.FindByEmail(input.Email)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email!"})
		return
	}

	if !common.ComparePassword(u.Password, input.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid password!"})
		return
	}
	token, err := GenerateToken(u.ID.Hex(), u.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create token!"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": token})
}
