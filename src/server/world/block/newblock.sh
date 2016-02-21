cat << EOF > $(echo $1 | tr '[A-Z]' '[a-z]').go
package blocks

type $1 struct {
}

func (g $1) Listeners() map[int]Listener {

	return map[int]Listener{}

}

func (g $1) Solid() bool {
	return false
}
EOF
