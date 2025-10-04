package main

import (
	"time"
)

// ユーザー
type User struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Email     string    `json:"email" gorm:"unique;not null"`
	Password  string    `json:"-" gorm:"not null"` // JSONには含めない
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// カテゴリ
type Category struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	UserID      uint      `json:"userId"`
	Name        string    `json:"name"`
	Type        string    `json:"type"` // income, expense
	Color       string    `json:"color"`
	Icon        string    `json:"icon"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"createdAt"`
}

// 取引記録
type Transaction struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	UserID      uint      `json:"userId"`
	Type        string    `json:"type"` // income, expense
	Amount      float64   `json:"amount"`
	CategoryID  uint      `json:"categoryId"`
	Category    Category  `json:"category" gorm:"foreignKey:CategoryID"`
	Description string    `json:"description"`
	Date        time.Time `json:"date"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// 認証リクエスト
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Name     string `json:"name" binding:"required"`
}

type AuthResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

// 月別集計
type MonthlySummary struct {
	Year         int     `json:"year"`
	Month        int     `json:"month"`
	TotalIncome  float64 `json:"totalIncome"`
	TotalExpense float64 `json:"totalExpense"`
	Balance      float64 `json:"balance"`
}

// カテゴリ別集計
type CategorySummary struct {
	CategoryID    uint    `json:"categoryId"`
	CategoryName  string  `json:"categoryName"`
	CategoryIcon  string  `json:"categoryIcon"`
	CategoryColor string  `json:"categoryColor"`
	Type          string  `json:"type"`
	TotalAmount   float64 `json:"totalAmount"`
	Count         int64   `json:"count"`
}

// 統計情報
type Stats struct {
	TotalIncome      float64 `json:"totalIncome"`
	TotalExpense     float64 `json:"totalExpense"`
	CurrentBalance   float64 `json:"currentBalance"`
	ThisMonthIncome  float64 `json:"thisMonthIncome"`
	ThisMonthExpense float64 `json:"thisMonthExpense"`
	TransactionCount int64   `json:"transactionCount"`
}

// 日別集計
type DailySummary struct {
	Date         string  `json:"date"`
	TotalIncome  float64 `json:"totalIncome"`
	TotalExpense float64 `json:"totalExpense"`
	Balance      float64 `json:"balance"`
}

// 月次予算
type Budget struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	UserID    uint      `json:"userId"`
	Year      int       `json:"year"`
	Month     int       `json:"month"`
	Amount    float64   `json:"amount"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// 固定費
type FixedExpense struct {
	ID             uint       `json:"id" gorm:"primaryKey"`
	UserID         uint       `json:"userId"`
	Name           string     `json:"name"`
	Amount         float64    `json:"amount"`
	Type           string     `json:"type" gorm:"default:expense"` // income, expense
	CategoryID     uint       `json:"categoryId"`
	Category       Category   `json:"category" gorm:"foreignKey:CategoryID"`
	Description    string     `json:"description"`
	IsActive       bool       `json:"isActive" gorm:"default:true"`
	AutoRegister   bool       `json:"autoRegister" gorm:"default:false"`
	RegisterDay    int        `json:"registerDay" gorm:"default:1"`
	LastRegistered *time.Time `json:"lastRegistered,omitempty"`
	CreatedAt      time.Time  `json:"createdAt"`
	UpdatedAt      time.Time  `json:"updatedAt"`
}

// 予算分析結果
type BudgetAnalysis struct {
	Year               int     `json:"year"`
	Month              int     `json:"month"`
	MonthlyBudget      float64 `json:"monthlyBudget"`
	TotalFixedExpenses float64 `json:"totalFixedExpenses"`
	CurrentSpending    float64 `json:"currentSpending"`
	RemainingBudget    float64 `json:"remainingBudget"`
	BudgetUtilization  float64 `json:"budgetUtilization"` // 使用率 (%)
	DaysRemaining      int     `json:"daysRemaining"`
	DailyAverage       float64 `json:"dailyAverage"` // 1日あたり使用可能金額
}

