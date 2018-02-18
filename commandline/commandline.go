package commandline

import (
	"os"
	"reflect"
	"log"
	"strings"
)

const (
	signature             		int = 10000000
	ENoError                    int = 0
	eNotEnoughArgs 				int = signature + 1
	eNotEnoughFieldsInStruct 	int = signature + 2
)

var errorStrings [2]string

func GetErrorString(ecode int) (string) {
	return errorStrings[ecode-signature-1]
}

func ProcessCommandline(it interface{},cmdcnt int, examplestr string) {
	if (len(os.Args) != cmdcnt) {
		log.Fatal(GetErrorString(eNotEnoughArgs)+examplestr)
	}
	val := reflect.ValueOf(it).Elem()
	numfields := val.NumField()
	if (numfields != cmdcnt) {
		log.Fatal(GetErrorString(eNotEnoughFieldsInStruct)+examplestr)
	}
	for i := 0; i < numfields; i++ {
		val.FieldByName(val.Type().Field(i).Name).SetString(strings.ToLower(os.Args[i]))
	}
}

func init() {
	errorStrings[0] = "Not enough arguments\r\n";
	errorStrings[1] = "Not enough fields in struct\r\n";
}

