package clickhouse

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/ClickHouse/clickhouse-go"
	"github.com/vijeyash1/genericgitevents/model"
)

func GetClickHouseConnection(url string) (*sql.DB, error) {
	connect, err := sql.Open("clickhouse", url)
	if err != nil {
		log.Fatal(err)
	}
	if err := connect.Ping(); err != nil {
		if exception, ok := err.(*clickhouse.Exception); ok {
			fmt.Printf("[%d] %s \n%s\n", exception.Code, exception.Message, exception.StackTrace)
		} else {
			fmt.Println(err)
		}
		return nil, err
	}
	return connect, nil
}

func CreateGitSchema(connect *sql.DB) {
	_, err := connect.Exec(`
		CREATE TABLE IF NOT EXISTS vijesh (
			id                UUID,
			commitedBy        String,
			commitedAt        DateTime,
			repository        String,
			commitstat        String,
			availablebranches String,
			commitmessage     String	
		) engine=File(TabSeparated)
	`)
	if err != nil {
		log.Fatal(err)
	}
}
func InsertGitEvent(connect *sql.DB, metrics model.Githubevent) {
	var (
		tx, _   = connect.Begin()
		stmt, _ = tx.Prepare("INSERT INTO vijesh (id, commitedBy, commitedAt, repository, commitstat, availablebranches, commitmessage) VALUES (?, ?, ?, ?, ?, ?, ?)")
	)

	defer stmt.Close()
	if _, err := stmt.Exec(
		metrics.Uuid,
		metrics.CommitedBy,
		metrics.CommitedAt,
		metrics.Repository,
		metrics.Commitstat,
		metrics.Availablebranches,
		metrics.Commitmessage,
	); err != nil {
		log.Fatal(err)
	}
	if err := tx.Commit(); err != nil {
		log.Fatal(err)
	}
}
