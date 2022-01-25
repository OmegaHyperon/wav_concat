// https://oralytics.com/category/go-lang/
// https://blogs.oracle.com/developers/post/connecting-to-oracle-databases-using-godror-and-sqlx

package ora_conn

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"

	_ "github.com/godror/godror"
	"github.com/jmoiron/sqlx"
)

// type userTablesSql struct {
// 	TABLE_NAME      string `db:"TABLE_NAME"`
// 	TABLESPACE_NAME string `db:"TABLESPACE_NAME"`
// }

type tSimpleRequest struct {
	Action string `json:"action"`
}

type tSaveRequest struct {
	Action  string `json:"action"`
	Id      int64  `json:"id"`
	Formula string `json:"formula"`
	Fname   string `json:"fname"`
	Status  int    `json:"status"`
}

type OraConn struct {
	Username string
	Password string
	Database string

	db *sqlx.DB

	InfoLog  *log.Logger
	ErrorLog *log.Logger
}

func (p *OraConn) saveLogInfo(msg ...string) {
	if p.InfoLog != nil {
		p.InfoLog.Printf("OraConn: %s", msg[0])
	} else {
		fmt.Printf("OraConn: %s\n", msg[0])
	}
}

func (p *OraConn) saveLogError(msg ...string) {
	if p.ErrorLog != nil {
		p.ErrorLog.Printf("OraConn: %s", msg[0])
	} else {
		fmt.Printf("OraConn: %s\n", msg[0])
	}
}

func (p *OraConn) ConnectToOracle() error {
	// Организует подключение к Oracle

	p.saveLogInfo(">>> Connect to Oracle...")
	db, err := sqlx.Connect("godror", p.Username+"/"+p.Password+"@"+p.Database)
	if err != nil {
		msgErr := fmt.Sprintf("Error of connecttion: %s", err)
		p.saveLogError(msgErr)
		return errors.New(msgErr)
	}
	// defer db.Close()
	db.Ping()

	//p.saveLogInfo(">>> Load all rows at once")
	//sqlStr := "select table_name, tablespace_name from user_tables where 1=1 "
	//table_recs := []userTablesSql{}
	//db.Select(&table_recs, sqlStr)
	//for i, r := range table_recs {
	//	tn, tsn := r.TABLE_NAME, r.TABLESPACE_NAME
	//	fmt.Println(i, tn, "/", tsn)
	//}

	//p.saveLogInfo(">>> Parsing each row separately")
	//rows, err := db.Queryx(sqlStr)
	//if err != nil {
	//	p.saveLogError(fmt.Sprintf("!!! Error processing query: %s", err))
	//	return errors.New("Error")
	//}
	//defer rows.Close()

	//table_rec := userTablesSql{}
	//for rows.Next() {
	//	err := rows.StructScan(&table_rec)
	//	if err != nil {
	//		fmt.Println(err)
	//	}
	//	fmt.Printf("%#v\n", table_rec)
	//	break
	//}

	p.db = db
	return nil
}

// Получить ID запроса для записи в БД
// Request:
// {
//	"action": "get_id"
// }
// Response:
// {
//  "res": "OK",
//	"value": 1111
// }
func (p *OraConn) GetId() (int64, error) {
	var resInt int64 = -1
	err := errors.New("Error in getID: Unknown error")

	strSR := &tSimpleRequest{
		Action: "get_id",
	}
	strB, _ := json.Marshal(strSR)

	retVal, err := p.execAction(string(strB))
	if err != nil {
		p.saveLogError("!!! Error in GetId: ", err.Error())

	} else {
		var dat map[string]interface{}
		if err = json.Unmarshal([]byte(retVal), &dat); err != nil {
			p.saveLogError(fmt.Sprintf("!!! Error in result of getId: %s", err))
		} else {
			res := strings.ToUpper(dat["res"].(string))
			if res != "OK" {
				p.saveLogError(fmt.Sprintf("!!! Error in action: %s\n%s", retVal, err))
			} else {
				err = nil
				if resTmp, errm := dat["value"]; !errm {
					resInt = -1
					err = errors.New("!!! Error in getID: Unknown result")
				} else {
					resInt = int64(resTmp.(float64))
				}
			}
		}

	}

	return resInt, err
}

// Запись запроса на генерации звука в историю
// Request:
// {
//	"action": "save",	-
// 	"id": 5345, 		- ID запроса
// 	"formula": "", 		- формула генерации файла
// 	"fname": "", 		- имя сформированного файла
// 	"status": 1			- успешное выполнение операции
// }
// Response:
// {
//  "res": "OK"
// }
func (p *OraConn) SaveRequest(id int64, formula, fName string, status bool) error {
	var sInt int = 0
	if status {
		sInt = 1
	}
	lFName := fmt.Sprintf("%s.wav", fName)

	strSR := &tSaveRequest{
		Action:  "save",
		Id:      id,
		Formula: formula,
		Fname:   lFName,
		Status:  sInt,
	}
	strB, _ := json.Marshal(strSR)
	_, err := p.execAction(string(strB))

	return err
}

// Call procedure to perform action
// https://forum.golangbridge.org/t/how-to-call-oracle-stored-procedure-with-custom-type-out-parameter/22241
func (p *OraConn) execAction(action string) (string, error) {

	/*
		CREATE OR REPLACE FUNCTION test_func (
			action IN VARCHAR2
		) RETURN VARCHAR2
		AS
			retval VARCHAR2(4000);
		BEGIN
			INSERT INTO tmp_aaa(str) VALUES(action);
			COMMIT;

			retval := '{"res":"OK"}';

			RETURN retval;
		END;
	*/

	var retVal string
	var retErr error
	query := "BEGIN :retVal := test_func(:action); commit; END;"

	stmt, err := p.db.Prepare(query)
	if err != nil {
		errMsg := fmt.Sprintf("!!! Error of execution: %s", err)
		p.saveLogError(errMsg)
		retErr = err
	} else {
		defer stmt.Close()

		qres, err := stmt.Exec(
			sql.Named("action", action),
			sql.Named("retVal", sql.Out{Dest: &retVal}),
		)
		if err != nil {
			errMsg := fmt.Sprintf("!!! Error of running %q: %+v", query, err)
			p.saveLogError(errMsg)
			retErr = err
		} else {
			fmt.Println("Res: ", qres, retVal)

			var dat map[string]interface{}
			if err := json.Unmarshal([]byte(retVal), &dat); err != nil {
				p.saveLogError(fmt.Sprintf("!!! Error in result: %s", err))
			}
			res := strings.ToUpper(dat["res"].(string))
			if res != "OK" {
				p.saveLogError(fmt.Sprintf("!!! Error in action: %s\n%s", retVal, err))
			}
		}
	}

	return retVal, retErr
}
