package main

import (
	"log"
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
	
	// 通常の取引からの集計（固定費から自動生成された取引も含む）
	db.Model(&Transaction{}).Where("user_id = ? AND type = ?", userID, "income").Select("COALESCE(SUM(amount), 0)").Scan(&stats.TotalIncome)
	db.Model(&Transaction{}).Where("user_id = ? AND type = ?", userID, "expense").Select("COALESCE(SUM(amount), 0)").Scan(&stats.TotalExpense)
	stats.CurrentBalance = stats.TotalIncome - stats.TotalExpense
	
	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	endOfMonth := startOfMonth.AddDate(0, 1, 0).Add(-time.Second)
	
	// 今月の取引（固定費から自動生成された取引も含む）
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
		
		// 通常の取引からの集計（固定費から自動生成された取引も含む）
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
	
	// 通常の取引からの集計（固定費から自動生成された取引も含む）
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
	
	// デバッグログ
	log.Printf("Category summaries for user %v, type %s: %d categories", userID, transactionType, len(summaries))
	for _, s := range summaries {
		log.Printf("Category: %s (ID: %d), Amount: %f, Count: %d", s.CategoryName, s.CategoryID, s.TotalAmount, s.Count)
	}
	
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

// 予算関連ハンドラー

// 指定月の予算取得
func getBudget(c *gin.Context) {
	userID, _ := c.Get("userID")
	year, _ := strconv.Atoi(c.Param("year"))
	month, _ := strconv.Atoi(c.Param("month"))
	
	var budget Budget
	if err := db.Where("user_id = ? AND year = ? AND month = ?", userID, year, month).First(&budget).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Budget not found"})
		return
	}
	
	c.JSON(http.StatusOK, budget)
}

// 予算設定
func createBudget(c *gin.Context) {
	userID, _ := c.Get("userID")
	var req BudgetRequest
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data: " + err.Error()})
		return
	}
	
	// 既存の予算があるかチェック
	var existingBudget Budget
	if err := db.Where("user_id = ? AND year = ? AND month = ?", userID, req.Year, req.Month).First(&existingBudget).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Budget already exists for this month"})
		return
	}
	
	budget := Budget{
		UserID: userID.(uint),
		Year:   req.Year,
		Month:  req.Month,
		Amount: req.Amount,
	}
	
	if err := db.Create(&budget).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create budget: " + err.Error()})
		return
	}
	
	c.JSON(http.StatusCreated, budget)
}

// 予算更新
func updateBudget(c *gin.Context) {
	userID, _ := c.Get("userID")
	id := c.Param("id")
	var budget Budget
	
	if err := db.Where("user_id = ?", userID).First(&budget, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Budget not found"})
		return
	}
	
	var req BudgetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data: " + err.Error()})
		return
	}
	
	budget.Year = req.Year
	budget.Month = req.Month
	budget.Amount = req.Amount
	
	if err := db.Save(&budget).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update budget: " + err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, budget)
}

