package main

import (
	"log"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var db *gorm.DB

func main() {
	// データベース初期化
	initDB()

	// Ginルーター設定
	r := gin.Default()

	// CORS設定
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
	}))

	// API routes
	api := r.Group("/api")
	{
		// 認証関連
		api.POST("/register", register)
		api.POST("/login", login)
		api.POST("/logout", logout)
		api.GET("/me", authMiddleware(), getCurrentUser)

		// 認証が必要なルート
		protected := api.Group("/")
		protected.Use(authMiddleware())
		{
			// 取引関連
			protected.GET("transactions", getTransactions)
			protected.POST("transactions", createTransaction)
			protected.PUT("transactions/:id", updateTransaction)
			protected.DELETE("transactions/:id", deleteTransaction)
			protected.GET("transactions/:id", getTransaction)

			// カテゴリ関連
			protected.GET("categories", getCategories)
			protected.POST("categories", createCategory)
			protected.PUT("categories/:id", updateCategory)
			protected.DELETE("categories/:id", deleteCategory)

			// 統計・集計
			protected.GET("stats", getStats)
			protected.GET("summary/monthly", getMonthlySummary)
			protected.GET("summary/category", getCategorySummary)
			protected.GET("summary/daily", getDailySummary)
		}
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("💰 MoneyTracker Server starting on port %s", port)
	r.Run(":" + port)
}

func initDB() {
	var err error

	// 開発環境ではSQLite
	db, err = gorm.Open(sqlite.Open("moneytracker.db"), &gorm.Config{})

	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// マイグレーション
	db.AutoMigrate(&User{}, &Category{}, &Transaction{})

	// 初期データ投入
	seedData()
}