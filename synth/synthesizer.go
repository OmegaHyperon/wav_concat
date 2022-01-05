package synth

/*
	https://github.com/youpy/go-wav

	https://mholt.github.io/json-to-go/

	libFonems:
	{
		"descr": "blah-blah-blah",
		"data": [
			{"morf": "welcome", "file": "welcome.wav"},
			{"morf": "one", "file": "one.wav"},
			{"morf": "end", "file": "end.wav"},
			{"morf": "silence/1", "file": "silence_1.wav"}
		]
	}

	formula:
	["welcone", "silence/1", "one", "silence/1", "end"]

*/

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"log"
	"path/filepath"

	"github.com/youpy/go-wav"
)

// Для хранения морфем
/* type t_libMorphemes struct {
	Descr string `json:"descr"`
	Data  []struct {
		Morf string `json:"fonem"` // Имя морфемы
		File string `json:"file"`  // Имя файла со звуком для морфемы
	} `json:"data"`
} */

type TFormula []string

type TInitSynth struct {
	PathLibMorf string
	PathResult  string
	ErrorLog    *log.Logger
	InfoLog     *log.Logger
}

// Объект синтезатор речи
type Synthesizer struct {
	numChannels     uint16 // Параметры результирующего файла
	sampleRate      uint32 // ~
	bitsPerSample   uint16 // ~
	pathLibMorfFile string // Имя файла словаря
	pathLibMorf     string // Путь к словарю и к файлам со звуками
	pathResult      string // Путь для записи результирующих файлов

	errorLog      	*log.Logger
	infoLog       	*log.Logger

	morphemesJSON 	string
	formula       	[]string            // Правило для синтеза речи
	morphemes     	map[string]*os.File // Словарь содержащий файлы для каждой фонемы
	wav_bytes     	map[string][]wav.Sample
	res           	[]wav.Sample // Результат обработки
}

// Загрузка всех доступных морфем в словарь
// На вход поступает JSON со словарем морфем
func (p *Synthesizer) loadMorphemes() bool {
	lres := true
	p.morphemes = make(map[string]*os.File)
	p.wav_bytes = make(map[string][]wav.Sample)

	lmfj := []byte(p.morphemesJSON)

	var dat map[string]interface{}
	if err := json.Unmarshal(lmfj, &dat); err != nil {
		p.errorLog.Printf("!!! Error in loadMorphemes: %s\n", err)
	}
	p.infoLog.Println("Morphemes: ", dat)
	dat1 := dat["data"].([]interface{})

	// https://gobyexample.com/json
	for _, idt := range dat1 {
		idt1 := idt.(map[string]interface{})

		fhndl, errf := os.Open(p.pathLibMorf + idt1["file"].(string))
		if errf != nil {
			p.errorLog.Printf("!!! Error of opening file: %s, %s\n", idt1["file"].(string), errf)
			lres = false
			break
		} else {
			p.infoLog.Println(">>> File is opened: ", idt1["file"].(string))
			defer fhndl.Close()

			p.morphemes[idt1["morf"].(string)] = fhndl
			smpld, errs := wav.NewReader(fhndl).ReadSamples()
			if errs == io.EOF {
				p.errorLog.Println("!!! Error of ReadSamples: ", idt1["morf"].(string))
				lres = false
				break
			}
			p.wav_bytes[idt1["morf"].(string)] = smpld
		}

		p.infoLog.Println(">>> Morphemes: ", p.morphemes)
	}

	return lres
}

func (p *Synthesizer) openLibMorfs() bool {
	lres := true
	lfname := filepath.Join(p.pathLibMorf, p.pathLibMorfFile)

	file, err := os.Open(lfname)
	if err != nil {
		p.errorLog.Printf("!!! Error in openLibMorfs: %s", err)
	}
	defer file.Close()

	ldata := make([]byte, 64)

	for {
		n, err := file.Read(ldata)
		if err == io.EOF { // если конец файла
			break // выходим из цикла
		}
		p.morphemesJSON += string(ldata[:n])
	}

	p.infoLog.Println(">>> Lib of morphemes was loaded: ", lfname)

	return lres
}

func (p *Synthesizer) loadFormula(formulaJSON string) bool {
	lfj := []byte(formulaJSON)
	var dat TFormula
	if err := json.Unmarshal(lfj, &dat); err != nil {
		panic(err)
	}
	p.formula = dat
	p.infoLog.Println(">>> Fomula is loaded: ", p.formula)

	return true
}

// Осуществляет сборку фразы по формуле
func (p *Synthesizer) assemble() bool {
	var res_data []wav.Sample

	p.infoLog.Println("Assemble the formula...")

	for _, i_word := range p.formula {
		res_data = append(res_data, p.wav_bytes[i_word]...)
	}

	p.res = res_data

	return true
}

// Запись файла с результатом сборки данных
func (p *Synthesizer) saveFile(resFName string) bool {
	lres := true
	lfname := filepath.Join(p.pathResult, fmt.Sprintf("%s.wav", resFName))

	file3, _ := os.Create(lfname)
	ln := uint32(len(p.res))
	wrtr := wav.NewWriter(file3, ln, p.numChannels, p.sampleRate, p.bitsPerSample)

	errw := wrtr.WriteSamples(p.res)
	if errw != nil {
		p.errorLog.Println("!!! Error of writing data: ", errw)
		lres = false
	}

	defer file3.Close()

	if lres {
		p.infoLog.Println("File is saved: ", lfname)
	}

	return lres
}

func (p *Synthesizer) Init(data TInitSynth) {
	// Initializing
	p.numChannels = 2
	p.sampleRate = 8000
	p.bitsPerSample = 16
	p.pathLibMorfFile = "lib_morf.dat"

	p.pathLibMorf = data.PathLibMorf // "./lib/"
	p.pathResult = data.PathResult   // "./result/"
	p.infoLog = data.InfoLog
	p.errorLog = data.ErrorLog
}

func (p *Synthesizer) Run(formu string, fname string) bool {
	// Главный обработчик
	// Обрабатывает формулу и генерирует файл с результирующей записью
	//
	// lformu := "[\"welcone\", \"silence/1\", \"one\", \"silence/1\", \"end\"]"
	// lformu = "[\"beep\", \"beeperr\", \"beep\", \"beeperr\"]"
	// Словарь фонем
	// lmorf := "{\"descr\": \"DDDDD\", \"data\": [{\"morf\": \"beep\", \"file\": \"./beep.wav\"}, {\"morf\": \"beeperr\", \"file\": \"beeperr.wav\"}]}"
	// syn := Synthesizer{}

	p.infoLog.Println("Start of the server...")

	lres := false
	if p.pathLibMorf != "" && p.pathResult != "" && p.infoLog != nil && p.errorLog != nil  {
		if p.openLibMorfs() {
			if p.loadFormula(formu) {
				if p.loadMorphemes() {
					if p.assemble() {
						if p.saveFile(fname) {
							lres = true
						}
					}
				}
			}
		}
	} else {
		fmt.Println("!!! Synthesizer wasn't fully inited.")
	}

	return lres
}

/* func test_main() {
	l_formula := "[\"beep\", \"beeperr\", \"beep\", \"beeperr\"]"
	l_res_fname := "./summary.wav"

	syn := Synthesizer{}
	syn.run(l_formula, l_res_fname)
} */
