package gosnake

import "log"

func infoln(v ...interface{}) {
	log.Println(v...)
}

func infof(format string, v ...interface{}) {
	log.Printf(format, v...)
}

func fatalln(v ...interface{}) {
	log.Fatalln(v...)
}

func fatalf(format string, v ...interface{}) {
	log.Fatalf(format, v...)
}

func panicln(v ...interface{}) {
	log.Panicln(v...)
}

func panicf(format string, v ...interface{}) {
	log.Panicf(format, v...)
}
