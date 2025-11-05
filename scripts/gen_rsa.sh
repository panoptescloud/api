#!/bin/sh

if [ -f "private.key" ] || [ -f "public.key" ]; then
    echo "Keys may already exist, check source directory mounted at $(pwd) and remove them to regenrate!"
    exit 1
fi

# Private key
openssl genrsa -out private.key 2048

# Public key
openssl rsa -in private.key -pubout -out public.key