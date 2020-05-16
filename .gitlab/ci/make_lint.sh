#!/bin/bash
#
# parameters: <required linter output file>
set -e

PKG=$1
LINTER_OUTPUT=$2

PKG_LIST=$(go list ${PKG}/... | grep -v /vendor/)

golint -set_exit_status ${PKG_LIST} | tee ${LINTER_OUTPUT}
stat=${PIPESTATUS[0]}
if [ $stat -ne 0 ]; then
 exit $stat
fi
git grep -i "todo" -- ":/" ":!.gitlab/ci/make_lint.sh" | tee -a ${LINTER_OUTPUT}
stat=${PIPESTATUS[0]}
if [ $stat -eq 0 ]; then
	echo "There should not be any TODOs in the code."
	echo "Remove them and add proper tasks to Jira instead."
 exit 1
fi
if [ $stat -eq 1 ]; then
	exit 0
fi
exit $stat
