package smtp

import (
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"net/smtp"
	"strings"

)

type Smtp struct {
	Address    string
	Username   string
	Password   string
	TLS        bool
	Anonymous  bool
	SkipVerify bool
}

//Compatible with old API
func New(address, username, password string) *Smtp {
	return &Smtp{
		Address:  address,
		Username: username,
		Password: password,
	}
}

//Support TLS and Anonymous
func NewSMTP(address, username, password string, tls, anonymous, skipVerify bool) *Smtp {
	return &Smtp{
		Address:    address,
		Username:   username,
		Password:   password,
		TLS:        tls,
		Anonymous:  anonymous,
		SkipVerify: skipVerify,
	}
}

func (this *Smtp) SendMail(from, tos, subject, body string, contentType ...string) error {
	var preferred_auths = [] string{"LOGIN", "PLAIN", "CRAM-MD5"}
	if this.Address == "" {
		return fmt.Errorf("address is necessary")
	}

	hp := strings.Split(this.Address, ":")
	if len(hp) != 2 {
		return fmt.Errorf("address format error")
	}

	arr := strings.Split(tos, ";")
	count := len(arr)
	safeArr := make([]string, 0, count)
	for i := 0; i < count; i++ {
		if arr[i] == "" {
			continue
		}
		safeArr = append(safeArr, arr[i])
	}

	if len(safeArr) == 0 {
		return fmt.Errorf("tos invalid")
	}

	tos = strings.Join(safeArr, ";")

	b64 := base64.NewEncoding("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/")

	header := make(map[string]string)
	header["From"] = from
	header["To"] = tos
	header["Subject"] = fmt.Sprintf("=?UTF-8?B?%s?=", b64.EncodeToString([]byte(subject)))
	header["MIME-Version"] = "1.0"

	ct := "text/plain; charset=UTF-8"
	if len(contentType) > 0 && contentType[0] == "html" {
		ct = "text/html; charset=UTF-8"
	}

	header["Content-Type"] = ct
	header["Content-Transfer-Encoding"] = "base64"

	message := ""
	for k, v := range header {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + b64.EncodeToString([]byte(body))

	var auth smtp.Auth = nil
	tlsconfig := &tls.Config{
		InsecureSkipVerify: this.SkipVerify,
		ServerName:         hp[0],
	}
	var loginSuccess = false
	// 修改login方式
	var c *smtp.Client
	var err error
	for _, method := range preferred_auths{
		switch method {
		case "PLAIN":
			auth = smtp.PlainAuth("", this.Username, this.Password, hp[0])
		case "CRAM-MD5":
			auth = smtp.CRAMMD5Auth(this.Username, this.Password)
		case "LOGIN":
			auth = &loginAuth{username: this.Username, password: this.Password, host: hp[0]}
		}
		// 生成connection
		c, err = smtp.Dial(this.Address)
		if this.TLS {
			if err = c.StartTLS(tlsconfig); err != nil {
				return err
			}
		}
		// 如果登陆失败
		if err = c.Auth(auth); err != nil {
			continue
		}else {
			loginSuccess = true
			break
		}

	}
	if loginSuccess {
		return sendMail(c, from, strings.Split(tos, ";"), []byte(message), this.SkipVerify)
	}else {
		return fmt.Errorf("Login failed: %v", err.Error())
	}

}

func sendMail(c *smtp.Client, from string, to []string, msg []byte, skipVerify bool) error {
	var err error
	if err = validateLine(from); err != nil {
		return err
	}
	for _, recp := range to {
		if err := validateLine(recp); err != nil {
			return err
		}
	}

	defer c.Close()
	if err = c.Mail(from); err != nil {
		return err
	}
	for _, addr := range to {
		if err = c.Rcpt(addr); err != nil {
			return err
		}
	}
	w, err := c.Data()
	if err != nil {
		return err
	}
	_, err = w.Write(msg)
	if err != nil {
		return err
	}
	err = w.Close()
	if err != nil {
		return err
	}
	return c.Quit()
}

func validateLine(line string) error {
	if strings.ContainsAny(line, "\n\r") {
		return fmt.Errorf("smtp: A line must not contain CR or LF")
	}
	return nil
}
