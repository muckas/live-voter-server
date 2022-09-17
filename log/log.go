package log

import (
	"fmt"
	"os"
	"time"
	"path/filepath"
	"reflect"
	"runtime"
)

var (
	LOGDIR string = "logs"
	LOGNAME string = "log"
)

func TraceCaller(skip int) string {
	pc := make([]uintptr, 15)
	n := runtime.Callers(skip, pc) // Callers( <number of callers back>, pc )
	frames := runtime.CallersFrames(pc[:n])
	frame, _ := frames.Next()
	return frame.Function
}

func textToLine(text []interface{}) string {
	var prev_string, is_string bool
	var line string = ""
	prev_string = false
	for arg_num, arg := range text {
		is_string = arg != nil && reflect.TypeOf(arg).Kind() == reflect.String
		if arg_num > 0 && !is_string && !prev_string {
			line += " "
		}
		line += fmt.Sprintf("%+v", arg)
		prev_string = is_string
	}
	return line
}

func logWrite(level string, line string) {
	var now time.Time = time.Now()
	var datetime string = now.Format("2006-01-02 15:04:05.000")
	var formatted_line string = fmt.Sprintf("%s : %s : %s : %s\n", datetime, level, TraceCaller(4), line)
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
