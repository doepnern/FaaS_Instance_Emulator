package db

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/doepnern/faas-simulator/logs"
)

const CreateResourceUsageTable = `CREATE TABLE IF NOT EXISTS resource_usage (
	id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
	resourceName TEXT NOT NULL,
	amount FLOAT NOT NULL,
	start INTEGER NOT NULL,
	until INTEGER NOT NULL,
	target TEXT NOT NULL
);`

const CreateResourceUsageIndex = `CREATE INDEX IF NOT EXISTS idx_resource_name ON resource_usage (resourceName);`

func (db *Database) AddResourceUsage(resourceName string, amount float64, start int64, until int64, target string) (sql.Result, error) {
	timingStart := time.Now()
	res, err := db.Connection.Exec("INSERT INTO resource_usage (resourceName, amount,start, until, target) VALUES (?, ?, ?, ?, ?)", resourceName, amount, start, until, target)
	logs.PerformanceLogger.Debug((fmt.Sprint("AddResourceUsage took: ", time.Since(timingStart))))
	return res, err
}

func (db *Database) RemoveStaleResourceUsage(currentTime int64) (sql.Result, error) {
	timingStart := time.Now()
	oneSecondAgo := currentTime - 1e9
	res, err := db.Connection.Exec("DELETE FROM resource_usage WHERE until < ? AND until < ?", currentTime, oneSecondAgo)
	logs.PerformanceLogger.Debug((fmt.Sprint("RemoveStaleResourceUsage took: ", time.Since(timingStart))))
	return res, err
}

func (db *Database) GetResourceUsageByResourceForTarget(instanceName string, currentTime int64) (map[string]float64, error) {
	timingStart := time.Now()
	rows, err := db.Connection.Query("SELECT resourceName, SUM(amount) FROM resource_usage WHERE target = ? AND until > ? GROUP BY resourceName", instanceName, currentTime)
	logs.PerformanceLogger.Debug((fmt.Sprint("GetResourceUsageByResourceForTarget took: ", time.Since(timingStart))))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	resourceUsage := make(map[string]float64)
	for rows.Next() {
		var resourceName string
		var totalAmount float64
		if err := rows.Scan(&resourceName, &totalAmount); err != nil {
			return nil, err
		}
		resourceUsage[resourceName] = totalAmount
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return resourceUsage, nil
}

// return the average resource usage for a target instance in a given time frame
func (db *Database) GetAverageResourceUsageInTimeFrame(targetInstance string, start int64, until int64) (map[string]float64, error) {
	timingStart := time.Now()
	rows, err := db.Connection.Query("SELECT resourceName, amount, start, until FROM resource_usage WHERE target = ? AND until > ? and start < ?", targetInstance, start, until)
	logs.PerformanceLogger.Debug((fmt.Sprint("GetAverageResourceUsageInTimeFrame took: ", time.Since(timingStart))))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	resourceUsage := make(map[string]float64)
	for rows.Next() {
		var resourceName string
		var amount float64
		var fromResource int64
		var untilResource int64
		if err := rows.Scan(&resourceName, &amount, &fromResource, &untilResource); err != nil {
			return nil, err
		}
		var fromTime = max(fromResource, start)
		var untilTime = min(untilResource, until)
		var percentage = float64(untilTime-fromTime) / float64(until-start)
		resourceUsage[resourceName] += amount * percentage
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return resourceUsage, nil

}

func InitResources(conn *sql.DB) {
	_, err := conn.Exec(CreateResourceUsageTable)
	if err != nil {
		panic(err)
	}
	_, err = conn.Exec(CreateResourceUsageIndex)
	if err != nil {
		panic(err)
	}
}
