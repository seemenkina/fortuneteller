# ZPD-training-service
Services for A/D CTF training.

## fortuneteller

Service for fortune-telling by books.

### Tags

- golang
- postgres
- crypto
- web

### Vulnerabilities

- Path traversal on golang by using unsafe method `filepath.Join()`. [Sploit.](./sploits/fortuneteller/path_traversal_exploit.py)
- AES in insecure ECB mode.  [Sploit.](./sploits/fortuneteller/aes_exploit.py)
- Usage low public exponent in RSA. [Sploit.](./sploits/fortuneteller/rsa_exploit.py)
- Insertion of Sensitive Information into Log File.

More details [here](./sploits/fortuneteller/README.md)
## Deploy

### Service

```bash
cd ./services/fortuneteller
docker-compose up -d
```

### Checker

The checker interface matches the description for ructf: `https://github.com/HackerDom/ructf-2017/wiki/Интерфейс-«проверяющая-система-чекеры»`

```bash
cd ./checkers/fortuneteller
python3 checker.py 
```

## Contributors

@seemenkina

