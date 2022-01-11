package tool

import (
	"fmt"
	"app/models"
	"io"
	"time"
	"strings"
	"math/rand"
	"net/mail"
	"crypto/md5"
	b64 "encoding/base64"
	e "app/pkg/e"
)
const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

/**
* Valid E-mail
*/
func ValidEmail(email string) bool {
    _, err := mail.ParseAddress(email)
    return err == nil
}

/**
*  string to MD5
*/
func ToMD5(str string) string  {
    w := md5.New()
    io.WriteString(w, str)
    md5str := fmt.Sprintf("%x", w.Sum(nil))
    return md5str
}

/**
*    Base64 + Salt
*/
func EncryptionData(bodyData string) string{
     base64String := b64.StdEncoding.EncodeToString([]byte(bodyData))
     return (e.SaltFirst + base64String + e.SaltAfter)
}

/**
*  Get sign value
*/
func GetSign(data models.Command) string{
     var allData = "body="+data.Body+"&etag="+data.Etag+"&extra="+data.Extra+"&method="+data.Method+"&time="+data.Time+"&to="+data.To
     return ToMD5(allData)
}

/**
* Random MD5
*/
func RandStringBytes(n int) string {
    b := make([]byte, n)
    for i := range b {
        b[i] = letterBytes[rand.Intn(len(letterBytes))]
    }
    return string(b)
}

/**
* Time format
*/
func ShortDur(d time.Duration) string {
    s := d.String()
    if strings.HasSuffix(s, "m0s") {
        s = s[:len(s)-2]
    }
    if strings.HasSuffix(s, "h0m") {
        s = s[:len(s)-2]
    }
    return s
}

/**
*  Remove last word of the string
*/
func RemoveLastRune(s string) string {
    r := []rune(s)
    return string(r[:len(r)-1])
}

/**
*  Get Time Now
*/
func GetTimeNow() string{
   dt := time.Now()
   formatted := fmt.Sprintf("%d-%02d-%02d %02d:%02d:%02d%",
               dt.Year(), dt.Month(), dt.Day(), dt.Hour(), dt.Minute(), dt.Second())
   return formatted
}