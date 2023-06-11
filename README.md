# Masked Email Bot

Currently, works only with Fastmail.com.

To build static binary you can use `podman` image (or `docker`):
```bash
podman build -t masked-email-bot .
$id = podman create masked-email-bot
podman cp $id:/usr/local/bin/masked-email-bot .
podman rm -v $id
```