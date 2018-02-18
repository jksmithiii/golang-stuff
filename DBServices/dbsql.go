// dbsql
package DBServices

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"log"
	"JKString"
	"strconv"
)

type TQueryResults struct {
	err error
	resultset string // json
	rowcount int
}

func (r *TQueryResults) QueryError() error {
	return r.err
}

func (r *TQueryResults) QueryResults() string {
	return r.resultset
}

func (r *TQueryResults) QueryRowCount() int {
	return r.rowcount
}

type TQuerySummaryData struct {
	ErrorStr string `json:"errorstr"`
	RowCount int    `json:"rowcount"`
}

type TQueryParams struct {
	Cols      string
	Tablename string
	Wherename string
	Whereval  string
	Orderval  string
	Ordersort int
	PageSize  int
	Page      int
}

func prepareInsert(it interface{}, tablename string) string {
	// all string values?
	val := reflect.ValueOf(it).Elem()
	res := ""
	var names string = "INSERT INTO " + tablename + " ("
	var values string = "VALUES ("
	numfields := val.NumField()
	for i := 0; i < numfields; i++ {
		valueField := val.Field(i)
		typeField := val.Type().Field(i)
		names = JKString.Commaindex(names+typeField.Name, i, numfields)
		values = JKString.Commaindex(values+JKString.UTF8SingleQuoted(JKString.ToString(valueField.Interface())), i, numfields)
	}
	names = names + ") "
	values = values + ")"
	res = names + values
	return res
}

func prepareUpdate(it interface{}, tablename string, wherename string) (res string, ok bool) {
	// handles only one identity field in where.
	val := reflect.ValueOf(it).Elem()
	res = "UPDATE " + tablename + " SET "
	numfields := val.NumField()
	var tstr string
	var cnt int = 0
	var wname string = strings.ToUpper(wherename)
	for i := 0; i < numfields; i++ {
		if strings.ToUpper(val.Type().Field(i).Name) == wname {
			wname = " WHERE " + wname + " = "
		} else {
			tstr = val.Type().Field(i).Name + "=" + JKString.QuoteStringIfNeeded(JKString.ToString(val.Field(i).Interface()))
			if cnt > 0 {
				tstr = "," + tstr
			}
			cnt++
		}
		res += tstr
	}
	res += wname
	ok = len(res) > 0
	return
}

func prepareDelete(tablename string, wherename string) (res string, ok bool) {
	// handles only one identity field equal to value in where.
	ok = (len(tablename) > 0) && (len(wherename) > 0)
	res = "DELETE FROM " + tablename + " WHERE " + wherename
	return
}

