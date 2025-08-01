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
}