// 予算削除
func deleteBudget(c *gin.Context) {
	userID, _ := c.Get("userID")
	id := c.Param("id")
	
	if err := db.Where("user_id = ?", userID).Delete(&Budget{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete budget: " + err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "Budget deleted successfully"})
}

// 固定費関連ハンドラー

// 固定費一覧取得
func getFixedExpenses(c *gin.Context) {
	userID, _ := c.Get("userID")
	var fixedExpenses []FixedExpense
	
	if err := db.Preload("Category").Where("user_id = ?", userID).Order("name ASC").Find(&fixedExpenses).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch fixed expenses: " + err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, fixedExpenses)
}

// 固定費追加
func createFixedExpense(c *gin.Context) {
	userID, _ := c.Get("userID")
	var req FixedExpenseRequest
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data: " + err.Error()})
		return
	}
	
	// デバッグログ
	log.Printf("Creating fixed expense - Name: %s, Amount: %f, Type: %s, CategoryID: %v", 
		req.Name, req.Amount, req.Type, req.CategoryID)
	
	fixedExpense := FixedExpense{
		UserID:      userID.(uint),
		Name:        req.Name,
		Amount:      req.Amount,
		Type:        req.Type,
		CategoryID:  req.CategoryID,
		Description: req.Description,
		IsActive:    true,
	}
	
	// IsActiveが明示的に設定されている場合は使用
	if req.IsActive != nil {
		fixedExpense.IsActive = *req.IsActive
	}
	
	if err := db.Create(&fixedExpense).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create fixed expense: " + err.Error()})
		return
	}
	
	// 固定費作成時に今月の取引を生成（重複チェック付き）
	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	endOfMonth := startOfMonth.AddDate(0, 1, 0).Add(-time.Second)
	
	description := "固定収支: " + fixedExpense.Name
	if fixedExpense.Type == "income" {
		description = "固定収入: " + fixedExpense.Name
	} else {
		description = "固定支出: " + fixedExpense.Name
	}
	
	// 今月に同じカテゴリ・同じ金額・同じタイプの取引が既に存在するかチェック
	var existingCount int64
	db.Model(&Transaction{}).Where(
		"user_id = ? AND category_id = ? AND type = ? AND amount = ? AND date BETWEEN ? AND ?",
		userID, fixedExpense.CategoryID, fixedExpense.Type, fixedExpense.Amount, startOfMonth, endOfMonth,
	).Count(&existingCount)
	
	// 既存の取引がない場合のみ新しい取引を生成
	if existingCount == 0 {
		transaction := Transaction{
			UserID:      userID.(uint),
			Type:        fixedExpense.Type,
			Amount:      fixedExpense.Amount,
			CategoryID:  fixedExpense.CategoryID,
			Description: description,
			Date:        startOfMonth,
		}
		
		if err := db.Create(&transaction).Error; err != nil {
			log.Printf("Failed to create transaction for fixed expense: %v", err)
		} else {
			log.Printf("Created transaction for fixed expense: %s, amount: %f", fixedExpense.Name, fixedExpense.Amount)
		}
	} else {
		log.Printf("Skipping transaction creation for fixed expense: %s (similar transaction already exists)", fixedExpense.Name)
	}
	
	// カテゴリ情報を含めて返す
	db.Preload("Category").First(&fixedExpense, fixedExpense.ID)
	c.JSON(http.StatusCreated, fixedExpense)
}

// 固定費更新
func updateFixedExpense(c *gin.Context) {
	userID, _ := c.Get("userID")
	id := c.Param("id")
	var fixedExpense FixedExpense
	
	if err := db.Where("user_id = ?", userID).First(&fixedExpense, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Fixed expense not found"})
		return
	}
	
	var req FixedExpenseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data: " + err.Error()})
		return
	}
	
	fixedExpense.Name = req.Name
	fixedExpense.Amount = req.Amount
	fixedExpense.Type = req.Type
	fixedExpense.CategoryID = req.CategoryID
	fixedExpense.Description = req.Description
	
	if req.IsActive != nil {
		fixedExpense.IsActive = *req.IsActive
	}
	
	if err := db.Save(&fixedExpense).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update fixed expense: " + err.Error()})
		return
	}
	
	// カテゴリ情報を含めて返す
	db.Preload("Category").First(&fixedExpense, fixedExpense.ID)
	c.JSON(http.StatusOK, fixedExpense)
}

// 固定費削除
func deleteFixedExpense(c *gin.Context) {
	userID, _ := c.Get("userID")
	id := c.Param("id")
	
	// 固定費が存在するかチェック
	var fixedExpense FixedExpense
	if err := db.Where("user_id = ?", userID).First(&fixedExpense, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Fixed expense not found"})
		return
	}
	
	// 固定費から自動生成された取引を削除
	var description string
	if fixedExpense.Type == "income" {
		description = "固定収入: " + fixedExpense.Name
	} else {
		description = "固定支出: " + fixedExpense.Name
	}
	
	// 関連する自動生成取引を削除
	if err := db.Where("user_id = ? AND category_id = ? AND type = ? AND description = ?", 
		userID, fixedExpense.CategoryID, fixedExpense.Type, description).Delete(&Transaction{}).Error; err != nil {
		log.Printf("Failed to delete related transactions for fixed expense %d: %v", fixedExpense.ID, err)
	}
	
	// 固定費を削除
	if err := db.Delete(&fixedExpense).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete fixed expense: " + err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "Fixed expense deleted successfully"})
}



