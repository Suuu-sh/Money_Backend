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

			// 予算関連
			protected.GET("budget/:year/:month", getBudget)
			protected.POST("budget", createBudget)
			protected.PUT("budget/:id", updateBudget)
			protected.DELETE("budget/:id", deleteBudget)

			// 固定費関連
			protected.GET("fixed-expenses", getFixedExpenses)
			protected.POST("fixed-expenses", createFixedExpense)
			protected.PUT("fixed-expenses/:id", updateFixedExpense)
			protected.DELETE("fixed-expenses/:id", deleteFixedExpense)

			// 予算分析関連
			protected.GET("budget/analysis/:year/:month", getBudgetAnalysis)
			protected.GET("budget/remaining/:year/:month", getRemainingBudget)
			protected.GET("budget/history", getBudgetHistory)
			protected.GET("budget/monthly-report/:year/:month", getMonthlyBudgetReport)
			protected.POST("budget/continue/:year/:month", continueBudgetSettings)
			
			// 固定収支の月次処理
			protected.POST("fixed-expenses/process-monthly", processMonthlyFixedTransactionsHandler)

			// カテゴリ別予算関連
			protected.GET("category-budgets/:year/:month", getCategoryBudgets)
			protected.POST("category-budgets", createCategoryBudget)
			protected.PUT("category-budgets/:id", updateCategoryBudget)
			protected.DELETE("category-budgets/:id", deleteCategoryBudget)
			protected.GET("category-budgets/analysis/:year/:month", getCategoryBudgetAnalysis)
		}
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("MoneyTracker Server starting on port %s", port)
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
	db.AutoMigrate(&User{}, &Category{}, &Transaction{}, &Budget{}, &FixedExpense{}, &CategoryBudget{})

	// 初期データ投入
	seedData()
}