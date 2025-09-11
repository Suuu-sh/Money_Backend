package main

import (
	"log"
	"time"
)

// 自動スケジューラーを開始する関数（バッチ処理として完全にバックグラウンドで実行）
func startScheduler() {
	log.Println("=== MoneyTracker Batch Scheduler Started ===")
	log.Println("Automatic monthly fixed transaction processing enabled")
	
	// goroutineで非同期実行
	go func() {
		for {
			now := time.Now()
			
			// 次の月の1日 00:00:00を計算
			nextMonth := now.AddDate(0, 1, 0)
			firstDayNextMonth := time.Date(nextMonth.Year(), nextMonth.Month(), 1, 0, 0, 0, 0, time.UTC)
			
			// 現在時刻から次の実行時刻までの待機時間を計算
			duration := firstDayNextMonth.Sub(now)
			
			log.Printf("[SCHEDULER] Next execution: %s (in %v)", 
				firstDayNextMonth.Format("2006-01-02 15:04:05"), duration)
			
			// 指定時間まで待機
			time.Sleep(duration)
			
			// 月次固定収支処理を実行
			log.Printf("[BATCH] Starting monthly fixed transaction processing for %s", 
				firstDayNextMonth.Format("2006-01"))
			
			startTime := time.Now()
			processMonthlyFixedTransactions()
			processingTime := time.Since(startTime)
			
			log.Printf("[BATCH] Monthly processing completed in %v", processingTime)
		}
	}()
}

// サーバー起動時に当月の処理が必要かチェックして実行する関数（バッチ処理）
func checkAndProcessCurrentMonth() {
	log.Println("[BATCH] Checking current month processing status...")
	
	now := time.Now()
	currentMonth := now.Format("2006-01")
	
	// 今月の固定収支処理が既に実行されているかチェック
	var processedCount int64
	
	// アクティブな固定収支の数を取得
	var activeFixedExpensesCount int64
	db.Model(&FixedExpense{}).Where("is_active = ?", true).Count(&activeFixedExpensesCount)
	
	if activeFixedExpensesCount == 0 {
		log.Printf("[BATCH] No active fixed expenses found for %s, skipping processing", currentMonth)
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
		log.Printf("[BATCH] Processing required for %s: %d/%d transactions processed", 
			currentMonth, processedCount, activeFixedExpensesCount)
		
		startTime := time.Now()
		processMonthlyFixedTransactions()
		processingTime := time.Since(startTime)
		
		log.Printf("[BATCH] Current month processing completed in %v", processingTime)
	} else {
		log.Printf("[BATCH] Processing already completed for %s: %d transactions processed", 
			currentMonth, processedCount)
	}
}