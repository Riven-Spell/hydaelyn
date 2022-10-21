package updates

func GetNeededUpdates(currentUpdate uint64) []DBUpdate {
	for k, v := range Versions {
		if v.Version > currentUpdate {
			return Versions[k:]
		}
	}

	return nil
}
