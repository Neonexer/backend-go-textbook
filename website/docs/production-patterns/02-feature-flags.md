---
title: "Feature Flags"
sidebar_position: 2
---

import Quiz from '@site/src/components/Quiz';

# Feature Flags

Новая фича готова, но включать сразу на всех пользователей страшно. Feature flags позволяют включать функциональность постепенно: 1% → 10% → 50% → 100%.

## Простая реализация

```go
type FeatureFlags struct {
    NewSearchEnabled  bool
    DarkModeEnabled   bool
    BetaCheckoutPct   int // 0-100
}

func (f *FeatureFlags) IsEnabled(flag string, userID int) bool {
    switch flag {
    case "beta_checkout":
        // Детерминированно для пользователя
        return userID%100 < f.BetaCheckoutPct
    }
    return false
}
```

## Источники флагов

- **Конфиг-файл** — простые статические флаги
- **БД / Redis** — динамическое переключение без передеплоя
- **LaunchDarkly / Flagsmith** — UI, аудит, A/B testing

## Процентный rollout

```go
func isInExperiment(userID string, pct int) bool {
    h := fnv.New32a()
    h.Write([]byte(userID))
    return int(h.Sum32()%100) < pct
}
```

Хеш гарантирует что пользователь всегда попадает в одну группу.

## Ключевые выводы

- Feature flags отделяют деплой от релиза
- Процентный rollout снижает риск
- Хеш от userID для консистентных групп

<Quiz quizId="pp-02-feature-flags" questions={[
  {id:"q1",question:"Как feature flag помогает при баге в новой фиче?",options:["Никак","Фича отключается мгновенно через конфиг — не нужно откатывать деплой","Флаг автоматически чинит баги","Флаг замедляет пользователей"],correctIndex:1,explanation:"Без флагов багнутую фичу нужно откатывать (деплой 10-15 минут). С флагом — выключил в конфиге, фича исчезла мгновенно."}
]} />
