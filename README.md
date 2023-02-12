# tg-service

## Описание

```
Микро сервис(тьфу тьфу тьфу) предназначен для проксирования событий из telegram на api url которые указаны в конфиге, 
с функцией обратного ответа, добавлена поддержка файлов
```

## Функции

- Поддержка нескольких(до 5) каналов слушателей
- Управление кнопками в ботах

## Build

GOOS=linux GOARCH=amd64 go build -o ./tg-service -a

## Запросы

POST http://localhost:5243/channel/test_bot где test_bot это name в конфге

```
{
    "text":"Привет я просто скучный текст",
    "chat_id":19223218410
}

{
    "file_url":"https://i.imgur.com/unQLJIb.jpg",
    "text":"Привет я картинка с котиком по урлу",
    "chat_id":19223218410
}

{
    "file_path":"/Volumes/work/test/Controller.file",
    "text":"Привет я файл",
    "chat_id":19223218410
}

{
    "file_path":"/Users/alex/Downloads/faf0f1as-960.jpg",
    "text":"Привет я картинка с котиком файлом",
    "chat_id":19223218410
}
```

