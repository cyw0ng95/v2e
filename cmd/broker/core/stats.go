package core

// GetMessageStats returns a copy of the current message statistics.
func (b *Broker) GetMessageStats() MessageStats {
	return b.bus.GetMessageStats()
}

// GetPerProcessStats returns a copy of current per-process stats.
func (b *Broker) GetPerProcessStats() map[string]PerProcessStats {
	return b.bus.GetPerProcessStats()
}

// GetMessageCount returns the total number of messages processed (sent + received).
func (b *Broker) GetMessageCount() int64 {
	return b.bus.GetMessageCount()
}
