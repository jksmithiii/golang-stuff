package Token

import (
	"time"
	"crypto/aes"
	"crypto/cipher"
	"encoding/json"
	b64 "encoding/base64"
	"math"
	"crypto/sha256"
	"DBServices"
	"JKString"
	"errors"
)

var (
	dbConn DBServices.TDBConnection  //"user:password@tcp(127.0.0.1:3306)/database"
)


type THeaderToken struct {
	ApiKey 			  string     `json:"apikey"`
	LoginTimeStamp    time.Time  `json:"logintimestamp"`
	LastActivityStamp time.Time  `json:"timestamp"`
	Timeout 		  int        `json:"timeout"`
}

const DefaultTimeOut int = 30*60
const key string = "123456789123456789"
const ExpiredTokensTable string = "ExpiredTokens"
var iv = []byte{35, 46, 57, 24, 85, 35, 24, 74, 87, 35, 88, 98, 66, 32, 14, 05}

func hashKey() ([]byte) {
	h := sha256.New()
	h.Write([]byte(key))
	return h.Sum(nil)
}

func SetConnection(driver,dbserver,dbname string) {
	dbConn.Set(driver,dbserver+"/"+dbname)
}

func (h *THeaderToken) Set(apikey string) {
	h.ApiKey = apikey
	h.LoginTimeStamp = time.Now().UTC()
	h.LastActivityStamp = h.LoginTimeStamp
	h.Timeout = DefaultTimeOut
}

func (h *THeaderToken) Encrypt() (string,error) {
	var (
		block cipher.Block
		err error
	)
	btoken,err := json.Marshal(h)
	if err != nil { return "",err }

	block, err = aes.NewCipher(hashKey())
	if err != nil { return "",err }
	plaintext := []byte(btoken)
	cfb := cipher.NewCFBEncrypter(block, iv)
	ciphertext := make([]byte, len(plaintext))
	cfb.XORKeyStream(ciphertext, plaintext)
	return b64.StdEncoding.EncodeToString(ciphertext),err
}

func (h *THeaderToken) Decrypt(token string) error {
	var ciphertext []byte
	block, err := aes.NewCipher(hashKey())
	if err != nil { return err }
	ciphertext,err = b64.StdEncoding.DecodeString(token)
	if err != nil { return err }
	cfb := cipher.NewCFBEncrypter(block, iv)
	plaintext := make([]byte, len(ciphertext))
	cfb.XORKeyStream(plaintext, ciphertext)
	err = json.Unmarshal([]byte(plaintext),&h)
	return err
}

func (h *THeaderToken) Expired(token string) (bool,error) {
	err := h.Decrypt(token)
	if err != nil { return false,err }
	d:= time.Since(h.LastActivityStamp)
	return int(math.Trunc(d.Seconds())) > DefaultTimeOut,err
}

func (h *THeaderToken) Update(token string) (string,error) {
	err := h.Decrypt(token)
	if err != nil { return "",err }
	h.LastActivityStamp = time.Now().UTC()
	return h.Encrypt()
}

func KillToken(token string) (error){
	if (token == "") {return errors.New("1:token is empty")}
	_, err := dbConn.GetDB().Exec("INSERT INTO "+ExpiredTokensTable+ " (token) VALUES ("+JKString.AddDoubleQuotes(token)+")")
	dbConn.SetLastError(err)
	return err
}

func TokenIsDead(token string) (bool) {
	var id int
	if (token == "") {return true} // dead = (empty string or in list )
	row := dbConn.GetDB().QueryRow("SELECT ID FROM "+ExpiredTokensTable+" WHERE token = "+JKString.AddDoubleQuotes(token))
	err := row.Scan(&id)
	return (err == nil) //token in dead list
}

