package logger

import (
	"context"
	"log/slog"
	"os"
	"time"
)

// 上下文键类型
type ctxKey string

// 上下文键常量
const (
	LoggerKey    ctxKey = "logger"     // 日志键
	RequestIDKey ctxKey = "request_id" // 请求ID键
)

type Env string

const (
	EnvDev  Env = "dev"
	EnvProd Env = "prod"
)

// 定义日志接口
// 定义日志接口，包含以下方法：
// - Info：记录信息级别日志
// - Error：记录错误级别日志
// - Warn：记录警告级别日志
// - Debug：记录调试级别日志
// - Fatal：记录严重错误级别日志
// - With：添加上下文信息
// 其中，Info、Error、Warn、Debug、Fatal 方法接收上下文和格式化字符串，支持可变参数。
// With 方法接收任意类型的键值对，用于添加上下文信息。
type Logger interface {
	Info(ctx context.Context, msg string, args ...any)
	Error(ctx context.Context, msg string, args ...any)
	Warn(ctx context.Context, msg string, args ...any)
	Debug(ctx context.Context, msg string, args ...any) Logger
	Fatal(ctx context.Context, msg string, args ...any) Logger
	With(args ...any) Logger // 添加上下文信息
}

// 实现日志接口

type defaultLogger struct {
	// 类似于继承，组合了 slog.Logger 的所有方法
	*slog.Logger
}

// 实现 Info 方法，调用 slog.Logger 的 InfoContext 方法，相当于重写了 slog.Logger 的 Info 方法
func (l *defaultLogger) Info(ctx context.Context, msg string, args ...any) {
	l.Logger.InfoContext(ctx, msg, args...)
}

// 实现 Error 方法，调用 slog.Logger 的 ErrorContext 方法，相当于重写了 slog.Logger 的 Error 方法
func (l *defaultLogger) Error(ctx context.Context, msg string, args ...any) {
	l.Logger.ErrorContext(ctx, msg, args...)
}

// 实现 Warn 方法，调用 slog.Logger 的 WarnContext 方法，相当于重写了 slog.Logger 的 Warn 方法
func (l *defaultLogger) Warn(ctx context.Context, msg string, args ...any) {
	l.Logger.WarnContext(ctx, msg, args...)
}

// 实现 Debug 方法，调用 slog.Logger 的 DebugContext 方法，相当于重写了 slog.Logger 的 Debug 方法
func (l *defaultLogger) Debug(ctx context.Context, msg string, args ...any) Logger {
	l.Logger.DebugContext(ctx, msg, args...)
	return l
}

// 实现 Fatal 方法，记录严重错误级别日志，并退出程序
func (l *defaultLogger) Fatal(ctx context.Context, msg string, args ...any) Logger {
	l.Logger.ErrorContext(ctx, "[FATAL] "+msg, args...)
	os.Exit(1)
	return l
}

// 实现 With 方法，添加上下文信息
func (l *defaultLogger) With(args ...any) Logger {
	// 返回log指针
	var log = l.Logger.With(args...)
	// 放入defaultLogger结构体中，指针依旧在，但是结构体是值
	// &取defaultLogger地址
	return &defaultLogger{log}
	// return &defaultLogger{l.Logger.With(args...)}
}

// 实现 NewLogger 函数，创建一个默认的 Logger 实例
func NewLogger(env Env) Logger {
	// handlerOptions是slog.Handler的配置选项，核心用途：指定日志级别和属性替换函数
	var opts *slog.HandlerOptions
	if env == EnvProd {
		opts = &slog.HandlerOptions{
			Level: slog.LevelInfo, // 生产环境日志级别为 Info，只记录Info级别及以上的日志
			ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
				// 如果属性键为 slog.TimeKey，则将属性值设置为当前时间
				if a.Key == slog.TimeKey {
					a.Value = slog.TimeValue(time.Now().UTC())
				}
				// 返回属性
				return a
			},
		}
	} else {
		opts = &slog.HandlerOptions{
			Level: slog.LevelDebug, // 开发环境日志级别为 Debug
			ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
				// 如果属性键为 slog.TimeKey，则将属性值设置为当前时间
				if a.Key == slog.TimeKey {
					a.Value = slog.TimeValue(time.Now().UTC())
				}
				// 返回属性
				return a
			},
		}
	}
	// 创建slog.Handler，用于格式化日志
	var h slog.Handler
	if env == EnvProd {
		// 生产环境使用 JSON 格式化日志
		h = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		// 开发环境使用文本格式化日志
		h = slog.NewTextHandler(os.Stdout, opts)
	}
	
	// 核心步骤：用自定义的 contextHandler 包装原生的 Handler，实现上下文自动提取
	h = contextHandler{h}

	return &defaultLogger{slog.New(h)}
}

// contextHandler 包装 slog.Handler，用于自动从 context 提取字段注入到每一条日志中
type contextHandler struct {
	slog.Handler // 匿名嵌入，相当于继承
}

// Handle 拦截每条日志输出，在写入前塞入 Context 中的变量
func (h contextHandler) Handle(ctx context.Context, r slog.Record) error {
	// 自动提取 RequestID
	if reqID := GetRequestID(ctx); reqID != "" {
		r.AddAttrs(slog.String("request_id", reqID))
	}
	// 后续如果加上了 User 解析，在这里也能自动提取
	if userID, ok := ctx.Value("current_user_id").(int64); ok {
		r.AddAttrs(slog.Int64("user_id", userID))
	}
	
	return h.Handler.Handle(ctx, r)
}

// context标准库的作用：传递上下文信息
// 比如：一个请求进来，我需要记录这个请求的ID，我就可以使用context来传递这个请求ID
// 这样我就可以在日志中记录这个请求ID
func WithRequestID(ctx context.Context, requestID string) context.Context {
	// context.WithValue返回一个新context，新context中包含旧context和新的键值对
	return context.WithValue(ctx, RequestIDKey, requestID)
}

// 获取请求ID
func GetRequestID(ctx context.Context) string {
	// ctx.Value返回context中存储的值，类型为any
	// RequestIDKey是context.Context中的键，类型为ctxKey
	// (string)是类型断言，将any类型转换为string类型
	// ok是类型断言的结果，如果类型断言成功，则ok为true，否则为false
	// 如果类型断言成功，则返回v的值
	// 如果类型断言失败，则返回空字符串
	if v, ok := ctx.Value(RequestIDKey).(string); ok {
		return v
	}
	return ""
}
