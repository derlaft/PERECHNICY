project=$(dirname $(readlink -f "$BASH_SOURCE"))
echo $project
export GOPATH=$project/vendor:$project
unset project
