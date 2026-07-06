---
title: "Конфигурация с Viper"
sidebar_position: 11
---

import Quiz from '@site/src/components/Quiz';

# Конфигурация с Viper

Пароли и URL'ы в коде — плохо. Переменные окружения — лучше. Viper — библиотека для управления конфигурацией: `.env`, YAML, JSON, флаги — всё в одном интерфейсе. 12-factor app, третье правило: «конфигурация в окружении».

## Установка

```bash
go get github.com/spf13/viper
```

## Базовая настройка

```go
import "github.com/spf13/viper"

func initConfig() {
    viper.SetConfigName("config")   // имя файла без расширения
    viper.SetConfigType("yaml")      // или json, toml
    viper.AddConfigPath(".")         // где искать
    viper.AddConfigPath("$HOME/.blog")

    // Переменные окружения имеют приоритет
    viper.AutomaticEnv()
    viper.SetEnvPrefix("BLOG")

    if err := viper.ReadInConfig(); err != nil {
        slog.Warn("конфиг не найден, использую env", "err", err)
    }
}
```

## config.yaml

```yaml
server:
  port: 8080
  read_timeout: 5s
  write_timeout: 10s

database:
  url: postgres://localhost:5432/blog
  max_connections: 20

auth:
  jwt_secret: ""  # переопределяется через BLOG_AUTH_JWT_SECRET
```

## Использование

```go
port := viper.GetInt("server.port")
dbURL := viper.GetString("database.url")
secret := viper.GetString("auth.jwt_secret")
timeout := viper.GetDuration("server.read_timeout")
```

## Приоритет: env > файл > default

```go
viper.SetDefault("server.port", 8080) // 3. Значение по умолчанию
// config.yaml: port: 9090             // 2. Файл конфигурации
// BLOG_SERVER_PORT=7070               // 1. Переменная окружения — побеждает
```

Viper автоматически мапит `BLOG_SERVER_PORT` → `server.port` через `SetEnvPrefix("BLOG")` + `AutomaticEnv()`.

## Структурирование конфигурации

```go
type Config struct {
    Server   ServerConfig
    Database DatabaseConfig
    Auth     AuthConfig
}

type ServerConfig struct {
    Port         int           `mapstructure:"port"`
    ReadTimeout  time.Duration `mapstructure:"read_timeout"`
    WriteTimeout time.Duration `mapstructure:"write_timeout"`
}

func LoadConfig() (*Config, error) {
    var cfg Config
    if err := viper.Unmarshal(&cfg); err != nil {
        return nil, err
    }
    return &cfg, nil
}
```

<Quiz quizId="p1-11-viper" questions={[
  {id:"q1",question:"Почему Viper лучше чем os.Getenv вручную?",options:["Он быстрее","Один интерфейс для env, YAML, JSON, флагов + автоматический маппинг в структуру + значения по умолчанию","Getenv не работает в Docker","Viper обязателен для 12-factor apps"],correctIndex:1,explanation:"Viper унифицирует все источники конфигурации. Переменная окружения переопределяет значение из YAML, которое переопределяет default. Плюс Unmarshal в структуру — не нужно парсить каждое поле отдельно."}
]} />
