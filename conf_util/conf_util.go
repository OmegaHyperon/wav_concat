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
resultdir = ./result/

*/

package conf_util

import (
	"fmt"
	"os"

	"gopkg.in/ini.v1"
)

type ConfUtil struct {
	ini_path string
	Port     int
	LogFile  string

	DB_username string
	DB_password string
	DB_conn     string

	Synth_libdir    string
	Synth_resultdir string
}

func (p *ConfUtil) LoadIniFile() int {
	p.ini_path = "/usr/local/etc/simple_api_golang/simple_api_golang.ini"
	p.LogFile = "./logs/simple_api_golang.log"

	fmt.Println("Path to ini-file:", p.ini_path)
	cfg, err := ini.Load(p.ini_path)
	if err != nil {
		fmt.Println("Fail to read file:", p.ini_path, err)
		os.Exit(1)
	}

	p.Port = cfg.Section("SERVER").Key("port").MustInt(8080)

	p.DB_username = cfg.Section("DB").Key("user").String()
	p.DB_password = cfg.Section("DB").Key("passwd").String()
	p.DB_conn = cfg.Section("DB").Key("conn_str").String()

	p.Synth_libdir = cfg.Section("SYNTH").Key("libdir").String()
	p.Synth_resultdir = cfg.Section("SYNTH").Key("resultdir").String()

	fmt.Println("Port: ", p.Port)
	fmt.Println("DB_username: ", p.DB_username)
	fmt.Println("DB_conn: ", p.DB_conn)

	fmt.Println("synth_libdir: ", p.Synth_libdir)
	fmt.Println("synth_resultdir: ", p.Synth_resultdir)

	return 1
}
