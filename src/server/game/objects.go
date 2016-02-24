package game

var (
	Blocks = BlocksMap{
		TILE_GROUND:   Ground{},
		TILE_WALL:     Wall{},
		TILE_BONES:    Ground{},
		TILE_MUSHROOM: Wall{},
	}

	Entities = EntityMap{
		ENTITY_BEAR:   NewBear,
		ENTITY_BOT:    NewBot,
		ENTITY_KTULHU: NewKtulhu,
	}
)

type BlocksMap map[byte]Block
type EntityMap map[byte]func() Entity

const (
	TILE_GROUND   = byte(0x0)
	TILE_WALL     = byte('#')
	TILE_BONES    = byte('=')
	TILE_MUSHROOM = byte('"')
	ENTITY_BEAR   = byte('b')
	ENTITY_BOT    = byte('A')
	ENTITY_KTULHU = byte('K')

	Solid = "SOLID"
)

type Block interface {
	Is() map[string]bool
	Has() map[string]byte
}

func (b BlocksMap) Is(block byte, quality string) bool {
	handle, found := b[block]
	if !found {
		return false
	}

	if isMap := handle.Is(); found && isMap != nil {

		result, found := handle.Is()[quality]
		return found && result
	}
	return false
}

func (b BlocksMap) Has(block byte, quality string) byte {
	handle, found := b[block]
	if !found {
		return 0
	}

	if hasMap := handle.Has(); hasMap != nil {
		result, _ := handle.Has()[quality]
		return result
	}
	return 0
}
