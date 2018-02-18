package DBServices

import (
	"database/sql"
	"strconv"
	"html/template"
)

func RowsAndPages(db *sql.DB, tablename string, pagesize int) (count int,numpages int) {
	numpages = 0
	count = 0
	row := db.QueryRow("SELECT COUNT(1) FROM "+tablename)
	if row.Scan(&count) != nil { return }
	numpages = count / pagesize
	if count % pagesize > 0 {numpages+=1}
	return
}

func BuildHTMLOption(db *sql.DB,tablename string,pagesize int, selected int,numpages int) (options template.HTML, err error) {
	var count int
	options = ""
	numpages = 0
	row := db.QueryRow("SELECT COUNT(1) FROM "+tablename)
	err = row.Scan(&count)
	if err != nil { return }
	numpages = count / pagesize
	if count % pagesize > 0 {numpages+=1}
	var (
		val string
		seltag string
		soptions string
	)
	soptions = "\r\n"
	for i:= 0; i < numpages; i++ {
		val = strconv.Itoa(i+1)
		if (i + 1 == selected) {
			seltag = "selected"
		} else {
			seltag = ""
		}
		soptions += "<option " + seltag + " value='" + val + "'>" + val + "</option>\r\n"
	}
	options = template.HTML(soptions)
	return
}
