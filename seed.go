package main

import (
	"golang.org/x/crypto/bcrypt"
)

// パスワードハッシュ化（seed用）
func hashPasswordForSeed(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func seedData() {
	// 開発環境でのみテストユーザーを作成
	var userCount int64
	db.Model(&User{}).Count(&userCount)
	if userCount == 0 {
		// テスト用ユーザー作成（開発環境のみ）
		hashedPassword, _ := hashPasswordForSeed("password123")
		testUser := User{
			Email:    "test@example.com",
			Password: hashedPassword,
			Name:     "テストユーザー",
		}
		
		db.Create(&testUser)
		
		// テストユーザー用のデフォルトカテゴリ作成
		createDefaultCategories(testUser.ID)
	}
	
	// 既存のすべてのユーザーに対して不足しているカテゴリを追加
	ensureAllUsersHaveDefaultCategories()
	
	// 既存のカテゴリのアイコンを絵文字から商用アイコンに更新
	updateEmojiIconsToCommercialIcons()
}

// 既存のカテゴリのアイコンを絵文字から商用アイコンに更新
func updateEmojiIconsToCommercialIcons() {
	// 絵文字から商用アイコンへのマッピング
	iconMapping := map[string]string{
		"💼": "briefcase",
		"🎁": "gift",
		"💻": "computer",
		"📈": "chart",
		"💵": "money",
		"💰": "money",
		"🍽️": "food",
		"🏠": "home",
		"⚡": "lightning",
		"📱": "phone",
		"🚗": "car",
		"🏥": "hospital",
		"🧴": "bottle",
		"👕": "shirt",
		"💄": "beauty",
		"📚": "book",
		"🎮": "game",
		"👥": "users",
		"🐷": "piggybank",
		"📄": "document",
	}
	
	// 各絵文字アイコンを商用アイコンに更新
	for emojiIcon, commercialIcon := range iconMapping {
		db.Model(&Category{}).Where("icon = ?", emojiIcon).Update("icon", commercialIcon)
	}
}

