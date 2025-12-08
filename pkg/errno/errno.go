package errno

import "fmt"

type AppErro struct {
	Code  int    `json:"code"`
	Msg   string `json:"msg"`
	Cause error  `json:"-"`
}

func (e *AppErro) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("code: %d, msg: %s, cause: %v", e.Code, e.Msg, e.Cause)
	}
	return fmt.Sprintf("code: %d, msg: %s", e.Code, e.Msg)
}

func (e *AppErro) Unwrap() error {
	return e.Cause
}

// 此处使用值接收方法，类似拷贝后再修改，防止全局错误被篡改
//
// 包装一个错误类型
func (e AppErro) WithCause(err error) *AppErro {
	return &AppErro{Cause: err}
}

func New(code int, msg string) *AppErro {
	return &AppErro{Code: code, Msg: msg}
}
