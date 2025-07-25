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
	ID          uint   `json:"id" gorm:"primaryKey"`
	UserID      uint   `json:"userId"`
	Name        string `json:"name"`
	Type        string `json:"type"` // income, expense
	Color       string `json:"color"`
	Icon        string `json:"icon"`
	Description string `json:"description"`
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
	CategoryID   uint    `json:"categoryId"`
	CategoryName string  `json:"categoryName"`
	CategoryIcon string  `json:"categoryIcon"`
	CategoryColor string `json:"categoryColor"`
	Type         string  `json:"type"`
	TotalAmount  float64 `json:"totalAmount"`
	Count        int64   `json:"count"`
}

// 統計情報
type Stats struct {
	TotalIncome     float64 `json:"totalIncome"`
	TotalExpense    float64 `json:"totalExpense"`
	CurrentBalance  float64 `json:"currentBalance"`
	ThisMonthIncome float64 `json:"thisMonthIncome"`
	ThisMonthExpense float64 `json:"thisMonthExpense"`
	TransactionCount int64  `json:"transactionCount"`
}

// 日別集計
type DailySummary struct {
	Date         string  `json:"date"`
	TotalIncome  float64 `json:"totalIncome"`
	TotalExpense float64 `json:"totalExpense"`
	Balance      float64 `json:"balance"`
}