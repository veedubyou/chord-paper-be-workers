#!/bin/sh

# This file creates/updates ./GIT-VERSION-FILE, for use by
# Github Action Workflows for docker deployment.  It generates a
# version number based on `git describe` and writes it into the above files.
# Keep in mind that these files are only changed during the docker build process;
# they exist in the repo with a placeholder string to prevent errors locally.
# Therefore, this script shouldn't even be run in dev, unless you feel like
# dirtying the repo.

GVF=GIT-VERSION-FILE
DEF_VER=v0.69.0

LF='
'

# First see if there is a version file (included in release tarballs),
# then try git-describe, then default.
if test -f version
then
        VN=$(cat version) || VN="$DEF_VER"
elif test -d ${GIT_DIR:-.git} -o -f .git &&
        VN=$(git describe --match "v[0-9]*" HEAD 2>/dev/null) &&
        case "$VN" in
        *$LF*) (exit 1) ;;
        v[0-9]*)
                git update-index -q --refresh
                test -z "$(git diff-index --name-only HEAD --)" ||
                VN="$VN-dirty" ;;
        esac
then
        VN=$(echo "$VN" | sed -e 's/-/./g');
else
        VN="$DEF_VER"
fi

VN=$(expr "$VN" : v*'\(.*\)')

if test -r $GVF
then
        VC=$(sed -e 's/^//' <$GVF)
else
        VC=unset
fi
test "$VN" = "$VC" || {
        echo >&2 "$VN"
        echo "$VN" > $GVF
}
