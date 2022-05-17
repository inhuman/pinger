# pinger

## Принцип работы

1.  Считывает переменные окружения которые имеют префикс `HOST_`, `LATENCY_`,
    `PERIOD_`. Из `HOST_*` берёт доменные имена, котоорые нужно проверять; из
    `LATENCY_*` длительность тайм-аута на проверку доступности; из `PERIOD_*` --
    время, через которое запускается проверка.
2.  Резолвит доменное имя.
3.  Пингует серверы, на которые указывет доменное имя, по ICMP.
4.  Посылает HTTP GET запрос.
5.  Проверяет срок действия сертификата.

## Метрики

Метрики имеют название *var_name*\_availability. Где _var_name_ — это часть переменной окружения без префикса. Например: если задана переменная `HOST_EXAMPLE=https//...`, то название метрики будет `example_availability`.

### Возможные значения

| Значение              | Описание                           |
| --------------------- | ---------------------------------- |
| `host_not_resolved`   | Ошибка при резолве доменного имени |
| `host_not_accessible` | Ошибка при пинге серверов          |
| `http_request_error`  | Ошибка при HTTP GET запросе        |
| `certificate_invalid` | Ошибка при проверке сертификата    |
| `timeout_exceed`      | Тайм-аут                           |
| `success`             | Ошибок при проверке не произошло   |
