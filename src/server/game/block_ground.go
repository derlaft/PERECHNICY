package game

type Ground struct {
}

func (b Ground) Is() map[string]bool {
	return map[string]bool{
		"Solid": false,
	}
}

func (b Ground) Has() map[string]byte {
	return nil
}
