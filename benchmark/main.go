package main

import (
	"benchmark/data"
	"context"
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/tnh9570/tnhGoFramework/dbm"
)

var targetDBServerHost = "192.168.100.230"
var targetDBServerPort = "3306"
var targetDBServerUser = "manager"
var targetDBServerPassword = "mgrsol123"

var operationSec = 10 * 60 //10 minutes
var Concurrency = 8
var ResultLogFile = "tde_benchmark_results.csv"

func main() {
	err := ConnectDB()
	if err != nil {
		panic(err)
	}

	err = CheckBaseLine(operationSec)
	if err != nil {
		panic(err)
	}

	err = CheckOverhead(operationSec)
	if err != nil {
		panic(err)
	}

	err = CheckWorstCase(operationSec)
	if err != nil {
		panic(err)
	}

	err = DecryptTable("h00000.TCUSTOMERPERSONAL")
	if err != nil {
		panic(err)
	}
}

func ConnectDB() error {
	err := dbm.CreateDBAdapter(
		targetDBServerHost,
		targetDBServerPort,
		targetDBServerUser,
		targetDBServerPassword)
	if err != nil {
		return err
	}

	err = dbm.Run([]*dbm.Statement{
		{
			Query: "USE h00000;",
		},
	})
	if err != nil {
		return err
	}

	return nil
}

func CheckBaseLine(operationSec int) error {
	err := StartLoadAndLogResult("CheckBaseLine", operationSec)
	if err != nil {
		return err
	}
	return nil
}

func CheckOverhead(operationSec int) error {
	err := EncryptTable("TCUSTOMERPERSONAL")
	if err != nil {
		return err
	}

	fmt.Println("Starting Warm-up Run (60s)...")
	err = StartLoadAndLogResult("WarmUpRun", 60)
	if err != nil {
		return err
	}

	log.Printf("Starting Overhead Measurement (%ds)...\n", operationSec)
	err = StartLoadAndLogResult("CheckOverhead", operationSec)
	if err != nil {
		return err
	}

	return nil
}

func EncryptTable(tableName string) error {
	alterTableQuery := `ALTER TABLE ` + tableName + ` ENCRYPTED=YES;`
	err := dbm.Run([]*dbm.Statement{{Query: alterTableQuery}})
	if err != nil {
		return err
	}

	return nil
}

// restart DB server + //10분 실행
func CheckWorstCase(operationSec int) error {
	err := StartLoadAndLogResult("CheckWorstCase", operationSec)
	if err != nil {
		return err
	}
	return nil
}

// operationSec 동안 아래 쿼리를 반복 실행
func StartLoadAndLogResult(stage string, operationSec int) error {
	var wg sync.WaitGroup
	// 쿼리 실행 횟수 및 지연 시간을 카운트할 변수
	var queryCount int64

	// 1. 시간 제한 설정: operationSec 이후에 신호를 보낼 채널 생성
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(operationSec)*time.Second)
	defer cancel()

	// 2. 동시 실행 스레드 (Goroutine) 시작
	for i := 0; i < Concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// Goroutine이 종료될 때까지 무한 루프
			for {
				select {
				case <-ctx.Done():
					// operationSec 시간이 끝나면 루프를 종료합니다.
					return

				default:

					// 3. 쿼리 실행 및 시간 측정
					for _, query := range data.QueryList {
						// dbm.Run은 *dbm.Statement 슬라이스를 받으므로, 단일 쿼리라도 슬라이스 형태로 전달합니다.
						err := dbm.Run([]*dbm.Statement{{Query: query}})

						if err != nil {
							// TODO: DB 오류 처리 로직 (오류가 발생하면 해당 Goroutine만 종료하거나 리턴)
							log.Printf("Query error: %v\n", err)
							return
						}

						// 4. 통계 업데이트 (Lock 필요)
						atomic.AddInt64(&queryCount, 1)
						// TotalLatency 업데이트는 동시성 문제로 인해 복잡해지므로,
						// 평균 지연 시간 대신 TPS(처리량)에 초점을 맞춥니다.
						log.Println(stage, "쿼리실행완료!")
					}
				}
			}
		}()
	}

	// 5. 모든 Goroutine이 끝날 때까지 대기
	wg.Wait()

	// 6. 결과 계산
	totalQueries := atomic.LoadInt64(&queryCount)

	// 벤치마크 총 시간
	totalTime := time.Duration(operationSec) * time.Second

	// TPS = 총 쿼리 수 / 총 시간
	tps := float64(totalQueries) / totalTime.Seconds()

	// 3. CSV 로그 기록
	err := LogResult(stage, totalQueries, totalTime, tps)
	if err != nil {
		return fmt.Errorf("logging failed: %w", err)
	}

	return nil
}

// LogResult 함수는 측정 결과를 CSV 파일에 기록합니다.
func LogResult(stage string, totalQueries int64, totalTime time.Duration, tps float64) error {
	// 파일 열기 (없으면 생성, 있으면 덧붙이기)
	file, err := os.OpenFile(ResultLogFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush() // 버퍼 비우기

	// 파일이 비어 있으면 헤더 작성 (첫 실행 시에만)
	if fi, _ := file.Stat(); fi.Size() == 0 {
		header := []string{"Timestamp", "Stage", "TableName", "TotalQueries", "TotalTime_s", "TPS", "AvgLatency_ms"}
		if err := writer.Write(header); err != nil {
			return fmt.Errorf("failed to write header: %w", err)
		}
	}

	// 기록할 데이터 준비
	record := []string{
		time.Now().Format("2006-01-02 15:04:05"),
		stage,
		strconv.FormatInt(totalQueries, 10),
		fmt.Sprintf("%.2f", totalTime.Seconds()),
		fmt.Sprintf("%.2f", tps),
	}

	if err := writer.Write(record); err != nil {
		return fmt.Errorf("failed to write record: %w", err)
	}

	return nil
}

func DecryptTable(tableName string) error {
	alterTableQuery := `ALTER TABLE ` + tableName + ` ENCRYPTED=NO;`
	err := dbm.Run([]*dbm.Statement{{Query: alterTableQuery}})
	if err != nil {
		return err
	}

	log.Printf("Table %s decrypted successfully.\n", tableName)
	return nil
}
