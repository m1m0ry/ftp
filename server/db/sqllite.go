package db

import (
	"database/sql"
	"encoding/json"
	"errors"
	"path/filepath"

	"github.com/m1m0ry/golang/ftp/server/common"
	_ "github.com/mattn/go-sqlite3"
)

func InitSqliteStore(filePath string) (Store, error) {
	db, err := sql.Open("sqlite3", "./sql.db")
	if err != nil {
		return nil, err
	}

	sql_table := `
    CREATE TABLE IF NOT EXISTS fileinfo(
        uid INTEGER PRIMARY KEY AUTOINCREMENT,
        name TEXT NOT NULL UNIQUE,
		size INTEGER NOT NULL
		offset INTEGER
		status INTEGER
		host TEXT
    );
    `

	db.Exec(sql_table)

	return &SqliteStore{
		sqldb: db,
	}, nil
}

type SqliteStore struct {
	sqldb *sql.DB
}

func (sqlite *SqliteStore) Close(filePath string) {
	sqlite.sqldb.Close()
}

func (sqlite *SqliteStore) Get(filePath string) []byte {
	filename := filepath.Base(filePath)
	rows, err := sqlite.sqldb.Query("SELECT * FROM fileinfo WHERE name=" + filename)
	checkErr(err)

	var (
		uid      int
		name     string
		filesize int64
		offset   int64
		status   bool
		host     string
	)

	for rows.Next() {
		err = rows.Scan(&uid, &name, &filesize, &offset, &status, &host)
		checkErr(err)
	}

	info := common.FileInfo{
		Filename: name,
		Filesize: filesize,
		Offset:   offset,
	}
	data, err := json.Marshal(info)
	return data
}

func (sqlite *SqliteStore) Post(filePath string, fileinfo interface{}) (err error) {
	info, ok := fileinfo.(common.FileInfo)
	if ok != true {
		return errors.New("无法识别该格式")
	}

	stmt, err := sqlite.sqldb.Prepare("INSERT INTO fileinfo(name, size, offset) values(?,?,?)")
	checkErr(err)

	res, err := stmt.Exec(info.Filename, info.Filesize, info.Offset)
	checkErr(err)

	_, err = res.LastInsertId()
	checkErr(err)

	return nil
}

func (sqlite *SqliteStore) Put(filePath string, offset int64) {
	filename := filepath.Base(filePath)
	stmt, err := sqlite.sqldb.Prepare("update fileinfo set offset=? where name=?")
	checkErr(err)

	_, err = stmt.Exec(offset, filename)
	checkErr(err)
}

func (sqlite *SqliteStore) IsDone(filePath string) int64 {
	filename := filepath.Base(filePath)
	rows, err := sqlite.sqldb.Query("SELECT offset FROM fileinfo WHERE name=" + filename)
	checkErr(err)

	var offset int64

	for rows.Next() {
		err = rows.Scan(&offset)
		checkErr(err)
	}

	return offset
}

func (sqlite *SqliteStore) Delete(filePath string) {
	filename := filepath.Base(filePath)
	stmt, err := sqlite.sqldb.Prepare("delete from fileinfo where name=?")
	checkErr(err)

	_, err = stmt.Exec(filename)
	checkErr(err)
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
