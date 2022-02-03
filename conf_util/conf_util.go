/*
Example of ini-file:

[SERVER]
port = 8081

[DB]
user = dbuser
passw = passwd
dbname = 127.0.0.1:1521/dbname
dbfunc = test_func

[SYNTH]
libdir = ./lib/

; Запись созданного файла в хранилище
saveresult = 0

; Хранилище для созданных файлов
resultdir = ./result/

*/

package conf_util

import (
	_ "embed"
	"fmt"
	"log"

	"gopkg.in/ini.v1"
)

type ConfUtil struct {
	Version  string
	Ini_path string
	Port     int
	LogFile  string

	DB_username  string
	DB_password  string
	DB_conn      string
	DB_func_name string

	Synth_libdir     string
	Synth_saveresult int
	Synth_resultdir  string

	ErrorLog *log.Logger
	InfoLog  *log.Logger
}

func (p *ConfUtil) saveLogInfo(msg ...string) {
	for _, item := range msg {
		if p.InfoLog != nil {
			p.InfoLog.Println(item)
		} else {
			fmt.Println(item)
		}
	}
}

func (p *ConfUtil) saveLogError(msg ...string) {
	for _, item := range msg {
		if p.ErrorLog != nil {
			p.ErrorLog.Println(item)
		} else {
			fmt.Print(item)
		}
	}
}

func (p *ConfUtil) LoadIniFile() {
	p.LogFile = "./logs/wav_concat_api.log"

	p.saveLogInfo("Path to ini-file: " + p.Ini_path)
	cfg, err := ini.Load(p.Ini_path)
	if err != nil {
		p.saveLogError("Fail to load ini-file: " + p.Ini_path + "\n" + err.Error())
		panic("Ini-file file is'nt found")
	}

	p.Port = cfg.Section("SERVER").Key("port").MustInt(8080)

	p.DB_username = cfg.Section("DB").Key("user").String()
	p.DB_password = cfg.Section("DB").Key("passw").String()
	p.DB_conn = cfg.Section("DB").Key("dbname").String()
	p.DB_func_name = cfg.Section("DB").Key("dbfunc").String()

	p.Synth_libdir = cfg.Section("SYNTH").Key("libdir").String()
	p.Synth_saveresult, err = cfg.Section("SYNTH").Key("saveresult").Int()
	if err != nil {
		p.Synth_saveresult = 0
	}
	p.Synth_resultdir = cfg.Section("SYNTH").Key("resultdir").String()
}
