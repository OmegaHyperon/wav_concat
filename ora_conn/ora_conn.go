// https://oralytics.com/category/go-lang/
// https://blogs.oracle.com/developers/post/connecting-to-oracle-databases-using-godror-and-sqlx

package ora_conn

import (
	"fmt"

	_ "github.com/godror/godror"
	"github.com/jmoiron/sqlx"
)

type user_tables_sql struct {
	TABLE_NAME      string `db:"TABLE_NAME"`
	TABLESPACE_NAME string `db:"TABLESPACE_NAME"`
}

// "time"
// "database/sql"
// godror "github.com/godror/godror"
// "github.com/jmoiron/sqlx"

func ConnectToOracle(username string, password string, database string) {
	// username = <username>
	// password := <password>
	// database := <database name>

	fmt.Println(">>> Connect to Oracle...")
	db, err := sqlx.Connect("godror", username+"/"+password+"@"+database)
	if err != nil {
		fmt.Println(err)
	}
	defer db.Close()

	fmt.Println(">>> Load all rows at once")
	sqlStr := "select table_name, tablespace_name from user_tables where 1=1 "
	table_recs := []user_tables_sql{}
	db.Select(&table_recs, sqlStr)
	//for i, r := range table_recs {
	//	tn, tsn := r.TABLE_NAME, r.TABLESPACE_NAME
	//	fmt.Println(i, tn, "/", tsn)
	//}

	fmt.Println(">>> Parsing each row separately")
	rows, err := db.Queryx(sqlStr)
	if err != nil {
		fmt.Println("!!! Error processing query: ")
		fmt.Println(err)
		return
	}
	defer rows.Close()

	table_rec := user_tables_sql{}
	for rows.Next() {
		err := rows.StructScan(&table_rec)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Printf("%#v\n", table_rec)
		break
	}

}
