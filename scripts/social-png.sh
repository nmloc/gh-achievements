#!/bin/sh
# Render the social preview PNG (GitHub repository card and og:image) from the
# generated SVG. Needs rsvg-convert (brew install librsvg / apt install librsvg2-bin).
set -eu
cd "$(dirname "$0")/.."
rsvg-convert -w 1280 -h 640 docs/social.svg -o docs/social.png
cp docs/social.png site/social.png
echo "wrote docs/social.png and site/social.png"
