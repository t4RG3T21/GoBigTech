package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger обертка над zap.Logger для удобного использования
type Logger struct {
	*zap.Logger
}

// New создает новый логгер с настройками в зависимости от уровня логирования
// serviceName - имя сервиса (будет добавлено в каждое сообщение)
// level - уровень логирования: debug, info, warn, error
func New(serviceName string, level string) (*Logger, error) {
	// Парсим уровень логирования
	zapLevel, err := parseLevel(level)
	if err != nil {
		return nil, err
	}

	// Определяем режим: development (debug) или production (остальные)
	isDevelopment := level == "debug"

	var config zap.Config
	if isDevelopment {
		// Development режим: консольный вывод с цветами
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		config.EncoderConfig.EncodeCaller = zapcore.ShortCallerEncoder
	} else {
		// Production режим: JSON формат
		config = zap.NewProductionConfig()
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	}

	// Устанавливаем уровень логирования
	config.Level = zap.NewAtomicLevelAt(zapLevel)

	// Добавляем имя сервиса в конфигурацию
	config.InitialFields = map[string]interface{}{
		"service": serviceName,
	}

	// Строим логгер
	logger, err := config.Build()
	if err != nil {
		return nil, err
	}

	return &Logger{Logger: logger}, nil
}

// parseLevel преобразует строковый уровень в zapcore.Level
// Всегда возвращает валидный уровень (по умолчанию InfoLevel)
func parseLevel(level string) (zapcore.Level, error) {
	switch level {
	case "debug":
		return zapcore.DebugLevel, nil
	case "info":
		return zapcore.InfoLevel, nil
	case "warn":
		return zapcore.WarnLevel, nil
	case "error":
		return zapcore.ErrorLevel, nil
	default:
		return zapcore.InfoLevel, nil // По умолчанию info
	}
}

// WithFields добавляет поля к логгеру
func (l *Logger) WithFields(fields ...zap.Field) *Logger {
	return &Logger{Logger: l.Logger.With(fields...)}
}

// WithError добавляет ошибку к логгеру
func (l *Logger) WithError(err error) *Logger {
	return &Logger{Logger: l.Logger.With(zap.Error(err))}
}
