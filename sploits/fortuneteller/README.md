#Уязвимости

## Path traversal
Можно получить доступ к ключам книги. 

Фикс: Использовать безопасный метод `join`

https://koumudi-garikipati.medium.com/go-lang-directory-traversal-567b586e5d0b

## AES ECB

## RSA 

https://crypto.stackexchange.com/questions/6713/low-public-exponent-attack-for-rsa

## Insertion of Sensitive Information into Log File

1) В логи попадает конфиденциальная информация (пароли, секреты/флаги и тп)
2) Логи by-design доступны внешнему наблюдателю через систему мониторинга

Фикс: маскирование значений секретов в логах / исправление самих сообщений

https://cwe.mitre.org/data/definitions/532.html
