package pkg

import (
	"github.com/sirupsen/logrus"
)

// Notifier 通知接口
type Notifier interface {
	SendInfo(title, message string)
	SendWarning(title, message string)
	SendFatalError(operation string, err error)
}

// SendDingTalkWarning 发送钉钉警告通知
func SendDingTalkWarning(title, message string) {
	// 这里需要延迟导入config包，避免循环依赖
	// 在实际使用中，这个函数会被替换为直接调用
	logrus.Warnf("DingTalk warning notification not implemented in notifier.go to avoid circular import")
}

// SendDingTalkInfo 发送钉钉信息通知
func SendDingTalkInfo(title, message string) {
	// 这里需要延迟导入config包，避免循环依赖
	// 在实际使用中，这个函数会被替换为直接调用
	logrus.Warnf("DingTalk info notification not implemented in notifier.go to avoid circular import")
}

// SendDingTalkFatalError 发送钉钉致命错误通知
func SendDingTalkFatalError(operation string, err error) {
	// 这里需要延迟导入config包，避免循环依赖
	// 在实际使用中，这个函数会被替换为直接调用
	logrus.Warnf("DingTalk fatal error notification not implemented in notifier.go to avoid circular import")
}
