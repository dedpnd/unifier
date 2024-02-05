# Как запустить ?
![Coverage](https://img.shields.io/badge/Coverage-70.3%25-brightgreen)

Все необходимые службы для работы находятся в docker-compose

```
docker-compose up -d
go run ./cmd/*
```

По завершение работы желательно отключить тестовую среду 
```
docker-compose down
```

# Описание среды выполнения

zookeeper - служба координации данных  
kafka - брокер сообщений  
kafka-workload - иммитатор нагрузки для kafka  
redpanda - интерфейс управления для система потоковых данных  
postgres - СУБД  

# Начало работы

Пользователь postgres - user | password  
Требуеться зарегистрировать пользователя 
Так же добавленно 1 правило унификации оно сразу настроенно на работы с тестовыми событиями

```
{
    "topicFrom": "events",
    "filter": {
        "regexp": "\"dstHost.ip\": \"10.10.10.10\""
    },
    "entityHash": [
        "srcHost.ip",
        "dstHost.port"
    ],
    "unifier": [
        {
            "name": "id",
            "type": "string",
            "expression": "auditEventLog"
        },
        {
            "name": "date",
            "type": "timestamp",
            "expression": "datetime"
        },
        {
            "name": "ipaddr",
            "type": "string",
            "expression": "srcHost.ip"
        },
        {
            "name": "category",
            "type": "string",
            "expression": "cat"
        }
    ],
    "extraProcess": [
        {
            "func": "__if",
            "args": "category, /Host/Connect/Host/Accept, high",
            "to": "category"
        },
        {
            "func": "__stringConstant",
            "args": "test",
            "to": "customString1"
        }
    ],
    "topicTo": "test"
}
```

Для просмотра топиков можно использовать консоль redpand - http://localhost:8081/topics  
Набор тестовых данных для топика входящих событий - `./kafka-perf-test/scripts/example.json`  
