package JKString

import (
	"fmt"
	"reflect"
	"strings"
	"time"
	"github.com/satori/go.uuid"
	"golang.org/x/crypto/bcrypt"
	"net"
	"errors"
	json "encoding/json"
	"strconv"
	"net/http"
	b64 "encoding/base64"
)

type THashSignature struct {
	Uuid uuid.UUID          `json:"uuid"`
	Email string            `json:"email"`
	Dt time.Time            `json:"timestamp"`
	Serverip net.IP         `json:"serverip"`
}

func LocalIPToString() (net.IP, error) {
	tt, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	for _, t := range tt {
		aa, err := t.Addrs()
		if err != nil {
			return nil, err
		}
		for _, a := range aa {
			ipnet, ok := a.(*net.IPNet)
			if !ok {
				continue
			}
			v4 := ipnet.IP.To4()
			if v4 == nil || v4[0] == 127 { // loopback address
				continue
			}
			return v4, nil
		}
	}
	return nil, errors.New("cannot find local IP address")
}


func Empty(aStr string) bool {
	return (aStr == "")
}

func Commaindex(aStr string, idx int, ubound int) (res string) {
	res = aStr
	if (idx + 1) < ubound {
		res = res + ","
	}
	return
}

func EncloseIn(aStr string, eStr string) (res string) {
	res = eStr + aStr + eStr
	return
}

func OrNULL(aStr string) (res string) {
	res = aStr
	if res == "<nil>" {
		res = "NULL"
	}
	return
}

func ToString(x interface{}) string {
	switch y := x.(type) {

	// Handle dates with special logic
	// This needs to come above the fmt.Stringer
	// test since time.Time's have a .String()
	// method
	case time.Time:
		return y.Format(time.RFC3339)

	// Handle type string
	case string:
		return y

	// Handle type with .String() method
	case fmt.Stringer:
		return y.String()

	// Handle type with .Error() method
	case error:
		return y.Error()

	}

	// Handle named string type
	if v := reflect.ValueOf(x); v.Kind() == reflect.String {
		return v.String()
	}

	// Fallback to fmt package for anything else like numeric types
	return fmt.Sprint(x)
}

func QuoteStringIfNeeded(x interface{}) string {
	// quote only if non numeric
	switch y := x.(type) {

	// Handle dates with special logic
	// This needs to come above the fmt.Stringer
	// test since time.Time's have a .String()
	// method
	case time.Time:
		return UTF8SingleQuoted(y.Format(time.RFC3339))

	// Handle type string
	case string:
		return UTF8SingleQuoted(y)

	// Handle type with .String() method
	case fmt.Stringer:
		return UTF8SingleQuoted(y.String())

	// Handle type with .Error() method
	case error:
		return y.Error()

	}

	// Handle named string type
	if v := reflect.ValueOf(x); v.Kind() == reflect.String {
		return UTF8SingleQuoted(v.String())
	}

	// Fallback to fmt package for anything else like numeric types
	return fmt.Sprint(x)
}

func UTF8SingleQuoted(aStr string) (res string) {
	res = "N'" + aStr + "'"
	return
}

func SpaceTZ(aStr string) (res string) {
	// convert RFC3339 to mssql edible timestamp
	// this needs some validation work
	res = aStr
	res = strings.Replace(res, "T", " ", 1)
	res = strings.Replace(res, "Z", " ", 1)
	return
}

func AddSingleQuotes(aStr string) (res string) {
	res = aStr
	if !strings.HasPrefix(aStr, "'") {
		res = "'" + res
	}
	if !strings.HasSuffix(aStr, "'") {
		res += "'"
	}
	return
}

func AddDoubleQuotes(aStr string) (res string) {
	res = aStr
	if !strings.HasPrefix(aStr, `"`) {
		res = `"` + res
	}
	if !strings.HasSuffix(aStr, `"`) {
		res += `"`
	}
	return
}

func FixConnectionString(aStr string) (res string) {
	res = strings.Replace(strings.ToLower(aStr), "provider=sqloledb.1;integrated security=sspi;", "Driver={SQL Server Native Client 11.0};", 1)
	res = strings.Replace(res, "data source=", "Server=", 1)
	return
}

const idsalt string = "bfdf9f4f-1330-4916-a333-b83604707ea1"

func MakeIDAndB64Hash() (id string, hash string) {
	var (err error
		bhash []byte)
	id = uuid.Must(uuid.NewV4()).String()
	bhash,err = bcrypt.GenerateFromPassword([]byte(id),bcrypt.DefaultCost)
	if err != nil {
		hash = ""
		id = ""
	} else {
		hash = b64.StdEncoding.EncodeToString(bhash)
	}
	return
}

func MakeIDAndB64HashExtended(email string) (id string, hash string) {
	var (err error
		bhash []byte
		bsig []byte
		sig THashSignature)
	id = ""
	hash = ""

	sig.Uuid = uuid.Must(uuid.NewV4())
	sig.Dt = time.Now().UTC()
	sig.Email = email
	sig.Serverip,err = LocalIPToString()
	if err != nil {return}
	bsig,err = json.Marshal(sig)
	if err != nil {return}
	bhash,err = bcrypt.GenerateFromPassword(bsig,bcrypt.DefaultCost)
	if err != nil {
		hash = ""
		id = ""
	} else {
		id = string(bsig)
		hash = b64.StdEncoding.EncodeToString(bhash)
	}
	return
}

func FmtHttpError(err error,httpstatus int) (string) {
	return strconv.Itoa(httpstatus)+":"+http.StatusText(httpstatus)+" "+fmt.Sprintf("%s",err)
}