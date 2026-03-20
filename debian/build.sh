#!/usr/bin/env bash
set -euo pipefail

dir=$(dirname "$0")
cd "$dir" || exit 1

root=$(cd .. && pwd)

version=$(git describe --tags --abbrev=0 2>/dev/null || echo "0.1.0")
version=${version#v}

changelog="$root/debian/changelog"
if [ ! -f "$changelog" ]; then
    echo "warning: changelog not found at $changelog" >&2
else
    if ! grep -q "^ssh-connect (${version})" "$changelog"; then
        if command -v dch >/dev/null 2>&1; then
            if [ -z "${DEBEMAIL:-}" ]; then
                DEBEMAIL=$(git config --get user.email || true)
                export DEBEMAIL
            fi
            dch --check-dirname-level=0 -v "${version}-1" "Automatic release"
        else
            echo "warning: dch not found, please update debian/changelog manually" >&2
        fi
    fi
fi

if ! command -v dpkg-buildpackage >/dev/null; then
    echo "dpkg-buildpackage not found; install build-essential debhelper" >&2
    exit 1
fi

cd ..
export DH_NO_DWZ=1
dpkg-buildpackage -us -uc -b

pkgroot=$(dirname "$0")/..
mkdir -p "$pkgroot/dist"
debfile=$(ls "$root/../ssh-connect_"*.deb 2>/dev/null || true)
if [ -n "$debfile" ]; then
    mv "$debfile" "$pkgroot/dist/"
    echo "package built and moved to dist/$(basename "$debfile")"
else
    echo "package built, see parent directory for .deb files"
fi

# rm -f "$root/../"ssh-connect_*.{buildinfo,changes,dsc,tar.gz} || true
# rm -f "$root/ssh-connect"
# rm -rf "$root/debian/ssh-connect"
# rm -f "$root/debian/debhelper-build-stamp" "$root/debian/files" \
#        "$root/debian/ssh-connect.postrm.debhelper" "$root/debian/ssh-connect.substvars"
# rm -rf "$root/debian/.debhelper"
