project=$(dirname $(readlink -f "$BASH_SOURCE"))
echo $project
export GOPATH=$project:$project/vendor
unset project
