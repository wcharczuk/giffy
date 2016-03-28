#! /bin/bash

set -e

TEST_FILTER=""
TEST_MODE=""
TEST_VERBOSITY=""
ARG_FILTER=""
ARG_PACKAGE=""
ARG_ROOT=""

for i in "$@"
do
case $i in
	--root=*)
	ARG_ROOT="${i#*=}"
	;;
	-f=*|--filter=*)
	ARG_FILTER="${i#*=}"
	;;
	-p=*|--package=*)
	ARG_PACKAGE="${i#*=}"
	;;
	--short)
	TEST_MODE="-short"
	;;
	--verbose)
	TEST_VERBOSITY="-v"
	;;
	--default)
	DEFAULT=YES
	;;
	*)
		# unknown option
	;;
esac
done

if [ ! -z "$ARG_ROOT" ]; then
	TEST_ROOT="$ARG_ROOT"
fi

if [ ! -z "$ARG_PACKAGE" ]; then
	TEST_ROOT="${TEST_ROOT}${ARG_PACKAGE}/"
fi

if [ ! -z "$ARG_FILTER" ]; then
	TEST_FILTER="-run ${ARG_FILTER}"
fi

genv -f="./_config/config.json" go test $TEST_VERBOSITY $TEST_FILTER "${TEST_ROOT}..."