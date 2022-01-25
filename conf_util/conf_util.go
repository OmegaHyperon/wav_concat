/*
Example of ini-file:

[SERVER]
port = 8081

[DB]
user = dbuser
passwd = passwd
conn_str = 127.0.0.1:1521/dbname

[SYNTH]
libdir = ./lib/

; Запись созданного файла в хранилище
saveresult = 0

; Хранилище для созданных файлов
resultdir = ./result/

*/

package conf_util

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"gopkg.in/ini.v1"
)

type ConfUtil struct {
	Version  string
	ini_path string
	Port     int
	LogFile  string

	DB_username string
	DB_password string
	DB_conn     string

	Synth_libdir     string
	Synth_saveresult int
	Synth_resultdir  string

	ErrorLog *log.Logger
	InfoLog  *log.Logger
}

func (p *ConfUtil) saveLogInfo(msg ...string) {
	if p.InfoLog != nil {
		p.InfoLog.Printf(msg[0])
	} else {
		fmt.Println(msg[0])
	}
}

func (p *ConfUtil) saveLogError(msg string) {
	if p.ErrorLog != nil {
		p.ErrorLog.Println(msg[0])
	} else {
		fmt.Println(msg[0])
	}
}

func (p *ConfUtil) loadVersionFile() string {
	lverfpath := "./version"
	lverf, lerr := os.Open(lverfpath)
	if lerr != nil {
		fmt.Println("Не найден файл с версией проекта: ", lverfpath)
		panic("Version file is'nt found")
	}
	verb, lerr2 := ioutil.ReadAll(lverf)
	if lerr2 != nil {
		fmt.Println("Не найден файл с версией проекта: ", lverfpath)
		panic("Version file is'nt found")
	}

	p.Version = string(verb)

	return p.Version
}

func (p *ConfUtil) LoadIniFile() {
	p.loadVersionFile()

	p.ini_path = "/usr/local/etc/simple_api_golang/simple_api_golang.ini"
	p.LogFile = "./logs/simple_api_golang.log"

	p.saveLogInfo("Path to ini-file: " + p.ini_path)
	cfg, err := ini.Load(p.ini_path)
	if err != nil {
		p.saveLogError("Fail to read file: " + p.ini_path + err.Error())
		os.Exit(1)
	}

	p.Port = cfg.Section("SERVER").Key("port").MustInt(8080)

	p.DB_username = cfg.Section("DB").Key("user").String()
	p.DB_password = cfg.Section("DB").Key("passwd").String()
	p.DB_conn = cfg.Section("DB").Key("conn_str").String()

	p.Synth_libdir = cfg.Section("SYNTH").Key("libdir").String()
	p.Synth_saveresult, err = cfg.Section("SYNTH").Key("saveresult").Int()
	if err != nil {
		p.Synth_saveresult = 0
	}
	p.Synth_resultdir = cfg.Section("SYNTH").Key("resultdir").String()
}
