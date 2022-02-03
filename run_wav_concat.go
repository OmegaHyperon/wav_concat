package main

// TODO: Сделать обработку параметра командной строки -с

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	_ "embed"
	"strings"
	"wav_concat/conf_util"
	"wav_concat/gcollect"
	"wav_concat/ora_conn"
	"wav_concat/route_table"
	"wav_concat/synth"

	_ "github.com/godror/godror"
)

//go:embed version
var version_res string
var Version string = strings.TrimSpace(version_res)

type tDataReqJSON struct {
	Formula []string `json:"formula"`
	Fname   string   `json:"fname"`
}

type tRespData struct {
	Status string `json:"status"`
	Res    string `json:"res"`
	Msg    string `json:"msg"`
}

type application struct {
	counter       int
	mutex         *sync.Mutex
	conf          *conf_util.ConfUtil
	rotable       *route_table.RoTable
	synthInitData synth.TInitSynth
	gcollect      *gcollect.GCollect
	oraConn       *ora_conn.OraConn

	infoLog  *log.Logger
	errorLog *log.Logger
}

func (a *application) headers(req *http.Request) string {
	var res string = ""

	for name, headers := range req.Header {
		for _, h := range headers {
			res += fmt.Sprintf("%v: %v\n", name, h)
		}
	}

	return res
}

func (a *application) prepareResp(status string, res string, msg string) string {
	// Prepare a json-responce for http-request
	var jtRespData []byte
	tRespData := tRespData{Status: status, Res: strconv.Itoa(a.counter), Msg: ""}
	jtRespData, err := json.Marshal(tRespData)
	if err != nil {
		jtRespData = []byte("{Status: \"ERROR\", Res: \"\", Msg: \"\"}")
		a.errorLog.Printf(fmt.Sprintf("!!! Error in prepareResp: %s", err.Error()))
	}

	return string(jtRespData)
}

func (a *application) echoString(w http.ResponseWriter, r *http.Request) {
	a.infoLog.Printf(">>> Start of echoString...")
	w.Header().Set("MYHEADER", "AAABBBCCCDDDEEEFFFGGG")
	fmt.Fprintf(w, "hello\n\n")
}

// Обработка команды /data
func (a *application) dataPost(w http.ResponseWriter, r *http.Request) {
	a.infoLog.Printf(">>> Start of dataPost...")

	lbody, lerr := io.ReadAll(r.Body)
	if lerr != nil {
		a.errorLog.Println("!!! Error in body: ", lerr)
	}

	lbodyj := []byte(string(lbody))

	var dat map[string]interface{}
	if err := json.Unmarshal(lbodyj, &dat); err != nil {
		panic(err)
	}
	lformulaj := dat["formula"].([]interface{})
	lfname := dat["fname"].(string)

	lformulab, err := json.Marshal(lformulaj)
	if err != nil {
		a.errorLog.Println("!!! dataPost: Error in converting of JSON: ", err, lformulab)
		fmt.Fprintf(w, a.prepareResp("ERROR", "NoData", ""))

	} else {
		lformula := string(lformulab)

		a.infoLog.Println("Details of request: ", lformula, lfname)

		lsyn := synth.Synthesizer{}
		lsyn.Init(a.synthInitData)
		lres := lsyn.Run(lformula, lfname)

		a.infoLog.Println("Result of run: ", lres)
		if lres {
			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "application/octet-stream")
			w.Header().Set("RESULTFNAME", lfname)
			lsyn.SaveStream(w)
		} else {
			fmt.Fprintf(w, a.prepareResp("ERROR", "NoData", ""))
		}

	}
}

func (a *application) saveLogInfo(msg ...string) {
	if a.infoLog != nil {
		a.infoLog.Printf(msg[0])
	} else {
		fmt.Println(msg[0])
	}
}

func (a *application) saveLogError(msg string) {
	if a.errorLog != nil {
		a.errorLog.Println(msg[0])
	} else {
		fmt.Println(msg[0])
	}
}

func (a *application) initPrint() {
	a.saveLogInfo("Version: " + a.conf.Version)

	a.saveLogInfo(fmt.Sprintf("Port: %d", a.conf.Port))
	a.saveLogInfo("DB_username: " + a.conf.DB_username)
	a.saveLogInfo("DB_conn: " + a.conf.DB_conn)

	a.saveLogInfo("Synth_libdir: " + a.conf.Synth_libdir)
	a.saveLogInfo(fmt.Sprintf("Synth_saveresult: %d", a.conf.Synth_saveresult))
	a.saveLogInfo("Synth_resultdir: " + a.conf.Synth_resultdir)
}

// Main
func main() {
	iniPathArg := flag.String(
		"conf",
		"/usr/local/etc/simple_api_golang/wav_concat_api.ini",
		"path to ini-file",
	)
	flag.Parse()

	appl := application{}

	appl.conf = &conf_util.ConfUtil{
		Ini_path: *iniPathArg,
		Version: Version,
	}
	appl.conf.LoadIniFile()

	f_log, err := os.OpenFile(appl.conf.LogFile, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		log.Fatal(err)
	}
	defer f_log.Close()

	mwh := io.MultiWriter(os.Stdout, f_log)

	// Используйте log.New() для создания логгера для записи информационных сообщений. Для этого нужно
	// три параметра: место назначения для записи логов (os.Stdout), строка
	// с префиксом сообщения (INFO или ERROR) и флаги, указывающие, какая
	// дополнительная информация будет добавлена. Обратите внимание, что флаги
	// соединяются с помощью оператора OR |.
	appl.infoLog = log.New(mwh, "INFO\t", log.Ldate|log.Ltime)

	// Создаем логгер для записи сообщений об ошибках таким же образом, но используем stderr как
	// место для записи и используем флаг log.Lshortfile для включения в лог
	// названия файла и номера строки где обнаружилась ошибка.
	appl.errorLog = log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	appl.conf.InfoLog = appl.infoLog
	appl.conf.ErrorLog = appl.errorLog
	appl.initPrint()

	appl.mutex = &sync.Mutex{}
	appl.counter = 0

	//
	appl.rotable = &route_table.RoTable{}
	(*appl.rotable).Init(
		appl.echoString,
		appl.dataPost)

	// appl.infoLog.Printf(">>> Test of Oracle...")
	appl.oraConn = &ora_conn.OraConn{
		Username: appl.conf.DB_username,
		Password: appl.conf.DB_password,
		Database: appl.conf.DB_conn,
		FuncName: appl.conf.DB_func_name,

		InfoLog:  appl.infoLog,
		ErrorLog: appl.errorLog,
	}
	appl.oraConn.ConnectToOracle()

	// Init for synthesizer
	appl.synthInitData.PathLibMorf = appl.conf.Synth_libdir
	appl.synthInitData.SaveResult = appl.conf.Synth_saveresult
	appl.synthInitData.PathResult = appl.conf.Synth_resultdir
	appl.synthInitData.InfoLog = appl.infoLog
	appl.synthInitData.ErrorLog = appl.errorLog
	appl.synthInitData.OraConn = appl.oraConn

	appl.infoLog.Printf("Start of the server...")

	appl.gcollect = &gcollect.GCollect{
		Is_stopped: 0,
		Ext:        "wav",
		PathDir:    "./result/",
		TimeOut:    600,

		InfoLog:  appl.infoLog,
		ErrorLog: appl.errorLog,
	}
	go appl.gcollect.StartLoop()

	router := http.HandlerFunc(appl.rotable.Serve)
	errserv := http.ListenAndServe(fmt.Sprintf(":%d", appl.conf.Port), router)
	appl.errorLog.Fatal(errserv)
}
