package controllers

import (
	"cowin-emailer/db"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type CreateUserInput struct {
	UID   string `json:"uid" binding:"required"`
	Email string `json:"email" binding:"required"`
}

type UpdateUserInput struct {
	Age      uint   `json:"age"`
	District string `json:"district"`
}

type DeleteUserInput struct {
	UID string `json:"uid" binding:"required"`
}

func CreateUser(c *gin.Context) {
	var input CreateUserInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if !CheckUID(input.UID) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthenticated User"})
		return
	}

	user := db.User{
		UID:      input.UID,
		Email:    input.Email,
		Age:      0,
		District: "",
	}

	db.DB.Create(&user)

	c.JSON(http.StatusCreated, gin.H{"data": user})
}

func UpdateUser(c *gin.Context) {
	var user db.User

	var input UpdateUserInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	uid := c.Request.Header.Get("Authorization")

	if !CheckUID(uid) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthenticated User"})
		return
	}

	if err := db.DB.Where("uid = ?", uid).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthenticated User"})
		return
	}

	new_user := db.User{
		UID:      user.UID,
		Email:    user.Email,
		District: input.District,
		Age:      input.Age,
	}
	db.DB.Model(&user).Updates(new_user)
	c.JSON(http.StatusOK, gin.H{"data": user})
}

func GetUser(c *gin.Context) {
	var user db.User

	uid := c.Request.Header.Get("Authorization")

	if err := db.DB.Where("uid = ?", uid).First(&user).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"user_exists": false})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": user, "user_exists": true})

}

func DeleteUser(c *gin.Context) {
	var input DeleteUserInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	uid := input.UID
	err := db.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("uid = ?", uid).Delete(&db.User{}).Error; err != nil {
			return err
		}
		if err := DeleteFirebaseUser(uid); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"user_deleted": false})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user_deleted": true})

}
