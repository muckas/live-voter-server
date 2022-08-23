package log

import (
	"fmt"
	"os"
	"time"
	"path/filepath"
)

var (
	LOGDIR string = "logs"
	LOGNAME string = "log"
)

func textToLine(text []interface{}) string {
	var line string = ""
	for _, s := range text {
		line += fmt.Sprint(s)
	}
	return line
}

func logWrite(level string, line string) {
	var now time.Time = time.Now()
	var datetime string = now.Format("2006-01-02 15:04:05.000")
	var formatted_line string = fmt.Sprintf("%s : %s : %s\n", datetime, level, line)
	fmt.Printf(formatted_line)

	var date string = now.Format("2006-01-02")
	var logfile string = fmt.Sprintf("%s-%s.log", date, LOGNAME)

	f, err := os.OpenFile(filepath.Join(LOGDIR, logfile), os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		fmt.Println(err)
	}
	defer f.Close()
	_, err = f.WriteString(formatted_line)
	if err != nil {
		fmt.Println(err)
	}
}

func Debug(text ...any) {
	logWrite("DEBUG", textToLine(text))
}

func Info(text ...any) {
	logWrite("INFO", textToLine(text))
}

func Warning(text ...any) {
	logWrite("WARNING", textToLine(text))
}

func Error(text ...any) {
	logWrite("ERROR", textToLine(text))
}

func Init(logdir string, logname string) {
	LOGDIR = logdir
	LOGNAME = logname
	err := os.Mkdir(logdir, 0666)
	if err != nil {
		fmt.Println(err)
	}
}
