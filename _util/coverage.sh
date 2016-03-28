#! /bin/sh

: ${ROOT:?"ROOT is required"}

echo "mode: set" > ./profile.cov

for dir in $(find ${ROOT} -maxdepth 10 -not -path '*/testdata' -type d);
do
if ls $dir/*.go &> /dev/null; then
	genv -f="./_config/config.json" go test -short -covermode=set -coverprofile=$dir/profile.tmp $dir
	if [ -f $dir/profile.tmp ]; then
		cat $dir/profile.tmp | tail -n +2 >> profile.cov
		rm $dir/profile.tmp
	fi
fi
done

genv -f="config.json" go tool cover -html=profile.cov

rm profile.cov