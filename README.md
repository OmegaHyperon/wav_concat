Сервис для генерации аудио-файлов по заданной последовательности звуков из библиотеки. Предоставляет HTTP-API для передачи последовательности для генерации.

Библиотека звуков:
- lib_morf.dat - описание библиотеки, формат JSON, пример: 
{
    "descr": "lib of morfs", 
    "data": [
        {"morf": "beep", "file": "./beep.wav"}, 
        {"morf": "beeperr", "file": "./beeperr.wav"}
    ]
}
- набор файлов, описанных в lib_morf.dat

Запрос генерации файлов:
http://127.0.0.1:8081/data
{
    "formula": ["beep", "beeperr", "beep", "beeperr"],
    "fname": "summary2"
}
formula - последовательность звуков для склейки, которые будут найдены в библиотеки 
fname - имя файла, в который будет записан результирующий файл
Ответ: {Status: "OK"}
Результирующий файл сохраняется в папке SYNTH.resultdir (см ini)