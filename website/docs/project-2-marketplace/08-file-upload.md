---
title: "Загрузка файлов"
sidebar_position: 8
---

import Quiz from '@site/src/components/Quiz';

# Загрузка файлов

Фото товаров, аватарки, документы — в этой главе настроим приём файлов через multipart/form-data, валидацию и хранение.

## Приём файла

Go парсит multipart автоматически:

```go
func (h *ProductHandler) UploadImage(w http.ResponseWriter, r *http.Request) {
    // 10 MB максимум в памяти, остальное во временный файл
    r.ParseMultipartForm(10 << 20)

    file, header, err := r.FormFile("image")
    if err != nil {
        writeError(w, http.StatusBadRequest, "file required")
        return
    }
    defer file.Close()

    // Валидация
    if header.Size > 5<<20 { // 5 MB
        writeError(w, http.StatusBadRequest, "file too large (max 5MB)")
        return
    }

    contentType := header.Header.Get("Content-Type")
    if contentType != "image/jpeg" && contentType != "image/png" {
        writeError(w, http.StatusBadRequest, "only JPEG and PNG allowed")
        return
    }

    // Читаем и сохраняем
    data, err := io.ReadAll(file)
    path, err := h.storage.Save(r.Context(), header.Filename, data)
    // ...
}
```

## Где хранить

| Вариант | Плюсы | Минусы |
|---------|-------|--------|
| Локальный диск | Просто, бесплатно | Не масштабируется на несколько инстансов |
| S3 / Yandex Object Storage | Масштабируется, CDN | Платно |
| В БД (bytea) | Транзакционность | БД не для файлов, медленно |

Для старта — локальный диск с абстракцией:

```go
type FileStorage interface {
    Save(ctx context.Context, filename string, data []byte) (string, error)
    Delete(ctx context.Context, path string) error
}

type LocalStorage struct {
    basePath string
}

func (s *LocalStorage) Save(ctx context.Context, filename string, data []byte) (string, error) {
    // Генерируем уникальное имя
    ext := filepath.Ext(filename)
    name := uuid.New().String() + ext
    path := filepath.Join(s.basePath, name)
    return "/uploads/" + name, os.WriteFile(path, data, 0644)
}
```

## Отдача файлов

```go
r.Get("/uploads/{filename}", func(w http.ResponseWriter, r *http.Request) {
    filename := chi.URLParam(r, "filename")
    http.ServeFile(w, r, filepath.Join("./uploads", filename))
})
```

Для продакшена — nginx перед приложением для отдачи статики, или CDN.

## Ключевые выводы

- `r.ParseMultipartForm(maxMemory)` перед `r.FormFile`
- Валидируй размер и тип файла на сервере
- Генерируй уникальные имена (uuid) чтобы избежать коллизий
- Абстракция `FileStorage` для лёгкой замены локального диска на S3

<Quiz quizId="p2-08-files" questions={[
  {id:"q1",question:"Почему нельзя полагаться на проверку типа файла только по расширению?",options:["Расширение легко подделать — проверяй Content-Type и сигнатуру файла","Можно, расширения достаточно","Go не поддерживает проверку расширений","Расширения не уникальны"],correctIndex:0,explanation:"Content-Type из заголовка тоже можно подделать. Для критичных случаев проверяй магические байты (первые байты файла): JPEG начинается с FF D8 FF, PNG с 89 50 4E 47."}
]} />
