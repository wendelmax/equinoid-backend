package logging

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// Logger representa o logger estruturado
type Logger struct {
	*logrus.Logger
}

// Fields representa campos adicionais para logs estruturados
type Fields map[string]interface{}

// NewLogger cria um novo logger estruturado
func NewLogger(level string) *Logger {
	logger := logrus.New()

	// Configurar nível de log
	logLevel, err := logrus.ParseLevel(level)
	if err != nil {
		logLevel = logrus.InfoLevel
	}
	logger.SetLevel(logLevel)

	// Configurar formatação JSON para produção
	logger.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat:   time.RFC3339,
		DisableHTMLEscape: true,
		FieldMap: logrus.FieldMap{
			logrus.FieldKeyTime:  "timestamp",
			logrus.FieldKeyLevel: "level",
			logrus.FieldKeyMsg:   "message",
			logrus.FieldKeyFunc:  "function",
			logrus.FieldKeyFile:  "file",
		},
	})

	// Configurar saída
	logger.SetOutput(os.Stdout)

	// Habilitar informações de função e arquivo em modo debug
	if logLevel <= logrus.DebugLevel {
		logger.SetReportCaller(true)
	}

	// Adicionar hooks customizados
	logger.AddHook(&ContextHook{})
	logger.AddHook(&CallerHook{})

	return &Logger{logger}
}

// WithFields adiciona campos estruturados ao log
func (l *Logger) WithFields(fields Fields) *logrus.Entry {
	return l.Logger.WithFields(logrus.Fields(fields))
}

// WithContext adiciona contexto ao log
func (l *Logger) WithContext(ctx context.Context) *logrus.Entry {
	return l.Logger.WithContext(ctx)
}

// WithError adiciona erro ao log
func (l *Logger) WithError(err error) *logrus.Entry {
	return l.Logger.WithError(err)
}

// WithField adiciona um campo ao log
func (l *Logger) WithField(key string, value interface{}) *logrus.Entry {
	return l.Logger.WithField(key, value)
}

// WithComponent adiciona o componente ao log
func (l *Logger) WithComponent(component string) *logrus.Entry {
	return l.WithField("component", component)
}

// WithUserID adiciona o ID do usuário ao log
func (l *Logger) WithUserID(userID uint) *logrus.Entry {
	return l.WithField("user_id", userID)
}

// WithEquinoID adiciona o ID do equino ao log
func (l *Logger) WithEquinoID(equinoID string) *logrus.Entry {
	return l.WithField("equino_id", equinoID)
}

// WithRequestID adiciona o ID da requisição ao log
func (l *Logger) WithRequestID(requestID string) *logrus.Entry {
	return l.WithField("request_id", requestID)
}

// WithHTTPRequest adiciona informações da requisição HTTP ao log
func (l *Logger) WithHTTPRequest(method, path, userAgent, ip string, statusCode int, duration time.Duration) *logrus.Entry {
	return l.WithFields(Fields{
		"http_method":      method,
		"http_path":        path,
		"http_user_agent":  userAgent,
		"http_ip":          ip,
		"http_status":      statusCode,
		"http_duration_ms": duration.Milliseconds(),
	})
}

// LogHTTPRequest registra uma requisição HTTP
func (l *Logger) LogHTTPRequest(method, path, userAgent, ip string, statusCode int, duration time.Duration) {
	entry := l.WithHTTPRequest(method, path, userAgent, ip, statusCode, duration)

	if statusCode >= 500 {
		entry.Error("HTTP request failed")
	} else if statusCode >= 400 {
		entry.Warn("HTTP request error")
	} else {
		entry.Info("HTTP request")
	}
}

// LogDatabaseQuery registra uma consulta ao banco de dados
func (l *Logger) LogDatabaseQuery(query string, duration time.Duration, rows int64) {
	l.WithFields(Fields{
		"db_query":       query,
		"db_duration_ms": duration.Milliseconds(),
		"db_rows":        rows,
	}).Debug("Database query")
}

// LogCacheOperation registra uma operação de cache
func (l *Logger) LogCacheOperation(operation, key string, hit bool, duration time.Duration) {
	l.WithFields(Fields{
		"cache_operation":   operation,
		"cache_key":         key,
		"cache_hit":         hit,
		"cache_duration_ms": duration.Milliseconds(),
	}).Debug("Cache operation")
}

// LogAuthentication registra uma tentativa de autenticação
func (l *Logger) LogAuthentication(userID uint, email string, success bool, reason string) {
	entry := l.WithFields(Fields{
		"auth_user_id": userID,
		"auth_email":   email,
		"auth_success": success,
		"auth_reason":  reason,
	})

	if success {
		entry.Info("Authentication successful")
	} else {
		entry.Warn("Authentication failed")
	}
}

// LogSecurityEvent registra um evento de segurança
func (l *Logger) LogSecurityEvent(eventType, description string, userID uint, ip string) {
	l.WithFields(Fields{
		"security_event": eventType,
		"security_desc":  description,
		"security_user":  userID,
		"security_ip":    ip,
	}).Warn("Security event")
}

