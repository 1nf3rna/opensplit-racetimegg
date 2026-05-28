package logging

import "log"

type Logger struct {
	debug bool
}

func NewLogger(debug bool) *Logger {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds | log.Lshortfile)

	return &Logger{
		debug: debug,
	}
}

func (l *Logger) Debug(component string, format string, v ...any) {
	if !l.debug {
		return
	}

	log.Printf("[DEBUG] "+component+": "+format, v...)
}

func (l *Logger) Info(component string, format string, v ...any) {
	log.Printf("[INFO] "+component+": "+format, v...)
}

func (l *Logger) Warn(component string, format string, v ...any) {
	log.Printf("[WARN] "+component+": "+format, v...)
}

func (l *Logger) Error(component string, format string, v ...any) {
	log.Printf("[ERROR] "+component+": "+format, v...)
}