// 予算分析関連ハンドラー

// 月次予算分析
func getBudgetAnalysis(c *gin.Context) {
	userID, _ := c.Get("userID")
	year, _ := strconv.Atoi(c.Param("year"))
	month, _ := strconv.Atoi(c.Param("month"))
	
	// 予算取得（存在しない場合はカテゴリ別予算の合計を使用）
	var budget Budget
	budgetAmount := float64(0)
	if err := db.Where("user_id = ? AND year = ? AND month = ?", userID, year, month).First(&budget).Error; err == nil {
		budgetAmount = budget.Amount
	}
	
	// 月次予算が設定されていない場合、カテゴリ別予算の合計を使用
	if budgetAmount == 0 {
		var categoryBudgetTotal float64
		db.Model(&CategoryBudget{}).Where("user_id = ? AND year = ? AND month = ?", userID, year, month).Select("COALESCE(SUM(amount), 0)").Scan(&categoryBudgetTotal)
		budgetAmount = categoryBudgetTotal
	}
	
	// 当月の支出取得（固定費から自動生成された取引も含む）
	startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 1, 0).Add(-time.Second)
	
	var currentSpending float64
	db.Model(&Transaction{}).Where("user_id = ? AND type = ? AND date BETWEEN ? AND ?", userID, "expense", startDate, endDate).Select("COALESCE(SUM(amount), 0)").Scan(&currentSpending)
	
	// 固定支出合計取得（表示用）- 固定収入は含めない
	var totalFixedExpenses float64
	db.Model(&FixedExpense{}).Where("user_id = ? AND type = ? AND is_active = ?", userID, "expense", true).Select("COALESCE(SUM(amount), 0)").Scan(&totalFixedExpenses)
	
	// 残り予算計算（固定費は既にcurrentSpendingに含まれているので重複計算しない）
	remainingBudget := budgetAmount - currentSpending
	
	// 予算使用率計算
	budgetUtilization := float64(0)
	if budgetAmount > 0 {
		budgetUtilization = (currentSpending / budgetAmount) * 100
	}
	
	// 残り日数計算
	now := time.Now()
	var daysRemaining int
	if now.Year() == year && int(now.Month()) == month {
		lastDayOfMonth := startDate.AddDate(0, 1, 0).Add(-24 * time.Hour)
		daysRemaining = int(lastDayOfMonth.Sub(now).Hours()/24) + 1
		if daysRemaining < 0 {
			daysRemaining = 0
		}
	} else {
		daysRemaining = int(endDate.Sub(startDate).Hours()/24) + 1
	}
	
	// 1日あたり使用可能金額計算
	dailyAverage := float64(0)
	if daysRemaining > 0 && remainingBudget > 0 {
		dailyAverage = remainingBudget / float64(daysRemaining)
	}
	
	analysis := BudgetAnalysis{
		Year:              year,
		Month:             month,
		MonthlyBudget:     budgetAmount,
		TotalFixedExpenses: totalFixedExpenses,
		CurrentSpending:   currentSpending,
		RemainingBudget:   remainingBudget,
		BudgetUtilization: budgetUtilization,
		DaysRemaining:     daysRemaining,
		DailyAverage:      dailyAverage,
	}
	
	c.JSON(http.StatusOK, analysis)
}

// 残り予算取得
func getRemainingBudget(c *gin.Context) {
	userID, _ := c.Get("userID")
	year, _ := strconv.Atoi(c.Param("year"))
	month, _ := strconv.Atoi(c.Param("month"))
	
	// 予算取得
	var budget Budget
	if err := db.Where("user_id = ? AND year = ? AND month = ?", userID, year, month).First(&budget).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Budget not found for this month"})
		return
	}
	
	// 固定支出合計取得（固定収入は含めない）
	var totalFixedExpenses float64
	db.Model(&FixedExpense{}).Where("user_id = ? AND type = ? AND is_active = ?", userID, "expense", true).Select("COALESCE(SUM(amount), 0)").Scan(&totalFixedExpenses)
	
	// 当月の支出取得
	startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 1, 0).Add(-time.Second)
	
	var currentSpending float64
	db.Model(&Transaction{}).Where("user_id = ? AND type = ? AND date BETWEEN ? AND ?", userID, "expense", startDate, endDate).Select("COALESCE(SUM(amount), 0)").Scan(&currentSpending)
	
	// 残り予算計算
	remainingBudget := budget.Amount - totalFixedExpenses - currentSpending
	
	c.JSON(http.StatusOK, gin.H{
		"remainingBudget": remainingBudget,
		"monthlyBudget":   budget.Amount,
		"fixedExpenses":   totalFixedExpenses,
		"currentSpending": currentSpending,
	})
}

