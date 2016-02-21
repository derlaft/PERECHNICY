package block

import (
	. "server/game"
	. "util"
)

type Listener func(pt Point, c *Control)
