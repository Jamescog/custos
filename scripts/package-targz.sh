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
OUT_DIR="${OUT_DIR:-$ROOT_DIR/dist/tar}"
STAGE_DIR="$OUT_DIR/${APP_NAME}_${VERSION}_linux_${ARCH}"
BIN_DIR="$ROOT_DIR/build/bin"

if [[ "$ARCH" == "arm64" ]]; then
  if [[ -z "${CC:-}" ]] && command -v aarch64-linux-gnu-gcc >/dev/null 2>&1; then
    export CC="aarch64-linux-gnu-gcc"
  fi
  if [[ -z "${CXX:-}" ]] && command -v aarch64-linux-gnu-g++ >/dev/null 2>&1; then
    export CXX="aarch64-linux-gnu-g++"
  fi

  if [[ -z "${CC:-}" ]] || ! command -v "$CC" >/dev/null 2>&1; then
    printf '%s\n' "ARCH=arm64 needs cgo cross-compiler (aarch64)." >&2
    printf '%s\n' "Install toolchain (Debian/Ubuntu): sudo apt-get install gcc-aarch64-linux-gnu g++-aarch64-linux-gnu" >&2
    printf '%s\n' "Or build on arm64 machine." >&2
    exit 1
  fi
fi

cd "$ROOT_DIR"
rm -rf "$STAGE_DIR"
mkdir -p "$STAGE_DIR"

if [[ -z "${SKIP_BUILD:-}" ]]; then
  wails build -platform "linux/$ARCH"
fi

BIN_PATH="$(find "$BIN_DIR" -maxdepth 1 -type f -perm -111 -name "${APP_NAME}*" | head -n 1)"
if [[ -z "${BIN_PATH:-}" ]]; then
  printf 'built binary not found in %s\n' "$BIN_DIR" >&2
  exit 1
fi

install -m 0755 "$BIN_PATH" "$STAGE_DIR/$APP_NAME"

if [[ -f "$ROOT_DIR/build/appicon.png" ]]; then
  install -m 0644 "$ROOT_DIR/build/appicon.png" "$STAGE_DIR/$APP_NAME.png"
fi

cat > "$STAGE_DIR/$APP_NAME.desktop" <<EOF
[Desktop Entry]
Version=1.0
Name=Custos
GenericName=Battery Monitor
Comment=System battery status and notifications
Keywords=Custos;Battery;Monitor;Power;
Exec=$APP_NAME
Terminal=false
Type=Application
Categories=Utility;HardwareSettings;System;
Icon=$APP_NAME
StartupNotify=false
EOF

cat > "$STAGE_DIR/install.sh" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail

APP_NAME="custos"
MODE="system"

while [[ $# -gt 0 ]]; do
  case "$1" in
    --user)
      MODE="user"
      shift
      ;;
    --system)
      MODE="system"
      shift
      ;;
    -h|--help)
      printf '%s\n' "Usage: ./install.sh [--user|--system]"
      exit 0
      ;;
    *)
      printf '%s\n' "unknown arg: $1" >&2
      exit 2
      ;;
  esac
done

ROOT="$(cd "$(dirname "$0")" && pwd)"

if [[ "$MODE" == "system" ]]; then
  if [[ "$(id -u)" -ne 0 ]]; then
    printf '%s\n' "--system requires root. Re-run: sudo ./install.sh --system" >&2
    exit 1
  fi

  install -m 0755 "$ROOT/$APP_NAME" "/usr/local/bin/$APP_NAME"
  install -Dm 0644 "$ROOT/$APP_NAME.desktop" "/usr/share/applications/$APP_NAME.desktop"
  if [[ -f "$ROOT/$APP_NAME.png" ]]; then
    install -Dm 0644 "$ROOT/$APP_NAME.png" "/usr/share/icons/hicolor/256x256/apps/$APP_NAME.png"
  fi

  update-desktop-database /usr/share/applications 2>/dev/null || true
  printf '%s\n' "Installed system-wide. Run: $APP_NAME"
  exit 0
