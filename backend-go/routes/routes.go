package routes

import (
	"parking-system-go/controllers"
	"parking-system-go/models"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func SetupRoutes(router *gin.Engine) {
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	api := router.Group("/api")
	{
		parkingController := controllers.NewParkingController(models.DB)
		reservationController := controllers.NewReservationController(models.DB)
		orderController := controllers.NewOrderController(models.DB)
		accessController := controllers.NewAccessControlController(models.DB)
		billingController := controllers.NewBillingController(models.DB)
		deviceController := controllers.NewDeviceController(models.DB)

		api.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"status":  "ok",
				"message": "Parking system is running",
			})
		})

		parking := api.Group("/parking")
		{
			parking.GET("/lots", parkingController.GetParkingLots)
			parking.GET("/lots/:id", parkingController.GetParkingLotByID)
			parking.GET("/spots", parkingController.GetParkingSpots)
			parking.GET("/spots/:id", parkingController.GetParkingSpotByID)
			parking.PUT("/spots/:id/status", parkingController.UpdateParkingSpotStatus)
		}

		reservations := api.Group("/reservations")
		{
			reservations.POST("", reservationController.CreateReservation)
			reservations.GET("", reservationController.GetReservations)
			reservations.GET("/:id", reservationController.GetReservationByID)
			reservations.PUT("/:id/cancel", reservationController.CancelReservation)
		}

		orders := api.Group("/orders")
		{
			orders.POST("", orderController.CreateOrder)
			orders.GET("", orderController.GetOrders)
			orders.GET("/:id", orderController.GetOrderByID)
			orders.POST("/:id/pay", orderController.PayOrder)
		}

		access := api.Group("/access")
		{
			access.POST("/entry", accessController.VehicleEntry)
			access.POST("/exit", accessController.VehicleExit)
			access.GET("/logs", accessController.GetAccessLogs)
		}

		billing := api.Group("/billing")
		{
			billing.POST("/calculate", billingController.CalculateFee)
			billing.GET("/rules", billingController.GetBillingRules)
			billing.PUT("/rules/:id", billingController.UpdateBillingRule)
		}

		devices := api.Group("/devices")
		{
			devices.GET("", deviceController.GetDevices)
			devices.GET("/:id", deviceController.GetDeviceByID)
			devices.POST("/:id/control", deviceController.ControlDevice)
			devices.GET("/:id/status", deviceController.GetDeviceStatus)
		}
	}
}
