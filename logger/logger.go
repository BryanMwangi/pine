package logger

import (
	"fmt"
	"log"
	"os"
	"time"
)

// we define the logger object that we will be using to log files
type logger struct {
	Filename  string
	MaxSize   int64
	LocalTime time.Time
	Size      int64
	file      *os.File
	size      int64
}

var (
	megabyte       = 1024 * 1024
	defaultMaxSize = 100
)

var (
	Reset  = "\033[0m"
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	White  = "\033[97m"
)

// we initialise this in the target project by calling logger.init and passing in
// the parameters required to store the log data
func Init(fileName string, maxSize int64) error {
	//if the file exists we continue set up to ensure all logs are written in the
	//suggested file
	log.SetOutput(&logger{
		Filename: fileName,
		MaxSize:  maxSize,
	})
	return nil
}

// a custom io writer that will write the log data
func (l *logger) Write(p []byte) (n int, err error) {
	writeLen := int64(len(p))
	if writeLen > l.max() {
		return 0, fmt.Errorf(
			"write length %d exceeds maximum item size %d", writeLen, l.max(),
		)
	}
	if l.file == nil {
		if err = l.openExistingOrNew(); err != nil {
			return 0, err
		}
	}
	n, err = l.file.Write(p)
	l.size += int64(n)

	return n, err
}

func (l *logger) max() int64 {
	if l.MaxSize == 0 {
		return int64(defaultMaxSize * megabyte)
	}
	return int64(l.MaxSize) * int64(megabyte)
}

func (l *logger) openExistingOrNew() error {
	info, err := os.Stat(l.Filename)
	if os.IsNotExist(err) {
		return l.openNew()
	}
	if err != nil {
		return l.openNew()
	}
	file, err := os.OpenFile(l.Filename, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return l.openNew()
	}
	l.file = file
	l.size = info.Size()
	return nil
}

func (l *logger) openNew() error {
	file, err := os.OpenFile(l.Filename, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		return fmt.Errorf("%s in Logger.openNew", err)
	}
	l.file = file
	l.size = 0
	return nil
}

func Info(message ...any) {
	fmt.Println(White + fmt.Sprint(message...) + Reset)
	log.Println("INFO: " + fmt.Sprint(message...))
}

func Error(message ...any) {
	fmt.Println(Red + fmt.Sprint(message...) + Reset)
	log.Println("ERROR: " + fmt.Sprint(message...))
}

func Warning(message ...any) {
	fmt.Println(Yellow + fmt.Sprint(message...) + Reset)
	log.Println("WARN : " + fmt.Sprint(message...))
}

func Success(message ...any) {
	fmt.Println(Green + fmt.Sprint(message...) + Reset)
	log.Println("SUCCESS: " + fmt.Sprint(message...))
}

func RuntimeError(message ...any) {
	fmt.Println(Red + fmt.Sprint(message...) + Reset)
	log.Println("RUNTIME ERROR: " + fmt.Sprint(message...))
}

func RuntimeInfo(message ...any) {
	fmt.Println(White + fmt.Sprint(message...) + Reset)
}
