#!/usr/bin/env sh
set -e

: "${GUACAMOLE_VERSION:?GUACAMOLE_VERSION must be set}"

OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m | sed 's/x86_64/amd64/;s/aarch64/arm64/')
VERSION="${GUACAMOLE_VERSION#v}"

curl -fsSL \
  "https://github.com/padok-team/guacamole/releases/download/${GUACAMOLE_VERSION}/guacamole_${VERSION}_${OS}_${ARCH}.tar.gz" \
  -o /tmp/guacamole.tar.gz

tar -xzf /tmp/guacamole.tar.gz -C /usr/local/bin guacamole
chmod +x /usr/local/bin/guacamole
guacamole --help >/dev/null
