---
title: "GORM"
sidebar_position: 2
---

import Quiz from '@site/src/components/Quiz';

# GORM — ORM для Go

В прошлой главе мы писали SQL вручную через pgx. Теперь посмотрим на GORM — популярный ORM, который генерирует SQL за тебя. В этой главе разберём: когда ORM помогает, а когда мешает, и как писать на GORM так, чтобы не было больно.

## Что такое GORM и зачем он нужен

GORM — это Object-Relational Mapper. Вместо `SELECT * FROM products WHERE id = ?` ты пишешь `db.First(&product, id)`. GORM генерирует SQL, мапит строки в структуры и обратно.

```go
import (
    "gorm.io/gorm"
    "gorm.io/driver/postgres"
)

db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
```

## Подходы: pgx vs GORM

| | pgx (нативный SQL) | GORM (ORM) |
|---|---|---|
| **Контроль** | Полный контроль над SQL | SQL генерируется автоматически |
| **Производительность** | Максимальная | Небольшой оверхед |
| **Порог входа** | Нужно знать SQL | Можно начать без глубокого SQL |
| **Сложные запросы** | Легко (пишешь SQL) | Трудно (борешься с ORM) |
| **Миграции** | Внешний инструмент | Встроенный AutoMigrate |
| **N+1 проблема** | Ручной контроль | Может возникнуть незаметно |

:::tip Когда что выбирать
**pgx** — если у тебя сложные запросы, высокая нагрузка, или ты хочешь полный контроль. **GORM** — для типовых CRUD-операций, когда 90% запросов это `SELECT/INSERT/UPDATE/DELETE` по первичному ключу.
:::

## Определение моделей

GORM использует структурные теги для маппинга:

```go
type Product struct {
    ID          uint           `gorm:"primaryKey"`
    Title       string         `gorm:"size:200;not null"`
    Description string         `gorm:"type:text"`
    Price       int            `gorm:"not null;check:price >= 0"`
    SellerID    uint           `gorm:"index;not null"`
    Seller      User           `gorm:"foreignKey:SellerID"`
    CreatedAt   time.Time
    UpdatedAt   time.Time
}
```

Теги GORM поверх обычных Go-структур — без магии кодогенерации (в отличие от Prisma, например).

## AutoMigrate — миграции из кода

GORM умеет создавать таблицы по структурам:

```go
db.AutoMigrate(&Product{}, &User{})
```

Это создаст таблицы `products` и `users` с колонками, соответствующими полям структур. Для dev-окружения — ок, для продакшена лучше использовать файлы миграций (глава 3).

## CRUD на GORM

### Create

```go
product := Product{Title: "Кроссовки", Price: 500000, SellerID: 1}
result := db.Create(&product)
// product.ID теперь содержит сгенерированный ID
```

### Read

```go
// По первичному ключу
var product Product
db.First(&product, 1)        // SELECT * FROM products WHERE id = 1
db.First(&product, "title = ?", "Кроссовки") // с условием

// Много записей
var products []Product
db.Where("price < ?", 100000).Find(&products)
db.Where("seller_id = ?", sellerID).Order("created_at desc").Find(&products)
```

### Update

```go
db.Model(&product).Updates(Product{Title: "Новое название", Price: 600000})
// UPDATE products SET title='...', price=600000 WHERE id = product.ID

// Или одно поле
db.Model(&product).Update("price", 450000)
```

### Delete

```go
db.Delete(&product, 1) // мягкое удаление если есть DeletedAt
// Или жёсткое:
db.Unscoped().Delete(&product, 1)
```

:::warning Мягкое удаление по умолчанию
Если в структуре есть поле `gorm.DeletedAt`, `Delete()` не удаляет строку, а проставляет дату удаления. Все последующие `Find` автоматически фильтруют «удалённые» записи. Удобно, но может удивить.
:::

## Связи и N+1 проблема

GORM делает связи удобными:

```go
type Product struct {
    // ...
    SellerID uint
    Seller   User `gorm:"foreignKey:SellerID"`
}

// Жадная загрузка
var products []Product
db.Preload("Seller").Find(&products)
```

