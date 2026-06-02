#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
APP_NAME="custos"
ARCH="${ARCH:-amd64}"
VERSION_RAW="${VERSION:-$(git -C "$ROOT_DIR" describe --tags --always --dirty 2>/dev/null || echo 0.0.0)}"
VERSION="${VERSION_RAW#v}"
if [[ ! "$VERSION" =~ ^[0-9] ]]; then
  VERSION="0.0.0+git.${VERSION}"
fi
VERSION="${VERSION//-dirty/+dirty}"
OUT_DIR="${OUT_DIR:-$ROOT_DIR/dist/deb}"
STAGE_DIR="$OUT_DIR/${APP_NAME}_${VERSION}_${ARCH}"
BIN_DIR="$ROOT_DIR/build/bin"

cd "$ROOT_DIR"
rm -rf "$STAGE_DIR"
mkdir -p \
  "$STAGE_DIR/DEBIAN" \
  "$STAGE_DIR/opt/$APP_NAME" \
  "$STAGE_DIR/usr/bin" \
  "$STAGE_DIR/usr/share/applications" \
  "$STAGE_DIR/usr/share/icons/hicolor/256x256/apps" \
  "$STAGE_DIR/usr/share/metainfo"

if [[ -z "${SKIP_BUILD:-}" ]]; then
  wails build -platform "linux/$ARCH"
fi

BIN_PATH="$(find "$BIN_DIR" -maxdepth 1 -type f -perm -111 -name "${APP_NAME}*" | head -n 1)"
if [[ -z "$BIN_PATH" ]]; then
  printf 'built binary not found in %s\n' "$BIN_DIR" >&2
  exit 1
fi

install -m 0755 "$BIN_PATH"           "$STAGE_DIR/opt/$APP_NAME/$APP_NAME"
ln -s "/opt/$APP_NAME/$APP_NAME"      "$STAGE_DIR/usr/bin/$APP_NAME"
install -m 0644 "$ROOT_DIR/build/appicon.png" \
  "$STAGE_DIR/usr/share/icons/hicolor/256x256/apps/$APP_NAME.png"

# ── .desktop file ──────────────────────────────────────────────────────────────
cat > "$STAGE_DIR/usr/share/applications/$APP_NAME.desktop" <<EOF
[Desktop Entry]
Version=1.0
Name=Custos
GenericName=Battery Monitor
Comment=System battery status and notifications
Keywords=Custos;Battery;Monitor;Power;
Exec=/opt/custos/custos
TryExec=/opt/custos/custos
Terminal=false
Type=Application
Categories=Utility;HardwareSettings;System;
Icon=custos
StartupNotify=false
EOF

# ── AppStream MetaInfo ─────────────────────────────────────────────────────────
# Required so GNOME Software / App Center shows the correct icon and metadata.
cat > "$STAGE_DIR/usr/share/metainfo/io.github.jamescog.custos.metainfo.xml" <<EOF
<?xml version="1.0" encoding="UTF-8"?>
<component type="desktop-application">
  <id>io.github.jamescog.custos</id>
  <launchable type="desktop-id">custos.desktop</launchable>
  <name>Custos</name>
  <summary>System battery status and notifications</summary>
  <description>
    <p>Custos monitors your battery level and sends desktop notifications when it
    reaches your configured upper or lower threshold.</p>
  </description>
  <url type="homepage">https://github.com/jamescog/custos</url>
  <metadata_license>MIT</metadata_license>
  <project_license>MIT</project_license>
  <categories>
    <category>Utility</category>
    <category>System</category>
  </categories>
  <provides>
    <binary>custos</binary>
  </provides>
  <releases>
    <release version="${VERSION}" date="$(date -u +%Y-%m-%d)"/>
  </releases>
</component>
EOF

# ── DEBIAN/control ─────────────────────────────────────────────────────────────
cat > "$STAGE_DIR/DEBIAN/control" <<EOF
Package: custos
Version: $VERSION
Section: utils
Priority: optional
Architecture: $ARCH
Maintainer: Jamescog <jamescog72@gmail.com>
Description: System battery monitor and notifier
 Custos monitors your battery and sends desktop notifications when the
 charge level reaches your configured upper or lower threshold.
EOF

# ── DEBIAN/postinst ────────────────────────────────────────────────────────────
# Refreshes the desktop DB and icon cache immediately after install so
# the app appears in the launcher/search without requiring a logout.
cat > "$STAGE_DIR/DEBIAN/postinst" <<'POSTINST'
#!/bin/sh
set -e

# Rebuild the desktop entry database
if command -v update-desktop-database >/dev/null 2>&1; then
    update-desktop-database -q /usr/share/applications || true
fi

# Rebuild the icon cache — try all known binary names
for cmd in gtk-update-icon-cache gtk-update-icon-cache-3.0 gtk4-update-icon-cache; do
    if command -v "$cmd" >/dev/null 2>&1; then
        "$cmd" -f -t /usr/share/icons/hicolor || true
        break
    fi
done

# Rebuild the AppStream cache so GNOME Software shows the right icon
if command -v appstreamcli >/dev/null 2>&1; then
    appstreamcli refresh --force 2>/dev/null || true
fi

exit 0
POSTINST
chmod 0755 "$STAGE_DIR/DEBIAN/postinst"

# ── DEBIAN/postrm ──────────────────────────────────────────────────────────────
cat > "$STAGE_DIR/DEBIAN/postrm" <<'POSTRM'
#!/bin/sh
set -e

if command -v update-desktop-database >/dev/null 2>&1; then
    update-desktop-database -q /usr/share/applications || true
fi

for cmd in gtk-update-icon-cache gtk-update-icon-cache-3.0 gtk4-update-icon-cache; do
    if command -v "$cmd" >/dev/null 2>&1; then
        "$cmd" -f -t /usr/share/icons/hicolor || true
        break
    fi
done

if command -v appstreamcli >/dev/null 2>&1; then
    appstreamcli refresh --force 2>/dev/null || true
fi

exit 0
POSTRM
chmod 0755 "$STAGE_DIR/DEBIAN/postrm"

dpkg-deb --build "$STAGE_DIR" "$OUT_DIR/${APP_NAME}_${VERSION}_${ARCH}.deb"
printf '%s\n' "$OUT_DIR/${APP_NAME}_${VERSION}_${ARCH}.deb"