// 支出予測
type SpendingPrediction struct {
	Year            int       `json:"year"`
	Month           int       `json:"month"`
	CurrentSpending float64   `json:"currentSpending"`
	PredictedTotal  float64   `json:"predictedTotal"`
	DailyAverage    float64   `json:"dailyAverage"`
	RemainingDays   int       `json:"remainingDays"`
	Confidence      string    `json:"confidence"`
	Trend           string    `json:"trend"`
	WeeklyPattern   []float64 `json:"weeklyPattern"`
	MonthlyProgress float64   `json:"monthlyProgress"`
}

// 予算履歴
type BudgetHistory struct {
	Year           int     `json:"year"`
	Month          int     `json:"month"`
	Budget         float64 `json:"budget"`
	ActualSpending float64 `json:"actualSpending"`
	FixedExpenses  float64 `json:"fixedExpenses"`
	SavingsRate    float64 `json:"savingsRate"`
	BudgetExceeded bool    `json:"budgetExceeded"`
}

// 予算設定リクエスト
type BudgetRequest struct {
	Year   int     `json:"year" binding:"required"`
	Month  int     `json:"month" binding:"required,min=1,max=12"`
	Amount float64 `json:"amount" binding:"required,min=0"`
}

// 固定費設定リクエスト
type FixedExpenseRequest struct {
	Name        string  `json:"name" binding:"required"`
	Amount      float64 `json:"amount" binding:"required,min=0"`
	Type        string  `json:"type" binding:"required,oneof=income expense"`
	CategoryID  uint    `json:"categoryId" binding:"required"`
	Description string  `json:"description"`
	IsActive    *bool   `json:"isActive,omitempty"`
}

// 固定収支設定リクエスト（固定費と同じ構造）
type FixedTransactionRequest struct {
	Name        string  `json:"name" binding:"required"`
	Amount      float64 `json:"amount" binding:"required,min=0"`
	Type        string  `json:"type" binding:"required,oneof=income expense"`
	CategoryID  uint    `json:"categoryId" binding:"required"`
	Description string  `json:"description"`
	IsActive    *bool   `json:"isActive,omitempty"`
}

// カテゴリ別予算
type CategoryBudget struct {
	ID              uint      `json:"id" gorm:"primaryKey"`
	UserID          uint      `json:"userId"`
	CategoryID      uint      `json:"categoryId"`
	Category        Category  `json:"category" gorm:"foreignKey:CategoryID"`
	Year            int       `json:"year"`
	Month           int       `json:"month"`
	Amount          float64   `json:"amount"`
	Spent           float64   `json:"spent" gorm:"-"`           // 計算フィールド
	Remaining       float64   `json:"remaining" gorm:"-"`       // 計算フィールド
	UtilizationRate float64   `json:"utilizationRate" gorm:"-"` // 計算フィールド
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

// カテゴリ別予算リクエスト
type CategoryBudgetRequest struct {
	CategoryID uint    `json:"categoryId" binding:"required"`
	Year       int     `json:"year" binding:"required"`
	Month      int     `json:"month" binding:"required,min=1,max=12"`
	Amount     float64 `json:"amount" binding:"required,min=0"`
}

// カテゴリ別予算分析
type CategoryBudgetAnalysis struct {
	CategoryID       uint    `json:"categoryId"`
	CategoryName     string  `json:"categoryName"`
	CategoryColor    string  `json:"categoryColor"`
	CategoryIcon     string  `json:"categoryIcon"`
	BudgetAmount     float64 `json:"budgetAmount"`
	SpentAmount      float64 `json:"spentAmount"`
	RemainingAmount  float64 `json:"remainingAmount"`
	UtilizationRate  float64 `json:"utilizationRate"`
	IsOverBudget     bool    `json:"isOverBudget"`
	TransactionCount int64   `json:"transactionCount"`
}
