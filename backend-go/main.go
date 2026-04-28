package main

import (
	"log"
	"parking-system-go/config"
	"parking-system-go/models"
	"parking-system-go/routes"
	"parking-system-go/utils"

	"github.com/gin-gonic/gin"
)

func main() {
	config.InitConfig()
	utils.InitLogger()
	models.InitDB()
	models.AutoMigrate()

	router := gin.Default()
	routes.SetupRoutes(router)

	port := config.GetConfig().Server.Port
	log.Printf("Server starting on port %s...", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