// LogBusinessEvent registra um evento de negócio
func (l *Logger) LogBusinessEvent(eventType, description string, userID uint, equinoID string, metadata Fields) {
	fields := Fields{
		"business_event":  eventType,
		"business_desc":   description,
		"business_user":   userID,
		"business_equino": equinoID,
	}

	// Adicionar metadados customizados
	for k, v := range metadata {
		fields["business_"+k] = v
	}

	l.WithFields(fields).Info("Business event")
}

// LogError registra um erro com contexto
func (l *Logger) LogError(err error, component string, metadata Fields) {
	entry := l.WithError(err).WithField("component", component)

	if metadata != nil {
		entry = entry.WithFields(logrus.Fields(metadata))
	}

	entry.Error("Application error")
}

// LogPanic registra um panic
func (l *Logger) LogPanic(panicValue interface{}, stack string) {
	l.WithFields(Fields{
		"panic_value": panicValue,
		"panic_stack": stack,
	}).Fatal("Application panic")
}

// ContextHook adiciona informações do contexto aos logs
type ContextHook struct{}

// Levels define os níveis em que o hook será executado
func (hook *ContextHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

// Fire é executado quando um log é registrado
func (hook *ContextHook) Fire(entry *logrus.Entry) error {
	if entry.Context != nil {
		// Extrair informações do contexto
		if requestID := entry.Context.Value("request_id"); requestID != nil {
			entry.Data["request_id"] = requestID
		}

		if userID := entry.Context.Value("user_id"); userID != nil {
			entry.Data["user_id"] = userID
		}

		if traceID := entry.Context.Value("trace_id"); traceID != nil {
			entry.Data["trace_id"] = traceID
		}
	}

	return nil
}

// CallerHook adiciona informações do caller aos logs
type CallerHook struct{}

// Levels define os níveis em que o hook será executado
func (hook *CallerHook) Levels() []logrus.Level {
	return []logrus.Level{
		logrus.PanicLevel,
		logrus.FatalLevel,
		logrus.ErrorLevel,
		logrus.WarnLevel,
	}
}

// Fire é executado quando um log é registrado
func (hook *CallerHook) Fire(entry *logrus.Entry) error {
	if entry.HasCaller() {
		return nil // Já tem informação de caller
	}

	// Obter informações do caller
	pc, file, line, ok := runtime.Caller(8) // Skip logrus + nossa abstração
	if !ok {
		return nil
	}

	function := runtime.FuncForPC(pc)
	if function == nil {
		return nil
	}

	functionName := function.Name()
	fileName := filepath.Base(file)

	// Adicionar informações aos dados do log
	entry.Data["caller"] = fmt.Sprintf("%s:%d", fileName, line)
	entry.Data["function"] = functionName

	return nil
}

// FileLoggerHook salva logs em arquivo
type FileLoggerHook struct {
	logPath string
	levels  []logrus.Level
}

// NewFileLoggerHook cria um novo hook para salvar logs em arquivo
func NewFileLoggerHook(logPath string, levels []logrus.Level) *FileLoggerHook {
	return &FileLoggerHook{
		logPath: logPath,
		levels:  levels,
	}
}

// Levels retorna os níveis que serão salvos em arquivo
func (hook *FileLoggerHook) Levels() []logrus.Level {
	return hook.levels
}

// Fire salva o log em arquivo
func (hook *FileLoggerHook) Fire(entry *logrus.Entry) error {
	// Criar diretório se não existir
	dir := filepath.Dir(hook.logPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// Abrir/criar arquivo de log
	file, err := os.OpenFile(hook.logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	// Formatar e escrever log
	formatter := &logrus.JSONFormatter{
		TimestampFormat: time.RFC3339,
	}

	serialized, err := formatter.Format(entry)
	if err != nil {
		return err
	}

	_, err = file.Write(serialized)
	return err
}

// SetupFileLogging adiciona logging em arquivo
func (l *Logger) SetupFileLogging(logPath string) error {
	// Hook para todos os logs
	allLevelsHook := NewFileLoggerHook(logPath, logrus.AllLevels)
	l.AddHook(allLevelsHook)

	// Hook específico para erros
	errorLogPath := strings.Replace(logPath, ".log", ".error.log", 1)
	errorHook := NewFileLoggerHook(errorLogPath, []logrus.Level{
		logrus.PanicLevel,
		logrus.FatalLevel,
		logrus.ErrorLevel,
	})
	l.AddHook(errorHook)

	return nil
}

// StructuredLogger interface para logging estruturado
type StructuredLogger interface {
	Debug(args ...interface{})
	Info(args ...interface{})
	Warn(args ...interface{})
	Error(args ...interface{})
	Fatal(args ...interface{})
	Panic(args ...interface{})

	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Fatalf(format string, args ...interface{})
	Panicf(format string, args ...interface{})

	WithFields(fields Fields) *logrus.Entry
	WithContext(ctx context.Context) *logrus.Entry
	WithError(err error) *logrus.Entry
	WithField(key string, value interface{}) *logrus.Entry
}

// Ensure Logger implements StructuredLogger
var _ StructuredLogger = (*Logger)(nil)