// 予算履歴取得（過去6ヶ月）
func getBudgetHistory(c *gin.Context) {
	userID, _ := c.Get("userID")
	
	var history []BudgetHistory
	now := time.Now()
	
	// 過去6ヶ月のデータを取得
	for i := 5; i >= 0; i-- {
		targetDate := now.AddDate(0, -i, 0)
		year := targetDate.Year()
		month := int(targetDate.Month())
		
		var budget Budget
		var budgetAmount float64 = 0
		if err := db.Where("user_id = ? AND year = ? AND month = ?", userID, year, month).First(&budget).Error; err == nil {
			budgetAmount = budget.Amount
		}
		
		// 固定支出合計取得（固定収入は含めない）
		var fixedExpenses float64
		db.Model(&FixedExpense{}).Where("user_id = ? AND type = ? AND is_active = ?", userID, "expense", true).Select("COALESCE(SUM(amount), 0)").Scan(&fixedExpenses)
		
		// 実際の支出取得
		startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
		endDate := startDate.AddDate(0, 1, 0).Add(-time.Second)
		
		var actualSpending float64
		db.Model(&Transaction{}).Where("user_id = ? AND type = ? AND date BETWEEN ? AND ?", userID, "expense", startDate, endDate).Select("COALESCE(SUM(amount), 0)").Scan(&actualSpending)
		
		// 貯蓄率計算
		savingsRate := float64(0)
		if budgetAmount > 0 {
			savingsRate = ((budgetAmount - actualSpending) / budgetAmount) * 100
		}
		
		// 予算超過チェック
		budgetExceeded := actualSpending > budgetAmount && budgetAmount > 0
		
		historyItem := BudgetHistory{
			Year:           year,
			Month:          month,
			Budget:         budgetAmount,
			ActualSpending: actualSpending,
			FixedExpenses:  fixedExpenses,
			SavingsRate:    savingsRate,
			BudgetExceeded: budgetExceeded,
		}
		
		history = append(history, historyItem)
	}
	
	c.JSON(http.StatusOK, history)
}

// カテゴリ別予算関連ハンドラー

// カテゴリ別予算一覧取得
func getCategoryBudgets(c *gin.Context) {
	userID, _ := c.Get("userID")
	year, _ := strconv.Atoi(c.Param("year"))
	month, _ := strconv.Atoi(c.Param("month"))
	
	var categoryBudgets []CategoryBudget
	if err := db.Preload("Category").Where("user_id = ? AND year = ? AND month = ?", userID, year, month).Find(&categoryBudgets).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch category budgets: " + err.Error()})
		return
	}
	
	// 各カテゴリ別予算の使用状況を計算
	startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 1, 0).Add(-time.Second)
	
	for i := range categoryBudgets {
		var spent float64
		db.Model(&Transaction{}).Where("user_id = ? AND category_id = ? AND type = ? AND date BETWEEN ? AND ?", 
			userID, categoryBudgets[i].CategoryID, "expense", startDate, endDate).
			Select("COALESCE(SUM(amount), 0)").Scan(&spent)
		
		categoryBudgets[i].Spent = spent
		categoryBudgets[i].Remaining = categoryBudgets[i].Amount - spent
		
		if categoryBudgets[i].Amount > 0 {
			categoryBudgets[i].UtilizationRate = (spent / categoryBudgets[i].Amount) * 100
		}
	}
	
	c.JSON(http.StatusOK, categoryBudgets)
}