Но без `Preload` при обращении к `product.Seller` GORM сделает **отдельный запрос** для каждого товара. Это N+1 проблема — 1 запрос на товары + N запросов на продавцов.

**Всегда явно указывай Preload для связей, которые понадобятся.**

## Когда GORM начинает мешать

Чем сложнее запрос, тем труднее выразить его через методы GORM:

```go
// GORM — читаемо
db.Where("price < ?", maxPrice).
   Where("status = ?", "active").
   Order("created_at desc").
   Limit(20).
   Find(&products)

// Но группировка с JOIN уже не так хороша
db.Table("products").
   Select("seller_id, AVG(price) as avg_price").
   Joins("JOIN users ON users.id = products.seller_id").
   Where("users.role = ?", "seller").
   Group("seller_id").
   Having("AVG(price) > ?", 100000).
   Find(&results)
```

Когда запрос перестаёт быть цепочкой методов и превращается в кусок SQL внутри `Raw()` — проще написать на pgx.

## Гибридный подход

Лучшее из двух миров: GORM для простых операций, pgx для сложных:

```go
type ProductRepo struct {
    db   *gorm.DB      // для простых CRUD
    pool *pgxpool.Pool  // для сложных запросов и транзакций
}

func (r *ProductRepo) FindByID(id int) (*Product, error) {
    var p Product
    return &p, r.db.First(&p, id).Error
}

func (r *ProductRepo) TopSellers(ctx context.Context) ([]SellerStat, error) {
    // Сложная аналитика — на чистом SQL
    rows, _ := r.pool.Query(ctx, `SELECT seller_id, COUNT(*), AVG(price) ...`)
    // ...
}
```

Не обязан выбирать что-то одно. GORM для 80% запросов, pgx для оставшихся 20%.

## Ключевые выводы

1. GORM хорош для типовых CRUD — меньше бойлерплейта
2. `AutoMigrate` для dev, файлы миграций для production
3. `Preload` обязателен для связей — иначе N+1 проблема
4. Сложные запросы — на чистом SQL через `db.Raw()` или pgx
5. Гибридный подход: GORM + pgx в одном проекте

В следующей главе настроим миграции через `golang-migrate` — правильный способ управлять схемой БД.

---

## Проверь себя

<Quiz
  quizId="p2-02-gorm"
  questions={[
    {
      id: "q1",
      question: "Что такое N+1 проблема в GORM и как её избежать?",
      options: [
        "Это когда делается N+1 запросов вместо одного с JOIN — решается через db.Preload()",
        "Это баг в GORM, который исправят в следующей версии",
        "Это ограничение PostgreSQL, не связанное с GORM",
        "Это когда N товаров имеют N+1 продавцов"
      ],
      correctIndex: 0,
      explanation: "Без Preload GORM делает 1 запрос на товары и N запросов на их продавцов (по одному на каждый товар). Preload('Seller') делает 2 запроса вместо N+1."
    },
    {
      id: "q2",
      question: "Когда GORM стоит заменить на чистый SQL?",
      options: [
        "Всегда — чистый SQL всегда лучше",
        "Когда запросы становятся сложнее цепочки Where/Order/Limit: JOIN'ы, GROUP BY, подзапросы",
        "Только если DBA требует",
        "GORM нельзя использовать с чистым SQL в одном проекте"
      ],
      correctIndex: 1,
      explanation: "Цепочки методов GORM хороши для простых CRUD. Но для GROUP BY с HAVING, сложных JOIN'ов или оконных функций проще написать SQL — и через db.Raw(), и через отдельный pgx-пул."
    },
    {
      id: "q3",
      question: "Почему AutoMigrate не стоит использовать в продакшене?",
      options: [
        "Он слишком медленный",
        "Он не умеет создавать индексы",
        "Он не даёт контроля над миграциями: нельзя откатить, нет версионирования, нет аудита изменений схемы",
        "Он работает только с SQLite"
      ],
      correctIndex: 2,
      explanation: "AutoMigrate подходит для локальной разработки. В продакшене нужны контролируемые, версионированные, откатываемые миграции с понятным апрувом."
    }
  ]}
/>
