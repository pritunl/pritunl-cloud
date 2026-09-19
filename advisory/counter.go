package advisory

type Counter struct {
	Count   int
	Max     int
	Pending int
}

func (c *Counter) Add(adv *Advisory, state string, dismissed bool) {
	if adv == nil || adv.Dismissed || dismissed {
		return
	}

	if !adv.Complete {
		c.Pending += 1
	}

	if state == Unreachable {
		return
	}

	if adv.Score >= High {
		c.Count += 1
	}
	if adv.Score > c.Max {
		c.Max = adv.Score
	}
}
