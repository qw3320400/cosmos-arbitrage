package common

import (
	"fmt"
	"runtime"
	"strconv"
	"time"
)

func Log(msg string) {
	_, file, line, _ := runtime.Caller(1)
	fmt.Println(time.Now().Format("2006-01-02 15:04:05"), file+":"+strconv.Itoa(line)+" "+msg)
}

func LogErr(msg string) {
	_, file, line, _ := runtime.Caller(1)
	fmt.Println(time.Now().Format("2006-01-02 15:04:05"), "{Error}", file+":"+strconv.Itoa(line)+" "+msg)
}

func StringIsIn(in string, arg ...string) bool {
	for _, v := range arg {
		if in == v {
			return true
		}
	}
	return false
}

func Errorf(err error, s string, arg ...interface{}) error {
	_, file, line, _ := runtime.Caller(1)
	if err != nil {
		return fmt.Errorf("\n\t"+err.Error()+"\n\t"+file+":"+strconv.Itoa(line)+" "+s, arg...)
	}
	return fmt.Errorf("\n\t"+file+":"+strconv.Itoa(line)+" "+s, arg...)
}
