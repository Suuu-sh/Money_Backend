package main

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// 取引一覧取得
func getTransactions(c *gin.Context) {
	userID, _ := c.Get("userID")
	var transactions []Transaction
	
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	transactionType := c.Query("type")
	categoryId := c.Query("categoryId")
	startDate := c.Query("startDate")
	endDate := c.Query("endDate")
	
	query := db.Preload("Category").Where("user_id = ?", userID).Order("date DESC, created_at DESC")
	
	if transactionType != "" {
		query = query.Where("type = ?", transactionType)
	}
	if categoryId != "" {
		query = query.Where("category_id = ?", categoryId)
	}
	if startDate != "" {
		query = query.Where("date >= ?", startDate)
	}
	if endDate != "" {
		query = query.Where("date <= ?", endDate)
	}
	
	offset := (page - 1) * limit
	query.Offset(offset).Limit(limit).Find(&transactions)
	
	c.JSON(http.StatusOK, transactions)
}

func createTransaction(c *gin.Context) {
	userID, _ := c.Get("userID")
	
	// リクエストデータを受け取るための構造体
	var req struct {
		Type        string  `json:"type" binding:"required"`
		Amount      float64 `json:"amount" binding:"required"`
		CategoryID  uint    `json:"categoryId" binding:"required"`
		Description string  `json:"description"`
		Date        string  `json:"date" binding:"required"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data: " + err.Error()})
		return
	}

	// 日付文字列をtime.Timeに変換
	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format: " + err.Error()})
		return
	}

	// Transactionオブジェクトを作成
	transaction := Transaction{
		UserID:      userID.(uint),
		Type:        req.Type,
		Amount:      req.Amount,
		CategoryID:  req.CategoryID,
		Description: req.Description,
		Date:        date,
	}

	if err := db.Create(&transaction).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create transaction: " + err.Error()})
		return
	}

	// カテゴリ情報を含めて返す
	db.Preload("Category").First(&transaction, transaction.ID)
	c.JSON(http.StatusCreated, transaction)
}

func updateTransaction(c *gin.Context) {
	userID, _ := c.Get("userID")
	id := c.Param("id")
	var transaction Transaction
	
	if err := db.Where("user_id = ?", userID).First(&transaction, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Transaction not found"})
		return
	}

	if err := c.ShouldBindJSON(&transaction); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db.Save(&transaction)
	db.Preload("Category").First(&transaction, transaction.ID)
	c.JSON(http.StatusOK, transaction)
}

func deleteTransaction(c *gin.Context) {
	userID, _ := c.Get("userID")
	id := c.Param("id")
	
	if err := db.Where("user_id = ?", userID).Delete(&Transaction{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Transaction deleted successfully"})
}

func getTransaction(c *gin.Context) {
	userID, _ := c.Get("userID")
	id := c.Param("id")
	var transaction Transaction
	
	if err := db.Preload("Category").Where("user_id = ?", userID).First(&transaction, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Transaction not found"})
		return
	}

	c.JSON(http.StatusOK, transaction)
}

func getCategories(c *gin.Context) {
	userID, _ := c.Get("userID")
	var categories []Category
	transactionType := c.Query("type")
	
	query := db.Where("user_id = ?", userID).Order("type ASC, name ASC")
	if transactionType != "" {
		query = query.Where("type = ?", transactionType)
	}
	
	query.Find(&categories)
	c.JSON(http.StatusOK, categories)
}

func createCategory(c *gin.Context) {
	userID, _ := c.Get("userID")
	var category Category
	if err := c.ShouldBindJSON(&category); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	category.UserID = userID.(uint)

	if err := db.Create(&category).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, category)
}

func updateCategory(c *gin.Context) {
	userID, _ := c.Get("userID")
	id := c.Param("id")
	var category Category
	
	if err := db.Where("user_id = ?", userID).First(&category, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Category not found"})
		return
	}

	if err := c.ShouldBindJSON(&category); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db.Save(&category)
	c.JSON(http.StatusOK, category)
}

func deleteCategory(c *gin.Context) {
	userID, _ := c.Get("userID")
	id := c.Param("id")
	
	var count int64
	db.Model(&Transaction{}).Where("user_id = ? AND category_id = ?", userID, id).Count(&count)
	if count > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot delete category with existing transactions"})
		return
	}
	
	if err := db.Where("user_id = ?", userID).Delete(&Category{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Category deleted successfully"})
}

func getStats(c *gin.Context) {
	userID, _ := c.Get("userID")
	var stats Stats
	
	db.Model(&Transaction{}).Where("user_id = ? AND type = ?", userID, "income").Select("COALESCE(SUM(amount), 0)").Scan(&stats.TotalIncome)
	db.Model(&Transaction{}).Where("user_id = ? AND type = ?", userID, "expense").Select("COALESCE(SUM(amount), 0)").Scan(&stats.TotalExpense)
	stats.CurrentBalance = stats.TotalIncome - stats.TotalExpense
	
	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	endOfMonth := startOfMonth.AddDate(0, 1, 0).Add(-time.Second)
	
	db.Model(&Transaction{}).Where("user_id = ? AND type = ? AND date BETWEEN ? AND ?", userID, "income", startOfMonth, endOfMonth).Select("COALESCE(SUM(amount), 0)").Scan(&stats.ThisMonthIncome)
	db.Model(&Transaction{}).Where("user_id = ? AND type = ? AND date BETWEEN ? AND ?", userID, "expense", startOfMonth, endOfMonth).Select("COALESCE(SUM(amount), 0)").Scan(&stats.ThisMonthExpense)
	
	db.Model(&Transaction{}).Where("user_id = ?", userID).Count(&stats.TransactionCount)
	
	c.JSON(http.StatusOK, stats)
}

func getMonthlySummary(c *gin.Context) {
	userID, _ := c.Get("userID")
	year, _ := strconv.Atoi(c.DefaultQuery("year", strconv.Itoa(time.Now().Year())))
	
	var summaries []MonthlySummary
	
	for month := 1; month <= 12; month++ {
		startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
		endDate := startDate.AddDate(0, 1, 0).Add(-time.Second)
		
		var summary MonthlySummary
		summary.Year = year
		summary.Month = month
		
		db.Model(&Transaction{}).Where("user_id = ? AND type = ? AND date BETWEEN ? AND ?", userID, "income", startDate, endDate).Select("COALESCE(SUM(amount), 0)").Scan(&summary.TotalIncome)
		db.Model(&Transaction{}).Where("user_id = ? AND type = ? AND date BETWEEN ? AND ?", userID, "expense", startDate, endDate).Select("COALESCE(SUM(amount), 0)").Scan(&summary.TotalExpense)
		summary.Balance = summary.TotalIncome - summary.TotalExpense
		
		summaries = append(summaries, summary)
	}
	
	c.JSON(http.StatusOK, summaries)
}

func getCategorySummary(c *gin.Context) {
	userID, _ := c.Get("userID")
	transactionType := c.DefaultQuery("type", "expense")
	startDate := c.Query("startDate")
	endDate := c.Query("endDate")
	
	query := `
		SELECT 
			c.id as category_id,
			c.name as category_name,
			c.icon as category_icon,
			c.color as category_color,
			c.type as type,
			COALESCE(SUM(t.amount), 0) as total_amount,
			COUNT(t.id) as count
		FROM categories c
		LEFT JOIN transactions t ON c.id = t.category_id AND t.type = ? AND t.user_id = ?
	`
	
	args := []interface{}{transactionType, userID}
	
	if startDate != "" && endDate != "" {
		query += " AND t.date BETWEEN ? AND ?"
		args = append(args, startDate, endDate)
	}
	
	query += " WHERE c.user_id = ? AND c.type = ? GROUP BY c.id ORDER BY total_amount DESC"
	args = append(args, userID, transactionType)
	
	var summaries []CategorySummary
	db.Raw(query, args...).Scan(&summaries)
	
	c.JSON(http.StatusOK, summaries)
}

func getDailySummary(c *gin.Context) {
	userID, _ := c.Get("userID")
	startDate := c.DefaultQuery("startDate", time.Now().AddDate(0, 0, -30).Format("2006-01-02"))
	endDate := c.DefaultQuery("endDate", time.Now().Format("2006-01-02"))
	
	query := `
		SELECT 
			DATE(date) as date,
			COALESCE(SUM(CASE WHEN type = 'income' THEN amount ELSE 0 END), 0) as total_income,
			COALESCE(SUM(CASE WHEN type = 'expense' THEN amount ELSE 0 END), 0) as total_expense,
			COALESCE(SUM(CASE WHEN type = 'income' THEN amount ELSE -amount END), 0) as balance
		FROM transactions 
		WHERE user_id = ? AND date BETWEEN ? AND ?
		GROUP BY DATE(date)
		ORDER BY date DESC
	`
	
	var summaries []DailySummary
	db.Raw(query, userID, startDate, endDate).Scan(&summaries)
	
	c.JSON(http.StatusOK, summaries)
}