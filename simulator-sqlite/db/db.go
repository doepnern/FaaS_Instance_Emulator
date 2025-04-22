package db

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/doepnern/faas-simulator/logs"
	_ "github.com/mattn/go-sqlite3"
)

const create string = `
  CREATE TABLE IF NOT EXISTS requests (
  id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
	targetInstance TEXT,
	receivedAt INTEGER NOT NULL,
	processedAt INTEGER DEFAULT NULL
  );
`

type Database struct {
	Connection *sql.DB
}

func New() (*Database, error) {
	var err error
	db, err := sql.Open("sqlite3", "file::memory:?cache=shared")
	db.SetConnMaxLifetime(-1)
	db.SetMaxOpenConns(1)
	db.Exec("PRAGMA journal_mode=MEMORY;")
	db.Exec("PRAGMA synchronous=OFF;")
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(create)
	if err != nil {
		panic(err)
	}
	InitResources(db)
	fmt.Println("Database initialized")
	return &Database{
		Connection: db,
	}, nil
}

func (db *Database) Close() {
	db.Connection.Close()
}

type Request struct {
	ID             int
	TargetInstance string
	ReceivedAt     string
	ProcessedAt    *string
}

func (db *Database) AddRequest(targetInstance string) (int, error) {
	timingStart := time.Now()
	var id int
	currentTime := time.Now().UnixNano()
	err := db.Connection.QueryRow("INSERT INTO requests (targetInstance, receivedAt) VALUES (?, ?) RETURNING id", targetInstance, currentTime).Scan(&id)
	logs.PerformanceLogger.Debug((fmt.Sprint("AddRequest took: ", time.Since(timingStart))))
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (db *Database) UpdateRequest(id int, processedAt int64) (sql.Result, error) {
	timingStart := time.Now()
	res, err := db.Connection.Exec("UPDATE requests SET processedAt = ? WHERE id = ?", processedAt, id)
	logs.PerformanceLogger.Debug((fmt.Sprint("UpdateRequest took: ", time.Since(timingStart))))
	return res, err
}

func (db *Database) GetInstanceOrder() ([]string, error) {
	timingStart := time.Now()
	rows, err := db.Connection.Query("SELECT DISTINCT targetInstance FROM requests WHERE processedAt IS NULL ORDER BY id")
	logs.PerformanceLogger.Debug((fmt.Sprint("GetInstanceOrder took: ", time.Since(timingStart))))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var instances []string
	for rows.Next() {
		var instance string
		if err := rows.Scan(&instance); err != nil {
			return nil, err
		}
		instances = append(instances, instance)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return instances, nil
}

func (db *Database) GetUnhandledRequestsByInstance(targetInstance string) ([]Request, error) {
	timingStart := time.Now()
	rows, err := db.Connection.Query("SELECT id, targetInstance, receivedAt, processedAt FROM requests WHERE processedAt IS NULL AND targetInstance = ? ORDER BY id", targetInstance)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var requests []Request
	for rows.Next() {
		var req Request
		if err := rows.Scan(&req.ID, &req.TargetInstance, &req.ReceivedAt, &req.ProcessedAt); err != nil {
			return nil, err
		}
		requests = append(requests, req)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	logs.PerformanceLogger.Debug((fmt.Sprint("GetUnhandledRequestsByInstance took: ", time.Since(timingStart))))
	return requests, nil
}

func (db *Database) GetHandledRequests(currentTime int64) ([]Request, error) {
	timingStart := time.Now()
	rows, err := db.Connection.Query("SELECT id, targetInstance, receivedAt, processedAt FROM requests WHERE processedAt IS NOT NULL AND processedAt < ? ORDER BY id", currentTime)
	logs.PerformanceLogger.Debug((fmt.Sprint("GetHandledRequests took: ", time.Since(timingStart))))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var requests []Request
	for rows.Next() {
		var req Request
		if err := rows.Scan(&req.ID, &req.TargetInstance, &req.ReceivedAt, &req.ProcessedAt); err != nil {
			return nil, err
		}
		requests = append(requests, req)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return requests, nil
}

func (db *Database) RemoveHandledRequests(currentTime int64) (sql.Result, error) {
	timingStart := time.Now()
	res, err := db.Connection.Exec("DELETE FROM requests WHERE processedAt IS NOT NULL AND processedAt < ?", currentTime)
	logs.PerformanceLogger.Debug((fmt.Sprint("RemoveHandledRequests took: ", time.Since(timingStart))))
	return res, err
}

func (db *Database) GetRequestsForInstance(instanceName string) ([]Request, error) {
	timingStart := time.Now()
	rows, err := db.Connection.Query("SELECT id, targetInstance, receivedAt, processedAt FROM requests WHERE targetInstance = ? AND processedAt IS NULL", instanceName)
	logs.PerformanceLogger.Debug((fmt.Sprint("GetRequestsForInstance took: ", time.Since(timingStart))))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var requests []Request
	for rows.Next() {
		var req Request
		if err := rows.Scan(&req.ID, &req.TargetInstance, &req.ReceivedAt, &req.ProcessedAt); err != nil {
			return nil, err
		}
		requests = append(requests, req)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return requests, nil
}

func (db *Database) GetAllRequests() ([]Request, error) {
	timingStart := time.Now()
	rows, err := db.Connection.Query("SELECT id, targetInstance, receivedAt, processedAt FROM requests")
	logs.PerformanceLogger.Debug((fmt.Sprint("GetAllRequests took: ", time.Since(timingStart))))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var requests []Request
	for rows.Next() {
		var req Request
		if err := rows.Scan(&req.ID, &req.TargetInstance, &req.ReceivedAt, &req.ProcessedAt); err != nil {
			return nil, err
		}
		requests = append(requests, req)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return requests, nil
}

func (db *Database) GetNotProcessedRequests() ([]Request, error) {
	timingStart := time.Now()
	rows, err := db.Connection.Query("SELECT id, targetInstance, receivedAt, processedAt FROM requests WHERE processedAt IS NULL")
	logs.PerformanceLogger.Debug((fmt.Sprint("GetNotProcessedRequests took: ", time.Since(timingStart))))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var requests []Request
	for rows.Next() {
		var req Request
		if err := rows.Scan(&req.ID, &req.TargetInstance, &req.ReceivedAt, &req.ProcessedAt); err != nil {
			return nil, err
		}
		requests = append(requests, req)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return requests, nil
}