// カテゴリ別予算作成
func createCategoryBudget(c *gin.Context) {
	userID, _ := c.Get("userID")
	var req CategoryBudgetRequest
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data: " + err.Error()})
		return
	}
	
	// 既存の予算があるかチェック
	var existingBudget CategoryBudget
	if err := db.Where("user_id = ? AND category_id = ? AND year = ? AND month = ?", 
		userID, req.CategoryID, req.Year, req.Month).First(&existingBudget).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Budget already exists for this category and month"})
		return
	}
	
	categoryBudget := CategoryBudget{
		UserID:     userID.(uint),
		CategoryID: req.CategoryID,
		Year:       req.Year,
		Month:      req.Month,
		Amount:     req.Amount,
	}
	
	if err := db.Create(&categoryBudget).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create category budget: " + err.Error()})
		return
	}
	
	// カテゴリ情報を含めて返す
	db.Preload("Category").First(&categoryBudget, categoryBudget.ID)
	c.JSON(http.StatusCreated, categoryBudget)
}

// カテゴリ別予算更新
func updateCategoryBudget(c *gin.Context) {
	userID, _ := c.Get("userID")
	id := c.Param("id")
	var categoryBudget CategoryBudget
	
	if err := db.Where("user_id = ?", userID).First(&categoryBudget, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Category budget not found"})
		return
	}
	
	var req CategoryBudgetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data: " + err.Error()})
		return
	}
	
	categoryBudget.CategoryID = req.CategoryID
	categoryBudget.Year = req.Year
	categoryBudget.Month = req.Month
	categoryBudget.Amount = req.Amount
	
	if err := db.Save(&categoryBudget).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update category budget: " + err.Error()})
		return
	}
	
	// カテゴリ情報を含めて返す
	db.Preload("Category").First(&categoryBudget, categoryBudget.ID)
	c.JSON(http.StatusOK, categoryBudget)
}

// カテゴリ別予算削除
func deleteCategoryBudget(c *gin.Context) {
	userID, _ := c.Get("userID")
	id := c.Param("id")
	
	if err := db.Where("user_id = ?", userID).Delete(&CategoryBudget{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete category budget: " + err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "Category budget deleted successfully"})
}

// カテゴリ別予算分析
func getCategoryBudgetAnalysis(c *gin.Context) {
	userID, _ := c.Get("userID")
	year, _ := strconv.Atoi(c.Param("year"))
	month, _ := strconv.Atoi(c.Param("month"))
	
	// カテゴリ別予算を取得
	var categoryBudgets []CategoryBudget
	db.Preload("Category").Where("user_id = ? AND year = ? AND month = ?", userID, year, month).Find(&categoryBudgets)
	
	// 予算が設定されていない場合は空の配列を返す
	if len(categoryBudgets) == 0 {
		c.JSON(http.StatusOK, []CategoryBudgetAnalysis{})
		return
	}
	
	var analysis []CategoryBudgetAnalysis
	startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 1, 0).Add(-time.Second)
	
	for _, budget := range categoryBudgets {
		var spentAmount float64
		var transactionCount int64
		
		db.Model(&Transaction{}).Where("user_id = ? AND category_id = ? AND type = ? AND date BETWEEN ? AND ?", 
			userID, budget.CategoryID, "expense", startDate, endDate).
			Select("COALESCE(SUM(amount), 0)").Scan(&spentAmount)
		
		db.Model(&Transaction{}).Where("user_id = ? AND category_id = ? AND type = ? AND date BETWEEN ? AND ?", 
			userID, budget.CategoryID, "expense", startDate, endDate).Count(&transactionCount)
		
		remainingAmount := budget.Amount - spentAmount
		utilizationRate := float64(0)
		if budget.Amount > 0 {
			utilizationRate = (spentAmount / budget.Amount) * 100
		}
		
		analysisItem := CategoryBudgetAnalysis{
			CategoryID:       budget.CategoryID,
			CategoryName:     budget.Category.Name,
			CategoryColor:    budget.Category.Color,
			CategoryIcon:     budget.Category.Icon,
			BudgetAmount:     budget.Amount,
			SpentAmount:      spentAmount,
			RemainingAmount:  remainingAmount,
			UtilizationRate:  utilizationRate,
			IsOverBudget:     spentAmount > budget.Amount,
			TransactionCount: transactionCount,
		}
		
		analysis = append(analysis, analysisItem)
	}
	
	c.JSON(http.StatusOK, analysis)
}