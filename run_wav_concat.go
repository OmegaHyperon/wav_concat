package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"wav_concat/conf_util"
	"wav_concat/ora_conn"
	"wav_concat/route_table"
	"wav_concat/synth"

	_ "github.com/godror/godror"
)

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
	errorLog      *log.Logger
	infoLog       *log.Logger
	rotable       *route_table.RoTable
	synthInitData synth.TInitSynth
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

// Тест подключения к БД
func (a *application) testOracle() {
	ora_conn.ConnectToOracle(a.conf.DB_username, a.conf.DB_password, a.conf.DB_conn)
	fmt.Println(">>> End of ConnectToOracle")
}

// Main
func main() {
	appl := application{}

	appl.conf = &conf_util.ConfUtil{}
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

	appl.mutex = &sync.Mutex{}
	appl.counter = 0

	//
	appl.rotable = &route_table.RoTable{}
	(*appl.rotable).Init(
		appl.echoString,
		appl.dataPost)

	// Init for synthesizer
	appl.synthInitData.PathLibMorf = appl.conf.Synth_libdir
	appl.synthInitData.SaveResult = appl.conf.Synth_saveresult
	appl.synthInitData.PathResult = appl.conf.Synth_resultdir
	appl.synthInitData.InfoLog = appl.infoLog
	appl.synthInitData.ErrorLog = appl.errorLog

	appl.infoLog.Printf("Start of the server...")
	// appl.errorLog.Printf("No errors at start!")

	appl.infoLog.Printf(">>> Test of Oracle...")
	appl.testOracle()

	// With Mux...
	// mux := http.NewServeMux()
	// mux.HandleFunc("/headers", appl.headersString)
	// mux.HandleFunc("/inc", appl.incrementCounter)
	// mux.HandleFunc("/rand", appl.randDigit)
	// mux.HandleFunc("/", appl.echoString)

	// var serv_url string = fmt.Sprintf(":%d", appl.conf.Port)
	// addr := flag.String("addr", serv_url, "Сетевой адрес веб-сервера")
	// srv := &http.Server{
	// 	Addr:     *addr,
	// 	ErrorLog: appl.errorLog,
	// 	Handler:  mux,
	// }
	// fmt.Println(">>> ", serv_url, appl.conf.Port)

	router := http.HandlerFunc(appl.rotable.Serve)
	errserv := http.ListenAndServe(fmt.Sprintf(":%d", appl.conf.Port), router)
	appl.errorLog.Fatal(errserv)
}
