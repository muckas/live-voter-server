package log

import (
	"fmt"
	"os"
	"time"
	"path/filepath"
)

var (
	LOGDIR string
	LOGNAME string
)

func logWrite(line string, level string) {
	var now time.Time
	var date, datetime, formatted_line, logfile string
	now = time.Now()
	datetime = now.Format("2006-01-02 15:04:05.000")
	formatted_line = fmt.Sprintf("%s : %s : %s\n", datetime, level, line)
	fmt.Printf(formatted_line)

	date = now.Format("2006-01-02")
	logfile = fmt.Sprintf("%s-%s.log", date, LOGNAME)

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

func Debug(line string) {
	logWrite(line, "DEBUG")
}

func Info(line string) {
	logWrite(line, "INFO")
}

func Warning(line string) {
	logWrite(line, "WARNING")
}

func Error(line string) {
	logWrite(line, "ERROR")
}

func Init(logdir string, logname string) {
	LOGDIR = logdir
	LOGNAME = logname
	err := os.Mkdir(logdir, 0666)
	if err != nil {
		fmt.Println(err)
	}
}