func execStatement(db *sql.DB, sqlStatement string, rw http.ResponseWriter) (res error) {
	_, res = db.Exec(sqlStatement)
	if res == nil {
		return
	}
	http.Error(rw, JKString.ToString(res)+": "+http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	return
}

func PrepareAndExec(db *sql.DB, it interface{}, tablename string, wherename string,
	whereval string, aUpdate bool, rw http.ResponseWriter) (res error) {
	res = nil
	sqlStatement, ok := prepareUpdate(it, tablename, wherename)
	if !ok {
		res = errors.New("error preparing update statement")
		http.Error(rw, JKString.ToString(res)+": "+http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	sqlStatement += JKString.QuoteStringIfNeeded(whereval)
	res = execStatement(db, sqlStatement, rw)
	return
}

func PrepareAndExecInsert(db *sql.DB, it interface{}, tablename string) (res error) {
	res = nil
	log.Println("PrepareAndExecInsert:1")
	sqlStatement:= prepareInsert(it,tablename)
	log.Println("PrepareAndExecInsert:2")
	if sqlStatement == "" {
		res = errors.New("error preparing insert statement")
		return
	}
	log.Println(sqlStatement)
	_, res = db.Exec(sqlStatement)
	log.Println("after exec")
	return res
}

func QueryRows(db *sql.DB, params TQueryParams) (res string, queryres TQuerySummaryData) {
	// simple query, returns json of row(s)
	queryres = TQuerySummaryData{"", 0}
	querystr := "select " + params.Cols + " from " + params.Tablename
	if (params.Wherename != "") && (params.Whereval != "") {
		querystr = querystr + " where " + params.Wherename + " = " + params.Whereval
	}
	if params.Orderval != "" {
		querystr = querystr + " order by " + params.Orderval
	}
	if params.Ordersort > 0 {
		querystr = querystr + " asc"
	} else if params.Ordersort < 0 {
		querystr = querystr + " desc"
	}
	//if (params.PageSize > 0) && (params.Page > 0) { // test some out of range and invalid value queries
	//	querystr = querystr + " limit " + inttostr(((page-1) * pagesize))+1) + ","+inttostr(pagesize)
	//}
	log.Println(querystr)
	if (db == nil) {
		log.Println("db is nil")
	}
	rows, dberr := db.Query(querystr)
	if dberr != nil {
		queryres.ErrorStr = JKString.ToString(dberr) + ": " + http.StatusText(http.StatusBadRequest)
		log.Println(queryres.ErrorStr)
		return
	}

	columns, dberr := rows.Columns()
	if dberr != nil {
		queryres.ErrorStr = JKString.ToString(dberr) + ": " + http.StatusText(http.StatusBadRequest)
		log.Println(queryres.ErrorStr)

		return
	}
  log.Println(columns)
	// Fetch rows
	pointers := make([]interface{}, len(columns))
	container := make([]interface{}, len(pointers))

	for i, _ := range pointers {
		pointers[i] = &container[i]
		columns[i] = strings.ToLower(columns[i])
	}

	res = "{"
	queryres.RowCount = 0
	for rows.Next() {
		queryres.RowCount += 1
		if queryres.RowCount > 1 {
			res += ",{"
		}
		rows.Scan(pointers...)
		for i, value := range container {
			if i > 0 {
				res += ","
			}
			res += JKString.AddDoubleQuotes(columns[i]) + ":" + JKString.AddDoubleQuotes(JKString.OrNULL(fmt.Sprintf("%s",value)))
		}
		res += "}"
	}

	if queryres.RowCount > 1 {
		res = "[" + res + "]"
	}
	return
}

func QueryRowsv2(db *sql.DB, params TQueryParams) TQueryResults {
	// simple query, returns json of row(s)
	querystr := "select " + params.Cols + " from " + params.Tablename
	if (params.Wherename != "") && (params.Whereval != "") {
		querystr = querystr + " where " + params.Wherename + " = " + params.Whereval
	}
	if params.Orderval != "" {
		querystr = querystr + " order by " + params.Orderval
	}
	if params.Ordersort > 0 {
		querystr = querystr + " asc"
	} else if params.Ordersort < 0 {
		querystr = querystr + " desc"
	}
	// MYSQL specific: fix this code goddamit its cool
	if (params.PageSize > 0) && (params.Page > 0) { // test some out of range and invalid value queries
	   querystr = querystr + " limit " + strconv.Itoa(((params.Page-1) * params.PageSize)+1) + ","+strconv.Itoa(params.PageSize)
	}

	var (results TQueryResults
		rows *sql.Rows
		columns []string
	)
	log.Println("From QueryRowsv2: "+querystr)
	rows, results.err = db.Query(querystr)
	if results.err != nil {return results}

	columns, results.err = rows.Columns()
	if results.err != nil {return results}

	// Fetch rows
	pointers := make([]interface{}, len(columns))
	container := make([]interface{}, len(pointers))

	for i, _ := range pointers {
		pointers[i] = &container[i]
		columns[i] = strings.ToLower(columns[i])
	}

	results.resultset = "{"
	results.rowcount = 0
	for rows.Next() {
		results.rowcount += 1
		if results.rowcount > 1 {
			results.resultset += ",{"
		}
		rows.Scan(pointers...)
		for i, value := range container {
			if i > 0 {results.resultset += ","}
			results.resultset += JKString.AddDoubleQuotes(columns[i]) + ":" + JKString.AddDoubleQuotes(JKString.OrNULL(fmt.Sprintf("%s",value)))
		}
		results.resultset += "}"
	}

	if results.rowcount > 1 {
		results.resultset = "[" + results.resultset + "]"
	}
	return results
}
