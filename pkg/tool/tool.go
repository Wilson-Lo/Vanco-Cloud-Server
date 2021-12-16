package tool

import (
	"fmt"
	"io"
	"math/rand"
	"net/mail"
	"crypto/md5"
	b64 "encoding/base64"
	e "app/pkg/e"
)
const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

//Valid E-mail
func ValidEmail(email string) bool {
    _, err := mail.ParseAddress(email)
    return err == nil
}

//string to MD5
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
*
*/
func RandStringBytes(n int) string {
    b := make([]byte, n)
    for i := range b {
        b[i] = letterBytes[rand.Intn(len(letterBytes))]
    }
    return string(b)
}