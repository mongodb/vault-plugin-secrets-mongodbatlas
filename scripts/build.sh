#!/usr/bin/env bash

DATABASE_TOOL="mongodbatlas-database-plugin"
SECRET_TOOL="vault-plugin-secrets-mongodbatlas"
DATABASE_DIR="plugins/database/mongodbatlas/"
SECRET_DIR="plugins/logical/mongodbatlas/cmd/"

# This script builds the application from source for multiple platforms.
set -e

# Get the parent directory of where this script is.
SOURCE="${BASH_SOURCE[0]}"
while [ -h "$SOURCE" ] ; do SOURCE="$(readlink "$SOURCE")"; done
DIR="$( cd -P "$( dirname "$SOURCE" )/.." && pwd )"

# Change into that directory
cd "$DIR"

# Set build tags
BUILD_TAGS="${BUILD_TAGS}:-${TOOL}"

# Get the git commit
GIT_COMMIT="$(git rev-parse HEAD)"
GIT_DIRTY="$(test -n "`git status --porcelain`" && echo "+CHANGES" || true)"

# Determine the arch/os combos we're building for
XC_ARCH=${XC_ARCH:-"386 amd64"}
XC_OS=${XC_OS:-linux darwin windows freebsd openbsd netbsd solaris}
XC_OSARCH=${XC_OSARCH:-"linux/386 linux/amd64 linux/arm linux/arm64 darwin/386 darwin/amd64 windows/386 windows/amd64 freebsd/386 freebsd/amd64 freebsd/arm openbsd/386 openbsd/amd64 openbsd/arm netbsd/386 netbsd/amd64 netbsd/arm solaris/amd64"}

GOPATH=${GOPATH:-$(go env GOPATH)}
case $(uname) in
    CYGWIN*)
        GOPATH="$(cygpath $GOPATH)"
        ;;
esac

# Delete the old dir
echo "==> Removing old directory..."
rm -f bin/*
rm -rf pkg/*
mkdir -p bin/

# If its dev mode, only build for our self
if [ "${VAULT_DEV_BUILD}x" != "x" ]; then
    XC_OS=$(go env GOOS)
    XC_ARCH=$(go env GOARCH)
    XC_OSARCH=$(go env GOOS)/$(go env GOARCH)
fi

# If its dev mode, only build for our self
if [ "${DOCKER_BUILD}x" != "x" ]; then
    XC_OS="linux"
    XC_ARCH="amd64"
    XC_OSARCH="linux/amd64"
fi


# Build Database Plugin tool
cd "${DIR}/${DATABASE_DIR}/${DATABASE_TOOL}"

# Delete the old dir
echo "==> Removing old directory..."
rm -f bin/*
rm -rf pkg/*
mkdir -p bin/

# Build!
echo "==> Building..."
gox \
    -osarch="${XC_OSARCH}" \
    -ldflags "-X github.com/hashicorp/${DATABASE_TOOL}/version.GitCommit='${GIT_COMMIT}${GIT_DIRTY}'" \
    -output "${DIR}/pkg/{{.OS}}_{{.Arch}}/${DATABASE_TOOL}" \
    -tags="${BUILD_TAGS}" \
    .

# Return to the home directory
cd "$DIR"


# Build Secret Engine
cd "${DIR}/${SECRET_DIR}/${SECRET_TOOL}"

# Delete the old dir
echo "==> Removing old directory..."
rm -f bin/*
rm -rf pkg/*
mkdir -p bin/

# Build!
echo "==> Building..."
gox \
    -osarch="${XC_OSARCH}" \
    -ldflags "-X github.com/hashicorp/${SECRET_TOOL}/version.GitCommit='${GIT_COMMIT}${GIT_DIRTY}'" \
    -output "${DIR}/pkg/{{.OS}}_{{.Arch}}/${SECRET_TOOL}" \
    -tags="${BUILD_TAGS}" \
    .

# Return to the home directory
cd "$DIR"

# Move all the compiled things to the $GOPATH/bin
OLDIFS=$IFS
IFS=: MAIN_GOPATH=($GOPATH)
IFS=$OLDIFS

# Copy our OS/Arch to the bin/ directory
DEV_PLATFORM="./pkg/$(go env GOOS)_$(go env GOARCH)"
for F in $(find ${DEV_PLATFORM} -mindepth 1 -maxdepth 1 -type f); do
    cp ${F} bin/
    cp ${F} ${MAIN_GOPATH}/bin/
done

if [ "${DOCKER_BUILD}x" = "x" ]; then
    # Zip and copy to the dist dir
    echo "==> Packaging..."
    for PLATFORM in $(find ./pkg -mindepth 1 -maxdepth 1 -type d); do
        OSARCH=$(basename ${PLATFORM})
        echo "--> ${OSARCH}"

        pushd $PLATFORM >/dev/null 2>&1
        zip ../${OSARCH}.zip ./*
        popd >/dev/null 2>&1
    done
fi

if [ "${VAULT_DEV_TEST_BUILD}x" = "x" ]; then
    # Zip and copy to the dist dir
    echo "==> Packaging..."
    for PLATFORM in $(find ./pkg -mindepth 1 -maxdepth 1 -type d); do
        OSARCH=$(basename ${PLATFORM})
        echo "--> ${OSARCH}"

        pushd $PLATFORM >/dev/null 2>&1
        zip ../${OSARCH}.zip ./*
        popd >/dev/null 2>&1
    done

    # Copy our OS/Arch to the bin/ directory
    DEV_PLATFORM="./pkg/linux_amd64"
    for F in $(find ${DEV_PLATFORM} -mindepth 1 -maxdepth 1 -type f); do
        cp ${F} bin/
        cp ${F} ${MAIN_GOPATH}/bin/
    done
fi

# Done!
echo
echo "==> Results:"
ls -hl bin/
