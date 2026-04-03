package main

import (
	"fmt"
	"main/src/auth"
	"main/src/common"
	"main/src/movie"
	"main/src/ticket"
	"main/src/user"

	"github.com/gin-gonic/gin"
)

func main() {
	common.LoadEnv()

	db := common.ConnectMongoDB()
	_ = db

	r := gin.Default()

	userRepo := user.NewRepository(db)
	userCtrl := user.NewController(userRepo)
	authCtrl := auth.NewController(userRepo)
	movieRepo := movie.NewRepository(db)
	movieCtrl := movie.NewController(movieRepo, userRepo)
	ticketRepo := ticket.NewRepository(db)
	ticketCtrl := ticket.NewController(ticketRepo, movieRepo, userRepo)

	port := common.GetEnv("PORT")
	fmt.Println("Server is running at http://" + port)

	r.POST("/api/register", userCtrl.Register)
	r.POST("/api/login", authCtrl.Login)

	admin := r.Group("/api")
	admin.Use(auth.JWTMiddleware(), auth.AdminMiddleware(userRepo))
	{
		admin.GET("/movies", movieCtrl.GetAll)
		admin.POST("/movies", movieCtrl.Create)
		admin.PUT("/movies/:id", movieCtrl.Update)
		admin.DELETE("/movies/:id", movieCtrl.Delete)
	}

	customer := r.Group("/api")
	customer.Use(auth.JWTMiddleware(), auth.CustomerMiddleware(userRepo))
	{
		customer.POST("/tickets", ticketCtrl.Create)
		customer.GET("/tickets", ticketCtrl.GetByUser)
	}

	r.GET("/api/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})
	r.Run(port)
}
