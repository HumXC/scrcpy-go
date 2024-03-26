#!/usr/bin/env sh

VERSION="v2.4"

URL="https://github.com/Genymobile/scrcpy/releases/download/$VERSION/scrcpy-server-$VERSION"

DIR="$(dirname "$(readlink -f "$0")")"

curl -L "$URL" -o "$DIR/scrcpy-server-$VERSION"

echo "Download of version $VERSION complete!"
