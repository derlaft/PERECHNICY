package game

type Wall struct {
}

func (b Wall) Is() map[string]bool {
	return map[string]bool{
		Solid: true,
	}
}

func (b Wall) Has() map[string]byte {
	return nil
}
