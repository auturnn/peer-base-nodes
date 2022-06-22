//Package utils contains function to be used across the app
//utils는 app전역에서 사용되는 기능들을 포함한다.
package utils

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"log"
	"strings"
)

//unit test시에도 내용을 변경하여 정상운용 가능하도록 설정
var logPn = log.Panic

//HandleError error 변수를 받아 nil이 아닐경우 log.Panic(err)을 발생시킨다.
func HandleError(err error) {
	if err != nil {
		logPn(err)
	}
}

//Splitter s는 잘라낼 대상 문자열, sep은 잘라낼 기준이 되는 문자, i는 변환된 slice에서 가져올 부분이다.
//strings.Split(s,sep)를 통해 반환된 문자열 slice의 i번째 문자열을 반환한다.
func Splitter(s string, sep string, i int) string {
	r := strings.Split(s, sep)
	if len(r)-1 < i {
		return ""
	}
	return r[i]
}

//ToBytes interface를 받고, Buffer를 이용하여 byte로 encoding한 뒤 반환한다.
func ToBytes(i interface{}) []byte {
	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)
	HandleError(encoder.Encode(i))
	return buf.Bytes()
}

//FromBytes interface와 byte로 이루어진 data를 가져와서 data를 encode하고 interface에 전달한다.
func FromBytes(i interface{}, data []byte) {
	decoder := gob.NewDecoder(bytes.NewReader(data))
	HandleError(decoder.Decode(i))
}

//Hash takes an interface and returns the hex encoding of the hash.
//Hash는 interface를 받아 sha256으로 encoding하여 string을 반환한다.
func Hash(i interface{}) string {
	s := fmt.Sprintf("%v", i)
	hash := sha256.Sum256([]byte(s))
	return fmt.Sprintf("%x", hash)
}

//ToJSON interface를 받아 json으로 Marshal하여 byte로 반환한다.
func ToJSON(i interface{}) []byte {
	r, err := json.Marshal(i)
	HandleError(err)
	return r
}
