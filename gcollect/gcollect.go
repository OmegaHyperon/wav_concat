package gcollect
// package main

/*
	Класс занимается сбором и удалением старых файлов

*/

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"syscall"
	"time"
)

// Класс-чистильщик
type GCollect struct {
	Is_stopped int
	Ext        string // Обрабатываемое расширение файлов
	PathDir    string // Путь к хранилищу файлов, где производится обработка
	TimeOut    int64  // Таймаут запрета удаления файла в секундах

	InfoLog    *log.Logger
	ErrorLog   *log.Logger
}

func (p *GCollect) saveLogInfo(msg ...string) {
	if p.InfoLog != nil {
		p.InfoLog.Println(msg[0])
	} else {
		fmt.Println(msg[0])
	}
}

func (p *GCollect) saveLogError(msg string) {
	if p.ErrorLog != nil {
		p.ErrorLog.Println(msg[0])
	} else {
		fmt.Println(msg[0])
	}
}

func (p *GCollect) statTimes(name string) (atime, mtime, ctime time.Time, err error) {
	// Отдает дату изменения (+) файла
	// В Linux три различных временных метки, связанные с файлом:
	// 	время последнего доступа к содержимому ( atime ),
	// 	время последнего изменения содержимого ( mtime ),
	// 	и время последнего изменения индекса (метаданные, ctime ).

	fi, err := os.Stat(name)
	if err != nil {
		p.saveLogError(fmt.Sprintf("!!! Ошибка в statTimes: %s", err))
		return
	}
	mtime = fi.ModTime()
	stat := fi.Sys().(*syscall.Stat_t)
	atime = time.Unix(int64(stat.Atim.Sec), int64(stat.Atim.Nsec))
	ctime = time.Unix(int64(stat.Ctim.Sec), int64(stat.Ctim.Nsec))
	
	return
}

func (p *GCollect) StartLoop() {
	p.saveLogInfo("Start of GCollect...")

	for p.Is_stopped == 0 {
		fmt.Println(">>> .")

		time.Sleep(time.Duration(p.TimeOut)*time.Millisecond)

		files, err := ioutil.ReadDir(p.PathDir)
		if err != nil {
			p.saveLogError(fmt.Sprintf("!!! Ошибка чтения директории: %s; %s", p.PathDir, err))
		}

		for _, file := range files {
			// fmt.Println(">>> Found file: ", file.Name(), file.IsDir())

			if !file.IsDir() {
				lfpath := filepath.Join(p.PathDir, file.Name())
				atime, _, _, serr := p.statTimes(lfpath)
				if serr == nil {
					if time.Now().Unix()-atime.Unix() > p.TimeOut && filepath.Ext(lfpath) == p.Ext {
						rerr := os.Remove(lfpath)
						if rerr != nil {
							p.saveLogError(fmt.Sprintf("!!! Error of file deleteing: %s; %s", lfpath, rerr))
						} else {
							p.saveLogInfo(fmt.Sprintf(">>> File was deleted: %s", lfpath))
						}
					}
				}
			}
		}

		time.Sleep(30000)
	}
}

// func main() {
// 	gcollect := GCollect{0, "wav", "./files", 60, nil, nil}
// 	gcollect.StartLoop()
// }
