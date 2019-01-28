#smtp
# 优化点
## 新增LoginAuth md5 认证方式
默认只支持plain的登陆方式，本次新增了LOGIN，CRAM-MD5两种认证
## 优化tls登陆
默认的做法为建立一个tls的连接，但其实在exchange这类邮件服务上是无法使用的，最兼容的做法是在调用STARTTLS手工开启tls  
## 提升了兼容性
参考了py的smtp包，三种登陆方式都会尝试，保证了对exchange与postfix两种邮件服务的兼容
## demo

```go
package main

import (
	"log"

	"github.com/toolkits/smtp"
)

func main() {
	//s := smtp.New("smtp.exmail.qq.com:25", "notify@a.com", "password")
	s := smtp.NewSMTP("smtp.exmail.qq.com:25", "notify@a.com", "password",false,false,false)
	log.Println(s.SendMail("notify@a.com", "ulric@b.com;rain@c.com", "这是subject", "这是body,<font color=red>red</font>"))
}
```