package main

import (
	"log"
	"time"
)

// 自動スケジューラーを開始する関数
func startScheduler() {
	log.Println("Starting automatic scheduler for monthly fixed transactions...")
	
	// goroutineで非同期実行
	go func() {
		for {
			now := time.Now()
			
			// 次の月の1日 00:00:00を計算
			nextMonth := now.AddDate(0, 1, 0)
			firstDayNextMonth := time.Date(nextMonth.Year(), nextMonth.Month(), 1, 0, 0, 0, 0, time.UTC)
			
			// 現在時刻から次の実行時刻までの待機時間を計算
			duration := firstDayNextMonth.Sub(now)
			
			log.Printf("Next monthly fixed transaction processing scheduled for: %s (in %v)", 
				firstDayNextMonth.Format("2006-01-02 15:04:05"), duration)
			
			// 指定時間まで待機
			time.Sleep(duration)
			
			// 月次固定収支処理を実行
			log.Println("Executing scheduled monthly fixed transaction processing...")
			processMonthlyFixedTransactions()
			log.Println("Scheduled monthly fixed transaction processing completed")
		}
	}()
}

// 開発・テスト用：即座に月次処理を実行する関数
func executeMonthlyProcessingNow() {
	log.Println("Executing monthly fixed transaction processing immediately (for testing)...")
	processMonthlyFixedTransactions()
	log.Println("Immediate monthly fixed transaction processing completed")
}

// 開発・テスト用：指定した時間後に月次処理を実行する関数
func scheduleMonthlyProcessingAfter(duration time.Duration) {
	log.Printf("Scheduling monthly fixed transaction processing to run in %v", duration)
	
	go func() {
		time.Sleep(duration)
		log.Println("Executing scheduled monthly fixed transaction processing...")
		processMonthlyFixedTransactions()
		log.Println("Scheduled monthly fixed transaction processing completed")
	}()
}

// サーバー起動時に当月の処理が必要かチェックして実行する関数
func checkAndProcessCurrentMonth() {
	log.Println("Checking if current month processing is needed...")
	
	now := time.Now()
	
	// 今月1日の日付を取得
	firstDayThisMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	
	// 今月の固定収支処理が既に実行されているかチェック
	var processedCount int64
	
	// アクティブな固定収支の数を取得
	var activeFixedExpensesCount int64
	db.Model(&FixedExpense{}).Where("is_active = ?", true).Count(&activeFixedExpensesCount)
	
	if activeFixedExpensesCount == 0 {
		log.Println("No active fixed expenses found, skipping current month processing")
		return
	}
	
	// 今月1日に作成された固定収支取引の数をチェック
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	endOfDay := startOfMonth.Add(24 * time.Hour).Add(-time.Second)
	
	db.Model(&Transaction{}).Where(
		"date BETWEEN ? AND ? AND (description LIKE ? OR description LIKE ?)",
		startOfMonth, endOfDay, "固定収入:%", "固定支出:%",
	).Count(&processedCount)
	
	// 処理済みの取引数がアクティブな固定収支数より少ない場合、処理を実行
	if processedCount < activeFixedExpensesCount {
		log.Printf("Current month processing needed. Found %d processed transactions, expected %d", 
			processedCount, activeFixedExpensesCount)
		processMonthlyFixedTransactions()
	} else {
		log.Printf("Current month processing already completed. Found %d processed transactions", processedCount)
	}
}