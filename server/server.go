package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"net/http"
)

const (
	httpPort = ":8888"
)

func main() {
	r := setupRouter()
	r.Run(httpPort)
}

func setupRouter() *gin.Engine {
	api := &Api{store: NewInMem()}

	r := gin.Default()

	r.POST("/user", api.Signup)

	r.GET("/user/:name", api.GetUser)

	authorized := r.Group("/message")
	authorized.Use(api.AuthMiddleware)
	{
		authorized.GET("", api.GetMessages)
		authorized.POST("", api.SendMessage)
	}

	return r
}

type Api struct {
	store
}

func (api *Api) Signup(c *gin.Context) {

	var user struct {
		Name   string `json:"name" binding:"required"`
		PubKey []byte `json:"pubKey" binding:"required"`
	}
	err := c.Bind(&user)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	if _, ok := api.GetUserPubKey(user.Name); ok {
		c.JSON(http.StatusBadRequest, gin.H{"message": "user already exists"})
		return
	}

	api.SetUserPubKey(user.Name, user.PubKey)
	apiKey := uuid.New().String()
	api.SetApiKeyUser(apiKey, user.Name)

	c.JSON(http.StatusOK, gin.H{"message": "registered user", "api_key": apiKey})

}
func (api *Api) AuthMiddleware(c *gin.Context) {
	APIKey := c.Request.Header.Get("X-API-Key")
	user, ok := api.Authenticate(APIKey)
	if !ok {
		c.JSON(http.StatusForbidden, gin.H{"message": "invalid api key"})
		return
	}
	c.Set("user", user)

	c.Next()
}

func (api *Api) GetUser(c *gin.Context) {
	userName := c.Params.ByName("name")
	pubKey, ok := api.GetUserPubKey(userName)
	if ok {
		c.JSON(http.StatusOK, gin.H{"pubKey": pubKey})
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"message": "user not found"})
	}
}

func (api *Api) SendMessage(c *gin.Context) {
	fromUser, _ := c.Get("user")

	var message struct {
		To  string `json:"to" binding:"required"`
		Msg []byte `json:"msg" binding:"required"`
	}
	err := c.Bind(&message)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	if _, ok := api.GetUserPubKey(message.To); !ok {
		c.JSON(http.StatusBadRequest, gin.H{"message": "receiver user does not exist"})
		return
	}

	api.AddMessage(fromUser.(string), SenderMessage{
		To:      message.To,
		Message: message.Msg,
	})

	c.JSON(http.StatusBadRequest, gin.H{"message": "message sent"})
}

func (api *Api) GetMessages(c *gin.Context) {
	toUser, _ := c.Get("user")

	msgs := api.GetUserMessages(toUser.(string))
	fmt.Println(msgs)
	c.JSON(http.StatusOK, gin.H{"messages": msgs})
}
