package dbclaim

func NormalizeLeaseSeconds(requested int32, fallback int32, maximum int32) int32 {
	if requested <= 0 {
		requested = fallback
	}
	if maximum > 0 && requested > maximum {
		requested = maximum
	}
	return requested
}

func Until(nowUnix int64, leaseSeconds int32) int64 {
	if leaseSeconds <= 0 {
		return nowUnix
	}
	return nowUnix + int64(leaseSeconds)
}
