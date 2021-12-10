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

func SendResetPWMail(){
  fmt.Println("SendResetPWMail!")
 // Receiver email address.
  to := []string{
    "lowilson180@gmail.com",
  }
  fmt.Println("SendResetPWMail222222222")
  // Message.
  message := []byte("This is a test email message.")
  fmt.Println("SendResetPWMail333333333")
  // Authentication.
  auth := smtp.PlainAuth("", from, password, smtpHost)
  // Sending email.
  fmt.Println("SendResetPWMail444444444")
  err := smtp.SendMail(smtpHost+":"+smtpPort, auth, from, to, message)
   fmt.Println("SendResetPWMail555555555")
  if err != nil {
    fmt.Println("error = ", err)
    return
  }

  fmt.Println("Email Sent Successfully!")
}

/*func SendMail(mailTo []string,subject string, body string ) error {
 fmt.Println("SendMail")
  //定義郵箱伺服器連線資訊，如果是阿里郵箱 pass填密碼，qq郵箱填授權碼
    mailConn := map[string]string {
        "user": from,
        "pass": password,
        "host": smtpHost,
        "port": smtpPort,
    }

    port, _ := strconv.Atoi(mailConn["port"]) //轉換埠型別為int

    m := gomail.NewMessage()
    m.SetHeader("From","XD Game" + "<" + mailConn["user"] + ">")  //這種方式可以新增別名，即“XD Game”， 也可以直接用<code>m.SetHeader("From",mailConn["user"])</code> 讀者可以自行實驗下效果
    m.SetHeader("To", mailTo...)  //傳送給多個使用者
    m.SetHeader("Subject", subject)  //設定郵件主題
    m.SetBody("text/html", body)     //設定郵件正文

    d := gomail.NewDialer(mailConn["host"], port, mailConn["user"], mailConn["pass"])
    fmt.Println("SendMail11111")
    err := d.DialAndSend(m)
    fmt.Println("SendMail222222")
    if(err != nil){
       fmt.Println("error ", err.Error())
    }else{
       fmt.Println("success")
    }
    return err

}*/

func SendGmail(){
  fmt.Println("SendGmail!")
   // Configuration
   	from := "lowilson180@gmail.com"
   	password := "XJIJ0vuXJIJ"
   	to := []string{"lowilson180@gmail.com"}
   	smtpHost := "smtp.gmail.com"
   	smtpPort := "587"

   	message := []byte("My super secret message.")

   	// Create authentication
   	auth := smtp.PlainAuth("", from, password, smtpHost)

   	// Send actual message
   	 fmt.Println("SendGmail222")
   	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, from, to, message)
    fmt.Println("SendGmail333")
   	if err != nil {
   		 fmt.Println("error ", err.Error())
   	}else{
         fmt.Println("send success")
   	}
}


func SendMailTest(){

	   fmt.Println("SendMailTest~~~~~~~~~~~")

    	var mailTo []string
    	mailTo = strings.Split("lowilson180@gmail.com;", ";")
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

    	//msg := []byte("astpi4 test")

    	header := make(map[string]string)
        header["Subject"] = "Aspeed System notification"
        header["MIME-Version"] = "1.0"
        header["Content-Type"] = "text/plain; charset=\"utf-8\""
        header["Content-Transfer-Encoding"] = "base64"
        message := ""
        for k, v := range header {
            message += fmt.Sprintf("%s: %s\r\n", k, v)
        }
        message += "\r\n" + base64.StdEncoding.EncodeToString([]byte("Hi Wilson, Test"))

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


