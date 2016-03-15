package client

import "strconv"

// UID is a user ID number limit that is currently sufficient.
// TODO(tmrts): Use big.Int if necessary later on.
type UID uint64

// ParseUID puts a given byte slice into a UID integer type.
func ParseUID(buf []byte) (UID, error) {
	n, err := strconv.ParseUint(string(buf), 10, 64)

	return UID(n), err
}

// UIDSet allows easy member tests where a collection of clients are kept.
type UIDSet map[UID]struct{}

// Contains check whether the given UID is in the set.
func (set UIDSet) Contains(n UID) bool {
	_, ok := set[n]

	return ok
}

// Add adds a new UID member to the UIDSet.
func (set *UIDSet) Add(n UID) {
	(*set)[n] = struct{}{}
}
