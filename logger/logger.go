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
	//check if the log file is ready or start a new one
	filename := l.Filename
	_, err := os.OpenFile(filename, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		return fmt.Errorf(err.Error() + " in logger.openExistingOrNew")
	}
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return l.openNew()
	}
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		// if we fail to open the old log file for some reason, just ignore
		// it and open a new log file.
		return l.openNew()
	}
	l.file = file
	l.size = info.Size()
	return nil
}

func (l *logger) openNew() error {
	filename := l.Filename
	_, err := os.OpenFile(filename, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		return fmt.Errorf(err.Error() + " in Logger.openNew")
	}
	return nil
}

func Info(message interface{}) {
	fmt.Println(White + message.(string) + Reset)
	log.Println("INFO: " + message.(string))
}

func Error(message interface{}) {
	fmt.Println(Red + message.(string) + Reset)
	log.Println("ERROR: " + message.(string))
}

func Warning(message interface{}) {
	fmt.Println(Yellow + message.(string) + Reset)
	log.Println("WARN : " + message.(string))
}

func Success(message interface{}) {
	fmt.Println(Green + message.(string) + Reset)
	log.Println("SUCCESS: " + message.(string))
}

func RuntimeError(message interface{}) {
	fmt.Println(Red + message.(string) + Reset)
}

func RuntimeInfo(message interface{}) {
	fmt.Println(White + message.(string) + Reset)
}
