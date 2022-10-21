package updates

type sortByVersion []DBUpdate

func (s sortByVersion) Len() int {
	return len(s)
}

func (s sortByVersion) Less(i, j int) bool {
	return s[i].Version < s[j].Version
}

func (s sortByVersion) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
