package main

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var jwtSecret = []byte("your-secret-key-change-in-production")

type Claims struct {
	UserID uint   `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

// パスワードハッシュ化
func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

// パスワード検証
func checkPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// JWTトークン生成
func generateToken(userID uint, email string) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// ユーザー登録
func register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 既存ユーザーチェック
	var existingUser User
	if err := db.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "このメールアドレスは既に登録されています"})
		return
	}

	// パスワードハッシュ化
	hashedPassword, err := hashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "パスワードの処理に失敗しました"})
		return
	}

	// ユーザー作成
	user := User{
		Email:    req.Email,
		Password: hashedPassword,
		Name:     req.Name,
	}

	if err := db.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ユーザーの作成に失敗しました"})
		return
	}

	// デフォルトカテゴリ作成
	ensureUserHasDefaultCategories(user.ID)

	// トークン生成
	token, err := generateToken(user.ID, user.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "トークンの生成に失敗しました"})
		return
	}

	c.JSON(http.StatusCreated, AuthResponse{
		Token: token,
		User:  user,
	})
}

// ログイン
func login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// ユーザー検索
	var user User
	if err := db.Where("email = ?", req.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "メールアドレスまたはパスワードが間違っています"})
		return
	}

	// パスワード検証
	if !checkPasswordHash(req.Password, user.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "メールアドレスまたはパスワードが間違っています"})
		return
	}

	// トークン生成
	token, err := generateToken(user.ID, user.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "トークンの生成に失敗しました"})
		return
	}

	c.JSON(http.StatusOK, AuthResponse{
		Token: token,
		User:  user,
	})
}

// ログアウト
func logout(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "ログアウトしました"})
}

// 現在のユーザー取得
func getCurrentUser(c *gin.Context) {
	userID, _ := c.Get("userID")
	
	var user User
	if err := db.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "ユーザーが見つかりません"})
		return
	}

	c.JSON(http.StatusOK, user)
}

// 認証ミドルウェア
func authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "認証トークンが必要です"})
			c.Abort()
			return
		}

		tokenString := strings.Replace(authHeader, "Bearer ", "", 1)
		
		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return jwtSecret, nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "無効なトークンです"})
			c.Abort()
			return
		}

		c.Set("userID", claims.UserID)
		c.Set("userEmail", claims.Email)
		c.Next()
	}
}

// 特定のユーザーに不足しているカテゴリを追加
func ensureUserHasDefaultCategories(userID uint) {
	// 必要なカテゴリのリスト
	requiredCategories := []Category{
		// 収入カテゴリ
		{UserID: userID, Name: "給与", Type: "income", Color: "#10B981", Icon: "briefcase", Description: "会社からの給与"},
		{UserID: userID, Name: "賞与", Type: "income", Color: "#F59E0B", Icon: "gift", Description: "ボーナス・賞与"},
		{UserID: userID, Name: "副業", Type: "income", Color: "#3B82F6", Icon: "computer", Description: "副業・フリーランス収入"},
		{UserID: userID, Name: "投資", Type: "income", Color: "#8B5CF6", Icon: "chart", Description: "株式・投資信託の利益"},
		{UserID: userID, Name: "その他収入", Type: "income", Color: "#6B7280", Icon: "money", Description: "その他の収入"},
		// 支出カテゴリ
		{UserID: userID, Name: "食費", Type: "expense", Color: "#EF4444", Icon: "food", Description: "食事・食材費"},
		{UserID: userID, Name: "住居費", Type: "expense", Color: "#F97316", Icon: "home", Description: "家賃・住宅ローン"},
		{UserID: userID, Name: "光熱費", Type: "expense", Color: "#EAB308", Icon: "lightning", Description: "電気・ガス・水道"},
		{UserID: userID, Name: "通信費", Type: "expense", Color: "#3B82F6", Icon: "phone", Description: "携帯・インターネット"},
		{UserID: userID, Name: "交通費", Type: "expense", Color: "#22C55E", Icon: "car", Description: "電車・バス・ガソリン"},
		{UserID: userID, Name: "医療費", Type: "expense", Color: "#EC4899", Icon: "hospital", Description: "病院・薬代"},
		{UserID: userID, Name: "日用品", Type: "expense", Color: "#84CC16", Icon: "bottle", Description: "洗剤・ティッシュなど"},
		{UserID: userID, Name: "衣服費", Type: "expense", Color: "#06B6D4", Icon: "shirt", Description: "洋服・靴・アクセサリー"},
		{UserID: userID, Name: "美容費", Type: "expense", Color: "#EC4899", Icon: "beauty", Description: "美容院・化粧品"},
		{UserID: userID, Name: "教育費", Type: "expense", Color: "#8B5CF6", Icon: "book", Description: "学費・書籍・習い事"},
		{UserID: userID, Name: "娯楽費", Type: "expense", Color: "#F59E0B", Icon: "game", Description: "映画・ゲーム・趣味"},
		{UserID: userID, Name: "交際費", Type: "expense", Color: "#EC4899", Icon: "users", Description: "飲み会・デート・プレゼント"},
		{UserID: userID, Name: "投資費", Type: "expense", Color: "#059669", Icon: "piggybank", Description: "株式・投資信託・積立投資"},
		{UserID: userID, Name: "その他支出", Type: "expense", Color: "#6B7280", Icon: "document", Description: "その他の支出"},
	}
	
	// 各カテゴリが存在するかチェックし、存在しない場合は作成
	for _, requiredCategory := range requiredCategories {
		var existingCategory Category
		result := db.Where("user_id = ? AND name = ? AND type = ?", userID, requiredCategory.Name, requiredCategory.Type).First(&existingCategory)
		
		if result.Error != nil {
			// カテゴリが存在しない場合は作成
			db.Create(&requiredCategory)
		}
	}
}

// 既存のすべてのユーザーに対して不足しているカテゴリを追加
func ensureAllUsersHaveDefaultCategories() {
	var users []User
	db.Find(&users)
	
	for _, user := range users {
		ensureUserHasDefaultCategories(user.ID)
	}
}

// デフォルトカテゴリ作成（後方互換性のため残す）
func createDefaultCategories(userID uint) {
	ensureUserHasDefaultCategories(userID)
}