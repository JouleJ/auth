package log

import (
    "auth/pkg/config"
)

import (
    "github.com/TheZeroSlave/zapsentry"
    "go.uber.org/zap"
    "go.uber.org/zap/zapcore"
)

func New(cfg *config.Config) (*zap.Logger, error) {
    logger, err := zap.NewProduction()
    if err != nil {
        return nil, err
    }

    zapsectry_cfg := zapsentry.Configuration {
        Level: zapcore.Level(cfg.LoggerLevel),
    }

    core, err := zapsentry.NewCore(zapsectry_cfg, zapsentry.NewSentryClientFromDSN(cfg.SentryDsn))
    if err != nil {
        return nil, err
    }

    return zapsentry.AttachCoreToLogger(core, logger), nil
}
