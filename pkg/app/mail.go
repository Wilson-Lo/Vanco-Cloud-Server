package app

import (
     "fmt"
     "net/smtp"
     "crypto/tls"
     "net"
     "strings"
     "errors"
     "encoding/base64"
    // "io/ioutil"
    //  "os/exec"
    // "gopkg.in/gomail.v2"
    // "strconv"
    // "os"
)

var from = "gomax_app@gomax-electronics.com.tw"
var password = "T6ba_4u4"
var smtpHost = "mse.gomax-electronics.com.tw"
var smtpPort = "25"

func SendMail(userMail string, url_link string){

	   fmt.Println("SendMail~~~~~~~~~~~")

    	var mailTo []string
    	mailTo = strings.Split( userMail + ";", ";")
    	mailTo = mailTo[:len(mailTo)-1]
    	fmt.Println(mailTo,len(mailTo))

    	if len(mailTo)==0{
    		return
    	}

        //addr := "mse.gomax-electronics.com.tw:25"
    	c, err := smtp.Dial("mse.gomax-electronics.com.tw:25")

    	if err != nil {
            fmt.Println("smtp.Dial error = " , err.Error())
            return
        }

        host, _, _ := net.SplitHostPort("mse.gomax-electronics.com.tw:25")

        defer c.Close()
    	if ok, _ := c.Extension("STARTTLS"); ok {
            config := &tls.Config{ServerName: host, InsecureSkipVerify: true}
            if err = c.StartTLS(config); err != nil {
                fmt.Println("call start tls")
                return
            }
        }

    	auth := NewLoginAuth("pi4control@gomax-electronics.com.tw","Iym#g052")

    	if ok, _ := c.Extension("AUTH"); ok {
                if err = c.Auth(auth); err != nil {
                    fmt.Println("check auth with err:", err)
                    return
                }
        }

    	from := "pi4control@gomax-electronics.com.tw"

    	if err = c.Mail(from); err != nil {
    	    fmt.Println("mail err:", err)
            return
        }

    	//to := []string{"pokemongoplayer20160817@gmail.com","kdjljfsjf@gmail.com"}
    	//fmt.Println(to)
    	//mailTo
        for _, addr := range mailTo {
            if err = c.Rcpt(addr); err != nil {
    			 fmt.Println("range err:", err)
                return
            }
        }
        w, err := c.Data()
        if err != nil {
    		fmt.Println("Data err:", err)
            return
        }

    	msg := []byte(url_link)

    	header := make(map[string]string)
        header["Subject"] = "Cloud System - Reset Password Mail"
        header["MIME-Version"] = "1.0"
        header["Content-Type"] = "text/plain; charset=\"utf-8\""
        header["Content-Transfer-Encoding"] = "base64"
        message := ""
        for k, v := range header {
            message += fmt.Sprintf("%s: %s\r\n", k, v)
        }
        message += "\r\n" + base64.StdEncoding.EncodeToString(msg)

        _, err = w.Write([]byte(message))

    	if err != nil {
    		fmt.Println("write err:", err)
            return
        }
        err = w.Close()
        if err != nil {
            return
        }
        c.Quit()
}

type LoginAuth struct {
    username, password string
}

func (a *LoginAuth) Start(server *smtp.ServerInfo) (string, []byte, error) {
    return "LOGIN", []byte{}, nil
}

func NewLoginAuth(username, password string) smtp.Auth {
    return &LoginAuth{username, password}
}

func (a *LoginAuth) Next(fromServer []byte, more bool) ([]byte, error) {
    if more {
        switch string(fromServer) {
        case "Username:":
            return []byte(a.username), nil
        case "Password:":
            return []byte(a.password), nil
        default:
            return nil, errors.New("Unknown fromServer")
        }
    }
    return nil, nil
}


