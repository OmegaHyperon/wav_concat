// package gcollect
package main

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
	is_stopped int
	ext        string // Обрабатываемое расширение файлов
	pathDir    string // Путь к хранилищу файлов, где производится обработка
	timeOut    int64  // Таймаут запрета удаления файла в секундах
	errorLog   *log.Logger
	infoLog    *log.Logger
}

func statTimes(name string) (atime, mtime, ctime time.Time, err error) {
	// Отдает дату изменения (+) файла
	// В Linux три различных временных метки, связанные с файлом:
	// 	время последнего доступа к содержимому ( atime ),
	// 	время последнего изменения содержимого ( mtime ),
	// 	и время последнего изменения индекса (метаданные, ctime ).

	fi, err := os.Stat(name)
	if err != nil {
		fmt.Println("!!! Ошибка в statTimes", err)
		return
	}
	mtime = fi.ModTime()
	stat := fi.Sys().(*syscall.Stat_t)
	atime = time.Unix(int64(stat.Atim.Sec), int64(stat.Atim.Nsec))
	ctime = time.Unix(int64(stat.Ctim.Sec), int64(stat.Ctim.Nsec))
	return
}

func (p *GCollect) startLoop() {
	fmt.Println("Start of GCollect...")

	for p.is_stopped == 0 {
		time.Sleep(time.Duration(p.timeOut)*time.Millisecond)

		files, err := ioutil.ReadDir(p.pathDir)
		if err != nil {
			fmt.Println("!!! Ошибка чтения директории: ", p.pathDir, err)
		}

		for _, file := range files {
			fmt.Println(">>> Found file: ", file.Name(), file.IsDir())

			if !file.IsDir() {
				lfpath := filepath.Join(p.pathDir, file.Name())
				fmt.Println(lfpath)

				atime, _, _, serr := statTimes(lfpath)
				if serr == nil {
					if time.Now().Unix()-atime.Unix() > p.timeOut && filepath.Ext(lfpath) == p.ext {
						rerr := os.Remove(lfpath)
						if rerr != nil {
							fmt.Println("!!! Error of file deleteing: ", lfpath, rerr)
						} else {
							fmt.Println(">>> File was deleted: ", lfpath)
						}
					}
				}
			}
		}
	}
}

func main() {
	gcollect := GCollect{0, "wav", "./files", 60, nil, nil}
	gcollect.startLoop()
}
