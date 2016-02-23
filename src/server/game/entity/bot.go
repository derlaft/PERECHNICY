package entity

import (
	"encoding/json"
	"figex/mio"
	"fmt"
	. "server/game"
	. "util"
)

var (
	handlers map[byte]Interrupt = map[byte]Interrupt{
		0x01: MoveInt,
		0x02: ScanInt,
		0xf0: AttackInt,
	}
)

const (
	DIRECTION_UP    = 0
	DIRECTION_RIGHT = 1
	DIRECTION_DOWN  = 2
	DIRECTION_LEFT  = 3
	VIDEO_SIZE      = 5
	QSZ             = VIDEO_SIZE * VIDEO_SIZE
	MOVEMENT_AP     = 1
	ENERGY_START    = 10000
	BOT_DAMAGE      = 4
	HP_START        = 100
)

type Interrupt func(*Bot, *Control)

type Bot struct {
	Prog   *mio.Prog
	Energy uint
	AP     uint
	HP     uint
}

type JSONOutput struct {
	Reg       [16]byte
	IP        int
	Inst      string
	Health    uint
	Energy    uint
	AP        uint
	Destroyed bool
	Map       [3][VIDEO_SIZE * VIDEO_SIZE]byte
}

func NewBot(prog *mio.Prog) *Bot {
	bot := Bot{
		Prog:   prog,
		Energy: ENERGY_START,
		HP:     HP_START,
	}
	return &bot
}

func (b *Bot) Health() uint {
	return b.HP
}

func (b *Bot) OnDamage(c *Control, dmg uint) {
	b.HP = uint(Max(0, int(b.HP)-int(dmg)))
}

func (b *Bot) Tick(c *Control) {

	inc, max := b.APParams(c)

	b.AP = (b.AP + inc) % max
	b.Energy -= 0

	if b.Energy == 0 || b.HP == 0 {
		//self-destruct
		c.Destroy()
	}

	retInt := b.Prog.Tick()
	if retInt != nil { //handle hwint
		handler := handlers[*retInt] //TODO: replace incorrect int with FLT
		if handler != nil {
			handler(b, c)
		}
	}
}

func (b *Bot) APParams(c *Control) (increment uint, max uint) {
	return 1, 25
}

func (b *Bot) Byte(c *Control) byte {
	return ENTITY_BOT
}

func MoveInt(b *Bot, c *Control) {

	if b.AP > MOVEMENT_AP {
		b.AP -= MOVEMENT_AP
	} else {
		b.Prog.Flt()
		return
	}

	var add Point

	switch b.Prog.State.Reg[0] % 4 {
	case DIRECTION_UP:
		add = Point{0, -1}
	case DIRECTION_RIGHT:
		add = Point{1, 0}
	case DIRECTION_DOWN:
		add = Point{0, 1}
	case DIRECTION_LEFT:
		add = Point{-1, 0}
	}

	if !c.Move(c.Location.Add(add)) {
		b.Prog.Flt()
	}
}

func (b *Bot) DumpRegisters() (out string) {

	header := ""

	regs := b.Prog.State.Reg

	for i, reg := range regs[:15] {
		header += fmt.Sprintf(":R%X ", i)
		out += fmt.Sprintf("&%02x ", reg)
	}

	header += "_GELIFOZ "
	out += fmt.Sprintf("%08b ", int64(regs[15]))

	header += "IP   "
	out += fmt.Sprintf("%04d ", b.Prog.State.IP)

	if b.Prog.State.IP < len(b.Prog.Prog) {
		header += "INST"
		out += b.Prog.Prog[b.Prog.State.IP].InstName + " "
	}

	return header + "\n" + out
}

// copy non-overlaping parts of array
func memcpy(arr []byte, from, to, n int) {
	for i := 0; i < n && from+i < len(arr) && to+i < len(arr); i++ {
		arr[to+i] = arr[from+i]
	}
}

func AttackInt(b *Bot, c *Control) {
	for pt := range EachPoint(
		c.Location.Add(Point{-1, -1}),
		c.Location.Add(Point{+1, +1}),
	) {
		e := c.Game.EntityAt(*pt)
		if e != nil {
			e.Entity.OnDamage(e, BOT_DAMAGE)
		}
	}

}

func ScanInt(b *Bot, c *Control) {

	//copy buffers
	for i := 1; i >= 0; i-- {
		memcpy(b.Prog.State.Mem[:], i*QSZ, (i+1)*QSZ, QSZ)
	}

	//do scan
	from := c.Location.Add(Point{-(VIDEO_SIZE / 2), -(VIDEO_SIZE / 2)})
	to := c.Location.Add(Point{2, 2})

	for pt := range EachPoint(from, to) {
		cell := ((VIDEO_SIZE/2 + pt.X - c.Location.X) * VIDEO_SIZE) +
			(VIDEO_SIZE/2 + pt.Y - c.Location.Y)
		b.Prog.State.Mem[cell] = c.Game.At(*pt)
	}

}

func (b *Bot) OnColission(me, he *Control) {

}

func (b *Bot) instName() string {
	if len(b.Prog.Prog) <= b.Prog.State.IP {
		return ""
	}
	return b.Prog.Prog[b.Prog.State.IP].InstName
}

func (b *Bot) JSON(c *Control) []byte {
	out := JSONOutput{
		Reg:       b.Prog.State.Reg,
		IP:        b.Prog.State.IP,
		Inst:      b.instName(),
		Map:       [3][VIDEO_SIZE * VIDEO_SIZE]byte{},
		Health:    b.Health(),
		Energy:    b.Energy,
		AP:        b.AP,
		Destroyed: c.Destroyed,
	}

	for i := 0; i < 3; i++ {
		out.Map[i] = [VIDEO_SIZE * VIDEO_SIZE]byte{}
		copy(out.Map[i][:], b.Prog.State.Mem[i*VIDEO_SIZE*VIDEO_SIZE:(i+1)*VIDEO_SIZE*VIDEO_SIZE])
	}

	ret, err := json.Marshal(out)
	if err != nil {
		panic(err)
	}
	return ret
}
