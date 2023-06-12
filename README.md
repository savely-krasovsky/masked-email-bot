# Masked Email Bot

Currently, works only with Fastmail.com.

To build static binary you can use this script:

```bash
#!/bin/sh
docker build -t masked-email-bot .
id=$(docker create masked-email-bot)
docker cp $id:/usr/local/bin/masked-email-bot .
docker rm -v $id
```

It requires `podman` or `docker`.