fi

mkdir -p \
  "$HOME/.local/bin" \
  "$HOME/.local/share/applications" \
  "$HOME/.local/share/icons/hicolor/256x256/apps"

install -m 0755 "$ROOT/$APP_NAME" "$HOME/.local/bin/$APP_NAME"
install -m 0644 "$ROOT/$APP_NAME.desktop" "$HOME/.local/share/applications/$APP_NAME.desktop"
if [[ -f "$ROOT/$APP_NAME.png" ]]; then
  install -m 0644 "$ROOT/$APP_NAME.png" "$HOME/.local/share/icons/hicolor/256x256/apps/$APP_NAME.png"
fi

update-desktop-database "$HOME/.local/share/applications" 2>/dev/null || true
printf '%s\n' "Installed for current user. Ensure ~/.local/bin in PATH. Run: $APP_NAME"
EOF
chmod 0755 "$STAGE_DIR/install.sh"

cat > "$STAGE_DIR/uninstall.sh" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail

APP_NAME="custos"
MODE="system"

while [[ $# -gt 0 ]]; do
  case "$1" in
    --user)
      MODE="user"
      shift
      ;;
    --system)
      MODE="system"
      shift
      ;;
    -h|--help)
      printf '%s\n' "Usage: ./uninstall.sh [--user|--system]"
      exit 0
      ;;
    *)
      printf '%s\n' "unknown arg: $1" >&2
      exit 2
      ;;
  esac
done

if [[ "$MODE" == "system" ]]; then
  if [[ "$(id -u)" -ne 0 ]]; then
    printf '%s\n' "--system requires root. Re-run: sudo ./uninstall.sh --system" >&2
    exit 1
  fi

  rm -f "/usr/local/bin/$APP_NAME"
  rm -f "/usr/share/applications/$APP_NAME.desktop"
  rm -f "/usr/share/icons/hicolor/256x256/apps/$APP_NAME.png"

  update-desktop-database /usr/share/applications 2>/dev/null || true
  printf '%s\n' "Uninstalled system-wide."
  exit 0
fi

rm -f "$HOME/.local/bin/$APP_NAME"
rm -f "$HOME/.local/share/applications/$APP_NAME.desktop"
rm -f "$HOME/.local/share/icons/hicolor/256x256/apps/$APP_NAME.png"

update-desktop-database "$HOME/.local/share/applications" 2>/dev/null || true
printf '%s\n' "Uninstalled for current user."
EOF
chmod 0755 "$STAGE_DIR/uninstall.sh"

cat > "$STAGE_DIR/README.txt" <<EOF
Custos - Linux release bundle

Run:
  ./$APP_NAME

Install:
  sudo ./install.sh

Install (no root):
  ./install.sh --user

Uninstall:
  sudo ./uninstall.sh

Uninstall (no root):
  ./uninstall.sh --user

Desktop integration (no root):
  mkdir -p ~/.local/bin ~/.local/share/applications ~/.local/share/icons/hicolor/256x256/apps
  cp $APP_NAME ~/.local/bin/
  cp $APP_NAME.desktop ~/.local/share/applications/
  cp $APP_NAME.png ~/.local/share/icons/hicolor/256x256/apps/ 2>/dev/null || true
  update-desktop-database ~/.local/share/applications 2>/dev/null || true

System dependencies (typical):
  - GTK3
  - WebKit2GTK
  - libayatana-appindicator / libappindicator (tray support depends on DE)

If app fails to start, run from terminal to see missing shared library name.
EOF

TARBALL="$OUT_DIR/${APP_NAME}_${VERSION}_linux_${ARCH}.tar.gz"
mkdir -p "$OUT_DIR"

tar -C "$OUT_DIR" -czf "$TARBALL" "$(basename "$STAGE_DIR")"
printf '%s\n' "$TARBALL"
