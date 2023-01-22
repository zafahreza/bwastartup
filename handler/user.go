package handler

import (
	"bwastartup/auth"
	"bwastartup/helper"
	"bwastartup/user"
	"cloud.google.com/go/storage"
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

type userHandler struct {
	userService user.Service
	authService auth.Service
}

func NewUserHandler(userService user.Service, authService auth.Service) *userHandler {
	return &userHandler{userService, authService}
}

func (h *userHandler) RegisterUser(c *gin.Context) {
	var input user.RegisterUserInput

	err := c.ShouldBindJSON(&input)
	if err != nil {
		errors := helper.FormatValidationError(err)
		errorMessage := gin.H{"errors": errors}

		response := helper.APIResponse("Register account failed", http.StatusUnprocessableEntity, "Error", errorMessage)
		c.JSON(http.StatusUnprocessableEntity, response)
		return
	}

	newUser, err := h.userService.RegisterUser(input)
	if err != nil {
		response := helper.APIResponse("Register account failed", http.StatusBadRequest, "Error", nil)
		c.JSON(http.StatusBadRequest, response)
		return
	}

	token, err := h.authService.GenerateToken(newUser.ID)
	if err != nil {
		response := helper.APIResponse("Register account failed", http.StatusBadRequest, "Error", nil)
		c.JSON(http.StatusBadRequest, response)
		return
	}

	formatter := user.FormatUser(newUser, token)
	response := helper.APIResponse("Account has been Registed", http.StatusOK, "Success", formatter)

	c.JSON(http.StatusOK, response)
}

func (h *userHandler) Login(c *gin.Context) {
	var input user.LoginInput

	err := c.ShouldBindJSON(&input)
	if err != nil {
		errors := helper.FormatValidationError(err)
		errorMessage := gin.H{"errors": errors}

		response := helper.APIResponse("Login Failed", http.StatusUnprocessableEntity, "Error", errorMessage)
		c.JSON(http.StatusUnprocessableEntity, response)
		return
	}

	loggedinUser, err := h.userService.Login(input)
	if err != nil {
		errorMessage := gin.H{"errors": err.Error()}
		response := helper.APIResponse("Login Failed", http.StatusBadRequest, "Error", errorMessage)
		c.JSON(http.StatusBadRequest, response)
		return
	}

	token, err := h.authService.GenerateToken(loggedinUser.ID)
	if err != nil {
		response := helper.APIResponse("Login failed", http.StatusBadRequest, "Error", nil)
		c.JSON(http.StatusBadRequest, response)
		return
	}

	formatter := user.FormatUser(loggedinUser, token)
	response := helper.APIResponse("Login Success", http.StatusOK, "Success", formatter)

	c.JSON(http.StatusOK, response)
}

func (h *userHandler) CheckEmailAvailability(c *gin.Context) {
	var input user.CheckEmailInput

	err := c.ShouldBindJSON(&input)
	if err != nil {
		errors := helper.FormatValidationError(err)
		errorMessage := gin.H{"errors": errors}

		response := helper.APIResponse("Email Checking Failed", http.StatusUnprocessableEntity, "Error", errorMessage)
		c.JSON(http.StatusUnprocessableEntity, response)
		return
	}

	isEmailAvailable, err := h.userService.IsEmailAvailable(input)
	if err != nil {
		errorMessage := gin.H{"errors": "server error"}

		response := helper.APIResponse("Email Checking Failed", http.StatusUnprocessableEntity, "Error", errorMessage)
		c.JSON(http.StatusUnprocessableEntity, response)
		return
	}

	data := gin.H{
		"is_available": isEmailAvailable,
	}

	metaMessage := "email has been registered"

	if isEmailAvailable {
		metaMessage = "email is available"
	}

	response := helper.APIResponse(metaMessage, http.StatusOK, "Success", data)

	c.JSON(http.StatusOK, response)
}

func (h *userHandler) UploadAvatar(c *gin.Context) {
	file, err := c.Request.MultipartReader()
	if err != nil {
		data := gin.H{"is_uploaded": false}
		response := helper.APIResponse("Upload avatar image failed", http.StatusBadRequest, "error", data)

		c.JSON(http.StatusBadRequest, response)
		return
	}

	var foundImage bool
	var fileName string
	currentUser := c.MustGet("currentUser").(user.User)
	userID := currentUser.ID
	for {
		next, err := file.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			data := gin.H{"is_uploaded": false}
			response := helper.APIResponse("Upload avatar image failed", http.StatusBadRequest, "error", data)

			c.JSON(http.StatusBadRequest, response)
			return
		}
		if next.FormName() == "avatar" {
			foundImage = true
			ctx := context.Background()
			client, err := storage.NewClient(ctx)
			if err != nil {
				data := gin.H{"is_uploaded": false}
				response := helper.APIResponse("Upload avatar image failed", http.StatusBadRequest, "error", data)

				c.JSON(http.StatusBadRequest, response)
				return
			}

			bucket := client.Bucket("donation_alert")
			w := bucket.Object(next.FileName()).NewWriter(ctx)
			if _, err := io.Copy(w, next); err != nil {
				data := gin.H{"is_uploaded": false}
				response := helper.APIResponse("Upload avatar image failed", http.StatusBadRequest, "error", data)

				c.JSON(http.StatusBadRequest, response)
				return
			}
			if err := w.Close(); err != nil {
				data := gin.H{"is_uploaded": false}
				response := helper.APIResponse("Upload avatar image failed", http.StatusBadRequest, "error", data)

				c.JSON(http.StatusBadRequest, response)
				return
			}

			acl := bucket.Object(next.FileName()).ACL()
			if err := acl.Set(c, storage.AllUsers, storage.RoleReader); err != nil {
				data := gin.H{"is_uploaded": false}
				response := helper.APIResponse("Upload avatar image failed", http.StatusBadRequest, "error", data)

				c.JSON(http.StatusBadRequest, response)
				return
			}
			fileName = next.FileName()
		}
	}
	if !foundImage {
		data := gin.H{"is_uploaded": false}
		response := helper.APIResponse("Upload avatar image failed", http.StatusBadRequest, "error", data)

		c.JSON(http.StatusBadRequest, response)
		return
	}

	//path := fmt.Sprintf("/home/muhammadfahrezam/bwastartup/images/%d-%s", userID, file.Filename)
	//pathName := fmt.Sprintf("%d-%s", userID, fileName)
	//err = h.userService.UploadToCloud(file, userID)
	if err != nil {
		data := gin.H{"is_uploaded": false}
		response := helper.APIResponse("Upload avatar image failed", http.StatusBadRequest, "error", data)

		c.JSON(http.StatusBadRequest, response)
		return
	}

	imageUrl := fmt.Sprintf("https://storage.googleapis.com/donation_alert/%s", fileName)
	_, err = h.userService.SaveAvatar(userID, imageUrl)
	if err != nil {
		data := gin.H{"is_uploaded": false}
		response := helper.APIResponse("Ups Upload avatar image failed", http.StatusBadRequest, "error", data)

		c.JSON(http.StatusBadRequest, response)
		return
	}

	data := gin.H{"is_uploaded": true}
	response := helper.APIResponse("Upload avatar image success", http.StatusOK, "success", data)

	c.JSON(http.StatusOK, response)
}

func (h *userHandler) FetchUser(c *gin.Context) {
	curretUser := c.MustGet("currentUser").(user.User)

	formatter := user.FormatUser(curretUser, "")

	response := helper.APIResponse("Fetch user data Success", http.StatusOK, "success", formatter)

	c.JSON(http.StatusOK, response)
}
